package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// QuotaChecker handles checking mobile data quota
type QuotaChecker struct {
	client *http.Client

	// tokenCache caches the dynamic xl-ku.my.id header (name+value) that the
	// site rotates. It is refreshed automatically when stale.
	tokenMu     sync.Mutex
	tokenHeader string
	tokenValue  string
	tokenExpiry time.Time
}

// NewQuotaChecker creates a new quota checker service
func NewQuotaChecker() *QuotaChecker {
	return &QuotaChecker{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// xlKuBaseURL is the endpoint origin for xl-ku.my.id
const xlKuBaseURL = "https://xl-ku.my.id"

// xlKuCheckPackageJS is the obfuscated JS file that contains the token
// builder. The site rotates token values here on every deploy.
const xlKuCheckPackageJS = xlKuBaseURL + "/xlkujs/check-package?v=1.0.1"

// xlKuTokenTTL is how long a scraped token is considered fresh before we
// re-scrape the JS for a new one.
const xlKuTokenTTL = 10 * time.Minute

// xlKuFallbackHeader / xlKuFallbackValue are used only if scraping fails
// and no previously cached value exists. Update when site rotates token.
const (
	xlKuFallbackHeader = "xl-5d966be717"
	xlKuFallbackValue  = "WLsZ_qz-ORj4O8j3ZeuxwOtLmCw3gDqErI8W0jZqW9EWZkgFAohKDsBRGzyMWNw2wfWPadd5YTCNJGXGZbZiLID1peKQjYSyDD5K_g0362P-XiM2rC3hndhnE5TRkire5-JK2GyRHemflXRxTH36DGmYwchFJZM5RX_3KSwuF0P3vCfr"
)

// ── JS obfuscation decoder helpers ───────────────────────────────────────────

// jsFuncBodyRe extracts all zero-arg function bodies from the obfuscated JS.
// Matches: function NAME(){ ... } where the body may contain nested {}
var jsFuncBodyRe = regexp.MustCompile(`function (\w+)\(\)\{`)

// parseJSFunctions scans src and returns a map of functionName -> body
// using a balanced-brace scanner so nested anonymous functions inside
// atob().map(function(c){...}) are handled correctly.
func parseJSFunctions(src string) map[string]string {
	out := make(map[string]string)
	idxs := jsFuncBodyRe.FindAllStringSubmatchIndex(src, -1)
	for _, idx := range idxs {
		name := src[idx[2]:idx[3]]
		start := idx[1] // just after the opening '{'
		depth := 1
		i := start
		for i < len(src) && depth > 0 {
			switch src[i] {
			case '{':
				depth++
			case '}':
				depth--
			}
			i++
		}
		body := strings.TrimSpace(src[start : i-1])
		out[name] = body
	}
	return out
}

// headerBuilderRe matches: function X(){return {[Y()]:Z()};}  or similar;
// we use it to find the zero-arg function that returns the header object.
var headerBuilderRe = regexp.MustCompile(`function (\w+)\(\)\{return \{\[(\w+)\(\)\]:(\w+)\(\)\}`)

// findHeaderBuilderFns searches src for the function that returns {[keyFn()]:valFn()}
// and returns (objFn, keyFn, valFn) names, or empty strings if not found.
func findHeaderBuilderFns(src string) (string, string, string) {
	m := headerBuilderRe.FindStringSubmatch(src)
	if len(m) == 4 {
		return m[1], m[2], m[3]
	}
	return "", "", ""
}

// decodeJSFunc recursively evaluates an obfuscated zero-arg JS function
// and returns the string value it produces at runtime.
func decodeJSFunc(name string, funcs map[string]string, depth int) string {
	if depth > 12 {
		return ""
	}
	body, ok := funcs[name]
	if !ok {
		return ""
	}

	// ── Pattern 1: push/join array builder ──────────────────────────────────
	// var arr=[]; arr.push(fn1()); arr.push(fn2()); return arr.join('');
	if strings.Contains(body, ".push(") && strings.Contains(body, ".join('')") {
		pushRe := regexp.MustCompile(`\.push\((\w+)\(\)\)`)
		matches := pushRe.FindAllStringSubmatch(body, -1)
		var sb strings.Builder
		for _, m := range matches {
			sb.WriteString(decodeJSFunc(m[1], funcs, depth+1))
		}
		return sb.String()
	}

	// ── Pattern 2: XOR atob decode ───────────────────────────────────────────
	// var v="b64"; var k=N; return atob(v).split('').map(function(c){return String.fromCharCode(c.charCodeAt(0)^k);}).join('');
	xorRe := regexp.MustCompile(`var \w+="([A-Za-z0-9+/=]+)";\s*var \w+=(\d+);\s*return atob`)
	if xm := xorRe.FindStringSubmatch(body); len(xm) == 3 && strings.Contains(body, "^") {
		if decoded, err := base64.StdEncoding.DecodeString(xm[1]); err == nil {
			var xorVal int
			fmt.Sscanf(xm[2], "%d", &xorVal)
			var sb strings.Builder
			for _, b := range decoded {
				sb.WriteByte(byte(int(b) ^ xorVal))
			}
			return sb.String()
		}
	}

	// ── Pattern 3: SUB atob decode ───────────────────────────────────────────
	// var v="b64"; var k=N; return atob(v).split('').map(function(c){return String.fromCharCode(c.charCodeAt(0)-k);}).join('');
	subRe := regexp.MustCompile(`var \w+="([A-Za-z0-9+/=]+)";\s*var \w+=(\d+);\s*return atob`)
	if sm := subRe.FindStringSubmatch(body); len(sm) == 3 && strings.Contains(body, "charCodeAt(0)-") {
		if decoded, err := base64.StdEncoding.DecodeString(sm[1]); err == nil {
			var subVal int
			fmt.Sscanf(sm[2], "%d", &subVal)
			var sb strings.Builder
			for _, b := range decoded {
				sb.WriteByte(byte(int(b) - subVal))
			}
			return sb.String()
		}
	}

	// ── Pattern 4: plain atob ────────────────────────────────────────────────
	// var v="b64"; return atob(v);
	atobRe := regexp.MustCompile(`"([A-Za-z0-9+/=]+)";\s*return atob`)
	if am := atobRe.FindStringSubmatch(body); len(am) == 2 {
		if decoded, err := base64.StdEncoding.DecodeString(am[1]); err == nil {
			return string(decoded)
		}
	}

	// ── Pattern 5: reverse ──────────────────────────────────────────────────
	// var v="literal"; return v.split('').reverse().join('');
	revRe := regexp.MustCompile(`var \w+="([^"]+)";\s*return \w+\.split\(''\)\.reverse`)
	if rm := revRe.FindStringSubmatch(body); len(rm) == 2 {
		r := []rune(rm[1])
		for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
		return string(r)
	}

	// ── Pattern 6: trim literal ──────────────────────────────────────────────
	// var v="literal"; return v.trim();
	trimRe := regexp.MustCompile(`var \w+="([^"]+)";\s*return \w+\.trim\(\)`)
	if tm := trimRe.FindStringSubmatch(body); len(tm) == 2 {
		return strings.TrimSpace(tm[1])
	}

	// ── Pattern 7: literal var return ───────────────────────────────────────
	// var v="literal"; return v;
	litRe := regexp.MustCompile(`var \w+="([^"]+)";\s*return \w+;?$`)
	if lm := litRe.FindStringSubmatch(body); len(lm) == 2 {
		return lm[1]
	}

	// ── Pattern 8: inline string literal ────────────────────────────────────
	// return "literal";
	inlRe := regexp.MustCompile(`^return "([^"]+)";?$`)
	if im := inlRe.FindStringSubmatch(strings.TrimSpace(body)); len(im) == 2 {
		return im[1]
	}

	// ── Pattern 9: concat of sub-calls ──────────────────────────────────────
	// return fn1()+fn2()+fn3();
	concatRe := regexp.MustCompile(`^return (.+?);?$`)
	if cm := concatRe.FindStringSubmatch(strings.TrimSpace(body)); len(cm) == 2 {
		callsRe := regexp.MustCompile(`(\w+)\(\)`)
		calls := callsRe.FindAllStringSubmatch(cm[1], -1)
		var sb strings.Builder
		for _, c := range calls {
			sb.WriteString(decodeJSFunc(c[1], funcs, depth+1))
		}
		if sb.Len() > 0 {
			return sb.String()
		}
	}

	return ""
}

// QuotaResponse represents the API response structure
type QuotaResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		SubsInfo struct {
			MSISDN      string `json:"msisdn"`
			Operator    string `json:"operator"`
			IDVerified  string `json:"id_verified"`
			NetType     string `json:"net_type"`
			Tenure      string `json:"tenure"`
			ExpDate     string `json:"exp_date"`
			GraceUntil  string `json:"grace_until"`
			Volte       struct {
				Device  bool `json:"device"`
				Area    bool `json:"area"`
				Simcard bool `json:"simcard"`
			} `json:"volte"`
		} `json:"subs_info"`
		PackageInfo struct {
			ErrorMessage interface{} `json:"error_message"`
			Packages     []struct {
				Name      string `json:"name"`
				Expiry    string `json:"expiry"`
				Timestamp int64  `json:"timestamp"`
				Quotas    []struct {
					Name      string      `json:"name"`
					Percent   interface{} `json:"percent"`
					Total     string      `json:"total"`
					Remaining string      `json:"remaining"`
				} `json:"quotas"`
			} `json:"packages"`
		} `json:"package_info"`
	} `json:"data"`
}

// NormalizePhoneNumber normalizes phone number to format 08xxxxx
func (qc *QuotaChecker) NormalizePhoneNumber(number string) string {
	// Remove spaces, dashes, and other non-numeric characters
	number = strings.ReplaceAll(number, " ", "")
	number = strings.ReplaceAll(number, "-", "")
	number = strings.ReplaceAll(number, "+", "")

	// Convert 628xxx to 08xxx
	if strings.HasPrefix(number, "628") {
		number = "0" + number[2:]
	} else if strings.HasPrefix(number, "62") {
		number = "0" + number[2:]
	}

	// Ensure it starts with 08
	if !strings.HasPrefix(number, "08") {
		return ""
	}

	return number
}

// CheckQuota checks the quota for a given phone number
func (qc *QuotaChecker) CheckQuota(phoneNumber string) (string, error) {
	// Normalize phone number
	normalized := qc.NormalizePhoneNumber(phoneNumber)
	if normalized == "" {
		return "", fmt.Errorf("Format nomor tidak valid. Gunakan format: 08xxx atau 628xxx")
	}

	// Build API URL — endpoint resmi dari front-end xl-ku.my.id
	url := fmt.Sprintf("%s/check/all-info/%s", xlKuBaseURL, normalized)

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("Gagal membuat request: %v", err)
	}

	// Dapatkan token header dinamis dari homepage (site merotasi token).
	headerName, headerValue, err := qc.getXLKuToken()
	if err != nil {
		return "", fmt.Errorf("Gagal mengambil token xl-ku: %v", err)
	}

	// Add headers to mimic the browser request exactly like the site's JS does.
	// xl-ku.my.id (Cloudflare) will answer 404 for requests without the
	// custom token header or without browser-like headers.
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://xl-ku.my.id/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Sec-Ch-Ua", `"Not;A=Brand";v="8", "Chromium";v="150", "Brave";v="150"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Linux"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set(headerName, headerValue)

	// Execute request
	resp, err := qc.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Gagal mengakses API: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Gagal membaca response: %v", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API mengembalikan error: HTTP %d", resp.StatusCode)
	}

	// Parse JSON response
	var quotaResp QuotaResponse
	if err := json.Unmarshal(body, &quotaResp); err != nil {
		return "", fmt.Errorf("Gagal parsing JSON response: %v", err)
	}

	// Check if API call was successful
	if !quotaResp.Success {
		return "", fmt.Errorf("API error: %s - %s", quotaResp.Code, quotaResp.Message)
	}

	// Format the response
	return qc.FormatQuotaResponse(&quotaResp), nil
}

// getXLKuToken returns the current dynamic header (name + value) that
// xl-ku.my.id expects. The site rotates the token by obfuscating it inside
// /xlkujs/check-package. We decode it at runtime and cache for xlKuTokenTTL.
func (qc *QuotaChecker) getXLKuToken() (string, string, error) {
	qc.tokenMu.Lock()
	defer qc.tokenMu.Unlock()

	// Return cached token if still fresh
	if qc.tokenHeader != "" && time.Now().Before(qc.tokenExpiry) {
		return qc.tokenHeader, qc.tokenValue, nil
	}

	headerName, value, err := qc.scrapeXLKuToken()
	if err != nil {
		// Fallback to last-known cached value
		if qc.tokenHeader != "" {
			return qc.tokenHeader, qc.tokenValue, nil
		}
		// Last resort: baked-in constant
		return xlKuFallbackHeader, xlKuFallbackValue, nil
	}

	qc.tokenHeader = headerName
	qc.tokenValue = value
	qc.tokenExpiry = time.Now().Add(xlKuTokenTTL)
	return headerName, value, nil
}

// scrapeXLKuToken fetches /xlkujs/check-package and decodes the obfuscated
// JS to extract the current xl-ku header name and value.
//
// How the site obfuscates the token:
//   - The token is NOT in the homepage HTML. It lives in /xlkujs/check-package.
//   - The JS defines a function that returns  {[keyBuilderFn()]: valueBuilderFn()}
//     using a computed property key  (the header name) and a computed value.
//   - Each builder is a chain of zero-arg functions that concatenate partial
//     strings produced via base64+XOR, base64+SUB, reverse, trim, or literals.
//   - We decode the chain recursively without executing JS.
func (qc *QuotaChecker) scrapeXLKuToken() (string, string, error) {
	req, err := http.NewRequest("GET", xlKuCheckPackageJS, nil)
	if err != nil {
		return "", "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36")
	req.Header.Set("Referer", xlKuBaseURL+"/")
	req.Header.Set("Accept", "*/*")

	resp, err := qc.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("fetch check-package JS: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("read check-package JS: %w", err)
	}
	// Normalise CRLF so our regexes work on every platform
	src := strings.ReplaceAll(string(raw), "\r\n", "\n")
	src = strings.ReplaceAll(src, "\r", "\n")

	// Parse all zero-arg function bodies from the obfuscated JS
	funcs := parseJSFunctions(src)
	if len(funcs) == 0 {
		return "", "", fmt.Errorf("no JS functions found in check-package")
	}

	// Locate the header-object builder: function X(){return {[keyFn()]:valFn()};}
	_, keyFn, valFn := findHeaderBuilderFns(src)
	if keyFn == "" || valFn == "" {
		return "", "", fmt.Errorf("header builder pattern not found in check-package JS")
	}

	headerName := decodeJSFunc(keyFn, funcs, 0)
	if headerName == "" {
		return "", "", fmt.Errorf("could not decode header key from JS function %s", keyFn)
	}
	headerValue := decodeJSFunc(valFn, funcs, 0)
	if headerValue == "" {
		return "", "", fmt.Errorf("could not decode header value from JS function %s", valFn)
	}

	return headerName, headerValue, nil
}

// FormatQuotaResponse formats the quota response into a readable message
func (qc *QuotaChecker) FormatQuotaResponse(resp *QuotaResponse) string {
	var sb strings.Builder

	// Header
	sb.WriteString("INFO KUOTA & PAKET\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Subscription Info
	subs := resp.Data.SubsInfo
	sb.WriteString("*INFORMASI KARTU*\n")
	sb.WriteString(fmt.Sprintf("Nomor: %s\n", subs.MSISDN))
	sb.WriteString(fmt.Sprintf("Provider: %s\n", subs.Operator))
	sb.WriteString(fmt.Sprintf("Jaringan: %s\n", subs.NetType))
	sb.WriteString(fmt.Sprintf("Lama Berlangganan: %s\n", subs.Tenure))
	sb.WriteString(fmt.Sprintf("Masa Aktif: %s\n", subs.ExpDate))
	sb.WriteString(fmt.Sprintf("Grace Period: %s\n\n", subs.GraceUntil))

	// VoLTE Info
	sb.WriteString("*VOLTE STATUS*\n")
	volteDevice := "Tidak"
	if subs.Volte.Device {
		volteDevice = "Ya"
	}
	volteArea := "Tidak"
	if subs.Volte.Area {
		volteArea = "Ya"
	}
	volteSimcard := "Tidak"
	if subs.Volte.Simcard {
		volteSimcard = "Ya"
	}
	sb.WriteString(fmt.Sprintf("Device: %s\n", volteDevice))
	sb.WriteString(fmt.Sprintf("Area: %s\n", volteArea))
	sb.WriteString(fmt.Sprintf("Simcard: %s\n\n", volteSimcard))

	// Package Info
	pkgInfo := resp.Data.PackageInfo
	if pkgInfo.ErrorMessage != nil {
		sb.WriteString(fmt.Sprintf("Error: %v\n", pkgInfo.ErrorMessage))
		return sb.String()
	}

	if len(pkgInfo.Packages) == 0 {
		sb.WriteString("Tidak ada paket aktif.\n")
		return sb.String()
	}

	sb.WriteString("*PAKET AKTIF*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for i, pkg := range pkgInfo.Packages {
		sb.WriteString(fmt.Sprintf("Paket %d: *%s*\n", i+1, pkg.Name))
		sb.WriteString(fmt.Sprintf("Berlaku hingga: %s\n\n", pkg.Expiry))

		if len(pkg.Quotas) > 0 {
			sb.WriteString("Detail Kuota:\n")
			for j, quota := range pkg.Quotas {
				// Convert percent to float
				percentStr := "0"
				switch v := quota.Percent.(type) {
				case float64:
					percentStr = fmt.Sprintf("%.1f", v)
				case int:
					percentStr = fmt.Sprintf("%d", v)
				case string:
					percentStr = v
				}

				sb.WriteString(fmt.Sprintf("\n%d. %s\n", j+1, quota.Name))
				sb.WriteString(fmt.Sprintf("   Total: %s\n", quota.Total))
				sb.WriteString(fmt.Sprintf("   Sisa: %s\n", quota.Remaining))
				sb.WriteString(fmt.Sprintf("   Terpakai: %s%%\n", percentStr))
			}
		}

		if i < len(pkgInfo.Packages)-1 {
			sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
		}
	}

	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("Data berhasil diambil\n")
	sb.WriteString(fmt.Sprintf("Waktu: %s\n", time.Now().Format("02-01-2006 15:04:05")))

	return sb.String()
}

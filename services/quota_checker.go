package services

import (
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
	tokenMu      sync.Mutex
	tokenHeader  string
	tokenValue   string
	tokenExpiry  time.Time
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

// xlKuTokenTTL is how long a scraped token is considered fresh before we
// re-scrape the homepage for a new one. The site rotates the token, so we
// must not cache it forever.
const xlKuTokenTTL = 10 * time.Minute

// xlKuFallbackHeader / xlKuFallbackValue are used only if scraping the
// homepage fails. Update these if the site changes its token scheme and
// scraping is temporarily broken.
const (
	xlKuFallbackHeader = "xl-88740fafdf"
	xlKuFallbackValue  = "4H-4XGlbBYrKNotrg6_yS2JPzkNRg91-oLSakD_60bpMNqxJXcrf_bhkEbuAUfRHUSxGVBgWbf2hrGW8OPbYQOAAjO-sjetMtdpNd_sV4VyU9Zvp7Io2Y2U9W2unETVnwzZ7Hbh1yw"
)

// tokenHeaderRe finds the custom "xl-xxxxxxxxxx": varName used in the
// check/all-info ajax settings block of the homepage. The site rotates this
// token (and the header name), so we scrape the current one from the HTML.
var tokenHeaderRe = regexp.MustCompile(`url:\s*baseUrl\s*\+\s*"check/all-info/[\s\S]*?headers:\s*\{[\s\S]*?"(xl-[0-9a-f]+)":\s*([a-zA-Z0-9_]+)`)

// varValueRe finds `var <name> = "<value>"` assignments in the homepage JS.
var varValueRe = regexp.MustCompile(`var\s+([a-zA-Z0-9_]+)\s*=\s*"([^"]+)"`)

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
// xl-ku.my.id expects. The site rotates this token, so we scrape it from the
// homepage HTML on each call (cached for xlKuTokenTTL to avoid hammering).
// If scraping fails we fall back to the last-known constant.
func (qc *QuotaChecker) getXLKuToken() (string, string, error) {
	qc.tokenMu.Lock()
	defer qc.tokenMu.Unlock()

	// Return cached token if still fresh
	if qc.tokenHeader != "" && time.Now().Before(qc.tokenExpiry) {
		return qc.tokenHeader, qc.tokenValue, nil
	}

	headerName, value, err := qc.scrapeXLKuToken()
	if err != nil {
		// Fallback to last-known constant if we have one cached
		if qc.tokenHeader != "" {
			return qc.tokenHeader, qc.tokenValue, nil
		}
		// No cached token and scrape failed -> use baked-in fallback
		return xlKuFallbackHeader, xlKuFallbackValue, nil
	}

	qc.tokenHeader = headerName
	qc.tokenValue = value
	qc.tokenExpiry = time.Now().Add(xlKuTokenTTL)
	return headerName, value, nil
}

// scrapeXLKuToken fetches the homepage and extracts the current xl-ku token
// header name + value from the embedded JS.
func (qc *QuotaChecker) scrapeXLKuToken() (string, string, error) {
	req, err := http.NewRequest("GET", xlKuBaseURL+"/", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := qc.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	html := string(body)

	// Find the header name + variable name used in the check/all-info settings
	m := tokenHeaderRe.FindStringSubmatch(html)
	if len(m) < 3 {
		return "", "", fmt.Errorf("tidak bisa menemukan header token xl-ku di homepage")
	}
	headerName := m[1]
	varName := m[2]

	// Find the variable's value: var <varName> = "value"
	valueRe := regexp.MustCompile(`var\s+` + regexp.QuoteMeta(varName) + `\s*=\s*"([^"]+)"`)
	vm := valueRe.FindStringSubmatch(html)
	if len(vm) < 2 {
		return "", "", fmt.Errorf("tidak bisa menemukan nilai token untuk %s", varName)
	}

	return headerName, vm[1], nil
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

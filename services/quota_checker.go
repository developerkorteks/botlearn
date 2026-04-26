package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// QuotaChecker handles checking mobile data quota
type QuotaChecker struct {
	client *http.Client
}

// NewQuotaChecker creates a new quota checker service
func NewQuotaChecker() *QuotaChecker {
	return &QuotaChecker{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
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

	// Build API URL
	url := fmt.Sprintf("https://xl-ku.my.id/end.php?check=package&number=%s&version=2", normalized)

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("Gagal membuat request: %v", err)
	}

	// Add headers to mimic browser request
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://bendith.my.id/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

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

	// Check HTTP status - API returns 201 (Created) instead of 200 (OK)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
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

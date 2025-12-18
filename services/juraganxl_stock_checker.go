package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

// JuraganXLStockChecker handles checking product stock from JuraganXL API
type JuraganXLStockChecker struct {
	client    *http.Client
	baseURL   string
	csrfToken string
	mu        sync.Mutex
}

// NewJuraganXLStockChecker creates a new JuraganXL stock checker service
func NewJuraganXLStockChecker() *JuraganXLStockChecker {
	// Create cookie jar for session management
	jar, _ := cookiejar.New(nil)
	
	return &JuraganXLStockChecker{
		client: &http.Client{
			Timeout: 60 * time.Second,
			Jar:     jar,
		},
		baseURL: "https://juraganxl.my.id",
	}
}

// ============================================================================
// DATA MODELS
// ============================================================================

// CSRFResponse represents the CSRF token response
type CSRFResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RegularStock represents XDA regular stock item
type RegularStock struct {
	Config          string         `json:"config"`
	Count           int            `json:"count"`
	QuotaAllocation map[string]int `json:"quota_allocation"` // Area1-4
}

// CircleStock represents XCLP circle stock item
type CircleStock struct {
	Config          string `json:"config"`
	Count           int    `json:"count"`
	QuotaAllocation int    `json:"quota_allocation"` // Simple int
}

// FlexMaxItem represents a FlexMax product
type FlexMaxItem struct {
	Name             string            `json:"name"`
	Price            string            `json:"price"`
	ProductCode      string            `json:"productCode"`
	Description      map[string]string `json:"description"` // Area-based
	Bonus            map[string]string `json:"bonus"`
	Status           string            `json:"status"` // TERSEDIA/HABIS
	BonusInstruction string            `json:"bonusInstruction"`
}

// FlexMaxResponse represents the FlexMax API response
type FlexMaxResponse struct {
	Flexmax map[string]FlexMaxItem `json:"flexmax"` // XL, XXL
}

// FlexManiaItem represents a FlexMania product
type FlexManiaItem struct {
	Name        string            `json:"name"`
	ProductCode string            `json:"productCode"`
	MainQuota   string            `json:"mainQuota"`
	Bonus       map[string]string `json:"bonus"`
}

// FlexManiaResponse represents the FlexMania API response
type FlexManiaResponse struct {
	FlexManiaTable map[string]FlexManiaItem `json:"flexManiaTable"`
}

// JuraganXLStockResult represents combined result from all APIs
type JuraganXLStockResult struct {
	Regular   []RegularStock
	Circle    []CircleStock
	FlexMax   FlexMaxResponse
	FlexMania FlexManiaResponse
	Timestamp time.Time
	Errors    []string // Track partial failures
}

// ============================================================================
// CSRF TOKEN HANDLER
// ============================================================================

// getCSRFToken fetches and stores the CSRF token
func (sc *JuraganXLStockChecker) getCSRFToken() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	csrfURL := fmt.Sprintf("%s/api/csrf-token", sc.baseURL)

	req, err := http.NewRequest("GET", csrfURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create CSRF request: %v", err)
	}

	// Add headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", sc.baseURL)

	resp, err := sc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get CSRF token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CSRF token request failed: HTTP %d", resp.StatusCode)
	}

	// Parse response
	var csrfResp CSRFResponse
	if err := json.NewDecoder(resp.Body).Decode(&csrfResp); err != nil {
		return fmt.Errorf("failed to parse CSRF response: %v", err)
	}

	if !csrfResp.Success {
		return fmt.Errorf("CSRF token initialization failed")
	}

	// Extract CSRF token from cookies
	parsedURL, _ := url.Parse(sc.baseURL)
	cookies := sc.client.Jar.Cookies(parsedURL)
	for _, cookie := range cookies {
		if cookie.Name == "csrf-token" {
			sc.csrfToken = cookie.Value
			return nil
		}
	}

	return fmt.Errorf("CSRF token not found in cookies")
}

// ============================================================================
// API FETCHERS
// ============================================================================

// makeRequest creates and executes an authenticated request
func (sc *JuraganXLStockChecker) makeRequest(endpoint string) ([]byte, error) {
	apiURL := fmt.Sprintf("%s/api/%s", sc.baseURL, endpoint)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", sc.baseURL)
	req.Header.Set("X-CSRF-Token", sc.csrfToken)

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	return body, nil
}

// FetchRegularStock fetches XDA regular stock
func (sc *JuraganXLStockChecker) FetchRegularStock() ([]RegularStock, error) {
	body, err := sc.makeRequest("regulers")
	if err != nil {
		return nil, err
	}

	var stocks []RegularStock
	if err := json.Unmarshal(body, &stocks); err != nil {
		return nil, fmt.Errorf("failed to parse regular stock: %v", err)
	}

	return stocks, nil
}

// FetchCircleStock fetches XCLP circle stock
func (sc *JuraganXLStockChecker) FetchCircleStock() ([]CircleStock, error) {
	body, err := sc.makeRequest("stocks-circle")
	if err != nil {
		return nil, err
	}

	var stocks []CircleStock
	if err := json.Unmarshal(body, &stocks); err != nil {
		return nil, fmt.Errorf("failed to parse circle stock: %v", err)
	}

	return stocks, nil
}

// FetchFlexMax fetches FlexMax products
func (sc *JuraganXLStockChecker) FetchFlexMax() (FlexMaxResponse, error) {
	body, err := sc.makeRequest("flexmax")
	if err != nil {
		return FlexMaxResponse{}, err
	}

	var response FlexMaxResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return FlexMaxResponse{}, fmt.Errorf("failed to parse FlexMax: %v", err)
	}

	return response, nil
}

// FetchFlexMania fetches FlexMania products
func (sc *JuraganXLStockChecker) FetchFlexMania() (FlexManiaResponse, error) {
	body, err := sc.makeRequest("flexmax-table")
	if err != nil {
		return FlexManiaResponse{}, err
	}

	var response FlexManiaResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return FlexManiaResponse{}, fmt.Errorf("failed to parse FlexMania: %v", err)
	}

	return response, nil
}

// ============================================================================
// MAIN CHECK STOCK
// ============================================================================

// CheckStock fetches stock from all APIs concurrently
func (sc *JuraganXLStockChecker) CheckStock() (*JuraganXLStockResult, error) {
	// Step 1: Get CSRF token first
	if err := sc.getCSRFToken(); err != nil {
		return nil, fmt.Errorf("failed to get CSRF token: %v", err)
	}

	result := &JuraganXLStockResult{
		Timestamp: time.Now(),
		Errors:    []string{},
	}

	// Step 2: Fetch all data concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Fetch Regular Stock (XDA)
	wg.Add(1)
	go func() {
		defer wg.Done()
		stocks, err := sc.FetchRegularStock()
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Regular Stock: %v", err))
		} else {
			result.Regular = stocks
		}
	}()

	// Fetch Circle Stock (XCLP)
	wg.Add(1)
	go func() {
		defer wg.Done()
		stocks, err := sc.FetchCircleStock()
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Circle Stock: %v", err))
		} else {
			result.Circle = stocks
		}
	}()

	// Fetch FlexMax
	wg.Add(1)
	go func() {
		defer wg.Done()
		flexmax, err := sc.FetchFlexMax()
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("FlexMax: %v", err))
		} else {
			result.FlexMax = flexmax
		}
	}()

	// Fetch FlexMania
	wg.Add(1)
	go func() {
		defer wg.Done()
		flexmania, err := sc.FetchFlexMania()
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("FlexMania: %v", err))
		} else {
			result.FlexMania = flexmania
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()

	// Check if all APIs failed
	if len(result.Errors) == 4 {
		return nil, fmt.Errorf("all APIs failed: %s", strings.Join(result.Errors, "; "))
	}

	return result, nil
}

// ============================================================================
// RESPONSE FORMATTER
// ============================================================================

// FormatStockResponse formats the stock response into a readable message
func (sc *JuraganXLStockChecker) FormatStockResponse(result *JuraganXLStockResult) string {
	var sb strings.Builder

	// Header
	sb.WriteString("CEK STOCK PRODUK JURAGANXL\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Summary
	regularReady := sc.countReady(result.Regular)
	circleReady := sc.countCircleReady(result.Circle)
	flexMaxReady := sc.countFlexMaxReady(result.FlexMax)
	flexManiaCount := len(result.FlexMania.FlexManiaTable)

	sb.WriteString("📊 RINGKASAN:\n")
	sb.WriteString(fmt.Sprintf("• XDA Ready: %d / %d\n", regularReady, len(result.Regular)))
	sb.WriteString(fmt.Sprintf("• XCLP Ready: %d / %d\n", circleReady, len(result.Circle)))
	sb.WriteString(fmt.Sprintf("• FlexMax: %d / 2\n", flexMaxReady))
	sb.WriteString(fmt.Sprintf("• FlexMania: %d produk\n\n", flexManiaCount))

	// Format each section
	sc.formatRegularStock(&sb, result.Regular)
	sc.formatCircleStock(&sb, result.Circle)
	sc.formatFlexMax(&sb, result.FlexMax)
	sc.formatFlexMania(&sb, result.FlexMania)

	// Footer
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("✅ Data berhasil diambil\n")
	sb.WriteString(fmt.Sprintf("🕒 Waktu: %s\n", result.Timestamp.Format("02-01-2006 15:04:05")))

	// Show errors if any
	if len(result.Errors) > 0 {
		sb.WriteString(fmt.Sprintf("\n⚠️ Partial errors: %d\n", len(result.Errors)))
	}

	return sb.String()
}

// countReady counts ready products in regular stock
func (sc *JuraganXLStockChecker) countReady(stocks []RegularStock) int {
	count := 0
	for _, stock := range stocks {
		if stock.Count > 0 {
			count++
		}
	}
	return count
}

// countCircleReady counts ready products in circle stock
func (sc *JuraganXLStockChecker) countCircleReady(stocks []CircleStock) int {
	count := 0
	for _, stock := range stocks {
		if stock.Count > 0 {
			count++
		}
	}
	return count
}

// countFlexMaxReady counts available FlexMax products
func (sc *JuraganXLStockChecker) countFlexMaxReady(response FlexMaxResponse) int {
	count := 0
	for _, item := range response.Flexmax {
		if strings.ToUpper(item.Status) == "TERSEDIA" {
			count++
		}
	}
	return count
}

// formatRegularStock formats XDA regular stock
func (sc *JuraganXLStockChecker) formatRegularStock(sb *strings.Builder, stocks []RegularStock) {
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("📱 STOK XDA (Reguler)\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Separate ready and empty
	var ready, empty []RegularStock
	for _, stock := range stocks {
		if stock.Count > 0 {
			ready = append(ready, stock)
		} else {
			empty = append(empty, stock)
		}
	}

	// Show ready products
	if len(ready) > 0 {
		sb.WriteString(fmt.Sprintf("✅ READY (%d):\n", len(ready)))
		for i, stock := range ready {
			sb.WriteString(fmt.Sprintf("%d. %s - Stock: %d\n", i+1, stock.Config, stock.Count))
			sb.WriteString(fmt.Sprintf("   Area1: %dGB | Area2: %dGB | Area3: %dGB | Area4: %dGB\n",
				stock.QuotaAllocation["Area1"],
				stock.QuotaAllocation["Area2"],
				stock.QuotaAllocation["Area3"],
				stock.QuotaAllocation["Area4"]))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("✅ READY: Tidak ada\n\n")
	}

	// Show empty products (limit to first 5)
	if len(empty) > 0 {
		limit := 5
		if len(empty) < limit {
			limit = len(empty)
		}
		sb.WriteString(fmt.Sprintf("❌ EMPTY (%d, showing %d):\n", len(empty), limit))
		for i := 0; i < limit; i++ {
			stock := empty[i]
			sb.WriteString(fmt.Sprintf("%d. %s - Stock: 0\n", i+1, stock.Config))
		}
		if len(empty) > limit {
			sb.WriteString(fmt.Sprintf("   ... dan %d produk lainnya\n", len(empty)-limit))
		}
		sb.WriteString("\n")
	}
}

// formatCircleStock formats XCLP circle stock
func (sc *JuraganXLStockChecker) formatCircleStock(sb *strings.Builder, stocks []CircleStock) {
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("🎯 STOK XCLP (Circle)\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Separate ready and empty
	var ready, empty []CircleStock
	for _, stock := range stocks {
		if stock.Count > 0 {
			ready = append(ready, stock)
		} else {
			empty = append(empty, stock)
		}
	}

	// Show ready products
	if len(ready) > 0 {
		sb.WriteString(fmt.Sprintf("✅ READY (%d):\n", len(ready)))
		for i, stock := range ready {
			sb.WriteString(fmt.Sprintf("%d. %s - Stock: %d | Kuota: %dGB\n",
				i+1, stock.Config, stock.Count, stock.QuotaAllocation))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("✅ READY: Tidak ada\n\n")
	}

	// Show empty products (limit to first 5)
	if len(empty) > 0 {
		limit := 5
		if len(empty) < limit {
			limit = len(empty)
		}
		sb.WriteString(fmt.Sprintf("❌ EMPTY (%d, showing %d):\n", len(empty), limit))
		for i := 0; i < limit; i++ {
			stock := empty[i]
			sb.WriteString(fmt.Sprintf("%d. %s - Kuota: %dGB\n", i+1, stock.Config, stock.QuotaAllocation))
		}
		if len(empty) > limit {
			sb.WriteString(fmt.Sprintf("   ... dan %d produk lainnya\n", len(empty)-limit))
		}
		sb.WriteString("\n")
	}
}

// formatFlexMax formats FlexMax products
func (sc *JuraganXLStockChecker) formatFlexMax(sb *strings.Builder, response FlexMaxResponse) {
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("🌟 FLEXMAX (Area)\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	if len(response.Flexmax) == 0 {
		sb.WriteString("Tidak ada data FlexMax\n\n")
		return
	}

	i := 1
	for _, item := range response.Flexmax {
		status := "✅"
		if strings.ToUpper(item.Status) != "TERSEDIA" {
			status = "❌"
		}

		sb.WriteString(fmt.Sprintf("%d. %s (%s) - %s %s\n",
			i, item.Name, item.ProductCode, item.Price, status))

		if item.Description != nil {
			sb.WriteString(fmt.Sprintf("   Area1: %s\n", item.Description["Area1"]))
			sb.WriteString(fmt.Sprintf("   Area2: %s\n", item.Description["Area2"]))
			sb.WriteString(fmt.Sprintf("   Area3: %s\n", item.Description["Area3"]))
			sb.WriteString(fmt.Sprintf("   Area4: %s\n", item.Description["Area4"]))
		}

		i++
	}
	sb.WriteString("\n")
}

// formatFlexMania formats FlexMania products
func (sc *JuraganXLStockChecker) formatFlexMania(sb *strings.Builder, response FlexManiaResponse) {
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("🔥 FLEXMANIA (Nasional)\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	if len(response.FlexManiaTable) == 0 {
		sb.WriteString("Tidak ada data FlexMania\n\n")
		return
	}

	// Sort by product code for consistent display
	codes := []string{"FM6", "FM10", "FM15", "FM20", "FM25", "FM32", "FM36"}
	i := 1
	for _, code := range codes {
		if item, exists := response.FlexManiaTable[code]; exists {
			sb.WriteString(fmt.Sprintf("%d. %s - %s\n", i, item.Name, item.MainQuota))
			i++
		}
	}
	sb.WriteString("\n")
}

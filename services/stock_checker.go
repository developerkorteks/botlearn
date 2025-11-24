package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// StockChecker handles checking product stock from multiple APIs
type StockChecker struct {
	client *http.Client
}

// NewStockChecker creates a new stock checker service
func NewStockChecker() *StockChecker {
	return &StockChecker{
		client: &http.Client{
			Timeout: 60 * time.Second, // Increase timeout for multiple concurrent requests
		},
	}
}

// ProductItem represents a product
type ProductItem struct {
	KodeProduk     string `json:"kode_produk"`
	NamaProduk     string `json:"nama_produk"`
	Deskripsi      string `json:"deskripsi"`
	HargaOriginal  int    `json:"harga_original"`
	HargaFinal     int    `json:"harga_final"`
	Stok           int    `json:"stok"`
	Type           string `json:"type"`
}

// StockResponse represents the API response
type StockResponse struct {
	Success       bool          `json:"success"`
	Type          string        `json:"type"`
	Title         string        `json:"title"`
	ProvidersUsed []string      `json:"providers_used"`
	Data          StockData     `json:"data"`
}

// StockData represents the data field in response
type StockData struct {
	Ready []ProductItem `json:"ready"`
	Empty []ProductItem `json:"empty"`
}

// StockResult represents combined result from all APIs
type StockResult struct {
	BPA StockResponse // Produk Harian
	XDA StockResponse // Produk BulananV2
	XLA StockResponse // Produk Bulanan
}

// FetchStock fetches stock from a specific API
func (sc *StockChecker) FetchStock(stockType string) (*StockResponse, error) {
	// Build API URL
	url := fmt.Sprintf("https://ics-store.my.id/api.php?action=fetchProducts&type=%s", stockType)

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Gagal membuat request: %v", err)
	}

	// Add headers
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8")
	req.Header.Set("Referer", "https://ics-store.my.id/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Chromium";v="142", "Brave";v="142", "Not_A Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Linux"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sec-gpc", "1")

	// Execute request
	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Gagal mengakses API: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Gagal membaca response: %v", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API mengembalikan error: HTTP %d", resp.StatusCode)
	}

	// Parse JSON response
	var stockResp StockResponse
	if err := json.Unmarshal(body, &stockResp); err != nil {
		return nil, fmt.Errorf("Gagal parsing JSON response: %v", err)
	}

	// Check if API call was successful
	if !stockResp.Success {
		return nil, fmt.Errorf("API error untuk type: %s", stockType)
	}

	return &stockResp, nil
}

// CheckStock fetches stock from all 3 APIs concurrently
func (sc *StockChecker) CheckStock() (*StockResult, error) {
	result := &StockResult{}
	
	// Use WaitGroup for concurrent API calls
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := []string{}

	// Fetch BPA (Harian)
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := sc.FetchStock("bpa")
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errors = append(errors, fmt.Sprintf("BPA: %v", err))
		} else {
			result.BPA = *resp
		}
	}()

	// Fetch XDA (BulananV2)
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := sc.FetchStock("xda")
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errors = append(errors, fmt.Sprintf("XDA: %v", err))
		} else {
			result.XDA = *resp
		}
	}()

	// Fetch XLA (Bulanan)
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := sc.FetchStock("xla")
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errors = append(errors, fmt.Sprintf("XLA: %v", err))
		} else {
			result.XLA = *resp
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()

	// Check if all APIs failed
	if len(errors) == 3 {
		return nil, fmt.Errorf("Semua API gagal: %s", strings.Join(errors, "; "))
	}

	return result, nil
}

// FormatStockResponse formats the stock response into a readable message
func (sc *StockChecker) FormatStockResponse(result *StockResult) string {
	var sb strings.Builder

	// Header
	sb.WriteString("CEK STOCK PRODUK\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Count total products
	totalReady := len(result.BPA.Data.Ready) + len(result.XDA.Data.Ready) + len(result.XLA.Data.Ready)
	totalEmpty := len(result.BPA.Data.Empty) + len(result.XDA.Data.Empty) + len(result.XLA.Data.Empty)

	sb.WriteString(fmt.Sprintf("Total Produk Ready: %d\n", totalReady))
	sb.WriteString(fmt.Sprintf("Total Produk Empty: %d\n\n", totalEmpty))

	// Format each API response
	sc.formatSingleStock(&sb, &result.BPA)
	sc.formatSingleStock(&sb, &result.XDA)
	sc.formatSingleStock(&sb, &result.XLA)

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("Data berhasil diambil\n")
	sb.WriteString(fmt.Sprintf("Waktu: %s\n", time.Now().Format("02-01-2006 15:04:05")))

	return sb.String()
}

// formatSingleStock formats a single stock response
func (sc *StockChecker) formatSingleStock(sb *strings.Builder, stock *StockResponse) {
	if stock.Title == "" {
		return // Skip if no data
	}

	sb.WriteString(fmt.Sprintf("*%s*\n", stock.Title))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Show ready products
	if len(stock.Data.Ready) > 0 {
		sb.WriteString(fmt.Sprintf("READY (%d):\n", len(stock.Data.Ready)))
		for i, product := range stock.Data.Ready {
			sb.WriteString(fmt.Sprintf("%d. %s - Stock: %d\n", i+1, product.NamaProduk, product.Stok))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("READY: Tidak ada\n\n")
	}

	// Show empty products (show all)
	if len(stock.Data.Empty) > 0 {
		sb.WriteString(fmt.Sprintf("EMPTY (%d):\n", len(stock.Data.Empty)))
		for i, product := range stock.Data.Empty {
			sb.WriteString(fmt.Sprintf("%d. %s - Stock: %d\n", i+1, product.NamaProduk, product.Stok))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("EMPTY: Tidak ada\n\n")
	}
}


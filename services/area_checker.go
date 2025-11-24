package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AreaChecker handles checking area/region information
type AreaChecker struct {
	client    *http.Client
	areaCache []AreaItem
	cacheTime time.Time
}

// AreaItem represents an area/region
type AreaItem struct {
	Value string `json:"value"` // L1, L2, L3, L4
	Label string `json:"label"` // Kab. Demak, Kota Jakarta, etc
}

// AreaListResponse represents the API response for area list
type AreaListResponse struct {
	Akrab []AreaItem `json:"akrab"`
}

// NewAreaChecker creates a new area checker service
func NewAreaChecker() *AreaChecker {
	return &AreaChecker{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		areaCache: []AreaItem{},
	}
}

// FetchAreaList fetches the area list from API
func (ac *AreaChecker) FetchAreaList() error {
	// Check cache (valid for 1 hour)
	if len(ac.areaCache) > 0 && time.Since(ac.cacheTime) < 1*time.Hour {
		return nil
	}

	// Build API URL
	url := "https://bendith.my.id/area_list.json"

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("Gagal membuat request: %v", err)
	}

	// Add headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://bendith.my.id/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Chromium";v="142", "Brave";v="142", "Not_A Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Linux"`)

	// Execute request
	resp, err := ac.client.Do(req)
	if err != nil {
		return fmt.Errorf("Gagal mengakses API: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Gagal membaca response: %v", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API mengembalikan error: HTTP %d", resp.StatusCode)
	}

	// Parse JSON response
	var areaResp AreaListResponse
	if err := json.Unmarshal(body, &areaResp); err != nil {
		return fmt.Errorf("Gagal parsing JSON response: %v", err)
	}

	// Update cache
	ac.areaCache = areaResp.Akrab
	ac.cacheTime = time.Now()

	return nil
}

// SearchArea searches for an area by name with fuzzy matching
func (ac *AreaChecker) SearchArea(query string) ([]AreaItem, error) {
	// Fetch area list
	if err := ac.FetchAreaList(); err != nil {
		return nil, err
	}

	// Normalize query (lowercase, trim spaces)
	query = strings.ToLower(strings.TrimSpace(query))
	
	if query == "" {
		return nil, fmt.Errorf("Query tidak boleh kosong")
	}

	// Search results
	exactMatches := []AreaItem{}
	partialMatches := []AreaItem{}
	fuzzyMatches := []AreaItem{}

	for _, area := range ac.areaCache {
		labelLower := strings.ToLower(area.Label)
		
		// Remove common prefixes for better matching
		labelClean := strings.TrimPrefix(labelLower, "kab. ")
		labelClean = strings.TrimPrefix(labelClean, "kota ")
		labelClean = strings.TrimPrefix(labelClean, "prov. ")
		
		queryClean := strings.TrimPrefix(query, "kab. ")
		queryClean = strings.TrimPrefix(queryClean, "kota ")
		queryClean = strings.TrimPrefix(queryClean, "prov. ")
		queryClean = strings.TrimPrefix(queryClean, "kabupaten ")
		queryClean = strings.TrimPrefix(queryClean, "provinsi ")

		// Exact match (after cleaning)
		if labelClean == queryClean {
			exactMatches = append(exactMatches, area)
			continue
		}

		// Contains match
		if strings.Contains(labelClean, queryClean) {
			partialMatches = append(partialMatches, area)
			continue
		}

		// Fuzzy match (check similarity for typo tolerance)
		if ac.isSimilar(labelClean, queryClean) {
			fuzzyMatches = append(fuzzyMatches, area)
		}
	}

	// Priority: exact > partial > fuzzy
	results := []AreaItem{}
	results = append(results, exactMatches...)
	results = append(results, partialMatches...)
	results = append(results, fuzzyMatches...)

	// Limit to top 10 results
	if len(results) > 10 {
		results = results[:10]
	}

	return results, nil
}

// isSimilar checks if two strings are similar (for typo tolerance)
func (ac *AreaChecker) isSimilar(s1, s2 string) bool {
	// Levenshtein distance based similarity
	if len(s1) == 0 || len(s2) == 0 {
		return false
	}

	// Make sure s1 is the longer string
	if len(s2) > len(s1) {
		s1, s2 = s2, s1
	}

	// If length difference is too big, not similar
	lenDiff := len(s1) - len(s2)
	if lenDiff > 3 { // Allow max 3 characters difference
		return false
	}

	// Calculate Levenshtein distance
	distance := ac.levenshteinDistance(s1, s2)
	
	// Calculate similarity ratio (1 - distance/maxLength)
	maxLen := len(s1)
	similarity := 1.0 - (float64(distance) / float64(maxLen))
	
	return similarity >= 0.7 // 70% similarity threshold
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func (ac *AreaChecker) levenshteinDistance(s1, s2 string) int {
	len1 := len(s1)
	len2 := len(s2)
	
	// Create matrix
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}
	
	// Initialize first row and column
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}
	
	// Fill matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}
	
	return matrix[len1][len2]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// CheckArea checks area information for a given query
func (ac *AreaChecker) CheckArea(query string) (string, error) {
	// Search for area
	results, err := ac.SearchArea(query)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", fmt.Errorf("Area tidak ditemukan untuk: %s", query)
	}

	// Format response
	return ac.FormatAreaResponse(query, results), nil
}

// FormatAreaResponse formats the area response into a readable message
func (ac *AreaChecker) FormatAreaResponse(query string, results []AreaItem) string {
	var sb strings.Builder

	// Header
	sb.WriteString("INFORMASI AREA\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Query info
	sb.WriteString(fmt.Sprintf("Pencarian: *%s*\n", query))
	sb.WriteString(fmt.Sprintf("Ditemukan: %d hasil\n\n", len(results)))

	// Results
	sb.WriteString("*HASIL PENCARIAN*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for i, area := range results {
		// Convert L4 to Area 4 format
		areaLevel := strings.Replace(area.Value, "L", "Area ", 1)
		
		sb.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, area.Label))
		sb.WriteString(fmt.Sprintf("   Area: %s\n", areaLevel))
		
		if i < len(results)-1 {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("Data berhasil diambil\n")
	sb.WriteString(fmt.Sprintf("Waktu: %s\n", time.Now().Format("02-01-2006 15:04:05")))

	return sb.String()
}

// getAreaLevelDescription returns description for area level
func (ac *AreaChecker) getAreaLevelDescription(level string) string {
	switch level {
	case "L1":
		return "Area Terpencil (Level 1)"
	case "L2":
		return "Area Jauh (Level 2)"
	case "L3":
		return "Area Sedang (Level 3)"
	case "L4":
		return "Area Dekat/Kota (Level 4)"
	default:
		return "Unknown"
	}
}

package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type CheckBugHTTPHeaders struct {
	Server      string `json:"server,omitempty"`
	CFRay       string `json:"cf_ray,omitempty"`
	CFCache     string `json:"cf_cache,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type CheckBugSNIResult struct {
	Supported *bool       `json:"supported,omitempty"`
	Info      interface{} `json:"info,omitempty"`
}

type CheckBugIPEntry struct {
	Cloudflare *bool              `json:"cloudflare,omitempty"`
	WhoisExcerpt string           `json:"whois_excerpt,omitempty"`
	SNI        CheckBugSNIResult  `json:"sni"`
}

type CheckBugDomainResult struct {
	Domain      string                       `json:"domain"`
	ResolvedIPs []string                     `json:"resolved_ips"`
	IPs         map[string]CheckBugIPEntry   `json:"ips"`
	HTTPStatus  *int                          `json:"http_status,omitempty"`
	HTTPHeaders CheckBugHTTPHeaders          `json:"http_headers"`
	DurationSec float64                      `json:"duration_s"`
	Error       string                       `json:"error,omitempty"`
}

type CheckBugService struct {
	Timeout time.Duration
	Retries int
	Threads int
}

func NewCheckBugService() *CheckBugService {
	return &CheckBugService{
		Timeout: 8 * time.Second,
		Retries: 2,
		Threads: 4,
	}
}

func (s *CheckBugService) resolve(domain string) []string {
	ips, _ := net.LookupIP(domain)
	set := map[string]struct{}{}
	for _, ip := range ips {
		set[ip.String()] = struct{}{}
	}
	res := make([]string, 0, len(set))
	for k := range set {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

func (s *CheckBugService) checkWhoisCloudflare(ip string) (*bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "whois", ip)
	out, err := cmd.CombinedOutput()
	if err != nil && ctx.Err() == context.DeadlineExceeded {
		return nil, "whois timed out"
	}
	low := strings.ToLower(string(out))
	excerpt := firstLines(string(out), 8)
	if strings.Contains(low, "cloudflare") || strings.Contains(low, "as13335") || strings.Contains(low, "cloudflarenet") {
		b := true
		return &b, excerpt
	}
	b := false
	return &b, excerpt
}

func (s *CheckBugService) sniCheck(ip, domain string) (supported *bool, info interface{}) {
	// Try TLS with SNI
	dialer := &net.Dialer{Timeout: s.Timeout}
	conf := &tls.Config{ServerName: domain, InsecureSkipVerify: true}
	var lastErr error
	for i := 0; i < s.Retries; i++ {
		conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(ip, "443"), conf)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i+1) * 300 * time.Millisecond)
			continue
		}
		state := conn.ConnectionState()
		conn.Close()
		if len(state.PeerCertificates) > 0 {
			cert := state.PeerCertificates[0]
			sans := cert.DNSNames
			ok := false
			for _, san := range sans {
				if san == domain || (strings.HasPrefix(san, "*.") && strings.HasSuffix(domain, san[1:])) {
					ok = true
					break
				}
			}
			return &ok, map[string]interface{}{"subject": cert.Subject.String(), "subjectAltName": sans, "issuer": cert.Issuer.String()}
		}
		// No cert info
		nilVal := (*bool)(nil)
		return nilVal, "no cert info"
	}
	return nil, fmt.Sprintf("sni error: %v", lastErr)
}

func (s *CheckBugService) httpStatus(domain string) (*int, CheckBugHTTPHeaders) {
	h := CheckBugHTTPHeaders{}
	client := &http.Client{Timeout: s.Timeout}
	// try https then http
	for _, scheme := range []string{"https://", "http://"} {
		req, _ := http.NewRequest("GET", scheme+domain, nil)
		req.Header.Set("User-Agent", "cfceks-go/1.0")
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		status := resp.StatusCode
		h.Server = resp.Header.Get("Server")
		h.CFRay = resp.Header.Get("CF-RAY")
		h.CFCache = resp.Header.Get("CF-Cache-Status")
		h.ContentType = resp.Header.Get("Content-Type")
		return &status, h
	}
	return nil, h
}

func (s *CheckBugService) InspectDomain(domain string) CheckBugDomainResult {
	start := time.Now()
	res := CheckBugDomainResult{Domain: domain, IPs: map[string]CheckBugIPEntry{}}
	ips := s.resolve(domain)
	res.ResolvedIPs = ips
	if len(ips) == 0 {
		res.Error = "resolve_failed"
		res.DurationSec = time.Since(start).Seconds()
		return res
	}
	for _, ip := range ips {
		entry := CheckBugIPEntry{}
		cf, excerpt := s.checkWhoisCloudflare(ip)
		entry.Cloudflare = cf
		entry.WhoisExcerpt = excerpt
		sniSup, sniInfo := s.sniCheck(ip, domain)
		entry.SNI = CheckBugSNIResult{Supported: sniSup, Info: sniInfo}
		res.IPs[ip] = entry
	}
	status, headers := s.httpStatus(domain)
	res.HTTPStatus = status
	res.HTTPHeaders = headers
	res.DurationSec = time.Since(start).Seconds()
	return res
}

func (s *CheckBugService) InspectDomains(domains []string) ([]CheckBugDomainResult, string) {
	// sequential simple to avoid excessive resource usage
	out := make([]CheckBugDomainResult, 0, len(domains))
	var buf bytes.Buffer
	for _, d := range domains {
		r := s.InspectDomain(strings.TrimSpace(d))
		out = append(out, r)
		buf.WriteString(formatPretty(r))
		buf.WriteString("\n")
	}
	return out, buf.String()
}

func firstLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n { lines = lines[:n] }
	return strings.Join(lines, "\n")
}

func formatPretty(r CheckBugDomainResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n*Hasil Pemeriksaan*: %s\n", r.Domain))
	if len(r.ResolvedIPs) == 0 {
		b.WriteString("- ❌ Resolve gagal\n")
		return b.String()
	}
	b.WriteString(fmt.Sprintf("- 🔎 IP: %s\n", strings.Join(r.ResolvedIPs, ", ")))
	for _, ip := range r.ResolvedIPs {
		entry := r.IPs[ip]
		cfStr := "⚠ Unknown"
		if entry.Cloudflare != nil {
			if *entry.Cloudflare { cfStr = "✅ Cloudflare" } else { cfStr = "❌ Non-Cloudflare" }
		}
		b.WriteString(fmt.Sprintf("  • %s → %s\n", ip, cfStr))
		if entry.SNI.Supported != nil {
			if *entry.SNI.Supported {
				b.WriteString("    - ✅ SNI: Didukung\n")
			} else {
				b.WriteString("    - ❌ SNI: Tidak\n")
			}
		} else {
			b.WriteString(fmt.Sprintf("    - ⚠ SNI: %v\n", entry.SNI.Info))
		}
	}
	if r.HTTPStatus != nil {
		b.WriteString(fmt.Sprintf("- 🌐 HTTP: %d\n", *r.HTTPStatus))
	} else {
		b.WriteString("- 🌐 HTTP: unreachable\n")
	}
	if ct := r.HTTPHeaders.ContentType; ct != "" {
		b.WriteString("  • content-type: "+ct+"\n")
	}
	if r.HTTPHeaders.Server != "" {
		b.WriteString("  • server: "+r.HTTPHeaders.Server+"\n")
	}
	if r.HTTPHeaders.CFRay != "" {
		b.WriteString("  • cf-ray: "+r.HTTPHeaders.CFRay+"\n")
	}
	if r.HTTPHeaders.CFCache != "" {
		b.WriteString("  • cf-cache-status: "+r.HTTPHeaders.CFCache+"\n")
	}
	b.WriteString(fmt.Sprintf("- ⏱ durasi: %.2fs\n", r.DurationSec))
	return b.String()
}

func (s *CheckBugService) ToJSON(results []CheckBugDomainResult) string {
	by, _ := json.MarshalIndent(results, "", "  ")
	return string(by)
}

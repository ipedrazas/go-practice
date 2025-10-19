package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ScanResult struct {
	Port    int    `json:"port"`
	Status  string `json:"status"` // open, closed, filtered
	Service string `json:"service,omitempty"`
	Banner  string `json:"banner,omitempty"`
	Latency int64  `json:"latency_ms"`
	Error   string `json:"error,omitempty"`
}

type ScanSummary struct {
	Target      string        `json:"target"`
	TotalPorts  int           `json:"total_ports"`
	OpenPorts   int           `json:"open_ports"`
	ClosedPorts int           `json:"closed_ports"`
	Filtered    int           `json:"filtered_ports"`
	Duration    time.Duration `json:"duration"`
	Results     []ScanResult  `json:"results"`
}

type OutputFormat string

const (
	TextFormat OutputFormat = "text"
	JSONFormat OutputFormat = "json"
	CSVFormat  OutputFormat = "csv"
)

func main() {
	var (
		target      = flag.String("t", "", "Target host to scan (required)")
		ports       = flag.String("p", "common", "Ports to scan (e.g., '80,443', '1-1000', 'common')")
		concurrency = flag.Int("c", 100, "Number of concurrent connections")
		timeout     = flag.Int("timeout", 1000, "Connection timeout in milliseconds")
		output      = flag.String("o", "text", "Output format (text, json, csv)")
		verbose     = flag.Bool("v", false, "Verbose output")
		help        = flag.Bool("h", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Scan network ports on target hosts.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nPort Examples:\n")
		fmt.Fprintf(os.Stderr, "  -p 80,443,8080     # Specific ports\n")
		fmt.Fprintf(os.Stderr, "  -p 1-1000          # Port range\n")
		fmt.Fprintf(os.Stderr, "  -p common          # Common ports\n")
		fmt.Fprintf(os.Stderr, "  -p 22,80,443,1-1024 # Mixed\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -t example.com -p 80,443\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -t localhost -p 1-1000 -c 50\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -t example.com -p common -o json\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *target == "" {
		fmt.Fprintf(os.Stderr, "Error: Target host is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Parse ports
	portList, err := parsePorts(*ports)
	if err != nil {
		log.Fatalf("Invalid port specification: %v", err)
	}

	// Validate output format
	format := OutputFormat(*output)
	switch format {
	case TextFormat, JSONFormat, CSVFormat:
		// Valid format
	default:
		log.Fatalf("Invalid output format: %s (use text, json, or csv)", *output)
	}

	// Create scanner
	scanner := &PortScanner{
		Target:      *target,
		Ports:       portList,
		Concurrency: *concurrency,
		Timeout:     time.Duration(*timeout) * time.Millisecond,
		Verbose:     *verbose,
	}

	// Run scan
	summary, err := scanner.Scan()
	if err != nil {
		log.Fatalf("Scan failed: %v", err)
	}

	// Output results
	if err := outputResults(summary, format); err != nil {
		log.Fatalf("Failed to output results: %v", err)
	}
}

type PortScanner struct {
	Target      string
	Ports       []int
	Concurrency int
	Timeout     time.Duration
	Verbose     bool
}

func (ps *PortScanner) Scan() (*ScanSummary, error) {
	startTime := time.Now()

	if ps.Verbose {
		fmt.Printf("Starting scan of %s for %d ports\n", ps.Target, len(ps.Ports))
		fmt.Printf("Concurrency: %d, Timeout: %v\n", ps.Concurrency, ps.Timeout)
	}

	// Resolve target hostname
	ipAddr, err := ps.resolveTarget()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target: %w", err)
	}

	// Perform scan
	results := ps.scanPorts(ipAddr)

	// Create summary
	summary := &ScanSummary{
		Target:   ps.Target,
		Results:  results,
		Duration: time.Since(startTime),
	}

	// Calculate statistics
	for _, result := range results {
		summary.TotalPorts++
		switch result.Status {
		case "open":
			summary.OpenPorts++
		case "closed":
			summary.ClosedPorts++
		case "filtered":
			summary.Filtered++
		}
	}

	if ps.Verbose {
		fmt.Printf("Scan completed in %v\n", summary.Duration)
	}

	return summary, nil
}

func (ps *PortScanner) resolveTarget() (string, error) {
	// Check if it's already an IP address
	if net.ParseIP(ps.Target) != nil {
		return ps.Target, nil
	}

	// Resolve hostname
	ips, err := net.LookupIP(ps.Target)
	if err != nil {
		return "", err
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no IP addresses found for %s", ps.Target)
	}

	// Use the first IPv4 address found
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip.String(), nil
		}
	}

	// Fallback to first IP if no IPv4 found
	return ips[0].String(), nil
}

func (ps *PortScanner) scanPorts(ipAddr string) []ScanResult {
	// Create channels
	jobs := make(chan int, len(ps.Ports))
	results := make(chan ScanResult, len(ps.Ports))

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < ps.Concurrency; i++ {
		wg.Add(1)
		go ps.worker(ipAddr, jobs, results, &wg)
	}

	// Send jobs
	go func() {
		for _, port := range ps.Ports {
			jobs <- port
		}
		close(jobs)
	}()

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var scanResults []ScanResult
	for result := range results {
		scanResults = append(scanResults, result)
	}

	// Sort results by port number
	sort.Slice(scanResults, func(i, j int) bool {
		return scanResults[i].Port < scanResults[j].Port
	})

	return scanResults
}

func (ps *PortScanner) worker(ipAddr string, jobs <-chan int, results chan<- ScanResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for port := range jobs {
		result := ps.scanPort(ipAddr, port)
		results <- result

		if ps.Verbose {
			status := "❌"
			if result.Status == "open" {
				status = "✅"
			}
			fmt.Printf("%s Port %d: %s (%v)\n", status, port, result.Status, time.Duration(result.Latency)*time.Millisecond)
		}
	}
}

func (ps *PortScanner) scanPort(ipAddr string, port int) ScanResult {
	startTime := time.Now()
	result := ScanResult{
		Port:   port,
		Status: "filtered", // Default status
	}

	address := fmt.Sprintf("%s:%d", ipAddr, port)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ps.Timeout)
	defer cancel()

	// Attempt TCP connection
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", address)

	latency := time.Since(startTime)
	result.Latency = latency.Milliseconds()

	if err != nil {
		// Determine error type
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				result.Status = "filtered"
			} else if strings.Contains(err.Error(), "connection refused") {
				result.Status = "closed"
			} else {
				result.Status = "filtered"
			}
		} else {
			result.Status = "filtered"
		}
		result.Error = err.Error()
		return result
	}

	// Port is open
	result.Status = "open"
	result.Service = getServiceName(port)

	// Try to grab banner
	banner := ps.grabBanner(conn, port)
	if banner != "" {
		result.Banner = banner
	}

	conn.Close()
	return result
}

func (ps *PortScanner) grabBanner(conn net.Conn, port int) string {
	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	// For certain services, send a probe
	switch port {
	case 21: // FTP
		conn.Write([]byte("HELP\r\n"))
	case 22: // SSH
		// SSH servers usually send banner immediately
	case 25: // SMTP
		conn.Write([]byte("EHLO test\r\n"))
	case 80, 8080: // HTTP
		conn.Write([]byte("GET / HTTP/1.0\r\nHost: test\r\n\r\n"))
	case 110: // POP3
		conn.Write([]byte("USER test\r\n"))
	case 143: // IMAP
		conn.Write([]byte("A001 CAPABILITY\r\n"))
	}

	// Read response
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return ""
	}

	banner := string(buffer[:n])
	// Clean up banner
	banner = strings.TrimSpace(banner)
	banner = strings.ReplaceAll(banner, "\r", "\\r")
	banner = strings.ReplaceAll(banner, "\n", "\\n")

	// Limit banner length
	if len(banner) > 100 {
		banner = banner[:100] + "..."
	}

	return banner
}

func parsePorts(portSpec string) ([]int, error) {
	var ports []int

	// Handle "common" keyword
	if portSpec == "common" {
		return getCommonPorts(), nil
	}

	// Split by comma
	parts := strings.Split(portSpec, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Handle ranges (e.g., "1-1000")
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid port range: %s", part)
			}

			start, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid start port: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid end port: %s", rangeParts[1])
			}

			if start < 1 || end > 65535 || start > end {
				return nil, fmt.Errorf("invalid port range: %d-%d", start, end)
			}

			for i := start; i <= end; i++ {
				ports = append(ports, i)
			}
		} else {
			// Single port
			port, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", part)
			}

			if port < 1 || port > 65535 {
				return nil, fmt.Errorf("port out of range: %d", port)
			}

			ports = append(ports, port)
		}
	}

	// Remove duplicates and sort
	return uniqueSortedPorts(ports), nil
}

func getCommonPorts() []int {
	return []int{
		21, 22, 23, 25, 53, 80, 110, 143, 443, 993, 995,
		135, 139, 445, 993, 995, 1723, 3306, 3389, 5432, 5900,
		8080, 8443, 9200, 27017,
	}
}

func uniqueSortedPorts(ports []int) []int {
	unique := make(map[int]bool)
	for _, port := range ports {
		unique[port] = true
	}

	var result []int
	for port := range unique {
		result = append(result, port)
	}

	sort.Ints(result)
	return result
}

func getServiceName(port int) string {
	services := map[int]string{
		21:   "ftp",
		22:   "ssh",
		23:   "telnet",
		25:   "smtp",
		53:   "dns",
		80:   "http",
		110:  "pop3",
		143:  "imap",
		443:  "https",
		993:  "imaps",
		995:  "pop3s",
		3306: "mysql",
		3389: "rdp",
		5432: "postgresql",
		5900: "vnc",
		8080: "http-alt",
		8443: "https-alt",
	}

	if service, exists := services[port]; exists {
		return service
	}
	return "unknown"
}

func outputResults(summary *ScanSummary, format OutputFormat) error {
	switch format {
	case TextFormat:
		return outputTextResults(summary)
	case JSONFormat:
		return outputJSONResults(summary)
	case CSVFormat:
		return outputCSVResults(summary)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func outputTextResults(summary *ScanSummary) error {
	fmt.Printf("Port Scan Results\n")
	fmt.Printf("================\n\n")
	fmt.Printf("Target: %s\n", summary.Target)
	fmt.Printf("Duration: %v\n", summary.Duration)
	fmt.Printf("Total Ports: %d\n", summary.TotalPorts)
	fmt.Printf("Open Ports: %d\n", summary.OpenPorts)
	fmt.Printf("Closed Ports: %d\n", summary.ClosedPorts)
	fmt.Printf("Filtered: %d\n", summary.Filtered)
	fmt.Printf("\n")

	// Show open ports first
	fmt.Printf("Open Ports:\n")
	openFound := false
	for _, result := range summary.Results {
		if result.Status == "open" {
			openFound = true
			line := fmt.Sprintf("  %d/tcp %s", result.Port, result.Service)
			if result.Banner != "" {
				line += fmt.Sprintf(" (%s)", result.Banner)
			}
			fmt.Println(line)
		}
	}
	if !openFound {
		fmt.Println("  No open ports found")
	}

	fmt.Printf("\n")

	// Show closed ports if verbose
	// This could be expanded based on user preference

	return nil
}

func outputJSONResults(summary *ScanSummary) error {
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func outputCSVResults(summary *ScanSummary) error {
	fmt.Println("port,status,service,banner,latency_ms,error")

	for _, result := range summary.Results {
		fmt.Printf("%d,%s,%s,%s,%d,%s\n",
			result.Port,
			result.Status,
			result.Service,
			result.Banner,
			result.Latency,
			result.Error)
	}

	return nil
}

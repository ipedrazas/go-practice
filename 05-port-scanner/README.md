# Exercise 5: Port Scanner Utility

## 🎯 Objective
Create a network port scanner that checks for open ports on target hosts using TCP connections.

## 📋 Main Focus Areas
- **Network programming** (`net` package)
- **Concurrency** (goroutines and channels)
- **Timeout handling** (`context` and `time` packages)
- **Parallel processing**
- **Command-line argument parsing**

## 🔧 What You'll Build
A port scanner that:
- Scans individual ports or port ranges
- Supports multiple concurrent connections
- Provides detailed scan results
- Handles timeouts gracefully
- Supports common port lists
- Generates reports in different formats

## 📝 Instructions

### Step 1: Basic TCP Scanner
Create a tool that:
1. Takes a target host and port as arguments
2. Attempts to establish TCP connection
3. Reports if port is open, closed, or filtered
4. Handles connection timeouts

### Step 2: Command-line Interface
Add support for these flags:
- `-t, --target`: Target host to scan
- `-p, --ports`: Ports to scan (e.g., "80,443", "1-1000", "common")
- `-c, --concurrency`: Number of concurrent connections
- `-t, --timeout`: Connection timeout in milliseconds
- `-o, --output`: Output format (text, json, csv)

### Step 3: Advanced Scanning
Implement these features:
- Port range scanning
- Common ports list (80, 443, 22, 21, 23, 25, 53, 110, 995, 993, 143, etc.)
- Service detection (basic banner grabbing)
- Ping check before scanning
- Scan results aggregation

### Step 4: Performance & Features
Add these capabilities:
- Configurable concurrency limits
- Progress indicators
- Rate limiting to avoid triggering security measures
- Results filtering (only open ports)
- Host resolution (hostname to IP)

## 🚀 Getting Started

```bash
# Navigate to exercise directory
cd 05-port-scanner

# Create your main.go file
touch main.go

# Run your solution
go run main.go -t example.com -p 80,443,8080

# Scan port range
go run main.go -t localhost -p 1-1000

# Scan common ports
go run main.go -t example.com -p common
```

## 💡 Implementation Tips

### Basic TCP Connection
```go
func scanPort(target string, port int, timeout time.Duration) (string, error) {
    address := fmt.Sprintf("%s:%d", target, port)

    conn, err := net.DialTimeout("tcp", address, timeout)
    if err != nil {
        return "closed", err
    }
    conn.Close()
    return "open", nil
}
```

### Concurrent Scanning with Worker Pool
```go
func scanPorts(target string, ports []int, concurrency int, timeout time.Duration) []ScanResult {
    jobs := make(chan int, len(ports))
    results := make(chan ScanResult, len(ports))

    // Start workers
    for i := 0; i < concurrency; i++ {
        go worker(target, jobs, results, timeout)
    }

    // Send jobs
    for _, port := range ports {
        jobs <- port
    }
    close(jobs)

    // Collect results
    var scanResults []ScanResult
    for i := 0; i < len(ports); i++ {
        result := <-results
        scanResults = append(scanResults, result)
    }

    return scanResults
}
```

### Common Ports List
```go
var commonPorts = []int{
    21, 22, 23, 25, 53, 80, 110, 143, 443, 993, 995,
    3306, 3389, 5432, 5900, 8080, 8443,
}
```

## 🧪 Test Cases
Test with different scenarios:
- Localhost scanning
- Public websites
- Invalid hostnames
- Port ranges vs individual ports
- Different concurrency levels
- Timeout behavior

## 📚 Go Concepts Covered
- `net` package for network operations
- Goroutines for concurrent execution
- Channels for communication
- Worker pool pattern
- Context and cancellation
- Timeouts and error handling
- Struct composition for results

## 🔗 Useful Resources
- [Go net package](https://pkg.go.dev/net)
- [Go concurrency patterns](https://blog.golang.org/concurrency-patterns)
- [TCP port scanning basics](https://nmap.org/book/man-port-scanning-basics.html)

## ✅ Success Criteria
Your tool should be able to:
- [ ] Scan individual ports and port ranges
- [ ] Handle concurrent connections efficiently
- [ ] Provide clear scan results
- [ ] Handle network errors and timeouts
- [ ] Support different output formats
- [ ] Be configurable for different use cases

## 🎁 Bonus Features
If you want an extra challenge:
- UDP scanning
- Service version detection
- Output to files
- IPv6 support
- Scan multiple hosts
- Integration with Shodan or other APIs
- GUI interface

When you're ready, check out the [solution](./solution/main.go) to see a complete implementation.
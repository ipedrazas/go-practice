# Exercise 3: Log Analyzer Tool

## 🎯 Objective
Create a CLI tool that parses and analyzes log files to extract useful information, patterns, and statistics.

## 📋 Main Focus Areas
- **File reading and processing** (`bufio`, `io` packages)
- **Regular expressions** (`regexp` package)
- **Text parsing and manipulation** (`strings` package)
- **Date/time handling** (`time` package)
- **Data aggregation and statistics**

## 🔧 What You'll Build
A log analyzer that:
- Parses various log formats (Apache, Nginx, custom formats)
- Extracts key metrics (requests per hour, error rates, popular pages)
- Identifies patterns and anomalies
- Generates summary reports
- Supports filtering and searching
- Outputs results in different formats

## 📝 Instructions

### Step 1: Basic Log Parsing
Create a tool that:
1. Reads log files line by line
2. Parses common log format (Apache/Nginx)
3. Extracts basic information (IP, timestamp, method, URL, status)
4. Counts total lines and requests

### Step 2: Command-line Interface
Add support for these flags:
- `-f, --file`: Log file to analyze
- `-p, --pattern`: Custom regex pattern for parsing
- `-s, --start`: Start time for filtering
- `-e, --end`: End time for filtering
- `-o, --output`: Output format (text, json, csv)
- `-t, --top`: Number of top results to show

### Step 3: Analysis Features
Implement these analyses:
- **Request statistics**: Total requests, requests per hour/day
- **Status codes**: Distribution of HTTP status codes
- **Top IPs**: Most frequent client IPs
- **Top pages**: Most requested URLs
- **Error analysis**: 4xx and 5xx error details
- **User agents**: Most common user agents

### Step 4: Advanced Features
Add these capabilities:
- Real-time log monitoring (tail -f functionality)
- Geographic IP analysis
- Bandwidth usage calculation
- Pattern detection (attack patterns, bots)
- Alerting for anomalies
- Multiple log file processing

## 🚀 Getting Started

```bash
# Navigate to exercise directory
cd 03-log-analyzer

# Create a sample log file
cat > sample.log << EOF
192.168.1.1 - - [10/Oct/2023:13:55:36 +0000] "GET /index.html HTTP/1.1" 200 2326
192.168.1.2 - - [10/Oct/2023:13:55:37 +0000] "GET /about.html HTTP/1.1" 200 1512
192.168.1.3 - - [10/Oct/2023:13:55:38 +0000] "GET /nonexistent.html HTTP/1.1" 404 209
EOF

# Create your main.go file
touch main.go

# Run your solution
go run main.go -f sample.log
```

## 💡 Implementation Tips

### Reading Files Efficiently
```go
file, err := os.Open(filename)
if err != nil {
    log.Fatal(err)
}
defer file.Close()

scanner := bufio.NewScanner(file)
for scanner.Scan() {
    line := scanner.Text()
    // Process line
}
```

### Regular Expression for Common Log Format
```go
logPattern := regexp.MustCompile(`^(\S+) \S+ \S+ \[([\w:/]+\s[+\-]\d{4})\] "(\S+) (\S+) (\S+)" (\d{3}) (\d+)`)

matches := logPattern.FindStringSubmatch(line)
if len(matches) > 0 {
    ip := matches[1]
    timestamp := matches[2]
    method := matches[3]
    url := matches[4]
    status := matches[6]
    size := matches[7]
}
```

### Time Parsing
```go
layout := "02/Jan/2006:15:04:05 -0700"
timestamp, err := time.Parse(layout, timeString)
```

### Data Aggregation
```go
type Stats struct {
    TotalRequests   int
    StatusCodes     map[int]int
    TopIPs          map[string]int
    RequestsPerHour map[string]int
}

stats := &Stats{
    StatusCodes:     make(map[int]int),
    TopIPs:          make(map[string]int),
    RequestsPerHour: make(map[string]int),
}
```

## 🧪 Test Cases
Create different log scenarios:
- Normal web server traffic
- High error rate scenarios
- Different log formats
- Large log files (performance testing)
- Corrupted or malformed log entries

## 📚 Go Concepts Covered
- `bufio.Scanner` for efficient file reading
- `regexp` package for pattern matching
- `time` package for date/time operations
- Maps for data aggregation
- String manipulation and parsing
- Structs for organizing data
- Error handling and validation

## 🔗 Useful Resources
- [Go regexp package](https://pkg.go.dev/regexp)
- [Go bufio package](https://pkg.go.dev/bufio)
- [Go time package](https://pkg.go.dev/time)
- [Common Log Format specification](https://httpd.apache.org/docs/1.3/logs.html#common)

## ✅ Success Criteria
Your tool should be able to:
- [ ] Parse common log formats correctly
- [ ] Generate useful statistics and summaries
- [ ] Support time-based filtering
- [ ] Handle large log files efficiently
- [ ] Provide multiple output formats
- [ ] Detect and report parsing errors

## 🎁 Bonus Features
If you want an extra challenge:
- Support for custom log formats
- Real-time monitoring with WebSocket updates
- Machine learning for anomaly detection
- Integration with Grafana/Prometheus
- Web dashboard for log visualization
- Log file compression support

When you're ready, check out the [solution](./solution/main.go) to see a complete implementation.
# jclog - JSON Log Viewer

A powerful command-line tool for viewing and formatting JSON log files. Supports multiple log formats, provides colorized output, filtering capabilities, and configuration file management.

[![Go Report Card](https://goreportcard.com/badge/github.com/techarm/jclog)](https://goreportcard.com/report/github.com/techarm/jclog)
[![License](https://img.shields.io/github/license/techarm/jclog)](https://github.com/techarm/jclog/blob/main/LICENSE)

## Features

- 🎨 Colorized output with different colors for log levels
- 🔍 Custom field filtering and exclusion
- 📝 Flexible output format configuration
- 🔧 Configuration file management with multiple profiles
- 🌲 Nested JSON parsing support
- 📦 Support for major logging frameworks

## Installation

```bash
go install github.com/techarm/jclog@latest
```

## Quick Start

1. Basic Usage:

```bash
# Read from file
jclog app.log

# Read from pipe
tail -f app.log | jclog
```

2. Using Filters:

```bash
# Show only INFO level logs
jclog --filter level=INFO app.log

# Exclude DEBUG level logs
jclog --exclude level=DEBUG app.log
```

3. Custom Output Format:

```bash
# Custom format string
jclog --format "{timestamp} [{level}] {message}" app.log

# Specify fields to display
jclog --fields timestamp,level,message,user app.log
```

## Configuration Management

1. Initialize Configuration:

```bash
jclog config init
```

2. Add New Profile:

```bash
jclog config add-profile --name prod \
    --format "{timestamp} [{level}] {message}" \
    --fields timestamp,level,message \
    --filter level=INFO
```

3. Use Profile:

```bash
jclog --profile prod app.log
```

## Supported Log Frameworks

### Logrus
```json
{"level":"info","msg":"Server is starting","time":"2024-03-20T10:00:00Z","service":"api"}
```

### Zap
```json
{"level":"INFO","ts":1647763200,"msg":"Connected to database","logger":"db","db_name":"users"}
```

### Zerolog
```json
{"level":"info","time":"2024-03-20T10:00:00Z","message":"Cache initialized","cache_size":1000}
```

### Bunyan
```json
{"name":"myapp","hostname":"server1","pid":12345,"level":30,"msg":"Request processed","time":"2024-03-20T10:00:00Z","v":0}
```

## Output Examples

Default Configuration:
```
2024-03-20T10:00:00Z [INFO] Server is starting
2024-03-20T10:00:00Z [ERROR] Failed to connect to database
2024-03-20T10:00:00Z [DEBUG] Cache hit ratio: 0.95
```

Custom Format:
```
[2024-03-20 10:00:00] INFO  - Server is starting (service: api)
[2024-03-20 10:00:00] ERROR - Failed to connect to database (error: timeout)
[2024-03-20 10:00:00] DEBUG - Cache hit ratio: 0.95 (cache: users)
```

## Advanced Features

1. Nested JSON Parsing:
   - Automatically parses nested JSON in message fields
   - Configurable parsing depth with `--max-depth`
   - Flattens nested structures for easy viewing

2. Field Aliases:
   - Supports common field aliases (e.g., msg/message, time/timestamp)
   - Automatically recognizes different log format conventions

3. Color Schemes:
   - INFO: Green
   - WARN: Yellow
   - ERROR: Red
   - DEBUG: Gray
   - TRACE: White

4. Pipeline Support:
   - Works seamlessly with Unix pipes
   - Real-time log processing with `tail -f`
   - Compatible with grep, awk, and other Unix tools

## Command Line Options

```bash
jclog [options] [file]

Options:
  --config string       Path to config file (default: ~/.jclog.json)
  --profile string     Configuration profile to use
  --format string      Output format template
  --fields strings     Fields to display
  --max-depth int      Maximum JSON parsing depth (default: 2)
  --hide-missing       Hide missing fields in format
  --filter strings     Filter conditions (field=value)
  --exclude strings    Exclude conditions (field=value)
```

## Common Use Cases

### 1. Monitoring Application Logs in Real-time

```bash
# Monitor application logs with custom format
tail -f /var/log/app.log | jclog --format "[{timestamp}] {level} - {message} (service: {service})"

# Monitor multiple log files
tail -f /var/log/app1.log /var/log/app2.log | jclog --filter service=api

# Watch for error logs only
tail -f /var/log/app.log | jclog --filter level=ERROR --fields timestamp,message,error
```

### 2. Log Analysis and Debugging

```bash
# Find all errors from a specific service
jclog --filter "level=ERROR" --filter "service=payment" logs/app.log

# Analyze slow requests
jclog --filter "duration_ms>1000" --fields timestamp,path,duration_ms,user_id logs/access.log

# Track user activity
jclog --filter "user_id=12345" --format "{timestamp} {action} by {user_id}" logs/audit.log
```

### 3. System Monitoring

```bash
# Monitor system metrics
jclog --fields timestamp,cpu_usage,memory_usage,disk_usage logs/metrics.log

# Alert on high resource usage
jclog --filter "cpu_usage>80" --filter "memory_usage>90" logs/metrics.log

# Track service health
jclog --fields service,status,health_check_latency logs/health.log
```

### 4. Security Audit

```bash
# Monitor failed login attempts
jclog --filter "event=login_failed" --fields timestamp,ip_address,username logs/auth.log

# Track permission changes
jclog --filter "action=permission_change" --fields timestamp,user,resource,old_perm,new_perm logs/audit.log
```

## Configuration Examples

### 1. Development Profile
```json
{
  "name": "dev",
  "format": "[{timestamp}] {level} {message}",
  "fields": ["timestamp", "level", "message", "caller"],
  "maxDepth": 3,
  "hideMissing": false,
  "filters": [],
  "excludes": []
}
```

### 2. Production Profile
```json
{
  "name": "prod",
  "format": "{timestamp} [{level}] {message} (service={service})",
  "fields": ["timestamp", "level", "message", "service", "request_id"],
  "maxDepth": 2,
  "hideMissing": true,
  "filters": ["level=ERROR", "level=WARN"],
  "excludes": ["level=DEBUG"]
}
```

### 3. Audit Profile
```json
{
  "name": "audit",
  "format": "{timestamp} - User:{user} Action:{action} Resource:{resource}",
  "fields": ["timestamp", "user", "action", "resource", "ip_address"],
  "maxDepth": 1,
  "hideMissing": true,
  "filters": ["type=audit"],
  "excludes": []
}
```

### 4. Metrics Profile
```json
{
  "name": "metrics",
  "format": "{timestamp} {service} - CPU:{cpu_usage}% MEM:{memory_usage}% DISK:{disk_usage}%",
  "fields": ["timestamp", "service", "cpu_usage", "memory_usage", "disk_usage"],
  "maxDepth": 1,
  "hideMissing": false,
  "filters": [],
  "excludes": []
}
```

### Complete Configuration File Example
```json
{
  "activeProfile": "prod",
  "profiles": {
    "dev": {
      "format": "[{timestamp}] {level} {message}",
      "fields": ["timestamp", "level", "message", "caller"],
      "maxDepth": 3,
      "hideMissing": false,
      "filters": [],
      "excludes": []
    },
    "prod": {
      "format": "{timestamp} [{level}] {message} (service={service})",
      "fields": ["timestamp", "level", "message", "service", "request_id"],
      "maxDepth": 2,
      "hideMissing": true,
      "filters": ["level=ERROR", "level=WARN"],
      "excludes": ["level=DEBUG"]
    },
    "audit": {
      "format": "{timestamp} - User:{user} Action:{action} Resource:{resource}",
      "fields": ["timestamp", "user", "action", "resource", "ip_address"],
      "maxDepth": 1,
      "hideMissing": true,
      "filters": ["type=audit"],
      "excludes": []
    },
    "metrics": {
      "format": "{timestamp} {service} - CPU:{cpu_usage}% MEM:{memory_usage}% DISK:{disk_usage}%",
      "fields": ["timestamp", "service", "cpu_usage", "memory_usage", "disk_usage"],
      "maxDepth": 1,
      "hideMissing": false,
      "filters": [],
      "excludes": []
    }
  }
}
```

## Contributing

Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change.

## License

MIT License

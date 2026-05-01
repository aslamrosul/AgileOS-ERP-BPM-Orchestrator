# AgileOS Network Security & DDoS Protection Guide

## 🛡️ Overview

AgileOS implements **enterprise-grade network security** with multiple layers of protection:
- **Rate Limiting** - Prevent API abuse and brute force attacks
- **IP Filtering** - Blacklist/whitelist management
- **Security Headers** - Industry-standard HTTP security headers
- **DDoS Protection** - Request throttling and traffic management
- **Audit Logging** - Security event tracking

## 🚦 Rate Limiting

### Global Rate Limit
**100 requests per minute per IP** for all endpoints

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1234567890
```

### Authentication Rate Limit
**5 login attempts per minute per IP** (Brute Force Protection)

Applies to:
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/register`

### Custom Rate Limits

You can create custom rate limiters for specific endpoints:

```go
// Example: 10 requests per 5 minutes for workflow creation
workflowLimiter := middleware.CustomRateLimit(middleware.RateLimitConfig{
    Period:  5 * time.Minute,
    Limit:   10,
    Message: "Workflow creation limit exceeded",
})

protected.POST("/workflow", workflowLimiter, workflowHandler.CreateWorkflow)
```

### Rate Limit Response

When limit is exceeded, API returns `429 Too Many Requests`:

```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests. Please try again later.",
  "retry_after": 1234567890
}
```

## 🚫 IP Filtering

### Blacklist Management

#### Automatic Blacklisting
IPs are automatically blacklisted when:
- Exceeding auth rate limit 3 times in 10 minutes
- Detected suspicious patterns
- Multiple failed authentication attempts

#### Manual Blacklisting

```go
import "agileos-backend/middleware"

// Blacklist IP for 24 hours
ipFilter := middleware.GetIPFilter()
ipFilter.BlacklistIP("192.168.1.100", 24 * time.Hour)
```

#### Blacklist Duration
- **Temporary**: 1-24 hours (automatic expiry)
- **Permanent**: Add to database for persistent blocking

### Whitelist Management

#### Default Whitelisted IPs
- `127.0.0.1` (localhost)
- `::1` (IPv6 localhost)

#### Add to Whitelist

```go
ipFilter := middleware.GetIPFilter()
ipFilter.WhitelistIP("203.0.113.10")  // Your office IP
```

### IP Filter Behavior

1. **Whitelisted IPs**: Always allowed, bypass all rate limits
2. **Blacklisted IPs**: Blocked with `403 Forbidden`
3. **Unknown IPs**: Subject to rate limiting

## 🔒 Security Headers

### Implemented Headers

| Header | Value | Purpose |
|--------|-------|---------|
| `X-Content-Type-Options` | `nosniff` | Prevent MIME type sniffing |
| `X-Frame-Options` | `DENY` | Prevent clickjacking |
| `X-XSS-Protection` | `1; mode=block` | Enable XSS filter |
| `Content-Security-Policy` | (see below) | Restrict resource loading |
| `Strict-Transport-Security` | `max-age=31536000` | Force HTTPS (production) |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Control referrer info |
| `Permissions-Policy` | `geolocation=(), microphone=(), camera=()` | Disable unnecessary features |

### Content Security Policy (CSP)

```
default-src 'self';
script-src 'self' 'unsafe-inline' 'unsafe-eval';
style-src 'self' 'unsafe-inline';
img-src 'self' data: https:;
font-src 'self' data:;
connect-src 'self' ws: wss:;
frame-ancestors 'none';
```

**What this does:**
- Only load scripts/styles from same origin
- Allow WebSocket connections
- Block embedding in iframes
- Prevent XSS attacks

## 📊 Security Audit Logging

### Logged Security Events

All security events are logged to audit trail:

```go
logger.LogSecurity("event_type", user_id, ip, metadata)
```

**Event Types:**
- `rate_limit_exceeded` - Global rate limit hit
- `auth_rate_limit_exceeded` - Login rate limit hit (HIGH severity)
- `ip_blacklisted` - IP added to blacklist
- `blocked_blacklisted_ip` - Request from blacklisted IP
- `untrusted_proxy_request` - Request from untrusted proxy

### Example Log Entry

```json
{
  "timestamp": "2026-04-30T10:15:30Z",
  "level": "WARN",
  "type": "security",
  "action": "auth_rate_limit_exceeded",
  "ip": "203.0.113.50",
  "metadata": {
    "limit": 5,
    "endpoint": "/api/v1/auth/login",
    "method": "POST",
    "user_agent": "Mozilla/5.0...",
    "severity": "HIGH"
  }
}
```

### Query Security Logs

```surql
-- Get all rate limit violations in last 24 hours
SELECT * FROM audit_trails
WHERE action = 'rate_limit_exceeded'
  AND timestamp > time::now() - 24h
ORDER BY timestamp DESC;

-- Get IPs with multiple failed login attempts
SELECT 
    metadata.ip AS ip,
    count() AS attempts
FROM audit_trails
WHERE action = 'auth_rate_limit_exceeded'
  AND timestamp > time::now() - 1h
GROUP BY metadata.ip
HAVING attempts > 3;
```

## 🌐 CORS Configuration

### Allowed Origins (Production)

```go
allowedOrigins := map[string]bool{
    "https://agileos.com": true,
    "https://www.agileos.com": true,
    "https://your-app.azurewebsites.net": true,
}
```

### Development Mode
In development, CORS allows all origins (`*`)

### Update Allowed Origins

Edit `middleware/security_headers.go`:

```go
allowedOrigins := map[string]bool{
    "https://your-domain.com": true,
}
```

## ☁️ Azure Network Security

### Network Security Group (NSG) Rules

#### Inbound Rules

| Priority | Name | Port | Protocol | Source | Action |
|----------|------|------|----------|--------|--------|
| 100 | Allow-HTTPS | 443 | TCP | Internet | Allow |
| 110 | Allow-HTTP | 80 | TCP | Internet | Allow |
| 120 | Allow-SSH | 22 | TCP | Your-IP | Allow |
| 900 | Deny-All | * | * | * | Deny |

#### Outbound Rules

| Priority | Name | Port | Protocol | Destination | Action |
|----------|------|------|-------------|--------|--------|
| 100 | Allow-Internet | * | * | Internet | Allow |

### Azure CLI Commands

```bash
# Create NSG
az network nsg create \
  --resource-group agileos-rg \
  --name agileos-nsg \
  --location eastus

# Allow HTTPS
az network nsg rule create \
  --resource-group agileos-rg \
  --nsg-name agileos-nsg \
  --name Allow-HTTPS \
  --priority 100 \
  --source-address-prefixes Internet \
  --destination-port-ranges 443 \
  --protocol Tcp \
  --access Allow

# Allow HTTP
az network nsg rule create \
  --resource-group agileos-rg \
  --nsg-name agileos-nsg \
  --name Allow-HTTP \
  --priority 110 \
  --source-address-prefixes Internet \
  --destination-port-ranges 80 \
  --protocol Tcp \
  --access Allow

# Allow SSH (from your IP only)
az network nsg rule create \
  --resource-group agileos-rg \
  --nsg-name agileos-nsg \
  --name Allow-SSH \
  --priority 120 \
  --source-address-prefixes YOUR_IP_ADDRESS \
  --destination-port-ranges 22 \
  --protocol Tcp \
  --access Allow

# Deny all other inbound traffic
az network nsg rule create \
  --resource-group agileos-rg \
  --nsg-name agileos-nsg \
  --name Deny-All \
  --priority 900 \
  --source-address-prefixes '*' \
  --destination-port-ranges '*' \
  --protocol '*' \
  --access Deny
```

### Associate NSG with Subnet

```bash
az network vnet subnet update \
  --resource-group agileos-rg \
  --vnet-name agileos-vnet \
  --name default \
  --network-security-group agileos-nsg
```

## 🔐 Database Security

### SurrealDB Network Isolation

#### Docker Compose Configuration

```yaml
services:
  agileos-db:
    image: surrealdb/surrealdb:v1.4.2
    networks:
      - agileos-internal  # Internal network only
    ports:
      - "127.0.0.1:8002:8000"  # Bind to localhost only
    # DO NOT expose to 0.0.0.0 in production!
```

#### Azure Container Instances

```bash
# Create internal network
az network vnet create \
  --resource-group agileos-rg \
  --name agileos-vnet \
  --address-prefix 10.0.0.0/16 \
  --subnet-name backend-subnet \
  --subnet-prefix 10.0.1.0/24

# Deploy SurrealDB (no public IP)
az container create \
  --resource-group agileos-rg \
  --name agileos-db \
  --image surrealdb/surrealdb:v1.4.2 \
  --vnet agileos-vnet \
  --subnet backend-subnet \
  --ip-address Private  # No public IP!
```

### Database Connection Security

```go
// Use internal Docker network hostname
dbURL := "ws://agileos-db:8000/rpc"  // NOT localhost!

// Use strong credentials
dbUser := os.Getenv("SURREAL_USER")  // From Azure Key Vault
dbPass := os.Getenv("SURREAL_PASS")  // From Azure Key Vault
```

## 🚨 DDoS Protection

### Layer 7 (Application Layer)

1. **Rate Limiting**: Throttle requests per IP
2. **IP Blacklisting**: Block malicious IPs
3. **Request Validation**: Reject malformed requests
4. **Connection Limits**: Max concurrent connections per IP

### Azure DDoS Protection

#### Enable Azure DDoS Protection Standard

```bash
# Create DDoS protection plan
az network ddos-protection create \
  --resource-group agileos-rg \
  --name agileos-ddos-plan \
  --location eastus

# Associate with VNet
az network vnet update \
  --resource-group agileos-rg \
  --name agileos-vnet \
  --ddos-protection true \
  --ddos-protection-plan agileos-ddos-plan
```

**Features:**
- Always-on traffic monitoring
- Automatic attack mitigation
- Real-time attack metrics
- DDoS rapid response support

### Cloudflare Integration (Optional)

For additional protection, use Cloudflare as CDN/WAF:

1. **DNS**: Point domain to Cloudflare
2. **SSL/TLS**: Full (strict) mode
3. **Firewall Rules**: Block known bad actors
4. **Rate Limiting**: Additional layer
5. **Bot Management**: Block malicious bots

## 📈 Monitoring & Alerts

### Metrics to Monitor

1. **Request Rate**: Requests per second per endpoint
2. **Rate Limit Hits**: Number of 429 responses
3. **Blacklisted IPs**: Active blacklist size
4. **Failed Auth Attempts**: Login failures per IP
5. **Response Times**: Latency under load

### Azure Monitor Alerts

```bash
# Alert on high rate limit violations
az monitor metrics alert create \
  --name high-rate-limit-violations \
  --resource-group agileos-rg \
  --scopes /subscriptions/.../agileos-backend \
  --condition "avg http_requests_total > 1000" \
  --window-size 5m \
  --evaluation-frequency 1m \
  --action email admin@agileos.com
```

### Log Analytics Queries

```kusto
// Rate limit violations by IP
AuditTrails
| where action == "rate_limit_exceeded"
| summarize count() by ip
| order by count_ desc

// Failed login attempts
AuditTrails
| where action == "auth_rate_limit_exceeded"
| where timestamp > ago(1h)
| project timestamp, ip, metadata
```

## 🧪 Testing Security

### Test Rate Limiting

```bash
# Test global rate limit (should get 429 after 100 requests)
for i in {1..110}; do
  curl -i http://localhost:8080/api/v1/workflows
done

# Test auth rate limit (should get 429 after 5 attempts)
for i in {1..7}; do
  curl -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"test","password":"wrong"}'
done
```

### Test Security Headers

```bash
curl -I http://localhost:8080/health

# Should see:
# X-Content-Type-Options: nosniff
# X-Frame-Options: DENY
# X-XSS-Protection: 1; mode=block
# Content-Security-Policy: ...
```

### Test IP Blacklist

```go
// In your test code
ipFilter := middleware.GetIPFilter()
ipFilter.BlacklistIP("192.168.1.100", 1 * time.Hour)

// Try accessing from that IP - should get 403
```

## 📚 Best Practices

### 1. Regular Security Audits
- Review audit logs weekly
- Check for unusual patterns
- Update blacklist/whitelist

### 2. Keep Dependencies Updated
```bash
go get -u golang.org/x/time/rate
go get -u github.com/ulule/limiter/v3
```

### 3. Use Environment Variables
Never hardcode:
- Database credentials
- API keys
- Allowed origins

### 4. Enable HTTPS in Production
```go
if gin.Mode() == gin.ReleaseMode {
    r.RunTLS(":443", "cert.pem", "key.pem")
}
```

### 5. Monitor Security Logs
Set up alerts for:
- Multiple rate limit violations
- Repeated failed logins
- Unusual traffic patterns

## ✅ Security Checklist

- [x] Rate limiting implemented (global + auth)
- [x] IP filtering (blacklist/whitelist)
- [x] Security headers configured
- [x] CORS properly configured
- [x] Audit logging for security events
- [ ] Azure NSG rules configured
- [ ] SurrealDB network isolated
- [ ] HTTPS enabled in production
- [ ] DDoS protection enabled
- [ ] Monitoring alerts configured
- [ ] Regular security audits scheduled

---

**Status**: ✅ Network Security Implementation Complete

**Next Steps**:
1. Configure Azure NSG rules
2. Enable Azure DDoS Protection
3. Set up monitoring alerts
4. Test rate limiting in production
5. Review security logs regularly

**Emergency Contact**: security@agileos.com

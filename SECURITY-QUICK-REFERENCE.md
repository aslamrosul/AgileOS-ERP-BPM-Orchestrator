# AgileOS Security Quick Reference Card

## 🚦 Rate Limits

| Endpoint | Limit | Window | Response |
|----------|-------|--------|----------|
| **All Endpoints** | 100 req | 1 min | 429 Too Many Requests |
| **Login/Register** | 5 req | 1 min | 429 + Security Log |
| **Custom** | Configurable | Configurable | 429 |

## 🔒 Security Headers

```
✓ X-Content-Type-Options: nosniff
✓ X-Frame-Options: DENY
✓ X-XSS-Protection: 1; mode=block
✓ Content-Security-Policy: (strict)
✓ Strict-Transport-Security: max-age=31536000
✓ Referrer-Policy: strict-origin-when-cross-origin
✓ Permissions-Policy: geolocation=(), microphone=(), camera=()
```

## 🚫 IP Management

### Blacklist IP
```go
ipFilter := middleware.GetIPFilter()
ipFilter.BlacklistIP("192.168.1.100", 24 * time.Hour)
```

### Whitelist IP
```go
ipFilter.WhitelistIP("203.0.113.10")
```

### Check Status
```go
isBlocked := ipFilter.IsBlacklisted("192.168.1.100")
isAllowed := ipFilter.IsWhitelisted("203.0.113.10")
```

## 📊 Security Logs

### Query Recent Violations
```surql
SELECT * FROM audit_trails
WHERE action IN ['rate_limit_exceeded', 'auth_rate_limit_exceeded']
  AND timestamp > time::now() - 1h
ORDER BY timestamp DESC;
```

### Top Offending IPs
```surql
SELECT 
    metadata.ip AS ip,
    count() AS violations
FROM audit_trails
WHERE action = 'rate_limit_exceeded'
  AND timestamp > time::now() - 24h
GROUP BY metadata.ip
ORDER BY violations DESC
LIMIT 10;
```

## ☁️ Azure NSG Ports

| Port | Protocol | Purpose | Source |
|------|----------|---------|--------|
| 443 | TCP | HTTPS | Internet |
| 80 | TCP | HTTP | Internet |
| 22 | TCP | SSH | Your IP Only |
| * | * | All Others | **DENY** |

## 🔐 Database Security

```yaml
# docker-compose.yml
services:
  agileos-db:
    ports:
      - "127.0.0.1:8002:8000"  # ✓ Localhost only
      # - "0.0.0.0:8002:8000"  # ✗ NEVER in production!
    networks:
      - agileos-internal  # ✓ Internal network
```

## 🧪 Quick Tests

### Test Rate Limit
```bash
# Should get 429 after 100 requests
for i in {1..110}; do curl http://localhost:8080/health; done
```

### Test Auth Rate Limit
```bash
# Should get 429 after 5 attempts
for i in {1..7}; do
  curl -X POST http://localhost:8080/api/v1/auth/login \
    -d '{"username":"test","password":"wrong"}'
done
```

### Check Security Headers
```bash
curl -I http://localhost:8080/health | grep -E "X-|Content-Security"
```

## 🚨 Emergency Response

### Block Attacking IP
```bash
# SSH to server
ssh admin@your-server

# Add to blacklist (requires code deployment)
# Or use Azure NSG:
az network nsg rule create \
  --resource-group agileos-rg \
  --nsg-name agileos-nsg \
  --name Block-Attacker \
  --priority 90 \
  --source-address-prefixes 203.0.113.50 \
  --destination-port-ranges '*' \
  --access Deny
```

### Check Active Attacks
```bash
# View recent security logs
docker logs agileos-backend | grep "rate_limit_exceeded"

# Count by IP
docker logs agileos-backend | grep "rate_limit_exceeded" | \
  grep -oP 'ip=[^ ]+' | sort | uniq -c | sort -rn
```

## 📞 Contacts

- **Security Team**: security@agileos.com
- **On-Call**: +62-xxx-xxx-xxxx
- **Azure Support**: portal.azure.com

---

**Last Updated**: 2026-04-30
**Version**: 1.0.0

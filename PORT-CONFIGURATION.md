# AgileOS BPM - Port Configuration Guide

## Port Mapping (Updated to Avoid Conflicts)

### External Ports (Host Machine)

| Service | Default Port | AgileOS Port | Purpose |
|---------|-------------|--------------|---------|
| **Nginx HTTP** | 80 | **8090** | Main entry point (HTTP) |
| **Nginx HTTPS** | 443 | **8443** | Main entry point (HTTPS) |
| **Frontend** | 3000 | **3001** | Next.js application |
| **Backend** | 8080 | **8081** | Go API server |
| **SurrealDB** | 8000 | **8002** | Database web interface |
| **NATS Client** | 4222 | **4223** | Message broker client |
| **NATS Monitor** | 8222 | **8223** | NATS monitoring |

### Internal Ports (Docker Network)

All services communicate internally using their default ports:
- Backend → SurrealDB: `ws://agileos-db:8000/rpc`
- Backend → NATS: `nats://agileos-nats:4222`
- Nginx → Backend: `http://agileos-backend:8080`
- Nginx → Frontend: `http://agileos-frontend:3000`

---

## Access URLs

### Development Mode

```bash
# Main application (via Nginx)
http://localhost:8090

# Direct access (for debugging)
Frontend:  http://localhost:3001
Backend:   http://localhost:8081
Database:  http://localhost:8002
NATS:      http://localhost:8223
```

### Production Mode (Azure)

```bash
# Via Nginx reverse proxy
http://YOUR_VM_IP:8090
https://YOUR_VM_IP:8443

# Or with custom domain
http://your-domain.com:8090
https://your-domain.com:8443
```

---

## Why These Ports?

### Common Port Conflicts

Many projects use these default ports:
- **80/443**: Web servers (Apache, Nginx, IIS)
- **3000**: React, Next.js, Node.js apps
- **8080**: Tomcat, Jenkins, Spring Boot
- **8000**: Django, Python apps
- **4222**: Other NATS instances

### AgileOS Solution

We shifted all ports slightly to avoid conflicts:
- **8090/8443**: Uncommon for web servers
- **3001**: Next.js alternative port
- **8081**: Alternative API port
- **8002**: Alternative database port
- **4223/8223**: Alternative NATS ports

---

## Changing Ports

### Option 1: Environment Variables (Recommended)

Edit `.env` file:

```bash
# Custom ports
NGINX_HTTP_PORT=9090
NGINX_HTTPS_PORT=9443
FRONTEND_EXTERNAL_PORT=3002
BACKEND_EXTERNAL_PORT=8082
SURREAL_EXTERNAL_PORT=8003
NATS_EXTERNAL_PORT=4224
NATS_MONITOR_PORT=8224
```

### Option 2: Direct Docker Compose Edit

Edit `docker-compose.prod.yml`:

```yaml
services:
  agileos-nginx:
    ports:
      - "9090:80"    # Change 8090 to your preferred port
      - "9443:443"   # Change 8443 to your preferred port
```

---

## Firewall Configuration

### Windows Firewall

```powershell
# Allow AgileOS ports
New-NetFirewallRule -DisplayName "AgileOS HTTP" -Direction Inbound -LocalPort 8090 -Protocol TCP -Action Allow
New-NetFirewallRule -DisplayName "AgileOS HTTPS" -Direction Inbound -LocalPort 8443 -Protocol TCP -Action Allow
```

### Linux (UFW)

```bash
# Allow AgileOS ports
sudo ufw allow 8090/tcp
sudo ufw allow 8443/tcp
```

### Azure Network Security Group

```bash
# Add inbound rules
az network nsg rule create \
  --resource-group agileos-rg \
  --nsg-name agileos-nsg \
  --name AllowHTTP \
  --priority 100 \
  --destination-port-ranges 8090 \
  --protocol Tcp \
  --access Allow

az network nsg rule create \
  --resource-group agileos-rg \
  --nsg-name agileos-nsg \
  --name AllowHTTPS \
  --priority 110 \
  --destination-port-ranges 8443 \
  --protocol Tcp \
  --access Allow
```

---

## Troubleshooting Port Conflicts

### Check if Port is in Use

**Windows:**
```powershell
netstat -ano | findstr :8090
```

**Linux/Mac:**
```bash
lsof -i :8090
netstat -tuln | grep 8090
```

### Find Process Using Port

**Windows:**
```powershell
# Get PID from netstat output, then:
tasklist | findstr <PID>
```

**Linux/Mac:**
```bash
lsof -i :8090
```

### Kill Process on Port

**Windows:**
```powershell
# Find PID first, then:
taskkill /PID <PID> /F
```

**Linux/Mac:**
```bash
kill -9 $(lsof -t -i:8090)
```

---

## Docker Port Mapping Explained

### Format: `HOST:CONTAINER`

```yaml
ports:
  - "8090:80"
    # ↑     ↑
    # Host  Container
    # Port  Port
```

- **Host Port (8090)**: Port on your machine/VM
- **Container Port (80)**: Port inside Docker container
- Traffic to `localhost:8090` → routed to container port `80`

### Multiple Instances

To run multiple AgileOS instances:

```yaml
# Instance 1
ports:
  - "8090:80"

# Instance 2 (different host ports)
ports:
  - "9090:80"
```

---

## Production Recommendations

### Standard Web Ports (80/443)

If AgileOS is the only web app on your server:

```yaml
agileos-nginx:
  ports:
    - "80:80"      # Standard HTTP
    - "443:443"    # Standard HTTPS
```

### Behind Load Balancer

If using Azure Load Balancer or Application Gateway:

```yaml
agileos-nginx:
  ports:
    - "8080:80"    # Internal port
  # Load balancer handles 80/443 → 8080
```

### Development vs Production

**Development** (avoid conflicts):
- Use custom ports (8090, 8443)
- Easy to run alongside other projects

**Production** (standard ports):
- Use 80/443 for better UX
- No need to specify port in URL
- Better for SEO and user experience

---

## Quick Reference Card

```
┌─────────────────────────────────────────┐
│  AgileOS BPM - Port Quick Reference     │
├─────────────────────────────────────────┤
│  Main App:     http://localhost:8090    │
│  API:          http://localhost:8081    │
│  Database UI:  http://localhost:8002    │
│  NATS Monitor: http://localhost:8223    │
└─────────────────────────────────────────┘
```

---

**Last Updated**: April 2026
**Version**: 1.1.0 (Port Configuration Update)

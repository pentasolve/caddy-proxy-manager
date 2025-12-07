<div align="center">

# 🚀 Caddy Proxy Manager

**A modern, lightweight reverse proxy manager powered by Caddy**

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/docker-ready-2496ED?logo=docker&logoColor=white)](https://hub.docker.com/r/pentasolve/caddy-proxy-manager)
[![Go](https://img.shields.io/badge/go-1.21+-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![Vue](https://img.shields.io/badge/vue-3.x-4FC08D?logo=vue.js&logoColor=white)](https://vuejs.org)

*The simplest way to manage your reverse proxies with automatic HTTPS*

</div>

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| 🔒 **Automatic HTTPS** | SSL certificates from Let's Encrypt, ZeroSSL, or custom |
| ⚡ **HTTP/3 Ready** | Modern protocols with QUIC support out of the box |
| 🎨 **Modern Dashboard** | Beautiful UI built with Vue 3 + Tailwind CSS |
| 🔄 **Real-time Updates** | WebSocket-based live configuration changes |
| 🌐 **Layer 4 Proxy** | TCP/UDP streaming for databases, game servers, etc. |
| 🛡️ **Access Control** | IP whitelisting and Basic Auth protection |
| ⚖️ **Load Balancing** | Multiple upstreams with health checks |
| 👥 **Multi-user** | Role-based access control with permissions |

---

## 🚀 Quick Start

### Docker Compose

```yaml
services:
  cpm:
    container_name: caddy-proxy-manager
    image: pentasolve/caddy-proxy-manager:latest
    ports:
      - "80:80"
      - "443:443"
      - "81:81"
    volumes:
      - cpm_data:/data
      - cpm_config:/config
    restart: unless-stopped

volumes:
  cpm_data:
  cpm_config:
```

```bash
docker compose up -d
```

### Docker Run

```bash
docker run -d \
  --name caddy-proxy-manager \
  -p 80:80 -p 443:443 -p 81:81 \
  -v cpm_data:/data \
  -v cpm_config:/config \
  --restart unless-stopped \
  pentasolve/caddy-proxy-manager:latest
```

---

## 🌐 Access

| Port | Description |
|:----:|-------------|
| `80` | HTTP (auto-redirects to HTTPS) |
| `443` | HTTPS with automatic certificates |
| `81` | **Admin Dashboard** |

**📍 Dashboard:** `http://your-server-ip:81`

> On first access, you'll be prompted to create an admin account.

---

## 📖 Usage

### Adding a Reverse Proxy

1. Login to the dashboard at port `81`
2. Navigate to **Proxy Hosts** → **Add Host**
3. Configure your proxy:
   - **Domain:** `app.example.com`
   - **Target:** `192.168.1.100:3000`
   - **SSL:** Enable for automatic HTTPS
4. Click **Save** - changes apply instantly!

### Adding a TCP/UDP Stream

1. Go to **Proxy Hosts** → **Streams (L4)** tab
2. Click **Add Stream**
3. Configure:
   - **Listen Port:** `3306`
   - **Protocol:** `TCP`
   - **Target:** `db-server:3306`

### SSL Options

| Provider | Description |
|----------|-------------|
| 🔄 **Auto** | Automatic selection |
| 🔐 **Let's Encrypt** | Free, trusted certificates |
| 🔐 **ZeroSSL** | Alternative free CA |
| 🔧 **Self-Signed** | For internal/development |
| 📁 **Custom** | Upload your own certificate |

---

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

---

<div align="center">

**Made with ❤️ by [Pentasolve](https://github.com/pentasolve)**

</div>


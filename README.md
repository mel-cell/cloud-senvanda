# Senvanda Cloud 🚀

Senvanda adalah platform PaaS (Platform-as-a-Service) mandiri yang dirancang untuk memudahkan deployment aplikasi web langsung dari repositori Git ke infrastruktur Docker pribadi Anda.

## 🌟 Fitur Utama

- **Heuristic Git Analysis**: Mendeteksi framework aplikasi secara otomatis (Node.js, Vue, Go, dsb).
- **Instant Deployment**: Build dan deploy aplikasi dalam hitungan detik menggunakan Docker API.
- **Smart Dashboard**: Kelola semua container Anda melalui antarmuka premium yang modern.
- **Auto-SSL & Reverse Proxy**: Integrasi dengan Caddy untuk provisioning HTTPS otomatis.
- **Self-Healing Infrastructure**: Sistem otomatis memulihkan record database jika terjadi ketidaksinkronan dengan Docker cluster.

## 🏗️ Arsitektur Proyek

Proyek ini dibangun dengan teknologi modern:

- **Frontend**: Vue 3 + Tailwind CSS + Lucide Icons.
- **Backend Orchestrator**: Go (Golang) + Echo Framework.
- **Database & Auth**: PocketBase (SQLite).
- **Core Engine**: Docker Engine API & Git Heuristics.

## 🚀 Memulai (Development)

### Prasyarat

- Docker & Docker Compose
- Go 1.21+
- Node.js & pnpm

### Langkah Instalasi

1. Clone repositori ini.
2. Jalankan infrastruktur dasar:
   ```bash
   cd infrastructure && docker-compose up -d
   ```
3. Jalankan Backend:
   ```bash
   cd backend && go run cmd/api/main.go
   ```
4. Jalankan Frontend:
   ```bash
   cd frontend && pnpm install && pnpm dev
   ```

## 🔒 Keamanan

Senvanda dirancang untuk dijalankan di jaringan internal atau VPS pribadi. Pastikan port management (PocketBase admin) tidak terekspos langsung ke publik tanpa proteksi firewall.

---

_Built with ❤️ for rapid deployment._

---

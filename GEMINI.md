# GEMINI Protocol: Senvanda Cloud Ecosystem

Dokumen ini berfungsi sebagai **Single Source of Truth (SSOT)** untuk pengembangan, pemeliharaan, dan evolusi ekosistem **Cloud Senvanda**. Seluruh keputusan teknis dan implementasi fitur harus merujuk pada prinsip-prinsip yang tertuang di sini.

---

## 1. Lima Pilar Utama (Core Principles)

Kami berpegang teguh pada 5 prinsip dasar dalam setiap baris kode dan konfigurasi yang ditulis:

### 1. Meningkatkan Kestabilan Sistem (Stability First)

- **Zero "Flaky" Builds**: CI/CD pipeline harus deterministik. Jika berjalan sekali, harus berjalan selamanya dengan hasil yang sama.
- **Graceful Handling**: Sistem harus bisa menangani kegagalan (self-healing) tanpa membuat user panik.
- **Testing**: Validasi perubahan infrastruktur sebelum merge.

### 2. Meningkatkan Keamanan Sistem (Security Hardening)

- **Internal First**: Ekspos seminimal mungkin port ke host machine. Komunikasi antar service harus via internal Docker network (`senvanda-network`).
- **Secret Management**: Jangan pernah hardcode credential. Gunakan Environment Variables (`.env`).
- **Access Control**: Batasi akses root dan privilege container hanya jika mutlak diperlukan.

### 3. Infrastruktur yang Tertata Rapi (Structured Infrastructure)

- **Infrastructure as Code (IaC)**: Semua konfigurasi (Docker, Caddy, App Config) harus tercatat dalam kode, bukan manual setup.
- **Centralized Config**: Konfigurasi terpusat di `infrastructure/` (docker-compose, Caddyfile).
- **Clean Directories**: Data persisten disimpan rapi dalam volume yang terdefinisi, bukan scatter file.

### 4. Menghindari Duplikasi (DRY - Don't Repeat Yourself)

- **Reuse Components**: Manfaatkan service yang sudah ada. Jangan spin-up service baru jika service yang ada bisa menangani logika tersebut (contoh: gunakan Caddy yang sudah ada untuk routing baru).
- **Unified Networking**: Gunakan DNS internal Docker (`ci.senvanda.local`, `git.senvanda.local`) daripada mapping IP/Host manual yang berulang.

### 5. Meningkatkan User Experience (Premium UX)

- **Status Kejelasan**: User harus selalu tahu apa yang sedang terjadi (Building, Deploying, Failed, Online).
- **Responsivitas**: UI harus terasa "hidu" (Real-time update via SSE/WebSocket).
- **Aestetika**: Desain harus modern, bersih, dan memanjakan mata (vibrant colors, smooth animations).

---

## 2. Arsitektur Sistem (System Overview)

Ekosistem Cloud Senvanda dibangun di atas containerisasi modern dengan orkestrasi mandiri.

### Komponen Utama:

1.  **The Guard (Caddy)**: Single Entry Point. Menangani reverse proxy untuk semua subdomain (`git.*`, `ci.*`, `api.*`) di port **9080** (HTTP) dan **443** (HTTPS).
2.  **The Brain (PocketBase/Backend)**: Control plane utama. Mengelola data user, project, dan orkestrasi deployment.
3.  **The Source (Gitea)**: Git hosting internal. Sumber kebenaran kode.
4.  **The Factory (Woodpecker CI)**: Engine CI/CD. Mendeteksi push, menjalankan build, dan mem-publish artifact.
5.  **The Warehouse (Registry)**: Local Docker Registry. Menyimpan image hasil build siap deploy.

### Alur Kerja (Workflow):

`Push Code` -> `Gitea Webhook` -> `Woodpecker Build` -> `Push to Registry` -> `Notify Backend` -> `Backend Pull & Deploy` -> `Update Caddy Route`

---

## 3. Status Progress (Current State)

### ✅ Completed (Stable)

- **Unified Networking**: Caddy menangani semua routing internal/eksternal via port 9080.
- **CI/CD Pipeline**: Automasi dari `git push` hingga `deployed` berfungsi penuh.
- **Real-time Dashboard**: Frontend menerima update status via PocketBase SSE.
- **Local Registry**: Image disimpan dan diambil secara lokal untuk efisiensi bandwidth.

### 🚧 In Progress / Maintenance

- **Monitoring**: Memastikan log mudah dibaca dan diakses saat debug.
- **Security Audit**: Memastikan isolasi container antar project user.

---

## 4. Aturan Pengembangan (Rules of Engagement)

1.  **Baca Dulu**: Sebelum mengubah `docker-compose.yml` atau `Caddyfile`, pahami dampaknya ke seluruh service.
2.  **Cek Log**: Jika ada error 400/500, cek log Caddy (`docker logs senvanda-caddy`) terlebih dahulu.
3.  **Konsistensi Port**: Gunakan port **9080** untuk komunikasi HTTP internal antar service via Caddy.
4.  **Dokumentasi**: Update file ini jika ada perubahan arsitektur besar.

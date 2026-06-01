# 🐹 Zakkistore SDK for Go

**Official B2B Client Library for Zakki Store API Gateway**

Pustaka Go (Golang) resmi untuk memudahkan integrasi layanan Host-to-Host (H2H) prabayar/pascabayar, payment gateway QRIS otomatis, perbankan Virtual Account (VA), Noktel OTP virtual, mining reward, dan gacha koin Zakki Store ke dalam proyek Go Anda (Gin, Fiber, Echo, native Go backend, bot Telegram/Discord, dll).

---

## 🚀 Instalasi & Inisialisasi

Instal pustaka dari terminal proyek Go Anda:

```bash
go get github.com/MrLow12/zakkistore-sdk-go
```

### Inisialisasi Klien

#### Mode 1: Inisialisasi Instan (Official Gateway by Default)
Sangat praktis! SDK otomatis mengarah ke gateway server resmi (`https://qris.zakki.store`).

```go
package main

import (
	"fmt"
	"log"
	zakkistore "github.com/MrLow12/zakkistore-sdk-go"
)

func main() {
	// Klien otomatis mengarah ke server resmi!
	zakki, err := zakkistore.New("API_TOKEN_MEMBER_ANDA")
	if err != nil {
		log.Fatalf("Gagal inisialisasi: %v", err)
	}

	// Contoh: Melakukan Health Check Server
	status, err := zakki.Status()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("Status Server: %v\n", status["status"])
}
```

#### Mode 2: Inisialisasi dengan Konfigurasi Kustom
Gunakan opsi ini jika Anda ingin melakukan kustomisasi base URL (migrasi domain) atau mengaktifkan fitur penarikan otomatis (Auto-Withdraw).

```go
zakki, err := zakkistore.NewWithConfig(zakkistore.Config{
	BaseURL:      "https://qris.zakki.store", // Domain custom/resmi
	Token:        "API_TOKEN_MEMBER_ANDA",
	IDUser:       "IBO99",
	Email:        "member@gmail.com",
	PIN:          "123456",                   // Wajib untuk tabung & tarik
	AutoWithdraw: true,                       // Aktifkan auto-withdrawal saldo bank!
})
```

---

## 🛠️ Fitur Unggulan

### 🔄 Auto-Withdraw Saldo VA
Jika opsi `AutoWithdraw: true` diaktifkan, SDK akan memicu penarikan dana VA bank otomatis secara *real-time* menjadi saldo utama aplikasi zakki store ketika fungsi `zakki.Checkbank()` dipanggil.

### 💡 Dual-Flow Pascabayar & Bebas Nominal
*   **Pascabayar (PLN/BPJS/PDAM):** Inquiry tagihan terlebih dahulu, lalu bayar dengan format tujuan `[ID_Pelanggan].[Nominal_Tagihan]` (Contoh: `122345678901.150000`).
*   **E-Wallet Bebas Nominal:** Kirim transfer E-Wallet nominal kustom dengan format tujuan `[No_HP].[Nominal]` (Contoh: `08123456789.25000`).

---

## 📑 Daftar Referensi Metode Lengkap & Struktur Pengelompokan (36 Fungsi Resmi)

Seluruh fungsi yang didukung oleh SDK ini dikelompokkan secara rapi ke dalam 7 kategori layanan utama demi mempermudah pemahaman dan integrasi:

### 1. ⚡ Layanan Payment Gateway (QRIS Topup) — [4 Fungsi]
*   **`zakki.Topup(nominal)`** — Membuat tiket pembayaran QRIS dinamis instan dengan nominal kode unik.
*   **`zakki.Cektopup(idtopup)`** — Mengecek status pembayaran tiket QRIS tertentu secara real-time.
*   **`zakki.Mytopup()`** — Mengambil seluruh riwayat transaksi topup QRIS akun Anda.
*   **`zakki.Cancel(idTransaksi, allPending)`** — Membatalkan satu atau seluruh tiket topup pending.

### 2. 🏪 Layanan Transaksi Host-to-Host (H2H) — [4 Fungsi]
*   **`zakki.Listkode(jenis, productType)`** — Mengambil katalog produk prabayar/pascabayar aktif beserta daftar harga beli.
*   **`zakki.H2H(params)`** — Mengirimkan order transaksi H2H (pulsa, paket data, PLN kustom, dll).
*   **`zakki.Cekh2h(idTrx)`** — Mengecek status transaksi, Serial Number (SN), dan harga beli riil dari order H2H.
*   **`zakki.Myh2h()`** — Mengambil 20 riwayat transaksi H2H terupdate milik akun Anda.

### 3. 🏦 Layanan Perbankan & Transfer Saldo VA — [8 Fungsi]
*   **`zakki.Checkbank()`** — Memeriksa detail Virtual Account (VA), saldo bank VA, serta memicu Auto-Withdraw jika diaktifkan.
*   **`zakki.Checkname(number)`** — Memverifikasi nama asli pemilik rekening Virtual Account tujuan sebelum melakukan transfer.
*   **`zakki.Transfer(params)`** — Mengirimkan saldo antar-VA member secara instan dan bebas biaya admin.
*   **`zakki.Tabung(jumlah)`** — Menyetorkan saldo aktif aplikasi ke rekening bank Virtual Account terhubung Anda.
*   **`zakki.Tarik(jumlah)`** — Menarik dana dari bank Virtual Account ke saldo aktif aplikasi Zakki Store Anda.
*   **`zakki.Checkmutasi(mutasiType)`** — Melihat riwayat mutasi tabung/tarik saldo bank VA (`all`, `tarik`, `tabung`).
*   **`zakki.Checktransfer(idtransfer)`** — Mengecek status pengiriman dana transfer tertentu secara detail.
*   **`zakki.Mytransfer(type)`** — Mengambil riwayat pengiriman dan penerimaan transfer saldo (`all`, `kirim`, `terima`).

### 4. 📱 Layanan Noktel Marketplace (OTP Virtual) — [5 Fungsi]
*   **`zakki.NoktelStok()`** — Memeriksa ketersediaan stok nomor virtual aktif per kategori layanan/aplikasi.
*   **`zakki.NoktelBuy(category)`** — Membeli nomor virtual baru untuk penerimaan kode verifikasi/OTP.
*   **`zakki.NoktelGetOtp(accountId)`** — Mengambil kode verifikasi/OTP yang masuk ke nomor virtual secara real-time.
*   **`zakki.NoktelCancel(invoiceId)`** — Membatalkan order nomor virtual yang pending OTP dan memicu auto-refund saldo.
*   **`zakki.NoktelHistory()`** — Mengambil daftar riwayat lengkap pemesanan nomor virtual.

### 5. ⛏️ Layanan Reward Komputasi SHA-256 (Mining) & Game — [5 Fungsi]
*   **`zakki.MiningStart()`** — Meminta challenge penambangan SHA-256 serta target kesulitan (difficulty) dari server.
*   **`zakki.MiningSubmit(nonce, signature)`** — Mengirimkan hasil kerja hashing SHA-256 (Proof-of-Work) untuk mendapatkan koin.
*   **`zakki.Cekmining(idmining)`** — Mengecek status audit dan persetujuan dari blok mining yang telah Anda selesaikan.
*   **`zakki.Mymining()`** — Melihat riwayat penambangan koin dan total reward hashing akun Anda.
*   **`zakki.Cekgacha()`** — Mengecek jumlah tiket gacha, riwayat kemenangan, dan detail koin keberuntungan Anda.

### 6. 🔒 Layanan Keamanan IP & Utilitas — [6 Fungsi]
*   **`zakki.Whitelistip(ip)`** — Mendaftarkan IP server/host Anda agar diizinkan melakukan transaksi H2H via API (Maksimal 3 IP).
*   **`zakki.Delwhitelistip(ip)`** — Menghapus alamat IP terdaftar dari whitelist API.
*   **`zakki.Cekmyip()`** — Mendeteksi alamat IP publik host/server Anda saat ini yang terbaca oleh sistem.
*   **`zakki.Cekip(ip)`** — Mengecek detail status IP whitelisting tertentu.
*   **`zakki.Leaderboard(limit, period)`** — Melihat daftar Sultan topup teraktif secara global.
*   **`zakki.Status()`** — Memeriksa beban CPU server, statistik finansial global, dan kesehatan sistem.

### 7. 🔗 Layanan Webhook Callback & Notifikasi Bot — [4 Fungsi]
*   **`zakki.Setcallback(site)`** — Memasang URL callback real-time untuk menerima laporan status transaksi H2H.
*   **`zakki.Delcallback()`** — Menghapus URL callback yang terpasang di sistem.
*   **`zakki.Setnotifbot(telegramId)`** — Memasang ID Telegram Anda untuk menerima notifikasi otomatis transaksi sukses/gagal.
*   **`zakki.Delnotifbot()`** — Menonaktifkan bot notifikasi Telegram.


## 🛡️ Protokol Keamanan API

> [!WARNING]
> **Selalu jalankan SDK ini di sisi backend (Server-side)!**
> Jangan pernah mengekspos API Token dan PIN Anda di sisi frontend / client-side publik demi mencegah potensi pencurian saldo.

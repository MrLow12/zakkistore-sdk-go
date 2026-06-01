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

## 📑 Daftar Referensi Metode Lengkap

SDK Go ini mendukung secara penuh seluruh **25 fungsi resmi** dengan nama dan perilaku yang konsisten dengan SDK versi Node.js (NPM), Python (PyPI), dan PHP (Composer):

### 1. Payment Gateway (QRIS Top Up)
*   `zakki.Topup(nominal int)` — Membuat QRIS dinamis instan dengan nominal kode unik.
*   `zakki.Cektopup(idtopup string)` — Cek status pembayaran QRIS.
*   `zakki.Cancel(idTransaksi string, allPending bool)` — Batalkan transaksi pending (Daftar pending, batal satu, atau batal massal).

### 2. Transaksi H2H
*   `zakki.Listkode(jenis, productType string)` — Katalog kode produk aktif, deskripsi, dan harga.
*   `zakki.H2H(params zakkistore.H2HParams)` — Mengirim order transaksi H2H.
*   `zakki.H2HSimple(kode, tujuan, refID string)` — Versi sederhana posisional untuk memicu order H2H.
*   `zakki.Cekh2h(idTrx string)` — Cek detail status pengisian, SN, dan harga beli order H2H.
*   `zakki.Myh2h()` — Mengambil 20 riwayat pembelian H2H terupdate.

### 3. Perbankan & Transfer VA
*   `zakki.Checkbank()` — Cek saldo, VA member, mutasi, dan pemicu Auto-Withdraw.
*   `zakki.Checkname(number string)` — Verifikasi nama asli pemilik VA Bank Zakki tujuan.
*   `zakki.Transfer(params zakkistore.TransferParams)` — Transfer saldo antar Virtual Account member Bank Zakki.
*   `zakki.TransferSimple(to string, amount int)` — Versi sederhana posisional untuk transfer saldo.
*   `zakki.Tabung(jumlah int)` — Menabung / deposit saldo dari aplikasi zakki store ke Bank (butuh PIN).
*   `zakki.Tarik(jumlah int)` — Menarik dana tabungan ke saldo aplikasi (butuh PIN).
*   `zakki.Checkmutasi(mutasiType string)` — Riwayat mutasi Tarik/Tabung (`tarik`, `tabung`, `all`).

### 4. Noktel Marketplace (OTP Virtual)
*   `zakki.NoktelStok()` — Cek stok nomor virtual yang ready.
*   `zakki.NoktelBuy(category string)` — Membeli nomor virtual baru untuk OTP.
*   `zakki.NoktelGetOtp(accountID string)` — Menarik kode OTP Telegram secara real-time.
*   `zakki.NoktelCancel(invoiceID string)` — Membatalkan nomor yang pending OTP & auto-refund.
*   `zakki.NoktelHistory()` — Mengambil daftar riwayat pembelian Noktel.

### 5. Reward Komputasi & Game
*   `zakki.Cekmining()` — Cek status kesulitan global, block reward, dan miner aktif.
*   `zakki.Mymining()` — Riwayat koin mining SHA256 milik akun Anda.
*   `zakki.Cekgacha()` — Statistik poin, kemenangan, dan keuntungan gacha member.

### 6. Keamanan & Utilitas
*   `zakki.Whitelistip(ip string)` — Whitelist IP server Anda untuk otorisasi API H2H.
*   `zakki.Delwhitelistip(ip string)` — Hapus IP server dari whitelist.
*   `zakki.Leaderboard(limit int, period string)` — Mengambil peringkat sultan topup teraktif.
*   `zakki.Status()` — Informasi beban CPU, metrik finansial, dan kesehatan sistem.

---

## 🛡️ Protokol Keamanan API

> [!WARNING]
> **Selalu jalankan SDK ini di sisi backend (Server-side)!**
> Jangan pernah mengekspos API Token dan PIN Anda di sisi frontend / client-side publik demi mencegah potensi pencurian saldo.

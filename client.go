package zakkistore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Config menyimpan parameter konfigurasi klien ZakkiStore SDK.
type Config struct {
	BaseURL      string
	Token        string
	IDUser       string
	Email        string
	PIN          string
	AutoWithdraw bool
}

// ZakkiStore adalah klien resmi untuk mengakses API Gateway B2B Zakki Store.
type ZakkiStore struct {
	baseURL        string
	token          string
	idUser         string
	email          string
	pin            string
	isAutoWithdraw bool
	httpClient     *http.Client
}

// H2HParams mendefinisikan parameter untuk transaksi H2H (Host-to-Host).
type H2HParams struct {
	Kode   string `json:"kode"`
	Tujuan string `json:"tujuan"`
	RefID  string `json:"refID"`
}

// TransferParams mendefinisikan parameter untuk transfer saldo antar Virtual Account member.
type TransferParams struct {
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

// New menginisialisasi klien ZakkiStore baru menggunakan default base URL Gateway Resmi.
func New(token string) (*ZakkiStore, error) {
	return NewWithConfig(Config{
		Token: token,
	})
}

// NewWithConfig menginisialisasi klien ZakkiStore baru dengan konfigurasi kustom lengkap.
func NewWithConfig(cfg Config) (*ZakkiStore, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("token wajib disertakan dalam konfigurasi SDK")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://qris.zakki.store"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &ZakkiStore{
		baseURL:        baseURL,
		token:          cfg.Token,
		idUser:         cfg.IDUser,
		email:          cfg.Email,
		pin:            cfg.PIN,
		isAutoWithdraw: cfg.AutoWithdraw,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// EnableAutoWithdraw mengaktifkan atau menonaktifkan fitur penarikan otomatis (Auto-Withdraw).
func (z *ZakkiStore) EnableAutoWithdraw(status bool) {
	z.isAutoWithdraw = status
}

// Request helper internal untuk HTTP request.
func (z *ZakkiStore) request(endpoint string, method string, data interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s%s", z.baseURL, endpoint)
	var reqBody []byte
	var err error

	if data != nil {
		if method == "GET" {
			var m map[string]interface{}
			b, _ := json.Marshal(data)
			json.Unmarshal(b, &m)
			
			queryParams := []string{}
			for k, v := range m {
				queryParams = append(queryParams, fmt.Sprintf("%s=%v", k, v))
			}
			if len(queryParams) > 0 {
				url = fmt.Sprintf("%s?%s", url, strings.Join(queryParams, "&"))
			}
		} else {
			reqBody, err = json.Marshal(data)
			if err != nil {
				return nil, fmt.Errorf("[ZakkiStore SDK Error] Gagal marshal payload: %w", err)
			}
		}
	}

	var req *http.Request
	if len(reqBody) > 0 {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("[ZakkiStore SDK Error] Gagal membuat request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[ZakkiStore SDK Error] Koneksi Gagal: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("[ZakkiStore SDK Error] Gagal membaca JSON response: %w", err)
	}

	if resp.StatusCode >= 400 {
		errMsg, ok := result["message"].(string)
		if !ok {
			errMsg = fmt.Sprintf("HTTP Error! Status: %d", resp.StatusCode)
		}
		if resp.StatusCode == 403 || strings.Contains(strings.ToLower(errMsg), "ip") {
			errMsg += "\n⚠️ [IP BLOCKED / UNREGISTERED] IP Anda diblokir atau belum terdaftar di whitelist API. Silakan hubungi developer via WhatsApp (https://wa.me/6283844082339) atau Telegram (https://t.me/zakki_store) untuk mendapatkan bantuan."
		}
		return nil, fmt.Errorf("[ZakkiStore SDK Error] %s", errMsg)
	}

	return result, nil
}

// ==========================================================
// --- 1. PAYMENT GATEWAY (QRIS TOPUP) ---
// ==========================================================

// Topup membuat invoice transaksi QRIS dinamis instan dengan nominal kode unik.
func (z *ZakkiStore) Topup(nominal int) (map[string]interface{}, error) {
	return z.request("/topup", "POST", map[string]interface{}{
		"token":   z.token,
		"nominal": nominal,
	})
}

// Cektopup memeriksa status pembayaran tagihan/topup QRIS berdasarkan ID topup.
func (z *ZakkiStore) Cektopup(idtopup string) (map[string]interface{}, error) {
	return z.request("/cektopup", "GET", map[string]string{
		"idtopup": idtopup,
	})
}

// Cektopup2 mendapatkan URL gambar struk digital dinamis (hologram receipt) berformat PNG.
func (z *ZakkiStore) Cektopup2(idtopup string) string {
	return fmt.Sprintf("%s/cektopup2?idtopup=%s", strings.TrimSuffix(z.baseURL, "/"), url.QueryEscape(idtopup))
}

// Cancel membatalkan transaksi pending (mendukung pembatalan massal atau spesifik ID).
func (z *ZakkiStore) Cancel(idTransaksi string, allPending bool) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"token": z.token,
	}
	if idTransaksi != "" {
		payload["id_transaksi"] = idTransaksi
	}
	if allPending {
		payload["all"] = true
	}
	return z.request("/cancel", "POST", payload)
}

// ==========================================================
// --- 2. TRANSAKSI H2H (HOST-TO-HOST) ---
// ==========================================================

// Listkode mengambil daftar katalog kode produk H2H, deskripsi, dan harga terupdate.
func (z *ZakkiStore) Listkode(jenis, productType string) (map[string]interface{}, error) {
	payload := map[string]string{}
	if jenis != "" {
		payload["jenis"] = jenis
	}
	if productType != "" {
		payload["type"] = productType
	}
	return z.request("/listkode", "GET", payload)
}

// H2H mengirimkan order transaksi pembelian produk prabayar/pascabayar H2H.
func (z *ZakkiStore) H2H(params H2HParams) (map[string]interface{}, error) {
	return z.request("/h2h", "POST", map[string]interface{}{
		"token":  z.token,
		"kode":   params.Kode,
		"tujuan": params.Tujuan,
		"refID":  params.RefID,
	})
}

// H2HSimple mengirim order H2H dengan parameter posisional (kode, tujuan, refID).
func (z *ZakkiStore) H2HSimple(kode, tujuan, refID string) (map[string]interface{}, error) {
	return z.H2H(H2HParams{Kode: kode, Tujuan: tujuan, RefID: refID})
}

// Cekh2h memeriksa status transaksi, Serial Number (SN), dan harga beli order H2H.
func (z *ZakkiStore) Cekh2h(idTrx string) (map[string]interface{}, error) {
	return z.request("/cekh2h", "GET", map[string]string{
		"id": idTrx,
	})
}

// Myh2h mengambil 20 riwayat pembelian H2H terupdate milik akun Anda.
func (z *ZakkiStore) Myh2h() (map[string]interface{}, error) {
	return z.request("/myh2h", "GET", map[string]string{
		"token": z.token,
	})
}

// ==========================================================
// --- 3. PERBANKAN & TRANSFER SALDO ---
// ==========================================================

// Checkbank memeriksa saldo VA, detail mutasi bank, dan memicu penarikan saldo otomatis (Auto-Withdraw).
func (z *ZakkiStore) Checkbank() (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"token": z.token,
	}
	if z.idUser != "" {
		payload["iduser"] = z.idUser
	} else if z.email != "" {
		payload["email"] = z.email
	}

	bankRes, err := z.request("/checkbank", "GET", payload)
	if err != nil {
		return nil, err
	}

	// Alur Auto-Withdraw VA Bank Otomatis
	if z.isAutoWithdraw {
		if data, ok := bankRes["data"].(map[string]interface{}); ok {
			if bankDetail, ok := data["bank_detail"].(map[string]interface{}); ok {
				var balance float64
				switch v := bankDetail["balance"].(type) {
				case float64:
					balance = v
				case int:
					balance = float64(v)
				}

				if balance > 0 {
					withdrawRes, err := z.Tarik(int(balance))
					if err == nil {
						updatedRes, err := z.request("/checkbank", "GET", payload)
						if err == nil {
							bankRes = updatedRes
							bankRes["auto_withdraw_executed"] = true
							bankRes["auto_withdraw_amount"] = int(balance)
							if msg, ok := withdrawRes["message"].(string); ok {
								bankRes["auto_withdraw_message"] = msg
							} else {
								bankRes["auto_withdraw_message"] = "Auto-withdraw berhasil dijalankan."
							}
						}
					} else {
						bankRes["auto_withdraw_executed"] = false
						bankRes["auto_withdraw_error"] = err.Error()
					}
				}
			}
		}
	}

	return bankRes, nil
}

// Checkname memverifikasi nama pemilik nomor Virtual Account (VA) Bank Zakki tujuan.
func (z *ZakkiStore) Checkname(number string) (map[string]interface{}, error) {
	return z.request("/checkname", "GET", map[string]string{
		"number": strings.TrimSpace(number),
	})
}

// Transfer mengirimkan saldo antar rekening Virtual Account member Bank Zakki.
func (z *ZakkiStore) Transfer(params TransferParams) (map[string]interface{}, error) {
	return z.request("/transfer", "POST", map[string]interface{}{
		"token":  z.token,
		"to":     params.To,
		"amount": params.Amount,
	})
}

// TransferSimple mentransfer saldo dengan parameter posisional (to, amount).
func (z *ZakkiStore) TransferSimple(to string, amount int) (map[string]interface{}, error) {
	return z.Transfer(TransferParams{To: to, Amount: amount})
}

// Tabung menabung/menyetorkan dana dari aplikasi utama ke rekening Virtual Account Bank (butuh PIN).
func (z *ZakkiStore) Tabung(jumlah int) (map[string]interface{}, error) {
	if z.pin == "" {
		return nil, fmt.Errorf("[ZakkiStore SDK Error] PIN transaksi diperlukan untuk melakukan transaksi tabung")
	}

	payload := map[string]interface{}{
		"token":  z.token,
		"jumlah": jumlah,
		"pin":    z.pin,
	}

	if z.idUser != "" {
		payload["iduser"] = z.idUser
	}
	if z.email != "" {
		payload["email"] = z.email
	}

	return z.request("/tabung", "POST", payload)
}

// Tarik menarik dana tabungan Virtual Account ke saldo aplikasi utama (butuh PIN).
func (z *ZakkiStore) Tarik(jumlah int) (map[string]interface{}, error) {
	if z.pin == "" {
		return nil, fmt.Errorf("[ZakkiStore SDK Error] PIN transaksi diperlukan untuk melakukan transaksi tarik")
	}

	payload := map[string]interface{}{
		"token":  z.token,
		"jumlah": jumlah,
		"pin":    z.pin,
	}

	if z.idUser != "" {
		payload["iduser"] = z.idUser
	}
	if z.email != "" {
		payload["email"] = z.email
	}

	return z.request("/tarik", "POST", payload)
}

// Checkmutasi mengambil daftar riwayat mutasi Tarik/Tabung berdasarkan tipe.
func (z *ZakkiStore) Checkmutasi(mutasiType string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"token": z.token,
		"type":  mutasiType,
	}

	if z.idUser != "" {
		payload["iduser"] = z.idUser
	}
	if z.email != "" {
		payload["email"] = z.email
	}

	return z.request("/checkmutasi", "GET", payload)
}

// ==========================================================
// --- 4. NOKTEL MARKETPLACE (OTP VIRTUAL) ---
// ==========================================================

// NoktelStok memeriksa stok nomor virtual yang ready untuk dipesan.
func (z *ZakkiStore) NoktelStok() (map[string]interface{}, error) {
	return z.request("/noktel/stok", "GET", map[string]string{
		"token": z.token,
	})
}

// NoktelBuy membeli nomor virtual baru berdasarkan kategori layanan.
func (z *ZakkiStore) NoktelBuy(category string) (map[string]interface{}, error) {
	return z.request("/noktel/buy", "POST", map[string]string{
		"token":    z.token,
		"category": strings.TrimSpace(category),
	})
}

// NoktelGetOtp menarik kode OTP Telegram secara real-time dari nomor yang dibeli.
func (z *ZakkiStore) NoktelGetOtp(accountID string) (map[string]interface{}, error) {
	return z.request("/noktel/getotp", "GET", map[string]string{
		"token":      z.token,
		"account_id": strings.TrimSpace(accountID),
	})
}

// NoktelCancel membatalkan nomor yang pending OTP dan melakukan auto-refund saldo.
func (z *ZakkiStore) NoktelCancel(invoiceID string) (map[string]interface{}, error) {
	return z.request("/noktel/cancel", "POST", map[string]string{
		"token":      z.token,
		"invoice_id": strings.TrimSpace(invoiceID),
	})
}

// NoktelHistory mengambil daftar riwayat transaksi pembelian nomor Noktel.
func (z *ZakkiStore) NoktelHistory() (map[string]interface{}, error) {
	return z.request("/noktel/history", "GET", map[string]string{
		"token": z.token,
	})
}

// ==========================================================
// --- 5. REWARD KOMPUTASI & GAME ---
// ==========================================================

// Cekmining memeriksa detail status transaksi mining koin spesifik berdasarkan ID.
func (z *ZakkiStore) Cekmining(idmining string) (map[string]interface{}, error) {
	if idmining == "" {
		return nil, fmt.Errorf("parameter idmining wajib diisi")
	}
	return z.request("/cekmining", "GET", map[string]string{
		"idmining": strings.TrimSpace(idmining),
	})
}

// Mymining mengambil riwayat koin hasil mining SHA256 milik akun Anda.
func (z *ZakkiStore) Mymining() (map[string]interface{}, error) {
	return z.request("/mymining", "GET", map[string]string{
		"token": z.token,
	})
}

// MiningStart meminta tantangan (challenge) mining baru dari server.
func (z *ZakkiStore) MiningStart() (map[string]interface{}, error) {
	return z.request("/mining/start", "GET", map[string]string{
		"token": z.token,
	})
}

// MiningSubmit mengirimkan nonce tebakan dan signature untuk divalidasi oleh server.
func (z *ZakkiStore) MiningSubmit(nonce interface{}, signature string) (map[string]interface{}, error) {
	if nonce == nil {
		return nil, fmt.Errorf("parameter nonce wajib disertakan")
	}
	if signature == "" {
		return nil, fmt.Errorf("parameter signature wajib disertakan")
	}
	return z.request("/mining/submit", "POST", map[string]interface{}{
		"token":     z.token,
		"nonce":     nonce,
		"signature": signature,
	})
}

// Cekgacha memeriksa statistik poin, kemenangan, dan keuntungan gacha member.
func (z *ZakkiStore) Cekgacha() (map[string]interface{}, error) {
	return z.request("/cekgacha", "GET", map[string]string{
		"token": z.token,
	})
}

// ==========================================================
// --- 6. UTILITY & SECURITY ---
// ==========================================================

// Whitelistip menambahkan IP server Anda ke whitelist API H2H.
func (z *ZakkiStore) Whitelistip(ip string) (map[string]interface{}, error) {
	return z.request("/whitelistip", "POST", map[string]string{
		"token": z.token,
		"ip":    strings.TrimSpace(ip),
	})
}

// Delwhitelistip menghapus IP server Anda dari whitelist API H2H.
func (z *ZakkiStore) Delwhitelistip(ip string) (map[string]interface{}, error) {
	return z.request("/delwhitelistip", "POST", map[string]string{
		"token": z.token,
		"ip":    strings.TrimSpace(ip),
	})
}

// Leaderboard mengambil daftar peringkat sultan topup member teraktif.
func (z *ZakkiStore) Leaderboard(limit int, period string) (map[string]interface{}, error) {
	return z.request("/leaderboard", "GET", map[string]interface{}{
		"limit":  limit,
		"period": strings.TrimSpace(period),
	})
}

// Status memeriksa metrik CPU, beban finansial, dan kesehatan sistem global.
func (z *ZakkiStore) Status() (map[string]interface{}, error) {
	return z.request("/status", "GET", nil)
}

// ==========================================================
// --- 7. METODE INTEGRASI BARU ---
// ==========================================================

// Setcallback mendaftarkan URL callback HTTPS Anda untuk menerima notifikasi otomatis.
func (z *ZakkiStore) Setcallback(site string) (map[string]interface{}, error) {
	return z.request("/setcallback", "GET", map[string]string{
		"token": z.token,
		"site":  strings.TrimSpace(site),
	})
}

// Delcallback menghapus URL callback yang terdaftar.
func (z *ZakkiStore) Delcallback() (map[string]interface{}, error) {
	return z.request("/delcallback", "GET", map[string]string{
		"token": z.token,
	})
}

// Setnotifbot mendaftarkan ID Telegram Anda untuk notifikasi laporan transaksi bot.
func (z *ZakkiStore) Setnotifbot(telegramID string) (map[string]interface{}, error) {
	return z.request("/setnotifbot", "GET", map[string]string{
		"token": z.token,
		"id":    strings.TrimSpace(telegramID),
	})
}

// Delnotifbot menghapus ID Telegram yang terdaftar untuk menonaktifkan notifikasi bot.
func (z *ZakkiStore) Delnotifbot() (map[string]interface{}, error) {
	return z.request("/delnotifbot", "GET", map[string]string{
		"token": z.token,
	})
}

// Checktransfer memverifikasi status transfer saldo antar member berdasarkan ID transfer.
func (z *ZakkiStore) Checktransfer(idtransfer string) (map[string]interface{}, error) {
	return z.request("/checktransfer", "GET", map[string]string{
		"idtransfer": strings.TrimSpace(idtransfer),
	})
}

// Mytransfer mengambil daftar riwayat transfer saldo masuk/keluar.
func (z *ZakkiStore) Mytransfer(transferType string) (map[string]interface{}, error) {
	return z.request("/mytransfer", "GET", map[string]string{
		"token": z.token,
		"type":  strings.TrimSpace(transferType),
	})
}

// Mytopup mengambil daftar riwayat topup sukses beserta total volume topup.
func (z *ZakkiStore) Mytopup() (map[string]interface{}, error) {
	return z.request("/mytopup", "GET", map[string]string{
		"token": z.token,
	})
}

// Cekmyip mengecek alamat IP publik server Anda yang terdeteksi oleh gateway.
func (z *ZakkiStore) Cekmyip() (map[string]interface{}, error) {
	return z.request("/cekmyip", "GET", nil)
}

// Cekip memverifikasi status keamanan IP tertentu (Aman/Whitelist/Blacklist).
func (z *ZakkiStore) Cekip(ip string) (map[string]interface{}, error) {
	return z.request("/cekip", "GET", map[string]string{
		"ip": strings.TrimSpace(ip),
	})
}

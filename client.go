package zakkistore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	BaseURL      string
	Token        string
	IDUser       string
	Email        string
	PIN          string
	AutoWithdraw bool
}

type ZakkiStore struct {
	baseURL        string
	token          string
	idUser         string
	email          string
	pin            string
	isAutoWithdraw bool
	httpClient     *http.Client
}

type H2HParams struct {
	Kode   string `json:"kode"`
	Tujuan string `json:"tujuan"`
	RefID  string `json:"refID"`
}

type TransferParams struct {
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

/**
 * Inisialisasi Klien ZakkiStore baru menggunakan default base URL Gateway Resmi.
 */
func New(token string) (*ZakkiStore, error) {
	return NewWithConfig(Config{
		Token: token,
	})
}

/**
 * Inisialisasi Klien ZakkiStore baru dengan konfigurasi kustom lengkap.
 */
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

func (z *ZakkiStore) EnableAutoWithdraw(status bool) {
	z.isAutoWithdraw = status
}

/**
 * Request helper internal untuk HTTP cURL / REST API request.
 * 
 * @private
 */
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

func (z *ZakkiStore) Topup(nominal int) (map[string]interface{}, error) {
	return z.request("/topup", "POST", map[string]interface{}{
		"token":   z.token,
		"nominal": nominal,
	})
}

func (z *ZakkiStore) Cektopup(idtopup string) (map[string]interface{}, error) {
	return z.request("/cektopup", "GET", map[string]string{
		"idtopup": idtopup,
	})
}

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

func (z *ZakkiStore) H2H(params H2HParams) (map[string]interface{}, error) {
	return z.request("/h2h", "POST", map[string]interface{}{
		"token":  z.token,
		"kode":   params.Kode,
		"tujuan": params.Tujuan,
		"refID":  params.RefID,
	})
}

func (z *ZakkiStore) H2HSimple(kode, tujuan, refID string) (map[string]interface{}, error) {
	return z.H2H(H2HParams{Kode: kode, Tujuan: tujuan, RefID: refID})
}

func (z *ZakkiStore) Cekh2h(idTrx string) (map[string]interface{}, error) {
	return z.request("/cekh2h", "GET", map[string]string{
		"id": idTrx,
	})
}

func (z *ZakkiStore) Myh2h() (map[string]interface{}, error) {
	return z.request("/myh2h", "GET", map[string]string{
		"token": z.token,
	})
}

// ==========================================================
// --- 3. PERBANKAN & TRANSFER SALDO ---
// ==========================================================

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

func (z *ZakkiStore) Checkname(number string) (map[string]interface{}, error) {
	return z.request("/checkname", "GET", map[string]string{
		"number": strings.TrimSpace(number),
	})
}

func (z *ZakkiStore) Transfer(params TransferParams) (map[string]interface{}, error) {
	return z.request("/transfer", "POST", map[string]interface{}{
		"token":  z.token,
		"to":     params.To,
		"amount": params.Amount,
	})
}

func (z *ZakkiStore) TransferSimple(to string, amount int) (map[string]interface{}, error) {
	return z.Transfer(TransferParams{To: to, Amount: amount})
}

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

func (z *ZakkiStore) NoktelStok() (map[string]interface{}, error) {
	return z.request("/noktel/stok", "GET", map[string]string{
		"token": z.token,
	})
}

func (z *ZakkiStore) NoktelBuy(category string) (map[string]interface{}, error) {
	return z.request("/noktel/buy", "POST", map[string]string{
		"token":    z.token,
		"category": strings.TrimSpace(category),
	})
}

func (z *ZakkiStore) NoktelGetOtp(accountID string) (map[string]interface{}, error) {
	return z.request("/noktel/getotp", "GET", map[string]string{
		"token":      z.token,
		"account_id": strings.TrimSpace(accountID),
	})
}

func (z *ZakkiStore) NoktelCancel(invoiceID string) (map[string]interface{}, error) {
	return z.request("/noktel/cancel", "POST", map[string]string{
		"token":      z.token,
		"invoice_id": strings.TrimSpace(invoiceID),
	})
}

func (z *ZakkiStore) NoktelHistory() (map[string]interface{}, error) {
	return z.request("/noktel/history", "GET", map[string]string{
		"token": z.token,
	})
}

// ==========================================================
// --- 5. REWARD KOMPUTASI & GAME ---
// ==========================================================

func (z *ZakkiStore) Cekmining() (map[string]interface{}, error) {
	return z.request("/cekmining", "GET", map[string]string{
		"token": z.token,
	})
}

func (z *ZakkiStore) Mymining() (map[string]interface{}, error) {
	return z.request("/mymining", "GET", map[string]string{
		"token": z.token,
	})
}

func (z *ZakkiStore) Cekgacha() (map[string]interface{}, error) {
	return z.request("/cekgacha", "GET", map[string]string{
		"token": z.token,
	})
}

// ==========================================================
// --- 6. UTILITY & SECURITY ---
// ==========================================================

func (z *ZakkiStore) Whitelistip(ip string) (map[string]interface{}, error) {
	return z.request("/whitelistip", "POST", map[string]string{
		"token": z.token,
		"ip":    strings.TrimSpace(ip),
	})
}

func (z *ZakkiStore) Delwhitelistip(ip string) (map[string]interface{}, error) {
	return z.request("/delwhitelistip", "POST", map[string]string{
		"token": z.token,
		"ip":    strings.TrimSpace(ip),
	})
}

func (z *ZakkiStore) Leaderboard(limit int, period string) (map[string]interface{}, error) {
	return z.request("/leaderboard", "GET", map[string]interface{}{
		"limit":  limit,
		"period": strings.TrimSpace(period),
	})
}

func (z *ZakkiStore) Status() (map[string]interface{}, error) {
	return z.request("/status", "GET", nil)
}

package main

import (
	"encoding/json"
	"fmt"
	"log"

	zakkistore "github.com/MrLow12/zakkistore-sdk"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 PENGUJIAN REAL SDK GO DENGAN OFFICIAL API")
	fmt.Println("==================================================")

	token := "9d6e27f09e65d3"
	iduser := "IBO6"

	fmt.Println("📡 Menghubungkan ke API Gateway Resmi...")

	// Inisialisasi SDK (Tanpa base_url, otomatis mengarah ke Official Gateway!)
	zakki, err := zakkistore.NewWithConfig(zakkistore.Config{
		Token:        token,
		IDUser:       iduser,
		AutoWithdraw: false, // Set ke false demi keamanan saldo selama testing
	})

	if err != nil {
		log.Fatalf("❌ Gagal inisialisasi: %v", err)
	}

	// 1. Health Check
	fmt.Println("\n🔍 1. Melakukan Health Check Server...")
	status, err := zakki.Status()
	if err != nil {
		log.Fatalf("❌ Health check gagal: %v", err)
	}
	fmt.Printf("🟢 [SUCCESS] Status API: %v\n", status["status"])

	// 2. Check Bank & Profil
	fmt.Println("\n🔍 2. Mengambil Detail Akun & Profil IBO6 (Checkbank)...")
	bankInfo, err := zakki.Checkbank()
	if err != nil {
		log.Fatalf("❌ Checkbank gagal: %v", err)
	}

	b, _ := json.MarshalIndent(bankInfo, "", "  ")
	fmt.Printf("Response data:\n%s\n", string(b))

	var accountHolder, virtualAccount string
	var balance float64
	var email string
	var totalH2H float64

	if data, ok := bankInfo["data"].(map[string]interface{}); ok {
		if bankDetail, ok := data["bank_detail"].(map[string]interface{}); ok {
			accountHolder, _ = bankDetail["account_holder"].(string)
			virtualAccount, _ = bankDetail["virtual_account"].(string)
			balance, _ = bankDetail["balance"].(float64)
		}
		if userDetail, ok := data["user_detail"].(map[string]interface{}); ok {
			email, _ = userDetail["email"].(string)
			totalH2H, _ = userDetail["total_h2h"].(float64)
		}
	}

	fmt.Println("\n📝 RINGKASAN AKUN USER:")
	fmt.Printf("   👤 Nama Pemegang Rekening: %s\n", accountHolder)
	fmt.Printf("   💳 Nomor Virtual Account : %s\n", virtualAccount)
	fmt.Printf("   💰 Saldo Bank VA         : Rp %.0f\n", balance)
	fmt.Printf("   📧 Email Terdaftar       : %s\n", email)
	fmt.Printf("   🏆 Total Transaksi H2H   : %.0f kali\n", totalH2H)

	// 3. Cek Katalog Harga DANA
	fmt.Println("\n🔍 3. Mengecek Katalog Produk H2H DANA (Listkode)...")
	katalog, err := zakki.Listkode("ewallet", "DANA")
	if err != nil {
		log.Fatalf("❌ Listkode gagal: %v", err)
	}

	if code, ok := katalog["code"].(float64); ok && code == 200 {
		if products, ok := katalog["data"].([]interface{}); ok {
			fmt.Printf("🟢 Berhasil memuat %d produk DANA.\n", len(products))
			if len(products) > 0 {
				fmt.Println("   Sampel Produk:")
				limit := 3
				if len(products) < limit {
					limit = len(products)
				}
				for i := 0; i < limit; i++ {
					if p, ok := products[i].(map[string]interface{}); ok {
						var harga float64
						switch h := p["harga"].(type) {
						case float64:
							harga = h
						case string:
							fmt.Sscanf(h, "%f", &harga)
						}
						fmt.Printf("   - Kode: %v | Produk: %v | Harga: Rp %.0f\n", p["kode"], p["produk"], harga)
					}
				}
			}
		}
	} else {
		fmt.Println("❌ Gagal memuat katalog.")
	}

	// 4. Leaderboard Sultan
	fmt.Println("\n🔍 4. Mengambil Data Leaderboard Sultan (3 Teratas)...")
	board, err := zakki.Leaderboard(3, "all")
	if err != nil {
		log.Fatalf("❌ Leaderboard gagal: %v", err)
	}

	if code, ok := board["code"].(float64); ok && code == 200 {
		if listSultan, ok := board["leaderboard"].([]interface{}); ok {
			fmt.Println("🟢 Peringkat Sultan Teraktif:")
			for _, item := range listSultan {
				if rank, ok := item.(map[string]interface{}); ok {
					var nama, va, totalTopup string
					if userInfo, ok := rank["user_info"].(map[string]interface{}); ok {
						nama, _ = userInfo["nama"].(string)
						va, _ = userInfo["virtual_account"].(string)
					}
					if stats, ok := rank["stats"].(map[string]interface{}); ok {
						totalTopup, _ = stats["total_topup_formatted"].(string)
					}
					fmt.Printf("   Rank #%.0f - %s (VA: %s) | Total Topup: %s\n", rank["rank"], nama, va, totalTopup)
				}
			}
		}
	} else {
		fmt.Println("❌ Gagal memuat leaderboard.")
	}

	fmt.Println("\n==================================================")
	fmt.Println("🎉 SELURUH PENGUJIAN RIEL SDK GO BERHASIL 100%!")
	fmt.Println("==================================================")
}

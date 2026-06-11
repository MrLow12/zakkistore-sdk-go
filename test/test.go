package main

import (
	"fmt"
	"os"
	"reflect"

	zakkistore "github.com/MrLow12/zakkistore-sdk-go"
)

func main() {
	fmt.Println("🧪 Menjalankan uji coba inisialisasi SDK Go...")

	// 1. Inisialisasi Mock dengan config lengkap
	zakki, err := zakkistore.NewWithConfig(zakkistore.Config{
		Token:        "mock_token_123",
		IDUser:       "mock_user_IBO99",
		PIN:          "123456",
		AutoWithdraw: false,
	})

	if err != nil {
		fmt.Printf("❌ Inisialisasi gagal: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Inisialisasi ZakkiStore Client berhasil!")

	// 2. Verifikasi tipe & metode menggunakan reflect
	t := reflect.TypeOf(zakki)
	fmt.Printf("\n🔍 Memverifikasi eksistensi 25 Native Methods (Total Terdeteksi: %d)...\n", t.NumMethod())

	expectedMethods := []string{
		"Topup", "Cektopup", "Cancel",
		"Listkode", "H2H", "H2HSimple", "Cekh2h", "Myh2h",
		"Checkbank", "Checkname", "Transfer", "TransferSimple", "Tabung", "Tarik", "Checkmutasi",
		"NoktelStok", "NoktelBuy", "NoktelGetOtp", "NoktelCancel", "NoktelHistory",
		"Cekmining", "Mymining", "Cekgacha",
		"Whitelistip", "Delwhitelistip", "Leaderboard", "Status",
	}

	for _, method := range expectedMethods {
		_, found := t.MethodByName(method)
		if found {
			fmt.Printf("  [OK] Metode '%s' terdeteksi dan aktif.\n", method)
		} else {
			fmt.Printf("  [FAIL] Metode '%s' TIDAK TERDETEKSI!\n", method)
			os.Exit(1)
		}
	}

	fmt.Println("\n🏆 Uji Coba Kepatuhan Metode & Struktur Go Sukses 100%!")
}

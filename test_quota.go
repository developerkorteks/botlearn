//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/nabilulilalbab/promote/services"
)

func main() {
	phone := "087817739901"
	if len(os.Args) > 1 {
		phone = os.Args[1]
	}

	fmt.Printf("🔍 Testing .checkkuota untuk nomor: %s\n", phone)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	qc := services.NewQuotaChecker()

	// Test normalisasi dulu
	normalized := qc.NormalizePhoneNumber(phone)
	fmt.Printf("📞 Normalisasi: %s → %s\n\n", phone, normalized)

	// Test CheckQuota
	fmt.Println("📡 Memanggil API xl-ku.my.id...")
	result, err := qc.CheckQuota(phone)
	if err != nil {
		fmt.Printf("❌ ERROR: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ BERHASIL! Hasil:")
	fmt.Println(result)
}

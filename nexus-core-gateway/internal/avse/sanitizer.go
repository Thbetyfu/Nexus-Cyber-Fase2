package avse

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"strings"
)

// SanitizeResult berisi data gambar yang sudah bersih dan metadatanya
type SanitizeResult struct {
	Data         []byte
	Format       string
	OriginalSize int
	CleanedSize  int
	RiskScore    int // 0-100 (Skor kecurigaan steganografi)
}

// AnalyzeRisk mendeduksi tingkat risiko berdasarkan rasio ukuran file vs resolusi
func AnalyzeRisk(input []byte) int {
	config, _, err := image.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return 50 // Default jika gagal baca
	}

	// Hitung rasio byte per pixel
	// Normal JPEG biasanya 0.1 - 0.5 bytes per pixel
	totalPixels := config.Width * config.Height
	if totalPixels == 0 {
		return 0
	}
	
	bytesPerPixel := float64(len(input)) / float64(totalPixels)
	
	// Jika rasio > 2.0, ada kemungkinan besar data tambahan disisipkan (High Entropy)
	if bytesPerPixel > 2.0 {
		return 90 // Sangat Mencurigakan
	} else if bytesPerPixel > 1.0 {
		return 60 // Cukup Mencurigakan
	}
	
	return 10 // Normal
}

// SanitizeImage melakukan pembersihan struktural (Fase 1 & 2)
func SanitizeImage(input []byte) (*SanitizeResult, error) {
	originalSize := len(input)

	// 1. Decode Config untuk cek resolusi (Tanpa load seluruh pixel ke RAM dulu)
	config, format, err := image.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image config: %v", err)
	}

	// [PHASE 4: ANTI-IMAGE BOMB]
	// Batasi resolusi maksimal (misal 4000x4000 = 16MP)
	if config.Width > 4000 || config.Height > 4000 {
		return nil, fmt.Errorf("image resolution too high (%dx%d), potential image-bomb detected", config.Width, config.Height)
	}

	// 2. Decode gambar (Hanya mengambil data pixel murni)
	img, _, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// 2. Persiapkan buffer untuk menulis ulang gambar
	var buf bytes.Buffer

	// 3. Re-Encode (Menulis ulang wadah file dari nol)
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		// Gunakan kualitas tinggi (95) agar tulisan tetap tajam
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95})
	case "png":
		// PNG bersifat lossless, tulisan akan tetap 100% tajam
		err = png.Encode(&buf, img)
	default:
		// Jika format lain, kita coba encode ke PNG demi keamanan maksimal
		err = png.Encode(&buf, img)
		format = "png (converted)"
	}

	if err != nil {
		return nil, fmt.Errorf("failed to re-encode image: %v", err)
	}

	cleanedData := buf.Bytes()
	risk := AnalyzeRisk(input)
	
	return &SanitizeResult{
		Data:         cleanedData,
		Format:       format,
		OriginalSize: originalSize,
		CleanedSize:  len(cleanedData),
		RiskScore:    risk,
	}, nil
}

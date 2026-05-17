package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-cyber/nexus-core-gateway/internal/avse"
	"github.com/nexus-cyber/nexus-core-gateway/internal/database"
	"github.com/nexus-cyber/nexus-core-gateway/internal/models"
	"github.com/nexus-cyber/nexus-core-gateway/internal/proxy"
	"github.com/nexus-cyber/nexus-core-gateway/pkg/logger"
)

// In-memory brute force tracker (Thread-Safe)
var failedAttempts sync.Map

func getCleanIP(remoteAddr string) string {
	if idx := strings.Index(remoteAddr, ":"); idx != -1 {
		return remoteAddr[:idx]
	}
	return remoteAddr
}

// uploadShieldHandler intercepts image uploads, inspects magic bytes,
// sanitizes steganography/EXIF data via AVSE, and forwards the clean file.
func uploadShieldHandler(px *proxy.NexusProxy, telemetry *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 1. Cek ukuran tubuh request (Maksimal 2.5MB untuk overhead form)
		r.Body = http.MaxBytesReader(w, r.Body, 25*1024*100)
		if err := r.ParseMultipartForm(2 << 20); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"status":"error","message":"File too large (Max 2MB)"}`))
			return
		}

		file, header, err := r.FormFile("photo")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"status":"error","message":"Missing photo parameter"}`))
			return
		}
		defer file.Close()

		// 2. Baca 512 byte pertama untuk inspeksi Magic Bytes (Mime Type Asli)
		buff := make([]byte, 512)
		n, err := file.Read(buff)
		if err != nil && err != io.EOF {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status":"error","message":"Failed to read file header"}`))
			return
		}
		// Reset cursor file
		_, _ = file.Seek(0, io.SeekStart)

		contentType := http.DetectContentType(buff[:n])

		// Hanya izinkan image/jpeg, image/png, image/gif, image/webp
		isAllowed := false
		allowedTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
		for _, t := range allowedTypes {
			if contentType == t {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			// LOG SERANGAN KE DATABASE (GORM)
			ip := getCleanIP(r.RemoteAddr)
			tLog := models.ThreatLog{
				Base:          models.Base{ID: uuid.New()},
				SourceIP:      r.RemoteAddr,
				Endpoint:      r.URL.Path,
				Method:        r.Method,
				Status:        "BLOCKED",
				ThreatType:    "MALICIOUS_FILE_UPLOAD_ATTACK",
				Severity:      5,
				PayloadSample: fmt.Sprintf("Filename: %s | Real MimeType: %s", header.Filename, contentType),
				UserAgent:     r.UserAgent(),
				LatencyMs:     1,
			}
			if database.DB != nil {
				database.DB.Create(&tLog)
			}

			// LOG EVENT KE SOC COMMAND CENTER
			telemetry.LogAIEvent(logger.AIEventLog{
				Timestamp:    time.Now(),
				Layer:        "Reflex-Upload-Guard",
				Status:       "ATTACK_BLOCKED",
				DetailAction: fmt.Sprintf("[FILE EXPLOIT] Blocked malicious upload from %s. File: %s. Real Mime: %s", ip, header.Filename, contentType),
			})

			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"status":"error","message":"Security Violation: Malicious file signature detected."}`))
			return
		}

		// 3. Baca seluruh isi file untuk sanitasi
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status":"error","message":"Failed to read file data"}`))
			return
		}

		// 4. Bersihkan gambar secara aktif menggunakan AVSE (Steganography & EXIF strip)
		cleanResult, err := avse.SanitizeImage(fileBytes)
		if err != nil {
			// Jika pembersihan gagal (misalnya gambar rusak/bomb)
			telemetry.LogAIEvent(logger.AIEventLog{
				Timestamp:    time.Now(),
				Layer:        "AVSE (Visual Shield)",
				Status:       "BLOCKED",
				DetailAction: fmt.Sprintf("Suspicious image validation failed from %s: %v", r.RemoteAddr, err),
			})
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"status":"error","message":"Image sanitization failed. File rejected."}`))
			return
		}

		// Kirim event keberhasilan sanitasi ke SOC
		telemetry.LogAIEvent(logger.AIEventLog{
			Timestamp:    time.Now(),
			Layer:        "AVSE (Visual Shield)",
			Status:       "SANITIZED",
			DetailAction: fmt.Sprintf("Visual Clean [%d%% Risk]: %d B -> %d B (%s)", cleanResult.RiskScore, cleanResult.OriginalSize, cleanResult.CleanedSize, cleanResult.Format),
		})

		// 5. Re-packing berkas yang sudah bersih dan kirim ke Portfolio backend
		targetHost := os.Getenv("TARGET_BACKEND")
		if targetHost == "" {
			targetHost = "http://portfolio:80"
		}
		targetURL := fmt.Sprintf("%s/api/upload", targetHost)

		bodyBuf := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuf)
		fileWriter, err := bodyWriter.CreateFormFile("photo", header.Filename)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status":"error","message":"Failed to prepare upload payload"}`))
			return
		}
		if _, err := fileWriter.Write(cleanResult.Data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status":"error","message":"Failed to write clean data"}`))
			return
		}
		bodyWriter.Close()

		targetReq, err := http.NewRequest(http.MethodPost, targetURL, bodyBuf)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status":"error","message":"Failed to create proxy upload request"}`))
			return
		}
		targetReq.Header.Set("Content-Type", bodyWriter.FormDataContentType())

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(targetReq)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`{"status":"error","message":"Failed to reach backend portfolio server"}`))
			return
		}
		defer resp.Body.Close()

		// Kembalikan respon dari portfolio ke client
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

// rewardUnlockHandler verifies password and returns reward link,
// blocks attackers with Autoban if they brute force.
func rewardUnlockHandler(telemetry *logger.Logger) http.HandlerFunc {
	type RequestData struct {
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"status":"error","message":"Method not allowed"}`))
			return
		}

		ip := getCleanIP(r.RemoteAddr)

		// 1. Cek apakah IP sudah ter-blacklist di database (DINONAKTIFKAN SEMENTARA SEBAGAI KOMENTAR)
		/*
		if database.IsIPBlacklisted(r.RemoteAddr) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"status":"error","message":"BANNED: Your IP is in the persistent blacklist due to multiple failed attempts."}`))
			return
		}
		*/

		var req RequestData
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"status":"error","message":"Invalid body"}`))
			return
		}
		_ = json.Unmarshal(bodyBytes, &req)

		// Ambil konfigurasi password & link hadiah
		correctPassword := os.Getenv("REWARD_PASSWORD")
		if correctPassword == "" {
			correctPassword = "nexus-cyber-secret" // Default
		}

		rewardLink := os.Getenv("REWARD_LINK")
		if rewardLink == "" {
			rewardLink = "https://shopee.co.id/m/shopee-kaget-nexus-success" // Default
		}

		// 2. Verifikasi Password
		if req.Password == correctPassword {
			// Sukses: Reset counter percobaan salah untuk IP ini
			failedAttempts.Delete(ip)

			telemetry.LogAIEvent(logger.AIEventLog{
				Timestamp:    time.Now(),
				Layer:        "Reward-Verifier",
				Status:       "SUCCESS",
				DetailAction: fmt.Sprintf("IP %s successfully unlocked the reward link using the correct access code.", ip),
			})

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "success",
				"link":    rewardLink,
			})
			return
		}

		// 3. Gagal: Naikkan counter percobaan salah
		attempts := 1
		if val, ok := failedAttempts.Load(ip); ok {
			attempts = val.(int) + 1
		}
		failedAttempts.Store(ip, attempts)

		// Catat ke threat_logs
		tLog := models.ThreatLog{
			Base:          models.Base{ID: uuid.New()},
			SourceIP:      r.RemoteAddr,
			Endpoint:      r.URL.Path,
			Method:        r.Method,
			Status:        "UNAUTHORIZED",
			ThreatType:    "REWARD_PASSWORD_BRUTE_FORCE",
			Severity:      3,
			PayloadSample: fmt.Sprintf("Attempted Password: '%s' | Fail Count: %d/5", req.Password, attempts),
			UserAgent:     r.UserAgent(),
			LatencyMs:     1,
		}
		if database.DB != nil {
			database.DB.Create(&tLog)
		}

		telemetry.LogAIEvent(logger.AIEventLog{
			Timestamp:    time.Now(),
			Layer:        "Reward-Verifier",
			Status:       "FAILED_ATTEMPT",
			DetailAction: fmt.Sprintf("Failed unlock attempt from %s. Tried: '%s' (%d/5 attempts)", ip, req.Password, attempts),
		})

		// 4. Trigger AUTOBAN jika gagal >= 5 kali (DINONAKTIFKAN SEMENTARA SEBAGAI KOMENTAR)
		/*
		if attempts >= 5 {
			// Simpan ban ke database
			blacklist := models.IntelBlacklist{
				Base:      models.Base{ID: uuid.New()},
				IPAddress: ip,
				Reason:    "Autoban: Exceeded maximum failed reward password attempts (5/5)",
				IsActive:  true,
			}
			if database.DB != nil {
				database.DB.Create(&blacklist)
			}

			telemetry.LogAIEvent(logger.AIEventLog{
				Timestamp:    time.Now(),
				Layer:        "Intel-Shield-Autoban",
				Status:       "IP_BANNED",
				DetailAction: fmt.Sprintf("[AUTOBAN] IP %s has been permanently banned after 5 failed attempts.", ip),
			})

			failedAttempts.Delete(ip)

			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"status":"error","message":"BANNED: Too many failed attempts. Your IP has been permanently blacklisted."}`))
			return
		}
		*/

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(fmt.Sprintf(`{"status":"error","message":"Incorrect Password. Attempt %d of 5"}`, attempts)))
	}
}

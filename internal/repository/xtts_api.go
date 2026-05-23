package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"vox-clone/internal/domain"
)

type xttsService struct {
	xttsURL   string // URL untuk /tts_to_audio/ → port 5002
	uploadURL string // URL untuk /upload_speaker/ → port 5003
	client    *http.Client
}

// NewXTTSService mendaftarkan client XTTS v2
// Sekarang butuh 2 URL: xttsURL (port 5002) dan uploadURL (port 5003)
func NewXTTSService(xttsURL, uploadURL string) domain.VoiceAIService {
	return &xttsService{
		xttsURL:   xttsURL,
		uploadURL: uploadURL,
		client:    &http.Client{Timeout: 120 * time.Second},
	}
}

// CloneVoice menyimpan file lokal DAN mengupload ke server Colab
func (s *xttsService) CloneVoice(ctx context.Context, name string, fileReader io.Reader, fileName string) (string, error) {
	// Step 1: Simpan file lokal dulu
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("gagal membuat folder uploads: %w", err)
	}

	uniqueFileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileName)
	localPath := filepath.Join(uploadDir, uniqueFileName)

	out, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file audio lokal: %w", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, fileReader); err != nil {
		return "", fmt.Errorf("gagal menyimpan data biner audio: %w", err)
	}

	// Step 2: Upload file ke server Colab (port 5003)
	uploadedName, err := s.uploadSpeakerToColab(ctx, localPath, uniqueFileName)
	if err != nil {
		return "", fmt.Errorf("gagal upload ke Colab: %w", err)
	}

	fmt.Printf("[XTTS-DEBUG] Speaker berhasil diupload: %s\n", uploadedName)

	// Kembalikan nama file (bukan path) sebagai voiceID yang disimpan di DB
	return uploadedName, nil
}

// uploadSpeakerToColab mengirim file audio ke mini upload server di Colab
func (s *xttsService) uploadSpeakerToColab(ctx context.Context, localPath, fileName string) (string, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuka file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(part, file); err != nil {
		return "", err
	}
	writer.Close()

	uploadURL := fmt.Sprintf("%s/upload_speaker/", s.uploadURL)
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("bypass-tunnel-reminder", "true")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal upload: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("[XTTS-DEBUG] Upload status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload error %d: %s", resp.StatusCode, string(resBody))
	}

	// Parse response untuk ambil filename
	var result struct {
		SpeakerPath string `json:"speaker_path"`
		Filename    string `json:"filename"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("gagal parse response upload: %w", err)
	}

	return result.SpeakerPath, nil
}

// GenerateTTS mengirim teks ke XTTS dan menerima audio
func (s *xttsService) GenerateTTS(ctx context.Context, voiceID, text string) (io.ReadCloser, error) {
	// voiceID sekarang = nama file di server Colab (bukan path lokal)
	jsonBody, _ := json.Marshal(map[string]string{
		"text":        text,
		"speaker_wav": voiceID,
		"language":    "en",
	})

	ttsURL := fmt.Sprintf("%s/tts_to_audio/", s.xttsURL)
	req, err := http.NewRequestWithContext(ctx, "POST", ttsURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("bypass-tunnel-reminder", "true")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal request TTS: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("[XTTS-DEBUG] TTS Status: %d\n", resp.StatusCode)
		fmt.Printf("[XTTS-DEBUG] TTS Error: %s\n", string(resBody))
		return nil, fmt.Errorf("XTTS error %d: %s", resp.StatusCode, string(resBody))
	}

	buf := new(bytes.Buffer)
	_, _ = io.Copy(buf, resp.Body)
	resp.Body.Close()
	return io.NopCloser(buf), nil
}

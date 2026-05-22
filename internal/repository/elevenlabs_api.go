package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"vox-clone/internal/domain"
)

type elevenLabsService struct {
	apiKey string
	client *http.Client
}

// NewElevenLabsService adalah constructor untuk layanan integrasi Cloud AI
func NewElevenLabsService(apiKey string) domain.ElevenLabsService {
	return &elevenLabsService{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (s *elevenLabsService) CloneVoice(ctx context.Context, name string, fileReader io.Reader, fileName string) (string, error) {
	url := "https://api.elevenlabs.io/v1/voices/add"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Set form field untuk nama voice
	_ = writer.WriteField("name", name)

	// Salin data binary audio ke dalam form multipart
	part, err := writer.CreateFormFile("files", fileName)
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(part, fileReader); err != nil {
		return "", err
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("xi-api-key", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("elevenlabs API error: %s", string(resBody))
	}

	var data struct {
		VoiceID string `json:"voice_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data.VoiceID, nil
}

func (s *elevenLabsService) GenerateTTS(ctx context.Context, voiceID, text string) (io.ReadCloser, error) {
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	payload := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_flash_v2_5", // Model andalan untuk kecepatan tinggi dan mendukung Bahasa Indonesia
	}
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("elevenlabs TTS error: %s", string(resBody))
	}

	// Mengembalikan response body berupa binary audio stream (.mp3) tanpa menutupnya dahulu
	return resp.Body, nil
}

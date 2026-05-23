package http

import (
	"io"
	"net/http"
	"vox-clone/internal/domain"

	"github.com/gin-gonic/gin"
)

type VoiceHandler struct {
	usecase domain.VoiceUsecase
}

// NewVoiceHandler mendaftarkan semua endpoint HTTP ke router Gin
func NewVoiceHandler(r *gin.Engine, u domain.VoiceUsecase) {
	handler := &VoiceHandler{usecase: u}

	// Grouping API untuk kerapihan endpoint
	api := r.Group("/api")
	{
		api.POST("/voices/clone", handler.CloneVoice)
		api.POST("/voices/tts", handler.GenerateTTS)
		api.GET("/voices", handler.GetAllVoices)
	}
}

// CloneVoice menangani request upload file / rekaman audio langsung dari dashboard
func (h *VoiceHandler) CloneVoice(c *gin.Context) {
	name := c.PostForm("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parameter 'name' tidak boleh kosong"})
		return
	}

	// Membaca file audio dari multipart form data (key: "audio")
	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file audio wajib disertakan dengan key 'audio'"})
		return
	}
	defer file.Close()

	// Meneruskan data ke layer usecase
	result, err := h.usecase.RegisterNewVoice(c.Request.Context(), name, file, header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Kloning suara berhasil didaftarkan!",
		"data":    result,
	})
}

// GenerateTTS menangani konversi teks menjadi audio dengan id suara tertentu
func (h *VoiceHandler) GenerateTTS(c *gin.Context) {
	var req struct {
		VoiceID string `json:"voice_id" binding:"required"`
		Text    string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data request tidak valid atau kurang"})
		return
	}

	audioStream, err := h.usecase.TextToSpeech(c.Request.Context(), req.VoiceID, req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer audioStream.Close()

	// Mengirimkan audio biner secara langsung (streaming) balik ke browser dashboard
	c.Header("Content-Type", "audio/mpeg")
	c.Header("Content-Disposition", "attachment; filename=result.mp3")

	if _, err := io.Copy(c.Writer, audioStream); err != nil {
		// Log error jika stream terputus di tengah jalan
		return
	}
}

// GetAllVoices mengambil daftar semua suara yang sudah pernah dikloning
func (h *VoiceHandler) GetAllVoices(c *gin.Context) {
	voices, err := h.usecase.GetAllVoices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Daftar suara berhasil diambil!",
		"data":    voices,
	})
}

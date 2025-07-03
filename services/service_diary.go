package services

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"web-diary-be/config"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// AnalyzeEmotion mengambil teks dan mengembalikan analisis emosi dan sentimen
func AnalyzeEmotion(text string) (string, string, error) {
	ctx := context.Background()
	
	if config.GeminiFlashAPIKey == "" {
		log.Println("Gemini Flash API Key is empty, aborting Gemini Flash analysis.")
		return "Unknown", "Neutral", nil
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(config.GeminiFlashAPIKey))
	if err != nil {
		log.Printf("Failed to create Gemini Flash client: %v", err)
		return "Unknown", "Neutral", err
	}
	defer client.Close()

	// --- PASTIKAN NAMA MODEL DI SINI SESUAI DENGAN YANG ANDA MAKSUD ---
	// Jika Anda ingin menggunakan Gemini Flash 1.5 Pro:
	model := client.GenerativeModel("gemini-2.0-flash")
	// Atau jika Anda ingin versi spesifik tanpa "latest"
	// model := client.GenerativeModel("gemini-flash-1.5-pro")

	// Jika Anda ingin menggunakan Gemini Flash 1.0 Pro (model yang lebih lama):
	// model := client.GenerativeModel("gemini-flash-pro") 
	// --- AKHIR PENTING ---

	prompt := `Analyze the following text for its dominant emotion and overall sentiment (positive, negative, neutral).
	Return the result in a JSON object with 'emotion' and 'sentiment' keys.
	Example: {"emotion": "joy", "sentiment": "positive"}
	Text: "` + text + `"`

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Error generating content from Gemini Flash: %v", err)
		return "Unknown", "Neutral", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		log.Println("No content returned from Gemini Flash. Defaulting to Neutral.")
		return "Neutral", "Neutral", nil
	}

	result := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			result += string(txt)
		}
	}

	var analysis struct {
		Emotion   string `json:"emotion"`
		Sentiment string `json:"sentiment"`
	}

	// Bersihkan result dari backtick dan blok markdown jika ada
	cleanResult := result
	if strings.HasPrefix(cleanResult, "```json") {
		cleanResult = strings.TrimPrefix(cleanResult, "```json")
	}
	if strings.HasPrefix(cleanResult, "```") {
		cleanResult = strings.TrimPrefix(cleanResult, "```")
	}
	cleanResult = strings.TrimSpace(cleanResult)
	if strings.HasSuffix(cleanResult, "```") {
		cleanResult = strings.TrimSuffix(cleanResult, "```")
	}

	err = json.Unmarshal([]byte(cleanResult), &analysis)
	if err != nil {
		log.Printf("Error unmarshalling Gemini Flash response: %v, Raw Response: %s", err, result)
		return "Unknown", "Neutral", nil
	}

	return analysis.Emotion, analysis.Sentiment, nil
}
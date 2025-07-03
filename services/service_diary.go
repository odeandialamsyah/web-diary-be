package services

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"web-diary-be/config"
)

// AnalyzeEmotion mengambil teks dan mengembalikan analisis emosi dan sentimen
func AnalyzeEmotion(text string) (string, string, error) {
	ctx := context.Background()
	
	if config.GeminiAPIKey == "" {
		log.Println("Gemini API Key is empty, aborting Gemini analysis.")
		return "Unknown", "Neutral", nil
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(config.GeminiAPIKey))
	if err != nil {
		log.Printf("Failed to create Gemini client: %v", err)
		return "Unknown", "Neutral", err
	}
	defer client.Close()

	// --- PASTIKAN NAMA MODEL DI SINI SESUAI DENGAN YANG ANDA MAKSUD ---
	// Jika Anda ingin menggunakan Gemini 1.5 Pro:
	model := client.GenerativeModel("gemini-1.5-pro-latest")
	// Atau jika Anda ingin versi spesifik tanpa "latest"
	// model := client.GenerativeModel("gemini-1.5-pro")

	// Jika Anda ingin menggunakan Gemini 1.0 Pro (model yang lebih lama):
	// model := client.GenerativeModel("gemini-pro") 
	// --- AKHIR PENTING ---

	prompt := `Analyze the following text for its dominant emotion and overall sentiment (positive, negative, neutral).
	Return the result in a JSON object with 'emotion' and 'sentiment' keys.
	Example: {"emotion": "joy", "sentiment": "positive"}
	Text: "` + text + `"`

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Error generating content from Gemini: %v", err)
		return "Unknown", "Neutral", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		log.Println("No content returned from Gemini. Defaulting to Neutral.")
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

	err = json.Unmarshal([]byte(result), &analysis)
	if err != nil {
		log.Printf("Error unmarshalling Gemini response: %v, Raw Response: %s", err, result)
		return "Unknown", "Neutral", nil
	}

	return analysis.Emotion, analysis.Sentiment, nil
}
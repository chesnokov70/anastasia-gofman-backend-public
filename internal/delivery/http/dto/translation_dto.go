package dto

type TranslationRequest struct {
	Text      string   `json:"text" binding:"required" example:"Hello, world!"`
	Languages []string `json:"languages" binding:"required" example:"ru,es,fr"`
}

type TranslationResponse map[string]string

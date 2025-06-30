package service

import (
	"anastasia_gofman_backend/internal/delivery/http/dto"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type translationService struct {
	apiKey string
}

type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

func NewTranslationService(apiKey string) TranslationService {
	return &translationService{
		apiKey: apiKey,
	}
}

func (s *translationService) TranslateText(text string, languages []string) (map[string]string, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is not configured")
	}

	systemPrompt := fmt.Sprintf(`You are a professional translator. Your task is to translate the given text into the specified languages. 

IMPORTANT INSTRUCTIONS:
1. Respond ONLY with a valid JSON object
2. The JSON keys should be the language codes provided
3. The JSON values should be the translated text
4. Do not include any markdown, explanations, or additional text
5. Do not use code blocks or formatting
6. Return only the raw JSON object

Target languages: %s

Example format: {"ru": "Привет мир", "es": "Hola mundo", "fr": "Bonjour le monde"}`, strings.Join(languages, ", "))

	userPrompt := fmt.Sprintf("Translate this text: %s", text)

	request := OpenAIRequest{
		Model:       "gpt-4.1",
		Temperature: 0.7,
		MaxTokens:   10000,
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
	}

	response, err := s.callOpenAI(request)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	content := strings.TrimSpace(response.Choices[0].Message.Content)

	var translations map[string]string
	if err := json.Unmarshal([]byte(content), &translations); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response as JSON: %w, content: %s", err, content)
	}

	for _, lang := range languages {
		if _, exists := translations[lang]; !exists {
			return nil, fmt.Errorf("translation for language '%s' not found in response", lang)
		}
	}

	return translations, nil
}

func (s *translationService) callOpenAI(request OpenAIRequest) (*OpenAIResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

func (s *translationService) AutoCompleteTranslation(existingText map[string]string, supportedLanguages []string, maxRetries int) (map[string]string, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is not configured")
	}

	result := make(map[string]string)

	for lang, text := range existingText {
		if strings.TrimSpace(text) != "" {
			result[lang] = text
		}
	}

	var missingLanguages []string
	var sourceText string

	for _, lang := range supportedLanguages {
		if text, exists := result[lang]; exists && strings.TrimSpace(text) != "" {
			if sourceText == "" {
				sourceText = text
			}
		} else {
			missingLanguages = append(missingLanguages, lang)
		}
	}

	if sourceText == "" {
		return nil, fmt.Errorf("no source text found in any supported language")
	}

	if len(missingLanguages) == 0 {
		return result, nil
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		translations, err := s.TranslateText(sourceText, missingLanguages)
		if err != nil {
			lastErr = err
			fmt.Printf("Translation attempt %d failed: %v\n", attempt+1, err)
			if attempt < maxRetries-1 {
				time.Sleep(time.Duration(attempt+1) * time.Second)
			}
			continue
		}

		for lang, translation := range translations {
			result[lang] = translation
		}
		return result, nil
	}

	return nil, fmt.Errorf("translation failed after %d attempts, last error: %w", maxRetries, lastErr)
}

func (s *translationService) AutoCompleteTranslatedTextDTO(textDTO dto.TranslatedTextDTO, maxRetries int) (dto.TranslatedTextDTO, error) {
	existingText := map[string]string{
		"en": textDTO.EN,
		"ru": textDTO.RU,
		"es": textDTO.ES,
	}

	supportedLanguages := []string{"en", "ru", "es"}

	completedText, err := s.AutoCompleteTranslation(existingText, supportedLanguages, maxRetries)
	if err != nil {
		return textDTO, err
	}

	return dto.TranslatedTextDTO{
		EN: completedText["en"],
		RU: completedText["ru"],
		ES: completedText["es"],
	}, nil
}

func (s *translationService) AutoCompleteEventTranslations(eventDTO *dto.CreateEventDTO, maxRetries int) error {
	var err error

	if s.hasAnyText(eventDTO.Title) {
		eventDTO.Title, err = s.AutoCompleteTranslatedTextDTO(eventDTO.Title, maxRetries)
		if err != nil {
			return fmt.Errorf("failed to translate title: %w", err)
		}
	}

	if s.hasAnyText(eventDTO.Description) {
		eventDTO.Description, err = s.AutoCompleteTranslatedTextDTO(eventDTO.Description, maxRetries)
		if err != nil {
			return fmt.Errorf("failed to translate description: %w", err)
		}
	}

	if s.hasAnyText(eventDTO.Location) {
		eventDTO.Location, err = s.AutoCompleteTranslatedTextDTO(eventDTO.Location, maxRetries)
		if err != nil {
			return fmt.Errorf("failed to translate location: %w", err)
		}
	}

	if s.hasAnyText(eventDTO.Language) {
		eventDTO.Language, err = s.AutoCompleteTranslatedTextDTO(eventDTO.Language, maxRetries)
		if err != nil {
			return fmt.Errorf("failed to translate language: %w", err)
		}
	}

	if s.hasAnyText(eventDTO.Format) {
		eventDTO.Format, err = s.AutoCompleteTranslatedTextDTO(eventDTO.Format, maxRetries)
		if err != nil {
			return fmt.Errorf("failed to translate format: %w", err)
		}
	}

	if s.hasAnyText(eventDTO.Venue) {
		eventDTO.Venue, err = s.AutoCompleteTranslatedTextDTO(eventDTO.Venue, maxRetries)
		if err != nil {
			return fmt.Errorf("failed to translate venue: %w", err)
		}
	}

	if s.hasAnyText(eventDTO.Organizer) {
		eventDTO.Organizer, err = s.AutoCompleteTranslatedTextDTO(eventDTO.Organizer, maxRetries)
		if err != nil {
			return fmt.Errorf("failed to translate organizer: %w", err)
		}
	}

	return nil
}

func (s *translationService) AutoCompleteEventWithPhotosTranslations(eventDTO *dto.CreateEventWithPhotosDTO, maxRetries int) error {
	var err error

	if s.hasAnyText(eventDTO.Title) {
		eventDTO.Title, err = s.AutoCompleteTranslatedTextDTO(eventDTO.Title, maxRetries)
		if err != nil {
			return fmt.Errorf("failed to translate title: %w", err)
		}
	}

	if s.hasAnyText(eventDTO.Description) {
		eventDTO.Description, err = s.AutoCompleteTranslatedTextDTO(eventDTO.Description, maxRetries)
		if err != nil {
			return fmt.Errorf("failed to translate description: %w", err)
		}
	}

	if s.hasAnyText(eventDTO.Location) {
		eventDTO.Location, err = s.AutoCompleteTranslatedTextDTO(eventDTO.Location, maxRetries)
		if err != nil {
			return fmt.Errorf("failed to translate location: %w", err)
		}
	}

	if s.hasAnyText(eventDTO.Language) {
		eventDTO.Language, err = s.AutoCompleteTranslatedTextDTO(eventDTO.Language, maxRetries)
		if err != nil {
			return fmt.Errorf("failed to translate language: %w", err)
		}
	}

	if s.hasAnyText(eventDTO.Format) {
		eventDTO.Format, err = s.AutoCompleteTranslatedTextDTO(eventDTO.Format, maxRetries)
		if err != nil {
			return fmt.Errorf("failed to translate format: %w", err)
		}
	}

	if s.hasAnyText(eventDTO.Venue) {
		eventDTO.Venue, err = s.AutoCompleteTranslatedTextDTO(eventDTO.Venue, maxRetries)
		if err != nil {
			return fmt.Errorf("failed to translate venue: %w", err)
		}
	}

	if s.hasAnyText(eventDTO.Organizer) {
		eventDTO.Organizer, err = s.AutoCompleteTranslatedTextDTO(eventDTO.Organizer, maxRetries)
		if err != nil {
			return fmt.Errorf("failed to translate organizer: %w", err)
		}
	}

	return nil
}

func (s *translationService) hasAnyText(textDTO dto.TranslatedTextDTO) bool {
	return strings.TrimSpace(textDTO.EN) != "" ||
		strings.TrimSpace(textDTO.RU) != "" ||
		strings.TrimSpace(textDTO.ES) != ""
}

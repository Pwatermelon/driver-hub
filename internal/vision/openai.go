package vision

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenAIScanner struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewOpenAIScanner(apiKey, baseURL, model string) *OpenAIScanner {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIScanner{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client:  &http.Client{Timeout: 90 * time.Second},
	}
}

func (o *OpenAIScanner) Mode() string { return "openai" }

const licensePrompt = `Ты извлекаешь данные российского водительского удостоверения с фото лицевой и оборотной стороны.
Верни ТОЛЬКО JSON без markdown со схемой:
{
  "last_name": "",
  "first_name": "",
  "middle_name": "",
  "birth_date": "YYYY-MM-DD",
  "series": "",
  "number": "",
  "issue_date": "YYYY-MM-DD",
  "expiry_date": "YYYY-MM-DD",
  "categories": ["B","C"],
  "issuer": "",
  "confidence": 0.0,
  "notes": ""
}
Правила:
- серии/номера как на документе;
- категории — массив латинских букв (A,B,C,D,CE...);
- даты строго YYYY-MM-DD;
- если поле не читается — пустая строка или [] и снизь confidence;
- не выдумывай данные.`

func (o *OpenAIScanner) ScanLicense(ctx context.Context, front, back *ImageInput) (*LicenseDraft, error) {
	if o.apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY не задан")
	}
	if front == nil || len(front.Data) == 0 {
		return nil, fmt.Errorf("нужно фото лицевой стороны")
	}
	if back == nil || len(back.Data) == 0 {
		return nil, fmt.Errorf("нужно фото оборотной стороны")
	}

	content := []any{
		map[string]any{"type": "text", "text": licensePrompt},
		imagePart(front),
		map[string]any{"type": "text", "text": "Выше — лицевая сторона. Ниже — оборот."},
		imagePart(back),
	}

	payload := map[string]any{
		"model": o.model,
		"messages": []any{
			map[string]any{
				"role":    "user",
				"content": content,
			},
		},
		"temperature": 0,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("vision api: %s", truncate(string(raw), 400))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("пустой ответ vision api")
	}

	contentText := strings.TrimSpace(parsed.Choices[0].Message.Content)
	contentText = strings.TrimPrefix(contentText, "```json")
	contentText = strings.TrimPrefix(contentText, "```")
	contentText = strings.TrimSuffix(contentText, "```")
	contentText = strings.TrimSpace(contentText)

	var draft LicenseDraft
	if err := json.Unmarshal([]byte(contentText), &draft); err != nil {
		return nil, fmt.Errorf("разбор JSON: %w", err)
	}
	draft.Source = "openai"
	draft.Categories = normalizeCategories(draft.Categories)
	if draft.Confidence == 0 {
		draft.Confidence = 0.7
	}
	return &draft, nil
}

func imagePart(img *ImageInput) map[string]any {
	ct := img.ContentType
	if ct == "" {
		ct = "image/jpeg"
	}
	b64 := base64.StdEncoding.EncodeToString(img.Data)
	return map[string]any{
		"type": "image_url",
		"image_url": map[string]any{
			"url":    fmt.Sprintf("data:%s;base64,%s", ct, b64),
			"detail": "high",
		},
	}
}

func normalizeCategories(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]bool{}
	for _, c := range in {
		c = strings.ToUpper(strings.TrimSpace(c))
		c = strings.ReplaceAll(c, " ", "")
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		out = append(out, c)
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

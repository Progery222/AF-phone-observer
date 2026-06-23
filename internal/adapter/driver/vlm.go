package driver

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mobilefarm/af/phone-observer/internal/config"
	"github.com/mobilefarm/af/phone-observer/internal/domain"
	"github.com/mobilefarm/af/phone-observer/internal/port"
)

type CascadingScreenAnalyzer struct {
	cfg      config.Config
	client   *http.Client
	log      port.Logger
	backends []string
	sem      chan struct{}
}

func NewCascadingScreenAnalyzer(cfg config.Config, client *http.Client, log port.Logger) *CascadingScreenAnalyzer {
	if client == nil {
		client = http.DefaultClient
	}
	maxConcurrency := cfg.VLMMaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = 2
	}
	return &CascadingScreenAnalyzer{
		cfg:      cfg,
		client:   client,
		log:      log,
		backends: configuredVLMBackends(cfg),
		sem:      make(chan struct{}, maxConcurrency),
	}
}

func (a *CascadingScreenAnalyzer) Configured() bool {
	return len(a.backends) > 0
}

func (a *CascadingScreenAnalyzer) Analyze(ctx context.Context, serial, platform string, screenshot []byte) (domain.VLMAnalysis, error) {
	if len(a.backends) == 0 {
		return domain.VLMAnalysis{}, domain.ErrVLMUnavailable
	}
	if len(screenshot) == 0 {
		return domain.VLMAnalysis{}, fmt.Errorf("%w: пустой screenshot", domain.ErrVLMFailed)
	}

	select {
	case a.sem <- struct{}{}:
		defer func() { <-a.sem }()
	case <-ctx.Done():
		return domain.VLMAnalysis{}, ctx.Err()
	}

	timeout := time.Duration(a.cfg.VLMTimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var lastErr error
	for _, backend := range a.backends {
		var (
			analysis domain.VLMAnalysis
			err      error
		)
		switch backend {
		case "vision_server":
			analysis, err = a.analyzeVisionServer(ctx, serial, platform, screenshot)
		case "ollama":
			analysis, err = a.analyzeOllama(ctx, platform, screenshot)
		case "openai":
			analysis, err = a.analyzeOpenAI(ctx, platform, screenshot)
		default:
			err = fmt.Errorf("%w: неизвестный backend %s", domain.ErrVLMUnavailable, backend)
		}
		if err == nil {
			return analysis, nil
		}
		lastErr = err
		if a.log != nil {
			a.log.Warn("vlm backend failed", "backend", backend, "error", err)
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return domain.VLMAnalysis{}, err
		}
	}
	if lastErr == nil {
		return domain.VLMAnalysis{}, domain.ErrVLMUnavailable
	}
	return domain.VLMAnalysis{}, fmt.Errorf("%w: %w", domain.ErrVLMUnavailable, lastErr)
}

func configuredVLMBackends(cfg config.Config) []string {
	raw := strings.Split(cfg.VLMBackends, ",")
	backends := make([]string, 0, len(raw))
	for _, item := range raw {
		item = strings.TrimSpace(strings.ToLower(item))
		switch item {
		case "":
			continue
		case "visionserver":
			item = "vision_server"
		}
		if item == "vision_server" && strings.TrimSpace(cfg.VisionServerURL) == "" {
			continue
		}
		if item == "openai" && strings.TrimSpace(cfg.OpenAIAPIKey) == "" {
			continue
		}
		backends = append(backends, item)
	}
	return backends
}

func (a *CascadingScreenAnalyzer) analyzeVisionServer(ctx context.Context, serial, platform string, screenshot []byte) (domain.VLMAnalysis, error) {
	if strings.TrimSpace(a.cfg.VisionServerURL) == "" {
		return domain.VLMAnalysis{}, domain.ErrVLMUnavailable
	}
	payload := map[string]any{
		"device_id":        serial,
		"frame_b64":        base64.StdEncoding.EncodeToString(screenshot),
		"platform":         domain.NormalizeDetectionPlatform(platform),
		"require_ocr":      true,
		"require_elements": true,
	}
	body, err := a.postJSON(ctx, strings.TrimRight(a.cfg.VisionServerURL, "/")+"/analyze", payload, nil)
	if err != nil {
		return domain.VLMAnalysis{}, err
	}
	var resp visionServerResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return domain.VLMAnalysis{}, fmt.Errorf("%w: %w", domain.ErrVLMFailed, err)
	}
	texts := append([]string{}, resp.ScreenState.Texts...)
	if resp.OCRText != "" {
		texts = append(texts, resp.OCRText)
	}
	for _, element := range resp.Elements {
		texts = append(texts, firstNonEmpty(element.Label, element.Text, element.ElementType))
	}
	return domain.VLMAnalysis{
		State:          domain.NormalizeVLMState(resp.ScreenState.ScreenType),
		Confidence:     resp.ScreenState.Confidence,
		Description:    firstNonEmpty(resp.OCRText, resp.ScreenState.ScreenType),
		Elements:       uniqueStrings(texts),
		MatchedSignals: []string{"vlm:vision_server", "vlm:" + resp.ScreenState.ScreenType},
		Flags:          flagsFromVLM(resp.ScreenState.ScreenType, strings.Join(texts, " ")),
		BackendUsed:    "vision_server",
		RawResponse:    string(body),
	}, nil
}

func (a *CascadingScreenAnalyzer) analyzeOllama(ctx context.Context, platform string, screenshot []byte) (domain.VLMAnalysis, error) {
	payload := map[string]any{
		"model":  firstNonEmpty(a.cfg.OllamaVLMModel, "qwen2.5vl:7b"),
		"prompt": vlmPrompt(platform),
		"images": []string{base64.StdEncoding.EncodeToString(screenshot)},
		"stream": false,
		"options": map[string]any{
			"temperature": 0.1,
			"num_predict": 800,
		},
	}
	body, err := a.postJSON(ctx, strings.TrimRight(firstNonEmpty(a.cfg.OllamaURL, "http://localhost:11434"), "/")+"/api/generate", payload, nil)
	if err != nil {
		return domain.VLMAnalysis{}, err
	}
	var resp struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return domain.VLMAnalysis{}, fmt.Errorf("%w: %w", domain.ErrVLMFailed, err)
	}
	analysis, err := parseVLMTextResponse(resp.Response, "ollama")
	if err != nil {
		return domain.VLMAnalysis{}, err
	}
	analysis.RawResponse = resp.Response
	return analysis, nil
}

func (a *CascadingScreenAnalyzer) analyzeOpenAI(ctx context.Context, platform string, screenshot []byte) (domain.VLMAnalysis, error) {
	if strings.TrimSpace(a.cfg.OpenAIAPIKey) == "" {
		return domain.VLMAnalysis{}, domain.ErrVLMUnavailable
	}
	imageURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(screenshot)
	payload := map[string]any{
		"model": firstNonEmpty(a.cfg.OpenAIModel, "gpt-5.4-mini"),
		"input": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{"type": "input_text", "text": vlmPrompt(platform)},
					{"type": "input_image", "image_url": imageURL},
				},
			},
		},
		"max_output_tokens": 800,
		"temperature":       0.1,
	}
	headers := map[string]string{"Authorization": "Bearer " + a.cfg.OpenAIAPIKey}
	body, err := a.postJSON(ctx, strings.TrimRight(firstNonEmpty(a.cfg.OpenAIBaseURL, "https://api.openai.com/v1"), "/")+"/responses", payload, headers)
	if err != nil {
		return domain.VLMAnalysis{}, err
	}
	rawText := openAIOutputText(body)
	if rawText == "" {
		return domain.VLMAnalysis{}, fmt.Errorf("%w: пустой ответ OpenAI", domain.ErrVLMFailed)
	}
	analysis, err := parseVLMTextResponse(rawText, "openai")
	if err != nil {
		return domain.VLMAnalysis{}, err
	}
	analysis.RawResponse = rawText
	return analysis, nil
}

func (a *CascadingScreenAnalyzer) postJSON(ctx context.Context, url string, payload any, headers map[string]string) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: %s: %s", domain.ErrVLMUnavailable, resp.Status, strings.TrimSpace(string(body)))
	}
	return body, nil
}

func parseVLMTextResponse(raw, backend string) (domain.VLMAnalysis, error) {
	jsonText := extractJSONObject(raw)
	if jsonText == "" {
		return domain.VLMAnalysis{}, fmt.Errorf("%w: ответ не содержит JSON", domain.ErrVLMFailed)
	}
	var parsed genericVLMResponse
	if err := json.Unmarshal([]byte(jsonText), &parsed); err != nil {
		return domain.VLMAnalysis{}, fmt.Errorf("%w: %w", domain.ErrVLMFailed, err)
	}
	elements := append([]string{}, parsed.VisibleText...)
	for _, element := range parsed.Elements {
		elements = append(elements, firstNonEmpty(element.Text, element.Label, element.Type))
	}
	state := domain.NormalizeVLMState(parsed.ScreenType)
	if state == "" {
		state = domain.ScreenStateUnknown
	}
	return domain.VLMAnalysis{
		State:           state,
		Confidence:      parsed.Confidence,
		Description:     firstNonEmpty(parsed.Description, parsed.CurrentScreen, parsed.ScreenType),
		Elements:        uniqueStrings(elements),
		MatchedSignals:  []string{"vlm:" + backend, "vlm:" + parsed.ScreenType},
		Flags:           domain.DetectionFlags{Captcha: parsed.HasCaptcha, Ban: parsed.HasBan, Error: parsed.HasError},
		SuggestedAction: parsed.SuggestedAction,
		BackendUsed:     backend,
		RawResponse:     raw,
	}, nil
}

func extractJSONObject(raw string) string {
	clean := strings.TrimSpace(raw)
	if strings.HasPrefix(clean, "```") {
		if idx := strings.Index(clean, "\n"); idx >= 0 {
			clean = clean[idx+1:]
		}
		clean = strings.TrimSuffix(clean, "```")
	}
	start := strings.Index(clean, "{")
	end := strings.LastIndex(clean, "}")
	if start < 0 || end <= start {
		return ""
	}
	return clean[start : end+1]
}

func openAIOutputText(body []byte) string {
	var resp struct {
		OutputText string `json:"output_text"`
		Output     []struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return ""
	}
	if strings.TrimSpace(resp.OutputText) != "" {
		return resp.OutputText
	}
	for _, output := range resp.Output {
		for _, content := range output.Content {
			if strings.TrimSpace(content.Text) != "" {
				return content.Text
			}
		}
	}
	return ""
}

func vlmPrompt(platform string) string {
	platform = domain.NormalizeDetectionPlatform(platform)
	return fmt.Sprintf(`Analyze this Android %s screen screenshot. Return only valid JSON:
{
  "screen_type": "feed|profile|search|settings|registration|login|captcha|ban_screen|notification|permission_dialog|loading|install|ad|error|other",
  "confidence": 0.0,
  "description": "brief description",
  "visible_text": ["text visible on screen"],
  "elements": [{"type":"button|input|text|icon","text":"label","bounds":[0,0,0,0]}],
  "has_captcha": false,
  "has_error": false,
  "has_ban": false,
  "suggested_action": "tap|type|wait|scroll|back"
}`, platform)
}

func flagsFromVLM(screenType, text string) domain.DetectionFlags {
	joined := strings.ToLower(screenType + " " + text)
	return domain.DetectionFlags{
		Captcha: strings.Contains(joined, "captcha") || strings.Contains(joined, "verify"),
		Ban:     strings.Contains(joined, "ban") || strings.Contains(joined, "suspend"),
		Error:   strings.Contains(joined, "error") || strings.Contains(joined, "something went wrong"),
	}
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

type visionServerResponse struct {
	ScreenState struct {
		ScreenType string   `json:"screen_type"`
		Confidence float64  `json:"confidence"`
		Texts      []string `json:"texts"`
	} `json:"screen_state"`
	Elements []struct {
		ElementType string  `json:"element_type"`
		Label       string  `json:"label"`
		Text        string  `json:"text"`
		Confidence  float64 `json:"confidence"`
	} `json:"elements"`
	OCRText string `json:"ocr_text"`
}

type genericVLMResponse struct {
	ScreenType      string           `json:"screen_type"`
	Confidence      float64          `json:"confidence"`
	Description     string           `json:"description"`
	CurrentScreen   string           `json:"current_screen"`
	VisibleText     []string         `json:"visible_text"`
	Elements        []genericElement `json:"elements"`
	HasCaptcha      bool             `json:"has_captcha"`
	HasError        bool             `json:"has_error"`
	HasBan          bool             `json:"has_ban"`
	SuggestedAction string           `json:"suggested_action"`
}

type genericElement struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Label string `json:"label"`
}

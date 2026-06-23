package domain

import (
	"strings"
	"time"
)

const (
	DetectionModeAuto = "auto"
	DetectionModeUI   = "ui"
	DetectionModeVLM  = "vlm"

	ScreenStateUnknown           = "unknown"
	ScreenStateLogin             = "login_screen"
	ScreenStateRegistration      = "registration_screen"
	ScreenStatePermissionRequest = "permission_request"
	ScreenStateAdsFullscreen     = "ads_fullscreen"
	ScreenStateInstall           = "install_screen"
	ScreenStateLoading           = "loading"
	ScreenStateMainFeed          = "main_feed"
	ScreenStateNotification      = "notification"
	ScreenStateCaptcha           = "captcha"
	ScreenStateBan               = "ban_screen"
	ScreenStateError             = "error_screen"
	ScreenStateProfile           = "profile_screen"
	ScreenStateSearch            = "search_screen"
	ScreenStateSettings          = "settings_screen"
)

type DetectStateOptions struct {
	Mode            string
	Platform        string
	UseScreenshot   bool
	StoreScreenshot bool
}

type DetectionFlags struct {
	Captcha bool `json:"captcha"`
	Ban     bool `json:"ban"`
	Error   bool `json:"error"`
}

type VLMAnalysis struct {
	State           string
	Confidence      float64
	Description     string
	Elements        []string
	MatchedSignals  []string
	Flags           DetectionFlags
	SuggestedAction string
	BackendUsed     string
	RawResponse     string
}

type ScreenDetection struct {
	Serial          string         `json:"serial,omitempty"`
	State           string         `json:"state"`
	Confidence      float64        `json:"confidence"`
	Source          string         `json:"source"`
	BackendUsed     string         `json:"backend_used"`
	Description     string         `json:"description"`
	Elements        []string       `json:"elements"`
	MatchedSignals  []string       `json:"matched_signals"`
	Flags           DetectionFlags `json:"flags"`
	SuggestedAction string         `json:"suggested_action"`
	PackageName     string         `json:"package_name"`
	ElementCount    int            `json:"element_count"`
	ScreenshotURL   string         `json:"screenshot_url"`
	MinioKey        string         `json:"minio_key"`
	VLMError        string         `json:"vlm_error,omitempty"`
	TakenAt         time.Time      `json:"taken_at"`
}

func NormalizeDetectionMode(raw string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "", DetectionModeAuto:
		return DetectionModeAuto, nil
	case DetectionModeUI:
		return DetectionModeUI, nil
	case DetectionModeVLM:
		return DetectionModeVLM, nil
	default:
		return "", ErrInvalidDetectMode
	}
}

func NormalizeDetectionPlatform(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return "android"
	}
	return raw
}

func DetectScreenFromUIDump(dump UIDump) ScreenDetection {
	ctx := detectionContextFromDump(dump)
	detection := ScreenDetection{
		Serial:         dump.Serial,
		State:          ScreenStateUnknown,
		Confidence:     0,
		Source:         "ui",
		BackendUsed:    "ui",
		Elements:       ctx.visible,
		MatchedSignals: []string{},
		PackageName:    dump.PackageName,
		ElementCount:   dump.ElementCount,
		TakenAt:        dump.TakenAt,
		Description:    "Экран прочитан, но состояние не распознано",
	}

	switch {
	case ctx.hasAny("captcha", "verify you", "prove you", "drag the slider", "select 2 images", "robot"):
		detection.set(ScreenStateCaptcha, 0.95, "Обнаружен экран проверки/captcha", "ui:captcha")
		detection.Flags.Captcha = true
	case ctx.hasAny("suspended", "banned", "violated", "account at risk", "перманентно", "заблокирован"):
		detection.set(ScreenStateBan, 0.95, "Обнаружен экран блокировки или риска аккаунта", "ui:ban")
		detection.Flags.Ban = true
	case ctx.hasAny("something went wrong", "try again", "error", "ошибка", "повторите"):
		detection.set(ScreenStateError, 0.88, "Обнаружен экран ошибки", "ui:error")
		detection.Flags.Error = true
	case strings.Contains(strings.ToLower(dump.PackageName), "systemui") || ctx.hasAny("notification", "уведомлен", "silent notifications"):
		detection.set(ScreenStateNotification, 0.86, "Обнаружен системный экран уведомлений", "ui:notification")
	case ctx.hasAny("allow", "разрешить", "while using", "only this time", "don't allow", "deny", "permission"):
		detection.set(ScreenStatePermissionRequest, 0.9, "Обнаружен запрос разрешений", "ui:permission")
	case ctx.hasAny("install", "установить"):
		detection.set(ScreenStateInstall, 0.9, "Обнаружен экран установки", "ui:install")
	case ctx.hasAdSignal() && ctx.hasCloseSignal():
		detection.set(ScreenStateAdsFullscreen, 0.82, "Обнаружена полноэкранная реклама", "ui:ad", "ui:close")
	case ctx.hasEmailInput() && ctx.hasPasswordInput():
		detection.set(ScreenStateLogin, 0.95, "Экран входа: найдены поля логина и пароля", "ui:email_input", "ui:password_input")
		if ctx.hasAny("log in", "login", "sign in", "войти") {
			detection.MatchedSignals = appendUnique(detection.MatchedSignals, "ui:login_button")
		}
	case ctx.hasAny("sign up", "create account", "register", "join", "зарегистр", "создать аккаунт"):
		detection.set(ScreenStateRegistration, 0.86, "Обнаружен экран регистрации", "ui:registration")
	case ctx.hasAny("for you", "home", "feed", "reels", "posts", "лента") || ctx.hasAny("for you") && ctx.hasAny("following"):
		detection.set(ScreenStateMainFeed, 0.78, "Обнаружена основная лента", "ui:feed")
	case ctx.hasAny("edit profile", "followers", "подписчики"):
		detection.set(ScreenStateProfile, 0.82, "Обнаружен экран профиля", "ui:profile")
	case ctx.hasAny("search", "discover", "explore", "поиск"):
		detection.set(ScreenStateSearch, 0.8, "Обнаружен экран поиска", "ui:search")
	case ctx.hasAny("settings", "settings and privacy", "manage account", "настройки"):
		detection.set(ScreenStateSettings, 0.82, "Обнаружен экран настроек", "ui:settings")
	case dump.ElementCount == 0 || ctx.hasAny("loading", "загрузка", "progress"):
		detection.set(ScreenStateLoading, 0.6, "Экран похож на загрузку или содержит мало UI-элементов", "ui:loading")
	}

	detection.Elements = appendUnique(detection.Elements, ctx.visible...)
	return detection
}

func MergeScreenDetections(ui ScreenDetection, vlm VLMAnalysis) ScreenDetection {
	if vlm.State == "" {
		return ui
	}
	vlmState := NormalizeVLMState(vlm.State)
	if vlmState == "" {
		vlmState = ScreenStateUnknown
	}
	if vlm.Confidence <= 0 {
		vlm.Confidence = 0.7
	}

	result := ui
	useVLM := ui.State == ScreenStateUnknown || vlm.Confidence >= ui.Confidence || vlm.Flags.Captcha || vlm.Flags.Ban || vlm.Flags.Error
	if useVLM {
		result.State = vlmState
		result.Confidence = clampConfidence(vlm.Confidence)
		result.Description = firstNonEmpty(vlm.Description, ui.Description)
		result.Flags = DetectionFlags{
			Captcha: ui.Flags.Captcha || vlm.Flags.Captcha,
			Ban:     ui.Flags.Ban || vlm.Flags.Ban,
			Error:   ui.Flags.Error || vlm.Flags.Error,
		}
		result.SuggestedAction = firstNonEmpty(vlm.SuggestedAction, ui.SuggestedAction)
	}
	result.Source = "hybrid"
	result.BackendUsed = vlm.BackendUsed
	result.Elements = appendUnique(result.Elements, vlm.Elements...)
	result.MatchedSignals = appendUnique(result.MatchedSignals, vlm.MatchedSignals...)
	if result.Description == "" {
		result.Description = "Экран проанализирован UI и VLM слоями"
	}
	return result
}

func NormalizeVLMState(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "", "unknown", "other":
		return ScreenStateUnknown
	case "login", "sign_in", "signin", "log_in", ScreenStateLogin:
		return ScreenStateLogin
	case "registration", "register", "signup", "sign_up", ScreenStateRegistration:
		return ScreenStateRegistration
	case "feed", "home", "fyp", "timeline", "news_feed", ScreenStateMainFeed:
		return ScreenStateMainFeed
	case "profile", ScreenStateProfile:
		return ScreenStateProfile
	case "search", "discover", "explore", ScreenStateSearch:
		return ScreenStateSearch
	case "settings", ScreenStateSettings:
		return ScreenStateSettings
	case ScreenStateCaptcha:
		return ScreenStateCaptcha
	case "ban", "banned", ScreenStateBan:
		return ScreenStateBan
	case "error", ScreenStateError:
		return ScreenStateError
	case ScreenStateNotification:
		return ScreenStateNotification
	case "permission", "permission_dialog", ScreenStatePermissionRequest:
		return ScreenStatePermissionRequest
	case ScreenStateLoading:
		return ScreenStateLoading
	case "install", ScreenStateInstall:
		return ScreenStateInstall
	case "ad", "ads", "advertisement", ScreenStateAdsFullscreen:
		return ScreenStateAdsFullscreen
	default:
		return raw
	}
}

func (d *ScreenDetection) set(state string, confidence float64, description string, signals ...string) {
	d.State = state
	d.Confidence = clampConfidence(confidence)
	d.Description = description
	d.MatchedSignals = appendUnique(d.MatchedSignals, signals...)
}

type detectionContext struct {
	joined  string
	visible []string
}

func detectionContextFromDump(dump UIDump) detectionContext {
	visible := make([]string, 0, len(dump.Elements)*3)
	textParts := make([]string, 0, len(dump.Elements)*5+1)
	if dump.PackageName != "" {
		textParts = append(textParts, dump.PackageName)
	}
	for _, element := range dump.Elements {
		fields := []string{element.Type, element.Text, element.ResourceID, element.ContentDesc, element.Hint}
		for _, field := range fields {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			textParts = append(textParts, field)
			if field != element.Type && field != element.ResourceID {
				visible = append(visible, field)
			}
		}
	}
	return detectionContext{
		joined:  strings.ToLower(strings.Join(textParts, " ")),
		visible: appendUnique(nil, visible...),
	}
}

func (c detectionContext) hasAny(values ...string) bool {
	for _, value := range values {
		if strings.Contains(c.joined, strings.ToLower(value)) {
			return true
		}
	}
	return false
}

func (c detectionContext) hasEmailInput() bool {
	return c.hasAny("email", "e-mail", "mail", "phone", "username", "логин", "телефон", "почта")
}

func (c detectionContext) hasPasswordInput() bool {
	return c.hasAny("password", "пароль")
}

func (c detectionContext) hasAdSignal() bool {
	return c.hasAny("advertisement", "sponsored", "skip ad") || c.hasExactVisible("ad", "реклама")
}

func (c detectionContext) hasCloseSignal() bool {
	return c.hasAny("close", "skip", "закрыть", "пропустить") || c.hasExactVisible("x", "×")
}

func (c detectionContext) hasExactVisible(values ...string) bool {
	for _, visible := range c.visible {
		for _, value := range values {
			if strings.EqualFold(strings.TrimSpace(visible), strings.TrimSpace(value)) {
				return true
			}
		}
	}
	return false
}

func appendUnique(values []string, additions ...string) []string {
	seen := make(map[string]struct{}, len(values)+len(additions))
	result := make([]string, 0, len(values)+len(additions))
	for _, value := range append(values, additions...) {
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

func clampConfidence(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

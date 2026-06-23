package domain

import "strings"

const (
	MatchExact    = "exact"
	MatchContains = "contains"
)

type FindElementQuery struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	ResourceID  string `json:"resource_id"`
	ContentDesc string `json:"content_desc"`
	Hint        string `json:"hint"`
	Match       string `json:"match"`
}

type FindElementResult struct {
	Found   bool      `json:"found"`
	Element UIElement `json:"element,omitempty"`
	FoundBy string    `json:"found_by"`
}

type WaitForElementResult struct {
	Error      string    `json:"error,omitempty"`
	Found      bool      `json:"found"`
	Element    UIElement `json:"element,omitempty"`
	FoundBy    string    `json:"found_by"`
	TimeoutSec int       `json:"timeout_sec,omitempty"`
	WaitTimeMS int64     `json:"wait_time_ms"`
	CheckCount int       `json:"check_count"`
}

func FindElement(elements []UIElement, query FindElementQuery) (FindElementResult, error) {
	if err := ValidateFindElementQuery(query); err != nil {
		return FindElementResult{}, err
	}
	query.Match = normalizeMatch(query.Match)

	searches := []struct {
		name  string
		value string
		match func(UIElement, string, string) bool
	}{
		{name: "resource_id", value: query.ResourceID, match: matchResourceID},
		{name: "text", value: query.Text, match: matchText},
		{name: "content_desc", value: query.ContentDesc, match: matchContentDesc},
		{name: "hint", value: query.Hint, match: matchHint},
		{name: "type", value: query.Type, match: matchType},
	}

	for _, search := range searches {
		if strings.TrimSpace(search.value) == "" {
			continue
		}
		for _, element := range elements {
			if search.match(element, search.value, query.Match) {
				return FindElementResult{Found: true, Element: element, FoundBy: search.name}, nil
			}
		}
	}
	return FindElementResult{}, ErrElementNotFound
}

func ValidateFindElementQuery(query FindElementQuery) error {
	if normalizeMatch(query.Match) == "" {
		return ErrInvalidElementQuery
	}
	if strings.TrimSpace(query.Type) == "" &&
		strings.TrimSpace(query.Text) == "" &&
		strings.TrimSpace(query.ResourceID) == "" &&
		strings.TrimSpace(query.ContentDesc) == "" &&
		strings.TrimSpace(query.Hint) == "" {
		return ErrInvalidElementQuery
	}
	return nil
}

func normalizeMatch(raw string) string {
	switch raw {
	case "", MatchExact:
		return MatchExact
	case MatchContains:
		return MatchContains
	default:
		return ""
	}
}

func matchResourceID(element UIElement, want, _ string) bool {
	return element.ResourceID == want
}

func matchText(element UIElement, want, mode string) bool {
	return matchString(element.Text, want, mode)
}

func matchContentDesc(element UIElement, want, mode string) bool {
	return matchString(element.ContentDesc, want, mode)
}

func matchHint(element UIElement, want, mode string) bool {
	return matchString(element.Hint, want, mode)
}

func matchType(element UIElement, want, _ string) bool {
	return strings.EqualFold(element.Type, want)
}

func matchString(got, want, mode string) bool {
	if mode == MatchContains {
		return strings.Contains(strings.ToLower(got), strings.ToLower(want))
	}
	return got == want
}

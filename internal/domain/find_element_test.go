package domain

import (
	"errors"
	"testing"
)

func TestFindElementExactMatchesByPriority(t *testing.T) {
	elements := []UIElement{
		{Type: "Button", Text: "Войти", ResourceID: "com.app:id/login", Bounds: Bounds{X1: 1, Y1: 2, X2: 3, Y2: 4}},
		{Type: "Button", Text: "Войти", ResourceID: "com.app:id/other", Bounds: Bounds{X1: 5, Y1: 6, X2: 7, Y2: 8}},
	}

	result, err := FindElement(elements, FindElementQuery{
		Text:       "Войти",
		ResourceID: "com.app:id/other",
		Match:      MatchExact,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Found || result.FoundBy != "resource_id" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.Element.ResourceID != "com.app:id/other" {
		t.Fatalf("unexpected element: %+v", result.Element)
	}
}

func TestFindElementContainsMatchesTextContentDescAndHint(t *testing.T) {
	elements := []UIElement{
		{Type: "TextView", Text: "Already signed in"},
		{Type: "ImageButton", ContentDesc: "Create new post"},
		{Type: "EditText", Hint: "Email address"},
	}

	for _, tc := range []struct {
		name    string
		query   FindElementQuery
		foundBy string
		want    string
	}{
		{name: "text", query: FindElementQuery{Text: "SIGNED", Match: MatchContains}, foundBy: "text", want: "Already signed in"},
		{name: "content desc", query: FindElementQuery{ContentDesc: "create", Match: MatchContains}, foundBy: "content_desc", want: "Create new post"},
		{name: "hint", query: FindElementQuery{Hint: "email", Match: MatchContains}, foundBy: "hint", want: "Email address"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := FindElement(elements, tc.query)
			if err != nil {
				t.Fatal(err)
			}
			if !result.Found || result.FoundBy != tc.foundBy {
				t.Fatalf("unexpected result: %+v", result)
			}
			if result.Element.Text != tc.want && result.Element.ContentDesc != tc.want && result.Element.Hint != tc.want {
				t.Fatalf("unexpected element: %+v", result.Element)
			}
		})
	}
}

func TestFindElementExactMatchesType(t *testing.T) {
	result, err := FindElement([]UIElement{{Type: "EditText"}}, FindElementQuery{Type: "EditText"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Found || result.FoundBy != "type" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestFindElementNotFoundAndValidation(t *testing.T) {
	if _, err := FindElement([]UIElement{{Text: "OK"}}, FindElementQuery{Text: "Missing"}); !errors.Is(err, ErrElementNotFound) {
		t.Fatalf("expected ErrElementNotFound, got %v", err)
	}
	if _, err := FindElement(nil, FindElementQuery{}); !errors.Is(err, ErrInvalidElementQuery) {
		t.Fatalf("expected ErrInvalidElementQuery, got %v", err)
	}
	if _, err := FindElement(nil, FindElementQuery{Text: "OK", Match: "fuzzy"}); !errors.Is(err, ErrInvalidElementQuery) {
		t.Fatalf("expected ErrInvalidElementQuery for invalid match, got %v", err)
	}
}

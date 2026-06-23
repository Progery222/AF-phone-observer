package domain

import (
	"testing"
	"time"
)

func TestDetectScreenFromUIDumpRecognizesLoginScreen(t *testing.T) {
	dump := UIDump{
		Serial:      "stub",
		PackageName: "com.instagram.android",
		TakenAt:     time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC),
		Elements: []UIElement{
			{Type: "EditText", Hint: "Email"},
			{Type: "EditText", Hint: "Password"},
			{Type: "Button", Text: "Log in"},
		},
		ElementCount: 3,
	}

	got := DetectScreenFromUIDump(dump)

	if got.State != "login_screen" {
		t.Fatalf("expected login_screen, got %+v", got)
	}
	if got.Confidence < 0.9 {
		t.Fatalf("expected high confidence, got %f", got.Confidence)
	}
	assertContainsString(t, got.Elements, "Email")
	assertContainsString(t, got.Elements, "Password")
	assertContainsString(t, got.MatchedSignals, "ui:email_input")
	assertContainsString(t, got.MatchedSignals, "ui:password_input")
	if got.PackageName != "com.instagram.android" || got.ElementCount != 3 {
		t.Fatalf("unexpected metadata: %+v", got)
	}
}

func TestDetectScreenFromUIDumpRecognizesPermissionBanAndUnknown(t *testing.T) {
	for _, tc := range []struct {
		name string
		dump UIDump
		want string
	}{
		{
			name: "permission",
			dump: UIDump{Elements: []UIElement{
				{Type: "TextView", Text: "Allow Instagram to take pictures?"},
				{Type: "Button", Text: "Allow"},
			}, ElementCount: 2},
			want: "permission_request",
		},
		{
			name: "ban",
			dump: UIDump{Elements: []UIElement{
				{Type: "TextView", Text: "Your account was suspended"},
			}, ElementCount: 1},
			want: "ban_screen",
		},
		{
			name: "unknown",
			dump: UIDump{Elements: []UIElement{
				{Type: "Button", Text: "OK"},
			}, ElementCount: 1},
			want: "unknown",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectScreenFromUIDump(tc.dump)
			if got.State != tc.want {
				t.Fatalf("expected %s, got %+v", tc.want, got)
			}
			if got.Description == "" {
				t.Fatalf("expected diagnostic description: %+v", got)
			}
		})
	}
}

func TestDetectScreenFromUIDumpDoesNotTreatTikTokFeedAddTextAsAd(t *testing.T) {
	dump := UIDump{
		Serial:      "R5GL2218DMR",
		PackageName: "com.zhiliaoapp.musically",
		Elements: []UIElement{
			{Type: "Button", ContentDesc: "Add or remove this video from Favorites."},
			{Type: "TextView", Text: "Explore"},
			{Type: "TextView", Text: "Following"},
			{Type: "TextView", Text: "For You"},
			{Type: "TextView", Text: "Home"},
			{Type: "TextView", Text: "Profile"},
		},
		ElementCount: 6,
	}

	got := DetectScreenFromUIDump(dump)

	if got.State != "main_feed" {
		t.Fatalf("expected main_feed, got %+v", got)
	}
	assertContainsString(t, got.MatchedSignals, "ui:feed")
}

func TestMergeScreenDetectionsPrefersConfidentVLMAndKeepsUISignals(t *testing.T) {
	ui := ScreenDetection{
		State:          "unknown",
		Confidence:     0.2,
		Source:         "ui",
		Elements:       []string{"OK"},
		MatchedSignals: []string{"ui:visible_text"},
		TakenAt:        time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC),
	}
	vlm := VLMAnalysis{
		State:           "login_screen",
		Confidence:      0.87,
		BackendUsed:     "ollama",
		Description:     "Login form is visible",
		Elements:        []string{"Email", "Password"},
		MatchedSignals:  []string{"vlm:login"},
		Flags:           DetectionFlags{Captcha: false, Ban: false, Error: false},
		SuggestedAction: "fill_login_fields",
	}

	got := MergeScreenDetections(ui, vlm)

	if got.State != "login_screen" || got.Source != "hybrid" || got.BackendUsed != "ollama" {
		t.Fatalf("unexpected merged detection: %+v", got)
	}
	assertContainsString(t, got.MatchedSignals, "ui:visible_text")
	assertContainsString(t, got.MatchedSignals, "vlm:login")
	assertContainsString(t, got.Elements, "Email")
	if got.SuggestedAction != "fill_login_fields" {
		t.Fatalf("unexpected suggested action: %q", got.SuggestedAction)
	}
}

func assertContainsString(t *testing.T, values []string, want string) {
	t.Helper()
	for _, value := range values {
		if value == want {
			return
		}
	}
	t.Fatalf("expected %q in %#v", want, values)
}

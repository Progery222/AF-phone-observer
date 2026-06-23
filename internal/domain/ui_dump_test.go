package domain

import (
	"testing"
	"time"
)

func TestParseUIDumpParsesElementsBoundsAndCenters(t *testing.T) {
	takenAt := time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC)
	xmlDump := `<hierarchy>
		<node text="Войти" content-desc="Create" resource-id="com.app:id/login" class="android.widget.Button" bounds="[200,500][600,580]" />
		<node text="" hint="Email" resource-id="com.app:id/email" class="android.widget.EditText" bounds="[100,200][700,300]" />
	</hierarchy>`

	dump, err := ParseUIDump("stub", xmlDump, takenAt)
	if err != nil {
		t.Fatal(err)
	}
	if dump.Serial != "stub" || dump.XMLDump != xmlDump || !dump.TakenAt.Equal(takenAt) {
		t.Fatalf("unexpected dump metadata: %+v", dump)
	}
	if dump.ElementCount != 2 || len(dump.Elements) != 2 {
		t.Fatalf("expected 2 elements, got count=%d len=%d", dump.ElementCount, len(dump.Elements))
	}

	button := dump.Elements[0]
	if button.Type != "Button" || button.Text != "Войти" || button.ResourceID != "com.app:id/login" || button.ContentDesc != "Create" {
		t.Fatalf("unexpected button: %+v", button)
	}
	if button.Bounds != (Bounds{X1: 200, Y1: 500, X2: 600, Y2: 580}) {
		t.Fatalf("unexpected bounds: %+v", button.Bounds)
	}
	if button.Center != (Point{X: 400, Y: 540}) {
		t.Fatalf("unexpected center: %+v", button.Center)
	}

	input := dump.Elements[1]
	if input.Type != "EditText" || input.Hint != "Email" {
		t.Fatalf("unexpected input: %+v", input)
	}
}

func TestParseUIDumpReturnsErrorForInvalidXML(t *testing.T) {
	if _, err := ParseUIDump("stub", `<hierarchy><node>`, time.Now().UTC()); err == nil {
		t.Fatal("expected invalid XML error")
	}
}

func TestParseUIDumpUsesRegexFallbackForMalformedXML(t *testing.T) {
	xmlDump := `<hierarchy><node text="OK" content-desc="Create" class="android.widget.Button" resource-id="stub:id/ok" bounds="[0,0][2,2]">`

	dump, err := ParseUIDump("stub", xmlDump, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if dump.ElementCount != 1 {
		t.Fatalf("expected 1 fallback element, got %d", dump.ElementCount)
	}
	got := dump.Elements[0]
	if got.Text != "OK" || got.ContentDesc != "Create" || got.ResourceID != "stub:id/ok" || got.Type != "Button" {
		t.Fatalf("unexpected fallback element: %+v", got)
	}
}

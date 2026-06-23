package domain

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var boundsPattern = regexp.MustCompile(`^\[(\d+),(\d+)\]\[(\d+),(\d+)\]$`)
var nodePattern = regexp.MustCompile(`<node\s+([^>]*)/?>`)
var attrPattern = regexp.MustCompile(`([A-Za-z0-9:-]+)="([^"]*)"`)

type UIDump struct {
	Serial       string      `json:"serial"`
	XMLDump      string      `json:"xml_dump"`
	Elements     []UIElement `json:"elements"`
	ElementCount int         `json:"element_count"`
	PackageName  string      `json:"package_name"`
	TakenAt      time.Time   `json:"taken_at"`
}

type UIElement struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	ResourceID  string `json:"resource_id"`
	ContentDesc string `json:"content_desc"`
	Hint        string `json:"hint"`
	Bounds      Bounds `json:"bounds"`
	Center      Point  `json:"center"`
}

type Bounds struct {
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func ParseUIDump(serial, xmlDump string, takenAt time.Time) (UIDump, error) {
	elements, err := parseUIDumpXML(xmlDump)
	if err != nil {
		fallbackElements := parseUIDumpRegexFallback(xmlDump)
		if len(fallbackElements) == 0 {
			return UIDump{}, err
		}
		elements = fallbackElements
	}
	return UIDump{
		Serial:       serial,
		XMLDump:      xmlDump,
		Elements:     elements,
		ElementCount: len(elements),
		PackageName:  extractPackageName(xmlDump),
		TakenAt:      takenAt,
	}, nil
}

func parseUIDumpXML(xmlDump string) ([]UIElement, error) {
	decoder := xml.NewDecoder(strings.NewReader(xmlDump))
	elements := make([]UIElement, 0)
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "node" {
			continue
		}
		element, err := parseUIElement(start)
		if err != nil {
			return nil, err
		}
		elements = append(elements, element)
	}
	return elements, nil
}

func parseUIElement(start xml.StartElement) (UIElement, error) {
	attrs := make(map[string]string, len(start.Attr))
	for _, attr := range start.Attr {
		attrs[attr.Name.Local] = attr.Value
	}
	return parseUIElementAttrs(attrs)
}

func parseUIElementAttrs(attrs map[string]string) (UIElement, error) {
	element := UIElement{
		Type:        shortClassName(attrs["class"]),
		Text:        attrs["text"],
		ResourceID:  attrs["resource-id"],
		ContentDesc: attrs["content-desc"],
		Hint:        attrs["hint"],
	}
	rawBounds := attrs["bounds"]
	if rawBounds != "" {
		bounds, err := parseBounds(rawBounds)
		if err != nil {
			return UIElement{}, err
		}
		element.Bounds = bounds
		element.Center = Point{
			X: (bounds.X1 + bounds.X2) / 2,
			Y: (bounds.Y1 + bounds.Y2) / 2,
		}
	}
	return element, nil
}

func parseUIDumpRegexFallback(xmlDump string) []UIElement {
	matches := nodePattern.FindAllStringSubmatch(xmlDump, -1)
	elements := make([]UIElement, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		attrs := make(map[string]string)
		for _, attr := range attrPattern.FindAllStringSubmatch(match[1], -1) {
			if len(attr) == 3 {
				attrs[attr[1]] = attr[2]
			}
		}
		if attrs["bounds"] == "" {
			continue
		}
		element, err := parseUIElementAttrs(attrs)
		if err != nil {
			continue
		}
		elements = append(elements, element)
	}
	return elements
}

func parseBounds(raw string) (Bounds, error) {
	match := boundsPattern.FindStringSubmatch(raw)
	if match == nil {
		return Bounds{}, fmt.Errorf("некорректные bounds: %s", raw)
	}
	values := make([]int, 4)
	for i := range values {
		parsed, err := strconv.Atoi(match[i+1])
		if err != nil {
			return Bounds{}, err
		}
		values[i] = parsed
	}
	return Bounds{X1: values[0], Y1: values[1], X2: values[2], Y2: values[3]}, nil
}

func shortClassName(raw string) string {
	if raw == "" {
		return ""
	}
	if idx := strings.LastIndexAny(raw, ".$"); idx >= 0 && idx+1 < len(raw) {
		return raw[idx+1:]
	}
	return raw
}

func extractPackageName(xmlDump string) string {
	const marker = `package="`
	start := strings.Index(xmlDump, marker)
	if start < 0 {
		return ""
	}
	start += len(marker)
	end := strings.Index(xmlDump[start:], `"`)
	if end <= 0 {
		return ""
	}
	return xmlDump[start : start+end]
}

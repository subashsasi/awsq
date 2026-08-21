package formatter

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	headers := []string{"ID", "NAME", "STATE"}
	rows := [][]string{
		{"i-123", "web-server", "running"},
		{"i-456", "api-server", "stopped"},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printJSON(headers, rows)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify it's valid JSON
	var result []map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("printJSON output is not valid JSON: %v", err)
	}

	// Verify content
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
	if result[0]["id"] != "i-123" {
		t.Errorf("expected id=i-123, got %q", result[0]["id"])
	}
	if result[0]["name"] != "web-server" {
		t.Errorf("expected name=web-server, got %q", result[0]["name"])
	}
	if result[1]["state"] != "stopped" {
		t.Errorf("expected state=stopped, got %q", result[1]["state"])
	}
}

func TestPrintCSV(t *testing.T) {
	headers := []string{"ID", "NAME"}
	rows := [][]string{
		{"i-123", "web-server"},
		{"i-456", "api-server"},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printCSV(headers, rows)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines (header + 2 rows), got %d", len(lines))
	}
	if lines[0] != "ID,NAME" {
		t.Errorf("expected header 'ID,NAME', got %q", lines[0])
	}
	if lines[1] != "i-123,web-server" {
		t.Errorf("expected row 'i-123,web-server', got %q", lines[1])
	}
}

func TestPrintTable(t *testing.T) {
	headers := []string{"ID", "NAME"}
	rows := [][]string{
		{"i-123", "web-server"},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printTable(headers, rows)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should contain header and data
	if !strings.Contains(output, "ID") {
		t.Error("table output missing header 'ID'")
	}
	if !strings.Contains(output, "NAME") {
		t.Error("table output missing header 'NAME'")
	}
	if !strings.Contains(output, "i-123") {
		t.Error("table output missing data 'i-123'")
	}
	if !strings.Contains(output, "web-server") {
		t.Error("table output missing data 'web-server'")
	}
}

func TestPrintTableEmpty(t *testing.T) {
	headers := []string{"ID", "NAME"}
	var rows [][]string

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printTable(headers, rows)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "No resources found") {
		t.Errorf("expected 'No resources found' for empty rows, got %q", output)
	}
}

func TestPrintTableTruncation(t *testing.T) {
	headers := []string{"DESC"}
	rows := [][]string{
		{"This is a very long description that should exceed the forty character column width limit set in the formatter"},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printTable(headers, rows)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should contain truncation character
	if !strings.Contains(output, "…") {
		t.Error("expected truncation character '…' for long values")
	}
}

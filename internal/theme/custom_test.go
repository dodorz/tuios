package theme

import (
	"os"
	"path/filepath"
	"testing"

	tint "github.com/lrstanley/bubbletint/v2"
)

// TestLoadCustomThemeFile_FullTheme tests loading a complete theme JSON file.
func TestLoadCustomThemeFile_FullTheme(t *testing.T) {
	dir := t.TempDir()
	themeJSON := `{
		"id": "test-full",
		"display_name": "Test Full Theme",
		"dark": true,
		"fg": "#d4d4d4",
		"bg": "#1e1e2e",
		"cursor": "#f5e0dc",
		"black": "#45475a",
		"red": "#f38ba8",
		"green": "#a6e3a1",
		"yellow": "#f9e2af",
		"blue": "#89b4fa",
		"purple": "#cba6f7",
		"cyan": "#94e2d5",
		"white": "#bac2de",
		"bright_black": "#585b70",
		"bright_red": "#f38ba8",
		"bright_green": "#a6e3a1",
		"bright_yellow": "#f9e2af",
		"bright_blue": "#89b4fa",
		"bright_purple": "#cba6f7",
		"bright_cyan": "#94e2d5",
		"bright_white": "#a6adc8"
	}`

	path := filepath.Join(dir, "test-full.json")
	if err := os.WriteFile(path, []byte(themeJSON), 0600); err != nil {
		t.Fatal(err)
	}

	theme, err := LoadCustomThemeFile(path)
	if err != nil {
		t.Fatalf("LoadCustomThemeFile failed: %v", err)
	}

	if theme.ID != "test-full" {
		t.Errorf("expected ID 'test-full', got %q", theme.ID)
	}
	if theme.DisplayName != "Test Full Theme" {
		t.Errorf("expected DisplayName 'Test Full Theme', got %q", theme.DisplayName)
	}
	if !theme.Dark {
		t.Error("expected Dark to be true")
	}

	// Verify all color fields are populated
	colors := []*tint.Color{
		theme.Fg, theme.Bg, theme.Cursor,
		theme.Black, theme.Red, theme.Green, theme.Yellow,
		theme.Blue, theme.Purple, theme.Cyan, theme.White,
		theme.BrightBlack, theme.BrightRed, theme.BrightGreen, theme.BrightYellow,
		theme.BrightBlue, theme.BrightPurple, theme.BrightCyan, theme.BrightWhite,
	}
	for i, c := range colors {
		if c == nil {
			t.Errorf("color at index %d is nil", i)
		}
	}
}

// TestLoadCustomThemeFile_Partial tests loading a minimal theme (only fg/bg).
func TestLoadCustomThemeFile_Partial(t *testing.T) {
	dir := t.TempDir()
	themeJSON := `{
		"id": "minimal-dark",
		"fg": "#c0c0c0",
		"bg": "#1a1a1a"
	}`

	path := filepath.Join(dir, "minimal-dark.json")
	if err := os.WriteFile(path, []byte(themeJSON), 0600); err != nil {
		t.Fatal(err)
	}

	theme, err := LoadCustomThemeFile(path)
	if err != nil {
		t.Fatalf("LoadCustomThemeFile failed: %v", err)
	}

	if theme.ID != "minimal-dark" {
		t.Errorf("expected ID 'minimal-dark', got %q", theme.ID)
	}

	// fillDefaults should have populated all ANSI colors
	colors := map[string]*tint.Color{
		"Cursor":      theme.Cursor,
		"Black":       theme.Black,
		"Red":         theme.Red,
		"Green":       theme.Green,
		"Yellow":      theme.Yellow,
		"Blue":        theme.Blue,
		"Purple":      theme.Purple,
		"Cyan":        theme.Cyan,
		"White":       theme.White,
		"BrightBlack": theme.BrightBlack,
		"BrightRed":   theme.BrightRed,
		"BrightGreen": theme.BrightGreen,
	}
	for name, c := range colors {
		if c == nil {
			t.Errorf("fillDefaults should have set %s, got nil", name)
		}
	}

	// Cursor should default to Fg color
	if theme.Cursor.R != theme.Fg.R || theme.Cursor.G != theme.Fg.G || theme.Cursor.B != theme.Fg.B {
		t.Error("Cursor should default to Fg color")
	}

	// Bright variants should default to normal variants
	if theme.BrightBlack.R != theme.Black.R {
		t.Error("BrightBlack should default to Black")
	}
}

// TestLoadCustomThemeFile_IDFromFilename tests ID derivation from filename.
func TestLoadCustomThemeFile_IDFromFilename(t *testing.T) {
	dir := t.TempDir()
	themeJSON := `{
		"fg": "#ffffff",
		"bg": "#000000"
	}`

	path := filepath.Join(dir, "My-Cool-Theme.json")
	if err := os.WriteFile(path, []byte(themeJSON), 0600); err != nil {
		t.Fatal(err)
	}

	theme, err := LoadCustomThemeFile(path)
	if err != nil {
		t.Fatalf("LoadCustomThemeFile failed: %v", err)
	}

	if theme.ID != "my-cool-theme" {
		t.Errorf("expected ID 'my-cool-theme' (derived from filename), got %q", theme.ID)
	}
	if theme.DisplayName != "my-cool-theme" {
		t.Errorf("expected DisplayName 'my-cool-theme', got %q", theme.DisplayName)
	}
}

// TestLoadCustomThemeFile_InvalidJSON tests graceful error on invalid JSON.
func TestLoadCustomThemeFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not valid json{{{"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadCustomThemeFile(path)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// TestLoadCustomThemes_EmptyDir tests loading from an empty directory.
func TestLoadCustomThemes_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	loaded, err := LoadCustomThemes(dir)
	if err != nil {
		t.Fatalf("LoadCustomThemes on empty dir should not error: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("expected 0 loaded themes, got %d", len(loaded))
	}
}

// TestLoadCustomThemes_IgnoresNonJSON tests that non-JSON files are skipped.
func TestLoadCustomThemes_IgnoresNonJSON(t *testing.T) {
	dir := t.TempDir()

	// Create non-JSON files
	for _, name := range []string{"readme.txt", "notes.md", ".hidden"} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("not a theme"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	loaded, err := LoadCustomThemes(dir)
	if err != nil {
		t.Fatalf("LoadCustomThemes should not error: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("expected 0 loaded themes, got %d", len(loaded))
	}
}

// TestLoadCustomThemes_Registration tests that loaded themes appear in TintIDs().
func TestLoadCustomThemes_Registration(t *testing.T) {
	dir := t.TempDir()
	themeJSON := `{
		"id": "test-registration-unique",
		"fg": "#ffffff",
		"bg": "#000000"
	}`

	path := filepath.Join(dir, "test-registration-unique.json")
	if err := os.WriteFile(path, []byte(themeJSON), 0600); err != nil {
		t.Fatal(err)
	}

	// Initialize registry first
	tint.NewDefaultRegistry()

	loaded, err := LoadCustomThemes(dir)
	if err != nil {
		t.Fatalf("LoadCustomThemes failed: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 loaded theme, got %d", len(loaded))
	}

	// Check that the theme appears in the registry
	ids := tint.TintIDs()
	found := false
	for _, id := range ids {
		if id == "test-registration-unique" {
			found = true
			break
		}
	}
	if !found {
		t.Error("custom theme 'test-registration-unique' not found in TintIDs()")
	}
}

// TestFillDefaults tests the fillDefaults function directly.
func TestFillDefaults(t *testing.T) {
	theme := &tint.Tint{}
	fillDefaults(theme)

	if theme.Fg == nil {
		t.Error("Fg should be set by fillDefaults")
	}
	if theme.Bg == nil {
		t.Error("Bg should be set by fillDefaults")
	}
	if theme.Cursor == nil {
		t.Error("Cursor should be set by fillDefaults")
	}
	if theme.Black == nil {
		t.Error("Black should be set by fillDefaults")
	}
	if theme.BrightWhite == nil {
		t.Error("BrightWhite should be set by fillDefaults")
	}
}

// TestCopyColor tests the copyColor helper.
func TestCopyColor(t *testing.T) {
	original := &tint.Color{R: 255, G: 128, B: 0, A: 255}
	copied := copyColor(original)

	if copied == original {
		t.Error("copyColor should return a different pointer")
	}
	if copied.R != original.R || copied.G != original.G || copied.B != original.B {
		t.Error("copyColor should copy values")
	}

	// Modifying copy should not affect original
	copied.R = 0
	if original.R == 0 {
		t.Error("modifying copy should not affect original")
	}

	// Test nil input
	if copyColor(nil) != nil {
		t.Error("copyColor(nil) should return nil")
	}
}

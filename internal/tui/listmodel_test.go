package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestListModel_Navigation verifies that cursor navigation wraps correctly.
func TestListModel_Navigation(t *testing.T) {
	lm := NewListModel([]string{"a", "b", "c"})
	lm.FooterText = "Cancel"

	// Start at cursor 0.
	if lm.Cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", lm.Cursor)
	}

	// Move down: 0 -> 1
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyDown})
	if lm.Cursor != 1 {
		t.Errorf("after down: cursor = %d, want 1", lm.Cursor)
	}

	// Move down: 1 -> 2
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyDown})
	if lm.Cursor != 2 {
		t.Errorf("after down: cursor = %d, want 2", lm.Cursor)
	}

	// Move down: 2 -> 3 (footer)
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyDown})
	if lm.Cursor != 3 {
		t.Errorf("after down: cursor = %d, want 3 (footer)", lm.Cursor)
	}

	// Move down: 3 -> 0 (wrap)
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyDown})
	if lm.Cursor != 0 {
		t.Errorf("after down from footer: cursor = %d, want 0", lm.Cursor)
	}

	// Move up: 0 -> 3 (wrap to footer)
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyUp})
	if lm.Cursor != 3 {
		t.Errorf("after up from 0: cursor = %d, want 3 (footer)", lm.Cursor)
	}

	// Move up: 3 -> 2
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyUp})
	if lm.Cursor != 2 {
		t.Errorf("after up: cursor = %d, want 2", lm.Cursor)
	}
}

// TestListModel_Select verifies that enter selects the correct item.
func TestListModel_Select(t *testing.T) {
	lm := NewListModel([]string{"first", "second", "third"})
	lm.FooterText = "Cancel"

	// Move down once to "second" (cursor 1).
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyDown})

	// Press enter.
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if lm.SelectedItem != "second" {
		t.Errorf("selected item = %q, want %q", lm.SelectedItem, "second")
	}
	if lm.SelectedIndex != 1 {
		t.Errorf("selected index = %d, want 1", lm.SelectedIndex)
	}
	if lm.Quitted {
		t.Error("should not be quitted after selecting an item")
	}
}

// TestListModel_FooterCancel verifies that pressing enter on the footer quits.
func TestListModel_FooterCancel(t *testing.T) {
	lm := NewListModel([]string{"a"})
	lm.FooterText = "Cancel"

	// Move down to footer.
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyDown})

	// Press enter on footer.
	lm, cmd := lm.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !lm.Quitted {
		t.Error("should be quitted after selecting footer")
	}
	if lm.SelectedItem != "" {
		t.Error("should not have selected item after footer")
	}
	if cmd == nil {
		t.Error("should return tea.Quit command after footer selection")
	}
}

// TestListModel_QuitKeys verifies that q, esc, and ctrl+c all quit.
func TestListModel_QuitKeys(t *testing.T) {
	quitKeys := []tea.KeyMsg{
		{Type: tea.KeyCtrlC},
		{Type: tea.KeyEsc},
		{Type: tea.KeyRunes, Runes: []rune("q")},
	}
	for _, key := range quitKeys {
		lm := NewListModel([]string{"a"})
		lm, cmd := lm.Update(key)
		if !lm.Quitted {
			t.Errorf("key %v: should quit", key)
		}
		if cmd == nil {
			t.Errorf("key %v: should return tea.Quit", key)
		}
	}
}

// TestListModel_View verifies that the view renders correctly.
func TestListModel_View(t *testing.T) {
	lm := NewListModel([]string{"item1", "item2"})
	lm.Title = "Test"
	lm.Header = "Choose one"
	lm.FooterText = "Done"
	lm.Help = "Navigate with arrows"

	view := lm.View()

	// Should contain title, header, items, footer, help.
	for _, expected := range []string{"Test", "Choose one", "item1", "item2", "Done", "Navigate with arrows"} {
		if !containsStr(view, expected) {
			t.Errorf("View should contain %q", expected)
		}
	}
}

// TestListModel_NoFooter verifies that without footer text, no footer is rendered.
func TestListModel_NoFooter(t *testing.T) {
	lm := NewListModel([]string{"a"})
	lm.FooterText = "" // No footer.

	view := lm.View()
	if containsStr(view, "Cancel") {
		t.Error("View should not contain footer text when FooterText is empty")
	}
}

// TestListModel_EmptyItems verifies that an empty list doesn't panic.
func TestListModel_EmptyItems(t *testing.T) {
	lm := NewListModel([]string{})
	lm.FooterText = "Cancel"

	// Should not panic.
	view := lm.View()
	if view == "" {
		t.Error("View should not be empty for empty items with footer")
	}

	// Press enter on footer.
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !lm.Quitted {
		t.Error("should quit when only footer exists")
	}
}

// TestListModel_KJKeys verifies that k and j keys work as up/down.
func TestListModel_KJKeys(t *testing.T) {
	lm := NewListModel([]string{"a", "b"})

	// j = down
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if lm.Cursor != 1 {
		t.Errorf("after j: cursor = %d, want 1", lm.Cursor)
	}

	// k = up
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if lm.Cursor != 0 {
		t.Errorf("after k: cursor = %d, want 0", lm.Cursor)
	}
}

// TestSelectorModel_Wrapper verifies that SelectorModel correctly wraps ListModel.
func TestSelectorModel_Wrapper(t *testing.T) {
	m := InitialSelectorModel([]string{"file1.yaml", "file2.yaml"})

	// Initial state.
	if m.Cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", m.Cursor)
	}
	if m.SelectedFile != "" {
		t.Errorf("initial selected file = %q, want empty", m.SelectedFile)
	}

	// Navigate and select — use pointer receiver type assertion.
	ptr := &m
	m1, _ := ptr.Update(tea.KeyMsg{Type: tea.KeyDown})
	ptr = m1.(*SelectorModel)
	m2, _ := ptr.Update(tea.KeyMsg{Type: tea.KeyEnter})
	ptr = m2.(*SelectorModel)

	if ptr.SelectedFile != "file2.yaml" {
		t.Errorf("selected file = %q, want %q", ptr.SelectedFile, "file2.yaml")
	}
	if ptr.Cursor != 1 {
		t.Errorf("cursor = %d, want 1", ptr.Cursor)
	}
}

// TestChartListModel_Wrapper verifies that ChartListModel correctly wraps ListModel.
func TestChartListModel_Wrapper(t *testing.T) {
	m := InitialChartListModel([]string{"app-a", "app-b", "app-c"})

	// Navigate and select — use pointer receiver type assertion.
	ptr := &m
	m1, _ := ptr.Update(tea.KeyMsg{Type: tea.KeyDown})
	ptr = m1.(*ChartListModel)
	m2, _ := ptr.Update(tea.KeyMsg{Type: tea.KeyDown})
	ptr = m2.(*ChartListModel)
	m3, _ := ptr.Update(tea.KeyMsg{Type: tea.KeyEnter})
	ptr = m3.(*ChartListModel)

	if ptr.SelectedChart != "app-c" {
		t.Errorf("selected chart = %q, want %q", ptr.SelectedChart, "app-c")
	}
	if ptr.Cursor != 2 {
		t.Errorf("cursor = %d, want 2", ptr.Cursor)
	}
}

// TestChartListModel_Cancel verifies that Cancel footer works for ChartListModel.
func TestChartListModel_Cancel(t *testing.T) {
	m := InitialChartListModel([]string{"app-a"})

	// Move to footer — use pointer receiver type assertion.
	ptr := &m
	m1, _ := ptr.Update(tea.KeyMsg{Type: tea.KeyDown})
	ptr = m1.(*ChartListModel)

	// Press enter on Cancel.
	m2, cmd := ptr.Update(tea.KeyMsg{Type: tea.KeyEnter})
	ptr = m2.(*ChartListModel)

	if !ptr.Quitted {
		t.Error("should be quitted after selecting Cancel")
	}
	if ptr.SelectedChart != "" {
		t.Error("should not have selected chart after Cancel")
	}
	if cmd == nil {
		t.Error("should return tea.Quit command")
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

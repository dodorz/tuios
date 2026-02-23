package input

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Gaurav-Gosain/tuios/internal/app"
)

// TestTransitionGuardCondition directly tests the guard condition that suppresses
// misparsed mouse-sequence fragments during the AllMotion→CellMotion transition.
// This prevents phantom keypresses in apps like kakoune/nano (issue #78).
func TestTransitionGuardCondition(t *testing.T) {
	// guardShouldBlock mirrors the condition in HandleTerminalModeKey
	guardShouldBlock := func(msg tea.KeyPressMsg, enteredAt time.Time) bool {
		return msg.Mod == 0 && msg.Text != "" && !enteredAt.IsZero() &&
			time.Since(enteredAt) < 150*time.Millisecond
	}

	tests := []struct {
		name        string
		key         tea.KeyPressMsg
		enteredAgo  time.Duration
		useZeroTime bool
		wantBlock   bool
	}{
		{
			name:       "digit during transition - block (mouse fragment)",
			key:        tea.KeyPressMsg{Code: '2', Text: "2"},
			enteredAgo: 10 * time.Millisecond,
			wantBlock:  true,
		},
		{
			name:       "letter M during transition - block (SGR terminator)",
			key:        tea.KeyPressMsg{Code: 'M', Text: "M"},
			enteredAgo: 50 * time.Millisecond,
			wantBlock:  true,
		},
		{
			name:       "semicolon during transition - block (SGR separator)",
			key:        tea.KeyPressMsg{Code: ';', Text: ";"},
			enteredAgo: 100 * time.Millisecond,
			wantBlock:  true,
		},
		{
			name:       "digit after transition window - pass through",
			key:        tea.KeyPressMsg{Code: '2', Text: "2"},
			enteredAgo: 200 * time.Millisecond,
			wantBlock:  false,
		},
		{
			name:       "ctrl+b during transition - pass through (has modifier)",
			key:        tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl},
			enteredAgo: 10 * time.Millisecond,
			wantBlock:  false,
		},
		{
			name:       "escape during transition - pass through (no Text)",
			key:        tea.KeyPressMsg{Code: tea.KeyEscape},
			enteredAgo: 10 * time.Millisecond,
			wantBlock:  false,
		},
		{
			name:        "zero timestamp - pass through (no mode switch recorded)",
			key:         tea.KeyPressMsg{Code: '2', Text: "2"},
			useZeroTime: true,
			wantBlock:   false,
		},
		{
			name:       "alt+key during transition - pass through (has modifier)",
			key:        tea.KeyPressMsg{Code: '1', Mod: tea.ModAlt, Text: "1"},
			enteredAgo: 10 * time.Millisecond,
			wantBlock:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var enteredAt time.Time
			if !tt.useZeroTime {
				enteredAt = time.Now().Add(-tt.enteredAgo)
			}

			got := guardShouldBlock(tt.key, enteredAt)
			if got != tt.wantBlock {
				t.Errorf("guardShouldBlock() = %v, want %v", got, tt.wantBlock)
			}
		})
	}
}

// TestTransitionGuardIntegration tests the full HandleTerminalModeKey function
// to verify that misparsed keys are actually suppressed during the transition.
func TestTransitionGuardIntegration(t *testing.T) {
	tests := []struct {
		name        string
		key         tea.KeyPressMsg
		enteredAgo  time.Duration
		shouldBlock bool
	}{
		{
			name:        "digit during transition - blocked",
			key:         tea.KeyPressMsg{Code: '2', Text: "2"},
			enteredAgo:  10 * time.Millisecond,
			shouldBlock: true,
		},
		{
			name:        "digit after transition - passes through",
			key:         tea.KeyPressMsg{Code: '2', Text: "2"},
			enteredAgo:  200 * time.Millisecond,
			shouldBlock: false,
		},
		{
			name:        "escape during transition - passes through",
			key:         tea.KeyPressMsg{Code: tea.KeyEscape},
			enteredAgo:  10 * time.Millisecond,
			shouldBlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &app.OS{
				Mode:                  app.TerminalMode,
				FocusedWindow:         -1, // no focused window
				TerminalModeEnteredAt: time.Now().Add(-tt.enteredAgo),
			}

			result, _ := HandleTerminalModeKey(tt.key, o)

			// If guard blocked: Mode stays TerminalMode (returned immediately).
			// If guard passed: hits "no focused window" → switches to WindowManagementMode.
			wasBlocked := result.Mode == app.TerminalMode
			if wasBlocked != tt.shouldBlock {
				t.Errorf("blocked = %v, want %v (key=%q, enteredAgo=%v)",
					wasBlocked, tt.shouldBlock, tt.key.String(), tt.enteredAgo)
			}
		})
	}
}

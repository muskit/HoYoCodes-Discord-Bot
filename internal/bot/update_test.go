package bot_test

import (
	"testing"

	"github.com/muskit/hoyocodes-discord-bot/internal/bot"
	"github.com/muskit/hoyocodes-discord-bot/internal/db"
)

func TestShouldNotify(t *testing.T) {
	tests := []struct {
		name     string
		sub      db.Subscription
		chg      bot.CodeChanges
		expected bool
	}{
		{
			name: "should notify on additions",
			sub: db.Subscription{
				AnnounceAdds: true,
				AnnounceRems: false,
			},
			chg: bot.CodeChanges{
				Added: [][]string{{"ABC123", "Description"}},
				Removed: [][]string{},
			},
			expected: true,
		},
		{
			name: "should notify on removals",
			sub: db.Subscription{
				AnnounceAdds: false,
				AnnounceRems: true,
			},
			chg: bot.CodeChanges{
				Added: [][]string{},
				Removed: [][]string{{"XYZ789", "Description"}},
			},
			expected: true,
		},
		{
			name: "should not notify when no changes",
			sub: db.Subscription{
				AnnounceAdds: true,
				AnnounceRems: true,
			},
			chg: bot.CodeChanges{
				Added: [][]string{},
				Removed: [][]string{},
			},
			expected: false,
		},
		{
			name: "should not notify when announcements are disabled",
			sub: db.Subscription{
				AnnounceAdds: false,
				AnnounceRems: false,
			},
			chg: bot.CodeChanges{
				Added: [][]string{{"ABC123", "Description"}},
				Removed: [][]string{{"XYZ789", "Description"}},
			},
			expected: false,
		},
		{
			name: "should notify on both additions and removals",
			sub: db.Subscription{
				AnnounceAdds: true,
				AnnounceRems: true,
			},
			chg: bot.CodeChanges{
				Added: [][]string{{"ABC123", "Description"}},
				Removed: [][]string{{"XYZ789", "Description"}},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bot.ShouldNotify(tt.sub, tt.chg)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

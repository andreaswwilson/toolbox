package todo

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
	"toolbox/internal/config"

	"github.com/stretchr/testify/assert"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func TestShouldMarkTodoAsDone(t *testing.T) {
	// Create a cxt with a debug logger
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	configData := &config.Config{Logger: logger}
	ctx = context.WithValue(ctx, config.ConfigKey, configData)

	t.Parallel()
	tests := []struct {
		name         string
		todo         *gitlab.Todo
		mergeRequest *gitlab.MergeRequest
		expected     bool
	}{
		{
			name:     "Todo with action name not review_requested",
			todo:     &gitlab.Todo{ActionName: "assigned", Target: &gitlab.TodoTarget{}},
			expected: false,
		},
		{
			name:     "Todo with merged target",
			todo:     &gitlab.Todo{ActionName: "review_requested", Target: &gitlab.TodoTarget{MergedAt: func() *time.Time { t := time.Now(); return &t }()}},
			expected: true,
		},
		{
			name:     "Todo with closed target",
			todo:     &gitlab.Todo{ActionName: "review_requested", Target: &gitlab.TodoTarget{State: "closed"}},
			expected: true,
		},
		{
			name:         "Todo with mergeable merge request",
			todo:         &gitlab.Todo{ActionName: "review_requested", Target: &gitlab.TodoTarget{}},
			mergeRequest: &gitlab.MergeRequest{DetailedMergeStatus: "mergeable"},
			expected:     true,
		},
		{
			name:         "Todo with non-mergeable merge request",
			todo:         &gitlab.Todo{ActionName: "review_requested", Target: &gitlab.TodoTarget{}},
			mergeRequest: &gitlab.MergeRequest{DetailedMergeStatus: "cannot_be_merged"},
			expected:     false,
		},
		{
			name:     "Todo with action name directly_addressed",
			todo:     &gitlab.Todo{ActionName: "directly_addressed", Target: &gitlab.TodoTarget{}},
			expected: false,
		},
		{
			name:         "Todo with action name directly_addressed and mergeable merge request",
			todo:         &gitlab.Todo{ActionName: "directly_addressed", Target: &gitlab.TodoTarget{}},
			mergeRequest: &gitlab.MergeRequest{DetailedMergeStatus: "mergeable"},
			expected:     false,
		},
		{
			name:         "Todo with action name review_requested and merge request status cannot_be_merged",
			todo:         &gitlab.Todo{ActionName: "review_requested", Target: &gitlab.TodoTarget{}},
			mergeRequest: &gitlab.MergeRequest{DetailedMergeStatus: "cannot_be_merged"},
			expected:     false,
		},
		{
			name:         "Todo with action name assigned and merge request status mergeable",
			todo:         &gitlab.Todo{ActionName: "assigned", Target: &gitlab.TodoTarget{}},
			mergeRequest: &gitlab.MergeRequest{DetailedMergeStatus: "mergeable"},
			expected:     false,
		},
		{
			name:         "Todo with action name review_requested and merge request status unknown",
			todo:         &gitlab.Todo{ActionName: "review_requested", Target: &gitlab.TodoTarget{}},
			mergeRequest: &gitlab.MergeRequest{DetailedMergeStatus: "unknown"},
			expected:     false,
		},
		{
			name:         "Todo and merge request are nil",
			todo:         nil,
			mergeRequest: nil,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := shouldMarkTodoAsDone(ctx, tt.todo, tt.mergeRequest)
			assert.Equal(t, tt.expected, result)
		})
	}
}

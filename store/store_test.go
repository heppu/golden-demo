package store_test

import (
	"testing"

	"github.com/heppu/golden-demo/store"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  store.TaskStatus
	}{
		{
			name:  "empty",
			input: "",
			want:  store.StatusUnknown,
		},
		{
			name:  "invalid",
			input: "asdasda",
			want:  store.StatusUnknown,
		},
		{
			name:  "waiting",
			input: "waiting",
			want:  store.StatusWaiting,
		},
		{
			name:  "working",
			input: "working",
			want:  store.StatusWorking,
		},
		{
			name:  "done",
			input: "done",
			want:  store.StatusDone,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := store.ParseStatus(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

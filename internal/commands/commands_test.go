package commands_test

import (
	"testing"
)

func TestMainCommandUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no args",
			args: []string{"satisficer"},
		},
		{
			name: "unknown command",
			args: []string{"satisficer", "unknown"},
		},
		{
			name: "create command with no args",
			args: []string{"satisficer", "create"},
		},
		{
			name: "build command with no args",
			args: []string{"satisficer", "build"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
		})
	}
}

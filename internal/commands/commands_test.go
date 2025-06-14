package commands_test

import (
	"bytes"
	"embed"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/fivethirty/satisficer/internal/commands"
)

//go:embed usage
var usageFS embed.FS

var noOpCommands = func() map[string]*commands.Command {
	result := make(map[string]*commands.Command)
	for name, cmd := range commands.SubCommands {
		result[name] = &commands.Command{
			Name:      cmd.Name,
			UsageText: cmd.UsageText,
			FlagSet:   cmd.FlagSet,
			Validate:  cmd.Validate,
			Run: func() error {
				return nil
			},
		}
	}
	return result
}()

// A lot of these tests cannot be run in parallel because they
// modify the globals. This is a bit of a hack both to avoid
// side effects and to test that stuff got logged.

func TestBadCommand(t *testing.T) {
	commands.SubCommands = noOpCommands
	tests := []struct {
		name      string
		args      []string
		usagePath string
	}{
		{
			name:      "unknown command",
			args:      []string{"satisficer", "unknown"},
			usagePath: "usage/main.txt",
		},
		{
			name:      "main: unknown flag",
			args:      []string{"satisficer", "--unknown"},
			usagePath: "usage/main.txt",
		},
		{
			name:      "sub: not enough arguments",
			args:      []string{"satisficer", "build", "arg1"},
			usagePath: "usage/build.txt",
		},
		{
			name:      "sub: unknown flag",
			args:      []string{"satisficer", "create", "--unknown"},
			usagePath: "usage/create.txt",
		},
		{
			name:      "sub: too many arguments",
			args:      []string{"satisficer", "create", "arg1", "arg2"},
			usagePath: "usage/create.txt",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			commands.DefaultWriter = buf
			err := commands.Execute(test.args)
			if !errors.Is(err, commands.ErrBadCommand) {
				t.Fatalf("expected ErrBadCommand but got %v", err)
			}

			expectedUsage, err := usageFS.ReadFile(test.usagePath)
			if err != nil {
				t.Fatal(err)
			}
			if buf.String() != string(expectedUsage) {
				t.Fatalf(
					"expected usage output to be %q but got %q",
					string(expectedUsage),
					buf.String(),
				)
			}
		})
	}
}

func TestHelp(t *testing.T) {
	commands.SubCommands = noOpCommands
	tests := []struct {
		name      string
		args      []string
		usagePath string
	}{
		{
			name:      "main",
			args:      []string{"satisficer"},
			usagePath: "usage/main.txt",
		},
		{
			name:      "create",
			args:      []string{"satisficer", "create"},
			usagePath: "usage/create.txt",
		},
		{
			name:      "build",
			args:      []string{"satisficer", "build"},
			usagePath: "usage/build.txt",
		},
		{
			name:      "serve",
			args:      []string{"satisficer", "serve"},
			usagePath: "usage/serve.txt",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			allArgs := [][]string{test.args}
			for _, arg := range []string{"-h", "-help", "--help"} {
				args := append(test.args, arg)
				allArgs = append(allArgs, args)
			}

			expectedUsage, err := usageFS.ReadFile(test.usagePath)
			if err != nil {
				t.Fatal(err)
			}
			expectedUsageStr := string(expectedUsage)

			for _, args := range allArgs {
				buf := bytes.NewBuffer(nil)
				commands.DefaultWriter = buf
				err := commands.Execute(args)
				if !errors.Is(err, commands.ErrHelp) || err == nil {
					t.Fatalf("expected ErrHelp but got %v", err)
				}
				if buf.String() != expectedUsageStr {
					t.Fatalf(
						"expected usage output to be %q but got %q",
						expectedUsageStr,
						buf.String(),
					)
				}
			}
		})
	}
}

func TestRealCreate(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "satisficer-test-create")
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	err := commands.Execute(
		[]string{"satisficer", "create", dir},
	)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
}

func TestRealCreateAndBuild(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "project")
	buildDir := filepath.Join(dir, "build")

	err := commands.Execute(
		[]string{"satisficer", "create", projectDir},
	)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	err = commands.Execute(
		[]string{"satisficer", "build", projectDir, buildDir},
	)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
}

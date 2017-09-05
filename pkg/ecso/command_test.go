package ecso

import (
	"fmt"
	"io"
	"testing"
)

func TestCommandFunc(t *testing.T) {
	called := false

	cmd := CommandFunc(func(ctx *CommandContext, r io.Reader, w io.Writer) error {
		called = true
		return nil
	})

	cmd.Execute(nil, nil, nil)

	if !called {
		t.Error("Expect command func to execute")
	}

	err := cmd.Validate(nil)
	if err != nil {
		t.Error(err)
	}
}

func TestCommandError(t *testing.T) {
	err := fmt.Errorf("an error")
	cmd := CommandError(err)
	got := cmd.Execute(nil, nil, nil)

	if err != got {
		t.Errorf("Expected CommandError to return provided error")
	}
}

func TestCommandContext(t *testing.T) {
	project := &Project{}
	preferences := &UserPreferences{}
	version := "1"

	ctx := NewCommandContext(project, preferences, version)

	if project != ctx.Project {
		t.Error("Expected projects to be equal")
	}

	if preferences != ctx.UserPreferences {
		t.Error("Expected preferences to be equal")
	}

	if version != ctx.EcsoVersion {
		t.Error("Expected versions to be equal")
	}
}

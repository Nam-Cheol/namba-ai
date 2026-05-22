package namba

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestCommandBehaviorCharacterizationPublicHelpMatrix(t *testing.T) {
	t.Parallel()

	usage := usageText()
	for _, definition := range publicTopLevelCommandDefinitions() {
		if !strings.Contains(usage, definition.UsageSummary) {
			t.Fatalf("top-level usage missing public command %q summary %q:\n%s", definition.Name, definition.UsageSummary, usage)
		}

		help, ok := commandUsageText(definition.Name)
		if !ok {
			t.Fatalf("expected command help for %q", definition.Name)
		}
		for _, want := range []string{"Usage:", fmt.Sprintf("namba %s", definition.Name)} {
			if !strings.Contains(help, want) {
				t.Fatalf("help for %q missing %q:\n%s", definition.Name, want, help)
			}
		}
	}

	if strings.Contains(usage, internalCreateCommandName) {
		t.Fatalf("top-level usage leaked internal command %q:\n%s", internalCreateCommandName, usage)
	}
	if _, ok := commandUsageText(internalCreateCommandName); ok {
		t.Fatalf("internal command %q should not have public help", internalCreateCommandName)
	}
}

func TestCommandBehaviorCharacterizationUnknownCommand(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	err := NewApp(stdout, &bytes.Buffer{}).Run(context.Background(), []string{"does-not-exist"})
	if err == nil {
		t.Fatal("expected unknown command to fail")
	}
	if !strings.Contains(err.Error(), "unknown command") || !strings.Contains(err.Error(), "NambaAI CLI") {
		t.Fatalf("unexpected unknown command error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("unknown command should not write stdout, got %q", stdout.String())
	}
}

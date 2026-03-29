package namba

import "testing"

func TestResolveVersionFallsBackToDev(t *testing.T) {
	t.Parallel()

	if got := resolveVersion("   "); got != devVersionLabel {
		t.Fatalf("resolveVersion(blank) = %q, want %q", got, devVersionLabel)
	}
}

func TestResolveVersionPreservesInjectedRelease(t *testing.T) {
	t.Parallel()

	if got := resolveVersion("v1.2.3"); got != "v1.2.3" {
		t.Fatalf("resolveVersion(release) = %q, want %q", got, "v1.2.3")
	}
}

func TestFormatVersionLine(t *testing.T) {
	t.Parallel()

	if got := formatVersionLine("v0.1.9"); got != "namba v0.1.9" {
		t.Fatalf("formatVersionLine(version) = %q, want %q", got, "namba v0.1.9")
	}
}

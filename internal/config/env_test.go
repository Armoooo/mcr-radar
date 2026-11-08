package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvLoadsValuesAndPreservesExistingEnvironment(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, ".env")
	if err := os.WriteFile(path, []byte("MCR_RADAR_TEST_ONE=from-file\nMCR_RADAR_TEST_TWO=\"quoted-value\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MCR_RADAR_TEST_ONE", "from-process")
	_ = os.Unsetenv("MCR_RADAR_TEST_TWO")
	t.Cleanup(func() { _ = os.Unsetenv("MCR_RADAR_TEST_TWO") })

	if err := LoadEnv(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("MCR_RADAR_TEST_ONE"); got != "from-process" {
		t.Fatalf("expected existing environment to win, got %q", got)
	}
	if got := os.Getenv("MCR_RADAR_TEST_TWO"); got != "quoted-value" {
		t.Fatalf("expected quoted value to load, got %q", got)
	}
}

package appenv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadExpandsEarlierEnvValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("CONFIG_PATH=/tmp/dashboard\nDASHBOARD_DB_PATH=${CONFIG_PATH}/database.db\n"), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CONFIG_PATH", "")
	t.Setenv("DASHBOARD_DB_PATH", "")
	os.Unsetenv("CONFIG_PATH")
	os.Unsetenv("DASHBOARD_DB_PATH")

	if err := Load(path); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if got := os.Getenv("DASHBOARD_DB_PATH"); got != "/tmp/dashboard/database.db" {
		t.Fatalf("expected expanded db path, got %q", got)
	}
}

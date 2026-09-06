package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCreatesDailyFiles(t *testing.T) {
	dir := t.TempDir()
	logger, err := New(Config{Dir: dir, Timezone: "Asia/Shanghai", RetentionDays: 30})
	if err != nil {
		t.Fatal(err)
	}
	logger.App.Print("test")
	logger.Close()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var app, access bool
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "app-") && filepath.Ext(entry.Name()) == ".log" {
			app = true
		}
		if strings.HasPrefix(entry.Name(), "access-") && filepath.Ext(entry.Name()) == ".log" {
			access = true
		}
	}
	if !app || !access {
		t.Fatalf("daily files missing: app=%v access=%v", app, access)
	}
}

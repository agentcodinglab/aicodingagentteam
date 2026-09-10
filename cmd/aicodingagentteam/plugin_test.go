package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdPlugin_NoArgs_PrintsUsage(t *testing.T) {
	cmdPlugin(nil)
}

func TestCmdPlugin_New_ScaffoldsDriver(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	cmdPlugin([]string{"new", "test-driver", "test"})

	drvPath := filepath.Join("pkg", "plugin", "examples", "test-driver", "driver.go")
	if _, err := os.Stat(drvPath); err != nil {
		t.Fatalf("scaffolded driver not found: %v", err)
	}
}

func TestCmdPlugin_Search_NoIndex(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	cmdPlugin([]string{"search", "test"})
}

func TestCmdPlugin_List_NoIndex(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	_ = os.Chdir(dir)

	cmdPlugin([]string{"list"})
}

func TestScaffoldDriver_CreatesValidGo(t *testing.T) {
	dir := t.TempDir()
	if err := scaffoldDriver(dir, "mydriver", "mybackend"); err != nil {
		t.Fatalf("scaffoldDriver: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "driver.go"))
	if err != nil {
		t.Fatalf("read driver.go: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "package mydriver") {
		t.Error("missing package declaration")
	}
	if !strings.Contains(content, `"mybackend"`) {
		t.Error("missing backend name")
	}
	if !strings.Contains(content, "plugin.Register") {
		t.Error("missing plugin.Register call")
	}
}

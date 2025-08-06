package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestLoadFrom_Success(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	envContent := []byte("DATABASE_URL=db_url\nPORT=1234\nAPP_ENV=ci\n")

	if err := os.WriteFile(envFile, envContent, 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.Unsetenv("DATABASE_URL")
	_ = os.Unsetenv("PORT")
	_ = os.Unsetenv("APP_ENV")

	cfg := LoadFrom(envFile)
	if cfg.DatabaseURL != "db_url" {
		t.Errorf("expected db_url, got %q", cfg.DatabaseURL)
	}
	if cfg.Port != "1234" {
		t.Errorf("expected 1234, got %q", cfg.Port)
	}
	if cfg.AppEnv != "ci" {
		t.Errorf("expected ci, got %q", cfg.AppEnv)
	}
}

func TestTryLoadFrom_Success(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	envContent := []byte("DATABASE_URL=test_url\nPORT=9999\nAPP_ENV=dev\n")

	if err := os.WriteFile(envFile, envContent, 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.Unsetenv("DATABASE_URL")
	_ = os.Unsetenv("PORT")
	_ = os.Unsetenv("APP_ENV")

	cfg, err := TryLoadFrom(envFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DatabaseURL != "test_url" {
		t.Errorf("expected test_url, got %q", cfg.DatabaseURL)
	}
	if cfg.Port != "9999" {
		t.Errorf("expected 9999, got %q", cfg.Port)
	}
	if cfg.AppEnv != "dev" {
		t.Errorf("expected dev, got %q", cfg.AppEnv)
	}
}

func TestTryLoadFrom_MissingRequired(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	envContent := []byte("PORT=8080\nAPP_ENV=test\n")

	if err := os.WriteFile(envFile, envContent, 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.Unsetenv("DATABASE_URL")

	_, err := TryLoadFrom(envFile)
	if err == nil {
		t.Fatal("expected error due to missing DATABASE_URL, got nil")
	}
}

func TestLoad_panicsOnMissingRequiredEnv(t *testing.T) {
	if os.Getenv("TEST_FATAL") == "1" {
		Load() // detta triggar log.Fatalf → os.Exit(1)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestLoad_panicsOnMissingRequiredEnv")
	cmd.Env = append(os.Environ(), "TEST_FATAL=1")
	cmd.Env = append(cmd.Env, "PORT=8080", "APP_ENV=test")         // men ingen DATABASE_URL
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", t.TempDir())) // säkerställ temp .env

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected Load() to exit with error, but it did not. Output:\n%s", output)
	}
}

func TestLoadFrom_panicsOnMissingRequiredEnv(t *testing.T) {
	if os.Getenv("TEST_FATAL_LOADFROM") == "1" {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		content := []byte("PORT=8080\nAPP_ENV=test\n") // Missing DATABASE_URL

		if err := os.WriteFile(envFile, content, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write env file: %v", err)
			os.Exit(2)
		}

		LoadFrom(envFile)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestLoadFrom_panicsOnMissingRequiredEnv")
	cmd.Env = append(os.Environ(), "TEST_FATAL_LOADFROM=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatalf("expected LoadFrom() to exit with error, but it did not. Output:\n%s", output)
	}
}

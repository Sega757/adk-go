package configurable

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFromConfig(t *testing.T) {
	resetRegistries(t)

	t.Run("HappyPathRegisteredAgentClass", func(t *testing.T) {
		resetRegistries(t)
		dir := t.TempDir()
		configPath := filepath.Join(dir, "valid.yaml")
		yamlContent := []byte(`
agent_class: LoopAgent
name: happy_agent
max_iterations: 3
`)
		if err := os.WriteFile(configPath, yamlContent, 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		ag, err := FromConfig(context.Background(), configPath)
		if err != nil {
			t.Fatalf("unexpected error from FromConfig: %v", err)
		}
		if ag == nil {
			t.Fatalf("expected non-nil agent")
		}
		if ag.Name() != "happy_agent" {
			t.Errorf("got agent name %q, want %q", ag.Name(), "happy_agent")
		}
	})

	t.Run("DefaultAgentClassFallback", func(t *testing.T) {
		resetRegistries(t)
		t.Setenv("GEMINI_API_KEY", "dummy-api-key")
		dir := t.TempDir()
		configPath := filepath.Join(dir, "default_class.yaml")
		yamlContent := []byte(`
name: default_llm_agent
model: gemini-1.5-flash
`)
		if err := os.WriteFile(configPath, yamlContent, 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		ag, err := FromConfig(context.Background(), configPath)
		if err != nil {
			t.Fatalf("unexpected error from FromConfig: %v", err)
		}
		if ag == nil {
			t.Fatalf("expected non-nil agent")
		}
		if ag.Name() != "default_llm_agent" {
			t.Errorf("got agent name %q, want %q", ag.Name(), "default_llm_agent")
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		resetRegistries(t)
		dir := t.TempDir()
		nonExistentPath := filepath.Join(dir, "does_not_exist.yaml")
		_, err := FromConfig(context.Background(), nonExistentPath)
		if err == nil {
			t.Fatalf("expected error for non-existent file, got nil")
		}
	})

	t.Run("InvalidYAML", func(t *testing.T) {
		resetRegistries(t)
		dir := t.TempDir()
		invalidPath := filepath.Join(dir, "invalid.yaml")
		if err := os.WriteFile(invalidPath, []byte("invalid: yaml: : content"), 0644); err != nil {
			t.Fatalf("failed to write invalid file: %v", err)
		}
		_, err := FromConfig(context.Background(), invalidPath)
		if err == nil {
			t.Fatalf("expected error for invalid YAML, got nil")
		}
	})

	t.Run("UnregisteredAgentClass", func(t *testing.T) {
		resetRegistries(t)
		dir := t.TempDir()
		unregisteredPath := filepath.Join(dir, "unregistered.yaml")
		yamlContent := []byte(`
agent_class: NonExistentAgentClass
name: test_agent
`)
		if err := os.WriteFile(unregisteredPath, yamlContent, 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}
		_, err := FromConfig(context.Background(), unregisteredPath)
		if err == nil {
			t.Fatalf("expected error for unregistered agent class, got nil")
		}
	})
}

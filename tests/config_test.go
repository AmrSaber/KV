package tests

import (
	"os"
	"path"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func readConfigFile(t *testing.T, tmpDir string) map[string]any {
	t.Helper()

	configPath := path.Join(tmpDir, "kv.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to parse config YAML: %v", err)
	}

	return parsed
}

func TestConfigDefaults(t *testing.T) {
	t.Run("defaults used for missing config values", func(t *testing.T) {
		tmpDir := SetupTestDB(t)
		configPath := path.Join(tmpDir, "kv.yaml")

		if err := os.WriteFile(configPath, []byte("history-length: 10\n"), 0o644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		output := RunKVSuccessStdout(t, "info", "--output", "yaml")

		var info struct {
			DataDir    string `yaml:"data-dir"`
			ConfigPath string `yaml:"config-path"`
			Config     struct {
				PruneHistoryAfterDays int    `yaml:"prune-history-after-days"`
				HistoryLength         int    `yaml:"history-length"`
			} `yaml:"config"`
		}
		if err := yaml.Unmarshal([]byte(output), &info); err != nil {
			t.Fatalf("Failed to parse info output: %v\nOutput:\n%s", err, output)
		}

		if info.Config.PruneHistoryAfterDays != 30 {
			t.Errorf("Expected prune-history-after-days = 30, got %d", info.Config.PruneHistoryAfterDays)
		}
		if info.Config.HistoryLength != 10 {
			t.Errorf("Expected history-length = 10, got %d", info.Config.HistoryLength)
		}
	})

	t.Run("default DB directory is overridden on load", func(t *testing.T) {
		tmpDir := SetupTestDB(t)
		configPath := path.Join(tmpDir, "kv.yaml")

		wrongDir := "/some/custom/path"
		yamlContent := "dbs:\n  default:\n    directory: " + wrongDir + "\n"
		if err := os.WriteFile(configPath, []byte(yamlContent), 0o644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		output := RunKVSuccessStdout(t, "info", "--output", "yaml")

		var info struct {
			DataDir string `yaml:"data-dir"`
			Config  struct {
				DBs map[string]struct {
					Directory string `yaml:"directory"`
				} `yaml:"dbs"`
			} `yaml:"config"`
		}
		if err := yaml.Unmarshal([]byte(output), &info); err != nil {
			t.Fatalf("Failed to parse info output: %v\nOutput:\n%s", err, output)
		}

		defaultDB, ok := info.Config.DBs["default"]
		if !ok {
			t.Fatal("default DB should be present in config")
		}

		if defaultDB.Directory != info.DataDir {
			t.Errorf("Default DB directory should be the data directory (%s), got %s", info.DataDir, defaultDB.Directory)
		}
	})

	t.Run("default DB is removed from config file on first write", func(t *testing.T) {
		tmpDir := SetupTestDB(t)
		configPath := path.Join(tmpDir, "kv.yaml")
		dataDir := path.Join(tmpDir, "kv")

		yamlContent := "dbs:\n  default:\n    directory: /custom/path\n"
		if err := os.WriteFile(configPath, []byte(yamlContent), 0o644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		RunKVSuccess(t, "--db", "mydb", "set", "k", "v")

		parsed := readConfigFile(t, tmpDir)

		dbs, ok := parsed["dbs"]
		if !ok {
			t.Fatal("Config file should contain 'dbs' after creating a custom DB")
		}

		dbMap, ok := dbs.(map[any]any)
		if !ok {
			t.Fatalf("dbs should be a map, got %T", dbs)
		}

		if _, exists := dbMap["default"]; exists {
			t.Error("default DB should not appear in the written config file")
		}

		mydbEntry, exists := dbMap["mydb"]
		if !exists {
			t.Fatal("mydb should appear in the written config file")
		}

		mydbConfig, ok := mydbEntry.(map[any]any)
		if !ok {
			t.Fatalf("mydb entry should be a map, got %T", mydbEntry)
		}

		dir, ok := mydbConfig["directory"]
		if !ok {
			t.Fatal("mydb should have a directory entry")
		}

		if dir != dataDir {
			t.Errorf("Expected mydb directory = %s, got %s", dataDir, dir)
		}
	})

	t.Run("dbs is omitted from config file when no custom DBs exist", func(t *testing.T) {
		tmpDir := SetupTestDB(t)

		RunKVSuccess(t, "--db", "tempdb", "set", "k", "v")

		parsed1 := readConfigFile(t, tmpDir)
		if _, ok := parsed1["dbs"]; !ok {
			t.Fatal("Config should contain 'dbs' after creating a custom DB")
		}

		RunKVSuccess(t, "db", "delete", "tempdb", "--prune")

		configPath := path.Join(tmpDir, "kv.yaml")
		written, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config file after deletion: %v", err)
		}

		if strings.Contains(string(written), "dbs:") {
			t.Error("Config file should not contain 'dbs:' when no custom DBs exist")
		}
	})
}

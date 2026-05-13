package common

import (
	"os"
	"path"

	gap "github.com/muesli/go-app-paths"
	"gopkg.in/yaml.v2"
)

const DefaultDBName = "default"

var cachedConfig *Config

type DBConfig struct {
	Directory string `json:"directory" yaml:"directory"`
}

type Config struct {
	PruneHistoryAfterDays int `json:"pruneHistoryAfterDays,omitempty" yaml:"prune-history-after-days,omitempty"`
	HistoryLength         int `json:"historyLength,omitempty" yaml:"history-length,omitempty"`

	DBs map[string]DBConfig `json:"dbs,omitempty" yaml:"dbs,omitempty"`

	CurrentDB string `json:"-" yaml:"-"`
}

func (config Config) GetDBPath(name string) string {
	dbDirectory := GetDataDirectory()
	if dbConfig, ok := config.DBs[name]; ok {
		dbDirectory = dbConfig.Directory
	}

	return path.Join(dbDirectory, name+".db")
}

func (config Config) GetCurrentDBPath() string {
	return config.GetDBPath(config.CurrentDB)
}

// RegisterDB adds DB if it does not exist, otherwise it's a no-op
func (config *Config) RegisterDB(name string) {
	// Skip existing DBs
	if _, ok := config.DBs[name]; ok {
		return
	}

	config.DBs[name] = DBConfig{Directory: GetDataDirectory()}
	config.write()
}

func (config *Config) SetDBDirectory(db string, directory string) {
	Assert(db != DefaultDBName, "Cannot update directory for default DB")

	dbConfig := config.DBs[db]
	dbConfig.Directory = directory
	config.DBs[db] = dbConfig

	config.write()
}

func (config *Config) RenameDB(db string, newName string) {
	Assert(db != DefaultDBName, "Cannot rename default DB")

	if _, ok := config.DBs[db]; !ok {
		Fail("%q DB does not exist", db)
	}

	if _, ok := config.DBs[newName]; ok {
		Fail("%q DB already exists", db)
	}

	config.DBs[newName] = config.DBs[db]
	delete(config.DBs, db)

	config.write()
}

func (config *Config) DeleteDB(db string) {
	Assert(db != DefaultDBName, "Cannot delete default DB")

	if _, ok := config.DBs[db]; !ok {
		Fail("%q DB does not exist", db)
	}

	delete(config.DBs, db)

	config.write()
}

// Write config to storage
func (config *Config) write() {
	originalDBs := config.DBs

	// Clone DBs and remove the default DB
	config.DBs = make(map[string]DBConfig)
	for key, value := range originalDBs {
		if key != DefaultDBName {
			config.DBs[key] = value
		}
	}

	err := os.WriteFile(GetConfigPath(), []byte(config.String()), 0o644)
	FailOn(err)

	config.DBs = originalDBs
}

func (config Config) String() string {
	output, err := yaml.Marshal(config)
	FailOn(err)

	return string(output)
}

func GetConfig() *Config {
	if cachedConfig != nil {
		return cachedConfig
	}

	cachedConfig = new(getDefaultConfig())

	configPath := GetConfigPath()
	configBytes, err := os.ReadFile(configPath)
	if !os.IsNotExist(err) {
		FailOn(err)

		err = yaml.Unmarshal(configBytes, cachedConfig)
		if err != nil {
			Fail("Invalid config YAML: %v", err)
		}
	}

	// Inject default DB
	cachedConfig.DBs[DefaultDBName] = DBConfig{Directory: GetDataDirectory()}

	// Set current DB
	cachedConfig.CurrentDB = DefaultDBName
	if envDB, found := os.LookupEnv("KV_DB"); found {
		cachedConfig.CurrentDB = envDB
	}

	ValidateDBName(cachedConfig.CurrentDB)

	return cachedConfig
}

func getDefaultConfig() Config {
	return Config{
		PruneHistoryAfterDays: 30,
		HistoryLength:         15,

		DBs: map[string]DBConfig{},
	}
}

func GetDataDirectory() string {
	scope := gap.NewScope(gap.User, "kv")

	dataDir, err := scope.DataPath("")
	FailOn(err)

	return dataDir
}

func GetConfigPath() string {
	scope := gap.NewScope(gap.User, "")

	// e.g. /home/some-user/.config/kv.yaml on linux
	configPath, err := scope.ConfigPath("kv.yaml")
	FailOn(err)

	return configPath
}

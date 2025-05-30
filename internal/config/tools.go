package config

type ToolConfig struct {
	Name      string   `yaml:"name"`
	ScriptURL string   `yaml:"script"`
	Actions   []string `yaml:"actions"`
	OS        []string `yaml:"os"`
}

type ToolsConfig struct {
	Tools []ToolConfig `yaml:"tools"`
}

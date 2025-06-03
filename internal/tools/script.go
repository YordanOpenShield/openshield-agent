package tools

type ScriptTool struct {
	*Tool

	script string `yaml:"script"`
}

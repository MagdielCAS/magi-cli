package project

// MagiConfig represents the .magi.yaml configuration file.
type MagiConfig struct {
	Actions      []Action               `mapstructure:"actions" yaml:"actions"`
	RulesPath    string                 `mapstructure:"rules_path" yaml:"rules_path"`
    Architecture string                 `mapstructure:"architecture" yaml:"architecture"`
    ProjectType  string                 `mapstructure:"project_type" yaml:"project_type"`
    Remaining    map[string]interface{} `mapstructure:",remain" yaml:",inline"`
}

// Action defines a project-specific action (e.g., creating a slice, service, etc.).
type Action struct {
	Name        string            `mapstructure:"name" yaml:"name"`
	Description string            `mapstructure:"description" yaml:"description"`
	Parameters  []ActionParameter `mapstructure:"parameters" yaml:"parameters"`
    Steps       []ActionStep      `mapstructure:"steps" yaml:"steps"`
}

// ActionParameter defines a parameter for an action.
type ActionParameter struct {
	Name        string `mapstructure:"name" yaml:"name"`
	Description string `mapstructure:"description" yaml:"description"`
	Type        string `mapstructure:"type" yaml:"type"` // string, boolean, etc.
    Required    bool   `mapstructure:"required" yaml:"required"`
}

// ActionStep defines a specific step in an action workflow.
type ActionStep struct {
    Tool        string            `mapstructure:"tool" yaml:"tool"`               // create_file, edit_file, etc.
    Instruction string            `mapstructure:"instruction" yaml:"instruction"` // What to do
    Parameters  map[string]string `mapstructure:"parameters" yaml:"parameters"`   // Additional static params if needed
}

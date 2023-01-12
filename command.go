package spec

type action string

type Command struct {
	Name            string            `yaml:"name"`
	Aliases         []string          `yaml:"aliases,omitempty"`
	Description     string            `yaml:"description,omitempty"`
	Group           string            `yaml:"group,omitempty"`
	Flags           map[string]string `yaml:"flags,omitempty"`
	PersistentFlags map[string]string `yaml:"persistentflags,omitempty"`
	Completion      struct {
		Flag          map[string][]action `yaml:"flag,omitempty"`
		Positional    [][]action          `yaml:"positional,omitempty"`
		PositionalAny []action            `yaml:"positionalany,omitempty"`
		Dash          [][]action          `yaml:"dash,omitempty"`
		DashAny       []action            `yaml:"dashany,omitempty"`
	} `yaml:"completion,omitempty"`
	Commands []Command `yaml:"commands,omitempty"`
}

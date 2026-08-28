package models

type DefaultConfig struct {
	Output      OutputConfig      `json:"output"`
	FromAliases map[string]string `json:"from_aliases"`
}

func NewDefaultConfig() *DefaultConfig {
	return &DefaultConfig{
		Output: OutputConfig{
			Color:     "blue",
			ShowStats: true,
		},
		FromAliases: make(map[string]string),
	}
}

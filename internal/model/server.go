package model

// Server represents an SSH server configuration.
type Server struct {
	Name string   `yaml:"name"`
	Host string   `yaml:"host"`
	User string   `yaml:"user"`
	Port int      `yaml:"port"`
	Key  string   `yaml:"key,omitempty"`
	Tags []string `yaml:"tags,omitempty"`
}

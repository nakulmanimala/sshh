package model

// Server represents an SSH server configuration.
type Server struct {
	Name string   `yaml:"name"`
	Host string   `yaml:"host"`
	User string   `yaml:"user"`
	Port int      `yaml:"port"`
	Key  string   `yaml:"key,omitempty"`
	Tags []string `yaml:"tags,omitempty"`
	// AWSManaged marks a server as having been added or IP-updated by AWS
	// sync, so a later sync can detect when its EC2 instance disappears.
	AWSManaged bool `yaml:"aws_managed,omitempty"`
}

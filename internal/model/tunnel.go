package model

// TunnelType represents the SSH port-forwarding mode.
type TunnelType string

const (
	TunnelLocal   TunnelType = "local"
	TunnelRemote  TunnelType = "remote"
	TunnelDynamic TunnelType = "dynamic"
)

// Tunnel represents a saved SSH tunnel template.
type Tunnel struct {
	Name       string     `yaml:"name"`
	SSHHost    string     `yaml:"ssh_host"`
	SSHUser    string     `yaml:"ssh_user,omitempty"`
	SSHPort    int        `yaml:"ssh_port,omitempty"`
	SSHKey     string     `yaml:"ssh_key,omitempty"`
	Type       TunnelType `yaml:"type"`
	LocalPort  int        `yaml:"local_port"`
	RemoteHost string     `yaml:"remote_host,omitempty"`
	RemotePort int        `yaml:"remote_port,omitempty"`
}

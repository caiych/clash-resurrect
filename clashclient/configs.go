package clashclient

import (
	"context"
)

// Configs defines scheme of clash API /configs.
type Configs struct {
	Port       int `json:"port"`
	SocksPort  int `json:"socks-port"`
	RedirPort  int `json:"redir-port"`
	TproxyPort int `json:"tproxy-port"`
	MixedPort  int `json:"mixed-port"`

	// Authentication
	AllowLan      bool   `json:"allow-lan"`
	BindAddress   string `json:"bind-address"`
	Mode          string `json:"mode"`      // TODO: enum
	LogLevel      string `json:"log-level"` // TODO: enum
	Ipv6          bool   `json:"ipv6"`
	InterfaceName string `json:"interface-name"`
}

// GetConfigs requests configs from clash.
func (c *Client) GetConfigs(ctx context.Context) (*Configs, error) {
	i, err := c.genericRequest(ctx, clashAPI{
		path:      "/configs",
		typeValue: &Configs{},
	})
	if err != nil {
		return nil, err
	}
	return i.(*Configs), nil
}

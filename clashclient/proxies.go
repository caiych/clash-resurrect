package clashclient

import "context"

// Proxies defines proxy config.
type Proxies struct {
	Proxies map[string]Proxy `json:"proxies"`
}

// Proxy defines a single proxy.
type Proxy struct {
	ProxyType string   `json:"type"`
	Child     []string `json:"all"`
	Current   string   `json:"now"`
}

// GetProxies gets current proxy config.
func (c *Client) GetProxies(ctx context.Context) (*Proxies, error) {
	i, err := c.genericRequest(ctx, clashAPI{
		path:      "/proxies",
		typeValue: &Proxies{},
	})
	if err != nil {
		return nil, err
	}
	return i.(*Proxies), nil
}

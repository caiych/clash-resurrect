package clashclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"gopkg.in/resty.v1"
)

// Client encapsulates a client of clash API.
type Client struct {
	Host string
	Port int
}

type clashAPI struct {
	path      string
	typeValue interface{}
}

type rootResponse struct {
	Hello string `json:"hello"`
}

// GetRoot sends requests to root path of clash API.
func (c *Client) GetRoot(ctx context.Context) error {
	i, err := c.genericRequest(ctx, clashAPI{
		path:      "/",
		typeValue: &rootResponse{},
	})
	if err != nil {
		return err
	}
	if i.(*rootResponse).Hello != "clash" {
		return errors.New("what?")
	}
	return nil
}

func (c *Client) genericRequest(ctx context.Context, api clashAPI) (interface{}, error) {
	resp, err := resty.R().SetContext(ctx).Get(c.url(api.path))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resp.Body(), api.typeValue); err != nil {
		return nil, err
	}
	return api.typeValue, nil
}

func (c *Client) url(path string) string {
	u := &url.URL{
		Scheme: "http",
		Path:   path,
		Host:   fmt.Sprintf("localhost:%d", c.Port),
	}
	return u.String()
}

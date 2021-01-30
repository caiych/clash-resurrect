package clashclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRoot(t *testing.T) {
	testcases := []struct {
		name string
		fn   http.HandlerFunc
		err  error
	}{
		{
			name: "ok",
			fn: func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte(`{"hello":"clash"}`))
			},
			err: nil,
		},
		{
			name: "timeout",
			fn: func(rw http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Second * 10)
				rw.Write([]byte(`{"hello":"clash"}`))
			},
			err: context.DeadlineExceeded,
		},
		{
			name: "invalid json",
			fn: func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte(`{"broken":`))
			},
			err: errors.New("unexpected end of JSON input"),
		},
		{
			name: "unexpected",
			fn: func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte(`{"hello":"clashX"}`))
			},
			err: errors.New("what?"),
		},
	}
	ctx := context.Background()
	for _, tc := range testcases {
		func() {
			ts := httptest.NewServer(tc.fn)
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Error(err)
			}
			port, err := strconv.Atoi(u.Port())

			if err != nil {
				t.Error(err)
			}
			c := &Client{
				Host: u.Hostname(),
				Port: port,
			}
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			if err := c.GetRoot(ctx); !strings.Contains(fmt.Sprintf("%v", err), fmt.Sprintf("%v", tc.err)) {
				t.Errorf("Testcase %v:\nExpected: %v\n Actual: %v", tc.name, tc.err, err)
			}
		}()
	}
}

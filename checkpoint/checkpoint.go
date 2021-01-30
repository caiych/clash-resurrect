package checkpoint

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/caiych/clash-resurrect/clashclient"
	log "github.com/sirupsen/logrus"
)

// Checkpoint models what's persistent on disk.
type Checkpoint struct {
	Proxies *clashclient.Proxies
	MTime   time.Time
}

// Read reads Checkpoint from given dir.
func Read(ctx context.Context, dir string) *Checkpoint {
	cleanSlate := &Checkpoint{
		Proxies: nil,
		MTime:   time.Now(),
	}
	p := proxiesPath(dir)
	// Check the stats, return if error(e.g. not exists)
	fs, err := os.Stat(p)
	if err != nil {
		return cleanSlate
	}

	f, err := os.Open(p)
	if err != nil {
		return cleanSlate
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return cleanSlate
	}

	ps := &clashclient.Proxies{}
	if err := json.Unmarshal(b, ps); err != nil {
		return cleanSlate
	}

	return &Checkpoint{
		Proxies: ps,
		MTime:   fs.ModTime(),
	}
}

// Write writes checkpoint to given dir.
func Write(ctx context.Context, dir string, ckpt *Checkpoint) {
	p := proxiesPath(dir)
	f, err := os.Create(p)
	if err != nil {
		log.Info(err)
	}
	defer f.Close()
	b, err := json.Marshal(ckpt.Proxies)
	if err != nil {
		log.Info(err)
	}
	f.Write(b)
}

func Update(ctx context.Context, c *clashclient.Client, ckpt *Checkpoint) {
	for p, v := range ckpt.Proxies.Proxies {
		if v.ProxyType != "Selector" {
			continue
		}
		// TODO: check if the childrens are the same.
		c.UpdateProxySelection(ctx, p, v.Current)
	}
}

func proxiesPath(dir string) string {
	return filepath.Join(dir, "proxies.json")
}

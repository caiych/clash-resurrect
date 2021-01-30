package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caiych/clash-resurrect/clashclient"
	"github.com/cakturk/go-netstat/netstat"
)

var (
	clashPort     = flag.Int("clash_api_port", 9090, "api port for clash.")
	checkpointDir = flag.String("checkpoint_dir", "/tmp/clash-resurrect/", "Dir for checkpoints.")
)

func main() {
	flag.Parse()

	if err := prepareCheckpointDir(); err != nil {
		log.Fatal(err)
	}
	if err := mainLoop(); err != nil {
		log.Fatal(err)
	}
}

func mainLoop() error {
	c := clashclient.Client{
		Port: *clashPort,
	}
	ctx := context.Background()
	tick := time.Tick(time.Minute)
	for {
		err := func() error {
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			if err := c.GetRoot(ctx); err != nil {
				if err := killClash(*clashPort); err != nil {
					log.Printf("Killing clash error: %v", err)
				}
			}
			return nil
		}()
		if err != nil {
			return err
		}
		<-tick
	}
}

func prepareCheckpointDir() error {
	p := *checkpointDir
	s, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(p, os.FileMode(0750))
		}
		return err
	}
	if !s.IsDir() {
		return fmt.Errorf("path %s exists as a file", p)
	}
	return nil
}

func killClash(port int) error {
	ss, err := netstat.TCP6Socks(func(ste *netstat.SockTabEntry) bool {
		return ste.LocalAddr.Port == uint16(port)
	})
	if err != nil {
		return err
	}
	pids := make(map[int]bool)
	for _, p := range ss {
		pids[p.Process.Pid] = true
	}
	if len(pids) != 1 {
		return fmt.Errorf("unexpected number of processes with port %d: %v", port, len(pids))
	}

	for pid := range pids {
		p, err := os.FindProcess(pid)
		if err != nil {
			return err
		}
		return p.Kill()
	}
	return nil
}

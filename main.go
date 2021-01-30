package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/caiych/clash-resurrect/checkpoint"
	"github.com/caiych/clash-resurrect/clashclient"
	"github.com/cakturk/go-netstat/netstat"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/prometheus/procfs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

var (
	clashPort     = flag.Int("clash_api_port", 9090, "api port for clash.")
	checkpointDir = flag.String("checkpoint_dir", "/tmp/clash-resurrect/", "Dir for checkpoints.")
)

func main() {
	flag.Parse()

	setupLogger()

	if err := prepareCheckpointDir(); err != nil {
		log.Fatal(err)
	}

	if err := mainLoop(); err != nil {
		log.Fatal(err)
	}
}

func mainLoop() error {
	c := clashclient.Client{
		Host: "localhost",
		Port: *clashPort,
	}
	ctx := context.Background()
	tick := time.Tick(10 * time.Second)
	for {
		err := func() error {
			{
				ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				if err := c.GetRoot(ctx); err != nil {
					if err := killClash(*clashPort); err != nil {
						log.Infof("Killing clash error: %v", err)
					}
				}
			}

			liveProxies, err := c.GetProxies(ctx)
			if err != nil {
				log.Infof("GetProxies failed, skipping: %v", err)
				return nil
			}
			ckpt := checkpoint.Read(ctx, *checkpointDir)

			liveTimestamp, err := clashStartupTime()
			if err != nil {
				log.Warningf("Couldn't decide clash startup time, err: %v", err)
				return nil
			}
			if liveTimestamp.Before(ckpt.MTime) {
				log.Infof("There's no server restart event after last checkpoint, saving checkpoint.")
				checkpoint.Write(ctx, *checkpointDir, &checkpoint.Checkpoint{
					Proxies: liveProxies,
					MTime:   time.Now(),
				})
				return nil
			}
			log.Warningf("Checkpoint(%v) is appears to be older than live config(%v(server restarted recently), updating.", ckpt.MTime, liveTimestamp)
			checkpoint.Update(ctx, &c, ckpt)
			checkpoint.Write(ctx, *checkpointDir, &checkpoint.Checkpoint{
				Proxies: ckpt.Proxies,
				MTime:   time.Now(),
			})
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

func findProcessByPort(port int) (*os.Process, error) {
	ss, err := netstat.TCP6Socks(func(ste *netstat.SockTabEntry) bool {
		return ste.LocalAddr.Port == uint16(port)
	})
	if err != nil {
		return nil, err
	}
	pids := make(map[int]bool)
	for _, p := range ss {
		if p.Process == nil {
			continue
		}
		pids[p.Process.Pid] = true
	}
	if len(pids) != 1 {
		return nil, fmt.Errorf("unexpected number of processes with port %d: %v", port, len(pids))
	}

	for pid := range pids {
		return os.FindProcess(pid)
	}
	return nil, nil
}

func killClash(port int) error {
	p, err := findProcessByPort(port)
	if err != nil {
		return err
	}
	return p.Kill()
}

func clashStartupTime() (time.Time, error) {
	p, err := findProcessByPort(*clashPort)
	if err != nil {
		return time.Time{}, err
	}
	pc, err := procfs.NewProc(p.Pid)
	if err != nil {
		return time.Time{}, err
	}
	s, err := pc.Stat()
	if err != nil {
		return time.Time{}, err
	}
	t, err := s.StartTime()
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(t), 0), nil
}

func setupLogger() {
	p := filepath.Join(*checkpointDir, "LOG")
	w, err := rotatelogs.New(
		p+".%Y%m%d",
		rotatelogs.WithLinkName(p),
		rotatelogs.WithMaxAge(time.Hour*24*7),
		rotatelogs.WithRotationTime(time.Hour*24*7),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.AddHook(lfshook.NewHook(w, &log.TextFormatter{}))
}

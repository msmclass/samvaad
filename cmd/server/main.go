// Copyright 2026 Samvaad Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/msmclass/samvaad/pkg/samvaad/logger"
	"github.com/msmclass/samvaad/pkg/samvaad/tracer/jaeger"

	"github.com/msmclass/samvaad/pkg/config"
	"github.com/msmclass/samvaad/pkg/routing"
	"github.com/msmclass/samvaad/pkg/rtc"
	"github.com/msmclass/samvaad/pkg/service"
	"github.com/msmclass/samvaad/pkg/telemetry/prometheus"
	"github.com/msmclass/samvaad/version"
)

var baseFlags = []cli.Flag{
	&cli.StringSliceFlag{
		Name:  "bind",
		Usage: "IP address to listen on, use flag multiple times to specify multiple addresses",
	},
	&cli.StringFlag{
		Name:  "config",
		Usage: "path to Samvaad config file",
	},
	&cli.StringFlag{
		Name:    "config-body",
		Usage:   "Samvaad config in YAML, typically passed in as an environment var in a container",
		Sources: cli.EnvVars("SAMVAAD_CONFIG"),
	},
	&cli.StringFlag{
		Name:  "key-file",
		Usage: "path to file that contains API keys/secrets",
	},
	&cli.StringFlag{
		Name:    "keys",
		Usage:   "api keys (key: secret\\n)",
		Sources: cli.EnvVars("SAMVAAD_KEYS"),
	},
	&cli.StringFlag{
		Name:    "region",
		Usage:   "region of the current node. Used by regionaware node selector",
		Sources: cli.EnvVars("SAMVAAD_REGION"),
	},
	&cli.StringFlag{
		Name:    "node-ip",
		Usage:   "IP address of the current node, used to advertise to clients. Automatically determined by default",
		Sources: cli.EnvVars("NODE_IP"),
	},
	&cli.StringFlag{
		Name:    "udp-port",
		Usage:   "UDP port(s) to use for WebRTC traffic",
		Sources: cli.EnvVars("UDP_PORT"),
	},
	&cli.StringFlag{
		Name:    "redis-host",
		Usage:   "host (incl. port) to redis server",
		Sources: cli.EnvVars("REDIS_HOST"),
	},
	&cli.StringFlag{
		Name:    "redis-password",
		Usage:   "password to redis",
		Sources: cli.EnvVars("REDIS_PASSWORD"),
	},
	&cli.StringFlag{
		Name:    "turn-cert",
		Usage:   "tls cert file for TURN server",
		Sources: cli.EnvVars("SAMVAAD_TURN_CERT"),
	},
	&cli.StringFlag{
		Name:    "turn-key",
		Usage:   "tls key file for TURN server",
		Sources: cli.EnvVars("SAMVAAD_TURN_KEY"),
	},
	&cli.StringFlag{
		Name:  "cpuprofile",
		Usage: "write CPU profile to `file`",
	},
	&cli.StringFlag{
		Name:  "memprofile",
		Usage: "write memory profile to `file`",
	},
	&cli.BoolFlag{
		Name:  "dev",
		Usage: "sets log-level to debug, console formatter, and /debug/pprof. insecure for production",
	},
	&cli.BoolFlag{
		Name:   "disable-strict-config",
		Usage:  "disables strict config parsing",
		Hidden: true,
	},
}

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	defer func() {
		if rtc.Recover(logger.GetLogger()) != nil {
			os.Exit(1)
		}
	}()

	generatedFlags, err := config.GenerateCLIFlags(baseFlags, true)
	if err != nil {
		fmt.Println(err)
	}

	cmd := &cli.Command{
		Name:        "samvaad-server",
		Usage:       "High performance SFU for Samvaad video conference by MSM Class (https://msmclass.in)",
		Description: "run without subcommands to start the server. Powered by MSM Class.",
		Flags:       append(baseFlags, generatedFlags...),
		Action:      startServer,
		Commands: []*cli.Command{
			{
				Name:   "generate-keys",
				Usage:  "generates an API key and secret pair",
				Action: generateKeys,
			},
			{
				Name:   "ports",
				Usage:  "print ports that server is configured to use",
				Action: printPorts,
			},
			{
				// this subcommand is deprecated, token generation is provided by CLI
				Name:   "create-join-token",
				Hidden: true,
				Usage:  "create a room join token for development use",
				Action: createToken,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "room",
						Usage:    "name of room to join",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "identity",
						Usage:    "identity of participant that holds the token",
						Required: true,
					},
					&cli.BoolFlag{
						Name:     "recorder",
						Usage:    "creates a hidden participant that can only subscribe",
						Required: false,
					},
				},
			},
			{
				Name:   "list-nodes",
				Usage:  "list all nodes",
				Action: listNodes,
			},
			{
				Name:   "help-verbose",
				Usage:  "prints app help, including all generated configuration flags",
				Action: helpVerbose,
			},
		},
		Version: version.Version,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Println(err)
	}
}

func getConfig(c *cli.Command) (*config.Config, error) {
	confString, err := getConfigString(c.String("config"), c.String("config-body"))
	if err != nil {
		return nil, err
	}

	strictMode := !c.Bool("disable-strict-config")

	conf, err := config.NewConfig(confString, strictMode, c, baseFlags)
	if err != nil {
		return nil, err
	}
	config.InitLoggerFromConfig(&conf.Logging)

	if conf.Development {
		logger.Infow("starting in development mode")

		if len(conf.Keys) == 0 {
			logger.Infow("no keys provided, using placeholder keys",
				"API Key", "devkey",
				"API Secret", "secret",
			)
			conf.Keys = map[string]string{
				"devkey": "secret",
			}
			shouldMatchRTCIP := false
			// when dev mode and using shared keys, we'll bind to localhost by default
			if conf.BindAddresses == nil {
				conf.BindAddresses = []string{
					"127.0.0.1",
					"::1",
				}
			} else {
				// if non-loopback addresses are provided, then we'll match RTC IP to bind address
				// our IP discovery ignores loopback addresses
				for _, addr := range conf.BindAddresses {
					ip := net.ParseIP(addr)
					if ip != nil && !ip.IsLoopback() && !ip.IsUnspecified() {
						shouldMatchRTCIP = true
					}
				}
			}
			if shouldMatchRTCIP {
				for _, bindAddr := range conf.BindAddresses {
					conf.RTC.IPs.Includes = append(conf.RTC.IPs.Includes, bindAddr+"/24")
				}
			}
		}
	}
	return conf, nil
}

func startServer(ctx context.Context, c *cli.Command) error {
	conf, err := getConfig(c)
	if err != nil {
		return err
	}
	if url := conf.Trace.JaegerURL; url != "" {
		jaeger.Configure(ctx, url, "samvaad")
	}

	// Apply low-resource mode tuning before any subsystem initialisation so that
	// GC pressure is bounded from the very first allocation.
	applyLowResourceMode(conf)

	// validate API key length
	err = conf.ValidateKeys()
	if err != nil {
		return err
	}

	if cpuProfile := c.String("cpuprofile"); cpuProfile != "" {
		if f, err := os.Create(cpuProfile); err != nil {
			return err
		} else {
			if err := pprof.StartCPUProfile(f); err != nil {
				f.Close()
				return err
			}
			defer func() {
				pprof.StopCPUProfile()
				f.Close()
			}()
		}
	}

	if memProfile := c.String("memprofile"); memProfile != "" {
		if f, err := os.Create(memProfile); err != nil {
			return err
		} else {
			defer func() {
				// run memory profile at termination
				runtime.GC()
				_ = pprof.WriteHeapProfile(f)
				_ = f.Close()
			}()
		}
	}

	currentNode, err := routing.NewLocalNode(conf)
	if err != nil {
		return err
	}

	if err := prometheus.Init(string(currentNode.NodeID()), currentNode.NodeType()); err != nil {
		return err
	}

	server, err := service.InitializeServer(conf, currentNode)
	if err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		for i := range 2 {
			sig := <-sigChan
			force := i > 0
			logger.Infow("exit requested, shutting down", "signal", sig, "force", force)
			go server.Stop(force)
		}
	}()

	return server.Start()
}

func getConfigString(configFile string, inConfigBody string) (string, error) {
	if inConfigBody != "" || configFile == "" {
		return inConfigBody, nil
	}

	outConfigBody, err := os.ReadFile(configFile)
	if err != nil {
		return "", err
	}

	return string(outConfigBody), nil
}

// applyLowResourceMode enforces hard GC and memory limits when
// low_resource_mode is enabled. It also propagates reduced packet buffer sizes
// from the LowResource config block into the active RTC config so that every
// downstream buffer factory picks up the smaller values.
//
// This function is intentionally idempotent and safe to call more than once.
func applyLowResourceMode(conf *config.Config) {
	if !conf.LowResourceMode {
		return
	}

	lr := conf.LowResource

	// --- Memory limit ---------------------------------------------------------
	if lr.MemLimitMiB > 0 {
		limitBytes := int64(lr.MemLimitMiB) * 1024 * 1024
		debug.SetMemoryLimit(limitBytes)
		logger.Infow("[low-resource] hard memory limit set",
			"limit_mib", lr.MemLimitMiB,
			"limit_bytes", limitBytes,
		)
	}

	// --- GC target ------------------------------------------------------------
	if lr.GCPercent >= 0 {
		old := debug.SetGCPercent(lr.GCPercent)
		logger.Infow("[low-resource] GC percent set",
			"new_pct", lr.GCPercent,
			"old_pct", old,
		)
	}

	// --- Packet buffer sizes --------------------------------------------------
	// Only override if the operator has not already set a smaller value manually.
	if lr.PacketBufferSizeVideo > 0 && conf.RTC.PacketBufferSizeVideo > lr.PacketBufferSizeVideo {
		logger.Infow("[low-resource] reducing video packet buffer",
			"from", conf.RTC.PacketBufferSizeVideo,
			"to", lr.PacketBufferSizeVideo,
		)
		conf.RTC.PacketBufferSizeVideo = lr.PacketBufferSizeVideo
	}
	if lr.PacketBufferSizeAudio > 0 && conf.RTC.PacketBufferSizeAudio > lr.PacketBufferSizeAudio {
		logger.Infow("[low-resource] reducing audio packet buffer",
			"from", conf.RTC.PacketBufferSizeAudio,
			"to", lr.PacketBufferSizeAudio,
		)
		conf.RTC.PacketBufferSizeAudio = lr.PacketBufferSizeAudio
	}

	logger.Infow("[low-resource] mode active — Samvaad running in sovereign edge mode 🇮🇳")
}


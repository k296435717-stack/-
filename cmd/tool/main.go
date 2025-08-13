package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var version = "0.1.0"

// simple task registry
var tasks = map[string]func(cfg *Config) error{
	"build": func(cfg *Config) error {
		logInfo(cfg, "task(build) start")
		time.Sleep(500 * time.Millisecond)
		logInfo(cfg, "task(build) done")
		return nil
	},
	"status": func(cfg *Config) error {
		out := map[string]any{"version": version, "time": time.Now().Format(time.RFC3339)}
		if cfg.Output == "json" {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(out)
		}
		fmt.Printf("Version:\t%s\nTime:\t%s\n", out["version"], out["time"])
		return nil
	},
}

type Config struct {
	ConfigPath string
	LogLevel   string
	Output     string
	Debug      bool
}

func main() {
	if len(os.Args) == 1 {
		printHelp()
		return
	}

	cmd := os.Args[1]
	if cmd == "--help" || cmd == "-h" || cmd == "help" {
		printHelp()
		return
	}
	if cmd == "--version" || cmd == "-V" {
		fmt.Println(version)
		return
	}

	// subcommand dispatch
	switch cmd {
	case "run":
		cfg, task, err := parseRunArgs(os.Args[2:])
		if err != nil {
			logError(nil, err.Error())
			os.Exit(1)
		}
		if task == "" {
			logError(cfg, "missing task name")
			os.Exit(1)
		}
		fn, ok := tasks[task]
		if !ok {
			logError(cfg, "unknown task: "+task)
			os.Exit(1)
		}
		logInfo(cfg, fmt.Sprintf("loading config: %s", cfg.ConfigPath))
		if err := fn(cfg); err != nil {
			logError(cfg, err.Error())
			os.Exit(1)
		}
		return
	default:
		logError(nil, "unknown command: "+cmd)
		printHelp()
		os.Exit(1)
	}
}

func parseRunArgs(args []string) (*Config, string, error) {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	cfg := &Config{}
	fs.StringVar(&cfg.ConfigPath, "config", "tool.yaml", "配置文件路径")
	fs.StringVar(&cfg.Output, "output", "", "输出格式(json)")
	fs.BoolVar(&cfg.Debug, "debug", false, "调试模式")
	if err := fs.Parse(args); err != nil {
		return nil, "", err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		return cfg, "", nil
	}
	return cfg, rest[0], nil
}

func printHelp() {
	help := `tool ` + version + `
用法:
  tool run <task> [选项]

命令:
  run         运行指定任务 (build, status)

选项:
  --config    指定配置文件 (默认: tool.yaml)
  --output    输出格式 (json)
  --debug     显示调试日志
  -h, --help  显示帮助
  -V, --version 显示版本
`
	fmt.Print(help)
}

func logInfo(cfg *Config, msg string)  { fmt.Println("[INFO]", msg) }
func logError(cfg *Config, msg string) { fmt.Fprintln(os.Stderr, "[ERROR]", msg) }

// simulate config load (placeholder)
func loadConfig(path string) error {
	if path == "" {
		return errors.New("empty path")
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

func expandPath(p string) string {
	if strings.HasPrefix(p, "~") {
		h, _ := os.UserHomeDir()
		return filepath.Join(h, strings.TrimPrefix(p, "~"))
	}
	return p
}

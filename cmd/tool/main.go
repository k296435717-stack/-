package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
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
	case "game":
		playGuessingGame()
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
  tool game            运行猜数字小游戏

命令:
  run         运行指定任务 (build, status)
  game        启动简单的猜数字游戏

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

func playGuessingGame() {
	rand.Seed(time.Now().UnixNano())
	target := rand.Intn(100) + 1
	reader := bufio.NewReader(os.Stdin)
	attempts := 0

	fmt.Println("🎲 猜数字游戏：请在 1 到 100 之间猜一个数字！")
	fmt.Println("输入 quit 退出游戏。最多 10 次尝试。")

	for {
		fmt.Print("请输入你的猜测: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			logError(nil, "读取输入失败")
			return
		}

		guessText := strings.TrimSpace(line)
		if strings.EqualFold(guessText, "quit") {
			fmt.Println("游戏结束，下次再来！")
			return
		}
		if guessText == "" {
			fmt.Println("请输入一个数字。")
			continue
		}

		guess, err := strconv.Atoi(guessText)
		if err != nil {
			fmt.Println("请输入有效的数字。")
			continue
		}

		attempts++
		if guess < target {
			fmt.Println("太小了，再试试。")
		} else if guess > target {
			fmt.Println("太大了，再试试。")
		} else {
			fmt.Printf("恭喜！你在 %d 次尝试内猜中了数字 %d！\n", attempts, target)
			return
		}

		if attempts >= 10 {
			fmt.Printf("达到最大尝试次数，正确答案是 %d。\n", target)
			return
		}
	}
}

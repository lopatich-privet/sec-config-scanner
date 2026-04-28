package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/lopatich-privet/sec-config-scanner/internal/analyzer"
	"github.com/lopatich-privet/sec-config-scanner/internal/output"
	"github.com/lopatich-privet/sec-config-scanner/internal/parser"
	"github.com/lopatich-privet/sec-config-scanner/internal/rules"
	"github.com/lopatich-privet/sec-config-scanner/server/grpc"
	httpserver "github.com/lopatich-privet/sec-config-scanner/server/http"

	flag "github.com/spf13/pflag"
)

type CLIConfig struct {
	silent       bool
	useStdin     bool
	useDirectory bool
	serverMode   bool
	grpcMode     bool
	serverPort   string
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Анализатор конфигурационных файлов на предмет уязвимостей безопасности\n\n")
		fmt.Fprintf(os.Stderr, "Использование:\n")
		fmt.Fprintf(os.Stderr, "  config-analyzer [флаги] <путь_к_файлу>\n")
		fmt.Fprintf(os.Stderr, "  config-analyzer [флаги] --dir <путь_к_директории>\n")
		fmt.Fprintf(os.Stderr, "  config-analyzer [флаги] --stdin\n")
		fmt.Fprintf(os.Stderr, "  config-analyzer --server [--port <порт>]\n")
		fmt.Fprintf(os.Stderr, "  config-analyzer --grpc [--port <порт>]\n\n")
		fmt.Fprintf(os.Stderr, "Флаги:\n")
		flag.PrintDefaults()
	}
}

func parseFlags() *CLIConfig {
	var cfg CLIConfig
	flag.BoolVarP(&cfg.silent, "silent", "s", false, "не выходить с ошибкой при наличии проблем")
	flag.BoolVar(&cfg.useStdin, "stdin", false, "прочитать конфигурацию из стандартного потока ввода")
	flag.BoolVar(&cfg.useDirectory, "dir", false, "прочитать все конфигурации из директории (рекурсивно)")
	flag.BoolVar(&cfg.serverMode, "server", false, "запустить HTTP сервер")
	flag.BoolVar(&cfg.grpcMode, "grpc", false, "запустить gRPC сервер")
	flag.StringVar(&cfg.serverPort, "port", "8080", "порт для сервера (по умолчанию: 8080)")
	flag.Parse()
	return &cfg
}

func runHTTPServer(port string) error {
	server := httpserver.NewServer(port)
	if err := server.Start(); err != nil {
		return fmt.Errorf("ошибка запуска HTTP сервера: %w", err)
	}
	return nil
}

func runGRPCServer(port string) error {
	server := grpc.NewServer(port)
	if err := server.Start(); err != nil {
		return fmt.Errorf("ошибка запуска gRPC сервера: %w", err)
	}
	return nil
}

func runDirectoryMode(dir string, silent bool) error {
	configs, err := parser.ParseDirectory(dir)
	if err != nil {
		return fmt.Errorf("ошибка парсинга директории: %w", err)
	}

	analyzerInstance := analyzer.NewAnalyzer(rules.GetFileModeRules())
	var allIssues []rules.Issue

	for _, config := range configs {
		issues := analyzerInstance.Analyze(config.Data)
		allIssues = append(allIssues, issues...)
	}

	out := output.NewOutput(allIssues)
	out.Print()

	if out.HasIssues() && !silent {
		os.Exit(1)
	}
	return nil
}

func parseConfig(useStdin bool, filePath string) (*parser.Config, error) {
	if useStdin {
		return parser.ParseFromStdin()
	}
	return parser.ParseFile(filePath)
}

func runSingleConfigMode(useStdin bool, filePath string, silent bool) error {
	config, err := parseConfig(useStdin, filePath)
	if err != nil {
		return err
	}

	analyzerInstance := analyzer.NewAnalyzer(rules.GetFileModeRules())
	issues := analyzerInstance.Analyze(config.Data)

	out := output.NewOutput(issues)
	out.Print()

	if out.HasIssues() && !silent {
		os.Exit(1)
	}
	return nil
}

func validateDirectoryArgs(args []string) {
	if len(args) != 0 {
		return
	}
	slog.Error("использование: config-analyzer --dir укажите путь к директории")
	flag.PrintDefaults()
	os.Exit(1)
}

func validateFileArgs(args []string) {
	if len(args) != 0 {
		return
	}
	slog.Error("использование: config-analyzer [--silent] [--stdin] укажите путь к файлу")
	flag.PrintDefaults()
	os.Exit(1)
}

func getFilePath(args []string, useStdin bool) string {
	if useStdin {
		return ""
	}
	return args[0]
}

func runClientMode(cfg *CLIConfig, args []string) error {
	if cfg.useDirectory {
		validateDirectoryArgs(args)
		return runDirectoryMode(args[0], cfg.silent)
	}

	if !cfg.useStdin {
		validateFileArgs(args)
	}
	filePath := getFilePath(args, cfg.useStdin)

	return runSingleConfigMode(cfg.useStdin, filePath, cfg.silent)
}

func main() {
	cfg := parseFlags()

	var err error
	switch {
	case cfg.serverMode:
		err = runHTTPServer(cfg.serverPort)
	case cfg.grpcMode:
		err = runGRPCServer(cfg.serverPort)
	default:
		args := flag.Args()
		err = runClientMode(cfg, args)
	}

	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

package main

import (
	"config-analyzer/internal/analyzer"
	"config-analyzer/internal/output"
	"config-analyzer/internal/parser"
	"config-analyzer/internal/rules"
	"config-analyzer/server/grpc"
	httpserver "config-analyzer/server/http"
	"fmt"
	"log/slog"
	"os"

	flag "github.com/spf13/pflag"
)

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

func main() {
	var silent bool
	var useStdin bool
	var useDirectory bool
	var serverMode bool
	var grpcMode bool
	var serverPort string

	flag.BoolVarP(&silent, "silent", "s", false, "не выходить с ошибкой при наличии проблем")
	flag.BoolVar(&useStdin, "stdin", false, "прочитать конфигурацию из стандартного потока ввода")
	flag.BoolVar(&useDirectory, "dir", false, "прочитать все конфигурации из директории (рекурсивно)")
	flag.BoolVar(&serverMode, "server", false, "запустить HTTP сервер")
	flag.BoolVar(&grpcMode, "grpc", false, "запустить gRPC сервер")
	flag.StringVar(&serverPort, "port", "8080", "порт для сервера (по умолчанию: 8080)")
	flag.Parse()

	// Режим HTTP сервера
	if serverMode {
		server := httpserver.NewServer(serverPort)
		if err := server.Start(); err != nil {
			slog.Error("ошибка запуска HTTP сервера", "error", err)
			os.Exit(1)
		}
		return
	}

	// Режим gRPC сервера
	if grpcMode {
		server := grpc.NewServer(serverPort)
		if err := server.Start(); err != nil {
			slog.Error("ошибка запуска gRPC сервера", "error", err)
			os.Exit(1)
		}
		return
	}

	args := flag.Args()

	// Режим директории
	if useDirectory {
		if len(args) == 0 {
			slog.Error("использование: config-analyzer --dir укажите путь к директории")
			flag.PrintDefaults()
			os.Exit(1)
		}

		configs, err := parser.ParseDirectory(args[0])
		if err != nil {
			slog.Error("ошибка парсинга директории", "error", err)
			os.Exit(1)
		}

		// Анализируем каждую конфигурацию с проверкой file permissions
		analyzerInstance := analyzer.NewAnalyzer(rules.GetFileModeRules())
		var allIssues []rules.Issue

		for _, config := range configs {
			issues := analyzerInstance.Analyze(config.Data)
			allIssues = append(allIssues, issues...)
		}

		// Выводим результат
		out := output.NewOutput(allIssues)
		out.Print()

		// Если есть проблемы и не включён silent режим, выходим с ошибкой
		if out.HasIssues() && !silent {
			os.Exit(1)
		}
		return
	}

	// Режим stdin или файл
	var config *parser.Config
	var err error

	if useStdin {
		config, err = parser.ParseFromStdin()
		if err != nil {
			slog.Error("ошибка чтения из stdin", "error", err)
			os.Exit(1)
		}
	} else {
		if len(args) == 0 {
			slog.Error("использование: config-analyzer [--silent] [--stdin] укажите путь к файлу")
			flag.PrintDefaults()
			os.Exit(1)
		}

		config, err = parser.ParseFile(args[0])
		if err != nil {
			slog.Error("ошибка парсинга файла", "error", err)
			os.Exit(1)
		}
	}

	// Создаём анализатор с проверкой file permissions для файлов
	analyzerInstance := analyzer.NewAnalyzer(rules.GetFileModeRules())

	// Анализируем конфиг
	issues := analyzerInstance.Analyze(config.Data)

	// Выводим результат
	out := output.NewOutput(issues)
	out.Print()

	// Если есть проблемы и не включён silent режим, выходим с ошибкой
	if out.HasIssues() && !silent {
		os.Exit(1)
	}
}

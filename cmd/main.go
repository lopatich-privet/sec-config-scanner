package main

import (
	"config-analyzer/internal/analyzer"
	"config-analyzer/internal/output"
	"config-analyzer/internal/parser"
	"config-analyzer/internal/rules"
	"config-analyzer/server/grpc"
	httpserver "config-analyzer/server/http"
	"log/slog"
	"os"

	flag "github.com/spf13/pflag"
)

func main() {
	var silent bool
	var useStdin bool
	var serverMode bool
	var grpcMode bool
	var serverPort string

	flag.BoolVarP(&silent, "silent", "s", false, "не выходить с ошибкой при наличии проблем")
	flag.BoolVar(&useStdin, "stdin", false, "прочитать конфигурацию из стандартного потока ввода")
	flag.BoolVar(&serverMode, "server", false, "запустить HTTP сервер")
	flag.BoolVar(&grpcMode, "grpc", false, "запустить gRPC сервер")
	flag.StringVar(&serverPort, "port", "8080", "порт для сервера")
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

	// Создаем анализатор с дефолтными правилами
	analyzerInstance := analyzer.NewAnalyzer(rules.GetDefaultRules())

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

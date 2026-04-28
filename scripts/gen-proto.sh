#!/bin/bash
# Генерация proto файлов
protoc --go_out=. --go_opt=module=github.com/lopatich-privet/sec-config-scanner \
	--go-grpc_out=. --go-grpc_opt=module=github.com/lopatich-privet/sec-config-scanner \
	api/analyzer.proto

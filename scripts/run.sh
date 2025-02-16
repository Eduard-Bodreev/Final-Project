#!/bin/bash

set -e

echo "Проверяем, установлен ли Go..."
if ! command -v go &> /dev/null
then
    echo "Go не найден! Установите Go и повторите попытку."
    exit 1
fi

echo "Запускаем сервер..."
go run ./cmd/main.go
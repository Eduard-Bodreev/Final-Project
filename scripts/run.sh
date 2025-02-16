#!/bin/bash

set -e

echo "Проверяем, установлен ли Go..."
if ! command -v go &> /dev/null; then
    echo "Go не найден! Установите Go и повторите попытку."
    exit 1
fi

echo "Запускаем сервер..."
go run ./cmd/main.go &
SERVER_PID=$!

echo "Ждем, что сервер запустился..."
sleep 30

echo "Проверяем доступность API..."
if ! curl -s --fail http://localhost:8080/health; then
    echo "✗ API сервер не отвечает. Возможно, он не запущен."
    cat server.log  # Вывод лога, если API не отвечает
    exit 1
fi

echo "Сервер запущен с PID $SERVER_PID"
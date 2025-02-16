#!/bin/bash

set -e  # Прекращать выполнение при ошибке

echo "Устанавливаем PostgreSQL..."
sudo apt update
sudo apt install -y postgresql postgresql-contrib unzip curl

echo "Запускаем PostgreSQL..."
sudo systemctl start postgresql
sudo systemctl enable postgresql

echo "Ожидаем запуск PostgreSQL..."
until pg_isready -h localhost -p 5432 -U postgres; do
    sleep 1
done
echo "PostgreSQL запущен!"

echo "Создаём базу данных и пользователя..."
sudo -u postgres psql <<EOF
CREATE DATABASE project_sem_1;
CREATE USER validator WITH PASSWORD 'val1dat0r';
GRANT ALL PRIVILEGES ON DATABASE project_sem_1 TO validator;
ALTER DATABASE project_sem_1 OWNER TO validator;
\c project_sem_1
CREATE TABLE IF NOT EXISTS prices (
    id SERIAL PRIMARY KEY,
    created_at DATE NOT NULL,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    price NUMERIC NOT NULL
);
EOF

echo "Настройка завершена!"
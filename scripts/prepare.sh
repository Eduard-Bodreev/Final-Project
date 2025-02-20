#!/bin/bash

echo "Устанавливаем зависимости Go..."
go mod tidy

DB_HOST=${POSTGRES_HOST:-localhost}
DB_PORT=${POSTGRES_PORT:-5432}
DB_USER=${POSTGRES_USER:-validator}
DB_PASSWORD=${POSTGRES_PASSWORD:-val1dat0r}
DB_NAME=${POSTGRES_DB:-project-sem-1}

echo "Ожидаем доступности PostgreSQL..."
until PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c '\q' 2>/dev/null; do
  >&2 echo "PostgreSQL ещё недоступен, ждём..."
  sleep 3
done
echo "PostgreSQL доступен!"

psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE project-sem-1 TO validator;"
psql -U postgres -d project-sem-1 -c "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO validator;"
psql -U postgres -d project-sem-1 -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO validator;"

echo "Создаём базу данных, если её нет..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME <<EOF
CREATE TABLE IF NOT EXISTS prices (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    price NUMERIC NOT NULL,
    create_date DATE NOT NULL
);
EOF
echo "Настройка завершена!"
#!/bin/bash

# Установка зависимостей для проекта
echo "Installing project dependencies..."
go mod tidy

# Проверка, установлен ли PostgreSQL, и его установка
echo "Checking if PostgreSQL is installed..."
if ! command -v psql &> /dev/null
then
    echo "PostgreSQL not found. Installing..."
    # Для macOS
    if [[ "$OSTYPE" == "darwin"* ]]; then
        brew install postgresql
    # Для Ubuntu/Debian
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        sudo apt update
        sudo apt install -y postgresql postgresql-contrib
    fi
else
    echo "PostgreSQL is already installed."
fi

# Запуск PostgreSQL если он не запущен
echo "Starting PostgreSQL..."
if ! pg_isready -q; then
    echo "PostgreSQL is not running. Starting it now..."
    sudo service postgresql start
else
    echo "PostgreSQL is already running."
fi

# Проверка, что база данных доступна
echo "Checking database connection..."
export PGPASSWORD=val1dat0r
psql -h localhost -U validator -d project-sem-1 -c '\q' > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "Database is accessible."
else
    echo "Database is not accessible. Please check the connection details."
    exit 1
fi

# Дальнейшая настройка приложения
echo "Preparing the application..."
# Дополнительные шаги по подготовке базы данных или проекта

# Завершение скрипта
echo "Preparation completed."
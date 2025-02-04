#!/bin/bash

# Выводим информацию о запуске Go сервера
echo "Starting Go server in background..."

# Запускаем Go сервер в фоновом режиме
go run main.go &

# Получаем PID последнего запущенного процесса (Go-сервера)
PID=$!

# Выводим информацию о процессе
echo "Go server started with PID: $PID"

# Выводим сообщение, что сервер работает в фоне, и скрипт продолжит выполнение
echo "Server is running in background, continuing with the pipeline..."

# Дополнительные задачи, если они необходимы
# Например, вы можете добавить команды для тестирования или других проверок

# Завершаем скрипт
echo "Run script completed."
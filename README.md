# Финальный проект 1 семестра

REST API сервис для загрузки и выгрузки данных о ценах.

## Требования к системе

- **OC**: Windows, Linux, macOS.
- **Go**: ^1.20.
- **PostgreSQL**: ^13.
- **Аппаратные требования**: 1Gb RAM 200Mb HDD

## Установка и запуск

1. Установить PostgreSQL.
2. Создать БД psql -U validator -d postgres -c "CREATE DATABASE project-sem-1;"
3. ./prepare.sh
4. ./run.sh

## Тестирование

Пример директории sample_data — это разархивированная версия файла sample_data.zip.

Для тестирования используйте скрипт tests.sh:

./tests.sh

Этот скрипт выполняет два теста:

Тестирование POST-запроса с файлом sample_data.zip.
Тестирование GET-запроса и проверка загрузки файла data.zip.

## Контакт

К кому можно обращаться в случае вопросов @bel0ruz

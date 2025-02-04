#!/bin/bash
go mod tidy
psql -U validator -d project-sem-1 -f db/migrations.sql
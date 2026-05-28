#!/bin/bash
echo "Menghapus semua data perjalanan..."
cd backend || exit
DB_HOST=localhost DB_PORT=5433 DB_USER=perjadin_user DB_PASSWORD=Yuxt2MaJxNhpfvJzCyNEbVTG4JgtpwXH DB_NAME=perjadin_db go run cmd/clear_perjadin/main.go
echo "Selesai!"

package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func Koneksi() *sql.DB {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/schoolmanagement")
	if err != nil {
		log.Println("Gagal koneksi ke database:", err)
		return nil
	}
	fmt.Println("Koneksi Berhasil")
	return db
}

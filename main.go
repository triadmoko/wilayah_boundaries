package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/lib/pq"              // PostgreSQL driver

	"github.com/joho/godotenv"
)

func main() {
	// Get command line arguments
	args := os.Args[1:]

	var keys = make(map[string]string)
	for _, arg := range args {
		split := strings.Split(arg, "=")
		keys[split[0]] = split[1]
	}
	for key, value := range keys {
		switch key {
		case "db":
			switch value {
			case "postgres":
				log.Println("Executing Postgres")
				ExecutePostgres()
			case "mysql":
				log.Println("Executing MySQL")
				ExecuteMySQL()
			}
		}
	}
}
func ExecutePostgres() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	var files []string
	err = filepath.Walk("./db/postgresql", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".sql" {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("failed to walk directory: %s", err)
	}

	// Buat koneksi ke database menggunakan env
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}
	defer db.Close()

	for _, file := range files {
		sqlScript, err := os.ReadFile(file)
		if err != nil {
			log.Printf("failed reading file %s: %s", file, err)
			continue
		}

		// Eksekusi script SQL
		_, err = db.Exec(string(sqlScript))
		if err != nil {
			log.Printf("failed executing SQL from file %s: %s", file, err)
			continue
		}
		fmt.Printf("Successfully executed SQL script from file: %s\n", file)
	}
}

func ExecuteMySQL() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	var files []string
	err = filepath.Walk("./db/mysql", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".sql" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("failed to walk directory: %s", err)
	}

	// Buat koneksi ke database menggunakan env
	dbURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}
	defer db.Close()

	for _, file := range files {
		sqlScript, err := os.ReadFile(file)
		if err != nil {
			log.Printf("failed reading file %s: %s", file, err)
			continue
		}
		// Split SQL into individual statements and execute each one
		statements := strings.Split(string(sqlScript), ");")
		for i, stmt := range statements {
			if strings.TrimSpace(stmt) == "" {
				continue
			}

			// Add back the closing parenthesis and semicolon
			if i < len(statements)-1 {
				stmt += ");"
			}

			_, err = db.Exec(stmt)
			if err != nil {
				log.Printf("Error executing statement: %v", err)
				continue
			}
		}
		fmt.Printf("Successfully executed SQL script from file: %s\n", file)
	}
}

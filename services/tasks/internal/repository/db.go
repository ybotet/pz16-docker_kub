package repository

import (
    "database/sql"
    "fmt"
    "os"
    
    _ "github.com/lib/pq"
)

func NewPostgresConnection() (*sql.DB, error) {
    host := os.Getenv("DB_HOST")
    if host == "" {
        host = "postgres"  // Cambiado de "localhost" a "postgres"
    }
    port := os.Getenv("DB_PORT")
    if port == "" {
        port = "5432"
    }
    user := os.Getenv("DB_USER")
    if user == "" {
        user = "tasksuser"
    }
    password := os.Getenv("DB_PASSWORD")
    if password == "" {
        password = "taskspass"
    }
    dbname := os.Getenv("DB_NAME")
    if dbname == "" {
        dbname = "tasksdb"
    }
    sslmode := os.Getenv("DB_SSLMODE")
    if sslmode == "" {
        sslmode = "disable"
    }
    
    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        host, port, user, password, dbname, sslmode)
    
    fmt.Printf("Conectando a PostgreSQL: %s\n", connStr) // Para debug
    
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("error opening database: %w", err)
    }
    
    err = db.Ping()
    if err != nil {
        return nil, fmt.Errorf("error connecting to database: %w", err)
    }
    
    return db, nil
}

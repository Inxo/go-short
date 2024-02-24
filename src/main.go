package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var (
	baseUrl      = os.Getenv("BASE_URL")
	databaseFile = os.Getenv("DATABASE_FILE")
)

func main() {
	// Проверяем, существует ли файл базы данных, если нет, то создаем его
	if _, err := os.Stat("data/database.db"); os.IsNotExist(err) {
		db, err := sql.Open("sqlite3", databaseFile)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// Создаем таблицу для хранения коротких и длинных URL
		_, err = db.Exec("CREATE TABLE urls (id INTEGER PRIMARY KEY, short TEXT, long TEXT)")
		if err != nil {
			log.Fatal(err)
		}
	}

	http.HandleFunc("/", redirectHandler)
	http.HandleFunc("/shorten", shortenHandler)
	http.ListenAndServe(":8080", nil)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	// Открываем соединение с базой данных
	db, err := sql.Open("sqlite3", "data/database.db")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Database connection error:", err)
		return
	}
	defer db.Close()

	// Получаем короткий URL из пути запроса
	shortURL := strings.TrimPrefix(r.URL.Path, "/")
	if shortURL == "" {
		http.NotFound(w, r)
		return
	}

	// Ищем длинный URL в базе данных по короткому URL
	var longURL string
	err = db.QueryRow("SELECT long FROM urls WHERE short = ?", shortURL).Scan(&longURL)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Перенаправляем на длинный URL
	http.Redirect(w, r, longURL, http.StatusFound)
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем длинный URL из формы
	longURL := r.FormValue("url")
	if longURL == "" {
		http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
		return
	}

	// Открываем соединение с базой данных
	db, err := sql.Open("sqlite3", "data/database.db")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Database connection error:", err)
		return
	}
	defer db.Close()

	// Генерируем короткий URL случайным образом
	shortURL := generateShortID()

	// Вставляем длинный URL в базу данных
	_, err = db.Exec("INSERT INTO urls (short, long) VALUES (?, ?)", shortURL, longURL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Database insertion error:", err)
		return
	}

	// Формируем короткий URL и отправляем его клиенту
	shortenedURL := baseUrl + shortURL
	fmt.Fprintf(w, "Shortened URL: %s", shortenedURL)
}

func generateShortID() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	shortURL := make([]byte, 6)
	for i := range shortURL {
		shortURL[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(shortURL)
}

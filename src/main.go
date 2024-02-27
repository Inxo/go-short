package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var (
	baseUrl      = os.Getenv("BASE_URL")
	databaseFile = os.Getenv("DATABASE_FILE")
)
var templates = template.Must(template.ParseFiles("templates/stats.html"))

type URLInfo struct {
	Short  string
	Long   string
	Clicks int
}

func statsHandler(w http.ResponseWriter, _ *http.Request) {
	// Открываем соединение с базой данных
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Database connection error:", err)
		return
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Println("Database close error:", err)
		}
	}(db)

	// Получаем данные о коротких ссылках из базы данных
	rows, err := db.Query("SELECT short, long, clicks FROM urls")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Database query error:", err)
		return
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println("rows close error:", err)
		}
	}(rows)

	// Создаем слайс для хранения информации о коротких ссылках
	var urlInfoList []URLInfo

	// Считываем данные из базы данных
	for rows.Next() {
		var urlInfo URLInfo
		if err := rows.Scan(&urlInfo.Short, &urlInfo.Long, &urlInfo.Clicks); err != nil {
			log.Println("Database scan error:", err)
			continue
		}
		urlInfoList = append(urlInfoList, urlInfo)
	}

	// Рендерим шаблон
	if err := templates.ExecuteTemplate(w, "stats.html", urlInfoList); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Template rendering error:", err)
	}
}

func main() {
	if len(databaseFile) == 0 {
		databaseFile = "data/database.db"
	}
	if len(baseUrl) == 0 {
		baseUrl = "http://localhost:8080/"
	}
	// Проверяем, существует ли файл базы данных, если нет, то создаем его
	if _, err := os.Stat(databaseFile); os.IsNotExist(err) {
		db, err := sql.Open("sqlite3", databaseFile)
		if err != nil {
			log.Fatal(err)
		}
		defer func(db *sql.DB) {
			err := db.Close()
			if err != nil {
				log.Println("Database close error:", err)
			}
		}(db)

		// Создаем таблицу для хранения коротких и длинных URL
		_, err = db.Exec("CREATE TABLE urls (id INTEGER PRIMARY KEY, short TEXT, long TEXT, clicks INTEGER DEFAULT 0)")
		if err != nil {
			log.Fatal(err)
		}
	}

	http.HandleFunc("/", redirectHandler)
	http.HandleFunc("/shorten", shortenHandler)
	// Добавляем обработчик для страницы статистики
	http.HandleFunc("/stats", statsHandler)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	// Открываем соединение с базой данных
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Database connection error:", err)
		return
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Println("Database close error:", err)
		}
	}(db)

	// Получаем короткий URL из пути запроса
	shortURL := strings.TrimPrefix(r.URL.Path, "/")
	if shortURL == "" {
		http.NotFound(w, r)
		return
	}
	// Обновляем счетчик переходов
	_, err = db.Exec("UPDATE urls SET clicks = clicks + 1 WHERE short = ?", shortURL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Database update error:", err)
		return
	}

	// Ищем длинный URL в базе данных по короткому URL
	var longURL string
	err = db.QueryRow("SELECT long FROM urls WHERE short = ?", shortURL).Scan(&longURL)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	unUrl, err := url.QueryUnescape(longURL)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Перенаправляем на длинный URL
	http.Redirect(w, r, unUrl, http.StatusFound)
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
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Database connection error:", err)
		return
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Println("Database close error:", err)
		}
	}(db)

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
	_, err = fmt.Fprintf(w, "Shortened URL: %s", shortenedURL)
	if err != nil {
		return
	}
}

func generateShortID() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	shortURL := make([]byte, 6)
	for {
		for i := range shortURL {
			shortURL[i] = alphabet[rand.Intn(len(alphabet))]
		}
		// Проверяем уникальность короткой ссылки
		if !shortURLExists(string(shortURL)) {
			break
		}
	}
	return string(shortURL)
}

func shortURLExists(shortURL string) bool {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		log.Println("Database connection error:", err)
		return true // В случае ошибки соединения, считаем, что короткая ссылка уже существует
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Println("Database close error:", err)
		}
	}(db)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM urls WHERE short = ?", shortURL).Scan(&count)
	if err != nil {
		log.Println("Database query error:", err)
		return true // В случае ошибки запроса, считаем, что короткая ссылка уже существует
	}

	return count > 0
}

package sqlite

// Импортируем необходимые пакеты для работы с базой данных SQLite и обработки ошибок.
// В этом коде:
import (
	"database/sql"                   // Стандартный пакет для работы с базами данных SQL в Go. Он предоставляет интерфейс для работы с любыми базами данных, поддерживающими SQL.
	"errors"                         // Стандартный пакет для работы с ошибками. Мы будем использовать его для создания и проверки ошибок.
	"fmt"                            // Стандартный пакет для форматированного вывода. Он используется для вывода строк, чисел и других данных в консоль.
	"url-shortener/internal/storage" // Пакет приложения, вероятно, содержит структуры и функции для работы с хранилищем данных.

	"github.com/mattn/go-sqlite3" // Внешний пакет для работы с SQLite. Он реализует драйвер для подключения Go-программы к базе данных SQLite.
)

// Storage - структура, которая представляет хранилище данных для работы с базой данных SQLite.
// В этой структуре содержится ссылка на объект типа *sql.DB, который используется для взаимодействия с базой данных.
// В этом коде:
type Storage struct {
	db *sql.DB // db - указатель на объект базы данных, который предоставляет интерфейс для выполнения SQL-запросов.
}

// New - функция, которая создает новое хранилище данных для работы с SQLite.
// Она открывает соединение с базой данных, создает таблицу и индекс, если они не существуют, и возвращает новый экземпляр Storage.
// В этом коде:
func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New" // Определяем строку, которая будет использоваться для указания контекста в сообщении об ошибке.

	// Открываем соединение с базой данных SQLite, используя путь к файлу базы данных.
	// sql.Open открывает базу данных и возвращает объект *sql.DB, который используется для взаимодействия с базой данных.
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		// Если ошибка при открытии базы данных, возвращаем ошибку с контекстом, добавленным с помощью %w.
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Готовим SQL-запрос для создания таблицы, если она не существует.
	// Этот запрос создает таблицу url с полями id, alias и url.
	// id - первичный ключ для уникальной идентификации каждой записи.
	// alias - текстовое поле, уникальное, не может быть пустым.
	// url - текстовое поле для хранения оригинального URL, не может быть пустым.
	// Создаем индекс для alias для ускорения поиска по этому полю.
	stmt, err := db.Prepare(
		`CREATE TABLE IF NOT EXISTS url(
			id INTEGER PRIMARY KEY,    
			alias TEXT NOT NULL UNIQUE, 
			url TEXT NOT NULL);         
			CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);) 
	`)

	if err != nil {
		// Если ошибка при подготовке запроса, возвращаем ошибку с контекстом.
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Выполняем подготовленный запрос для создания таблицы и индекса.
	_, err = stmt.Exec()
	if err != nil {
		// Если ошибка при выполнении запроса, возвращаем ошибку с контекстом.
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Возвращаем новый экземпляр Storage с открытым соединением db.
	return &Storage{db: db}, nil
}

// SaveURL - метод, который сохраняет новый URL в базу данных с уникальным псевдонимом.
// Он выполняет SQL-запрос для добавления записи в таблицу `url`, а затем возвращает ID вставленной строки или ошибку, если она возникла.
// В этом коде:
func (s *Storage) SaveURL(urlToSave, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL" // Определяем строку для контекста ошибки, которая будет добавлена к ошибке, если она произойдет.

	// Готовим SQL-запрос для вставки нового URL и псевдонима в таблицу `url`.
	// Используем `?` для параметризированных запросов, чтобы избежать SQL-инъекций.
	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		// Если не удалось подготовить запрос, возвращаем ошибку с контекстом.
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Выполняем подготовленный запрос, передавая urlToSave и alias в качестве параметров.
	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		// Проверяем, если ошибка связана с нарушением уникальности псевдонима.
		// Если ошибка типа sqlite3.Error и код ошибки соответствует уникальному ограничению,
		// то возвращаем ошибку с контекстом, что URL уже существует.
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		// Если ошибка другая, возвращаем её с контекстом.
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Получаем ID последней вставленной записи в базе данных.
	id, err := res.LastInsertId()
	if err != nil {
		// Если не удалось получить ID вставленной строки, возвращаем ошибку с контекстом.
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	// Возвращаем ID вставленной записи.
	return id, nil
}

// GetURL - метод, который извлекает URL по псевдониму из базы данных.
// Он выполняет SQL-запрос для получения URL, связанного с заданным псевдонимом, и возвращает его или ошибку.
// В этом коде:
func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL" // Строка, определяющая контекст ошибки для удобства отладки.

	// Готовим SQL-запрос для выборки URL по псевдониму.
	// Используем параметризированный запрос для предотвращения SQL-инъекций.
	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		// Если не удалось подготовить запрос, возвращаем ошибку с контекстом.
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	// Выполняем запрос и пытаемся получить результат в переменную resURL.
	var resURL string
	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		// Если ошибок не связаны с отсутствием строк, то возвращаем ошибку с контекстом.
		if errors.Is(err, sql.ErrNoRows) {
			// Если строки не найдены, возвращаем ошибку, что URL с таким псевдонимом не найден.
			return "", storage.ErrURLNotFound
		}
		// В случае других ошибок, возвращаем ошибку с контекстом.
		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	// Если URL найден, возвращаем его.
	return resURL, nil
}

// func (s *Storage) DeleteURL(alias string) error {
	func (s *Storage) DeleteURL(alias string) (int64, error) {
		const fn = "storage.sqlite.DeleteURL"
	
		result, err := s.db.Exec("DELETE FROM url WHERE alias = ?", alias)
		if err != nil {
			return 0, fmt.Errorf("%s: execute statement %w", fn, err)
		}
	
		rowsAffected, err := result.RowsAffected() // Считаем сколько удалили
		if err != nil {
			return 0, fmt.Errorf("%s: get rows affected: %w", fn, err) // Возвращаем 0 и ошибку
		}
	
		return rowsAffected, nil
	}
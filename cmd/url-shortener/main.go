package main

import (
	// Пакет log/slog используется для логирования
	"log/slog"
	"net/http"

	// Пакет os предоставляет функции для работы с операционной системой (например, чтение переменных окружения)
	"os"
	// Импортируем модуль конфигурации приложения
	"url-shortener/internal/config"
	// Импортируем middleware (промежуточный обработчик) для логирования HTTP-запросов
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/save"
	mwLogger "url-shortener/internal/http-server/middleware/logger"

	// Импортируем кастомный обработчик логирования slogpretty для красивого форматирования логов
	"url-shortener/internal/lib/logger/handlers/slogpretty"
	// Импортируем вспомогательный пакет sl для работы с логами
	"url-shortener/internal/lib/logger/sl"
	// Импортируем пакет для работы с хранилищем SQLite
	"url-shortener/internal/storage/sqlite"
	// Импортируем роутер chi v5 для работы с HTTP-маршрутизацией
	"github.com/go-chi/chi/v5"
	// Импортируем middleware из chi для различных вспомогательных функций (например, логирования, восстановления после паники)
	"github.com/go-chi/chi/v5/middleware"
)

const (
	// Определяем строковые константы для различных сред выполнения приложения
	envLocal = "local" // Локальная среда разработки (используется при запуске на локальном компьютере)
	envDev   = "dev"   // Среда для разработки (может использоваться на удалённом сервере для тестирования)
	envProd  = "prod"  // Продакшен-среда (используется в боевом окружении)
)

func main() {
	// TODO: init config: cleanenv

	// Вызываем функцию MustLoad из пакета config.
	// Она загружает конфигурацию приложения (например, из файла или переменных окружения).
	// MustLoad() - это функция, которая, скорее всего, вызывает log.Fatal при ошибке,
	// поэтому нам не нужно проверять ошибку отдельно.
	cfg := config.MustLoad()

	// TODO: init logger: slog

	// Вызываем функцию setupLogger, передавая в неё переменную среды cfg.Env.
	// setupLogger – это кастомная функция, которая настраивает логгер в зависимости от среды (local, dev, prod).
	// Возвращает объект log, который мы будем использовать для логирования событий.
	log := setupLogger(cfg.Env)

	// Вызываем метод Info у объекта log.
	// log.Info() – это метод логгера, который записывает информационное сообщение.
	// slog.String("env", cfg.Env) – добавляет в лог строковый параметр "env" со значением из конфигурации.
	log.Info("starting url-shortener", slog.String("env", cfg.Env), slog.String("version", "123"))

	// Вызываем метод Debug у логгера log.
	// log.Debug() записывает отладочное сообщение, но оно будет видно только если включён debug-уровень логирования.
	log.Debug("debug messages are enabled")

	// Вызываем метод Error у логгера log.
	// log.Error() записывает сообщение об ошибке.
	log.Error("error messages are enabled")

	// TODO: init storage: sqlLite

	// Вызываем функцию sqlite.New(), передавая путь к файлу базы данных из конфигурации.
	// sqlite.New() возвращает объект storage (хранилище) и ошибку err.
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		// Если err не nil (т.е. произошла ошибка), логируем её через log.Error().
		// sl.Err(err) – это вспомогательная функция для форматирования ошибки в логах.
		log.Error("failed to init storage", sl.Err(err))

		// os.Exit(1) – завершает выполнение программы с кодом ошибки 1 (общепринятое значение ошибки).
		os.Exit(1)
	}

	// _ = storage – временная заглушка, чтобы компилятор не ругался на неиспользуемую переменную.
	// В будущем здесь будет код работы с хранилищем.
	_ = storage

	// TODO: init router: chi

	// Создаём новый HTTP-роутер, вызывая chi.NewRouter().
	// chi.NewRouter() возвращает объект router, который будет обрабатывать входящие HTTP-запросы.
	router := chi.NewRouter()

	// Подключаем middleware (промежуточные обработчики, которые выполняются перед основным обработчиком запроса).

	// middleware.RequestID – это встроенный middleware из chi, который добавляет уникальный идентификатор (UUID) к каждому HTTP-запросу.
	router.Use(middleware.RequestID)

	// middleware.Logger – логирует входящие HTTP-запросы (метод, URL, время обработки и код ответа).
	router.Use(middleware.Logger)

	// mwLogger.New(log) – кастомный middleware, который использует наш логгер log для логирования запросов.
	router.Use(mwLogger.New(log))

	// middleware.Recoverer – встроенный middleware из chi, который обрабатывает паники внутри обработчиков.
	// Если в коде произойдёт panic, сервер не упадёт, а вернёт клиенту 500 Internal Server Error.
	router.Use(middleware.Recoverer)

	// middleware.URLFormat – встроенный middleware, который позволяет работать с URL-форматами.
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.Auth.User: cfg.Auth.Password, 
		}))

		r.Post("/", save.New(log, storage))
		r.Delete("/{alias}", delete.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	// TODO: run server
	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")

}

// setupLogger принимает строковый параметр env (среду выполнения)
// и возвращает указатель на объект slog.Logger.
func setupLogger(env string) *slog.Logger {
	// Объявляем переменную log, которая будет хранить указатель на объект slog.Logger.
	var log *slog.Logger

	// switch проверяет значение переменной env и выбирает соответствующий блок кода.
	switch env {
	case envLocal:
		// Если среда - локальная (local), используем кастомный логгер.
		log = setupPrettySlog() // setupPrettySlog() — самописная функция для красивого вывода логов в локальной среде.

	case envDev:
		// Если среда - dev (разработка), создаём JSON-логгер.
		log = slog.New( // slog.New создаёт новый объект логгера.
			slog.NewJSONHandler( // slog.NewJSONHandler создаёт обработчик логов, который выводит данные в JSON-формате.
				os.Stdout, // os.Stdout — стандартный вывод (терминал или лог-файл, если перенаправить вывод).
				&slog.HandlerOptions{Level: slog.LevelDebug}, // Указываем уровень логирования - Debug.
			),
		)

	case envProd:
		// Если среда - продакшен (prod), создаём JSON-логгер с уровнем Info (меньше подробностей).
		log = slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelInfo}, // В продакшене уровень логирования - Info (без debug).
			),
		)
	}

	// Возвращаем объект логгера.
	return log
}

// setupPrettySlog создаёт и настраивает логгер с красивым форматированием для локальной среды.
// Возвращает указатель на объект slog.Logger.
func setupPrettySlog() *slog.Logger {
	// Создаём объект настроек PrettyHandlerOptions из пакета slogpretty.
	// PrettyHandlerOptions отвечает за стилизацию логов (например, добавление цветов, форматирование строк).
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{ // Вложенные настройки для slog.
			Level: slog.LevelDebug, // Устанавливаем уровень логирования на Debug (показывает все сообщения).
		},
	}

	// Создаём обработчик логов с красивым форматированием.
	// opts.NewPrettyHandler(os.Stdout) создаёт новый обработчик, который пишет логи в стандартный вывод (терминал).
	handler := opts.NewPrettyHandler(os.Stdout)

	// Создаём новый логгер, передавая в него обработчик handler.
	// slog.New(handler) возвращает объект slog.Logger, который будет использовать этот обработчик.
	return slog.New(handler)
}

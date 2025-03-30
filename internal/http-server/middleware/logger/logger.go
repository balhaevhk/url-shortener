package logger

// Импортируем необходимые пакеты для работы с логированием, HTTP-сервером и промежуточным ПО (middleware).
import (
	// log/slog - стандартный пакет для логирования в Go, используется для создания и управления логами.
	"log/slog"

	// net/http - стандартный пакет для работы с HTTP-сервером и клиентом в Go.
	// Содержит все основные типы и методы для реализации HTTP-сервера.
	"net/http"

	// time - стандартный пакет для работы с временем, используется для работы с длительностью и временем в приложении.
	"time"

	// github.com/go-chi/chi/v5/middleware - сторонний пакет для промежуточного ПО в библиотеке chi.
	// Включает в себя набор полезных middleware для обработки запросов в веб-приложениях на базе chi.
	"github.com/go-chi/chi/v5/middleware"
)

// New - функция, которая возвращает middleware для логирования HTTP-запросов.
// Входной параметр log - это уже настроенный логгер (slog.Logger).
// Функция создает новый обработчик запросов, который будет логировать информацию о запросах и их ответах.
func New(log *slog.Logger) func(next http.Handler) http.Handler {
	// Возвращаем функцию, которая принимает следующий обработчик HTTP-запросов (next) и возвращает новый обработчик.
	return func(next http.Handler) http.Handler {
		// Создаем новый логгер, добавляя к нему метку, что это компонент "middleware/logger".
		// Это полезно для организации логов, чтобы знать, откуда пришло сообщение (из middleware).
		log := log.With(
			slog.String("component", "middleware/logger"),
		)

		// Логируем, что middleware для логирования включено.
		log.Info("logger middleware enabled")

		// Создаем функцию-обработчик для HTTP-запросов.
		// Эта функция будет логировать информацию о запросах и обрабатывать их.
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Создаем лог-обработчик для каждого запроса, добавляя в лог информацию о запросе:
			// метод запроса, путь, удаленный адрес, user-agent и ID запроса.
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)

			// Используем middleware.NewWrapResponseWriter для того, чтобы обернуть стандартный ResponseWriter.
			// Это позволяет отслеживать статус ответа и количество отправленных байтов.
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Фиксируем текущее время, чтобы затем вычислить продолжительность обработки запроса.
			t1 := time.Now()

			// Отложенно логируем завершение обработки запроса (это будет выполнено после того, как запрос будет обработан).
			defer func() {
				// Логируем информацию о завершении запроса: статус, количество байтов и продолжительность.
				entry.Info("request completed",
					slog.Int("status", ww.Status()),                  // Статус ответа.
					slog.Int("bytes", ww.BytesWritten()),             // Количество отправленных байтов.
					slog.String("duration", time.Since(t1).String()), // Время выполнения запроса.
				)
			}()

			// Передаем запрос следующему обработчику в цепочке.
			// Это вызовет обработку запроса в дальнейшем middleware или основном обработчике.
			next.ServeHTTP(ww, r)
		}

		// Возвращаем новый обработчик в виде http.HandlerFunc.
		// Это необходимо для того, чтобы соответствовать интерфейсу http.Handler.
		return http.HandlerFunc(fn)
	}
}

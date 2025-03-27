package config

// Подключаем стандартные библиотеки и сторонние пакеты
import (
	"log"  // Стандартная библиотека для логирования. Предназначена для вывода сообщений в консоль или в файл.
	"os"   // Стандартная библиотека для работы с операционной системой, например, для работы с файловой системой, переменными окружения и т.д.
	"time" // Стандартная библиотека для работы с временем: функции для работы с временем, длительностью и датой.

	// Сторонние библиотеки
	"github.com/ilyakaznacheev/cleanenv"  // cleanenv — библиотека для простого и удобного парсинга конфигурационных файлов и переменных окружения.
	"github.com/joho/godotenv"          // godotenv — библиотека для загрузки переменных окружения из .env файлов в приложение.
)


// Config - структура для хранения конфигурации приложения.
// В ней содержатся параметры, такие как среда выполнения, путь к базе данных и настройки HTTP-сервера.
type Config struct {
	// Env - указывает на среду выполнения приложения (например, "local", "dev", "prod").
	// Это поле будет считываться как из конфигурационного файла (YAML), так и из переменных окружения.
	// Если переменная окружения ENV не установлена, по умолчанию будет использоваться значение "local".
	// Также данное поле обязательно для заполнения (env-required:"true").
	Env string `yaml:"env" env:"ENV" env-default:"local" env-required:"true"`

	// StoragePath - путь к файлу базы данных, который будет использоваться для хранения данных.
	// Это поле обязано быть задано в переменных окружения, и его значение не может быть пустым.
	StoragePath string `yaml:"storage_path" env-required:"true"`

	// HTTPServer - структура, содержащая конфигурацию для HTTP-сервера.
	// В конфигурационном файле (YAML) и переменных окружения будет указано под полем "http_server".
	// Эта структура содержит настройки для работы с сервером (например, адрес, таймауты и т.д.).
	HTTPServer HTTPServer `yaml:"http_server"`
}


// HTTPServer - структура для хранения конфигурации HTTP-сервера.
// Включает параметры, такие как адрес, таймауты и другие настройки для работы с сервером.
type HTTPServer struct {
	// Adress - адрес, на котором будет слушать HTTP-сервер.
	// По умолчанию указывается "localhost:8080". Это значение будет использовано, если в конфигурации или переменных окружения не указано другое.
	Adress string `yaml:"address" env-default:"localhost:8080"`

	// Timeout - общий таймаут для запросов к серверу. Указывает максимальное время ожидания для ответа.
	// По умолчанию установлено значение 4 секунды.
	// Это значение будет использоваться, если в конфигурации или переменных окружения не указано другое.
	Timeout time.Duration `yaml:"timeout" env-default:"4"`

	// IdleTimeout - время бездействия соединения. Указывает максимальное время, в течение которого соединение может оставаться неактивным.
	// Если в конфигурации или переменных окружения не указано другое значение, используется значение 60 секунд.
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60"`
}


// MustLoad - функция для загрузки конфигурации приложения.
// 1. Загружает переменные окружения из файла .env.
// 2. Получает путь к конфигурационному файлу из переменной окружения CONFIG_PATH.
// 3. Проверяет существование конфигурационного файла.
// 4. Читает конфигурацию с помощью cleanenv и возвращает структуру с данными конфигурации.
func MustLoad() *Config {
	// Загружаем переменные окружения из файла .env.
	// godotenv.Load загружает все переменные из файла ".env" в окружение приложения.
	// В случае ошибки загрузки .env файла программа завершится с выводом сообщения об ошибке.
	err := godotenv.Load("../../.env")		
	if err != nil {
		log.Fatal("Error loading .env file") // Если возникла ошибка при загрузке, выводим сообщение и завершаем программу.
	}

	// Получаем путь к конфигурационному файлу из переменной окружения CONFIG_PATH.
	// Если переменная окружения не установлена, то это будет пустая строка.
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set") // Если переменная окружения пустая, выводим ошибку и завершаем выполнение.
	}

	// Проверяем, существует ли файл конфигурации по указанному пути.
	// Если файл не существует, выводим ошибку с указанием пути к отсутствующему файлу.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath) // Завершаем выполнение, если файл не найден.
	}

	// Создаем переменную для хранения конфигурации.
	var cfg Config

	// Читаем конфигурацию из файла с помощью библиотеки cleanenv.
	// cleanenv.ReadConfig читает данные из файла конфигурации и заполняет структуру cfg.
	// Если произошла ошибка при чтении, выводим сообщение с ошибкой и завершаем программу.
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err) // Завершаем выполнение с ошибкой, если не удается прочитать конфигурацию.
	}

	// Возвращаем указатель на загруженную структуру конфигурации.
	return &cfg
}


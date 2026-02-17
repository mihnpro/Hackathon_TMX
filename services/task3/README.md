# Task 3: Веб-приложение для анализа маршрутов локомотивов

## Описание

Комплексное веб-приложение для анализа, визуализации и прогнозирования параметров эксплуатации локомотивов. Интегрирует три основные задачи анализа данных с интерактивным пользовательским интерфейсом и взаимодействует с ML-моделью предсказания износа колес.

**Основные функции:**
- 🚂 Анализ маршрутов локомотивов и определение веток депо
- 📊 Определение самых популярных маршрутов для каждого локомотива
- 🗺️ Интерактивная визуализация на карте с тепловой картой посещений
- 🤖 Интеграция с ML-моделью для прогнозирования износа

---

## Технологический стек

### Backend
| Компонент | Технология | Версия |
|-----------|-----------|--------|
| **Language** | Go | 1.24.3 |
| **Web Framework** | Gin | 1.11.0 |
| **CORS** | Gin CORS | 1.7.6 |
| **ID Generation** | UUID | 1.6.0 |
| **Container** | Docker | - |

### Frontend
| Компонент | Технология |
|-----------|-----------|
| **Markup** | HTML5 |
| **Styling** | CSS3 |
| **Scripting** | JavaScript (Vanilla) |
| **Maps** | Leaflet.js |
| **Charts** | Chart.js |

---

## Архитектура

```
Go Backend (Gin Server)
    ↓
┌─────────────────────────────────────┐
│     Internal Services Layer         │
├─────────────────────────────────────┤
│  • AlgorithmService (Task 1)        │
│  • MostPopularTripService (Task 2)  │
│  • VisualizationService (Task 3)    │
│  • MLIntegrationService (Task 2 ML) │
└─────────────────────────────────────┘
    ↓
┌─────────────────────────────────────┐
│    Domain Models                    │
├─────────────────────────────────────┤
│  • Locomotive, Station, Route       │
│  • Trip, Direction, Predictions     │
└─────────────────────────────────────┘
    ↓
┌─────────────────────────────────────┐
│   Data Layer (CSV files)            │
├─────────────────────────────────────┤
│  • locomotives_displacement.csv     │
│  • station_info.csv                 │
└─────────────────────────────────────┘
```

---

## Структура проекта

```
task3/
├── cmd/
│   ├── main.go                    # CLI инструмент для обработки данных
│   └── web/
│       └── main.go                # Веб-сервер (главная точка входа)
├── internal/
│   ├── domain/                    # Модели данных
│   │   ├── locomotive.go
│   │   ├── station.go
│   │   ├── trip.go
│   │   ├── route.go
│   │   ├── direction.go
│   │   ├── improvedBranch.go
│   │   ├── routePoint.go
│   │   ├── jsRoutes.go
│   │   ├── jsStations.go
│   │   ├── mapConfig.go
│   │   ├── stationStats.go
│   │   ├── locomotiveDirectionStats.go
│   │   ├── locomotiveRoute.go
│   │   ├── record.go
│   │   └── ml/
│   │       ├── errors.go
│   │       └── predictions.go
│   ├── services/                  # Бизнес-логика
│   │   ├── 3.1.go                # Алгоритм определения веток
│   │   ├── 3.2.go                # Определение популярных маршрутов
│   │   ├── 3.3.go                # Генерация карт
│   │   ├── helpers_funcs.go
│   │   └── ml_integration.go      # Интеграция с ML сервисом
│   └── transport/                 # HTTP handlers и маршруты
│       ├── handlers/
│       │   ├── handler.go         # Базовый handler
│       │   ├── task1.go           # Обработчик Task 1
│       │   ├── task2.go           # Обработчик Task 2
│       │   ├── task3.go           # Обработчик Task 3
│       │   └── ml.go              # Обработчик ML запросов
│       ├── models/
│       │   ├── requests/
│       │   │   └── task2.go       # Типы запросов
│       │   └── responses/
│       │       ├── task1.go       # Типы ответов Task 1
│       │       ├── task2.go       # Типы ответов Task 2
│       │       ├── task3.go       # Типы ответов Task 3
│       │       └── ml.go          # Типы ответов ML
│       └── routes/
│           └── routes.go          # Конфигурация маршрутов
├── frontend/                      # Статические файлы (HTML/CSS/JS)
│   ├── index.html                # Главная страница
│   ├── shared/                   # Общие ресурсы
│   │   ├── css/
│   │   │   └── common.css
│   │   └── js/
│   │       └── utils.js
│   ├── task1/                    # Интерфейс Task 1
│   │   ├── index.html
│   │   ├── css/style.css
│   │   └── js/app.js
│   ├── task2/                    # Интерфейс Task 2
│   │   ├── index.html
│   │   ├── css/style.css
│   │   └── js/app.js
│   ├── task3/                    # Интерфейс Task 3
│   │   ├── index.html
│   │   ├── css/style.css
│   │   └── js/app.js
│   └── ml/                       # Интерфейс ML (прогнозы)
│       ├── index.html
│       ├── css/style.css
│       └── js/app.js
├── data/
│   ├── locomotives_displacement.csv    # Данные о маршрутах
│   ├── station_info.csv                # Информация о станциях
│   └── stations_map_with_heat.html     # Сгенерированные карты
├── maps/                         # Динамически сгенерированные карты
├── uploads/                      # Загруженные файлы пользователями
├── go.mod                        # Go модули
├── Dockerfile                    # Docker конфигурация
└── docker-compose.yaml           # (в services/) Compose конфигурация
```

---

## REST API Эндпоинты

### Health Check
```
GET /health
```
Проверка статуса сервера.

**Ответ:**
```json
{"status": "ok"}
```

---

### Task 1: Анализ веток депо

#### Получить все депо
```
GET /api/v1/task1/depots
```
Список всех депо в системе.

**Ответ:**
```json
{
  "depots": ["940006", "940008", "940009"]
}
```

---

#### Получить ветки конкретного депо
```
GET /api/v1/task1/depots/:depo/branches
```

**Параметры:**
- `:depo` - ID депо

**Ответ:**
```json
{
  "depo": "940006",
  "branches": [
    {
      "id": "branch_1",
      "endpoints": ["station_1", "station_5"],
      "from_depot": "station_1",
      "intermediate_stations": ["station_2", "station_3", "station_4"]
    }
  ]
}
```

---

#### Получить анализ всех веток
```
GET /api/v1/task1/branches
```

**Ответ:**
```json
{
  "total_branches": 15,
  "branches_by_depot": {
    "940006": 5,
    "940008": 4,
    "940009": 6
  }
}
```

---

### Task 2: Популярные маршруты

#### Получить популярные маршруты всех локомотивов
```
GET /api/v1/popular-direction
```

**Ответ:**
```json
{
  "locomotives": [
    {
      "series": "2TE25A",
      "number": 1,
      "depo": "940006",
      "popular_branch": "branch_1",
      "visits_count": 45,
      "avg_distance_km": 120.5
    }
  ]
}
```

---

#### Получить популярный маршрут конкретного локомотива
```
GET /api/v1/locomotives/:series/:number/popular-direction
```

**Параметры:**
- `:series` - серия локомотива (например, `2TE25A`)
- `:number` - номер локомотива (например, `1`)

**Ответ:**
```json
{
  "locomotive": {
    "series": "2TE25A",
    "number": 1
  },
  "popular_direction": {
    "branch_id": "branch_1",
    "depot": "940006",
    "visits": 45,
    "endpoints": ["station_1", "station_5"],
    "intermediate_stations": 4
  }
}
```

---

### Task 3: Визуализация и создание карт

#### Получить доступные депо
```
GET /api/v1/task3/depots
```

**Ответ:**
```json
{
  "available_depots": ["940006", "940008", "940009"],
  "depot_count": 3
}
```

---

#### Получить информацию о депо
```
GET /api/v1/task3/depots/:depo
```

**Параметры:**
- `:depo` - ID депо

**Ответ:**
```json
{
  "depo_id": "940006",
  "locomotives_count": 25,
  "total_routes": 150,
  "stations": 45,
  "branches": 5
}
```

---

#### Сгенерировать карты
```
POST /api/v1/task3/generate
```

**Тело запроса:**
```json
{
  "depot": "940006",
  "max_locomotives": 10
}
```

**Ответ:**
```json
{
  "status": "success",
  "maps_generated": 10,
  "map_files": [
    "map_940006_2TE25A_1.html",
    "map_940006_2TE25A_2.html"
  ],
  "map_url": "/maps/map_940006_2TE25A_1.html"
}
```

---

### ML Integration: Прогноз износа

#### Сделать предсказание
```
POST /api/v1/ml/predict
```

**Тело запроса:**
```json
{
  "locomotive_series": "2TE25A",
  "locomotive_number": 1,
  "depo": "940006",
  "steel_num": "45",
  "mileage_start": 150000.5
}
```

**Ответ:**
```json
{
  "prediction": 0.85,
  "locomotive": {
    "series": "2TE25A",
    "number": 1
  },
  "confidence": "high"
}
```

---

#### Загрузить файл данных
```
POST /api/v1/ml/upload
Content-Type: multipart/form-data
```

**Параметры:**
- `file` - CSV файл с данными локомотивов

**Ответ:**
```json
{
  "status": "success",
  "file_size": 1024000,
  "filename": "upload_123456.csv",
  "message": "Файл успешно загружен"
}
```

---

#### Проверить доступность ML сервиса
```
GET /api/v1/ml/health
```

**Ответ:**
```json
{
  "status": "healthy",
  "ml_service": "available",
  "version": "1.0.0"
}
```

---

#### Получить информацию о модели
```
GET /api/v1/ml/info
```

**Ответ:**
```json
{
  "model_name": "CatBoost Wear Predictor",
  "features": 15,
  "version": "1.0.0",
  "last_update": "2024-02-15"
}
```

---

## Frontend Страницы

| URL | Описание |
|-----|---------|
| `/` | Главная страница с навигацией |
| `/task1` | Анализ веток депо и маршрутов |
| `/task2` | Анализ популярных маршрутов |
| `/task3` | Интерактивная карта с визуализацией |
| `/ml` | Интерфейс для предсказания износа |

---

## Запуск приложения

### Только Backend (Go)
```bash
cd cmd/web
go run main.go
```

Сервер запустится на `http://localhost:8080`

### С Docker Compose (полный стек)
```bash
cd ..
docker-compose up
```

**Доступные сервисы:**
- Web App: `http://localhost:8080`
- Task 2 ML API: `http://localhost:8000`

### С помощью CLI инструмента
```bash
# Выполнить все задачи
go run cmd/main.go -task=all

# Задача 1: Анализ веток
go run cmd/main.go -task=1

# Задача 2: Популярные маршруты
go run cmd/main.go -task=2

# Задача 3: Визуализация для конкретного депо
go run cmd/main.go -task=3 -depo=940006 -max=10
```

---

## Переменные окружения

| Переменная | Описание | По умолчанию |
|-----------|---------|------------|
| `WEAR_PREDICTION_URL` | URL ML сервиса (Task 2) | `http://localhost:8000` |
| `PORT` | Порт веб-сервера | `8080` |
| `DATA_PATH` | Путь к файлу данных локомотивов | `./data/locomotives_displacement.csv` |
| `STATION_INFO_PATH` | Путь к информации о станциях | `./data/station_info.csv` |

---

## Интеграция с Task 2 (ML)

Приложение автоматически интегрируется с ML сервисом Task 2:

1. **При старте:** проверяет доступность ML сервиса
2. **При предсказании:** отправляет запросы к `/predict` эндпоинту Task 2
3. **Обработка ошибок:** корректно обрабатывает недоступность ML сервиса
4. **Кэширование:** кэширует результаты предсказаний

---

## Производительность

- **Время отклика API:** 50-200ms
- **Время генерации карты:** 500-2000ms
- **Поддерживаемые одновременные пользователи:** 100+
- **Максимальное количество локомотивов на карте:** зависит от размера данных

---

## Особенности

✅ Полнофункциональный анализ маршрутов  
✅ Интерактивные визуализации на карте  
✅ Интеграция с ML-моделью  
✅ Поддержка загрузки данных  
✅ CORS для кросс-доменных запросов  
✅ Docker контейнеризация  
✅ Отзывчивый веб-интерфейс  
✅ REST API для программной интеграции  


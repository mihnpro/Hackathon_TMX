# Task 2: Модель прогнозирования ресурса колеса локомотива

## Описание

Система предсказывает **интенсивность изнашивания колес локомотивов** на основе данных об эксплуатации. Приложение предоставляет REST API для получения предсказаний и использует машинное обучение для анализа истории использования.

**Критерий оценки:** Средняя квадратичная ошибка (MSE) предсказания интенсивности износа.

---

## Технологический стек

| Компонент | Технология | Версия |
|-----------|-----------|--------|
| **Framework** | FastAPI | 0.104.1 |
| **Web Server** | Uvicorn | 0.24.0 |
| **Machine Learning** | CatBoost | 1.2.5 |
| **Data Processing** | Pandas | 2.0.3 |
| **Numerical Computing** | NumPy | 1.24.3 |
| **Validation** | Pydantic | 2.5.0 |
| **Model Serialization** | Joblib | 1.3.2 |
| **Containerization** | Docker | - |
| **ML Pipeline** | Scikit-learn | 1.3.0 |

---

## Структура проекта

```
wear_prediction/
├── app.py                      # Основное приложение FastAPI
├── requirements.txt            # Зависимости Python
├── Dockerfile                  # Docker конфигурация
├── models/
│   ├── catboost_model_latest.pkl       # Обученная модель CatBoost
│   └── feature_columns_latest.pkl      # Список признаков для модели
├── data/
│   └── raw/
│       └── service_dates.csv          # История ремонтов локомотивов
└── test_data/
    └── test1.json              # Примеры тестовых данных
```

---

## Как работает система

### 1. **Загрузка модели**
- При старте приложения загружается обученная модель CatBoost
- Загружается набор признаков, необходимых для предсказания
- Загружаются исторические данные о ремонтах

### 2. **Подготовка данных**
Система принимает данные о локомотиве и:
- Рассчитывает статистику ремонтов (всего ремонтов, типы ремонтов, обточки)
- Нормализует данные о типе стали
- Вычисляет дополнительные признаки:
  - `repairs_per_100k` - количество ремонтов на 100,000 км пути
  - `turning_per_100k` - количество обточек на 100,000 км пути
  - `turning_ratio` - доля обточек среди всех ремонтов

### 3. **Предсказание**
- Модель анализирует признаки локомотива
- Возвращает предсказанное значение интенсивности износа

---

## REST API Эндпоинты

### `/predict` - POST
**Описание:** Получить предсказание интенсивности износа колеса

**Параметры тела запроса:**
```json
{
  "locomotive_series": "2TE25A",
  "locomotive_number": 1,
  "depo": "940006",
  "steel_num": "45",
  "mileage_start": 150000.5
}
```

**Параметры:**
- `locomotive_series` (string, обязательный) - модель/серия локомотива
- `locomotive_number` (integer, обязательный) - номер локомотива
- `depo` (string, обязательный) - депо приписки
- `steel_num` (string, обязательный) - марка стали колеса
- `mileage_start` (float, обязательный) - пройденное расстояние в км

**Ответ (успешный):**
```json
{
  "prediction": 0.85,
  "locomotive": {
    "series": "2TE25A",
    "number": 1,
    "depo": "940006"
  },
  "confidence": "high",
  "computation_time_ms": 45
}
```

**HTTP коды:**
- `200 OK` - успешное предсказание
- `422 Unprocessable Entity` - некорректные данные
- `500 Internal Server Error` - ошибка сервера

---

## Примеры использования

### cURL
```bash
curl -X POST "http://localhost:8000/predict" \
  -H "Content-Type: application/json" \
  -d '{
    "locomotive_series": "2TE25A",
    "locomotive_number": 1,
    "depo": "940006",
    "steel_num": "45",
    "mileage_start": 150000.5
  }'
```

### Python (requests)
```python
import requests

data = {
    "locomotive_series": "2TE25A",
    "locomotive_number": 1,
    "depo": "940006",
    "steel_num": "45",
    "mileage_start": 150000.5
}

response = requests.post("http://localhost:8000/predict", json=data)
prediction = response.json()["prediction"]
print(f"Интенсивность износа: {prediction}")
```

### JavaScript (fetch)
```javascript
const data = {
    locomotive_series: "2TE25A",
    locomotive_number: 1,
    depo: "940006",
    steel_num: "45",
    mileage_start: 150000.5
};

fetch('http://localhost:8000/predict', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
})
.then(res => res.json())
.then(data => console.log('Прогноз:', data.prediction));
```

---

## Запуск приложения

### Локально (Python)
```bash
cd wear_prediction
pip install -r requirements.txt
python -m uvicorn app:app --host 0.0.0.0 --port 8000
```

### Docker
```bash
docker build -t wear-prediction .
docker run -p 8000:8000 wear-prediction
```

### Docker Compose
```bash
cd ..
docker-compose up wear_prediction
```

Приложение будет доступно по адресу: `http://localhost:8000`

---

## Документация API

Интерактивная документация Swagger доступна по адресу:
```
http://localhost:8000/docs
```

Альтернативная документация ReDoc:
```
http://localhost:8000/redoc
```

---

## Переменные окружения

| Переменная | Описание | По умолчанию |
|-----------|---------|------------|
| `MODEL_PATH` | Путь к файлу модели | `models/catboost_model_latest.pkl` |
| `SERVICE_DATES_PATH` | Путь к файлу ремонтов | `data/raw/service_dates.csv` |

---

## Логирование

Приложение использует Python `logging` с уровнем INFO. Логи выводятся в консоль и включают:
- Загрузку модели и данных
- Время обработки запросов
- Ошибки валидации

---

## Производительность

- **Время ответа:** ~50-100ms (зависит от нагрузки)
- **Поддерживаемых одновременных подключений:** 100+
- **Использование памяти:** ~200-300MB

---

## Особенности

✅ Быстрые предсказания в реальном времени  
✅ Параллельная обработка запросов  
✅ Валидация входных данных  
✅ Подробное логирование  
✅ Docker контейнеризация для легкого развертывания  
✅ Интеграция с историческими данными о ремонтах  

---

## Интеграция с Task 3

Task 3 (веб-приложение) взаимодействует с этим сервисом через REST API для:
- Получения предсказаний износа
- Визуализации результатов на интерфейсе
- Анализа трендов износа по разным локомотивам

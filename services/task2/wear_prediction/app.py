from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import pandas as pd
import numpy as np
import joblib
import uvicorn
from pathlib import Path
from typing import List
import logging
import time

# ========== НАСТРОЙКА ЛОГИРОВАНИЯ ==========
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# ========== ЗАГРУЗКА МОДЕЛИ ==========
MODEL_PATH = Path(__file__).parent / 'models' / 'catboost_model_latest.pkl'
FEATURES_PATH = Path(__file__).parent / 'models' / 'feature_columns_latest.pkl'
SERVICE_DATES_PATH = Path(__file__).parent / 'data' / 'raw' / 'service_dates.csv'

# Проверка наличия файлов
if not MODEL_PATH.exists():
    raise FileNotFoundError(f"Модель не найдена: {MODEL_PATH}")
if not FEATURES_PATH.exists():
    raise FileNotFoundError(f"Файл признаков не найден: {FEATURES_PATH}")
if not SERVICE_DATES_PATH.exists():
    raise FileNotFoundError(f"Файл с ремонтами не найден: {SERVICE_DATES_PATH}")

# Загрузка модели и данных
logger.info("Загрузка модели...")
model = joblib.load(MODEL_PATH)
expected_features = joblib.load(FEATURES_PATH)
logger.info(f"Модель загружена. Признаков: {len(expected_features)}")

logger.info("Загрузка данных о ремонтах...")
service_dates = pd.read_csv(SERVICE_DATES_PATH)
logger.info(f"Загружено {len(service_dates)} записей о ремонтах")

# ========== СОЗДАНИЕ APP ==========
app = FastAPI(title="Wear Intensity Predictor", version="1.0.0")

# ========== PYDANTIC МОДЕЛИ (ТОЛЬКО ПОЛЯ ИЗ ТЗ) ==========
class WheelInput(BaseModel):
    locomotive_series: str
    locomotive_number: int
    depo: str
    steel_num: str
    mileage_start: float

# ========== ФУНКЦИИ ДЛЯ РАБОТЫ С РЕМОНТАМИ ==========
def get_repair_stats(series: str, number: int) -> dict:
    """
    Получить статистику ремонтов для локомотива из service_dates.csv
    """
    # Фильтруем записи для данного локомотива
    mask = (service_dates['locomotive_series'] == series) & \
           (service_dates['locomotive_number'] == number)
    repairs = service_dates[mask]
    
    # Считаем статистику
    if len(repairs) == 0:
        # Если ремонтов нет, возвращаем нули
        return {
            'total_repairs': 0,
            'repair_type_1': 0,
            'repair_type_2': 0,
            'repair_type_3': 0,
            'turning_count': 0,
            'unique_service_dates': 0
        }
    
    # Считаем количество каждого типа ремонта
    return {
        'total_repairs': len(repairs),
        'repair_type_1': (repairs['service_type'] == '1').sum(),
        'repair_type_2': (repairs['service_type'] == '2').sum(),
        'repair_type_3': (repairs['service_type'] == '3').sum(),
        'turning_count': (repairs['service_type'] == 'обточка').sum(),
        'unique_service_dates': repairs['service_date'].nunique()
    }

# ========== ФУНКЦИИ ПОДГОТОВКИ ДАННЫХ ==========
def prepare_data(df: pd.DataFrame) -> pd.DataFrame:
    """
    Подготовка данных для модели:
    - расчёт дополнительных признаков
    - обработка steel_num
    - приведение к нужному порядку колонок
    """
    df = df.copy()
    
    # Рассчитываем дополнительные признаки
    df['repairs_per_100k'] = df['total_repairs'] / (df['mileage_start'] / 100000 + 1)
    df['turning_per_100k'] = df['turning_count'] / (df['mileage_start'] / 100000 + 1)
    df['turning_ratio'] = df['turning_count'] / (df['total_repairs'] + 1)
    
    # Обработка steel_num
    df['steel_num'] = df['steel_num'].astype(str).str.replace('.0', '', regex=False).str.strip()
    df['steel_num'] = df['steel_num'].replace('', 'unknown').replace('nan', 'unknown')
    
    # Удаляем locomotive_number (не нужен для модели)
    if 'locomotive_number' in df.columns:
        df = df.drop(columns=['locomotive_number'])
    
    # Заполняем пропуски и бесконечности
    df = df.fillna(0)
    df = df.replace([np.inf, -np.inf], 0)
    
    # Добавляем недостающие признаки (если вдруг каких-то нет)
    for col in expected_features:
        if col not in df.columns:
            df[col] = 0
    
    # Возвращаем только нужные колонки в правильном порядке
    return df[expected_features]

# ========== ЭНДПОИНТЫ ==========
@app.get("/")
def root():
    return {
        "message": "Wear Intensity Predictor API",
        "status": "ok",
        "version": "1.0.0"
    }

@app.get("/health")
def health():
    return {
        "status": "healthy",
        "model_loaded": True,
        "service_dates_loaded": len(service_dates) > 0
    }

@app.get("/info")
def info():
    """Информация о модели для клиентов"""
    return {
        "model_type": "CatBoost",
        "features_count": len(expected_features),
        "features": expected_features[:10],
        "version": "1.0.0",
        "input_format": {
            "type": "array",
            "items": {
                "locomotive_series": "string",
                "locomotive_number": "integer",
                "depo": "string",
                "steel_num": "string",
                "mileage_start": "number"
            }
        },
        "output_format": "array of numbers"
    }

@app.post("/predict", response_model=List[float])
def predict(items: List[WheelInput]):
    """
    Предсказать интенсивность износа для списка колёс
    
    Ожидает массив объектов с полями:
    - locomotive_series: str
    - locomotive_number: int
    - depo: str
    - steel_num: str
    - mileage_start: float
    
    Возвращает массив предсказаний
    """
    start_time = time.time()
    logger.info(f"Получен запрос с {len(items)} объектами")
    
    # Ограничение на размер запроса
    if len(items) > 1000:
        raise HTTPException(
            status_code=400,
            detail="Слишком много объектов. Максимум 1000 за запрос"
        )
    
    try:
        # Обогащаем данные статистикой о ремонтах
        enriched_data = []
        for item in items:
            data = item.dict()
            # Добавляем данные о ремонтах
            repair_stats = get_repair_stats(
                data['locomotive_series'],
                data['locomotive_number']
            )
            data.update(repair_stats)
            enriched_data.append(data)
        
        # Преобразуем в DataFrame
        df = pd.DataFrame(enriched_data)
        logger.info(f"DataFrame создан, колонки: {df.columns.tolist()}")
        
        # Подготавливаем данные для модели
        df_prepared = prepare_data(df)
        logger.info(f"Данные подготовлены, форма: {df_prepared.shape}")
        
        # Предсказание
        predictions = model.predict(df_prepared).tolist()
        
        # Время обработки
        proc_time = (time.time() - start_time) * 1000
        logger.info(f"Запрос обработан за {proc_time:.2f} мс")
        
        return predictions
        
    except Exception as e:
        logger.error(f"Ошибка при обработке запроса: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

# ========== ЗАПУСК ==========
if __name__ == "__main__":
    uvicorn.run("app:app", host="0.0.0.0", port=8000, reload=True)
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import pandas as pd
import numpy as np
import joblib
import uvicorn
from pathlib import Path
from typing import List, Optional

# ========== ЗАГРУЗКА МОДЕЛИ ==========
MODEL_PATH = Path(__file__).parent / 'models' / 'catboost_model_latest.pkl'
FEATURES_PATH = Path(__file__).parent / 'models' / 'feature_columns_latest.pkl'

if not MODEL_PATH.exists() or not FEATURES_PATH.exists():
    raise FileNotFoundError("Модель или файл признаков не найдены")

model = joblib.load(MODEL_PATH)
expected_features = joblib.load(FEATURES_PATH)
print(f"Модель загружена. Признаков: {len(expected_features)}")

# ========== СОЗДАНИЕ APP ==========
app = FastAPI(title="Wear Intensity Predictor", version="1.0.0")

# ========== PYDANTIC МОДЕЛИ ==========
class WheelInput(BaseModel):
    locomotive_series: str
    locomotive_number: int
    depo: str
    steel_num: str
    mileage_start: float
    total_repairs: Optional[int] = 0
    repair_type_1: Optional[int] = 0
    repair_type_2: Optional[int] = 0
    repair_type_3: Optional[int] = 0
    turning_count: Optional[int] = 0
    unique_service_dates: Optional[int] = 0

class PredictionRequest(BaseModel):
    data: List[WheelInput]

class PredictionResponse(BaseModel):
    predictions: List[float]

# ========== ФУНКЦИИ ПОДГОТОВКИ ==========
def prepare_data(df: pd.DataFrame) -> pd.DataFrame:
    df = df.copy()
    
    # Рассчитываем дополнительные признаки
    df['repairs_per_100k'] = df['total_repairs'] / (df['mileage_start'] / 100000 + 1)
    df['turning_per_100k'] = df['turning_count'] / (df['mileage_start'] / 100000 + 1)
    df['turning_ratio'] = df['turning_count'] / (df['total_repairs'] + 1)
    
    # Обработка steel_num
    df['steel_num'] = df['steel_num'].astype(str).str.replace('.0', '', regex=False).str.strip()
    
    # Удаляем locomotive_number (не нужен для модели)
    df = df.drop(columns=['locomotive_number'])
    
    # Заполняем пропуски
    df = df.fillna(0)
    df = df.replace([np.inf, -np.inf], 0)
    
    # Добавляем недостающие признаки
    for col in expected_features:
        if col not in df.columns:
            df[col] = 0
    
    return df[expected_features]

# ========== ЭНДПОИНТЫ ==========
@app.get("/")
def root():
    return {"message": "API работает", "status": "ok"}

@app.get("/health")
def health():
    return {"status": "healthy"}

@app.post("/predict", response_model=PredictionResponse)
def predict(request: PredictionRequest):
    try:
        # Преобразуем в DataFrame
        input_data = [item.dict() for item in request.data]
        df = pd.DataFrame(input_data)
        
        # Подготавливаем данные
        df_prepared = prepare_data(df)
        
        # Предсказание
        predictions = model.predict(df_prepared).tolist()
        
        return PredictionResponse(predictions=predictions)
    
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

# ========== ЗАПУСК ==========
if __name__ == "__main__":
    uvicorn.run("app:app", host="0.0.0.0", port=8000)
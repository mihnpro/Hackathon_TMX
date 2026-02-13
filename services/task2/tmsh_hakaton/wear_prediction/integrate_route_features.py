"""
integrate_route_features.py
Интеграция данных из 3 задачи (маршруты) в модель
"""

import pandas as pd
import numpy as np
from datetime import datetime
from math import radians, cos, sin, asin, sqrt
import os

print("="*60)
print("ИНТЕГРАЦИЯ МАРШРУТНЫХ ДАННЫХ")
print("="*60)

# ============================================
# 1. ЗАГРУЗКА ДАННЫХ
# ============================================
print("\n1. Загрузка данных...")

# Данные модели
X_train = pd.read_csv('data/splits/X_train.csv')
X_test = pd.read_csv('data/splits/X_test.csv')

# Данные о перемещениях
displacement = pd.read_csv('data/locomotives_displacement.csv')
station_info = pd.read_csv('data/station_info.csv')

print(f"   X_train: {X_train.shape}")
print(f"   X_test: {X_test.shape}")
print(f"   displacement: {displacement.shape}")
print(f"   station_info: {station_info.shape}")

# ============================================
# 2. ПРЕДОБРАБОТКА ДАННЫХ О ПЕРЕМЕЩЕНИЯХ
# ============================================
print("\n2. Предобработка данных о перемещениях...")

# Преобразуем datetime
displacement['datetime'] = pd.to_datetime(displacement['datetime'])

# Сортируем по времени для каждого локомотива
displacement = displacement.sort_values(['locomotive_series', 'locomotive_number', 'datetime'])

print(f"   Диапазон дат: {displacement['datetime'].min()} - {displacement['datetime'].max()}")
print(f"   Уникальных локомотивов: {displacement[['locomotive_series', 'locomotive_number']].drop_duplicates().shape[0]}")

# ============================================
# 3. ФУНКЦИЯ ДЛЯ РАСЧЕТА РАССТОЯНИЯ МЕЖДУ СТАНЦИЯМИ
# ============================================
def haversine_distance(lat1, lon1, lat2, lon2):
    """Расчет расстояния между двумя точками на сфере (в км)"""
    R = 6371  # Радиус Земли в км
    
    lat1, lon1, lat2, lon2 = map(radians, [lat1, lon1, lat2, lon2])
    
    dlat = lat2 - lat1
    dlon = lon2 - lon1
    
    a = sin(dlat/2)**2 + cos(lat1) * cos(lat2) * sin(dlon/2)**2
    c = 2 * asin(sqrt(a))
    
    return R * c

# Добавляем координаты к перемещениям
print("\n3. Добавление координат к перемещениям...")

# Создаем словарь координат станций
station_coords = station_info.set_index('station')[['latitude', 'longitude']].to_dict('index')

# Добавляем координаты для каждой станции
displacement['lat'] = displacement['station'].map(lambda x: station_coords.get(x, {}).get('latitude', np.nan))
displacement['lon'] = displacement['station'].map(lambda x: station_coords.get(x, {}).get('longitude', np.nan))

# Удаляем записи без координат
before = len(displacement)
displacement = displacement.dropna(subset=['lat', 'lon'])
print(f"   Удалено записей без координат: {before - len(displacement)}")

# ============================================
# 4. ФУНКЦИЯ ДЛЯ АНАЛИЗА МАРШРУТОВ ОДНОГО ЛОКОМОТИВА
# ============================================
def analyze_locomotive_routes(loco_data):
    """Анализ маршрутов для одного локомотива"""
    
    results = {}
    
    # Базовые статистики
    results['total_visits'] = len(loco_data)
    results['unique_stations'] = loco_data['station'].nunique()
    
    # Временной период
    time_span = (loco_data['datetime'].max() - loco_data['datetime'].min()).days
    results['days_active'] = max(1, time_span)  # Избегаем деления на 0
    
    # Интенсивность
    results['visits_per_day'] = results['total_visits'] / results['days_active']
    
    # Определяем поездки (выезд из депо и возвращение)
    # Упрощенно: считаем каждую смену станции новой поездкой
    loco_data = loco_data.sort_values('datetime')
    loco_data['prev_station'] = loco_data['station'].shift(1)
    loco_data['station_changed'] = loco_data['station'] != loco_data['prev_station']
    
    results['num_trips'] = loco_data['station_changed'].sum()
    results['trips_per_day'] = results['num_trips'] / results['days_active']
    
    # Расчет расстояний (если есть координаты)
    distances = []
    for i in range(1, len(loco_data)):
        if pd.notna(loco_data.iloc[i-1]['lat']) and pd.notna(loco_data.iloc[i]['lat']):
            dist = haversine_distance(
                loco_data.iloc[i-1]['lat'], loco_data.iloc[i-1]['lon'],
                loco_data.iloc[i]['lat'], loco_data.iloc[i]['lon']
            )
            distances.append(dist)
    
    if distances:
        results['avg_trip_distance'] = np.mean(distances)
        results['max_trip_distance'] = np.max(distances)
        results['total_distance'] = np.sum(distances)
    else:
        results['avg_trip_distance'] = 0
        results['max_trip_distance'] = 0
        results['total_distance'] = 0
    
    # Географические характеристики
    if len(loco_data) > 0 and 'lat' in loco_data.columns:
        results['avg_latitude'] = loco_data['lat'].mean()
        results['avg_longitude'] = loco_data['lon'].mean()
        results['lat_span'] = loco_data['lat'].max() - loco_data['lat'].min()
        results['lon_span'] = loco_data['lon'].max() - loco_data['lon'].min()
    else:
        results['avg_latitude'] = 0
        results['avg_longitude'] = 0
        results['lat_span'] = 0
        results['lon_span'] = 0
    
    return results

# ============================================
# 5. РАСЧЕТ ПРИЗНАКОВ ДЛЯ КАЖДОГО ЛОКОМОТИВА
# ============================================
print("\n4. Расчет признаков для каждого локомотива...")

# Группируем по локомотивам
loco_groups = displacement.groupby(['locomotive_series', 'locomotive_number'])

route_features = []

for (series, number), group in loco_groups:
    if len(group) < 2:  # Пропускаем локомотивы с одной записью
        continue
    
    features = analyze_locomotive_routes(group)
    features['locomotive_series'] = series
    features['locomotive_number'] = number
    route_features.append(features)

route_df = pd.DataFrame(route_features)
print(f"\n   Рассчитано признаков для {len(route_df)} локомотивов")
print(f"   Признаки: {route_df.columns.tolist()}")

# ============================================
# 6. СОЗДАНИЕ ДОПОЛНИТЕЛЬНЫХ ГЕОГРАФИЧЕСКИХ ПРИЗНАКОВ
# ============================================
print("\n5. Создание географических признаков...")

# Климатические зоны по широте
def get_climate_zone(lat):
    if lat > 60:
        return 'arctic'
    elif lat > 50:
        return 'north'
    elif lat > 40:
        return 'temperate'
    else:
        return 'south'

route_df['climate_zone'] = route_df['avg_latitude'].apply(get_climate_zone)

# Горная местность (по разбросу высот/координат)
route_df['is_mountainous'] = (route_df['lat_span'] > 2).astype(int)  # Если размах широты большой

# Интенсивность использования
route_df['usage_intensity'] = pd.qcut(route_df['visits_per_day'], 
                                       q=5, 
                                       labels=['very_low', 'low', 'medium', 'high', 'very_high'])

print(f"   Добавлены признаки: climate_zone, is_mountainous, usage_intensity")

# ============================================
# 7. ОБЪЕДИНЕНИЕ С ТЕКУЩИМИ ДАННЫМИ
# ============================================
print("\n6. Объединение с данными модели...")

# Для X_train
print("   Обработка X_train...")
X_train_with_routes = X_train.merge(
    route_df.drop(columns=['locomotive_number']),  # Убираем номер, оставляем только серию для объединения
    on=['locomotive_series'],
    how='left'
)

# Для X_test
print("   Обработка X_test...")
X_test_with_routes = X_test.merge(
    route_df.drop(columns=['locomotive_number']),
    on=['locomotive_series'],
    how='left'
)

print(f"\n   X_train после объединения: {X_train_with_routes.shape}")
print(f"   X_test после объединения: {X_test_with_routes.shape}")

# ============================================
# 8. ЗАПОЛНЕНИЕ ПРОПУСКОВ
# ============================================
print("\n7. Заполнение пропусков...")

# Колонки, которые могли не объединиться
route_columns = ['total_visits', 'unique_stations', 'days_active', 'visits_per_day',
                 'num_trips', 'trips_per_day', 'avg_trip_distance', 'max_trip_distance',
                 'total_distance', 'avg_latitude', 'avg_longitude', 'lat_span', 'lon_span',
                 'climate_zone', 'is_mountainous', 'usage_intensity']

for col in route_columns:
    if col in X_train_with_routes.columns:
        if X_train_with_routes[col].dtype in ['float64', 'int64']:
            X_train_with_routes[col] = X_train_with_routes[col].fillna(0)
            X_test_with_routes[col] = X_test_with_routes[col].fillna(0)
        else:
            X_train_with_routes[col] = X_train_with_routes[col].fillna('unknown')
            X_test_with_routes[col] = X_test_with_routes[col].fillna('unknown')

print(f"   Пропуски после заполнения: {X_train_with_routes.isnull().sum().sum()}")

# ============================================
# 9. СОХРАНЕНИЕ НОВЫХ ДАННЫХ
# ============================================
print("\n8. Сохранение новых данных...")

# Создаем папку для расширенных данных
os.makedirs('data/enriched', exist_ok=True)

# Сохраняем
X_train_with_routes.to_csv('data/enriched/X_train_enriched.csv', index=False)
X_test_with_routes.to_csv('data/enriched/X_test_enriched.csv', index=False)

# Также сохраняем отдельно признаки для анализа
route_df.to_csv('data/enriched/locomotive_route_features.csv', index=False)

print(f"\n✅ Сохранено:")
print(f"   - data/enriched/X_train_enriched.csv")
print(f"   - data/enriched/X_test_enriched.csv")
print(f"   - data/enriched/locomotive_route_features.csv")

# ============================================
# 10. СТАТИСТИКА НОВЫХ ПРИЗНАКОВ
# ============================================
print("\n9. Статистика новых признаков:")
print("-"*40)

for col in ['total_visits', 'visits_per_day', 'avg_trip_distance', 'total_distance']:
    if col in X_train_with_routes.columns:
        print(f"\n{col}:")
        print(f"  Среднее: {X_train_with_routes[col].mean():.2f}")
        print(f"  Медиана: {X_train_with_routes[col].median():.2f}")
        print(f"  Мин: {X_train_with_routes[col].min():.2f}")
        print(f"  Макс: {X_train_with_routes[col].max():.2f}")

# Категориальные признаки
print("\nclimate_zone распределение:")
print(X_train_with_routes['climate_zone'].value_counts())

print("\nis_mountainous распределение:")
print(X_train_with_routes['is_mountainous'].value_counts())

print("\nusage_intensity распределение:")
print(X_train_with_routes['usage_intensity'].value_counts())

print("\n" + "="*60)
print("✅ ИНТЕГРАЦИЯ ЗАВЕРШЕНА!")
print("="*60)
print("""
Теперь используй в модели:
X_train = pd.read_csv('data/enriched/X_train_enriched.csv')
X_test = pd.read_csv('data/enriched/X_test_enriched.csv')

Новые признаки добавлены! 🚀
""")
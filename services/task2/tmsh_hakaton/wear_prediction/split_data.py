"""
split_data.py
Разделение данных на обучающую и проверочную выборки (80/20)
Сохраняем отдельные файлы для воспроизводимости
"""

import pandas as pd
import numpy as np
from sklearn.model_selection import train_test_split
import os

# ============== 1. ЗАГРУЖАЕМ ФИНАЛЬНЫЕ ДАННЫЕ ==============
print("📥 Загрузка подготовленных данных...")
df = pd.read_csv('data/processed/train_dataset_final.csv')

print(f"Всего записей: {len(df)}")
print(f"Колонки: {df.columns.tolist()}")
print(f"Целевая переменная: wear_intensity")

# ============== 2. ОТДЕЛЯЕМ ПРИЗНАКИ ОТ ЦЕЛЕВОЙ ==============
print("\n🔪 Разделяем на признаки (X) и целевую (y)...")
X = df.drop('wear_intensity', axis=1)
y = df['wear_intensity']

print(f"Размер X (признаки): {X.shape}")
print(f"Размер y (целевая): {y.shape}")

# ============== 3. ДЕЛИМ НА ТРЕНИРОВОЧНУЮ И ТЕСТОВУЮ ==============
print("\n✂️ Разделяем на train/test (80/20)...")
X_train, X_test, y_train, y_test = train_test_split(
    X, y,
    test_size=0.2,           # 20% на проверку
    random_state=42,         # фиксируем seed для воспроизводимости
    shuffle=True            # перемешиваем перед разделением
)

print(f"\n✅ РАЗДЕЛЕНИЕ ЗАВЕРШЕНО:")
print(f"   Обучающая выборка (80%): {len(X_train)} записей")
print(f"   Проверочная выборка (20%): {len(X_test)} записей")
print(f"   Всего: {len(X_train) + len(X_test)} записей")

# ============== 4. СОБИРАЕМ ОБРАТНО ПОЛНЫЕ ТАБЛИЦЫ ==============
print("\n📦 Формируем полные таблицы для сохранения...")
train_df = X_train.copy()
train_df['wear_intensity'] = y_train

test_df = X_test.copy()
test_df['wear_intensity'] = y_test

print(f"Train shape: {train_df.shape}")
print(f"Test shape: {test_df.shape}")

# ============== 5. ПРОВЕРЯЕМ ПРОПОРЦИИ ==============
print("\n📊 Проверка распределения целевой переменной:")
print(f"Обучающая выборка - среднее: {y_train.mean():.4f}, std: {y_train.std():.4f}")
print(f"Проверочная выборка - среднее: {y_test.mean():.4f}, std: {y_test.std():.4f}")
print(f"Разница в средних: {abs(y_train.mean() - y_test.mean()):.4f}")

# ============== 6. СОХРАНЯЕМ В ФАЙЛЫ ==============
print("\n💾 Сохранение в файлы...")
os.makedirs('data/splits', exist_ok=True)

train_df.to_csv('data/splits/train.csv', index=False)
test_df.to_csv('data/splits/test.csv', index=False)

# Также сохраняем отдельно X и y для удобства
pd.DataFrame(X_train).to_csv('data/splits/X_train.csv', index=False)
pd.DataFrame(X_test).to_csv('data/splits/X_test.csv', index=False)
pd.DataFrame(y_train, columns=['wear_intensity']).to_csv('data/splits/y_train.csv', index=False)
pd.DataFrame(y_test, columns=['wear_intensity']).to_csv('data/splits/y_test.csv', index=False)

print("\n✅ Файлы сохранены:")
print("   📁 data/splits/train.csv        - полная обучающая выборка")
print("   📁 data/splits/test.csv         - полная проверочная выборка")
print("   📁 data/splits/X_train.csv      - только признаки (обучение)")
print("   📁 data/splits/X_test.csv       - только признаки (проверка)")
print("   📁 data/splits/y_train.csv      - только целевая (обучение)")
print("   📁 data/splits/y_test.csv       - только целевая (проверка)")

# ============== 7. БЫСТРАЯ ПРОВЕРКА ==============
print("\n🔍 Быстрая проверка сохранённых файлов:")
check_train = pd.read_csv('data/splits/train.csv')
check_test = pd.read_csv('data/splits/test.csv')

print(f"train.csv - загружено: {check_train.shape}, wear_intensity: {'wear_intensity' in check_train.columns}")
print(f"test.csv - загружено: {check_test.shape}, wear_intensity: {'wear_intensity' in check_test.columns}")

# ============== 8. ИНСТРУКЦИЯ ДЛЯ МОДЕЛИ ==============
print("\n" + "="*60)
print("🎯 ГОТОВО К ОБУЧЕНИЮ!")
print("="*60)
print("""
Для обучения модели используй:

from sklearn.model_selection import train_test_split
import pandas as pd

# ВАРИАНТ 1: Загрузить готовые разделённые файлы
X_train = pd.read_csv('data/splits/X_train.csv')
X_test = pd.read_csv('data/splits/X_test.csv')
y_train = pd.read_csv('data/splits/y_train.csv').squeeze()
y_test = pd.read_csv('data/splits/y_test.csv').squeeze()

# ВАРИАНТ 2: Загрузить полные таблицы и разделить снова (не рекомендуется)
# df = pd.read_csv('data/processed/train_dataset_final.csv')
# X = df.drop('wear_intensity', axis=1)
# y = df['wear_intensity']
# X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

Категориальные признаки для модели:
- locomotive_series
- depo
- steel_num
""")
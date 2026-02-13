"""
prepare_data.py
Объединение данных о колёсах и ремонтах в одну таблицу
"""

import pandas as pd
import numpy as np
import os

# ============== 1. СОЗДАЁМ ПАПКИ ==============
os.makedirs('data/processed', exist_ok=True)

# ============== 2. ЗАГРУЖАЕМ ДАННЫЕ ==============
print("📥 Загрузка данных...")
wear = pd.read_csv('data/wear_data_train.csv')
service = pd.read_csv('data/service_dates.csv')

# ============== 3. СМОТРИМ, ЧТО ЗАГРУЗИЛИ ==============
print("\n" + "="*50)
print("ДАННЫЕ О КОЛЁСАХ (wear_data_train)")
print("="*50)
print(f"Размер: {wear.shape}")
print(f"Колонки: {wear.columns.tolist()}")
print("\nПервые 2 строки:")
print(wear.head(2))
print("\nИнформация:")
print(wear.info())

print("\n" + "="*50)
print("ДАННЫЕ О РЕМОНТАХ (service_dates)")
print("="*50)
print(f"Размер: {service.shape}")
print(f"Колонки: {service.columns.tolist()}")
print(f"\nУникальные типы ремонтов: {service['service_type'].unique()}")
print("\nПервые 2 строки:")
print(service.head(2))

# ============== 4. АГРЕГАЦИЯ РЕМОНТОВ ПО ЛОКОМОТИВАМ ==============
print("\n" + "="*50)
print("АГРЕГАЦИЯ РЕМОНТОВ")
print("="*50)

repair_stats = service.groupby(['locomotive_series', 'locomotive_number']).agg(
    # Общее количество ремонтов
    total_repairs=('service_type', 'count'),
    
    # Количество ремонтов каждого типа
    repair_type_1=('service_type', lambda x: (x == '1').sum()),
    repair_type_2=('service_type', lambda x: (x == '2').sum()),
    repair_type_3=('service_type', lambda x: (x == '3').sum()),
    turning_count=('service_type', lambda x: (x == 'обточка').sum()),
    
    # Количество уникальных дат ремонта
    unique_service_dates=('service_date', 'nunique'),
    
    # Первый и последний ремонт (может пригодиться)
    first_repair_date=('service_date', 'min'),
    last_repair_date=('service_date', 'max')
).reset_index()

print(f"Получили агрегированные данные по {len(repair_stats)} локомотивам")
print("\nПервые 2 строки агрегированных ремонтов:")
print(repair_stats.head(2))

# ============== 5. ПРОВЕРЯЕМ ДУБЛИКАТЫ ==============
print("\n" + "="*50)
print("ПРОВЕРКА ДУБЛИКАТОВ")
print("="*50)

wear_duplicates = wear['wheel_id'].duplicated().sum()
print(f"Дубликатов wheel_id в wear_data_train: {wear_duplicates}")

repair_duplicates = repair_stats.duplicated(subset=['locomotive_series', 'locomotive_number']).sum()
print(f"Дубликатов локомотивов в агрегированных ремонтах: {repair_duplicates}")

# ============== 6. СОЕДИНЯЕМ ТАБЛИЦЫ ==============
print("\n" + "="*50)
print("СОЕДИНЕНИЕ ТАБЛИЦ")
print("="*50)

# Left join: оставляем все колёса, добавляем информацию о ремонтах
train_data = wear.merge(
    repair_stats,
    on=['locomotive_series', 'locomotive_number'],
    how='left'
)

print(f"Размер до соединения: {wear.shape}")
print(f"Размер после соединения: {train_data.shape}")

# ============== 7. ОБРАБАТЫВАЕМ ПРОПУСКИ ==============
print("\n" + "="*50)
print("ОБРАБОТКА ПРОПУСКОВ")
print("="*50)

# Колонки, в которых могут быть пропуски (у колёс без ремонтов)
repair_columns = [
    'total_repairs', 'repair_type_1', 'repair_type_2', 'repair_type_3',
    'turning_count', 'unique_service_dates', 'first_repair_date', 'last_repair_date'
]

# Проверяем пропуски до обработки
print("Пропуски ДО обработки:")
print(train_data[repair_columns].isnull().sum())

# Заполняем числовые колонки нулями
for col in ['total_repairs', 'repair_type_1', 'repair_type_2', 'repair_type_3', 
            'turning_count', 'unique_service_dates']:
    train_data[col] = train_data[col].fillna(0).astype(int)

# Даты оставляем как есть (NaN значит "не было ремонтов")
print("\nПропуски ПОСЛЕ обработки:")
print(train_data[repair_columns].isnull().sum())

# ============== 8. ПРОВЕРЯЕМ РЕЗУЛЬТАТ ==============
print("\n" + "="*50)
print("ИТОГОВАЯ ТАБЛИЦА")
print("="*50)
print(f"Размер: {train_data.shape}")
print(f"Колонки: {train_data.columns.tolist()}")
print(f"\nТипы данных:")
print(train_data.dtypes)
print(f"\nПервые 3 строки:")
print(train_data.head(3))
print(f"\nСтатистика по числовым колонкам:")
print(train_data.describe())

# ============== 9. ПРОВЕРЯЕМ ЛОГИКУ ==============
print("\n" + "="*50)
print("ПРОВЕРКА ЛОГИКИ")
print("="*50)

# Проверяем 1: У колёс без ремонтов должны быть нули
no_repairs = train_data[train_data['total_repairs'] == 0]
print(f"Колёс без ремонтов: {len(no_repairs)}")
if len(no_repairs) > 0:
    print("Пример колеса без ремонтов:")
    print(no_repairs[['wheel_id', 'total_repairs', 'turning_count']].head(1))

# Проверяем 2: total_repairs = сумма ремонтов по типам + обточки
train_data['sum_repair_types'] = (train_data['repair_type_1'] + 
                                   train_data['repair_type_2'] + 
                                   train_data['repair_type_3'] + 
                                   train_data['turning_count'])
mismatch = train_data[train_data['total_repairs'] != train_data['sum_repair_types']]
print(f"\nНесовпадений total_repairs с суммой типов: {len(mismatch)}")
if len(mismatch) > 0:
    print("Пример несовпадения:")
    print(mismatch[['wheel_id', 'total_repairs', 'repair_type_1', 'repair_type_2', 
                    'repair_type_3', 'turning_count', 'sum_repair_types']].head(1))

# Удаляем служебную колонку
train_data = train_data.drop('sum_repair_types', axis=1)

# ============== 10. СОХРАНЯЕМ РЕЗУЛЬТАТ ==============
print("\n" + "="*50)
print("СОХРАНЕНИЕ")
print("="*50)

output_path = 'data/processed/train_dataset.csv'
train_data.to_csv(output_path, index=False)
print(f"✅ Данные сохранены в {output_path}")
print(f"Размер сохранённого файла: {train_data.shape}")

# ============== 11. КОРОТКИЙ ОТЧЁТ ==============
print("\n" + "="*50)
print("📊 ИТОГОВЫЙ ОТЧЁТ")
print("="*50)
print(f"Всего записей о колёсах: {len(train_data)}")
print(f"Уникальных локомотивов: {train_data['locomotive_number'].nunique()}")
print(f"Уникальных серий локомотивов: {train_data['locomotive_series'].nunique()}")
print(f"Уникальных депо: {train_data['depo'].nunique()}")
print(f"Уникальных плавок: {train_data['steel_num'].nunique()}")
print(f"\nСреднее количество ремонтов на локомотив: {train_data['total_repairs'].mean():.2f}")
print(f"Среднее количество обточек: {train_data['turning_count'].mean():.2f}")
print(f"Колёс без ремонтов: {len(train_data[train_data['total_repairs'] == 0])}")
print(f"Колёс с обточками: {len(train_data[train_data['turning_count'] > 0])}")
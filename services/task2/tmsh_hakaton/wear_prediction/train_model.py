import pandas as pd
from catboost import CatBoostRegressor
from sklearn.metrics import mean_squared_error
import numpy as np



# Загружаем признаки
X_train = pd.read_csv('data/splits/X_train.csv')
X_test = pd.read_csv('data/splits/X_test.csv')

# Загружаем целевую переменную и преобразуем в ряд
y_train = pd.read_csv('data/splits/y_train.csv').squeeze()
y_test = pd.read_csv('data/splits/y_test.csv').squeeze()


# 1. Удали явные дубликаты
print(f"Дубликатов ДО: {X_train.duplicated().sum()}")
X_train = X_train.drop_duplicates()
y_train = y_train.loc[X_train.index]  # Синхронизируем
print(f"Дубликатов ПОСЛЕ: {X_train.duplicated().sum()}")

# 2. Отсей выбросы (оставь 99 процентиль)
q99 = y_train.quantile(0.99)
mask = y_train <= q99
X_train = X_train[mask]
y_train = y_train[mask]
print(f"Удалено выбросов: {(~mask).sum()}")

# 3. Создай новые признаки
for df in [X_train, X_test]:
    # Интенсивность ремонтов
    df['repairs_per_100k'] = df['total_repairs'] / (df['mileage_start'] / 100000 + 1)
    # Интенсивность обточек
    df['turning_per_100k'] = df['turning_count'] / (df['mileage_start'] / 100000 + 1)
    # Доля обточек
    df['turning_ratio'] = df['turning_count'] / (df['total_repairs'] + 1)

# 4. Редкие категории в steel_num объедини в "other"
threshold = 100  # Минимум 100 примеров
value_counts = X_train['steel_num'].value_counts()
rare_values = value_counts[value_counts < threshold].index
X_train['steel_num'] = X_train['steel_num'].replace(rare_values, 'other')
X_test['steel_num'] = X_test['steel_num'].replace(rare_values, 'other')
print(f"Уникальных steel_num после объединения: {X_train['steel_num'].nunique()}")

# ====== ПРЕОБРАЗОВАНИЕ steel_num В СТРОКУ ======
print("Преобразование steel_num в строковый тип...")
X_train['steel_num'] = X_train['steel_num'].astype(str).str.replace('.0', '', regex=False).str.strip()
X_test['steel_num'] = X_test['steel_num'].astype(str).str.replace('.0', '', regex=False).str.strip()

# Проверка размеров
assert len(X_train) == len(y_train), "X_train и y_train разного размера!"
assert len(X_test) == len(y_test), "X_test и y_test разного размера!"

# Проверка пропусков
print(f"Пропуски в X_train: {X_train.isnull().sum().sum()}")
print(f"Пропуски в X_test: {X_test.isnull().sum().sum()}")
print(f"Пропуски в y_train: {y_train.isnull().sum()}")
print(f"Пропуски в y_test: {y_test.isnull().sum()}")

cat_features = ['locomotive_series', 'depo', 'steel_num']


for col in cat_features:
    if col in X_train.columns:
        print(f"  ✅ {col} - {X_train[col].dtype}")
    else:
        print(f"{col} - НЕ НАЙДЕН!")



model = CatBoostRegressor(
    iterations=2000,           # Увеличь
    learning_rate=0.1,        # Уменьши (медленнее, но точнее)
    depth=10,                    # Увеличь глубину
    cat_features=cat_features,
    eval_metric='RMSE',
    random_seed=42,
    verbose=100,
    early_stopping_rounds=100,
    l2_leaf_reg=5,              # Увеличь регуляризацию
    one_hot_max_size=100,       # Для steel_num с 7689 категориями
    bootstrap_type='Bernoulli',
    #subsample=0.7                # Уменьши для борьбы с переобучением
)


model.fit(
    X_train, y_train,
    eval_set=(X_test, y_test),
    plot=False,
    verbose=50
)



# Предсказания
y_pred_train = model.predict(X_train)
y_pred_test = model.predict(X_test)

# Метрики
mse_train = mean_squared_error(y_train, y_pred_train)
mse_test = mean_squared_error(y_test, y_pred_test)
rmse_train = np.sqrt(mse_train)
rmse_test = np.sqrt(mse_test)

print(f"MSE на обучении: {mse_train:.4f}")
print(f"MSE на проверке: {mse_test:.4f}")
print(f"RMSE на обучении: {rmse_train:.4f}")
print(f"RMSE на проверке: {rmse_test:.4f}")


# Качество модели
print(f"\n🎯 Качество модели (RMSE): {rmse_test:.4f}")
if rmse_test < 0.2:
    print("⭐ Отличная модель!")
elif rmse_test < 0.3:
    print("👍 Хорошая модель")
elif rmse_test < 0.4:
    print("👌 Приемлемая модель")
else:
    print("⚠️ Модель нужно улучшать")



# Получаем важность признаков
feature_importance = pd.DataFrame({
    'feature': X_train.columns,
    'importance': model.feature_importances_
}).sort_values('importance', ascending=False)

# Добавляем процент
feature_importance['importance_percent'] = (feature_importance['importance'] / feature_importance['importance'].sum() * 100).round(1)

# Выводим топ-10 признаков
print("Топ-10 самых важных признаков:")
for idx, row in feature_importance.head(10).iterrows():
    print(f"  {row['feature']:25} {row['importance']:8.0f} ({row['importance_percent']}%)")
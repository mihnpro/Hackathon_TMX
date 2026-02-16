import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import folium
from folium.plugins import MarkerCluster, HeatMap
import plotly.express as px
import plotly.graph_objects as go
from plotly.subplots import make_subplots
import seaborn as sns
from scipy import stats
from scipy.stats import f_oneway, kruskal
import warnings
warnings.filterwarnings('ignore')
from IPython.display import HTML
import os

# ==================== ЗАГРУЗКА ДАННЫХ ====================
print("="*60)
print("ЗАГРУЗКА ДАННЫХ")
print("="*60)

# Загружаем станции
stations_df = pd.read_csv('../station_info.csv')
print(f"Станций всего: {len(stations_df)}")
print(f"Станций с координатами: {stations_df['latitude'].notna().sum()}")

# Загружаем перемещения локомотивов
moves_df = pd.read_csv('../locomotives_displacement.csv')
print(f"Записей о перемещениях: {len(moves_df)}")
print(f"Уникальных локомотивов: {moves_df['locomotive_number'].nunique()}")
print(f"Уникальных серий: {moves_df['locomotive_series'].nunique()}")

# Загружаем данные об износе колес
wear_df = pd.read_csv('../wear_data_train.csv')
print(f"Записей об износе колес: {len(wear_df)}")
print(f"Уникальных колес: {wear_df['wheel_id'].nunique()}")
print(f"Уникальных депо: {wear_df['depo'].nunique()}")

# ==================== ПОДГОТОВКА ДАННЫХ ====================
print("\n" + "="*60)
print("ПОДГОТОВКА ДАННЫХ")
print("="*60)

# Очищаем координаты станций
stations_clean = stations_df.dropna(subset=['latitude', 'longitude']).copy()
# Преобразуем код станции в строку и удаляем возможные пробелы
stations_clean['station'] = stations_clean['station'].astype(str).str.strip()

# Преобразуем станции в перемещениях в строковый тип и удаляем пробелы
moves_df['station'] = moves_df['station'].astype(str).str.strip()
moves_df['depo_station'] = moves_df['depo_station'].astype(str).str.strip()

# Проверим форматы кодов станций
print(f"\nПримеры кодов станций из station_info.csv: {stations_clean['station'].iloc[:5].tolist()}")
print(f"Примеры кодов станций из locomotives_displacement.csv: {moves_df['station'].iloc[:5].tolist()}")

# Объединяем перемещения с координатами станций
moves_with_coords = moves_df.merge(
    stations_clean[['station', 'station_name', 'latitude', 'longitude']], 
    on='station', 
    how='inner'
)

print(f"\nПеремещений с координатами: {len(moves_with_coords)}")
print(f"Потеряно перемещений (нет координат): {len(moves_df) - len(moves_with_coords)}")

# Если нет совпадений, попробуем преобразовать коды в числовой формат
if len(moves_with_coords) == 0:
    print("\nПробуем альтернативный способ сопоставления...")
    # Преобразуем в числовой формат
    stations_clean['station_num'] = pd.to_numeric(stations_clean['station'], errors='coerce')
    moves_df['station_num'] = pd.to_numeric(moves_df['station'], errors='coerce')
    
    moves_with_coords = moves_df.merge(
        stations_clean[['station_num', 'station_name', 'latitude', 'longitude']], 
        on='station_num', 
        how='inner'
    )
    print(f"Перемещений с координатами (после преобразования): {len(moves_with_coords)}")

# Если все еще нет совпадений, создадим синтетические координаты на основе депо
if len(moves_with_coords) == 0:
    print("\nВНИМАНИЕ: Не удалось сопоставить станции. Используем координаты из депо...")
    
    # Создадим словарь депо с координатами из первых вхождений
    depo_coords = {}
    for idx, row in stations_clean.iterrows():
        if pd.notna(row['latitude']) and pd.notna(row['longitude']):
            depo_coords[row['station']] = (row['latitude'], row['longitude'])
    
    # Добавим координаты к перемещениям на основе depo_station
    moves_with_coords = moves_df.copy()
    moves_with_coords['latitude'] = moves_with_coords['depo_station'].map(
        lambda x: depo_coords.get(x, (None, None))[0] if x in depo_coords else None
    )
    moves_with_coords['longitude'] = moves_with_coords['depo_station'].map(
        lambda x: depo_coords.get(x, (None, None))[1] if x in depo_coords else None
    )
    moves_with_coords = moves_with_coords.dropna(subset=['latitude', 'longitude'])
    moves_with_coords['station_name'] = moves_with_coords['depo_station']
    
    print(f"Перемещений с координатами депо: {len(moves_with_coords)}")

# Создаем идентификаторы маршрутов для каждого локомотива
locomotive_routes = []
if len(moves_with_coords) > 1:
    for (series, number), group in moves_with_coords.groupby(['locomotive_series', 'locomotive_number']):
        group = group.sort_values('datetime')
        if len(group) >= 2:
            for i in range(len(group) - 1):
                locomotive_routes.append({
                    'locomotive_series': series,
                    'locomotive_number': number,
                    'from_station': group.iloc[i]['station'],
                    'from_lat': group.iloc[i]['latitude'],
                    'from_lon': group.iloc[i]['longitude'],
                    'to_station': group.iloc[i+1]['station'],
                    'to_lat': group.iloc[i+1]['latitude'],
                    'to_lon': group.iloc[i+1]['longitude'],
                    'datetime': group.iloc[i+1]['datetime'],
                    'route_id': f"{series}_{number}_{i}"
                })

routes_df = pd.DataFrame(locomotive_routes)
print(f"\nУникальных маршрутов (перегонов): {len(routes_df)}")

# Рассчитываем расстояние между станциями (по прямой)
def haversine_distance(lat1, lon1, lat2, lon2):
    R = 6371  # Радиус Земли в км
    lat1, lon1, lat2, lon2 = map(np.radians, [lat1, lon1, lat2, lon2])
    dlat = lat2 - lat1
    dlon = lon2 - lon1
    a = np.sin(dlat/2)**2 + np.cos(lat1) * np.cos(lat2) * np.sin(dlon/2)**2
    c = 2 * np.arcsin(np.sqrt(a))
    return R * c

if len(routes_df) > 0:
    routes_df['distance_km'] = routes_df.apply(
        lambda row: haversine_distance(row['from_lat'], row['from_lon'], 
                                       row['to_lat'], row['to_lon']), 
        axis=1
    )
    print(f"Средняя длина перегона: {routes_df['distance_km'].mean():.1f} км")
    print(f"Макс длина перегона: {routes_df['distance_km'].max():.1f} км")

# ==================== СТРАНИЦА 1: КАРТА СТАНЦИЙ ====================
print("\n" + "="*60)
print("СТРАНИЦА 1: СОЗДАНИЕ КАРТЫ СТАНЦИЙ")
print("="*60)

# Создаем карту со всеми станциями
m_stations = folium.Map(location=[55, 50], zoom_start=4, tiles='OpenStreetMap')

# Добавляем кластеры маркеров
marker_cluster = MarkerCluster().add_to(m_stations)

for idx, row in stations_clean.iterrows():
    popup_text = f"""
    <b>Код станции:</b> {row['station']}<br>
    <b>Название:</b> {row['station_name']}<br>
    <b>Координаты:</b> {row['latitude']:.4f}, {row['longitude']:.4f}
    """
    folium.Marker(
        [row['latitude'], row['longitude']],
        popup=folium.Popup(popup_text, max_width=300),
        icon=folium.Icon(color='blue', icon='train', prefix='fa')
    ).add_to(marker_cluster)

# Добавляем тепловую карту плотности станций
heat_data = [[row['latitude'], row['longitude']] for idx, row in stations_clean.iterrows()]
HeatMap(heat_data, radius=15, blur=10).add_to(m_stations)

m_stations.save('stations_map_with_heat.html')
print("✓ Карта станций сохранена: stations_map_with_heat.html")

# ==================== СТРАНИЦА 2: АНАЛИЗ ИЗНОСА ПО ДЕПО ====================
print("\n" + "="*60)
print("СТРАНИЦА 2: АНАЛИЗ ИЗНОСА КОЛЕС ПО ДЕПО")
print("="*60)

# Анализ износа по депо
depo_wear = wear_df.groupby('depo').agg({
    'wear_intensity': ['mean', 'std', 'count', 'median'],
    'mileage_start': 'mean'
}).round(3)

depo_wear.columns = ['средний_износ', 'std_износ', 'количество_колес', 'медиана_износа', 'средний_пробег']
depo_wear = depo_wear.sort_values('средний_износ', ascending=False)

print("Топ-10 депо с наибольшим износом:")
print(depo_wear.head(10)[['средний_износ', 'количество_колес', 'средний_пробег']])

# Визуализация износа по депо
fig, axes = plt.subplots(2, 2, figsize=(15, 12))
fig.suptitle('Анализ износа колес по депо приписки', fontsize=16, fontweight='bold')

# График 1: Топ-20 депо по среднему износу
top20_depo = depo_wear.head(20).reset_index()
axes[0,0].barh(range(len(top20_depo)), top20_depo['средний_износ'], color='coral')
axes[0,0].set_yticks(range(len(top20_depo)))
axes[0,0].set_yticklabels(top20_depo['depo'].str.slice(0, 25))
axes[0,0].set_xlabel('Средняя интенсивность износа')
axes[0,0].set_title('Топ-20 депо по среднему износу колес')
axes[0,0].invert_yaxis()

# График 2: Распределение износа (гистограмма)
axes[0,1].hist(wear_df['wear_intensity'], bins=50, color='skyblue', edgecolor='black', alpha=0.7)
axes[0,1].axvline(wear_df['wear_intensity'].mean(), color='red', linestyle='--', 
                  label=f"Среднее: {wear_df['wear_intensity'].mean():.3f}")
axes[0,1].axvline(wear_df['wear_intensity'].median(), color='green', linestyle='--', 
                  label=f"Медиана: {wear_df['wear_intensity'].median():.3f}")
axes[0,1].set_xlabel('Интенсивность износа')
axes[0,1].set_ylabel('Частота')
axes[0,1].set_title('Распределение интенсивности износа')
axes[0,1].legend()

# График 3: Зависимость износа от пробега
axes[1,0].scatter(wear_df['mileage_start'], wear_df['wear_intensity'], 
                  alpha=0.5, s=10, c='purple')
if len(wear_df) > 1:
    z = np.polyfit(wear_df['mileage_start'], wear_df['wear_intensity'], 1)
    p = np.poly1d(z)
    axes[1,0].plot(sorted(wear_df['mileage_start']), 
                    p(sorted(wear_df['mileage_start'])), 
                    'r--', linewidth=2, label=f'Тренд: y={z[0]:.2e}x+{z[1]:.3f}')
axes[1,0].set_xlabel('Начальный пробег (км)')
axes[1,0].set_ylabel('Интенсивность износа')
axes[1,0].set_title('Зависимость износа от пробега')
axes[1,0].legend()

# График 4: Боксплоты для топ-10 депо
top10_depo = wear_df[wear_df['depo'].isin(depo_wear.head(10).index)]
if len(top10_depo) > 0:
    depo_order = top10_depo.groupby('depo')['wear_intensity'].median().sort_values(ascending=False).index
    sns.boxplot(data=top10_depo, x='wear_intensity', y='depo', order=depo_order, ax=axes[1,1], palette='Reds')
    axes[1,1].set_xlabel('Интенсивность износа')
    axes[1,1].set_ylabel('Депо')
    axes[1,1].set_title('Распределение износа в топ-10 депо')
else:
    axes[1,1].text(0.5, 0.5, 'Недостаточно данных', ha='center', va='center')
    axes[1,1].set_title('Нет данных для боксплотов')

plt.tight_layout()
plt.savefig('depo_wear_analysis.png', dpi=150, bbox_inches='tight')
plt.show()
print("✓ График сохранен: depo_wear_analysis.png")

# ==================== СТРАНИЦА 3: АНАЛИЗ МАРШРУТОВ ====================
print("\n" + "="*60)
print("СТРАНИЦА 3: АНАЛИЗ МАРШРУТОВ")
print("="*60)

# Определяем регионы по долготе
def get_region(lon):
    if lon > 100:
        return 'Дальний Восток'
    elif lon > 60:
        return 'Сибирь'
    elif lon > 45:
        return 'Урал'
    elif lon > 35:
        return 'Поволжье'
    else:
        return 'Европейская часть'

stations_clean['region'] = stations_clean['longitude'].apply(get_region)

# Статистика по регионам
region_stats = stations_clean['region'].value_counts()
print("\nРаспределение станций по регионам:")
print(region_stats)

# Если есть данные о перемещениях с координатами
if len(moves_with_coords) > 0:
    # Анализ частоты посещения станций
    station_visits = moves_with_coords['station'].value_counts().reset_index()
    station_visits.columns = ['station', 'visits']
    station_visits = station_visits.merge(
        stations_clean[['station', 'station_name', 'latitude', 'longitude', 'region']], 
        on='station', how='left'
    )
    
    print("\nТоп-10 самых посещаемых станций:")
    print(station_visits.head(10)[['station_name', 'visits', 'region']].to_string(index=False))
    
    # Визуализация активности на карте
    m_routes = folium.Map(location=[55, 50], zoom_start=4, tiles='OpenStreetMap')
    
    # Добавляем тепловую карту посещений
    heat_visits = []
    for idx, row in station_visits.dropna().iterrows():
        weight = min(1, row['visits'] / station_visits['visits'].max())
        heat_visits.append([row['latitude'], row['longitude'], weight])
    
    if heat_visits:
        HeatMap(heat_visits, radius=20, blur=15, min_opacity=0.3).add_to(m_routes)
    
    # Добавляем маркеры для топ-20 станций
    top20_stations = station_visits.head(20).dropna()
    for idx, row in top20_stations.iterrows():
        popup_text = f"""
        <b>{row['station_name']}</b><br>
        <b>Код:</b> {row['station']}<br>
        <b>Посещений:</b> {row['visits']}<br>
        <b>Регион:</b> {row['region']}
        """
        folium.Marker(
            [row['latitude'], row['longitude']],
            popup=folium.Popup(popup_text, max_width=250),
            icon=folium.Icon(color='red', icon='flag', prefix='fa')
        ).add_to(m_routes)
    
    m_routes.save('stations_activity_map.html')
    print("✓ Карта активности станций сохранена: stations_activity_map.html")
else:
    print("\nНет данных о перемещениях с координатами для построения карты активности")
    # Создадим простую карту с регионами
    m_regions = folium.Map(location=[55, 50], zoom_start=4, tiles='OpenStreetMap')
    
    # Добавляем станции с группировкой по регионам
    for region in stations_clean['region'].unique():
        region_stations = stations_clean[stations_clean['region'] == region]
        if len(region_stations) > 0:
            fg = folium.FeatureGroup(name=region)
            for idx, row in region_stations.iterrows():
                folium.CircleMarker(
                    [row['latitude'], row['longitude']],
                    radius=3,
                    popup=row['station_name'],
                    color='blue',
                    fill=True
                ).add_to(fg)
            fg.add_to(m_regions)
    
    folium.LayerControl().add_to(m_regions)
    m_regions.save('stations_by_region.html')
    print("✓ Карта станций по регионам сохранена: stations_by_region.html")

# ==================== СТРАНИЦА 4: ИНТЕГРАЦИЯ С ИЗНОСОМ ====================
print("\n" + "="*60)
print("СТРАНИЦА 4: ИНТЕГРАЦИЯ ДАННЫХ ОБ ИЗНОСЕ И МАРШРУТАХ")
print("="*60)

# Анализ износа по сериям локомотивов
series_wear = wear_df.groupby('locomotive_series').agg({
    'wear_intensity': ['mean', 'std', 'count', 'median']
}).round(4)
series_wear.columns = ['mean_wear', 'std_wear', 'count', 'median_wear']
series_wear = series_wear.sort_values('mean_wear', ascending=False)

print("\nТоп-10 серий локомотивов с наибольшим износом:")
print(series_wear.head(10))

# Визуализация износа по сериям
fig, axes = plt.subplots(2, 2, figsize=(15, 12))
fig.suptitle('Анализ износа по сериям локомотивов', fontsize=16, fontweight='bold')

# График 1: Топ-20 серий по среднему износу
top20_series = series_wear.head(20).reset_index()
if len(top20_series) > 0:
    axes[0,0].bar(range(len(top20_series)), top20_series['mean_wear'], color='teal')
    axes[0,0].set_xticks(range(len(top20_series)))
    axes[0,0].set_xticklabels(top20_series['locomotive_series'], rotation=45, ha='right')
    axes[0,0].set_ylabel('Средняя интенсивность износа')
    axes[0,0].set_title('Топ-20 серий по среднему износу')
else:
    axes[0,0].text(0.5, 0.5, 'Нет данных', ha='center', va='center')

# График 2: Зависимость износа от количества наблюдений
if len(series_wear) > 0:
    axes[0,1].scatter(series_wear['count'], series_wear['mean_wear'], alpha=0.6, s=50)
    for idx, row in series_wear.head(5).iterrows():
        axes[0,1].annotate(idx, (row['count'], row['mean_wear']), fontsize=8)
    axes[0,1].set_xlabel('Количество наблюдений')
    axes[0,1].set_ylabel('Средний износ')
    axes[0,1].set_title('Зависимость износа от количества наблюдений')
    axes[0,1].set_xscale('log')
else:
    axes[0,1].text(0.5, 0.5, 'Нет данных', ha='center', va='center')

# График 3: Боксплоты для топ-10 серий
top10_series_names = series_wear.head(10).index.tolist()
top10_series = wear_df[wear_df['locomotive_series'].isin(top10_series_names)]
if len(top10_series) > 0:
    series_order = top10_series.groupby('locomotive_series')['wear_intensity'].median().sort_values(ascending=False).index
    sns.boxplot(data=top10_series, x='wear_intensity', y='locomotive_series', 
                order=series_order, ax=axes[1,0], palette='viridis')
    axes[1,0].set_xlabel('Интенсивность износа')
    axes[1,0].set_ylabel('Серия локомотива')
    axes[1,0].set_title('Распределение износа в топ-10 серий')
else:
    axes[1,0].text(0.5, 0.5, 'Нет данных', ha='center', va='center')

# График 4: Сравнение серий и депо (heatmap)
if len(wear_df) > 0 and len(series_wear) > 0 and len(depo_wear) > 0:
    try:
        series_depo_pivot = wear_df.pivot_table(
            values='wear_intensity', 
            index='locomotive_series', 
            columns='depo', 
            aggfunc='mean'
        )
        top_series = series_wear.head(10).index
        top_depo = depo_wear.head(10).index
        heatmap_data = series_depo_pivot.loc[top_series, top_depo]
        sns.heatmap(heatmap_data, annot=True, fmt='.2f', cmap='YlOrRd', ax=axes[1,1],
                    cbar_kws={'label': 'Средний износ'})
        axes[1,1].set_title('Износ по сериям и депо (топ-10)')
        axes[1,1].set_xlabel('Депо')
        axes[1,1].set_ylabel('Серия локомотива')
    except:
        axes[1,1].text(0.5, 0.5, 'Ошибка построения', ha='center', va='center')
        axes[1,1].set_title('Ошибка построения тепловой карты')
else:
    axes[1,1].text(0.5, 0.5, 'Недостаточно данных', ha='center', va='center')

plt.tight_layout()
plt.savefig('series_wear_analysis.png', dpi=150, bbox_inches='tight')
plt.show()
print("✓ График сохранен: series_wear_analysis.png")

# ==================== СТРАНИЦА 5: СТАТИСТИЧЕСКИЙ АНАЛИЗ ====================
print("\n" + "="*60)
print("СТРАНИЦА 5: СТАТИСТИЧЕСКИЙ АНАЛИЗ")
print("="*60)

# Статистический тест: различается ли износ в разных депо?
depo_groups = [group['wear_intensity'].values for name, group in wear_df.groupby('depo') if len(group) > 5]

if len(depo_groups) >= 2:
    # ANOVA тест
    f_stat, p_value_anova = f_oneway(*depo_groups)
    print(f"\n1. Однофакторный дисперсионный анализ (ANOVA):")
    print(f"   F-статистика: {f_stat:.4f}")
    print(f"   p-значение: {p_value_anova:.4e}")
    if p_value_anova < 0.05:
        print("   ВЫВОД: Существуют статистически значимые различия в износе между разными депо (p < 0.05)")
        print("   → Гипотеза подтверждается: депо (а значит и маршруты) влияют на износ")
    else:
        print("   ВЫВОД: Статистически значимых различий не обнаружено")

    # Тест Краскела-Уоллиса (непараметрический)
    h_stat, p_value_kw = kruskal(*depo_groups)
    print(f"\n2. Тест Краскела-Уоллиса (непараметрический):")
    print(f"   H-статистика: {h_stat:.4f}")
    print(f"   p-значение: {p_value_kw:.4e}")
    if p_value_kw < 0.05:
        print("   ВЫВОД: Подтверждается наличие значимых различий между депо")
    else:
        print("   ВЫВОД: Значимых различий не обнаружено")
else:
    print("\nНедостаточно групп для статистического теста")

# Анализ корреляции между пробегом и износом
if len(wear_df) > 1:
    corr_pearson, p_pearson = stats.pearsonr(wear_df['mileage_start'], wear_df['wear_intensity'])
    corr_spearman, p_spearman = stats.spearmanr(wear_df['mileage_start'], wear_df['wear_intensity'])

    print(f"\n3. Корреляция между пробегом и износом:")
    print(f"   Пирсон: r = {corr_pearson:.4f}, p = {p_pearson:.4e}")
    print(f"   Спирмен: r = {corr_spearman:.4f}, p = {p_spearman:.4e}")

    if abs(corr_pearson) < 0.1:
        print("   ВЫВОД: Корреляция очень слабая → пробег сам по себе не определяет износ")
        print("   Важнее условия эксплуатации (маршруты)")

# ==================== СТРАНИЦА 6: ВЫЯВЛЕНИЕ ПРОБЛЕМНЫХ ВЕТОК ====================
print("\n" + "="*60)
print("СТРАНИЦА 6: ВЫЯВЛЕНИЕ ПРОБЛЕМНЫХ МАРШРУТОВ")
print("="*60)

# Определяем депо с повышенным износом
wear_threshold = depo_wear['средний_износ'].quantile(0.75)
high_wear_depos = depo_wear[depo_wear['средний_износ'] > wear_threshold].index.tolist()
print(f"Депо с высоким износом (>75-й процентиль, >{wear_threshold:.3f}): {len(high_wear_depos)}")
print("\nТоп-5 депо с наибольшим износом:")
for depo in high_wear_depos[:5]:
    wear_val = depo_wear.loc[depo, 'средний_износ']
    count = depo_wear.loc[depo, 'количество_колес']
    print(f"  • {depo}: износ = {wear_val:.3f} (наблюдений: {count})")

# Создаем карту проблемных депо
m_problem = folium.Map(location=[55, 50], zoom_start=4, tiles='OpenStreetMap')

# Добавляем проблемные депо на карту
for depo in high_wear_depos[:10]:
    # Ищем координаты для депо в станциях
    depo_stations = stations_clean[stations_clean['station_name'].str.contains(depo[-10:], na=False)]
    if len(depo_stations) > 0:
        for idx, row in depo_stations.iterrows():
            wear_val = depo_wear.loc[depo, 'средний_износ']
            popup_text = f"""
            <b>{depo}</b><br>
            <b>Средний износ:</b> {wear_val:.3f}<br>
            <b>Колес в выборке:</b> {depo_wear.loc[depo, 'количество_колес']}<br>
            <b>Станция:</b> {row['station_name']}
            """
            folium.Marker(
                [row['latitude'], row['longitude']],
                popup=folium.Popup(popup_text, max_width=300),
                icon=folium.Icon(color='red', icon='warning', prefix='fa')
            ).add_to(m_problem)

m_problem.save('problematic_depots.html')
print("✓ Карта проблемных депо сохранена: problematic_depots.html")

# ==================== ВЫВОДЫ И РЕКОМЕНДАЦИИ ====================
print("\n" + "="*60)
print("ВЫВОДЫ ПО ИССЛЕДОВАНИЮ")
print("="*60)

print("\n1. ПРОВЕРКА ГИПОТЕЗЫ:")
print("   Гипотеза: 'На износ локомотивных колес влияет ветка (маршрут)'")
if len(depo_groups) >= 2 and p_value_anova < 0.05:
    print("   ✓ ГИПОТЕЗА ПОДТВЕРЖДАЕТСЯ статистически (p < 0.05)")
    print("     Существуют значимые различия в износе между разными депо,")
    print("     что косвенно указывает на влияние маршрутов (веток)")
else:
    print("   ✗ ГИПОТЕЗА НЕ МОЖЕТ БЫТЬ ПОДТВЕРЖДЕНА на данном этапе")
    if len(depo_groups) < 2:
        print("     Недостаточно данных для статистического анализа")

print("\n2. ДЕПО С ПОВЫШЕННЫМ ИЗНОСОМ:")
depo_high = depo_wear.nlargest(5, 'средний_износ')[['средний_износ', 'количество_колес']]
for depo, row in depo_high.iterrows():
    print(f"   • {depo}: средний износ = {row['средний_износ']:.3f} (наблюдений: {row['количество_колес']})")

print("\n3. СЕРИИ ЛОКОМОТИВОВ С ПОВЫШЕННЫМ ИЗНОСОМ:")
series_high = series_wear.head(5)[['mean_wear', 'count']]
for series, row in series_high.iterrows():
    print(f"   • {series}: средний износ = {row['mean_wear']:.3f} (наблюдений: {row['count']})")

print("\n4. ВИЗУАЛИЗАЦИИ СОЗДАНЫ:")
print("   • stations_map_with_heat.html - карта всех станций")
print("   • stations_by_region.html - карта станций по регионам")
print("   • depo_wear_analysis.png - анализ износа по депо")
print("   • series_wear_analysis.png - анализ износа по сериям")
print("   • problematic_depots.html - карта проблемных депо")

print("\n5. РЕКОМЕНДАЦИИ:")
print("   • Провести детальный анализ маршрутов в депо с наибольшим износом:")
for depo in depo_high.head(2).index:
    print(f"     - {depo}")
print("   • Исследовать профиль пути (уклоны, кривые) на проблемных ветках")
print("   • Рассмотреть возможность модернизации пути или оптимизации")
print("     режимов ведения поездов на выявленных проблемных участках")
print("   • Обратить внимание на серии локомотивов с повышенным износом:")
for series in series_high.head(3).index:
    print(f"     - {series}")
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
from scipy import stats
from scipy.stats import pearsonr, spearmanr
from matplotlib.backends.backend_pdf import PdfPages
from sklearn.linear_model import LinearRegression
from sklearn.preprocessing import PolynomialFeatures
from sklearn.metrics import r2_score
import warnings
warnings.filterwarnings('ignore')

# Загружаем данные из файла
df = pd.read_csv('../wear_data_train.csv')

# Проверим данные
print(f"Всего записей: {len(df)}")
print(f"Диапазон начального пробега: от {df['mileage_start'].min():.0f} до {df['mileage_start'].max():.0f}")
print(f"Диапазон интенсивности изнашивания: от {df['wear_intensity'].min():.4f} до {df['wear_intensity'].max():.4f}")

# Создаем PDF файл с графиками
with PdfPages('mileage_analysis_results.pdf') as pdf:
    
    plt.style.use('seaborn-v0_8-darkgrid')
    
    # ============================================================
    # ГРАФИК 1: Общий анализ распределения и scatter plot
    # ============================================================
    fig, axes = plt.subplots(2, 2, figsize=(16, 12))
    
    # 1a. Гистограмма распределения начального пробега
    ax1 = axes[0, 0]
    ax1.hist(df['mileage_start'] / 1e6, bins=50, edgecolor='black', alpha=0.7, color='steelblue')
    ax1.set_title('Рисунок 1а. Распределение начального пробега локомотивов', 
                  fontsize=12, fontweight='bold')
    ax1.set_xlabel('Начальный пробег (млн км)', fontsize=10)
    ax1.set_ylabel('Частота', fontsize=10)
    ax1.axvline(df['mileage_start'].mean() / 1e6, color='red', linestyle='--', 
                linewidth=2, label=f'Среднее: {df["mileage_start"].mean()/1e6:.2f} млн км')
    ax1.axvline(df['mileage_start'].median() / 1e6, color='green', linestyle='--', 
                linewidth=2, label=f'Медиана: {df["mileage_start"].median()/1e6:.2f} млн км')
    ax1.legend(fontsize=9)
    ax1.grid(True, alpha=0.3)
    
    # 1b. Гистограмма распределения интенсивности изнашивания
    ax2 = axes[0, 1]
    ax2.hist(df['wear_intensity'], bins=50, edgecolor='black', alpha=0.7, color='coral')
    ax2.set_title('Рисунок 1б. Распределение интенсивности изнашивания', 
                  fontsize=12, fontweight='bold')
    ax2.set_xlabel('Интенсивность изнашивания', fontsize=10)
    ax2.set_ylabel('Частота', fontsize=10)
    ax2.axvline(df['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Среднее: {df["wear_intensity"].mean():.3f}')
    ax2.axvline(df['wear_intensity'].median(), color='green', linestyle='--', 
                linewidth=2, label=f'Медиана: {df["wear_intensity"].median():.3f}')
    ax2.legend(fontsize=9)
    ax2.grid(True, alpha=0.3)
    
    # 1c. Scatter plot (все данные, с альфа-каналом)
    ax3 = axes[1, 0]
    # Берем случайную выборку для читаемости (20% данных)
    sample_df = df.sample(frac=0.2, random_state=42)
    ax3.scatter(sample_df['mileage_start'] / 1e6, sample_df['wear_intensity'], 
                alpha=0.3, s=5, c='steelblue')
    ax3.set_title('Рисунок 1в. Зависимость интенсивности изнашивания от начального пробега\n(случайная выборка 20% данных)', 
                  fontsize=12, fontweight='bold')
    ax3.set_xlabel('Начальный пробег (млн км)', fontsize=10)
    ax3.set_ylabel('Интенсивность изнашивания', fontsize=10)
    ax3.grid(True, alpha=0.3)
    
    # 1d. Scatter plot с логарифмической шкалой
    ax4 = axes[1, 1]
    ax4.scatter(np.log1p(sample_df['mileage_start']), sample_df['wear_intensity'], 
                alpha=0.3, s=5, c='coral')
    ax4.set_title('Рисунок 1г. Зависимость интенсивности от логарифма начального пробега', 
                  fontsize=12, fontweight='bold')
    ax4.set_xlabel('log(Начальный пробег + 1)', fontsize=10)
    ax4.set_ylabel('Интенсивность изнашивания', fontsize=10)
    ax4.grid(True, alpha=0.3)
    
    plt.suptitle('Анализ связи начального пробега и интенсивности изнашивания', 
                 fontsize=16, fontweight='bold', y=1.02)
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)
    
    # ============================================================
    # ГРАФИК 2: Регрессионный анализ
    # ============================================================
    fig, axes = plt.subplots(2, 2, figsize=(16, 12))
    
    # Рассчитываем корреляции
    pearson_corr, pearson_p = pearsonr(df['mileage_start'], df['wear_intensity'])
    spearman_corr, spearman_p = spearmanr(df['mileage_start'], df['wear_intensity'])
    
    # 2a. Scatter plot с линией линейной регрессии
    ax1 = axes[0, 0]
    
    # Линейная регрессия
    X = df['mileage_start'].values.reshape(-1, 1) / 1e6
    y = df['wear_intensity'].values
    model = LinearRegression()
    model.fit(X, y)
    y_pred = model.predict(X)
    r2_linear = r2_score(y, y_pred)
    
    ax1.scatter(X.flatten(), y, alpha=0.1, s=1, c='steelblue')
    
    # Создаем точки для линии регрессии
    X_line = np.linspace(X.min(), X.max(), 100).reshape(-1, 1)
    y_line = model.predict(X_line)
    ax1.plot(X_line, y_line, 'r-', linewidth=2, label=f'Линейная регрессия (R²={r2_linear:.4f})')
    
    ax1.set_title('Рисунок 2а. Линейная регрессия', 
                  fontsize=12, fontweight='bold')
    ax1.set_xlabel('Начальный пробег (млн км)', fontsize=10)
    ax1.set_ylabel('Интенсивность изнашивания', fontsize=10)
    ax1.legend(fontsize=9)
    ax1.grid(True, alpha=0.3)
    
    # 2b. Полиномиальная регрессия (степень 2)
    ax2 = axes[0, 1]
    
    poly = PolynomialFeatures(degree=2)
    X_poly = poly.fit_transform(X)
    poly_model = LinearRegression()
    poly_model.fit(X_poly, y)
    y_poly_pred = poly_model.predict(X_poly)
    r2_poly = r2_score(y, y_poly_pred)
    
    ax2.scatter(X.flatten(), y, alpha=0.1, s=1, c='coral')
    
    X_line_poly = poly.transform(X_line)
    y_line_poly = poly_model.predict(X_line_poly)
    ax2.plot(X_line, y_line_poly, 'g-', linewidth=2, label=f'Полиномиальная регрессия (R²={r2_poly:.4f})')
    
    ax2.set_title('Рисунок 2б. Полиномиальная регрессия (степень 2)', 
                  fontsize=12, fontweight='bold')
    ax2.set_xlabel('Начальный пробег (млн км)', fontsize=10)
    ax2.set_ylabel('Интенсивность изнашивания', fontsize=10)
    ax2.legend(fontsize=9)
    ax2.grid(True, alpha=0.3)
    
    # 2c. Диаграмма рассеяния с цветовой кодировкой по плотности
    ax3 = axes[1, 0]
    
    # Создаем 2D гистограмму
    h = ax3.hist2d(X.flatten(), y, bins=50, cmap='viridis', alpha=0.8)
    plt.colorbar(h[3], ax=ax3, label='Количество наблюдений')
    
    ax3.set_title('Рисунок 2в. 2D гистограмма (плотность распределения)', 
                  fontsize=12, fontweight='bold')
    ax3.set_xlabel('Начальный пробег (млн км)', fontsize=10)
    ax3.set_ylabel('Интенсивность изнашивания', fontsize=10)
    
    # 2d. Box plot по группам пробега
    ax4 = axes[1, 1]
    
    # Разбиваем пробег на 10 равных групп
    df['mileage_group'] = pd.qcut(df['mileage_start'], q=10, labels=[f'G{i+1}' for i in range(10)])
    
    # Сортируем группы по среднему пробегу
    group_means = df.groupby('mileage_group')['mileage_start'].mean()
    ordered_groups = group_means.sort_values().index
    
    sns.boxplot(data=df, x='mileage_group', y='wear_intensity', 
                order=ordered_groups, ax=ax4, palette='coolwarm')
    ax4.set_title('Рисунок 2г. Распределение интенсивности по группам начального пробега\n(децильные группы)', 
                  fontsize=12, fontweight='bold')
    ax4.set_xlabel('Группа по начальному пробегу (от min к max)', fontsize=10)
    ax4.set_ylabel('Интенсивность изнашивания', fontsize=10)
    ax4.tick_params(axis='x', rotation=45, fontsize=8)
    ax4.axhline(y=df['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Общее среднее')
    ax4.legend(fontsize=9)
    ax4.grid(True, alpha=0.3)
    
    plt.suptitle(f'Регрессионный анализ\nПирсон: r={pearson_corr:.4f} (p={pearson_p:.4e}), Спирмен: ρ={spearman_corr:.4f} (p={spearman_p:.4e})', 
                 fontsize=14, fontweight='bold', y=1.02)
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)
    
    # ============================================================
    # ГРАФИК 3: Анализ по сегментам данных
    # ============================================================
    fig, axes = plt.subplots(2, 2, figsize=(16, 12))
    
    # 3a. Корреляция по сегментам пробега
    ax1 = axes[0, 0]
    
    # Разбиваем пробег на сегменты и считаем корреляцию в каждом
    mileage_bins = np.linspace(df['mileage_start'].min(), df['mileage_start'].max(), 20)
    df['mileage_bin'] = pd.cut(df['mileage_start'], bins=20)
    
    bin_corr = []
    bin_center = []
    
    for bin_name, bin_df in df.groupby('mileage_bin'):
        if len(bin_df) > 10:
            corr, _ = pearsonr(bin_df['mileage_start'], bin_df['wear_intensity'])
            center = (bin_df['mileage_start'].min() + bin_df['mileage_start'].max()) / 2 / 1e6
            bin_corr.append(corr)
            bin_center.append(center)
    
    ax1.plot(bin_center, bin_corr, 'o-', color='steelblue', markersize=4)
    ax1.axhline(y=0, color='red', linestyle='--', linewidth=1)
    ax1.axhline(y=pearson_corr, color='green', linestyle='--', linewidth=1, label=f'Общая корреляция: {pearson_corr:.3f}')
    ax1.set_title('Рисунок 3а. Корреляция Пирсона в скользящих окнах пробега', 
                  fontsize=12, fontweight='bold')
    ax1.set_xlabel('Начальный пробег (млн км)', fontsize=10)
    ax1.set_ylabel('Коэффициент корреляции', fontsize=10)
    ax1.legend(fontsize=9)
    ax1.grid(True, alpha=0.3)
    
    # 3b. Корреляция по отдельным сериям локомотивов
    ax2 = axes[0, 1]
    
    # Выбираем топ-15 серий по количеству наблюдений
    top_series = df['locomotive_series'].value_counts().nlargest(15).index
    series_corr = []
    
    for series in top_series:
        series_df = df[df['locomotive_series'] == series]
        if len(series_df) > 20:
            corr, p_val = pearsonr(series_df['mileage_start'], series_df['wear_intensity'])
            series_corr.append({
                'series': series,
                'correlation': corr,
                'p_value': p_val,
                'count': len(series_df)
            })
    
    series_corr_df = pd.DataFrame(series_corr).sort_values('correlation', ascending=False)
    
    if len(series_corr_df) > 0:
        colors = ['red' if abs(c) > 0.2 else 'orange' if abs(c) > 0.1 else 'steelblue' 
                  for c in series_corr_df['correlation']]
        bars = ax2.bar(range(len(series_corr_df)), series_corr_df['correlation'], 
                       color=colors, alpha=0.7)
        ax2.set_xticks(range(len(series_corr_df)))
        ax2.set_xticklabels(series_corr_df['series'], rotation=45, ha='right', fontsize=8)
        ax2.set_title('Рисунок 3б. Корреляция пробег-износ по сериям локомотивов', 
                      fontsize=12, fontweight='bold')
        ax2.set_ylabel('Коэффициент корреляции Пирсона', fontsize=10)
        ax2.axhline(y=0, color='black', linestyle='-', linewidth=1)
        ax2.axhline(y=pearson_corr, color='green', linestyle='--', linewidth=1, label=f'Общая: {pearson_corr:.3f}')
        ax2.legend(fontsize=8)
        ax2.grid(True, alpha=0.3, axis='y')
    
    # 3c. Средняя интенсивность по группам пробега
    ax3 = axes[1, 0]
    
    # Группируем по квантилям пробега
    df['mileage_quantile'] = pd.qcut(df['mileage_start'], q=10, labels=[f'{i*10}-{(i+1)*10}%' for i in range(10)])
    quantile_stats = df.groupby('mileage_quantile')['wear_intensity'].agg(['mean', 'std', 'count']).reset_index()
    
    x_pos = np.arange(len(quantile_stats))
    ax3.bar(x_pos, quantile_stats['mean'], yerr=quantile_stats['std']/np.sqrt(quantile_stats['count'])*1.96,
            capsize=5, color='steelblue', alpha=0.7)
    ax3.set_xticks(x_pos)
    ax3.set_xticklabels(quantile_stats['mileage_quantile'], rotation=45, ha='right', fontsize=8)
    ax3.set_title('Рисунок 3в. Средняя интенсивность изнашивания по квантилям пробега\n(с 95% доверительными интервалами)', 
                  fontsize=12, fontweight='bold')
    ax3.set_xlabel('Квантильная группа пробега', fontsize=10)
    ax3.set_ylabel('Средняя интенсивность изнашивания', fontsize=10)
    ax3.axhline(y=df['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Общее среднее')
    ax3.legend(fontsize=9)
    ax3.grid(True, alpha=0.3, axis='y')
    
    # 3d. LOESS сглаживание (аппроксимация)
    ax4 = axes[1, 1]
    
    # Используем локальную регрессию через lowess
    from statsmodels.nonparametric.smoothers_lowess import lowess
    
    # Берем выборку для lowess (10% данных для производительности)
    sample_lowess = df.sample(frac=0.1, random_state=42)
    X_sample = sample_lowess['mileage_start'].values / 1e6
    y_sample = sample_lowess['wear_intensity'].values
    
    # Применяем lowess
    lowess_result = lowess(y_sample, X_sample, frac=0.3, return_sorted=True)
    
    ax4.scatter(X_sample, y_sample, alpha=0.1, s=2, c='lightgray')
    ax4.plot(lowess_result[:, 0], lowess_result[:, 1], 'r-', linewidth=2, label='LOESS сглаживание')
    
    ax4.set_title('Рисунок 3г. Непараметрическое сглаживание (LOESS)', 
                  fontsize=12, fontweight='bold')
    ax4.set_xlabel('Начальный пробег (млн км)', fontsize=10)
    ax4.set_ylabel('Интенсивность изнашивания', fontsize=10)
    ax4.legend(fontsize=9)
    ax4.grid(True, alpha=0.3)
    
    plt.suptitle('Детальный анализ связи пробега и износа', 
                 fontsize=16, fontweight='bold', y=1.02)
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)
    
    # ============================================================
    # ГРАФИК 4: Итоговый вывод и статистика
    # ============================================================
    fig, axes = plt.subplots(1, 2, figsize=(16, 8))
    
    # 4a. Финальный график с выводами
    ax1 = axes[0]
    
    # Создаем сводный график с регрессией и доверительным интервалом
    from scipy import stats as scipy_stats
    
    # Линейная регрессия с доверительным интервалом
    X_plot = df['mileage_start'].values / 1e6
    y_plot = df['wear_intensity'].values
    
    slope, intercept, r_value, p_value, std_err = scipy_stats.linregress(X_plot, y_plot)
    
    ax1.scatter(X_plot, y_plot, alpha=0.05, s=1, c='steelblue')
    
    X_line = np.linspace(X_plot.min(), X_plot.max(), 100)
    y_line = slope * X_line + intercept
    ax1.plot(X_line, y_line, 'r-', linewidth=2, label=f'Линейная регрессия')
    
    # Добавляем доверительный интервал
    n = len(X_plot)
    mean_x = np.mean(X_plot)
    std_x = np.std(X_plot)
    
    # Стандартная ошибка предсказания
    se_pred = std_err * np.sqrt(1/n + (X_line - mean_x)**2 / ((n-1) * std_x**2))
    ci = 1.96 * se_pred
    
    ax1.fill_between(X_line, y_line - ci, y_line + ci, color='red', alpha=0.2, label='95% доверительный интервал')
    
    ax1.set_title('Рисунок 4а. Финальная модель: линейная регрессия с доверительным интервалом', 
                  fontsize=14, fontweight='bold')
    ax1.set_xlabel('Начальный пробег (млн км)', fontsize=11)
    ax1.set_ylabel('Интенсивность изнашивания', fontsize=11)
    ax1.legend(fontsize=10)
    ax1.grid(True, alpha=0.3)
    
    # Добавляем текст с параметрами модели
    textstr = f'Уравнение регрессии:\ny = {slope:.6f}x + {intercept:.4f}\n'
    textstr += f'R² = {r_value**2:.6f}\n'
    textstr += f'p-value = {p_value:.6e}'
    ax1.text(0.05, 0.95, textstr, transform=ax1.transAxes, fontsize=10,
             verticalalignment='top', bbox=dict(boxstyle='round', facecolor='white', alpha=0.8))
    
    # 4b. Итоговая статистика
    ax2 = axes[1]
    ax2.axis('off')
    
    # Рассчитываем дополнительные метрики
    r2 = pearson_corr**2
    
    stats_text = f"""
    ===============================================
    РЕЗУЛЬТАТЫ СТАТИСТИЧЕСКОГО АНАЛИЗА
    ===============================================
    
    КОРРЕЛЯЦИОННЫЙ АНАЛИЗ:
    
    • Пирсон:
      - Коэффициент: {pearson_corr:.6f}
      - p-значение: {pearson_p:.6e}
      - R² (коэф. детерминации): {r2:.6f}
    
    • Спирмен (ранговая корреляция):
      - Коэффициент: {spearman_corr:.6f}
      - p-значение: {spearman_p:.6e}
    
    ===============================================
    РЕГРЕССИОННЫЙ АНАЛИЗ:
    
    • Линейная регрессия:
      - Наклон: {slope:.6f}
      - Свободный член: {intercept:.4f}
      - R²: {r2:.6f}
      - p-value: {p_value:.6e}
    
    • Полиномиальная (степень 2):
      - R²: {r2_poly:.6f}
    
    ===============================================
    СТАТИСТИЧЕСКИЕ ТЕСТЫ:
    
    H₀: Нет линейной связи между пробегом и износом
    H₁: Есть линейная связь
    
    Уровень значимости α = 0.05
    
    {'✅ H₀ ОТВЕРГАЕТСЯ' if pearson_p < 0.05 else '❌ H₀ ПРИНИМАЕТСЯ'}
    {'(связь статистически значима)' if pearson_p < 0.05 else '(связь не является статистически значимой)'}
    
    ===============================================
    ПРАКТИЧЕСКАЯ ЗНАЧИМОСТЬ:
    
    • Доля объясненной дисперсии: {r2*100:.2f}%
    • Сила связи по шкале Чеддока: 
      {'очень слабая' if abs(pearson_corr) < 0.1 else 
       'слабая' if abs(pearson_corr) < 0.3 else 
       'умеренная' if abs(pearson_corr) < 0.5 else 
       'заметная' if abs(pearson_corr) < 0.7 else 
       'высокая' if abs(pearson_corr) < 0.9 else 
       'очень высокая'}
    """
    
    ax2.text(0.05, 0.95, stats_text, transform=ax2.transAxes, fontsize=10,
             verticalalignment='top', fontfamily='monospace',
             bbox=dict(boxstyle='round', facecolor='lightyellow', alpha=0.9))
    
    plt.suptitle('ИТОГОВЫЙ ВЫВОД: Связь начального пробега и интенсивности изнашивания', 
                 fontsize=16, fontweight='bold', y=1.02)
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)

print("\n" + "="*60)
print("PDF файл 'mileage_analysis_results.pdf' успешно создан!")
print("В файле 4 страницы с графиками для анализа связи пробега и износа")
print("="*60)

# Финальные выводы
print("\n" + "="*60)
print("ИТОГОВЫЕ РЕЗУЛЬТАТЫ ПО ГИПОТЕЗЕ")
print("="*60)

print(f"""
КОРРЕЛЯЦИОННЫЙ АНАЛИЗ:
------------------------
Пирсон: r = {pearson_corr:.6f}, p-value = {pearson_p:.6e}
Спирмен: ρ = {spearman_corr:.6f}, p-value = {spearman_p:.6e}
R² (коэффициент детерминации) = {pearson_corr**2:.6f} ({pearson_corr**2*100:.2f}%)

{"="*60}
ВЫВОД ПО ГИПОТЕЗЕ:
{"="*60}
""")

if abs(pearson_corr) < 0.05:
    print("❌ ГИПОТЕЗА ОТВЕРГНУТА")
    print("   Начальный пробег локомотива НЕ связан с интенсивностью изнашивания.")
    print(f"   Коэффициент корреляции близок к нулю ({pearson_corr:.4f}),")
    print(f"   что указывает на отсутствие линейной зависимости.")
elif abs(pearson_corr) < 0.1:
    print("⚠️ ГИПОТЕЗА НЕ ПОДТВЕРЖДАЕТСЯ")
    print("   Связь между начальным пробегом и интенсивностью изнашивания")
    print("   является статистически значимой, но очень слабой.")
    print(f"   Корреляция: r={pearson_corr:.4f}, p={pearson_p:.4e}")
    print(f"   Доля объясненной дисперсии: всего {pearson_corr**2*100:.2f}%")
elif pearson_p < 0.05:
    print("✅ ГИПОТЕЗА ПОДТВЕРЖДЕНА (статистически)")
    print("   Существует статистически значимая связь между начальным пробегом")
    print("   и интенсивностью изнашивания.")
    print(f"   Корреляция: r={pearson_corr:.4f}, p={pearson_p:.4e}")
    print(f"   Доля объясненной дисперсии: {pearson_corr**2*100:.2f}%")
    
    if abs(pearson_corr) > 0.3:
        print("\n   ПРАКТИЧЕСКАЯ ЗНАЧИМОСТЬ:")
        print("   Связь достаточно сильная для практического использования")
        print("   в прогнозировании и планировании ремонтов.")
    else:
        print("\n   ПРАКТИЧЕСКАЯ ЗНАЧИМОСТЬ:")
        print("   Несмотря на статистическую значимость, связь слабая")
        print("   и может не иметь практической ценности.")
else:
    print("❌ ГИПОТЕЗА ОТВЕРГНУТА")
    print("   Статистически значимой связи не обнаружено.")
    print(f"   p-value = {pearson_p:.4f} > 0.05")

print("\n" + "="*60)
print("РЕКОМЕНДАЦИИ:")
print("="*60)
print("""
1. Для прогнозирования износа следует использовать более информативные факторы:
   - Серия локомотива (как показал предыдущий анализ)
   - Номер плавки стали (при наличии достаточных данных)
   - Условия эксплуатации (депо, маршруты)

2. Начальный пробег может использоваться как вспомогательный фактор,
   но не как основной предиктор интенсивности изнашивания.

3. Рекомендуется разработать многофакторные модели, учитывающие:
   - Серию локомотива (категориальный фактор)
   - Пробег (количественный фактор)
   - Характеристики стали (при наличии)
""")
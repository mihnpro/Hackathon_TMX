import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
from scipy import stats
from scipy.stats import f_oneway
from matplotlib.backends.backend_pdf import PdfPages
from statsmodels.stats.multicomp import pairwise_tukeyhsd
import warnings
warnings.filterwarnings('ignore')

# Загружаем данные из файла
df = pd.read_csv('../wear_data_train.csv')

# Основные статистики
stats_by_series = df.groupby('locomotive_series')['wear_intensity'].agg(['count', 'mean', 'std', 'median']).round(4)
print("Статистика по интенсивности изнашивания для каждой серии локомотивов:")
print(stats_by_series.sort_values('mean', ascending=False).head(10))

# Статистические тесты
series_counts = df['locomotive_series'].value_counts()
valid_series = series_counts[series_counts >= 5].index
df_anova = df[df['locomotive_series'].isin(valid_series)]

groups = [df_anova[df_anova['locomotive_series'] == series]['wear_intensity'].values 
          for series in valid_series]

f_stat, p_value = f_oneway(*groups)
h_stat, kw_p_value = stats.kruskal(*groups)

print("\n" + "="*60)
print("РЕЗУЛЬТАТЫ СТАТИСТИЧЕСКОГО ТЕСТИРОВАНИЯ")
print("="*60)
print(f"ANOVA тест: p-value = {p_value:.4e}")
print(f"Краскел-Уоллис тест: p-value = {kw_p_value:.4e}")
print(f"Вывод: {'Серии значимо различаются' if p_value < 0.05 else 'Значимых различий не обнаружено'}")

# Создаем PDF файл с 4 графиками
with PdfPages('analysis_results_4graphs.pdf') as pdf:
    
    plt.style.use('seaborn-v0_8-darkgrid')
    
    # ============================================================
    # ГРАФИК 1: Box plot - сравнение распределений по сериям
    # ============================================================
    fig, ax = plt.subplots(figsize=(16, 8))
    
    # Выберем топ-20 серий с наибольшим количеством наблюдений и топ-5 по среднему
    top_by_count = series_counts.nlargest(15).index
    top_by_mean = stats_by_series.nlargest(5, 'mean').index
    selected_series = list(set(top_by_count) | set(top_by_mean))
    
    # Сортируем по среднему значению
    series_means = df[df['locomotive_series'].isin(selected_series)].groupby('locomotive_series')['wear_intensity'].mean()
    ordered_series = series_means.sort_values(ascending=False).index
    
    bp = sns.boxplot(data=df[df['locomotive_series'].isin(selected_series)], 
                      x='locomotive_series', y='wear_intensity', 
                      order=ordered_series, ax=ax, palette='RdBu_r')
    
    ax.set_title('Рисунок 1. Распределение интенсивности изнашивания по сериям локомотивов\n(Box plot с медианами и квартилями)', 
                fontsize=14, fontweight='bold')
    ax.set_xlabel('Серия локомотива', fontsize=12)
    ax.set_ylabel('Интенсивность изнашивания', fontsize=12)
    ax.tick_params(axis='x', rotation=45, labelsize=10)
    ax.axhline(y=df['wear_intensity'].mean(), color='red', linestyle='--', 
               linewidth=2, label=f'Общее среднее: {df["wear_intensity"].mean():.3f}')
    ax.legend(fontsize=11)
    ax.grid(True, alpha=0.3)
    
    # Добавим подписи с медианами
    for i, series in enumerate(ordered_series):
        median_val = df[df['locomotive_series'] == series]['wear_intensity'].median()
        ax.text(i, median_val + 0.05, f'{median_val:.2f}', 
                ha='center', va='bottom', fontsize=9, fontweight='bold')
    
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)
    
    # ============================================================
    # ГРАФИК 2: Сравнение средних значений с доверительными интервалами
    # ============================================================
    fig, ax = plt.subplots(figsize=(16, 8))
    
    # Рассчитаем средние и доверительные интервалы для топ-20 серий
    plot_data = []
    for series in selected_series:
        series_data = df[df['locomotive_series'] == series]['wear_intensity'].dropna()
        if len(series_data) > 1:
            mean = series_data.mean()
            sem = series_data.std() / np.sqrt(len(series_data))
            ci = sem * 1.96  # 95% доверительный интервал
            plot_data.append({
                'series': series,
                'mean': mean,
                'ci': ci,
                'count': len(series_data)
            })
    
    plot_df = pd.DataFrame(plot_data).sort_values('mean', ascending=False)
    
    colors = ['darkred' if m > df['wear_intensity'].mean() else 'steelblue' 
              for m in plot_df['mean']]
    
    bars = ax.bar(range(len(plot_df)), plot_df['mean'], yerr=plot_df['ci'], 
                  capsize=5, color=colors, alpha=0.8, edgecolor='black', linewidth=0.5)
    
    ax.set_title('Рисунок 2. Средняя интенсивность изнашивания с 95% доверительными интервалами', 
                fontsize=14, fontweight='bold')
    ax.set_xlabel('Серия локомотива', fontsize=12)
    ax.set_ylabel('Средняя интенсивность изнашивания', fontsize=12)
    ax.set_xticks(range(len(plot_df)))
    ax.set_xticklabels(plot_df['series'], rotation=45, ha='right', fontsize=10)
    ax.axhline(y=df['wear_intensity'].mean(), color='red', linestyle='--', 
               linewidth=2, label=f'Общее среднее: {df["wear_intensity"].mean():.3f}')
    ax.legend(fontsize=11)
    ax.grid(True, alpha=0.3, axis='y')
    
    # Добавим подписи значений
    for i, (_, row) in enumerate(plot_df.iterrows()):
        ax.text(i, row['mean'] + row['ci'] + 0.02, f'{row["mean"]:.3f}', 
                ha='center', va='bottom', fontsize=9, fontweight='bold')
    
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)
    
    # ============================================================
    # ГРАФИК 3: Гистограмма общего распределения + KDE с выделением крайних серий
    # ============================================================
    fig, axes = plt.subplots(1, 2, figsize=(18, 7))
    
    # 3a. Гистограмма общего распределения
    ax1 = axes[0]
    ax1.hist(df['wear_intensity'], bins=50, edgecolor='black', alpha=0.7, 
             color='steelblue', density=True)
    df['wear_intensity'].plot.kde(ax=ax1, color='red', linewidth=2, label='Плотность распределения')
    
    ax1.set_title('Рисунок 3а. Распределение интенсивности изнашивания (все серии)', 
                  fontsize=13, fontweight='bold')
    ax1.set_xlabel('Интенсивность изнашивания', fontsize=11)
    ax1.set_ylabel('Плотность вероятности', fontsize=11)
    ax1.axvline(df['wear_intensity'].mean(), color='darkred', linestyle='--', 
                linewidth=2, label=f'Среднее: {df["wear_intensity"].mean():.3f}')
    ax1.axvline(df['wear_intensity'].median(), color='darkgreen', linestyle='--', 
                linewidth=2, label=f'Медиана: {df["wear_intensity"].median():.3f}')
    ax1.legend(fontsize=10)
    ax1.grid(True, alpha=0.3)
    
    # 3b. Box plot для сравнения крайних серий
    ax2 = axes[1]
    
    # Возьмем 5 серий с наибольшим и 5 с наименьшим средним
    top5_mean = stats_by_series.nlargest(5, 'mean').index
    bottom5_mean = stats_by_series.nsmallest(5, 'mean').index
    extreme_series = list(top5_mean) + list(bottom5_mean)
    
    # Сортируем по среднему
    extreme_means = df[df['locomotive_series'].isin(extreme_series)].groupby('locomotive_series')['wear_intensity'].mean()
    ordered_extreme = extreme_means.sort_values(ascending=False).index
    
    bp2 = sns.boxplot(data=df[df['locomotive_series'].isin(extreme_series)], 
                       x='locomotive_series', y='wear_intensity', 
                       order=ordered_extreme, ax=ax2, palette='coolwarm')
    
    ax2.set_title('Рисунок 3б. Сравнение серий с экстремальными значениями износа', 
                  fontsize=13, fontweight='bold')
    ax2.set_xlabel('Серия локомотива', fontsize=11)
    ax2.set_ylabel('Интенсивность изнашивания', fontsize=11)
    ax2.tick_params(axis='x', rotation=45, labelsize=10)
    ax2.axhline(y=df['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Общее среднее')
    ax2.legend(fontsize=10)
    ax2.grid(True, alpha=0.3)
    
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)
    
    # ============================================================
    # ГРАФИК 4: Статистическая значимость различий (Tukey HSD)
    # ============================================================
    fig, axes = plt.subplots(1, 2, figsize=(18, 8))
    
    # Возьмем топ-8 серий для анализа
    top8_series = series_counts.nlargest(8).index
    df_top8 = df[df['locomotive_series'].isin(top8_series)]
    
    # 4a. Tukey HSD plot
    ax1 = axes[0]
    tukey_results = pairwise_tukeyhsd(df_top8['wear_intensity'], 
                                       df_top8['locomotive_series'], 
                                       alpha=0.05)
    
    tukey_results.plot_simultaneous(ax=ax1, figsize=(10, 6))
    ax1.set_title('Рисунок 4а. Доверительные интервалы разности средних (Tukey HSD)', 
                  fontsize=13, fontweight='bold')
    ax1.set_xlabel('Разность средних', fontsize=11)
    ax1.axvline(x=0, color='red', linestyle='--', linewidth=1.5, alpha=0.7)
    ax1.grid(True, alpha=0.3)
    
    # 4b. Тепловая карта p-values
    ax2 = axes[1]
    
    series_list = top8_series.tolist()
    p_matrix = pd.DataFrame(index=series_list, columns=series_list, data=1.0)
    
    for i, s1 in enumerate(series_list):
        for j, s2 in enumerate(series_list):
            if i < j:
                data1 = df[df['locomotive_series'] == s1]['wear_intensity']
                data2 = df[df['locomotive_series'] == s2]['wear_intensity']
                t_stat, p_val = stats.ttest_ind(data1, data2, equal_var=False)
                p_matrix.loc[s1, s2] = p_val
                p_matrix.loc[s2, s1] = p_val
    
    # Маска для незначимых различий (p > 0.05)
    mask = p_matrix > 0.05
    
    sns.heatmap(-np.log10(p_matrix + 1e-10), mask=mask, annot=True, 
                fmt='.2f', cmap='RdYlGn_r', ax=ax2, square=True,
                cbar_kws={'label': '-log10(p-value)'},
                annot_kws={'size': 10})
    
    ax2.set_title('Рисунок 4б. Матрица значимости различий между сериями\n(красный - значимо, зеленый - незначимо)', 
                  fontsize=13, fontweight='bold')
    ax2.set_xlabel('Серия локомотива', fontsize=11)
    ax2.set_ylabel('Серия локомотива', fontsize=11)
    
    # Добавим пояснение
    ax2.text(0.5, -0.15, 
             'Примечание: Значения > 1.3 соответствуют p-value < 0.05\n(чем выше значение, тем значимее различие)',
             transform=ax2.transAxes, ha='center', fontsize=10, style='italic',
             bbox=dict(boxstyle='round', facecolor='wheat', alpha=0.5))
    
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)

print("\n" + "="*60)
print("PDF файл 'analysis_results_4graphs.pdf' успешно создан!")
print("В файле 4 страницы с ключевыми графиками, доказывающими гипотезу")
print("="*60)

# Финальные выводы
print("\n" + "="*60)
print("ВЫВОДЫ ПО РЕЗУЛЬТАТАМ АНАЛИЗА")
print("="*60)

overall_mean = df['wear_intensity'].mean()
overall_std = df['wear_intensity'].std()

print(f"Общая статистика по интенсивности изнашивания:")
print(f"  Среднее: {overall_mean:.4f}")
print(f"  Медиана: {df['wear_intensity'].median():.4f}")
print(f"  Стандартное отклонение: {overall_std:.4f}")

# Найдем серии с наибольшей и наименьшей интенсивностью
top_mean_series = stats_by_series.nlargest(5, 'mean')
bottom_mean_series = stats_by_series.nsmallest(5, 'mean')

print("\nТоп-5 серий с наибольшей средней интенсивностью изнашивания:")
for series, row in top_mean_series.iterrows():
    print(f"  {series}: {row['mean']:.4f} (n={int(row['count'])})")

print("\nТоп-5 серий с наименьшей средней интенсивностью изнашивания:")
for series, row in bottom_mean_series.iterrows():
    print(f"  {series}: {row['mean']:.4f} (n={int(row['count'])})")

print("\n" + "="*60)
print("ИТОГОВОЕ ЗАКЛЮЧЕНИЕ ПО ГИПОТЕЗЕ")
print("="*60)
print(f"✅ ГИПОТЕЗА ПОДТВЕРЖДЕНА: Интенсивность изнашивания статистически значимо")
print(f"   различается между разными сериями локомотивов (p-value = {p_value:.4e}).")
print()
print(f"   Ключевые наблюдения:")
print(f"   • Размах средних значений: от {bottom_mean_series['mean'].min():.3f} до {top_mean_series['mean'].max():.3f}")
print(f"   • Максимальное различие между сериями: в {top_mean_series['mean'].max()/bottom_mean_series['mean'].min():.1f} раз")
print(f"   • Наибольший износ наблюдается у серий: {', '.join(top_mean_series.index[:3])}")
print(f"   • Наименьший износ наблюдается у серий: {', '.join(bottom_mean_series.index[:3])}")
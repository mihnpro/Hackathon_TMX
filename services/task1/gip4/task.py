import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
from scipy import stats
from scipy.stats import f_oneway, kruskal
from matplotlib.backends.backend_pdf import PdfPages
from statsmodels.stats.multicomp import pairwise_tukeyhsd
import warnings
warnings.filterwarnings('ignore')

# Загружаем данные из файла
df = pd.read_csv('../wear_data_train.csv')

# Проверим данные
print(f"Всего записей: {len(df)}")
print(f"Количество уникальных депо: {df['depo'].nunique()}")

# Статистика по депо
depo_stats = df.groupby('depo')['wear_intensity'].agg(['count', 'mean', 'std', 'median', 'min', 'max']).round(4)
depo_stats = depo_stats.sort_values('count', ascending=False)
print("\nТоп-20 депо по количеству наблюдений:")
print(depo_stats.head(20))

# Выбираем депо с достаточным количеством наблюдений для анализа
min_observations = 30
valid_depos = depo_stats[depo_stats['count'] >= min_observations].index.tolist()
print(f"\nДепо с >= {min_observations} наблюдениями: {len(valid_depos)}")

# Создаем PDF файл с графиками
with PdfPages('depo_analysis_results.pdf') as pdf:
    
    plt.style.use('seaborn-v0_8-darkgrid')
    
    # ============================================================
    # ГРАФИК 1: Общий анализ распределения по депо
    # ============================================================
    fig, axes = plt.subplots(2, 2, figsize=(16, 12))
    
    # 1a. Топ-20 депо по количеству наблюдений
    ax1 = axes[0, 0]
    top20_depos = depo_stats.head(20)
    
    colors = ['steelblue' if i < 10 else 'coral' for i in range(20)]
    bars = ax1.bar(range(20), top20_depos['count'].values, color=colors, alpha=0.7)
    ax1.set_xticks(range(20))
    ax1.set_xticklabels(top20_depos.index, rotation=45, ha='right', fontsize=8)
    ax1.set_title('Рисунок 1а. Топ-20 депо по количеству наблюдений', 
                  fontsize=12, fontweight='bold')
    ax1.set_xlabel('Депо', fontsize=10)
    ax1.set_ylabel('Количество наблюдений', fontsize=10)
    ax1.grid(True, alpha=0.3, axis='y')
    
    # Добавим подписи значений
    for i, v in enumerate(top20_depos['count'].values):
        ax1.text(i, v + 5, str(v), ha='center', va='bottom', fontsize=7)
    
    # 1b. Распределение интенсивности изнашивания (все данные)
    ax2 = axes[0, 1]
    ax2.hist(df['wear_intensity'], bins=50, edgecolor='black', alpha=0.7, color='steelblue')
    ax2.set_title('Рисунок 1б. Распределение интенсивности изнашивания (все данные)', 
                  fontsize=12, fontweight='bold')
    ax2.set_xlabel('Интенсивность изнашивания', fontsize=10)
    ax2.set_ylabel('Частота', fontsize=10)
    ax2.axvline(df['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Среднее: {df["wear_intensity"].mean():.3f}')
    ax2.axvline(df['wear_intensity'].median(), color='green', linestyle='--', 
                linewidth=2, label=f'Медиана: {df["wear_intensity"].median():.3f}')
    ax2.legend(fontsize=9)
    ax2.grid(True, alpha=0.3)
    
    # 1c. Box plot для топ-15 депо по количеству наблюдений
    ax3 = axes[1, 0]
    top15_depos = depo_stats.head(15).index.tolist()
    df_top15 = df[df['depo'].isin(top15_depos)]
    
    # Сортируем депо по средней интенсивности
    depo_means = df_top15.groupby('depo')['wear_intensity'].mean()
    ordered_depos = depo_means.sort_values(ascending=False).index
    
    sns.boxplot(data=df_top15, x='depo', y='wear_intensity', 
                order=ordered_depos, ax=ax3, palette='RdBu_r')
    ax3.set_title('Рисунок 1в. Распределение интенсивности изнашивания по топ-15 депо', 
                  fontsize=12, fontweight='bold')
    ax3.set_xlabel('Депо', fontsize=10)
    ax3.set_ylabel('Интенсивность изнашивания', fontsize=10)
    ax3.tick_params(axis='x', rotation=45, labelsize=8)  # Исправлено: fontsize -> labelsize
    ax3.axhline(y=df['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Общее среднее')
    ax3.legend(fontsize=9)
    ax3.grid(True, alpha=0.3)
    
    # 1d. Violin plot для топ-15 депо
    ax4 = axes[1, 1]
    sns.violinplot(data=df_top15, x='depo', y='wear_intensity', 
                   order=ordered_depos, ax=ax4, palette='mako')
    ax4.set_title('Рисунок 1г. Violin plot для топ-15 депо', 
                  fontsize=12, fontweight='bold')
    ax4.set_xlabel('Депо', fontsize=10)
    ax4.set_ylabel('Интенсивность изнашивания', fontsize=10)
    ax4.tick_params(axis='x', rotation=45, labelsize=8)  # Исправлено: fontsize -> labelsize
    ax4.axhline(y=df['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Общее среднее')
    ax4.legend(fontsize=9)
    ax4.grid(True, alpha=0.3)
    
    plt.suptitle('Анализ влияния депо приписки на интенсивность изнашивания', 
                 fontsize=16, fontweight='bold', y=1.02)
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)
    
    # ============================================================
    # ГРАФИК 2: Сравнение средних значений по депо
    # ============================================================
    fig, axes = plt.subplots(2, 2, figsize=(16, 12))
    
    # 2a. Средние значения с доверительными интервалами
    ax1 = axes[0, 0]
    
    # Рассчитываем средние и доверительные интервалы для топ-20 депо
    plot_data = []
    for depo in depo_stats.head(20).index:
        depo_data = df[df['depo'] == depo]['wear_intensity'].dropna()
        if len(depo_data) > 1:
            mean = depo_data.mean()
            sem = depo_data.std() / np.sqrt(len(depo_data))
            ci = sem * 1.96  # 95% доверительный интервал
            plot_data.append({
                'depo': depo,
                'mean': mean,
                'ci': ci,
                'count': len(depo_data)
            })
    
    plot_df = pd.DataFrame(plot_data).sort_values('mean', ascending=False)
    
    colors = ['darkred' if m > df['wear_intensity'].mean() else 'steelblue' 
              for m in plot_df['mean']]
    
    bars = ax1.bar(range(len(plot_df)), plot_df['mean'], yerr=plot_df['ci'], 
                   capsize=5, color=colors, alpha=0.7, edgecolor='black', linewidth=0.5)
    ax1.set_xticks(range(len(plot_df)))
    ax1.set_xticklabels(plot_df['depo'], rotation=45, ha='right', fontsize=8)
    ax1.set_title('Рисунок 2а. Средняя интенсивность изнашивания по топ-20 депо\nс 95% доверительными интервалами', 
                  fontsize=12, fontweight='bold')
    ax1.set_xlabel('Депо', fontsize=10)
    ax1.set_ylabel('Средняя интенсивность изнашивания', fontsize=10)
    ax1.axhline(y=df['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Общее среднее: {df["wear_intensity"].mean():.3f}')
    ax1.legend(fontsize=9)
    ax1.grid(True, alpha=0.3, axis='y')
    
    # 2b. Топ-10 депо с наибольшей интенсивностью
    ax2 = axes[0, 1]
    top10_high = depo_stats.nlargest(10, 'mean')
    
    bars_high = ax2.bar(range(len(top10_high)), top10_high['mean'].values, 
                         yerr=top10_high['std'].values/np.sqrt(top10_high['count'].values)*1.96,
                         capsize=5, color='darkred', alpha=0.7)
    ax2.set_xticks(range(len(top10_high)))
    ax2.set_xticklabels(top10_high.index, rotation=45, ha='right', fontsize=9)
    ax2.set_title('Рисунок 2б. Топ-10 депо с наибольшей интенсивностью изнашивания', 
                  fontsize=12, fontweight='bold')
    ax2.set_xlabel('Депо', fontsize=10)
    ax2.set_ylabel('Средняя интенсивность изнашивания', fontsize=10)
    ax2.axhline(y=df['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Общее среднее')
    ax2.legend(fontsize=9)
    ax2.grid(True, alpha=0.3, axis='y')
    
    # 2c. Топ-10 депо с наименьшей интенсивностью
    ax3 = axes[1, 0]
    top10_low = depo_stats.nsmallest(10, 'mean')
    
    bars_low = ax3.bar(range(len(top10_low)), top10_low['mean'].values, 
                        yerr=top10_low['std'].values/np.sqrt(top10_low['count'].values)*1.96,
                        capsize=5, color='steelblue', alpha=0.7)
    ax3.set_xticks(range(len(top10_low)))
    ax3.set_xticklabels(top10_low.index, rotation=45, ha='right', fontsize=9)
    ax3.set_title('Рисунок 2в. Топ-10 депо с наименьшей интенсивностью изнашивания', 
                  fontsize=12, fontweight='bold')
    ax3.set_xlabel('Депо', fontsize=10)
    ax3.set_ylabel('Средняя интенсивность изнашивания', fontsize=10)
    ax3.axhline(y=df['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Общее среднее')
    ax3.legend(fontsize=9)
    ax3.grid(True, alpha=0.3, axis='y')
    
    # 2d. Размах средних значений
    ax4 = axes[1, 1]
    
    # Группируем депо по диапазонам средней интенсивности
    bins = [0, 0.4, 0.5, 0.6, 0.7, 0.8, 1.0, 1.5, 2.0, 10.0]
    labels = ['<0.4', '0.4-0.5', '0.5-0.6', '0.6-0.7', '0.7-0.8', '0.8-1.0', '1.0-1.5', '1.5-2.0', '>2.0']
    
    depo_stats_copy = depo_stats.copy()
    depo_stats_copy['mean_range'] = pd.cut(depo_stats_copy['mean'], bins=bins, labels=labels)
    range_counts = depo_stats_copy['mean_range'].value_counts().sort_index()
    
    colors_range = plt.cm.RdYlGn_r(np.linspace(0.2, 0.8, len(range_counts)))
    bars_range = ax4.bar(range(len(range_counts)), range_counts.values, color=colors_range, alpha=0.7)
    ax4.set_xticks(range(len(range_counts)))
    ax4.set_xticklabels(range_counts.index, rotation=45, ha='right')
    ax4.set_title('Рисунок 2г. Распределение депо по средней интенсивности изнашивания', 
                  fontsize=12, fontweight='bold')
    ax4.set_xlabel('Диапазон средней интенсивности', fontsize=10)
    ax4.set_ylabel('Количество депо', fontsize=10)
    ax4.grid(True, alpha=0.3, axis='y')
    
    # Добавим подписи
    for i, v in enumerate(range_counts.values):
        ax4.text(i, v + 0.5, str(v), ha='center', va='bottom', fontsize=9)
    
    plt.suptitle('Сравнение средних значений интенсивности изнашивания по депо', 
                 fontsize=16, fontweight='bold', y=1.02)
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)
    
    # ============================================================
    # ГРАФИК 3: Статистический анализ (ANOVA, Kruskal-Wallis)
    # ============================================================
    
    # Подготовка данных для статистических тестов
    df_valid = df[df['depo'].isin(valid_depos)].copy()
    
    # Группируем данные по депо для ANOVA
    # Берем только первые 30 депо для производительности
    test_depos = valid_depos[:30] if len(valid_depos) > 30 else valid_depos
    groups = [df_valid[df_valid['depo'] == depo]['wear_intensity'].values 
              for depo in test_depos]
    
    # Проводим статистические тесты
    if len(groups) >= 2:
        f_stat, anova_p = f_oneway(*groups)
        h_stat, kw_p = kruskal(*groups)
    else:
        f_stat, anova_p = 0, 1.0
        h_stat, kw_p = 0, 1.0
    
    fig, axes = plt.subplots(2, 2, figsize=(16, 12))
    
    # 3a. Box plot для топ-20 депо с подсветкой статистически значимых отличий
    ax1 = axes[0, 0]
    
    top20_valid = valid_depos[:20] if len(valid_depos) > 20 else valid_depos
    df_top20_valid = df[df['depo'].isin(top20_valid)]
    
    # Рассчитываем p-value для каждого депо относительно общего среднего
    depo_pvalues = []
    for depo in top20_valid:
        depo_data = df[df['depo'] == depo]['wear_intensity']
        if len(depo_data) > 1:
            t_stat, p_val = stats.ttest_1samp(depo_data, df['wear_intensity'].mean())
            depo_pvalues.append(p_val)
        else:
            depo_pvalues.append(1.0)
    
    # Сортируем депо по среднему
    depo_means_valid = df_top20_valid.groupby('depo')['wear_intensity'].mean()
    ordered_depos_valid = depo_means_valid.sort_values(ascending=False).index
    
    # Создаем палитру на основе p-value
    palette_colors = ['red' if p < 0.05 else 'steelblue' for p in depo_pvalues]
    
    # Убедимся, что количество цветов соответствует количеству депо
    if len(palette_colors) > len(ordered_depos_valid):
        palette_colors = palette_colors[:len(ordered_depos_valid)]
    elif len(palette_colors) < len(ordered_depos_valid):
        palette_colors.extend(['steelblue'] * (len(ordered_depos_valid) - len(palette_colors)))
    
    sns.boxplot(data=df_top20_valid, x='depo', y='wear_intensity', 
                order=ordered_depos_valid, ax=ax1, palette=palette_colors)
    ax1.set_title('Рисунок 3а. Топ-20 депо (красные - значимо отличаются от общего среднего, p<0.05)', 
                  fontsize=12, fontweight='bold')
    ax1.set_xlabel('Депо', fontsize=10)
    ax1.set_ylabel('Интенсивность изнашивания', fontsize=10)
    ax1.tick_params(axis='x', rotation=45, labelsize=7)  # Исправлено: fontsize -> labelsize
    ax1.axhline(y=df['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Общее среднее')
    ax1.legend(fontsize=9)
    ax1.grid(True, alpha=0.3)
    
    # 3b. Tukey HSD plot для топ-10 депо
    ax2 = axes[0, 1]
    
    top10_valid = valid_depos[:10] if len(valid_depos) > 10 else valid_depos
    df_top10_valid = df[df['depo'].isin(top10_valid)]
    
    if len(df_top10_valid['depo'].unique()) >= 2:
        tukey_results = pairwise_tukeyhsd(df_top10_valid['wear_intensity'], 
                                           df_top10_valid['depo'], 
                                           alpha=0.05)
        
        tukey_results.plot_simultaneous(ax=ax2, figsize=(10, 6))
        ax2.set_title('Рисунок 3б. Доверительные интервалы разности средних (Tukey HSD)\nдля топ-10 депо', 
                      fontsize=12, fontweight='bold')
        ax2.set_xlabel('Разность средних', fontsize=10)
        ax2.axvline(x=0, color='red', linestyle='--', linewidth=1.5, alpha=0.7)
        ax2.grid(True, alpha=0.3)
    
    # 3c. Тепловая карта p-values для попарных сравнений
    ax3 = axes[1, 0]
    
    # Создаем матрицу p-values для топ-10 депо
    depo_list = top10_valid
    p_matrix = pd.DataFrame(index=depo_list, columns=depo_list, data=1.0)
    
    for i, d1 in enumerate(depo_list):
        for j, d2 in enumerate(depo_list):
            if i < j:
                data1 = df[df['depo'] == d1]['wear_intensity']
                data2 = df[df['depo'] == d2]['wear_intensity']
                if len(data1) > 1 and len(data2) > 1:
                    t_stat, p_val = stats.ttest_ind(data1, data2, equal_var=False)
                    p_matrix.loc[d1, d2] = p_val
                    p_matrix.loc[d2, d1] = p_val
    
    # Сокращаем названия депо для читаемости
    short_names = {d: d.replace('ТЧЭ-', '').replace('ТЧ-', '')[:10] for d in depo_list}
    p_matrix_short = p_matrix.rename(index=short_names, columns=short_names)
    
    mask = p_matrix_short > 0.05
    sns.heatmap(-np.log10(p_matrix_short + 1e-10), mask=mask, annot=True, 
                fmt='.1f', cmap='RdYlGn_r', ax=ax3, square=True,
                cbar_kws={'label': '-log10(p-value)'},
                annot_kws={'size': 8})
    ax3.set_title('Рисунок 3в. Матрица значимости различий между депо\n(значения >1.3 соответствуют p<0.05)', 
                  fontsize=12, fontweight='bold')
    ax3.set_xlabel('Депо', fontsize=9)
    ax3.set_ylabel('Депо', fontsize=9)
    
    # 3d. Статистика тестов
    ax4 = axes[1, 1]
    ax4.axis('off')
    
    # Расчет eta-squared (размер эффекта)
    eta_squared = None
    if len(groups) >= 2:
        ss_between = sum([len(g) * (np.mean(g) - df_valid['wear_intensity'].mean())**2 for g in groups])
        ss_total = sum((df_valid['wear_intensity'] - df_valid['wear_intensity'].mean())**2)
        eta_squared = ss_between / ss_total if ss_total > 0 else 0
    
    stats_text = f"""
    ===============================================
    РЕЗУЛЬТАТЫ СТАТИСТИЧЕСКИХ ТЕСТОВ
    ===============================================
    
    ДАННЫЕ:
    • Всего депо: {df['depo'].nunique()}
    • Депо с ≥{min_observations} наблюдениями: {len(valid_depos)}
    • Всего наблюдений в анализе: {len(df_valid)}
    
    ===============================================
    ДИСПЕРСИОННЫЙ АНАЛИЗ (ANOVA):
    
    H₀: Средняя интенсивность изнашивания 
        одинакова во всех депо
    H₁: Есть различия между депо
    
    • F-статистика: {f_stat:.4f}
    • p-значение: {anova_p:.4e}
    
    {'✅ H₀ ОТВЕРГАЕТСЯ' if anova_p < 0.05 else '❌ H₀ ПРИНИМАЕТСЯ'}
    {'(есть статистически значимые различия)' if anova_p < 0.05 else '(нет статистически значимых различий)'}
    
    ===============================================
    НЕПАРАМЕТРИЧЕСКИЙ ТЕСТ (Краскел-Уоллис):
    
    • H-статистика: {h_stat:.4f}
    • p-значение: {kw_p:.4e}
    
    {'✅ Есть различия' if kw_p < 0.05 else '❌ Нет различий'}
    
    ===============================================
    РАЗМЕР ЭФФЕКТА:
    
    • Eta-squared (η²): {eta_squared:.6f if eta_squared else 'N/A'}
    • Доля дисперсии, объясняемая депо: 
      {eta_squared*100:.2f}% дисперсии
    
    ===============================================
    ДЕПО С ЭКСТРЕМАЛЬНЫМИ ЗНАЧЕНИЯМИ:
    """
    
    # Добавляем информацию о крайних депо
    top3_high = depo_stats.nlargest(3, 'mean').index.tolist()
    top3_low = depo_stats.nsmallest(3, 'mean').index.tolist()
    
    for i, d in enumerate(top3_high):
        stats_text += f"\n    Топ-{i+1} по износу: {d}: {depo_stats.loc[d, 'mean']:.3f}"
    
    for i, d in enumerate(top3_low):
        stats_text += f"\n    Топ-{i+1} по минимуму: {d}: {depo_stats.loc[d, 'mean']:.3f}"
    
    ax4.text(0.05, 0.95, stats_text, transform=ax4.transAxes, fontsize=9,
             verticalalignment='top', fontfamily='monospace',
             bbox=dict(boxstyle='round', facecolor='lightyellow', alpha=0.9))
    
    plt.suptitle('Статистический анализ различий между депо', 
                 fontsize=16, fontweight='bold', y=1.02)
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)
    
    # ============================================================
    # ГРАФИК 4: Итоговый анализ и выводы
    # ============================================================
    fig, axes = plt.subplots(2, 2, figsize=(16, 12))
    
    # 4a. Ранжирование депо по средней интенсивности
    ax1 = axes[0, 0]
    
    # Берем топ-30 депо для ранжирования
    top30_depos = depo_stats.nlargest(30, 'mean')
    
    colors_rank = plt.cm.RdYlGn_r(np.linspace(0.1, 0.9, len(top30_depos)))
    bars_rank = ax1.barh(range(len(top30_depos)), top30_depos['mean'].values, 
                          color=colors_rank, alpha=0.8)
    ax1.set_yticks(range(len(top30_depos)))
    ax1.set_yticklabels(top30_depos.index, fontsize=7)
    ax1.set_title('Рисунок 4а. Ранжирование топ-30 депо по средней интенсивности изнашивания', 
                  fontsize=12, fontweight='bold')
    ax1.set_xlabel('Средняя интенсивность изнашивания', fontsize=10)
    ax1.set_ylabel('Депо', fontsize=10)
    ax1.axvline(x=df['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Общее среднее')
    ax1.legend(fontsize=9)
    ax1.grid(True, alpha=0.3, axis='x')
    
    # Добавим подписи значений
    for i, (_, row) in enumerate(top30_depos.iterrows()):
        ax1.text(row['mean'] + 0.01, i, f'{row["mean"]:.3f}', 
                 va='center', fontsize=7)
    
    # 4b. Диаграмма размаха (разброс средних по депо)
    ax2 = axes[0, 1]
    
    # Группируем депо по географическому признаку (первые буквы)
    df['depo_region'] = df['depo'].str.extract(r'^([А-Я]+)')[0]
    
    region_stats = df.groupby('depo_region')['wear_intensity'].agg(['mean', 'std', 'count']).round(4)
    region_stats = region_stats[region_stats['count'] >= 10].sort_values('mean', ascending=False)
    
    if len(region_stats) > 0:
        colors_region = plt.cm.Set3(np.linspace(0, 1, len(region_stats)))
        bars_region = ax2.bar(range(len(region_stats)), region_stats['mean'].values, 
                               yerr=region_stats['std'].values/np.sqrt(region_stats['count'].values)*1.96,
                               capsize=5, color=colors_region, alpha=0.7)
        ax2.set_xticks(range(len(region_stats)))
        ax2.set_xticklabels(region_stats.index, rotation=45, ha='right')
        ax2.set_title('Рисунок 4б. Средняя интенсивность по регионам депо', 
                      fontsize=12, fontweight='bold')
        ax2.set_xlabel('Регион', fontsize=10)
        ax2.set_ylabel('Средняя интенсивность изнашивания', fontsize=10)
        ax2.axhline(y=df['wear_intensity'].mean(), color='red', linestyle='--', 
                    linewidth=2, label=f'Общее среднее')
        ax2.legend(fontsize=9)
        ax2.grid(True, alpha=0.3, axis='y')
    
    # 4c. Финальный box plot с выделением значимых депо
    ax3 = axes[1, 0]
    
    # Выбираем 5 депо с наибольшими и 5 с наименьшими средними
    extreme_depos = list(depo_stats.nlargest(5, 'mean').index) + list(depo_stats.nsmallest(5, 'mean').index)
    df_extreme = df[df['depo'].isin(extreme_depos)]
    
    # Сортируем по среднему
    extreme_means = df_extreme.groupby('depo')['wear_intensity'].mean()
    ordered_extreme = extreme_means.sort_values(ascending=False).index
    
    # Создаем два цвета для двух групп
    colors_extreme = ['darkred']*5 + ['darkblue']*5
    
    sns.boxplot(data=df_extreme, x='depo', y='wear_intensity', 
                order=ordered_extreme, ax=ax3, palette=colors_extreme)
    ax3.set_title('Рисунок 4в. Контрастные группы: депо с max и min износом', 
                  fontsize=12, fontweight='bold')
    ax3.set_xlabel('Депо', fontsize=10)
    ax3.set_ylabel('Интенсивность изнашивания', fontsize=10)
    ax3.tick_params(axis='x', rotation=45, labelsize=9)  # Исправлено: fontsize -> labelsize
    ax3.axhline(y=df['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Общее среднее')
    ax3.legend(fontsize=9)
    ax3.grid(True, alpha=0.3)
    
    # 4d. Итоговое заключение
    ax4 = axes[1, 1]
    ax4.axis('off')
    
    # Формулируем вывод
    if anova_p < 0.05:
        conclusion = "ГИПОТЕЗА ПОДТВЕРЖДЕНА"
        conclusion_color = "green"
        conclusion_text = "Депо приписки статистически значимо влияет на интенсивность изнашивания."
    else:
        conclusion = "ГИПОТЕЗА НЕ ПОДТВЕРЖДЕНА"
        conclusion_color = "red"
        conclusion_text = "Статистически значимого влияния депо приписки на интенсивность изнашивания не обнаружено."
    
    # Оцениваем практическую значимость
    if eta_squared:
        if eta_squared < 0.01:
            practical = "Очень слабый эффект"
        elif eta_squared < 0.06:
            practical = "Слабый эффект"
        elif eta_squared < 0.14:
            practical = "Средний эффект"
        else:
            practical = "Сильный эффект"
    else:
        practical = "Не оценивался"
    
    final_text = f"""
    ╔══════════════════════════════════════════════════════════╗
    ║                     ИТОГОВОЕ ЗАКЛЮЧЕНИЕ                  ║
    ╚══════════════════════════════════════════════════════════╝
    
    {conclusion}
    {'='*50}
    {conclusion_text}
    
    ===============================================
    КЛЮЧЕВЫЕ ПОКАЗАТЕЛИ:
    ===============================================
    
    • p-value (ANOVA): {anova_p:.4e}
    • p-value (Kruskal-Wallis): {kw_p:.4e}
    • Размер эффекта (η²): {eta_squared:.4f if eta_squared else 'N/A'}
    • Практическая значимость: {practical}
    
    ===============================================
    ДИАПАЗОН ЗНАЧЕНИЙ:
    ===============================================
    
    • Минимальная средняя: {depo_stats['mean'].min():.3f}
    • Максимальная средняя: {depo_stats['mean'].max():.3f}
    • Размах: {depo_stats['mean'].max() - depo_stats['mean'].min():.3f}
    • Отношение max/min: {(depo_stats['mean'].max() / depo_stats['mean'].min()):.2f}
    
    ===============================================
    ВЫВОД:
    ===============================================
    
    {'✅ Депо приписки является важным фактором,' if anova_p < 0.05 and eta_squared and eta_squared > 0.01 else '⚠️ Влияние депо статистически значимо, но слабое,' if anova_p < 0.05 else '❌ Депо не является определяющим фактором'}
    {' объясняющим ' + f'{eta_squared*100:.1f}%' if eta_squared else ''} дисперсии.
    
    Рекомендуется {'учитывать депо при планировании ремонтов' if anova_p < 0.05 and eta_squared and eta_squared > 0.01 else 'провести дополнительный анализ с учетом других факторов'}.
    """
    
    ax4.text(0.05, 0.95, final_text, transform=ax4.transAxes, fontsize=10,
             verticalalignment='top', fontfamily='monospace',
             bbox=dict(boxstyle='round', facecolor='lightgreen' if anova_p < 0.05 else 'lightcoral', alpha=0.9))
    
    plt.suptitle('ИТОГОВЫЙ ВЫВОД: Влияние депо приписки на интенсивность изнашивания', 
                 fontsize=16, fontweight='bold', y=1.02)
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)

print("\n" + "="*60)
print("PDF файл 'depo_analysis_results.pdf' успешно создан!")
print("В файле 4 страницы с графиками для анализа влияния депо приписки")
print("="*60)

# Финальные выводы в консоли
print("\n" + "="*60)
print("ИТОГОВЫЕ РЕЗУЛЬТАТЫ ПО ГИПОТЕЗЕ")
print("="*60)

print(f"""
СТАТИСТИЧЕСКИЕ ТЕСТЫ:
------------------------
ANOVA: F={f_stat:.4f}, p-value={anova_p:.4e}
Kruskal-Wallis: H={h_stat:.4f}, p-value={kw_p:.4e}
""")

if eta_squared:
    print(f"Размер эффекта (η²): {eta_squared:.4f} ({eta_squared*100:.2f}% дисперсии)")

print("\n" + "="*60)
print("ВЫВОД ПО ГИПОТЕЗЕ:")
print("="*60)

if anova_p < 0.05:
    if eta_squared and eta_squared > 0.01:
        print("✅ ГИПОТЕЗА ПОДТВЕРЖДЕНА")
        print("   Депо приписки статистически значимо влияет на интенсивность изнашивания.")
        print(f"   Различия между депо объясняют {eta_squared*100:.2f}% вариации в износе.")
        print(f"   Размах средних значений: от {depo_stats['mean'].min():.3f} до {depo_stats['mean'].max():.3f}")
    else:
        print("⚠️ ГИПОТЕЗА ЧАСТИЧНО ПОДТВЕРЖДЕНА")
        print("   Влияние депо статистически значимо, но практическая значимость низкая.")
        print(f"   Доля объясненной дисперсии: всего {eta_squared*100:.2f}%")
else:
    print("❌ ГИПОТЕЗА ОТВЕРГНУТА")
    print("   Депо приписки НЕ оказывает статистически значимого влияния")
    print("   на интенсивность изнашивания.")

print("\n" + "="*60)
print("РЕКОМЕНДАЦИИ:")
print("="*60)
print("""
1. Для планирования ремонтов рекомендуется учитывать:
   - Серию локомотива (основной фактор)
   - Индивидуальные особенности эксплуатации
   - Качество стали

2. Депо может использоваться как вспомогательный фактор,
   особенно для депо с экстремальными значениями износа.

3. Рекомендуется провести дополнительный анализ:
   - Взаимодействие депо и серии локомотива
   - Сезонные и географические факторы
   - Условия эксплуатации в разных депо
""")
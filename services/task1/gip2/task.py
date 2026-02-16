import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
from scipy import stats
from scipy.stats import f_oneway, pearsonr, spearmanr
from matplotlib.backends.backend_pdf import PdfPages
from statsmodels.stats.multicomp import pairwise_tukeyhsd
from sklearn.preprocessing import LabelEncoder
import warnings
warnings.filterwarnings('ignore')

# Загружаем данные из файла
df = pd.read_csv('../wear_data_train.csv')

# Очистка данных: удаляем строки с пропущенными значениями steel_num
df_clean = df.dropna(subset=['steel_num']).copy()
print(f"Всего записей: {len(df)}")
print(f"Записей с номером плавки: {len(df_clean)}")
print(f"Пропущено значений: {len(df) - len(df_clean)}")

# Основные статистики по steel_num
print("\n" + "="*60)
print("ОСНОВНЫЕ СТАТИСТИКИ ПО НОМЕРУ ПЛАВКИ")
print("="*60)
print(f"Количество уникальных номеров плавок: {df_clean['steel_num'].nunique()}")
print(f"Диапазон значений: от {df_clean['steel_num'].min():.0f} до {df_clean['steel_num'].max():.0f}")

# Статистика по интенсивности изнашивания для каждой плавки
steel_stats = df_clean.groupby('steel_num')['wear_intensity'].agg(['count', 'mean', 'std', 'median']).round(4)
steel_stats = steel_stats.sort_values('count', ascending=False)
print("\nТоп-10 плавок по количеству наблюдений:")
print(steel_stats.head(10))

print("\nТоп-10 плавок с наибольшей средней интенсивностью изнашивания:")
print(steel_stats.sort_values('mean', ascending=False).head(10))

print("\nТоп-10 плавок с наименьшей средней интенсивностью изнашивания:")
print(steel_stats.sort_values('mean', ascending=True).head(10))

# Создаем PDF файл с графиками
with PdfPages('steel_analysis_results.pdf') as pdf:
    
    plt.style.use('seaborn-v0_8-darkgrid')
    
    # ============================================================
    # ГРАФИК 1: Распределение интенсивности изнашивания по плавкам
    # ============================================================
    fig, axes = plt.subplots(2, 2, figsize=(16, 12))
    
    # 1a. Гистограмма распределения интенсивности по плавкам
    ax1 = axes[0, 0]
    ax1.hist(df_clean['wear_intensity'], bins=50, edgecolor='black', alpha=0.7, color='steelblue')
    ax1.set_title('Рисунок 1а. Распределение интенсивности изнашивания\n(все наблюдения)', 
                  fontsize=12, fontweight='bold')
    ax1.set_xlabel('Интенсивность изнашивания', fontsize=10)
    ax1.set_ylabel('Частота', fontsize=10)
    ax1.axvline(df_clean['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Среднее: {df_clean["wear_intensity"].mean():.3f}')
    ax1.axvline(df_clean['wear_intensity'].median(), color='green', linestyle='--', 
                linewidth=2, label=f'Медиана: {df_clean["wear_intensity"].median():.3f}')
    ax1.legend(fontsize=9)
    ax1.grid(True, alpha=0.3)
    
    # 1b. Box plot для топ-20 плавок по количеству наблюдений
    ax2 = axes[0, 1]
    top_steel = steel_stats.nlargest(20, 'count').index
    df_top = df_clean[df_clean['steel_num'].isin(top_steel)]
    
    # Преобразуем steel_num в строки для боксплота
    df_top['steel_num_str'] = df_top['steel_num'].astype(int).astype(str)
    
    # Сортируем по среднему значению
    steel_means = df_top.groupby('steel_num_str')['wear_intensity'].mean()
    ordered_steel = steel_means.sort_values(ascending=False).index
    
    sns.boxplot(data=df_top, x='steel_num_str', y='wear_intensity', 
                order=ordered_steel, ax=ax2, palette='viridis')
    ax2.set_title('Рисунок 1б. Распределение интенсивности изнашивания\nпо топ-20 плавкам (по количеству наблюдений)', 
                  fontsize=12, fontweight='bold')
    ax2.set_xlabel('Номер плавки стали', fontsize=10)
    ax2.set_ylabel('Интенсивность изнашивания', fontsize=10)
    ax2.tick_params(axis='x', rotation=45, labelsize=8)
    ax2.axhline(y=df_clean['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Общее среднее')
    ax2.legend(fontsize=9)
    ax2.grid(True, alpha=0.3)
    
    # 1c. Scatter plot: номер плавки vs интенсивность
    ax3 = axes[1, 0]
    # Берем случайную выборку для читаемости (10% данных)
    sample_df = df_clean.sample(frac=0.1, random_state=42)
    ax3.scatter(sample_df['steel_num'], sample_df['wear_intensity'], 
                alpha=0.5, s=10, c='steelblue')
    ax3.set_title('Рисунок 1в. Зависимость интенсивности изнашивания от номера плавки\n(случайная выборка 10% данных)', 
                  fontsize=12, fontweight='bold')
    ax3.set_xlabel('Номер плавки стали', fontsize=10)
    ax3.set_ylabel('Интенсивность изнашивания', fontsize=10)
    ax3.grid(True, alpha=0.3)
    
    # 1d. Violin plot для топ-15 плавок
    ax4 = axes[1, 1]
    top15_steel = steel_stats.nlargest(15, 'count').index
    df_top15 = df_clean[df_clean['steel_num'].isin(top15_steel)]
    df_top15['steel_num_str'] = df_top15['steel_num'].astype(int).astype(str)
    
    steel_means15 = df_top15.groupby('steel_num_str')['wear_intensity'].mean()
    ordered_steel15 = steel_means15.sort_values(ascending=False).index
    
    sns.violinplot(data=df_top15, x='steel_num_str', y='wear_intensity', 
                   order=ordered_steel15, ax=ax4, palette='mako')
    ax4.set_title('Рисунок 1г. Violin plot для топ-15 плавок\n(по количеству наблюдений)', 
                  fontsize=12, fontweight='bold')
    ax4.set_xlabel('Номер плавки стали', fontsize=10)
    ax4.set_ylabel('Интенсивность изнашивания', fontsize=10)
    ax4.tick_params(axis='x', rotation=45, labelsize=8)
    ax4.axhline(y=df_clean['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Общее среднее')
    ax4.legend(fontsize=9)
    ax4.grid(True, alpha=0.3)
    
    plt.suptitle('Анализ влияния номера плавки стали на интенсивность изнашивания', 
                 fontsize=16, fontweight='bold', y=1.02)
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)
    
    # ============================================================
    # ГРАФИК 2: Корреляционный анализ
    # ============================================================
    fig, axes = plt.subplots(2, 2, figsize=(16, 12))
    
    # Рассчитываем корреляции
    pearson_corr, pearson_p = pearsonr(df_clean['steel_num'], df_clean['wear_intensity'])
    spearman_corr, spearman_p = spearmanr(df_clean['steel_num'], df_clean['wear_intensity'])
    
    # 2a. Scatter plot с линией регрессии
    ax1 = axes[0, 0]
    # Берем случайную выборку для читаемости
    sample_df2 = df_clean.sample(frac=0.05, random_state=42)
    ax1.scatter(sample_df2['steel_num'], sample_df2['wear_intensity'], 
                alpha=0.5, s=5, c='steelblue')
    
    # Добавляем линию регрессии
    z = np.polyfit(df_clean['steel_num'], df_clean['wear_intensity'], 1)
    p = np.poly1d(z)
    ax1.plot(df_clean['steel_num'].sort_values(), 
             p(df_clean['steel_num'].sort_values()), 
             'r-', linewidth=2, label=f'Линия регрессии (коэф.={z[0]:.6f})')
    
    ax1.set_title('Рисунок 2а. Scatter plot с линией регрессии', 
                  fontsize=12, fontweight='bold')
    ax1.set_xlabel('Номер плавки стали', fontsize=10)
    ax1.set_ylabel('Интенсивность изнашивания', fontsize=10)
    ax1.legend(fontsize=9)
    ax1.grid(True, alpha=0.3)
    
    # 2b. Гистограмма распределения интенсивности для разных диапазонов плавок
    ax2 = axes[0, 1]
    
    # Разбиваем плавки на квантильные группы
    df_clean['steel_group'] = pd.qcut(df_clean['steel_num'], q=5, labels=['Q1 (мин)', 'Q2', 'Q3', 'Q4', 'Q5 (макс)'])
    
    for i, group in enumerate(['Q1 (мин)', 'Q2', 'Q3', 'Q4', 'Q5 (макс)']):
        group_data = df_clean[df_clean['steel_group'] == group]['wear_intensity']
        ax2.hist(group_data, bins=30, alpha=0.5, label=f'{group} (n={len(group_data)})', density=True)
    
    ax2.set_title('Рисунок 2б. Распределение интенсивности по квантильным группам плавок', 
                  fontsize=12, fontweight='bold')
    ax2.set_xlabel('Интенсивность изнашивания', fontsize=10)
    ax2.set_ylabel('Плотность вероятности', fontsize=10)
    ax2.legend(fontsize=8)
    ax2.grid(True, alpha=0.3)
    
    # 2c. Box plot по квантильным группам
    ax3 = axes[1, 0]
    sns.boxplot(data=df_clean, x='steel_group', y='wear_intensity', ax=ax3, palette='Set2')
    ax3.set_title('Рисунок 2в. Сравнение интенсивности по квантильным группам плавок', 
                  fontsize=12, fontweight='bold')
    ax3.set_xlabel('Группа плавок (по квантилям)', fontsize=10)
    ax3.set_ylabel('Интенсивность изнашивания', fontsize=10)
    ax3.axhline(y=df_clean['wear_intensity'].mean(), color='red', linestyle='--', 
                linewidth=2, label=f'Общее среднее')
    ax3.legend(fontsize=9)
    ax3.grid(True, alpha=0.3)
    
    # 2d. Тепловая карта корреляций (для разных серий)
    ax4 = axes[1, 1]
    
    # Рассчитываем корреляции для топ-10 серий
    top_series = df_clean['locomotive_series'].value_counts().nlargest(10).index
    corr_data = []
    
    for series in top_series:
        series_df = df_clean[df_clean['locomotive_series'] == series]
        if len(series_df) > 5:
            pearson_c, pearson_p = pearsonr(series_df['steel_num'], series_df['wear_intensity'])
            corr_data.append({
                'series': series,
                'correlation': pearson_c,
                'p_value': pearson_p,
                'count': len(series_df)
            })
    
    corr_df = pd.DataFrame(corr_data).sort_values('correlation', ascending=False)
    
    if len(corr_df) > 0:
        colors = ['red' if abs(c) > 0.1 else 'steelblue' for c in corr_df['correlation']]
        bars = ax4.bar(range(len(corr_df)), corr_df['correlation'], color=colors, alpha=0.7)
        ax4.set_xticks(range(len(corr_df)))
        ax4.set_xticklabels(corr_df['series'], rotation=45, ha='right', fontsize=8)
        ax4.set_title('Рисунок 2г. Корреляция Пирсона между номером плавки и износом\nпо отдельным сериям локомотивов', 
                      fontsize=12, fontweight='bold')
        ax4.set_xlabel('Серия локомотива', fontsize=10)
        ax4.set_ylabel('Коэффициент корреляции', fontsize=10)
        ax4.axhline(y=0, color='black', linestyle='-', linewidth=1)
        ax4.axhline(y=0.1, color='red', linestyle='--', linewidth=1, alpha=0.5)
        ax4.axhline(y=-0.1, color='red', linestyle='--', linewidth=1, alpha=0.5)
        ax4.grid(True, alpha=0.3, axis='y')
        
        # Добавим подписи значений
        for i, (_, row) in enumerate(corr_df.iterrows()):
            ax4.text(i, row['correlation'] + (0.02 if row['correlation'] >= 0 else -0.05), 
                     f'{row["correlation"]:.3f}', ha='center', va='bottom' if row['correlation'] >= 0 else 'top', 
                     fontsize=8, fontweight='bold')
    
    plt.suptitle(f'Корреляционный анализ\nПирсон: r={pearson_corr:.4f} (p={pearson_p:.4e}), Спирмен: ρ={spearman_corr:.4f} (p={spearman_p:.4e})', 
                 fontsize=14, fontweight='bold', y=1.02)
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)
    
    # ============================================================
    # ГРАФИК 3: ANOVA анализ для плавок с достаточным количеством наблюдений
    # ============================================================
    
    # Выбираем плавки с минимум 10 наблюдениями
    valid_steel = steel_stats[steel_stats['count'] >= 10].index
    df_anova_steel = df_clean[df_clean['steel_num'].isin(valid_steel)]
    
    print(f"\nКоличество плавок с >=10 наблюдениями: {len(valid_steel)}")
    print(f"Всего наблюдений в этих плавках: {len(df_anova_steel)}")
    
    fig, axes = plt.subplots(2, 2, figsize=(16, 12))
    
    # Подготовка данных для ANOVA
    groups = [df_anova_steel[df_anova_steel['steel_num'] == steel]['wear_intensity'].values 
              for steel in valid_steel[:20]]  # Ограничим первыми 20 для анализа
    
    if len(groups) >= 2:
        f_stat, p_value = f_oneway(*groups)
        h_stat, kw_p_value = stats.kruskal(*groups)
        
        # 3a. Bar plot средних значений для топ-20 плавок
        ax1 = axes[0, 0]
        top20_steel = steel_stats.nlargest(20, 'count').index
        df_top20 = df_clean[df_clean['steel_num'].isin(top20_steel)]
        
        means = df_top20.groupby('steel_num')['wear_intensity'].mean()
        stds = df_top20.groupby('steel_num')['wear_intensity'].std()
        counts = df_top20.groupby('steel_num')['wear_intensity'].count()
        
        x_pos = np.arange(len(means))
        colors = ['coral' if m > df_clean['wear_intensity'].mean() else 'steelblue' for m in means]
        
        ax1.bar(x_pos, means, yerr=stds/np.sqrt(counts)*1.96, capsize=5, 
                color=colors, alpha=0.7, edgecolor='black', linewidth=0.5)
        ax1.set_xticks(x_pos)
        ax1.set_xticklabels([f'{int(s)}' for s in means.index], rotation=45, ha='right', fontsize=8)
        ax1.set_title('Рисунок 3а. Средняя интенсивность изнашивания по топ-20 плавкам\n(с 95% доверительными интервалами)', 
                      fontsize=12, fontweight='bold')
        ax1.set_xlabel('Номер плавки стали', fontsize=10)
        ax1.set_ylabel('Средняя интенсивность изнашивания', fontsize=10)
        ax1.axhline(y=df_clean['wear_intensity'].mean(), color='red', linestyle='--', 
                    linewidth=2, label=f'Общее среднее')
        ax1.legend(fontsize=9)
        ax1.grid(True, alpha=0.3, axis='y')
        
        # 3b. Tukey HSD plot для топ-10 плавок
        ax2 = axes[0, 1]
        top10_steel = steel_stats.nlargest(10, 'count').index
        df_top10 = df_clean[df_clean['steel_num'].isin(top10_steel)]
        
        # Преобразуем steel_num в строки для Tukey
        df_top10['steel_num_str'] = df_top10['steel_num'].astype(int).astype(str)
        
        if len(df_top10['steel_num_str'].unique()) >= 2:
            tukey_results = pairwise_tukeyhsd(df_top10['wear_intensity'], 
                                               df_top10['steel_num_str'], 
                                               alpha=0.05)
            
            tukey_results.plot_simultaneous(ax=ax2, figsize=(10, 6))
            ax2.set_title('Рисунок 3б. Доверительные интервалы разности средних (Tukey HSD)\nдля топ-10 плавок', 
                          fontsize=12, fontweight='bold')
            ax2.set_xlabel('Разность средних', fontsize=10)
            ax2.axvline(x=0, color='red', linestyle='--', linewidth=1.5, alpha=0.7)
            ax2.grid(True, alpha=0.3)
        
        # 3c. Распределение интенсивности по плавкам (violin plot)
        ax3 = axes[1, 0]
        sns.violinplot(data=df_top10, x='steel_num_str', y='wear_intensity', 
                       order=df_top10.groupby('steel_num_str')['wear_intensity'].mean().sort_values(ascending=False).index,
                       ax=ax3, palette='mako')
        ax3.set_title('Рисунок 3в. Распределение интенсивности по топ-10 плавкам', 
                      fontsize=12, fontweight='bold')
        ax3.set_xlabel('Номер плавки стали', fontsize=10)
        ax3.set_ylabel('Интенсивность изнашивания', fontsize=10)
        ax3.tick_params(axis='x', rotation=45, fontsize=8)
        ax3.axhline(y=df_clean['wear_intensity'].mean(), color='red', linestyle='--', 
                    linewidth=2, label=f'Общее среднее')
        ax3.legend(fontsize=9)
        ax3.grid(True, alpha=0.3)
        
        # 3d. Статистика тестов
        ax4 = axes[1, 1]
        ax4.axis('off')
        
        stats_text = f"""
        РЕЗУЛЬТАТЫ СТАТИСТИЧЕСКИХ ТЕСТОВ
        
        ANOVA (для плавок с ≥10 наблюдениями):
        • F-статистика: {f_stat:.4f}
        • p-значение: {p_value:.4e}
        
        Краскел-Уоллис (непараметрический):
        • H-статистика: {h_stat:.4f}
        • p-значение: {kw_p_value:.4e}
        
        Корреляционный анализ (все данные):
        • Пирсон: r = {pearson_corr:.4f} (p={pearson_p:.4e})
        • Спирмен: ρ = {spearman_corr:.4f} (p={spearman_p:.4e})
        
        {'✅ СТАТИСТИЧЕСКИ ЗНАЧИМЫЕ РАЗЛИЧИЯ ОБНАРУЖЕНЫ' if p_value < 0.05 or kw_p_value < 0.05 else '❌ СТАТИСТИЧЕСКИ ЗНАЧИМЫХ РАЗЛИЧИЙ НЕ ОБНАРУЖЕНО'}
        
        Интерпретация:
        • Корреляция {'слабая' if abs(pearson_corr) < 0.1 else 'умеренная' if abs(pearson_corr) < 0.3 else 'сильная'}
        • {'Зависимость есть' if p_value < 0.05 else 'Зависимости нет'} между плавками
        """
        
        ax4.text(0.1, 0.9, stats_text, transform=ax4.transAxes, fontsize=11,
                 verticalalignment='top', fontfamily='monospace',
                 bbox=dict(boxstyle='round', facecolor='lightyellow', alpha=0.8))
        
    plt.suptitle('ANOVA анализ влияния номера плавки на интенсивность изнашивания', 
                 fontsize=16, fontweight='bold', y=1.02)
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)
    
    # ============================================================
    # ГРАФИК 4: Сравнение плавок с контролем по серии локомотива
    # ============================================================
    fig, axes = plt.subplots(2, 2, figsize=(16, 12))
    
    # Выбираем топ-3 серии по количеству наблюдений
    top3_series = df_clean['locomotive_series'].value_counts().nlargest(3).index
    
    for idx, series in enumerate(top3_series):
        if idx < 2:
            ax = axes[0, idx]
        else:
            ax = axes[1, idx-2]
        
        series_df = df_clean[df_clean['locomotive_series'] == series].copy()
        
        # Выбираем плавки с минимум 5 наблюдениями в этой серии
        steel_counts = series_df['steel_num'].value_counts()
        valid_steel_series = steel_counts[steel_counts >= 5].index
        series_df_filtered = series_df[series_df['steel_num'].isin(valid_steel_series)]
        
        if len(series_df_filtered) > 0 and len(series_df_filtered['steel_num'].unique()) >= 2:
            series_df_filtered['steel_num_str'] = series_df_filtered['steel_num'].astype(int).astype(str)
            
            # Сортируем по среднему
            steel_means_series = series_df_filtered.groupby('steel_num_str')['wear_intensity'].mean()
            ordered_steel_series = steel_means_series.sort_values(ascending=False).index
            
            sns.boxplot(data=series_df_filtered, x='steel_num_str', y='wear_intensity', 
                        order=ordered_steel_series, ax=ax, palette='Set3')
            ax.set_title(f'Серия: {series}\n(плавки с ≥5 наблюдениями)', 
                         fontsize=12, fontweight='bold')
            ax.set_xlabel('Номер плавки стали', fontsize=9)
            ax.set_ylabel('Интенсивность изнашивания', fontsize=9)
            ax.tick_params(axis='x', rotation=45, fontsize=7)
            ax.axhline(y=series_df['wear_intensity'].mean(), color='red', linestyle='--', 
                       linewidth=1.5, label=f'Среднее по серии: {series_df["wear_intensity"].mean():.3f}')
            ax.legend(fontsize=7, loc='upper right')
            ax.grid(True, alpha=0.3)
            
            # Рассчитаем корреляцию для этой серии
            if len(series_df_filtered) > 5:
                p_corr, p_p = pearsonr(series_df_filtered['steel_num'], series_df_filtered['wear_intensity'])
                s_corr, s_p = spearmanr(series_df_filtered['steel_num'], series_df_filtered['wear_intensity'])
                ax.text(0.02, 0.98, f'Пирсон: r={p_corr:.3f} (p={p_p:.3f})\nСпирмен: ρ={s_corr:.3f} (p={s_p:.3f})', 
                        transform=ax.transAxes, fontsize=7, verticalalignment='top',
                        bbox=dict(boxstyle='round', facecolor='white', alpha=0.8))
    
    # Если осталась пустая ячейка
    if len(top3_series) < 3:
        axes[1, 1].axis('off')
        axes[1, 1].text(0.5, 0.5, 'Недостаточно данных\nдля третьей серии', 
                        ha='center', va='center', fontsize=12, style='italic')
    elif len(top3_series) == 3:
        # Добавим четвертый график - сводная статистика
        ax4 = axes[1, 1]
        
        # Собираем данные по всем сериям
        summary_data = []
        for series in df_clean['locomotive_series'].value_counts().nlargest(5).index:
            series_df = df_clean[df_clean['locomotive_series'] == series]
            if len(series_df) > 10:
                p_corr, p_p = pearsonr(series_df['steel_num'], series_df['wear_intensity'])
                summary_data.append({
                    'series': series,
                    'correlation': p_corr,
                    'p_value': p_p,
                    'count': len(series_df)
                })
        
        summary_df = pd.DataFrame(summary_data).sort_values('correlation', ascending=False)
        
        if len(summary_df) > 0:
            bars = ax4.bar(range(len(summary_df)), summary_df['correlation'], 
                           color=['red' if p < 0.05 else 'steelblue' for p in summary_df['p_value']])
            ax4.set_xticks(range(len(summary_df)))
            ax4.set_xticklabels(summary_df['series'], rotation=45, ha='right', fontsize=9)
            ax4.set_title('Корреляция плавка-износ по основным сериям\n(красный - значимая p<0.05)', 
                          fontsize=12, fontweight='bold')
            ax4.set_ylabel('Коэффициент корреляции Пирсона', fontsize=9)
            ax4.axhline(y=0, color='black', linestyle='-', linewidth=1)
            ax4.grid(True, alpha=0.3, axis='y')
    
    plt.suptitle('Анализ влияния номера плавки в разрезе отдельных серий локомотивов', 
                 fontsize=16, fontweight='bold', y=1.02)
    plt.tight_layout()
    pdf.savefig(fig)
    plt.close(fig)

print("\n" + "="*60)
print("PDF файл 'steel_analysis_results.pdf' успешно создан!")
print("В файле 4 страницы с графиками для анализа влияния номера плавки")
print("="*60)

# Финальные выводы
print("\n" + "="*60)
print("ВЫВОДЫ ПО РЕЗУЛЬТАТАМ АНАЛИЗА")
print("="*60)

# Рассчитываем все необходимые статистики
pearson_corr, pearson_p = pearsonr(df_clean['steel_num'], df_clean['wear_intensity'])
spearman_corr, spearman_p = spearmanr(df_clean['steel_num'], df_clean['wear_intensity'])

# ANOVA для плавок с достаточным количеством наблюдений
valid_steel = steel_stats[steel_stats['count'] >= 10].index[:20]  # Первые 20
groups = [df_clean[df_clean['steel_num'] == steel]['wear_intensity'].values 
          for steel in valid_steel if len(df_clean[df_clean['steel_num'] == steel]) >= 5]

if len(groups) >= 2:
    f_stat, anova_p = f_oneway(*groups)
    h_stat, kw_p = stats.kruskal(*groups)
else:
    anova_p = 1.0
    kw_p = 1.0

print(f"\nКорреляционный анализ:")
print(f"  Пирсон: r = {pearson_corr:.4f}, p-value = {pearson_p:.4e}")
print(f"  Спирмен: ρ = {spearman_corr:.4f}, p-value = {spearman_p:.4e}")

print(f"\nДисперсионный анализ (ANOVA для плавок с ≥10 наблюдениями):")
print(f"  p-value = {anova_p:.4e}")

print(f"\nНепараметрический тест (Краскел-Уоллис):")
print(f"  p-value = {kw_p:.4e}")

print("\n" + "="*60)
print("ИТОГОВОЕ ЗАКЛЮЧЕНИЕ ПО ГИПОТЕЗЕ")
print("="*60)

if anova_p < 0.05 or kw_p < 0.05 or pearson_p < 0.05:
    if abs(pearson_corr) > 0.1:
        print(f"✅ ГИПОТЕЗА ЧАСТИЧНО ПОДТВЕРЖДЕНА: Номер плавки стали оказывает слабое, но статистически значимое влияние")
        print(f"   на интенсивность изнашивания (p={pearson_p:.4e}, r={pearson_corr:.3f}).")
        print()
        print(f"   Однако корреляция очень слабая (|r|={abs(pearson_corr):.3f}), что указывает на то,")
        print(f"   что номер плавки не является основным фактором, определяющим износ.")
    else:
        print(f"❌ ГИПОТЕЗА НЕ ПОДТВЕРЖДЕНА: Хотя статистические тесты показывают значимость,")
        print(f"   практическая значимость отсутствует из-за очень слабой корреляции (|r|={abs(pearson_corr):.3f}).")
else:
    print(f"❌ ГИПОТЕЗА ОТВЕРГНУТА: Номер плавки стали НЕ оказывает статистически значимого влияния")
    print(f"   на интенсивность изнашивания (p={anova_p:.4f} для ANOVA).")

print()
print(f"Ключевые наблюдения:")
print(f"• Коэффициент детерминации R² = {pearson_corr**2:.4f} (менее {pearson_corr**2*100:.1f}% вариации объясняется номером плавки)")
print(f"• Разброс средних значений между плавками: от {steel_stats['mean'].min():.3f} до {steel_stats['mean'].max():.3f}")
print(f"• {'Наблюдается' if pearson_corr > 0 else 'Не наблюдается'} устойчивой тенденции зависимости износа от номера плавки")
// Глобальные переменные
let apiData = null;
let currentView = 'cards';
let currentFilter = 'all';
let searchTerm = '';

// Загрузка данных при старте
document.addEventListener('DOMContentLoaded', () => {
    console.log('Страница загружена, начинаем загрузку данных...');
    fetchData();
    setupEventListeners();
});

// Получение данных с API
async function fetchData() {
    showLoading();
    try {
        console.log('Запрос к API:', '/api/v1/popular-direction');
        const response = await fetch('/api/v1/popular-direction');
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        apiData = await response.json();
        console.log('Данные получены:', apiData);
        
        updateFilterButtons();
        renderContent();
        hideLoading();
    } catch (error) {
        console.error('Error fetching data:', error);
        showError();
    }
}

// Обновление кнопок фильтрации
function updateFilterButtons() {
    if (!apiData || !apiData.depots) return;
    
    const container = document.getElementById('filterButtons');
    if (!container) return;
    
    const depots = apiData.depots.map(d => d.depo_code);
    
    let html = '<button class="filter-btn active" data-filter="all">Все депо</button>';
    depots.forEach(depo => {
        html += `<button class="filter-btn" data-filter="${depo}">Депо ${depo}</button>`;
    });
    
    container.innerHTML = html;
    
    // Добавляем обработчики
    container.querySelectorAll('.filter-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            container.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            currentFilter = btn.dataset.filter;
            renderContent();
        });
    });
}

// Настройка обработчиков событий
function setupEventListeners() {
    // Поиск
    const searchInput = document.getElementById('locomotiveSearch');
    if (searchInput) {
        searchInput.addEventListener('input', (e) => {
            searchTerm = e.target.value.toLowerCase();
            renderContent();
        });
    }
    
    // Виды отображения
    document.querySelectorAll('.view-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            document.querySelectorAll('.view-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            currentView = btn.dataset.view;
            renderContent();
        });
    });
}

// Показать загрузку
function showLoading() {
    const loading = document.getElementById('loading');
    const error = document.getElementById('error');
    const content = document.getElementById('content');
    
    if (loading) loading.style.display = 'block';
    if (error) error.style.display = 'none';
    if (content) content.innerHTML = '';
}

// Скрыть загрузку
function hideLoading() {
    const loading = document.getElementById('loading');
    if (loading) loading.style.display = 'none';
}

// Показать ошибку
function showError() {
    const loading = document.getElementById('loading');
    const error = document.getElementById('error');
    
    if (loading) loading.style.display = 'none';
    if (error) error.style.display = 'block';
}

// Отрисовка контента
function renderContent() {
    const content = document.getElementById('content');
    
    if (!apiData) {
        content.innerHTML = '<div class="loading">Нет данных для отображения</div>';
        return;
    }
    
    let filteredDepots = filterDepots();
    
    switch(currentView) {
        case 'cards':
            content.innerHTML = renderCardsView(filteredDepots);
            break;
        case 'table':
            content.innerHTML = renderTableView(filteredDepots);
            break;
        case 'stats':
            content.innerHTML = renderStatsView();
            break;
        default:
            content.innerHTML = renderCardsView(filteredDepots);
    }
}

// Фильтрация депо
function filterDepots() {
    if (!apiData || !apiData.depots) return [];
    
    let depots = apiData.depots;
    
    if (currentFilter !== 'all') {
        depots = depots.filter(d => d.depo_code === currentFilter);
    }
    
    return depots;
}

// Фильтрация локомотивов
function filterLocomotives(locomotives) {
    if (!searchTerm) return locomotives;
    
    return locomotives.filter(loco => 
        loco.model.toLowerCase().includes(searchTerm) ||
        loco.number.toLowerCase().includes(searchTerm)
    );
}

// Рендер карточек
function renderCardsView(depots) {
    let html = '';
    
    depots.forEach(depot => {
        const filteredLocomotives = filterLocomotives(depot.locomotives);
        
        html += `
            <div class="depot-card">
                <div class="depot-header" onclick="toggleDepot(this)">
                    <h3>
                        <i class="fas fa-warehouse"></i>
                        Депо ${depot.depo_code}
                    </h3>
                    <div class="depot-stats">
                        <div class="depot-stat">
                            <span class="stat-value">${depot.locomotive_count || 0}</span>
                            <span class="stat-label">локомотивов</span>
                        </div>
                        <div class="depot-stat">
                            <span class="stat-value">${depot.directions_count || 0}</span>
                            <span class="stat-label">направлений</span>
                        </div>
                    </div>
                </div>
                
                <div class="directions-badge">
                    ${depot.available_directions ? depot.available_directions.map(dir => `
                        <span class="badge">
                            <i class="fas fa-arrow-right"></i>
                            ${dir.name} (${dir.prefix})
                        </span>
                    `).join('') : ''}
                </div>
                
                <div class="locomotives-table" style="display: none;">
                    <table>
                        <thead>
                            <tr>
                                <th>Модель</th>
                                <th>Номер</th>
                                <th>Поездок</th>
                                <th>Популярное направление</th>
                                <th>Посещений</th>
                                <th>%</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${filteredLocomotives.map(loco => `
                                <tr onclick="showLocomotiveDetails('${loco.model}', '${loco.number}')">
                                    <td>${loco.model || ''}</td>
                                    <td>${loco.number || ''}</td>
                                    <td>${loco.total_trips || 0}</td>
                                    <td>${loco.most_popular?.direction_name || 'Нет данных'}</td>
                                    <td>${loco.most_popular?.visits || 0}</td>
                                    <td>${loco.most_popular?.percentage ? loco.most_popular.percentage.toFixed(1) : 0}%</td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>
            </div>
        `;
    });
    
    return html;
}

// Рендер таблицы
function renderTableView(depots) {
    let html = '<div class="locomotives-table"><table><thead><tr>';
    html += '<th>Депо</th><th>Модель</th><th>Номер</th><th>Поездок</th>';
    html += '<th>Популярное направление</th><th>Посещений</th><th>%</th></tr></thead><tbody>';
    
    depots.forEach(depot => {
        const filteredLocomotives = filterLocomotives(depot.locomotives);
        
        filteredLocomotives.forEach(loco => {
            html += `
                <tr onclick="showLocomotiveDetails('${loco.model}', '${loco.number}')">
                    <td>${depot.depo_code}</td>
                    <td>${loco.model || ''}</td>
                    <td>${loco.number || ''}</td>
                    <td>${loco.total_trips || 0}</td>
                    <td>${loco.most_popular?.direction_name || 'Нет данных'}</td>
                    <td>${loco.most_popular?.visits || 0}</td>
                    <td>${loco.most_popular?.percentage ? loco.most_popular.percentage.toFixed(1) : 0}%</td>
                </tr>
            `;
        });
    });
    
    html += '</tbody></table></div>';
    return html;
}

// Рендер статистики
function renderStatsView() {
    if (!apiData) return '';
    
    setTimeout(() => {
        createDirectionChart();
        createDepotChart();
    }, 100);
    
    return `
        <div class="stats-container">
            <div class="stat-card">
                <h3><i class="fas fa-chart-pie"></i> Топ-10 направлений</h3>
                <canvas id="directionChart"></canvas>
            </div>
            <div class="stat-card">
                <h3><i class="fas fa-chart-bar"></i> Распределение по депо</h3>
                <canvas id="depotChart"></canvas>
            </div>
            <div class="stat-card">
                <h3><i class="fas fa-chart-line"></i> Детальная статистика</h3>
                <div style="padding: 20px;">
                    <p><strong>Всего локомотивов:</strong> ${apiData.overall_stats?.total_locomotives || 0}</p>
                    <p><strong>С любимым направлением:</strong> ${apiData.overall_stats?.locomotives_with_favorite || 0} (${apiData.overall_stats?.locomotives_with_favorite_percent ? apiData.overall_stats.locomotives_with_favorite_percent.toFixed(1) : 0}%)</p>
                    <p><strong>На одном направлении:</strong> ${apiData.overall_stats?.locomotives_single_direction || 0} (${apiData.overall_stats?.locomotives_single_direction_percent ? apiData.overall_stats.locomotives_single_direction_percent.toFixed(1) : 0}%)</p>
                    <p><strong>Всего направлений:</strong> ${apiData.direction_popularity?.length || 0}</p>
                    <p><strong>Всего депо:</strong> ${apiData.depots?.length || 0}</p>
                </div>
            </div>
        </div>
    `;
}

// Создание графика направлений
function createDirectionChart() {
    const ctx = document.getElementById('directionChart');
    if (!ctx || !apiData.direction_popularity) return;
    
    const topDirections = apiData.direction_popularity.slice(0, 10);
    
    new Chart(ctx, {
        type: 'pie',
        data: {
            labels: topDirections.map(d => d.direction_name),
            datasets: [{
                data: topDirections.map(d => d.count),
                backgroundColor: [
                    '#667eea', '#764ba2', '#f39c12', '#e74c3c', '#2ecc71',
                    '#3498db', '#9b59b6', '#1abc9c', '#f1c40f', '#e67e22'
                ]
            }]
        },
        options: {
            responsive: true,
            plugins: {
                legend: {
                    position: 'bottom'
                }
            }
        }
    });
}

// Создание графика депо
function createDepotChart() {
    const ctx = document.getElementById('depotChart');
    if (!ctx || !apiData.depots) return;
    
    new Chart(ctx, {
        type: 'bar',
        data: {
            labels: apiData.depots.map(d => `Депо ${d.depo_code}`),
            datasets: [{
                label: 'Количество локомотивов',
                data: apiData.depots.map(d => d.locomotive_count),
                backgroundColor: '#667eea'
            }]
        },
        options: {
            responsive: true,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });
}

// Переключение отображения депо
function toggleDepot(element) {
    const table = element.parentElement.querySelector('.locomotives-table');
    if (table) {
        table.style.display = table.style.display === 'none' ? 'block' : 'none';
    }
}

// Показать детали локомотива
function showLocomotiveDetails(model, number) {
    if (!apiData || !apiData.depots) return;
    
    for (const depot of apiData.depots) {
        const loco = depot.locomotives.find(l => l.model === model && l.number === number);
        if (loco) {
            showModal(loco, depot.depo_code);
            break;
        }
    }
}

// Показать модальное окно
function showModal(loco, depoCode) {
    const modal = document.getElementById('locomotiveModal');
    const modalBody = document.getElementById('modalBody');
    
    if (!modal || !modalBody) return;
    
    const directionsList = loco.direction_visits ? loco.direction_visits.map(dir => `
        <li>
            <span class="direction-name">${dir.direction_name}</span>
            <span class="direction-stats">${dir.visits} поездок (${dir.percentage.toFixed(1)}%)</span>
        </li>
    `).join('') : '';
    
    modalBody.innerHTML = `
        <div class="loco-detail">
            <div class="detail-row">
                <span class="detail-label">Модель:</span>
                <span class="detail-value">${loco.model}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Номер:</span>
                <span class="detail-value">${loco.number}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Депо:</span>
                <span class="detail-value">${depoCode}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Всего поездок:</span>
                <span class="detail-value">${loco.total_trips}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Посещено направлений:</span>
                <span class="detail-value">${loco.visited_directions ? loco.visited_directions.length : 0}</span>
            </div>
            <h3>Самое популярное направление:</h3>
            ${loco.most_popular ? `
                <div class="detail-row">
                    <span class="detail-label">${loco.most_popular.direction_name}:</span>
                    <span class="detail-value">${loco.most_popular.visits} поездок (${loco.most_popular.percentage.toFixed(1)}%)</span>
                </div>
            ` : '<p>Нет данных</p>'}
            
            <h3>Все посещенные направления:</h3>
            <ul class="directions-list">
                ${directionsList}
            </ul>
        </div>
    `;
    
    modal.classList.add('show');
}

// Закрыть модальное окно
function closeModal() {
    const modal = document.getElementById('locomotiveModal');
    if (modal) {
        modal.classList.remove('show');
    }
}

// Закрытие по клику вне модального окна
window.onclick = function(event) {
    const modal = document.getElementById('locomotiveModal');
    if (event.target === modal) {
        closeModal();
    }
}

// Глобальные функции для HTML
window.toggleDepot = toggleDepot;
window.showLocomotiveDetails = showLocomotiveDetails;
window.closeModal = closeModal;
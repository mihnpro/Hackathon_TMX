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
    const searchInput = document.getElementById('locomotiveSearch');
    if (searchInput) {
        searchInput.addEventListener('input', (e) => {
            searchTerm = e.target.value.toLowerCase();
            renderContent();
        });
    }
    
    document.querySelectorAll('.view-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            document.querySelectorAll('.view-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            currentView = btn.dataset.view;
            renderContent();
        });
    });
    
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            closeModal();
        }
    });
}

function showLoading() {
    const loading = document.getElementById('loading');
    const error = document.getElementById('error');
    const content = document.getElementById('content');
    
    if (loading) loading.style.display = 'block';
    if (error) error.style.display = 'none';
    if (content) content.innerHTML = '';
}

function hideLoading() {
    const loading = document.getElementById('loading');
    if (loading) loading.style.display = 'none';
}

function showError() {
    const loading = document.getElementById('loading');
    const error = document.getElementById('error');
    
    if (loading) loading.style.display = 'none';
    if (error) error.style.display = 'block';
}

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

function filterDepots() {
    if (!apiData || !apiData.depots) return [];
    let depots = apiData.depots;
    if (currentFilter !== 'all') {
        depots = depots.filter(d => d.depo_code === currentFilter);
    }
    return depots;
}

function filterLocomotives(locomotives) {
    if (!searchTerm) return locomotives;
    return locomotives.filter(loco => 
        loco.model.toLowerCase().includes(searchTerm) ||
        loco.number.toLowerCase().includes(searchTerm)
    );
}

// Рендер карточек с информацией о маршрутах
function renderCardsView(depots) {
    let html = '';
    
    depots.forEach(depot => {
        const filteredLocomotives = filterLocomotives(depot.locomotives);
        
        html += `
            <div class="depot-section">
                <div class="depot-title" onclick="toggleDepot(this)">
                    <h2>
                        Депо ${depot.depo_code}
                    </h2>
                    <span class="badge">${filteredLocomotives.length} локомотивов</span>
                    <i class="fas fa-chevron-down"></i>
                </div>
                
                <div class="branches-grid" style="display: none;">
                    ${filteredLocomotives.map(loco => renderLocoCard(loco, depot.depo_code)).join('')}
                </div>
            </div>
        `;
    });
    
    return html;
}

// Рендер карточки одного локомотива с маршрутами
function renderLocoCard(loco, depoCode) {
    const mostPopular = loco.most_popular;
    const popularPercentage = mostPopular ? mostPopular.percentage.toFixed(1) : 0;
    const directions = loco.directions || [];
    const topDirections = [...directions]
        .sort((a, b) => b.visits - a.visits)
        .slice(0, 3);
    
    return `
        <div class="branch-card" onclick="showLocomotiveDetails('${loco.model}', '${loco.number}')">
            <div class="branch-header">
                <span class="branch-id">${loco.model}</span>
                <span class="branch-length">№ ${loco.number}</span>
            </div>
            
            <div class="branch-route">
                <div class="route-label">
                    Статистика поездок:
                </div>
                <div class="route-stats">
                    <div class="stat-item">
                        <span class="stat-label">Всего:</span>
                        <span class="stat-value">${loco.total_trips || 0}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Направлений:</span>
                        <span class="stat-value">${directions.length}</span>
                    </div>
                </div>
            </div>
            
            ${mostPopular ? `
                <div class="popular-route">
                    <div class="route-label">
                        Самое популярное:
                    </div>
                    <div class="popular-direction">
                        <span class="direction-name">${mostPopular.direction_name}</span>
                        <span class="direction-percentage">${popularPercentage}%</span>
                    </div>
                </div>
            ` : ''}
            
            ${topDirections.length > 0 ? `
                <div class="other-routes">
                    <div class="route-label">
                        Другие маршруты:
                    </div>
                    <div class="routes-list">
                        ${topDirections.map(dir => `
                            <div class="route-item">
                                <span class="route-name">${dir.name}</span>
                                <span class="route-percentage">${dir.percentage.toFixed(1)}%</span>
                            </div>
                        `).join('')}
                        ${directions.length > 3 ? `
                            <div class="route-item text-muted">
                                ... и еще ${directions.length - 3}
                            </div>
                        ` : ''}
                    </div>
                </div>
            ` : ''}
            
            <div class="example-path">
                Активность: ${getActivityLevel(loco.total_trips)}
            </div>
        </div>
    `;
}

function getActivityLevel(trips) {
    if (trips > 30) return 'Высокая';
    if (trips > 15) return 'Средняя';
    return 'Низкая';
}

// Рендер таблицы
function renderTableView(depots) {
    let html = '<div class="table-container"><table class="data-table">';
    html += '<thead><tr><th>Депо</th><th>Модель</th><th>Номер</th><th>Поездок</th><th>Популярное направление</th><th>%</th></tr></thead><tbody>';
    
    depots.forEach(depot => {
        const filteredLocomotives = filterLocomotives(depot.locomotives);
        
        html += `
            <tr class="depot-row" onclick="toggleDepotRows('${depot.depo_code}')">
                <td colspan="6">
                    Депо ${depot.depo_code} (${filteredLocomotives.length} локомотивов)
                    <i class="fas fa-chevron-down" style="float: right;"></i>
                </td>
            </tr>
        `;
        
        filteredLocomotives.forEach(loco => {
            const mostPopular = loco.most_popular;
            html += `
                <tr class="depot-${depot.depo_code}-rows" onclick="showLocomotiveDetails('${loco.model}', '${loco.number}')">
                    <td>${depot.depo_code}</td>
                    <td>${loco.model || ''}</td>
                    <td>${loco.number || ''}</td>
                    <td>${loco.total_trips || 0}</td>
                    <td>${mostPopular ? mostPopular.direction_name : 'Нет данных'}</td>
                    <td>${mostPopular ? mostPopular.percentage.toFixed(1) + '%' : '-'}</td>
                </tr>
            `;
        });
    });
    
    html += '</tbody></table></div>';
    return html;
}

function toggleDepotRows(depoCode) {
    const rows = document.querySelectorAll(`.depot-${depoCode}-rows`);
    const icon = event.currentTarget.querySelector('.fa-chevron-down');
    
    rows.forEach(row => {
        row.style.display = row.style.display === 'none' ? 'table-row' : 'none';
    });
    
    if (icon) {
        icon.style.transform = icon.style.transform === 'rotate(-90deg)' ? 'rotate(0deg)' : 'rotate(-90deg)';
    }
}

// Рендер статистики
function renderStatsView() {
    if (!apiData) return '';
    
    setTimeout(() => {
        createDepotChart();
        createTripsDistributionChart();
        createTopLocomotivesChart();
    }, 100);
    
    return `
        <div class="stats-container">
            <div class="stat-card">
                <h3>Распределение локомотивов по депо</h3>
                <canvas id="depotChart"></canvas>
            </div>
            <div class="stat-card">
                <h3>Распределение поездок по депо</h3>
                <canvas id="tripsChart"></canvas>
            </div>
            <div class="stat-card full-width">
                <h3>Топ-10 самых активных локомотивов</h3>
                <canvas id="topLocomotivesChart"></canvas>
            </div>
            <div class="stat-card full-width">
                <h3>Рейтинг депо по активности</h3>
                <div class="depot-ranking">
                    ${renderDepotRanking()}
                </div>
            </div>
        </div>
    `;
}

function renderDepotRanking() {
    if (!apiData || !apiData.depots) return '';
    
    const sortedDepots = [...apiData.depots].sort((a, b) => b.locomotive_count - a.locomotive_count);
    
    return sortedDepots.map((depot, index) => {
        const totalTrips = depot.locomotives.reduce((sum, loco) => sum + (loco.total_trips || 0), 0);
        const avgTrips = totalTrips > 0 ? (totalTrips / depot.locomotive_count).toFixed(1) : '0';
        
        return `
            <div class="depot-rank-item">
                <span class="rank">${index + 1}</span>
                <span class="depot-name">Депо ${depot.depo_code}</span>
                <div class="depot-stats">
                    <div class="stat">
                        <span class="stat-label">локомотивов</span>
                        <span class="stat-value">${depot.locomotive_count}</span>
                    </div>
                    <div class="stat">
                        <span class="stat-label">поездок</span>
                        <span class="stat-value">${totalTrips}</span>
                    </div>
                    <div class="stat">
                        <span class="stat-label">среднее</span>
                        <span class="stat-value">${avgTrips}</span>
                    </div>
                </div>
            </div>
        `;
    }).join('');
}

// Создание графика депо
function createDepotChart() {
    const canvas = document.getElementById('depotChart');
    if (!canvas || !apiData.depots) return;
    
    const ctx = canvas.getContext('2d');
    const existingChart = Chart.getChart(canvas);
    if (existingChart) existingChart.destroy();
    
    new Chart(ctx, {
        type: 'bar',
        data: {
            labels: apiData.depots.map(d => `Депо ${d.depo_code}`),
            datasets: [{
                label: 'Количество локомотивов',
                data: apiData.depots.map(d => d.locomotive_count),
                backgroundColor: '#667eea',
                borderRadius: 6
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: { display: false },
                tooltip: {
                    callbacks: {
                        label: (context) => `Локомотивов: ${context.raw}`
                    }
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'Количество локомотивов'
                    }
                }
            }
        }
    });
}

// Создание графика распределения поездок
function createTripsDistributionChart() {
    const canvas = document.getElementById('tripsChart');
    if (!canvas || !apiData.depots) return;
    
    const ctx = canvas.getContext('2d');
    const existingChart = Chart.getChart(canvas);
    if (existingChart) existingChart.destroy();
    
    const depotData = apiData.depots.map(depot => {
        const totalTrips = depot.locomotives.reduce((sum, loco) => sum + (loco.total_trips || 0), 0);
        return {
            depot: depot.depo_code,
            trips: totalTrips
        };
    }).filter(d => d.trips > 0);
    
    if (depotData.length === 0) return;
    
    new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: depotData.map(d => `Депо ${d.depot}`),
            datasets: [{
                data: depotData.map(d => d.trips),
                backgroundColor: [
                    '#667eea', '#764ba2', '#f39c12', '#e74c3c', '#2ecc71',
                    '#3498db', '#9b59b6', '#1abc9c', '#f1c40f', '#e67e22'
                ],
                borderWidth: 0
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: { position: 'bottom' },
                tooltip: {
                    callbacks: {
                        label: (context) => {
                            const value = context.raw;
                            const total = context.dataset.data.reduce((a, b) => a + b, 0);
                            const percentage = ((value / total) * 100).toFixed(1);
                            return `${context.label}: ${value} поездок (${percentage}%)`;
                        }
                    }
                }
            }
        }
    });
}

// Создание графика топ локомотивов
function createTopLocomotivesChart() {
    const canvas = document.getElementById('topLocomotivesChart');
    if (!canvas || !apiData.depots) return;
    
    const ctx = canvas.getContext('2d');
    const existingChart = Chart.getChart(canvas);
    if (existingChart) existingChart.destroy();
    
    const allLocomotives = [];
    apiData.depots.forEach(depot => {
        depot.locomotives.forEach(loco => {
            allLocomotives.push({
                name: `${loco.model}-${loco.number}`,
                trips: loco.total_trips || 0,
                depot: depot.depo_code,
                model: loco.model,
                number: loco.number
            });
        });
    });
    
    const topLocomotives = allLocomotives
        .sort((a, b) => b.trips - a.trips)
        .slice(0, 10);
    
    if (topLocomotives.length === 0) return;
    
    new Chart(ctx, {
        type: 'bar',
        data: {
            labels: topLocomotives.map(l => l.name),
            datasets: [{
                label: 'Количество поездок',
                data: topLocomotives.map(l => l.trips),
                backgroundColor: '#667eea',
                borderRadius: 6
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            onClick: (event, elements) => {
                if (elements && elements.length > 0) {
                    const index = elements[0].index;
                    const loco = topLocomotives[index];
                    showLocomotiveDetails(loco.model, loco.number);
                }
            },
            plugins: {
                legend: { display: false },
                tooltip: {
                    callbacks: {
                        label: (context) => {
                            const loco = topLocomotives[context.dataIndex];
                            return [`Поездок: ${context.raw}`, `Депо: ${loco.depot}`];
                        }
                    }
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'Количество поездок'
                    }
                },
                x: {
                    ticks: {
                        maxRotation: 45,
                        minRotation: 45
                    }
                }
            }
        }
    });
}

function toggleDepot(element) {
    const grid = element.parentElement.querySelector('.branches-grid');
    const icon = element.querySelector('.fa-chevron-down');
    
    if (grid) {
        grid.style.display = grid.style.display === 'none' ? 'grid' : 'none';
        if (icon) {
            icon.style.transform = grid.style.display === 'grid' ? 'rotate(0deg)' : 'rotate(-90deg)';
        }
    }
}

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

// Показать модальное окно с детальной информацией
function showModal(loco, depoCode) {
    const modal = document.getElementById('locomotiveModal');
    const modalBody = document.getElementById('modalBody');
    
    if (!modal || !modalBody) return;
    
    const activityLevel = getActivityLevel(loco.total_trips);
    const activityColor = loco.total_trips > 30 ? '#2ecc71' : loco.total_trips > 15 ? '#f39c12' : '#e74c3c';
    const activityPercent = Math.min(100, (loco.total_trips / 50) * 100);
    const directions = loco.directions || [];
    
    modalBody.innerHTML = `
        <div class="loco-detail">
            <div class="detail-section">
                <h3>Информация о локомотиве</h3>
                <div class="detail-grid">
                    <div class="detail-item">
                        <span class="label">Модель</span>
                        <span class="value">${loco.model}</span>
                    </div>
                    <div class="detail-item">
                        <span class="label">Номер</span>
                        <span class="value">${loco.number}</span>
                    </div>
                    <div class="detail-item">
                        <span class="label">Депо приписки</span>
                        <span class="value">${depoCode}</span>
                    </div>
                    <div class="detail-item">
                        <span class="label">Всего поездок</span>
                        <span class="value">${loco.total_trips}</span>
                    </div>
                </div>
            </div>
            
            <div class="detail-section">
                <h3>Анализ активности</h3>
                <div style="margin: 15px 0;">
                    <div style="display: flex; justify-content: space-between; margin-bottom: 5px;">
                        <span>Уровень активности</span>
                        <span style="color: ${activityColor}; font-weight: 600;">${activityLevel}</span>
                    </div>
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: ${activityPercent}%; background: ${activityColor};"></div>
                    </div>
                </div>
                <div class="stats-list" style="margin-top: 15px;">
                    <li style="display: flex; justify-content: space-between; padding: 8px 0;">
                        <span>Среднее по депо:</span>
                        <span class="text-success">${calculateDepotAverage(depoCode)} поездок</span>
                    </li>
                    <li style="display: flex; justify-content: space-between; padding: 8px 0;">
                        <span>Место в депо:</span>
                        <span class="text-warning">${calculateLocoRank(loco, depoCode)} из ${getDepotLocomotiveCount(depoCode)}</span>
                    </li>
                </div>
            </div>
            
            ${directions.length > 0 ? `
                <div class="detail-section">
                    <h3>Все направления</h3>
                    <div class="directions-table">
                        ${directions.map(dir => `
                            <div class="direction-row">
                                <span class="direction-name">${dir.name}</span>
                                <div class="direction-stats">
                                    <span class="direction-visits">${dir.visits} поездок</span>
                                    <span class="direction-percentage">${dir.percentage.toFixed(1)}%</span>
                                    <div class="progress-bar mini">
                                        <div class="progress-fill" style="width: ${dir.percentage}%; background: #667eea;"></div>
                                    </div>
                                </div>
                            </div>
                        `).join('')}
                    </div>
                </div>
            ` : ''}
            
            <div class="detail-section">
                <h3>Дополнительная информация</h3>
                <div style="text-align: center; padding: 10px;">
                    <span class="badge" style="background: #667eea; color: white; padding: 8px 20px;">
                        Последняя активность: ${loco.last_trip || 'Нет данных'}
                    </span>
                </div>
            </div>
        </div>
    `;
    
    modal.classList.add('show');
    document.body.style.overflow = 'hidden';
}

function calculateDepotAverage(depoCode) {
    const depot = apiData.depots.find(d => d.depo_code === depoCode);
    if (!depot || depot.locomotives.length === 0) return '0';
    const total = depot.locomotives.reduce((sum, loco) => sum + (loco.total_trips || 0), 0);
    return (total / depot.locomotives.length).toFixed(1);
}

function calculateLocoRank(loco, depoCode) {
    const depot = apiData.depots.find(d => d.depo_code === depoCode);
    if (!depot) return '?';
    const sorted = [...depot.locomotives].sort((a, b) => (b.total_trips || 0) - (a.total_trips || 0));
    const rank = sorted.findIndex(l => l.model === loco.model && l.number === loco.number) + 1;
    return rank || '?';
}

function getDepotLocomotiveCount(depoCode) {
    const depot = apiData.depots.find(d => d.depo_code === depoCode);
    return depot ? depot.locomotives.length : '?';
}

function closeModal() {
    const modal = document.getElementById('locomotiveModal');
    if (modal) {
        modal.classList.remove('show');
        document.body.style.overflow = '';
    }
}

window.onclick = function(event) {
    const modal = document.getElementById('locomotiveModal');
    if (event.target === modal) {
        closeModal();
    }
}

// Глобальные функции для HTML
window.toggleDepot = toggleDepot;
window.toggleDepotRows = toggleDepotRows;
window.showLocomotiveDetails = showLocomotiveDetails;
window.closeModal = closeModal;
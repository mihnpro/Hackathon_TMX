// Глобальные переменные
let apiData = null;
let currentView = 'cards';
let currentDepo = 'all';
let searchTerm = '';

// Загрузка данных при старте
document.addEventListener('DOMContentLoaded', () => {
    console.log('Задание 1: загрузка данных...');
    fetchData();
    setupEventListeners();
});

// Получение данных с API
async function fetchData() {
    showLoading();
    try {
        console.log('Запрос к API: /api/v1/task1/branches');
        const response = await fetch('/api/v1/task1/branches');
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        apiData = await response.json();
        console.log('Данные получены:', apiData);
        
        updateDepoSelect();
        updateStatsHeader();
        renderContent();
        hideLoading();
    } catch (error) {
        console.error('Error fetching data:', error);
        showError();
    }
}

// Обновление выпадающего списка депо
function updateDepoSelect() {
    const select = document.getElementById('depoSelect');
    if (!select || !apiData || !apiData.depots) return;
    
    let options = '<option value="all">Все депо</option>';
    
    apiData.depots.forEach(depot => {
        options += `<option value="${depot.depo_code}">Депо ${depot.depo_code} (${depot.branch_count} веток)</option>`;
    });
    
    select.innerHTML = options;
    
    // Добавляем обработчик
    select.addEventListener('change', (e) => {
        currentDepo = e.target.value;
        renderContent();
    });
}

// Обновление статистики в шапке
function updateStatsHeader() {
    if (!apiData || !apiData.overall_stats) return;
    
    const stats = apiData.overall_stats;
    
    document.getElementById('totalDepots').textContent = stats.total_depots || 0;
    document.getElementById('totalBranches').textContent = stats.total_branches || 0;
    document.getElementById('totalTerminals').textContent = stats.total_terminals || 0;
    document.getElementById('avgBranches').textContent = stats.avg_branches_per_depo 
        ? stats.avg_branches_per_depo.toFixed(1) 
        : '0';
}

// Настройка обработчиков событий
function setupEventListeners() {
    // Поиск
    const searchInput = document.getElementById('branchSearch');
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
    
    if (currentDepo !== 'all') {
        depots = depots.filter(d => d.depo_code === currentDepo);
    }
    
    return depots;
}

// Фильтрация веток по поиску
function filterBranches(branches) {
    if (!searchTerm) return branches;
    
    return branches.filter(branch => {
        // Поиск по ID ветки
        if (branch.branch_id.toLowerCase().includes(searchTerm)) return true;
        
        // Поиск по станциям в маршруте
        if (branch.core_stations.some(s => s.toLowerCase().includes(searchTerm))) return true;
        
        // Поиск по конечным станциям
        if (branch.terminals.some(t => t.station.toLowerCase().includes(searchTerm))) return true;
        
        return false;
    });
}

// Рендер карточек
function renderCardsView(depots) {
    let html = '';
    
    depots.forEach(depot => {
        const filteredBranches = filterBranches(depot.branches);
        
        html += `
            <div class="depot-section">
                <div class="depot-title" onclick="toggleDepotBranches(this)">
                    <h2>Депо ${depot.depo_code}</h2>
                    <span class="badge">${depot.branch_count} веток</span>
                    <i class="fas fa-chevron-down"></i>
                </div>
                
                <div class="branches-grid" style="display: grid;">
                    ${filteredBranches.map(branch => renderBranchCard(branch)).join('')}
                </div>
            </div>
        `;
    });
    
    return html;
}

// Рендер карточки ветки
function renderBranchCard(branch) {
    // Ограничиваем отображение станций
    const displayStations = branch.core_stations.slice(0, 8);
    const hasMore = branch.core_stations.length > 8;
    
    // Топ-3 конечных станции
    const topTerminals = branch.terminals.slice(0, 3);
    
    return `
        <div class="branch-card" onclick="showBranchDetails('${branch.branch_id}')">
            <div class="branch-header">
                <span class="branch-id">${branch.branch_id}</span>
                <span class="branch-length">${branch.station_count} станций</span>
            </div>
            
            <div class="branch-route">
                <div class="route-label">Основной маршрут:</div>
                <div class="route-stations">
                    ${displayStations.map(s => `<span class="station-badge">${s}</span>`).join('')}
                    ${hasMore ? '<span class="station-badge">...</span>' : ''}
                </div>
            </div>
            
            <div class="terminals-section">
                <div class="terminals-title">Конечные станции:</div>
                <div class="terminals-list">
                    ${topTerminals.map(t => `
                        <div class="terminal-item">
                            <span class="terminal-station">${t.station}</span>
                            <span class="terminal-frequency">${t.frequency.toFixed(1)}%</span>
                        </div>
                    `).join('')}
                    ${branch.terminals.length > 3 ? 
                        `<div class="terminal-item">
                            <span>и еще ${branch.terminals.length - 3}...</span>
                        </div>` : ''}
                </div>
            </div>
            
            ${branch.example_path && branch.example_path.length > 0 ? `
                <div class="example-path">
                    <i class="fas fa-route"></i> Пример: 
                    <span>${branch.example_path.slice(0, 5).join(' → ')}${branch.example_path.length > 5 ? '...' : ''}</span>
                </div>
            ` : ''}
        </div>
    `;
}

// Рендер таблицы
function renderTableView(depots) {
    let html = '<div class="table-container"><table class="data-table"><thead><tr>';
    html += '<th>Депо</th><th>ID ветки</th><th>Станций</th><th>Основной маршрут</th><th>Конечные станции</th></tr></thead><tbody>';
    
    depots.forEach(depot => {
        const filteredBranches = filterBranches(depot.branches);
        
        // Строка с названием депо
        html += `
            <tr class="depo-row">
                <td colspan="5"><strong>Депо ${depot.depo_code}</strong> (${filteredBranches.length} веток)</td>
            </tr>
        `;
        
        filteredBranches.forEach(branch => {
            const routePreview = branch.core_stations.slice(0, 5).join(' → ') + 
                (branch.core_stations.length > 5 ? '...' : '');
            
            const terminalsPreview = branch.terminals.slice(0, 3).map(t => 
                `${t.station} (${t.frequency.toFixed(0)}%)`
            ).join(', ') + (branch.terminals.length > 3 ? '...' : '');
            
            html += `
                <tr onclick="showBranchDetails('${branch.branch_id}')">
                    <td>${depot.depo_code}</td>
                    <td><span class="branch-id">${branch.branch_id}</span></td>
                    <td>${branch.station_count}</td>
                    <td>${routePreview}</td>
                    <td>${terminalsPreview}</td>
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
        createDepotChart();
        createTerminalsChart();
    }, 100);
    
    return `
        <div class="stats-container">
            <div class="stat-card">
                <h3><i class="fas fa-chart-pie"></i> Распределение веток по депо</h3>
                <canvas id="depotChart"></canvas>
            </div>
            
            <div class="stat-card">
                <h3><i class="fas fa-chart-bar"></i> Топ-10 самых длинных веток</h3>
                <canvas id="longestBranchesChart"></canvas>
            </div>
            
            <div class="stat-card">
                <h3><i class="fas fa-trophy"></i> Самые длинные ветки</h3>
                <ul class="longest-branches-list">
                    ${apiData.longest_branches.slice(0, 10).map((branch, index) => `
                        <li onclick="showBranchDetails('${branch.branch_id}')" style="cursor: pointer;">
                            <span class="rank">#${index + 1}</span>
                            <span class="depo">${branch.depo_code}</span>
                            <span class="route">${branch.route_string.substring(0, 30)}${branch.route_string.length > 30 ? '...' : ''}</span>
                            <span class="length">${branch.length}</span>
                        </li>
                    `).join('')}
                </ul>
            </div>
            
            <div class="stat-card">
                <h3><i class="fas fa-chart-line"></i> Детальная статистика</h3>
                <div style="padding: 20px;">
                    <p><strong>Всего депо:</strong> ${apiData.overall_stats.total_depots}</p>
                    <p><strong>Всего веток:</strong> ${apiData.overall_stats.total_branches}</p>
                    <p><strong>Всего конечных станций:</strong> ${apiData.overall_stats.total_terminals}</p>
                    <p><strong>Среднее веток на депо:</strong> ${apiData.overall_stats.avg_branches_per_depo.toFixed(2)}</p>
                    <p><strong>Самая длинная ветка:</strong> ${apiData.longest_branches[0]?.length || 0} станций</p>
                </div>
            </div>
        </div>
    `;
}

// Создание графика депо
function createDepotChart() {
    const ctx = document.getElementById('depotChart');
    if (!ctx || !apiData || !apiData.depots) return;
    
    new Chart(ctx, {
        type: 'bar',
        data: {
            labels: apiData.depots.map(d => `Депо ${d.depo_code}`),
            datasets: [{
                label: 'Количество веток',
                data: apiData.depots.map(d => d.branch_count),
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

// Создание графика длинных веток
function createTerminalsChart() {
    const ctx = document.getElementById('longestBranchesChart');
    if (!ctx || !apiData || !apiData.longest_branches) return;
    
    const topBranches = apiData.longest_branches.slice(0, 10);
    
    new Chart(ctx, {
        type: 'bar',
        data: {
            labels: topBranches.map(b => `${b.depo_code}`),
            datasets: [{
                label: 'Количество станций',
                data: topBranches.map(b => b.length),
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

// Переключение отображения веток депо
function toggleDepotBranches(element) {
    const grid = element.nextElementSibling;
    const icon = element.querySelector('i');
    
    if (grid.style.display === 'none') {
        grid.style.display = 'grid';
        icon.className = 'fas fa-chevron-down';
    } else {
        grid.style.display = 'none';
        icon.className = 'fas fa-chevron-right';
    }
}

// Показать детали ветки
function showBranchDetails(branchId) {
    // Ищем ветку во всех депо
    for (const depot of apiData.depots) {
        const branch = depot.branches.find(b => b.branch_id === branchId);
        if (branch) {
            showModal(branch, depot.depo_code);
            break;
        }
    }
}

// Показать модальное окно
function showModal(branch, depoCode) {
    const modal = document.getElementById('branchModal');
    const modalBody = document.getElementById('modalBody');
    
    // Формируем отображение маршрута со стрелками
    const stationsFlow = branch.core_stations.map((s, i) => `
        <span class="station-node">${s}</span>
        ${i < branch.core_stations.length - 1 ? '<span class="arrow">→</span>' : ''}
    `).join('');
    
    // Таблица конечных станций
    const terminalsTable = `
        <table class="terminals-table">
            <thead>
                <tr>
                    <th>Станция</th>
                    <th>Посещений</th>
                    <th>Частота</th>
                    <th>Визуализация</th>
                </tr>
            </thead>
            <tbody>
                ${branch.terminals.map(t => `
                    <tr>
                        <td><strong>${t.station}</strong></td>
                        <td>${t.visits}</td>
                        <td>${t.frequency.toFixed(1)}%</td>
                        <td style="width: 200px;">
                            <div class="progress-bar">
                                <div class="progress-fill" style="width: ${t.frequency}%;">${t.frequency.toFixed(0)}%</div>
                            </div>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
    
    // Примеры путей
    const examplesList = branch.example_path && branch.example_path.length > 0 
        ? `<p>${branch.example_path.join(' → ')}</p>`
        : '<p>Нет данных</p>';
    
    modalBody.innerHTML = `
        <div class="branch-detail">
            <div class="detail-section">
                <h3><i class="fas fa-info-circle"></i> Общая информация</h3>
                <p><strong>Депо:</strong> ${depoCode}</p>
                <p><strong>ID ветки:</strong> ${branch.branch_id}</p>
                <p><strong>Количество станций:</strong> ${branch.station_count}</p>
            </div>
            
            <div class="detail-section">
                <h3><i class="fas fa-route"></i> Основной маршрут</h3>
                <div class="stations-flow">
                    ${stationsFlow}
                </div>
            </div>
            
            <div class="detail-section">
                <h3><i class="fas fa-flag-checkered"></i> Конечные станции</h3>
                ${terminalsTable}
            </div>
            
            <div class="detail-section">
                <h3><i class="fas fa-eye"></i> Пример пути</h3>
                ${examplesList}
            </div>
        </div>
    `;
    
    modal.classList.add('show');
}

// Закрыть модальное окно
function closeModal() {
    const modal = document.getElementById('branchModal');
    if (modal) {
        modal.classList.remove('show');
    }
}

// Закрытие по клику вне модального окна
window.onclick = function(event) {
    const modal = document.getElementById('branchModal');
    if (event.target === modal) {
        closeModal();
    }
}

// Глобальные функции для HTML
window.toggleDepotBranches = toggleDepotBranches;
window.showBranchDetails = showBranchDetails;
window.closeModal = closeModal;
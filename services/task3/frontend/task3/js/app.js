// Глобальные переменные
let depots = [];
let currentMaps = null;
let currentDepo = '';

// Загрузка данных при старте
document.addEventListener('DOMContentLoaded', () => {
    console.log('Задание 3: загрузка данных...');
    loadDepots();
    setupEventListeners();
});

// Загрузка списка депо
async function loadDepots() {
    try {
        const response = await fetch('/api/v1/task3/depots');
        if (!response.ok) throw new Error('Failed to load depots');
        
        const data = await response.json();
        depots = data.depots || [];
        
        updateDepoSelect();
    } catch (error) {
        console.error('Error loading depots:', error);
        showError('Не удалось загрузить список депо');
    }
}

// Обновление выпадающего списка депо
function updateDepoSelect() {
    const select = document.getElementById('depoSelect');
    if (!select) return;
    
    let options = '<option value="">Выберите депо</option>';
    if (depots && depots.length > 0) {
        depots.forEach(depo => {
            options += `<option value="${depo}">Депо ${depo}</option>`;
        });
    } else {
        options = '<option value="">Нет доступных депо</option>';
    }
    
    select.innerHTML = options;
}

// Настройка обработчиков событий
function setupEventListeners() {
    // Выбор депо
    const depoSelect = document.getElementById('depoSelect');
    if (depoSelect) {
        depoSelect.addEventListener('change', async (e) => {
            const depoId = e.target.value;
            if (depoId) {
                await loadDepotInfo(depoId);
            } else {
                const depoInfo = document.getElementById('depoInfo');
                if (depoInfo) depoInfo.style.display = 'none';
            }
        });
    }
    
    // Генерация карт
    const generateBtn = document.getElementById('generateBtn');
    if (generateBtn) {
        generateBtn.addEventListener('click', generateMaps);
    }
    
    // Настраиваем обработчики для основных вкладок
    setupMainTabsListeners();
}

// Настройка слушателей для основных вкладок
function setupMainTabsListeners() {
    const mainTabs = document.querySelectorAll('.tab-btn[data-tab="overview"], .tab-btn[data-tab="heatmap"]');
    mainTabs.forEach(btn => {
        btn.removeEventListener('click', handleMainTabClick);
        btn.addEventListener('click', handleMainTabClick);
    });
}

// Обработчик клика по основным вкладкам
function handleMainTabClick(e) {
    e.preventDefault();
    e.stopPropagation();
    
    const tabId = e.currentTarget.dataset.tab;
    console.log('Клик по основной вкладке:', tabId);
    
    if (tabId) {
        // Деактивируем все вкладки локомотивов
        document.querySelectorAll('.loco-tab').forEach(btn => {
            btn.classList.remove('active');
        });
        
        // Активируем основную вкладку
        activateTab(tabId);
    }
}

// Загрузка информации о депо
async function loadDepotInfo(depoId) {
    try {
        const response = await fetch(`/api/v1/task3/depots/${depoId}`);
        if (!response.ok) throw new Error('Failed to load depot info');
        
        const info = await response.json();
        
        const depoRegion = document.getElementById('depoRegion');
        const depoCount = document.getElementById('depoCount');
        const depoInfo = document.getElementById('depoInfo');
        
        if (depoRegion) {
            depoRegion.innerHTML = `<i class="fas fa-map-pin"></i> ${info.region || 'Неизвестно'}`;
        }
        if (depoCount) {
            depoCount.innerHTML = `<i class="fas fa-train"></i> ${info.locomotive_count || 0} локомотивов`;
        }
        if (depoInfo) {
            depoInfo.style.display = 'flex';
        }
        
        currentDepo = depoId;
    } catch (error) {
        console.error('Error loading depot info:', error);
    }
}

// Генерация карт
async function generateMaps() {
    const depoSelect = document.getElementById('depoSelect');
    const maxLocomotivesInput = document.getElementById('maxLocomotives');
    
    const depoId = depoSelect ? depoSelect.value : '';
    const maxLocomotives = maxLocomotivesInput ? parseInt(maxLocomotivesInput.value) || 10 : 10;
    
    if (!depoId) {
        alert('Пожалуйста, выберите депо');
        return;
    }
    
    // Показываем загрузку
    showLoading();
    hideError();
    hideEmptyState();
    
    try {
        const response = await fetch('/api/v1/task3/generate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                depo_id: depoId,
                max_locomotives: maxLocomotives
            })
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to generate maps');
        }
        
        currentMaps = await response.json();
        
        // Отображаем карты
        displayMaps(currentMaps);
        showSuccess(`Карты сгенерированы для депо ${depoId}`);
        
        // Добавляем кнопку информации
        addInfoButton();
        
    } catch (error) {
        console.error('Error generating maps:', error);
        showError(error.message || 'Ошибка при генерации карт');
        showEmptyState();
    } finally {
        hideLoading();
    }
}

// Отображение карт
function displayMaps(maps) {
    // Проверяем наличие данных
    if (!maps || !maps.maps) {
        console.error('Нет данных для отображения');
        showError('Нет данных для отображения карт');
        return;
    }
    
    // Показываем контейнер с картами
    const mapsTabs = document.getElementById('mapsTabs');
    const emptyState = document.getElementById('emptyState');
    
    if (mapsTabs) mapsTabs.style.display = 'block';
    if (emptyState) emptyState.style.display = 'none';
    
    // Устанавливаем URL для iframe
    const overviewFrame = document.getElementById('overviewFrame');
    const heatmapFrame = document.getElementById('heatmapFrame');
    
    if (overviewFrame && maps.maps.overview) {
        overviewFrame.src = maps.maps.overview;
    }
    
    if (heatmapFrame && maps.maps.heatmap) {
        heatmapFrame.src = maps.maps.heatmap;
    }
    
    // Создаем вкладки для локомотивов
    if (maps.maps.locomotives && Array.isArray(maps.maps.locomotives)) {
        createLocomotiveTabs(maps.maps.locomotives);
    } else {
        console.warn('Нет данных о локомотивах');
    }
    
    // Перенастраиваем обработчики для основных вкладок
    setupMainTabsListeners();
    
    // Активируем первую вкладку (общая карта)
    setTimeout(() => {
        activateTab('overview');
    }, 100);
}

// Создание вкладок для локомотивов
function createLocomotiveTabs(locomotives) {
    const container = document.getElementById('locomotivesTabs');
    const panesContainer = document.getElementById('locomotivesPanes');
    
    if (!container || !panesContainer) {
        console.error('Контейнеры для вкладок не найдены');
        return;
    }
    
    container.innerHTML = '';
    panesContainer.innerHTML = '';
    
    // Проверяем, что locomotives - массив и не пустой
    if (!locomotives || !Array.isArray(locomotives) || locomotives.length === 0) {
        console.warn('Нет локомотивов для отображения');
        container.innerHTML = '<div class="no-locomotives">Нет локомотивов для отображения</div>';
        return;
    }
    
    // Используем цикл for...of для безопасного перебора
    let index = 0;
    for (const loco of locomotives) {
        try {
            // Создаем кнопку вкладки
            const tabBtn = document.createElement('button');
            tabBtn.className = 'loco-tab';
            tabBtn.dataset.tab = `loco-${index}`;
            
            // Безопасно получаем модель и номер
            const model = loco.model || 'Локомотив';
            const number = loco.number || index;
            tabBtn.innerHTML = `<i class="fas fa-train"></i> ${model}-${number}`;
            
            // Добавляем обработчик клика
            tabBtn.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                
                const tabId = e.currentTarget.dataset.tab;
                console.log('Клик по вкладке локомотива:', tabId);
                
                // Деактивируем все основные вкладки
                document.querySelectorAll('.tab-btn[data-tab="overview"], .tab-btn[data-tab="heatmap"]').forEach(btn => {
                    btn.classList.remove('active');
                });
                
                // Активируем вкладку локомотива
                activateTab(tabId);
            });
            
            container.appendChild(tabBtn);
            
            // Создаем панель с iframe
            const pane = document.createElement('div');
            pane.className = 'tab-pane';
            pane.id = `loco-${index}Tab`;
            
            const iframe = document.createElement('iframe');
            iframe.className = 'map-frame';
            
            // Безопасно устанавливаем URL
            if (loco.url) {
                iframe.src = loco.url;
            } else {
                iframe.src = 'about:blank';
                console.warn(`Локомотив ${index} не имеет URL`);
            }
            
            iframe.setAttribute('data-loco-index', index);
            
            pane.appendChild(iframe);
            panesContainer.appendChild(pane);
            
            index++;
        } catch (error) {
            console.error(`Ошибка при создании вкладки для локомотива ${index}:`, error);
        }
    }
}

// Активация вкладки
function activateTab(tabId) {
    console.log('Активация вкладки:', tabId);
    
    // Проверяем, что tabId - строка
    if (typeof tabId !== 'string') {
        console.error('Некорректный идентификатор вкладки:', tabId);
        return;
    }
    
    // Деактивируем все кнопки вкладок
    document.querySelectorAll('.tab-btn, .loco-tab').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Скрываем все панели
    document.querySelectorAll('.tab-pane').forEach(pane => {
        pane.classList.remove('active');
    });
    
    // Активируем нужную кнопку
    const targetBtn = document.querySelector(`.tab-btn[data-tab="${tabId}"], .loco-tab[data-tab="${tabId}"]`);
    if (targetBtn) {
        targetBtn.classList.add('active');
    } else {
        console.warn('Кнопка вкладки не найдена:', tabId);
    }
    
    // Определяем ID панели для отображения
    let paneId;
    if (tabId === 'overview') {
        paneId = 'overviewTab';
    } else if (tabId === 'heatmap') {
        paneId = 'heatmapTab';
    } else {
        paneId = `${tabId}Tab`; // для локомотивов: loco-0Tab, loco-1Tab и т.д.
    }
    
    // Показываем нужную панель
    const targetPane = document.getElementById(paneId);
    if (targetPane) {
        targetPane.classList.add('active');
        console.log('Активирована панель:', paneId);
    } else {
        console.warn('Панель не найдена:', paneId);
    }
}

// Показать загрузку
function showLoading() {
    const loading = document.getElementById('loading');
    if (loading) loading.style.display = 'block';
}

// Скрыть загрузку
function hideLoading() {
    const loading = document.getElementById('loading');
    if (loading) loading.style.display = 'none';
}

// Показать ошибку
function showError(message) {
    const errorEl = document.getElementById('error');
    const errorMessage = document.getElementById('errorMessage');
    
    if (errorEl) errorEl.style.display = 'block';
    if (errorMessage) errorMessage.textContent = message || 'Произошла ошибка';
}

// Скрыть ошибку
function hideError() {
    const errorEl = document.getElementById('error');
    if (errorEl) errorEl.style.display = 'none';
}

// Показать пустое состояние
function showEmptyState() {
    const emptyState = document.getElementById('emptyState');
    const mapsTabs = document.getElementById('mapsTabs');
    
    if (emptyState) emptyState.style.display = 'block';
    if (mapsTabs) mapsTabs.style.display = 'none';
}

// Скрыть пустое состояние
function hideEmptyState() {
    const emptyState = document.getElementById('emptyState');
    if (emptyState) emptyState.style.display = 'none';
}

// Показать успех
function showSuccess(message) {
    console.log('✅', message);
}

// Добавление кнопки информации
function addInfoButton() {
    // Проверяем, существует ли уже кнопка
    if (document.querySelector('.info-btn')) {
        return;
    }
    
    const controlPanel = document.querySelector('.control-panel');
    if (!controlPanel) return;
    
    const infoBtn = document.createElement('button');
    infoBtn.className = 'info-btn';
    infoBtn.innerHTML = '<i class="fas fa-info-circle"></i> Инфо';
    infoBtn.onclick = showGenerationInfo;
    infoBtn.title = 'Информация о генерации';
    infoBtn.style.cssText = `
        background: #4CAF50;
        color: white;
        border: none;
        padding: 12px 20px;
        border-radius: 10px;
        font-size: 1em;
        font-weight: 600;
        cursor: pointer;
        display: flex;
        align-items: center;
        gap: 8px;
        transition: all 0.3s;
    `;
    controlPanel.appendChild(infoBtn);
}

// Показать информацию о генерации
function showGenerationInfo() {
    if (!currentMaps) return;
    
    const modal = document.getElementById('infoModal');
    const body = document.getElementById('infoModalBody');
    
    if (!modal || !body) return;
    
    const locomotives = currentMaps.maps?.locomotives || [];
    
    const locoList = locomotives.map(loco => 
        `<li>
            <i class="fas fa-train"></i>
            <a href="${loco.url || '#'}" target="_blank">${loco.model || 'Локомотив'}-${loco.number || '??'} (${loco.trip_count || 0} поездок)</a>
        </li>`
    ).join('');
    
    body.innerHTML = `
        <div class="info-item">
            <span class="info-label">Депо:</span>
            <span class="info-value">${currentMaps.depot_id || 'Неизвестно'}</span>
        </div>
        <div class="info-item">
            <span class="info-label">Сгенерировано:</span>
            <span class="info-value">${currentMaps.generated_at ? new Date(currentMaps.generated_at).toLocaleString() : 'Неизвестно'}</span>
        </div>
        <div class="info-item">
            <span class="info-label">Карты:</span>
            <span class="info-value">
                <a href="${currentMaps.maps?.overview || '#'}" target="_blank">Общая</a> | 
                <a href="${currentMaps.maps?.heatmap || '#'}" target="_blank">Тепловая</a>
            </span>
        </div>
        <h4>Локомотивы (${locomotives.length}):</h4>
        <ul class="maps-list">
            ${locoList || '<li>Нет данных о локомотивах</li>'}
        </ul>
    `;
    
    modal.classList.add('show');
}

// Закрыть модальное окно
function closeInfoModal() {
    const modal = document.getElementById('infoModal');
    if (modal) modal.classList.remove('show');
}

// Добавляем стили для кнопки информации и сообщения об отсутствии локомотивов
const style = document.createElement('style');
style.textContent = `
    .info-btn:hover {
        transform: translateY(-2px);
        box-shadow: 0 10px 20px rgba(76, 175, 80, 0.3);
    }
    
    .no-locomotives {
        padding: 20px;
        text-align: center;
        color: #666;
        font-style: italic;
    }
`;
document.head.appendChild(style);

// Глобальные функции
window.activateTab = activateTab;
window.closeInfoModal = closeInfoModal;
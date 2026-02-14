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
    
    let options = '<option value="">Выберите депо</option>';
    depots.forEach(depo => {
        options += `<option value="${depo}">Депо ${depo}</option>`;
    });
    
    select.innerHTML = options;
}

// Настройка обработчиков событий
function setupEventListeners() {
    // Выбор депо
    document.getElementById('depoSelect').addEventListener('change', async (e) => {
        const depoId = e.target.value;
        if (depoId) {
            await loadDepotInfo(depoId);
        } else {
            document.getElementById('depoInfo').style.display = 'none';
        }
    });
    
    // Генерация карт
    document.getElementById('generateBtn').addEventListener('click', generateMaps);
}

// Загрузка информации о депо
async function loadDepotInfo(depoId) {
    try {
        const response = await fetch(`/api/v1/task3/depots/${depoId}`);
        if (!response.ok) throw new Error('Failed to load depot info');
        
        const info = await response.json();
        
        document.getElementById('depoRegion').innerHTML = `<i class="fas fa-map-pin"></i> ${info.region}`;
        document.getElementById('depoCount').innerHTML = `<i class="fas fa-train"></i> ${info.locomotive_count} локомотивов`;
        document.getElementById('depoInfo').style.display = 'flex';
        
        currentDepo = depoId;
    } catch (error) {
        console.error('Error loading depot info:', error);
    }
}

// Генерация карт
async function generateMaps() {
    const depoId = document.getElementById('depoSelect').value;
    const maxLocomotives = parseInt(document.getElementById('maxLocomotives').value) || 10;
    
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
    // Показываем контейнер с картами
    document.getElementById('mapsTabs').style.display = 'block';
    document.getElementById('emptyState').style.display = 'none';
    
    // Устанавливаем URL для iframe
    document.getElementById('overviewFrame').src = maps.maps.overview;
    document.getElementById('heatmapFrame').src = maps.maps.heatmap;
    
    // Создаем вкладки для локомотивов
    createLocomotiveTabs(maps.maps.locomotives);
    
    // Активируем первую вкладку
    activateTab('overview');
}

// Создание вкладок для локомотивов
function createLocomotiveTabs(locomotives) {
    const container = document.getElementById('locomotivesTabs');
    const panesContainer = document.getElementById('locomotivesPanes');
    
    container.innerHTML = '';
    panesContainer.innerHTML = '';
    
    locomotives.forEach((loco, index) => {
        // Создаем кнопку вкладки
        const tabBtn = document.createElement('button');
        tabBtn.className = 'loco-tab';
        tabBtn.dataset.tab = `loco-${index}`;
        tabBtn.innerHTML = `<i class="fas fa-train"></i> ${loco.model}-${loco.number}`;
        tabBtn.onclick = () => activateTab(`loco-${index}`);
        container.appendChild(tabBtn);
        
        // Создаем панель с iframe
        const pane = document.createElement('div');
        pane.className = 'tab-pane';
        pane.id = `loco-${index}Tab`;
        
        const iframe = document.createElement('iframe');
        iframe.className = 'map-frame';
        iframe.src = loco.url;
        
        pane.appendChild(iframe);
        panesContainer.appendChild(pane);
    });
}

// Активация вкладки
function activateTab(tabId) {
    // Деактивируем все кнопки
    document.querySelectorAll('.tab-btn, .loco-tab').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Скрываем все панели
    document.querySelectorAll('.tab-pane').forEach(pane => {
        pane.classList.remove('active');
    });
    
    // Активируем нужную кнопку
    let selector = `.tab-btn[data-tab="${tabId}"]`;
    if (document.querySelector(selector)) {
        document.querySelector(selector).classList.add('active');
    } else {
        // Возможно это вкладка локомотива
        document.querySelectorAll('.loco-tab').forEach((btn, index) => {
            if (`loco-${index}` === tabId) {
                btn.classList.add('active');
            }
        });
    }
    
    // Показываем нужную панель
    const paneId = tabId === 'overview' ? 'overviewTab' :
                   tabId === 'heatmap' ? 'heatmapTab' :
                   `${tabId}Tab`;
    
    document.getElementById(paneId)?.classList.add('active');
}

// Показать загрузку
function showLoading() {
    document.getElementById('loading').style.display = 'block';
}

// Скрыть загрузку
function hideLoading() {
    document.getElementById('loading').style.display = 'none';
}

// Показать ошибку
function showError(message) {
    const errorEl = document.getElementById('error');
    document.getElementById('errorMessage').textContent = message;
    errorEl.style.display = 'block';
}

// Скрыть ошибку
function hideError() {
    document.getElementById('error').style.display = 'none';
}

// Показать пустое состояние
function showEmptyState() {
    document.getElementById('emptyState').style.display = 'block';
    document.getElementById('mapsTabs').style.display = 'none';
}

// Скрыть пустое состояние
function hideEmptyState() {
    document.getElementById('emptyState').style.display = 'none';
}

// Показать успех
function showSuccess(message) {
    // Можно добавить уведомление
    console.log('✅', message);
}

// Показать информацию о генерации
function showGenerationInfo() {
    if (!currentMaps) return;
    
    const modal = document.getElementById('infoModal');
    const body = document.getElementById('infoModalBody');
    
    const locoList = currentMaps.maps.locomotives.map(loco => 
        `<li>
            <i class="fas fa-train"></i>
            <a href="${loco.url}" target="_blank">${loco.model}-${loco.number} (${loco.trip_count} поездок)</a>
        </li>`
    ).join('');
    
    body.innerHTML = `
        <div class="info-item">
            <span class="info-label">Депо:</span>
            <span class="info-value">${currentMaps.depot_id}</span>
        </div>
        <div class="info-item">
            <span class="info-label">Сгенерировано:</span>
            <span class="info-value">${new Date(currentMaps.generated_at).toLocaleString()}</span>
        </div>
        <div class="info-item">
            <span class="info-label">Карты:</span>
            <span class="info-value">
                <a href="${currentMaps.maps.overview}" target="_blank">Общая</a> | 
                <a href="${currentMaps.maps.heatmap}" target="_blank">Тепловая</a>
            </span>
        </div>
        <h4>Локомотивы (${currentMaps.maps.locomotives.length}):</h4>
        <ul class="maps-list">
            ${locoList}
        </ul>
    `;
    
    modal.classList.add('show');
}

// Закрыть модальное окно
function closeInfoModal() {
    document.getElementById('infoModal').classList.remove('show');
}

// Добавляем кнопку информации в панель управления
function addInfoButton() {
    const controlPanel = document.querySelector('.control-panel');
    const infoBtn = document.createElement('button');
    infoBtn.className = 'info-btn';
    infoBtn.innerHTML = '<i class="fas fa-info-circle"></i>';
    infoBtn.onclick = showGenerationInfo;
    infoBtn.title = 'Информация о генерации';
    controlPanel.appendChild(infoBtn);
}

// Вызываем после успешной генерации
const originalGenerateMaps = generateMaps;
generateMaps = async function() {
    await originalGenerateMaps();
    if (currentMaps) {
        addInfoButton();
    }
};

// Глобальные функции
window.activateTab = activateTab;
window.closeInfoModal = closeInfoModal;
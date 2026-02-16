// frontend/ml/js/app.js

class MLPredictor {
    constructor() {
        this.apiBase = '/api/v1/ml';
        this.currentFile = null;
        this.lastInputData = null;
        this.initElements();
        this.bindEvents();
        this.checkServiceHealth();
    }

    initElements() {
        // Формы
        this.manualForm = document.getElementById('manualForm');
        this.fileInput = document.getElementById('fileInput');
        this.fileUploadArea = document.getElementById('fileUploadArea');
        
        // Результаты
        this.singleResult = document.getElementById('singleResult');
        this.batchResult = document.getElementById('batchResult');
        this.predictionValue = document.getElementById('predictionValue');
        this.resultsBody = document.getElementById('resultsBody');
        this.resultsTable = document.getElementById('resultsTable');
        this.recordCount = document.getElementById('recordCount');
        
        // Статус
        this.serviceStatus = document.getElementById('serviceStatus');
        this.modelInfo = document.getElementById('modelInfo');
        this.loadingIndicator = document.getElementById('loadingIndicator');
        this.errorMessage = document.getElementById('errorMessage');
        this.errorText = document.getElementById('errorText');
        
        // Кнопки
        this.predictButton = document.querySelector('#manualForm button[type="submit"]');
        this.fileButton = document.querySelector('.file-upload-area .btn-secondary');
        
        // ✨ Флаг для предотвращения множественных инициализаций
        this.isInitialized = false;
    }

    bindEvents() {
        // Предотвращаем множественное добавление обработчиков
        if (this.isInitialized) return;
        
        // Ручная форма
        this.manualForm.addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleManualSubmit();
        });

        // ✨ Используем один обработчик для fileInput через делегирование
        document.addEventListener('change', (e) => {
            if (e.target.id === 'fileInput') {
                this.handleFileSelect(e.target.files[0]);
            }
        });

        // Drag & drop с удалением старых классов
        const onDragOver = (e) => {
            e.preventDefault();
            this.fileUploadArea.classList.add('dragover');
        };
        
        const onDragLeave = () => {
            this.fileUploadArea.classList.remove('dragover');
        };
        
        const onDrop = (e) => {
            e.preventDefault();
            this.fileUploadArea.classList.remove('dragover');
            const file = e.dataTransfer.files[0];
            if (file) {
                this.handleFileSelect(file);
            }
        };
        
        // ✨ Удаляем старые обработчики перед добавлением новых
        this.fileUploadArea.removeEventListener('dragover', this._dragOverHandler);
        this.fileUploadArea.removeEventListener('dragleave', this._dragLeaveHandler);
        this.fileUploadArea.removeEventListener('drop', this._dropHandler);
        
        // Сохраняем ссылки на обработчики для возможности удаления
        this._dragOverHandler = onDragOver;
        this._dragLeaveHandler = onDragLeave;
        this._dropHandler = onDrop;
        
        this.fileUploadArea.addEventListener('dragover', onDragOver);
        this.fileUploadArea.addEventListener('dragleave', onDragLeave);
        this.fileUploadArea.addEventListener('drop', onDrop);

        // Валидация полей в реальном времени
        const inputs = document.querySelectorAll('#manualForm input');
        const validationHandler = () => this.validateManualForm();
        
        // ✨ Удаляем старые обработчики
        inputs.forEach(input => {
            input.removeEventListener('input', validationHandler);
            input.addEventListener('input', validationHandler);
        });

        this.isInitialized = true;
    }

    // Валидация формы
    validateManualForm() {
        const series = document.getElementById('locomotiveSeries').value.trim();
        const number = document.getElementById('locomotiveNumber').value;
        const depo = document.getElementById('depo').value.trim();
        const steel = document.getElementById('steelNum').value.trim();
        const mileage = document.getElementById('mileageStart').value;
        
        const isValid = series && number > 0 && depo && steel && mileage >= 0;
        
        if (this.predictButton) {
            this.predictButton.disabled = !isValid;
        }
        return isValid;
    }

    // Проверка здоровья сервиса
    async checkServiceHealth() {
        try {
            this.updateStatus('checking', 'Проверка подключения...');
            
            const response = await fetch(`${this.apiBase}/health`, {
                // ✨ Добавляем заголовки против кэширования
                headers: {
                    'Cache-Control': 'no-cache',
                    'Pragma': 'no-cache'
                }
            });
            
            if (response.ok) {
                this.updateStatus('healthy', 'ML сервис доступен');
                await this.loadModelInfo();
            } else {
                this.updateStatus('unhealthy', 'ML сервис недоступен');
            }
        } catch (error) {
            this.updateStatus('unhealthy', 'Ошибка подключения к ML сервису');
            console.error('Health check failed:', error);
        }
    }

    // Загрузка информации о модели
    async loadModelInfo() {
        try {
            const response = await fetch(`${this.apiBase}/info`, {
                headers: {
                    'Cache-Control': 'no-cache',
                    'Pragma': 'no-cache'
                }
            });
            const info = await response.json();
            
            this.modelInfo.innerHTML = `
                <strong>Модель:</strong> ${info.model_type || 'CatBoost'} | 
                <strong>Признаков:</strong> ${info.features_count || 13} | 
                <strong>Версия:</strong> ${info.version || '1.0.0'}
            `;
        } catch (error) {
            console.error('Failed to load model info:', error);
        }
    }

    // Обновление статуса
    updateStatus(status, message) {
        this.serviceStatus.className = `status-indicator ${status}`;
        this.serviceStatus.innerHTML = `
            <i class="fas fa-circle"></i>
            <span>${message}</span>
        `;
    }

    // Обработка выбора файла
    handleFileSelect(file) {
        if (!file) return;

        // Проверка расширения
        const ext = file.name.split('.').pop().toLowerCase();
        if (!['json', 'jsonl'].includes(ext)) {
            this.showError('Пожалуйста, загрузите файл в формате JSON или JSONL');
            return;
        }

        // Проверка размера (10MB)
        if (file.size > 10 * 1024 * 1024) {
            this.showError('Файл слишком большой. Максимальный размер: 10MB');
            return;
        }

        this.currentFile = file;
        this.updateFileUploadUI();
    }

    // Обновление UI загрузки файла
    updateFileUploadUI() {
        // ✨ Очищаем содержимое и создаем новое
        this.fileUploadArea.innerHTML = '';
        
        if (this.currentFile) {
            const fileSize = (this.currentFile.size / 1024).toFixed(2);
            
            // Создаем элементы через createElement для лучшей производительности
            const icon = document.createElement('i');
            icon.className = 'fas fa-check-circle';
            
            const title = document.createElement('h3');
            title.textContent = 'Файл загружен';
            
            const fileInfo = document.createElement('div');
            fileInfo.className = 'file-info';
            fileInfo.innerHTML = `
                <span class="file-name">
                    <i class="fas fa-file-code"></i>
                    ${this.currentFile.name}
                </span>
                <span class="file-size">${fileSize} KB</span>
                <span class="remove-file">
                    <i class="fas fa-times"></i>
                </span>
            `;
            
            const processButton = document.createElement('button');
            processButton.className = 'btn btn-success';
            processButton.innerHTML = '<i class="fas fa-calculator"></i> Рассчитать износ для файла';
            
            const hint = document.createElement('p');
            hint.className = 'file-hint';
            hint.textContent = 'Нажмите "Рассчитать" для обработки или удалите файл';
            
            this.fileUploadArea.appendChild(icon);
            this.fileUploadArea.appendChild(title);
            this.fileUploadArea.appendChild(fileInfo);
            this.fileUploadArea.appendChild(processButton);
            this.fileUploadArea.appendChild(hint);
            
            // Добавляем обработчики
            const removeBtn = fileInfo.querySelector('.remove-file');
            removeBtn.onclick = (e) => {
                e.stopPropagation();
                this.removeFile();
            };
            
            processButton.onclick = () => this.processFile();
            
            this.fileUploadArea.classList.add('has-file');
        } else {
            // Создаем стандартный UI
            this.fileUploadArea.innerHTML = `
                <i class="fas fa-cloud-upload-alt"></i>
                <h3>Загрузите JSON файл</h3>
                <p>Перетащите файл сюда или кликните для выбора</p>
                <input type="file" id="fileInput" accept=".json,.jsonl" hidden>
                <button class="btn btn-secondary" onclick="document.getElementById('fileInput').click()">
                    <i class="fas fa-folder-open"></i> Выбрать файл
                </button>
                <p class="file-hint">Максимум 1000 записей, формат JSON или JSONL</p>
            `;
            
            this.fileUploadArea.classList.remove('has-file');
            this.fileInput = document.getElementById('fileInput');
        }
    }

    // Удаление файла
    removeFile() {
        this.currentFile = null;
        // ✨ Очищаем значение input
        if (this.fileInput) {
            this.fileInput.value = '';
        }
        this.updateFileUploadUI();
    }

    // Обработка файла
    async processFile() {
        if (!this.currentFile) return;

        const formData = new FormData();
        formData.append('file', this.currentFile);

        this.showLoading();
        this.hideError();

        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 30000); // 30 сек таймаут
            
            const response = await fetch(`${this.apiBase}/upload`, {
                method: 'POST',
                body: formData,
                signal: controller.signal,
                headers: {
                    'Cache-Control': 'no-cache',
                    'Pragma': 'no-cache'
                }
            });

            clearTimeout(timeoutId);

            const data = await response.json();

            if (response.ok) {
                if (data.predictions) {
                    this.lastInputData = data.inputs || null;
                    if (data.predictions.length === 1) {
                        this.showSingleResult(data.predictions[0]);
                    } else {
                        this.showBatchResults(data.predictions, data.count, data.inputs);
                    }
                    this.showSuccess('Файл успешно обработан');
                    
                    // ✨ Не удаляем файл после успешной обработки
                    // Но можно предложить удалить
                }
            } else {
                this.showError(data.error || 'Ошибка при обработке файла');
            }
        } catch (error) {
            if (error.name === 'AbortError') {
                this.showError('Превышено время ожидания ответа от сервера');
            } else {
                this.showError('Ошибка сети: ' + error.message);
            }
        } finally {
            this.hideLoading();
        }
    }

    // Обработка ручного ввода
    async handleManualSubmit() {
        if (!this.validateManualForm()) return;

        const data = [{
            locomotive_series: document.getElementById('locomotiveSeries').value.trim(),
            locomotive_number: parseInt(document.getElementById('locomotiveNumber').value),
            depo: document.getElementById('depo').value.trim(),
            steel_num: document.getElementById('steelNum').value.trim(),
            mileage_start: parseFloat(document.getElementById('mileageStart').value)
        }];

        this.lastInputData = data;
        await this.sendPredictionRequest(data);
    }

    // Отправка запроса на предсказание
    async sendPredictionRequest(data) {
        this.showLoading();
        this.hideError();

        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 30000);
            
            const response = await fetch(`${this.apiBase}/predict`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Cache-Control': 'no-cache',
                    'Pragma': 'no-cache'
                },
                body: JSON.stringify(data),
                signal: controller.signal
            });

            clearTimeout(timeoutId);

            const result = await response.json();

            if (response.ok) {
                if (result.predictions) {
                    if (result.predictions.length === 1) {
                        this.showSingleResult(result.predictions[0]);
                    } else {
                        this.showBatchResults(result.predictions, result.count, data);
                    }
                } else if (Array.isArray(result)) {
                    if (result.length === 1) {
                        this.showSingleResult(result[0]);
                    } else {
                        this.showBatchResults(result, result.length, data);
                    }
                }
            } else {
                this.showError(result.error || 'Ошибка при выполнении запроса');
            }
        } catch (error) {
            if (error.name === 'AbortError') {
                this.showError('Превышено время ожидания ответа от сервера');
            } else {
                this.showError('Ошибка сети: ' + error.message);
            }
        } finally {
            this.hideLoading();
        }
    }

    // Показать результат для одного объекта
    showSingleResult(prediction) {
        this.singleResult.style.display = 'block';
        this.batchResult.style.display = 'none';
        this.predictionValue.textContent = prediction.toFixed(6);
        
        // Добавляем данные из формы в результат
        const series = document.getElementById('locomotiveSeries').value;
        const number = document.getElementById('locomotiveNumber').value;
        const depo = document.getElementById('depo').value;
        const steel = document.getElementById('steelNum').value;
        const mileage = document.getElementById('mileageStart').value;
        
        // Добавляем информацию о входных данных
        const inputInfo = document.createElement('div');
        inputInfo.className = 'input-info';
        inputInfo.style.marginTop = '20px';
        inputInfo.style.padding = '15px';
        inputInfo.style.background = '#f8f9ff';
        inputInfo.style.borderRadius = '10px';
        inputInfo.style.fontSize = '0.9em';
        inputInfo.innerHTML = `
            <h4 style="margin-bottom:10px; color:#555;">Входные данные:</h4>
            <p><strong>Серия:</strong> ${series} | <strong>Номер:</strong> ${number} | 
            <strong>Депо:</strong> ${depo} | <strong>Сталь:</strong> ${steel} | 
            <strong>Пробег:</strong> ${mileage} км</p>
        `;
        
        // Удаляем старую информацию если есть
        const oldInfo = this.singleResult.querySelector('.input-info');
        if (oldInfo) oldInfo.remove();
        
        this.singleResult.appendChild(inputInfo);
        this.singleResult.classList.add('slide-in');
        
        setTimeout(() => {
            this.singleResult.classList.remove('slide-in');
        }, 300);
    }

    // Показать результаты для нескольких объектов
    showBatchResults(predictions, count, inputs = null) {
        this.batchResult.style.display = 'block';
        this.singleResult.style.display = 'none';
        this.recordCount.textContent = `${count} записей`;

        // Заполняем таблицу
        this.resultsBody.innerHTML = '';
        
        predictions.forEach((pred, index) => {
            const row = document.createElement('tr');
            
            if (inputs && inputs[index]) {
                // Если есть исходные данные
                const input = inputs[index];
                row.innerHTML = `
                    <td>${input.locomotive_series || '-'}</td>
                    <td>${input.locomotive_number || '-'}</td>
                    <td>${input.depo || '-'}</td>
                    <td>${input.steel_num || '-'}</td>
                    <td>${input.mileage_start?.toFixed(1) || '-'}</td>
                    <td><strong>${pred.toFixed(6)}</strong></td>
                `;
            } else {
                // Если нет исходных данных
                row.innerHTML = `
                    <td colspan="5" style="text-align:center; color:#999;">
                        Запись ${index + 1} (данные не сохранены)
                    </td>
                    <td><strong>${pred.toFixed(6)}</strong></td>
                `;
            }
            
            this.resultsBody.appendChild(row);
        });

        this.batchResult.classList.add('slide-in');
        setTimeout(() => {
            this.batchResult.classList.remove('slide-in');
        }, 300);
    }

    // Показать успех
    showSuccess(message) {
        console.log('Success:', message);
    }

    // Показать загрузку
    showLoading() {
        this.loadingIndicator.style.display = 'block';
        if (this.predictButton) this.predictButton.disabled = true;
        if (this.fileButton) this.fileButton.disabled = true;
    }

    // Скрыть загрузку
    hideLoading() {
        this.loadingIndicator.style.display = 'none';
        this.validateManualForm();
        if (this.fileButton) this.fileButton.disabled = false;
    }

    // Показать ошибку
    showError(message) {
        this.errorText.textContent = message;
        this.errorMessage.style.display = 'flex';
        
        // Автоматически скрываем через 5 секунд
        setTimeout(() => {
            this.hideError();
        }, 5000);
    }

    // Скрыть ошибку
    hideError() {
        this.errorMessage.style.display = 'none';
    }
}

// Функции для копирования примеров
function copyExample(id) {
    const examples = [
        `{
  "locomotive_series": "VL80",
  "locomotive_number": 123,
  "depo": "Depo1",
  "steel_num": "Steel1",
  "mileage_start": 50000
}`,
        `[
  {
    "locomotive_series": "VL80",
    "locomotive_number": 123,
    "depo": "Depo1",
    "steel_num": "Steel1",
    "mileage_start": 50000
  },
  {
    "locomotive_series": "VL85",
    "locomotive_number": 456,
    "depo": "Depo2",
    "steel_num": "Steel2",
    "mileage_start": 75000
  }
]`,
        `{"locomotive_series":"VL80","locomotive_number":123,"depo":"Depo1","steel_num":"Steel1","mileage_start":50000}
{"locomotive_series":"VL85","locomotive_number":456,"depo":"Depo2","steel_num":"Steel2","mileage_start":75000}`
    ];

    navigator.clipboard.writeText(examples[id - 1]).then(() => {
        alert('✅ Пример скопирован в буфер обмена');
    });
}

// Функции для работы с результатами
function downloadResults() {
    const table = document.getElementById('resultsTable');
    const rows = table.querySelectorAll('tr');
    
    let csv = [];
    rows.forEach(row => {
        const cells = row.querySelectorAll('th, td');
        const rowData = [];
        cells.forEach(cell => {
            rowData.push('"' + cell.textContent + '"');
        });
        csv.push(rowData.join(','));
    });
    
    const blob = new Blob(['\uFEFF' + csv.join('\n')], { type: 'text/csv;charset=utf-8;' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'ml_predictions.csv';
    a.click();
    
    // ✨ Очищаем URL
    setTimeout(() => {
        window.URL.revokeObjectURL(url);
    }, 100);
}

function copyResults() {
    const table = document.getElementById('resultsTable');
    const rows = table.querySelectorAll('tr');
    
    let text = '';
    rows.forEach(row => {
        const cells = row.querySelectorAll('th, td');
        cells.forEach(cell => {
            text += cell.textContent + '\t';
        });
        text += '\n';
    });
    
    navigator.clipboard.writeText(text).then(() => {
        alert('✅ Результаты скопированы в буфер обмена');
    });
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    window.predictor = new MLPredictor();
});
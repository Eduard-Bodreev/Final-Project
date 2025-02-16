# **Финальный проект 1 семестра**
REST API сервис для загрузки и выгрузки данных о ценах.


## **Требования к системе**
**Операционная система:**  
- Ubuntu 22.04 (или совместимые дистрибутивы Linux)  
- Windows 10/11 (через WSL)  

**Необходимые компоненты:**  
- Go `1.21+`  
- PostgreSQL `14+`  
- `curl`, `unzip`, `tar`, `bash` (для работы скриптов)

---

## **⚙ Установка и запуск**
### **1️ Клонирование репозитория**
Сначала склонируйте репозиторий с GitHub:
```bash
git clone https://github.com/Eduard-Bodreev/Final-Project.git
cd Final-Project
```


### **2 Установить зависимости и подготовить базу данных**
```bash
./scripts/prepare.sh
```
Скрипт:
- Устанавливает PostgreSQL (если не установлен)
- Создает базу данных `project_sem_1`
- Создает таблицу `prices`
- Создает пользователя `validator`
- Настраивает права доступа

---

### **3 Запустить сервер**
```bash
./scripts/run.sh
```
После запуска сервер будет доступен по адресу:
```
http://localhost:8080
```
Если порт `8080` уже используется, заверши процесс (`kill -9 <PID>`) или измени порт в `main.go`.

---

## ** API эндпоинты**
### **POST `/api/v0/prices`**
**Описание:** Загружает архив с ценами в базу данных.  
**Формат запроса:**  
```bash
curl -X POST http://localhost:8080/api/v0/prices -F "file=@sample_data.zip"
```
**Формат ответа (`JSON`):**
```json
{
  "total_items": 12,
  "total_categories": 5,
  "total_price": 4740.39
}
```

---

### **GET `/api/v0/prices`**
**Описание:** Выгружает все данные из базы в виде `data.csv` внутри ZIP-архива.  
**Формат запроса:**  
```bash
curl -X GET http://localhost:8080/api/v0/prices -o output.zip
```
**Файл `data.csv` (пример содержимого):**
```csv
id,created_at,name,category,price
1,2024-01-01,iPhone 13,Electronics,799.99
2,2024-01-02,Nike Air Max,Shoes,129.99
...
```

---

## **Тестирование**
### **Запуск тестов**
```bash
./scripts/tests.sh 1
```
### **Что проверяется:**
**POST** `/api/v0/prices`: загрузка ZIP-файла с `data.csv`  
**GET** `/api/v0/prices`: скачивание ZIP-архива с `data.csv`  
**PostgreSQL**: правильность сохранения данных  

Если все тесты пройдены, скрипт выведет:
```
✓ Все проверки пройдены успешно
```

---

## **Директория `sample_data`**
Пример данных для тестирования.  
Файл `sample_data.zip` содержит:
- `data.csv` — набор цен в формате:
```csv
id,name,category,price,create_date
1,iPhone 13,Electronics,799.99,2024-01-01
...
```

---

## **Контакты**
Если у вас есть вопросы по проекту, свяжитесь со мной:
📩 Telegram: @AdDadaya

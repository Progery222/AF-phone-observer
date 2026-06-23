```markdown
# phone-observer — полное описание сервиса

## Что это?

Сервис, который **смотрит на экран телефона** и предоставляет данные о том, что там происходит. Это **глаза** всей системы.

---

## Взаимодействие с другими сервисами

```
┌─────────────────────────────────────────────────────────────────┐
│                       Другие сервисы                           │
├─────────────────────────────────────────────────────────────────┤
│  Оркестратор:      "Что на экране?"                           │
│  Recovery-Engine:  "Дай скриншот и UI-иерархию"              │
│  Executor:         "Найди координаты кнопки"                  │
│  Analytics:        "Собери данные о текущем экране"          │
└─────────────────────────────┬───────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      phone-observer                            │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────────┐ │
│  │  Screenshot  │  │   UI Dump    │  │  Element Finder     │ │
│  │  (скриншот)  │  │ (иерархия)   │  │  (поиск элементов) │ │
│  └──────────────┘  └──────────────┘  └─────────────────────┘ │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────────┐ │
│  │  Wait For    │  │ Detect State │  │  OCR / Vision       │ │
│  │  (ожидание)  │  │ (состояние)  │  │  (распознавание)   │ │
│  └──────────────┘  └──────────────┘  └─────────────────────┘ │
└─────────────────────────────┬───────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Телефон (Android)                         │
│                                                                 │
│  - adb shell screencap    (скриншот)                          │
│  - adb shell uiautomator  (UI-иерархия)                      │
│  - adb shell dumpsys      (системная информация)              │
└─────────────────────────────────────────────────────────────────┘
```

---

## Что делает Observer

### 1. Делает скриншот

**Команда:**
```
Оркестратор → Observer: "Телефон #7, покажи, что у тебя на экране"
```

**Что делает Observer:**
- Запускает `adb shell screencap /sdcard/screen.png`
- Скачивает файл на сервер
- Сохраняет в MinIO (облачное хранилище)
- Возвращает ссылку

**Ответ:**
```json
{
  "screenshot_url": "https://minio.example.com/screenshots/phone7_2026-06-22.jpg",
  "minio_key": "screenshots/phone7_2026-06-22.jpg",
  "size_bytes": 245760,
  "resolution": {"width": 1080, "height": 1920},
  "taken_at": "2026-06-22T10:00:00Z"
}
```

---

### 2. Собирает UI-иерархию (XML)

**Команда:**
```
Оркестратор → Observer: "Телефон #7, расскажи, какие кнопки на экране"
```

**Что делает Observer:**
- Запускает `adb shell uiautomator dump`
- Получает XML-дерево всех элементов
- Парсит и возвращает структурированные данные

**Ответ (XML):**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<hierarchy>
  <node text="Добро пожаловать!" bounds="[100,50][700,150]" class="TextView"/>
  <EditText hint="Email" bounds="[100,200][700,300]" resource-id="com.app:id/email"/>
  <EditText hint="Пароль" bounds="[100,330][700,430]" resource-id="com.app:id/password"/>
  <Button text="Войти" bounds="[200,500][600,580]" resource-id="com.app:id/login"/>
  <Button text="Зарегистрироваться" bounds="[200,600][600,680]"/>
</hierarchy>
```

**Ответ (JSON):**
```json
{
  "elements": [
    {
      "type": "TextView",
      "text": "Добро пожаловать!",
      "bounds": {"x1": 100, "y1": 50, "x2": 700, "y2": 150},
      "center": {"x": 400, "y": 100}
    },
    {
      "type": "EditText",
      "hint": "Email",
      "resource_id": "com.app:id/email",
      "bounds": {"x1": 100, "y1": 200, "x2": 700, "y2": 300},
      "center": {"x": 400, "y": 250}
    },
    {
      "type": "Button",
      "text": "Войти",
      "resource_id": "com.app:id/login",
      "bounds": {"x1": 200, "y1": 500, "x2": 600, "y2": 580},
      "center": {"x": 400, "y": 540}
    }
  ]
}
```

---

### 3. Ищет конкретный элемент на экране

**Команда:**
```
Оркестратор → Observer: "Телефон #7, найди кнопку 'Войти'"
```

**Что делает Observer:**
- Ищет в UI-иерархии по тексту или resource-id
- Возвращает координаты и информацию об элементе

**Ответ:**
```json
{
  "found": true,
  "element": {
    "type": "Button",
    "text": "Войти",
    "resource_id": "com.app:id/login",
    "bounds": {"x1": 200, "y1": 500, "x2": 600, "y2": 580},
    "center": {"x": 400, "y": 540}
  },
  "found_by": "text"
}
```

---

### 4. Ждёт появления элемента

**Команда:**
```
Оркестратор → Observer: "Телефон #7, жди, пока не появится кнопка 'Далее'"
```

**Что делает Observer:**
- Каждые 0.5-1 секунду проверяет UI-иерархию
- Ищет элемент по тексту или resource-id
- Если появился → возвращает координаты
- Если таймаут (30 сек) → ошибка

**Ответ (элемент найден):**
```json
{
  "found": true,
  "element": {
    "type": "Button",
    "text": "Далее",
    "center": {"x": 500, "y": 400}
  },
  "wait_time_ms": 3200,
  "check_count": 7
}
```

**Ответ (таймаут):**
```json
{
  "found": false,
  "timeout_sec": 30,
  "reason": "Элемент не появился за 30 секунд"
}
```

---

### 5. Определяет состояние экрана

**Команда:**
```
Оркестратор → Observer: "Телефон #7, определи, где мы сейчас"
```

**Что делает Observer:**
- Анализирует UI-иерархию и скриншот
- Распознаёт типичные экраны

**Ответ:**
```json
{
  "state": "login_screen",
  "confidence": 0.95,
  "screenshot_url": "https://minio/phone7_2026-06-22.jpg",
  "elements": {
    "email_field": true,
    "password_field": true,
    "login_button": true
  },
  "description": "Экран входа: поля Email, Пароль, кнопка Войти"
}
```

**Типичные состояния:**

| Что видит | Определение |
|-----------|-------------|
| Поле ввода логина + пароля | `login_screen` |
| Кнопка "Разрешить" + запрос прав | `permission_request` |
| Реклама на весь экран + крестик | `ads_fullscreen` |
| Кнопка "Установить" | `install_screen` |
| Пустой экран + логотип | `loading` |
| Лента с постами | `main_feed` |
| Системное уведомление | `notification` |

---

## API Observer

### Основные методы

| Метод | Описание |
|-------|----------|
| `POST /screenshot` | Сделать скриншот |
| `POST /dump-ui` | Собрать UI-иерархию |
| `POST /find-element` | Найти элемент |
| `POST /wait-for-element` | Ждать появления элемента |
| `POST /detect-state` | Определить состояние |
| `GET /screen/{serial}` | Получить текущий скриншот |
| `GET /ui/{serial}` | Получить текущую UI-иерархию |
| `DELETE /cache/{serial}` | Очистить кеш экрана |

---

### Детали методов

#### POST /screenshot

**Запрос:**
```json
{
  "serial": "phone_7",
  "store_in_minio": true,
  "timeout_sec": 10
}
```

**Ответ:**
```json
{
  "screenshot_url": "https://minio/screenshots/phone7_2026-06-22.jpg",
  "minio_key": "screenshots/phone7_2026-06-22.jpg",
  "size_bytes": 245760,
  "resolution": {"width": 1080, "height": 1920},
  "taken_at": "2026-06-22T10:00:00Z"
}
```

---

#### POST /dump-ui

**Запрос:**
```json
{
  "serial": "phone_7",
  "format": "json",
  "timeout_sec": 10
}
```

**Ответ:**
```json
{
  "elements": [
    {
      "type": "Button",
      "text": "Войти",
      "resource_id": "com.app:id/login",
      "center": {"x": 400, "y": 540},
      "bounds": {"x1": 200, "y1": 500, "x2": 600, "y2": 580}
    }
  ],
  "element_count": 15,
  "taken_at": "2026-06-22T10:00:00Z"
}
```

---

#### POST /find-element

**Запрос:**
```json
{
  "serial": "phone_7",
  "element": {
    "type": "button",
    "text": "Войти",
    "resource_id": "com.app:id/login"
  },
  "timeout_sec": 10
}
```

**Ответ:**
```json
{
  "found": true,
  "element": {
    "type": "Button",
    "text": "Войти",
    "resource_id": "com.app:id/login",
    "bounds": {"x1": 200, "y1": 500, "x2": 600, "y2": 580},
    "center": {"x": 400, "y": 540}
  },
  "found_by": "text"
}
```

---

#### POST /wait-for-element

**Запрос:**
```json
{
  "serial": "phone_7",
  "element": {
    "text": "Далее"
  },
  "timeout_sec": 30,
  "check_interval_ms": 500
}
```

**Ответ:**
```json
{
  "found": true,
  "element": {
    "center": {"x": 500, "y": 400}
  },
  "wait_time_ms": 3200,
  "check_count": 7
}
```

---

#### POST /detect-state

**Запрос:**
```json
{
  "serial": "phone_7",
  "use_screenshot": true
}
```

**Ответ:**
```json
{
  "state": "login_screen",
  "confidence": 0.95,
  "screenshot_url": "https://minio/phone7_2026-06-22.jpg",
  "description": "Экран входа: поля Email, Пароль, кнопка Войти"
}
```

---

## Пример использования в сценарии

### Сценарий: вход в Instagram

```
1. Оркестратор → Executor: "Открой Instagram"
   Executor: Нажимает на иконку
   ↓
2. Оркестратор → Observer: "Что на экране?"
   Observer: Делает скриншот
   Observer: Собирает UI-иерархию
   Observer: {"state": "login_screen", "elements": ["Email", "Пароль", "Войти"]}
   ↓
3. Оркестратор → Observer: "Найди поле Email"
   Observer: {"center": {"x": 400, "y": 250}}
   ↓
4. Оркестратор → Executor: "Нажми на (400, 250)"
   Executor: ✅ "Готово"
   ↓
5. Оркестратор → Executor: "Введи 'phone7@gmail.com'"
   Executor: ✅ "Готово"
   ↓
6. Оркестратор → Observer: "Найди поле Пароль"
   Observer: {"center": {"x": 400, "y": 380}}
   ↓
7. Оркестратор → Executor: "Нажми на (400, 380)"
   Executor: ✅ "Готово"
   ↓
8. Оркестратор → Executor: "Введи 'MyPass123'"
   Executor: ✅ "Готово"
   ↓
9. Оркестратор → Observer: "Найди кнопку 'Войти'"
   Observer: {"center": {"x": 400, "y": 540}}
   ↓
10. Оркестратор → Executor: "Нажми на (400, 540)"
    Executor: ✅ "Готово"
    ↓
11. Оркестратор → Observer: "Что на экране?"
    Observer: {"state": "main_feed", "description": "Лента Instagram"}
    ↓
12. Оркестратор: ✅ "Вход выполнен!"
```

---

## Почему Observer — отдельный микросервис

### 1. Работает напрямую с Android
- Делает скриншоты (`screencap`)
- Собирает UI-иерархию (`uiautomator dump`)
- Использует системные команды (`dumpsys`)
- Это отдельный слой, который может меняться без затрагивания других сервисов

### 2. Нужен разным сервисам
- **Оркестратору** — чтобы понимать, что происходит
- **Recovery-Engine** — чтобы анализировать ошибки
- **Analytics** — чтобы собирать данные
- **Executor** — может запрашивать координаты

### 3. Можно заменить реализацию
```
Сегодня: uiautomator (чистый Android)
Завтра: Appium (кроссплатформенный)
Послезавтра: Airtest (для игр)
```
Observer скрывает эти детали от всех остальных

### 4. Кеширование и оптимизация
- Кеширует UI-иерархию на 1-2 секунды
- Не делает скриншот, если не изменился экран
- Оптимизирует запросы к Android

---

## Структура данных

```go
type Screenshot struct {
    Serial      string    `json:"serial"`
    URL         string    `json:"url"`
    MinIOKey    string    `json:"minio_key"`
    SizeBytes   int       `json:"size_bytes"`
    Width       int       `json:"width"`
    Height      int       `json:"height"`
    TakenAt     time.Time `json:"taken_at"`
}

type UIElement struct {
    Type       string `json:"type"`        // Button, EditText, TextView
    Text       string `json:"text"`        // видимый текст
    ResourceID string `json:"resource_id"` // android:id
    Hint       string `json:"hint"`        // placeholder
    Bounds     Bounds `json:"bounds"`      // координаты
    Center     Point  `json:"center"`      // центр
    Children   []UIElement `json:"children"`
}

type Bounds struct {
    X1 int `json:"x1"`
    Y1 int `json:"y1"`
    X2 int `json:"x2"`
    Y2 int `json:"y2"`
}

type Point struct {
    X int `json:"x"`
    Y int `json:"y"`
}

type ScreenState struct {
    State       string   `json:"state"`        // login_screen, ads_fullscreen, main_feed
    Confidence  float64  `json:"confidence"`
    Elements    []string `json:"elements"`     // список найденных элементов
    Description string   `json:"description"`
    Screenshot  string   `json:"screenshot"`
}
```

---

## Итог

**Observer — это глаза телефона:**

1. **Смотрит** на экран (скриншот, UI-иерархия)
2. **Находит** кнопки и элементы
3. **Ждёт**, когда появится нужный элемент
4. **Определяет**, какой сейчас экран
5. **Отдаёт** данные другим сервисам

**Оркестратор говорит:** "Что там?"  
**Observer отвечает:** "Вот кнопка 'Войти' на (400, 540), нажимай сюда."
```
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
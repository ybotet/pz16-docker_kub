# Практическое занятие №15: Деплой приложения на VPS. Настройка systemd

## Описание

Цель данной практической работы — научиться публиковать сервис на удалённой машине (VPS) и управлять им через systemd. В рамках работы сервис `tasks` разворачивается в среде Linux (WSL2), настраивается как systemd-сервис с автоматическим запуском и перезапуском, а также подключается к необходимым зависимостям (PostgreSQL, Redis, RabbitMQ).

---

## Структура проекта

![alt text](<public/Снимок экрана 2026-04-12 235207.png>)

---

## Инфраструктура

### Среда выполнения

В качестве "VPS" используется WSL2 (Windows Subsystem for Linux) с Ubuntu 22.04. Это обеспечивает полноценную среду Linux с поддержкой systemd.

**Проверка окружения:**
```bash
hostname -I
# 172.20.240.123
```

## Установленные компоненты

| Компонент | Версия | Назначение |
|---|---:|---|
| Ubuntu | 22.04 | Операционная система |
| systemd | 249 | Системный менеджер |
| PostgreSQL | 14.22 | База данных |
| Redis | 7.0.8 | Кэш/хранилище |
| RabbitMQ | 3.9.13 | Брокер сообщений |
| Go binary | - | Сервис `tasks` |

## Структура директорий

```bash
/opt/tasks/
├── bin/
│   └── tasks              # Бинарный файл сервиса
└── (рабочая директория)

/etc/tasks/
└── tasks.env              # Переменные окружения

/etc/systemd/system/
└── tasks.service          # systemd unit-файл

```

###
```bash
# Создание директорий
sudo mkdir -p /opt/tasks/bin
sudo mkdir -p /etc/tasks

# Создание пользователя для сервиса
sudo useradd --system --no-create-home --shell /usr/sbin/nologin tasksuser

# Установка прав
sudo chown -R tasksuser:tasksuser /opt/tasks
```

## Переменные окружения

Файл `/etc/tasks/tasks.env` содержит конфигурацию сервиса:

```bash
# Сервер
REST_PORT=8082

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=tasksuser
DB_PASSWORD=taskspass
DB_NAME=tasksdb

# Redis
REDIS_ADDR=localhost:6379

# RabbitMQ
RABBIT_URL=amqp://guest:guest@localhost:5672/
QUEUE_NAME=task_events

# Логирование
LOG_LEVEL=info
```

### Права доступа

```bash
sudo chmod 600 /etc/tasks/tasks.env   # Только для чтения root
sudo chown root:root /etc/tasks/tasks.env
```

## Systemd Unit-файл

Файл `/etc/systemd/system/tasks.service`:

```ini
[Unit]
Description=Tasks Service (REST + GraphQL)
After=network.target postgresql.service redis-server.service rabbitmq-server.service
Wants=postgresql.service redis-server.service rabbitmq-server.service

[Service]
Type=simple
User=tasksuser
Group=tasksuser
WorkingDirectory=/opt/tasks
EnvironmentFile=/etc/tasks/tasks.env
ExecStart=/opt/tasks/bin/tasks
Restart=always
RestartSec=5
NoNewPrivileges=true
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

### Объяснение ключевых параметров

| Параметр | Значение | Описание |
|---|---|---|
| `After=` | `network.target postgresql.service ...` | Сервис запускается после указанных зависимостей. |
| `Wants=` | `postgresql.service ...` | Мягкая зависимость: systemd попробует запустить эти сервисы. |
| `User=` | `tasksuser` | Запуск от непривилегированного пользователя для безопасности. |
| `EnvironmentFile=` | `/etc/tasks/tasks.env` | Загрузка переменных окружения из файла. |
| `Restart=` | `always` | Автоматический перезапуск при падении. |
| `RestartSec=` | `5` | Пауза 5 секунд перед перезапуском. |
| `NoNewPrivileges=` | `true` | Запрещает повышение привилегий. |
| `LimitNOFILE=` | `65535` | Максимум открытых файловых дескрипторов. |

## Управление сервисом

### Основные команды

```bash
# Перезагрузка конфигурации systemd
sudo systemctl daemon-reload

# Запуск сервиса
sudo systemctl start tasks

# Остановка сервиса
sudo systemctl stop tasks

# Перезапуск сервиса
sudo systemctl restart tasks

# Проверка статуса
sudo systemctl status tasks

# Включение автозапуска при загрузке системы
sudo systemctl enable tasks

# Отключение автозапуска
sudo systemctl disable tasks
```

### Статус сервиса

```bash
sudo systemctl status tasks
```

![alt text](<public/Снимок экрана 2026-04-13 030535.png>)

## Логирование (journalctl)
### Просмотр логов

![alt text](<public/Снимок экрана 2026-04-13 030808.png>)

## Проверка работоспособности
### Health check

![alt text](<public/Снимок экрана 2026-04-13 031038.png>) 

![alt text](<public/Снимок экрана 2026-04-13 031111.png>) 

![alt text](<public/Снимок экрана 2026-04-13 031119.png>)

## Обновление сервиса
```bash
# 1. Остановить сервис
sudo systemctl stop tasks

# 2. Сделать резервную копию текущего бинарника
sudo cp /opt/tasks/bin/tasks /opt/tasks/bin/tasks.backup

# 3. Скопировать новую версию
sudo cp /tmp/tasks /opt/tasks/bin/tasks

# 4. Установить права
sudo chmod 755 /opt/tasks/bin/tasks
sudo chown tasksuser:tasksuser /opt/tasks/bin/tasks

# 5. Запустить сервис
sudo systemctl start tasks

# 6. Проверить статус
sudo systemctl status tasks
```

## База данных PostgreSQL

### Создание таблицы

```sql
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    done BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### Права доступа

```sql
GRANT ALL PRIVILEGES ON TABLE tasks TO tasksuser;
GRANT ALL PRIVILEGES ON SCHEMA public TO tasksuser;
```


## Контрольные вопросы

1. Зачем нужен systemd и чем он лучше "запуска в screen/tmux"?

systemd — это системный менеджер (PID 1), который управляет сервисами на уровне ОС. В отличие от screen/tmux:

- Автоматический перезапуск при падении (Restart=always)
- Автозапуск при загрузке системы (systemctl enable)
- Централизованные логи (journalctl)
- Управление зависимостями (After=, Wants=)
- Безопасность (запуск от непривилегированного пользователя)

2. Почему не стоит запускать сервис от root?

По соображениям безопасности. Если сервис скомпрометирован, злоумышленник получит полный контроль над системой. Запуск от выделенного пользователя (tasksuser) ограничивает потенциальный ущерб.

3. Зачем хранить env-конфиг в /etc/, а не в репозитории?

- Разделение кода и конфигурации (12-factor app)
- Защита секретов (пароли, токены) — они не попадают в репозиторий
- Возможность изменять конфигурацию без перекомпиляции
- Разные конфигурации для разных сред (dev/staging/prod)

4. Как посмотреть логи сервиса, если он упал?

```bash

# Последние 50 строк логов
sudo journalctl -u tasks -n 50 --no-pager

# Логи с момента последнего запуска
sudo journalctl -u tasks --since "5 minutes ago"

# Логи в реальном времени
sudo journalctl -u tasks -f
```

5. Что даёт Restart=always и RestartSec?

- Restart=always — systemd автоматически перезапускает сервис, если процесс завершился (упал)
- RestartSec=5 — задержка в 5 секунд перед попыткой перезапуска, предотвращая бесконечный цикл быстрых перезапусков#   p z 1 6 - d o c k e r _ k u b  
 #   p z 1 6 - d o c k e r _ k u b  
 
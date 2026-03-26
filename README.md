# PZ7-DOCKER-GO

## Оглавление

1. [Описание проекта](#1-описание-проекта)
2. [Структура проекта](#2-структура-проекта)
3. [Dockerfile для сервиса auth](#3-dockerfile-для-сервиса-auth)
   - 3.1. Содержимое Dockerfile
   - 3.2. Пояснение стадий сборки
4. [Dockerfile для сервиса tasks](#4-dockerfile-для-сервиса-tasks)
   - 4.1. Содержимое Dockerfile
   - 4.2. Пояснение стадий сборки
5. [.dockerignore](#5-dockerignore)
6. [Docker Compose](#6-docker-compose)
   - 6.1. Полный файл docker-compose.yml
   - 6.2. Описание сервисов
   - 6.3. Сеть Docker
7. [Переменные окружения](#7-переменные-окружения)
8. [Команды для сборки и запуска](#8-команды-для-сборки-и-запуска)
9. [Проверка работоспособности](#9-проверка-работоспособности)
10. [Контрольные вопросы](#11-контрольные-вопросы)

---

## 1. Описание проекта

Данный проект представляет собой контейнеризацию двух микросервисов — **auth** (сервис аутентификации) и **tasks** (сервис управления задачами), а также вспомогательных компонентов: базы данных PostgreSQL и прокси-сервера nginx с поддержкой HTTPS.

Проект демонстрирует:

- Использование multi-stage сборки для Go-приложений
- Организацию взаимодействия между контейнерами в единой сети
- Настройку reverse proxy с SSL-терминацией
- Управление конфигурацией через переменные окружения
- Запуск всей инфраструктуры с помощью Docker Compose

**Технологии:** Docker, Docker Compose, Go, PostgreSQL, nginx, gRPC, JWT.

---

## 2. Структура проекта

![alt text](<public/Снимок экрана 2026-03-27 004752.png>)

---

## 3. Dockerfile для сервиса auth

### 3.1. Содержимое Dockerfile

**Путь:** `services/auth/Dockerfile`

![alt text](<public/Снимок экрана 2026-03-27 005159.png>)

### 3.2. Пояснение стадий сборки

| Стадия      | Образ                | Назначение                                                                                                                |
| ----------- | -------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| **builder** | `golang:1.25-alpine` | Содержит компилятор Go и все инструменты для сборки. Здесь происходит загрузка зависимостей и компиляция бинарного файла. |
| **runner**  | `alpine:latest`      | Минимальный образ Linux (~5 МБ). Содержит только скомпилированный бинарник и CA-сертификаты.                              |

**Преимущества multi-stage build:**

- Итоговый образ занимает ~15 МБ вместо ~300 МБ
- В финальный образ не попадают исходные коды и инструменты компиляции
- Повышается безопасность за счет уменьшения векторов атак
- Ускоряется сборка благодаря кэшированию зависимостей

---

## 4. Dockerfile для сервиса tasks

### 4.1. Содержимое Dockerfile

**Путь:** `services/tasks/Dockerfile`

![alt text](<public/Снимок экрана 2026-03-27 005816.png>)

### 4.2. Пояснение стадий сборки

Аналогично сервису auth, используется двухстадийная сборка. Это обеспечивает:

- Единообразие процесса сборки для всех Go-сервисов
- Возможность переиспользования кэша зависимостей
- Минимальный размер финального образа
- Независимость от версий Go на хост-машине

## 5. .dockerignore

**Путь:** `services/auth/.dockerignore и services/tasks/.dockerignore`

```text
.git
*.log
bin/
tmp/
*.exe
*.test
coverage.txt
.idea/
.vscode/
```

Назначение: исключает из контекста сборки ненужные файлы (логи, бинарники, служебные папки), что ускоряет сборку и уменьшает размер контекста, передаваемого демону Docker. Также предотвращает случайное попадание в образ чувствительных данных.

## 6. Docker Compose

### 6.1. Полный файл docker-compose.yml

**Путь:** `deploy/docker-compose.yml`

![alt text](<public/Снимок экрана 2026-03-27 010311.png>)

![alt text](<public/Снимок экрана 2026-03-27 010321.png>)

![alt text](<public/Снимок экрана 2026-03-27 010327.png>)

![alt text](<public/Снимок экрана 2026-03-27 010337.png>)

## 6.2. Descripción de servicios

| Servicio | Imagen                      | Propósito                                      | Puertos                   |
| -------- | --------------------------- | ---------------------------------------------- | ------------------------- |
| postgres | postgres:15                 | Base de datos para almacenar usuarios y tareas | 5432                      |
| auth     | construido desde Dockerfile | Servicio de autenticación (JWT, gRPC)          | 8081 (HTTP), 50051 (gRPC) |
| tasks    | construido desde Dockerfile | Servicio de gestión de tareas                  | 8082                      |
| nginx    | nginx:alpine                | Reverse proxy con HTTPS                        | 8443                      |

## 6.3. Red Docker

Todos los servicios están unidos en una red personalizada `pz7-network` con driver `bridge`. Esto permite que los contenedores se comuniquen entre sí por el nombre del servicio, en lugar de por dirección IP. Docker Compose configura automáticamente un servidor DNS integrado que resuelve los nombres de los servicios en las direcciones IP de los contenedores correspondientes.

## 7. Variables de entorno

### Servicio auth

| Variable         | Valor                                | Propósito                     |
| ---------------- | ------------------------------------ | ----------------------------- |
| `AUTH_HTTP_PORT` | 8081                                 | Puerto para solicitudes HTTP  |
| `AUTH_GRPC_PORT` | 50051                                | Puerto para solicitudes gRPC  |
| `JWT_SECRET`     | your-secret-key-change-in-production | Clave secreta para firmar JWT |

### Servicio tasks

| Variable         | Valor      | Propósito                                                     |
| ---------------- | ---------- | ------------------------------------------------------------- |
| `AUTH_GRPC_ADDR` | auth:50051 | Dirección del servidor gRPC de auth (por nombre del servicio) |
| `TASKS_PORT`     | 8082       | Puerto para solicitudes HTTP                                  |
| `DB_HOST`        | postgres   | Host de la base de datos (por nombre del servicio)            |
| `DB_PORT`        | 5432       | Puerto de PostgreSQL                                          |
| `DB_USER`        | tasksuser  | Usuario de la base de datos                                   |
| `DB_PASSWORD`    | taskspass  | Contraseña de la base de datos                                |
| `DB_NAME`        | tasksdb    | Nombre de la base de datos                                    |
| `DB_SSLMODE`     | disable    | Deshabilitar SSL para la conexión a la base de datos          |

## 8. Команды для сборки и запуска

### Сборка всех образов и запуск контейнеров:

```bash

cd deploy
docker-compose up -d --build
```

### Остановка всех контейнеров:

```bash
docker-compose down
```

### Просмотр статуса контейнеров:

```bash

docker-compose ps
```

### Просмотр логов:

```bash

# Логи всех сервисов
docker-compose logs -f

# Логи конкретного сервиса
docker-compose logs auth --tail 20
docker-compose logs tasks --tail 20
docker-compose logs nginx --tail 20
```

## 9. Проверка работоспособности

![alt text](<public/Снимок экрана 2026-03-16 030221.png>)

![alt text](<public/Снимок экрана 2026-03-16 035442.png>)

---

## 10. Контрольные вопросы

1. Чем отличается Docker image от container?

Docker image — это неизменяемый шаблон (образ), содержащий всё необходимое для запуска приложения: операционную систему, зависимости, код и настройки. Он доступен только для чтения и может храниться в реестре (Docker Hub). Docker container — это запущенный экземпляр образа. Контейнер имеет собственное изолированное окружение, файловую систему (слой записи поверх образа) и существует только во время выполнения. Один образ может породить множество контейнеров.

2. Зачем нужен multi-stage build?

Multi-stage build позволяет создавать небольшие и безопасные образы. В первой стадии (builder) используются все инструменты для компиляции (Go, компилятор, зависимости), а во второй (runner) копируется только скомпилированный бинарник в минимальный образ (Alpine). Это уменьшает размер финального образа с сотен мегабайт до 10-20 МБ, исключает из финального образа исходный код и утилиты сборки, что повышает безопасность, и ускоряет развертывание за счет меньшего объема данных для передачи.

3. Почему нельзя хранить секреты внутри Dockerfile?

Секреты (JWT_SECRET, пароли БД, API-ключи) нельзя хранить в Dockerfile по нескольким причинам:

- Dockerfile попадает в систему контроля версий (Git) и становится доступен всем разработчикам и в истории коммитов
- Секреты становятся частью образа и могут быть извлечены командой docker history --no-trunc
- Невозможно использовать разные секреты для разных окружений (dev, staging, production) без пересборки образа

Правильный подход — передавать секреты через переменные окружения (как в docker-compose.yml) или использовать Docker Secrets в Swarm/Kubernetes.

4. Почему внутри docker-сети нельзя обращаться к другому контейнеру через localhost?

В Docker каждый контейнер имеет собственное сетевое пространство (network namespace). localhost (или 127.0.0.1) внутри контейнера ссылается только на сам этот контейнер, а не на хост-машину и не на другие контейнеры. Чтобы контейнеры могли общаться друг с другом, их нужно подключить к одной пользовательской сети и обращаться по именам сервисов (например, auth:50051). Docker Compose автоматически настраивает DNS-резолюцию между сервисами в одной сети, поэтому имена сервисов резолвятся в актуальные IP-адреса контейнеров.

5. Зачем нужен .dockerignore?

.dockerignore исключает ненужные файлы и папки из контекста сборки, который передается демону Docker при выполнении docker build. Это:

- Ускоряет сборку за счет уменьшения объема данных, передаваемых демону
- Уменьшает размер контекста сборки
- Предотвращает случайное копирование в образ чувствительных данных (например, .git, локальных логов, бинарников, файлов с секретами)
- Уменьшает вероятность ошибок, связанных с включением ненужных файлов в образ

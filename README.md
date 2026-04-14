# Практическое занятие №16: Публикация приложения в Kubernetes

## Цель работы

Цель данной практической работы — научиться контейнеризировать приложение и разворачивать его в Kubernetes с использованием минимальных манифестов (Deployment, Service, ConfigMap).

---

## Структура проекта

![alt text](<public/Снимок экрана 2026-04-14 032937.png>)

## Используемый Kubernetes стенд

Для выполнения работы использовался **Kind (Kubernetes in Docker)** — инструмент для запуска локального кластера Kubernetes внутри Docker-контейнеров.

**Создание кластера:**
```bash
kind create cluster --name tasks-cluster --config deploy/k8s/kind-config.yaml
```

**Конфигурация кластера (kind-config.yaml):**
```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
```

**Проверка доступа:**
```bash
kubectl cluster-info --context kind-tasks-cluster
kubectl get nodes
``` 

![alt text](<public/Снимок экрана 2026-04-14 031120.png>)

## Подготовка Docker-образа

### Dockerfile

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o tasks ./services/task/cmd/task

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/tasks .

EXPOSE 8082

CMD ["./tasks"]
```

### Сборка образа

```bash
docker build -t techip-tasks:0.1 .
```

### Загрузка образа в Kind

```bash
kind load docker-image techip-tasks:0.1 --name tasks-cluster
```

![alt text](<public/Снимок экрана 2026-04-14 031713.png>)

## Манифесты Kubernetes

### ConfigMap (`configmap.yaml`)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tasks-config
data:
  REST_PORT: "8082"
  LOG_LEVEL: "info"
  DB_HOST: "postgres"
  DB_PORT: "5432"
  DB_USER: "tasksuser"
  DB_PASSWORD: "taskspass"
  DB_NAME: "tasksdb"
  REDIS_ADDR: "redis:6379"
  RABBIT_URL: "amqp://guest:guest@rabbitmq:5672/"
  QUEUE_NAME: "task_events"
```

### Deployment (`deployment.yaml`)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tasks
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tasks
  template:
    metadata:
      labels:
        app: tasks
    spec:
      containers:
      - name: tasks
        image: techip-tasks:0.1
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8082
        envFrom:
        - configMapRef:
            name: tasks-config
        readinessProbe:
          httpGet:
            path: /health
            port: 8082
          initialDelaySeconds: 5
          periodSeconds: 10
        livenessProbe:
          httpGet:
            path: /health
            port: 8082
          initialDelaySeconds: 15
          periodSeconds: 20
```

### Service (`service.yaml`)

```yaml
apiVersion: v1
kind: Service
metadata:
  name: tasks
spec:
  selector:
    app: tasks
  ports:
  - port: 8082
    targetPort: 8082
  type: ClusterIP
```

### Деплой зависимостей

Были созданы отдельные манифесты для каждого сервиса с `Deployment` и `Service`.

## Применение манифестов

```bash
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/postgres.yaml
kubectl apply -f deploy/k8s/redis.yaml
kubectl apply -f deploy/k8s/rabbitmq.yaml
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml
```

## Проверка состояния

![alt text](<public/Снимок экрана 2026-04-14 032341.png>)

![alt text](<public/Снимок экрана 2026-04-14 024954.png>)

![alt text](<public/Снимок экрана 2026-04-14 024914.png>) 

![alt text](<public/Снимок экрана 2026-04-14 024908.png>) 


## Ответы на контрольные вопросы

1. Чем Pod отличается от Deployment?

Pod — это минимальная единица в Kubernetes, содержащая один или несколько контейнеров. Deployment — это контроллер, который управляет Pod'ами: обеспечивает нужное количество реплик, выполняет обновления и откаты.

2. Зачем нужен Service и почему нельзя "ходить прямо в Pod"?

Pod'ы в Kubernetes эфемерны — они могут создаваться, удаляться и менять IP-адреса. Service предоставляет стабильный IP-адрес и DNS-имя, а также балансирует нагрузку между несколькими Pod'ами.

3. Чем readiness probe отличается от liveness probe?

- Readiness probe проверяет, готов ли Pod принимать трафик. Если проверка не проходит, Pod исключается из Service.
- Liveness probe проверяет, жив ли контейнер. Если проверка не проходит, Kubernetes перезапускает контейнер.

4. Зачем нужен ConfigMap и чем он отличается от Secret?

ConfigMap хранит несекретные данные конфигурации (переменные окружения, параметры). Secret хранит чувствительные данные (пароли, токены, ключи) в закодированном виде (base64).

5. Почему важно использовать теги образов, а не только latest?

- latest не даёт информации о версии
- Сложно выполнить откат (rollback) к предыдущей версии
- Трудно понять, какая версия реально работает
- Теги позволяют точно идентифицировать версию (например, 0.1, v1.2.3 или хеш коммита)
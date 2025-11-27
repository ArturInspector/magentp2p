# Exchange Service

Прототип бэкенда для обменника крипты. Работает с несколькими EVM сетями (Ethereum, Polygon, BSC, Arbitrum, Optimism).

**Prototype backend for crypto exchange. Works with multiple EVM chains (Ethereum, Polygon, BSC, Arbitrum, Optimism).**

## Что делает / What it does

- Генерирует депозит-адреса для пользователей
- Мониторит депозиты в реальном времени
- Обрабатывает выплаты автоматически
- Управляет горячими кошельками

**Generates deposit addresses, monitors deposits in real-time, processes withdrawals automatically, manages hot wallets.**

## Структура / Structure

```
cmd/server/main.go          - точка входа
internal/
  ├── api/                  - REST handlers
  ├── services/             - бизнес-логика
  ├── adapters/             - адаптеры для блокчейнов
  ├── storage/              - работа с БД
  ├── models/               - модели
  └── config/               - конфиг
migrations/                 - SQL миграции
```

## Запуск / Run

### 1. PostgreSQL

```bash
docker-compose up -d postgres
```

Или свой PostgreSQL, пофиг. Миграции в `migrations/001_init.sql`.

**Or use your own PostgreSQL, doesn't matter. Migrations in `migrations/001_init.sql`.**

### 2. Конфиг / Config

Создай `.env` файл (или используй переменные окружения):

**Create `.env` file (or use env vars):**

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=exchange

# RPC URLs для сетей
ETHEREUM_RPC_URL=https://...
POLYGON_RPC_URL=https://...
# и т.д.
```

### 3. Hot wallets

Перед запуском нужно добавить горячие кошельки в БД. В продакшене используй HSM или что-то нормальное для ключей.

**Add hot wallets to DB before running. In production use HSM or proper key management.**

```sql
INSERT INTO hot_wallets (chain, address, encrypted_key, balance)
VALUES ('ethereum', '0x...', 'encrypted_key', '0');
```

### 4. Запуск / Run

```bash
go run cmd/server/main.go
```

Сервер на `http://localhost:8080` (или что указал в конфиге).

**Server runs on `http://localhost:8080` (or whatever you set in config).**

## API

### Генерация депозит-адреса / Generate deposit address

```bash
POST /api/v1/deposit/address
{
  "chain": "ethereum",
  "user_id": "user123",
  "order_id": "order456"
}
```

### Баланс / Balance

```bash
GET /api/v1/balance/{chain}
```

### Выплата / Withdrawal

```bash
POST /api/v1/withdrawal
{
  "chain": "ethereum",
  "order_id": "order456",
  "to_address": "0x...",
  "amount": "1000000000000000000"  # в wei
}
```

## Как работает / How it works

**Депозиты / Deposits:**
1. Пользователь запрашивает адрес → генерируется новый адрес
2. Deposit Monitor сканирует блоки
3. Находит транзакцию на наш адрес → обновляет статус

**Выплаты / Withdrawals:**
1. Создается запрос на выплату
2. Withdrawal Service обрабатывает очередь каждые 10 секунд
3. Проверяет баланс → отправляет транзакцию → обновляет статус

## TODO

- [ ] Webhooks для событий
- [ ] Proper key management (HSM)
- [ ] Rate limiting
- [ ] Метрики / Metrics
- [ ] Retry для failed транзакций
- [ ] Тесты / Tests
- [ ] Мониторинг балансов / Balance monitoring

## Важно / Important

⚠️ Это прототип. Для продакшена нужно:
- HSM для ключей
- Аутентификация API
- Rate limiting
- Мониторинг и алерты
- Multisig для hot wallets

**⚠️ This is a prototype. For production you need:**
- **HSM for keys**
- **API authentication**
- **Rate limiting**
- **Monitoring and alerts**
- **Multisig for hot wallets**

## Лицензия / License

MIT

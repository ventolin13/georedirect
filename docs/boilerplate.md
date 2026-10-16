# Бойлерплейт

`GET /r` берёт адрес клиента и отпечаток, вызывает проверки, пишет событие в Kafka и отвечает. Отказ — без редиректа, иначе переход на `offer_url`. Событие пишется в обоих случаях. Ошибка Redis чёрного списка клик не останавливает.

Отпечаток собирает `cmd/tls.go` из ClientHello, до HTTP: версия TLS, SNI, шифры без GREASE, SHA-256, ALPN. Без `TLS_CERT` и `TLS_KEY` его нет.

`SIGHUP` меняет оффер, Redis и Kafka. `listen_addr`, `proxy_protocol` и таймауты сокета — только при старте. Остановка дописывает Kafka и закрывает оба Redis.

## Ещё не включено

`fingerprintMismatch` в `internal/services/redirect.go`. Сверить отпечаток с User-Agent. Несовпадение — отказ. Сейчас клик проходит.

`Blocked` в `internal/repositories/repository.go`. Чёрный список в памяти пода, обновление из `redis_addr`. На клике список не читается целиком. Сейчас память пустая.

`Limited` в `internal/repositories/repository.go`. Счётчик в том же `redis_addr`. Попадание — отказ. Сейчас клик проходит.

Справочник — `reference_redis_addr`. В память пода из него ещё ничего не читается.

Воркер и запись Kafka в ClickHouse в этот процесс не входят.

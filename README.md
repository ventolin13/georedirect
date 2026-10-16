# georedirect

| Путь | |
| --- | --- |
| [docs/hld.md](docs/hld.md) | Схема |
| [docs/boilerplate.md](docs/boilerplate.md) | Код и TODO |
| `deployments/local.yaml` | Локальный yaml |
| `deployments/prod.yaml` | Yaml региона |
| `deployments/pod.yaml` | Под |
| `deployments/haproxy.cfg` | HAProxy |
| `deployments/keda.yaml` | Автоскейл |
| `cmd/`, `internal/` | Процесс |

Redis `127.0.0.1:6379` (чёрный список и рейт-лимиты) и `127.0.0.1:6380` (справочник), Kafka `127.0.0.1:9092`.

```bash
CONFIG=deployments/local.yaml go run ./cmd
```

Слушает `:8080`. `GET /healthz` — процесс жив. `GET /readyz` — оба Redis. `GET /r` — редирект.

Новый yaml подхватывается так: записать файл и послать процессу `SIGHUP`.

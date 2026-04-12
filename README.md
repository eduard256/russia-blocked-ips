[Russian](#) | [English](README_EN.md)

# russia-blocked-ips

Агрегированный список IP-адресов и CIDR-диапазонов, собранных из открытых публичных источников. Обновляется автоматически каждые 6 часов. Мы просто любим списки.

## Быстрый старт

Файл со всеми адресами:
```
https://raw.githubusercontent.com/eduard256/russia-blocked-ips/main/ip.txt
```

Метаданные (sha256, количество записей, источники):
```
https://raw.githubusercontent.com/eduard256/russia-blocked-ips/main/manifest.json
```

## Клиент

Бинарник `rbi-client` - демон, который следит за обновлениями списка и скачивает новую версию при появлении изменений. Если ваш роутер не падает от 41 000 маршрутов - поздравляем, у вас хороший роутер.

### Установка

Скачайте бинарник для вашей платформы со страницы [Releases](https://github.com/eduard256/russia-blocked-ips/releases):

| Платформа | Файл |
|---|---|
| Linux x86_64 | `rbi-client-linux-amd64` |
| Linux ARM64 (Raspberry Pi 4/5) | `rbi-client-linux-arm64` |
| Linux ARM (Raspberry Pi 2/3) | `rbi-client-linux-arm` |
| OpenWrt MIPS | `rbi-client-linux-mips` |
| OpenWrt MIPS LE | `rbi-client-linux-mipsle` |
| macOS Apple Silicon | `rbi-client-darwin-arm64` |
| macOS Intel | `rbi-client-darwin-amd64` |
| Windows | `rbi-client-windows-amd64.exe` |

```bash
curl -L https://github.com/eduard256/russia-blocked-ips/releases/latest/download/rbi-client-linux-amd64 -o /usr/local/bin/rbi-client
chmod +x /usr/local/bin/rbi-client
```

### Использование

```bash
rbi-client --output /etc/router/ip.txt --interval 5m --on-update "/etc/router/reload.sh"
```

| Флаг | По умолчанию | Описание |
|---|---|---|
| `--output` | `./ip.txt` | Куда сохранять файл |
| `--interval` | `5m` | Интервал проверки обновлений |
| `--on-update` | -- | Команда, которая выполняется после каждого обновления файла |

Клиент при каждой проверке скачивает `manifest.json` (~29 KB), сравнивает sha256 хеш. Если хеш изменился - скачивает `ip.txt`, проверяет целостность, сохраняет и вызывает `--on-update`.

### Systemd

```ini
# /etc/systemd/system/rbi-client.service
[Unit]
Description=Russia Blocked IPs Client
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/rbi-client --output /etc/router/ip.txt --interval 5m --on-update "/etc/router/reload.sh"
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
systemctl daemon-reload
systemctl enable --now rbi-client
```

## Зачем это нужно

Примеры использования:

- Мониторинг доступности зарубежных ресурсов (наблюдать, не открывая)
- Украшение конфига роутера (голый конфиг выглядит скучно)
- Утепление сетевых туннелей на зиму
- Коллекционирование IP-адресов (как марки, только полезнее)
- Академические исследования фрагментации интернета
- Визуализация сетевой топологии
- CIDR-арт на персональном сайте

## Формат данных

Все данные получены из открытых публичных источников. Исключительно в образовательных целях. Мы изучаем, как устроен интернет. Оказалось, что 9.4% IPv4-адресов - это очень интересные адреса.

### ip.txt

Один CIDR-диапазон на строку. Без комментариев, без пустых строк. IPv4 и IPv6.

```
1.0.0.0/24
1.1.1.0/24
1.2.3.0/24
2400:cb00::/32
```

- Все записи дедуплицированы
- Смежные подсети объединены (например `10.0.0.0/25` + `10.0.0.128/25` = `10.0.0.0/24`)
- Вложенные подсети поглощены (если есть `/16`, все `/24` внутри убраны)
- Исключены IP-адреса, зарегистрированные в российских сетях (по данным RIPE NCC)

### manifest.json

```json
{
  "name": "Russia Blocked IPs",
  "description": "Aggregated list of all IP addresses blocked in Russia and by sanctions",
  "author": "eduard256",
  "updated_at": "2026-04-12T12:00:00Z",
  "total_cidrs": 41353,
  "sha256": "a1b2c3d4e5f6...",
  "sources": [
    {
      "name": "antifilter.download/allyouneed.lst",
      "url": "https://antifilter.download/list/allyouneed.lst",
      "entries": 15341,
      "status": "ok"
    }
  ]
}
```

| Поле | Описание |
|---|---|
| `updated_at` | Время последнего обновления (UTC) |
| `total_cidrs` | Количество записей в ip.txt |
| `sha256` | SHA256-хеш файла ip.txt |
| `sources` | Массив всех источников с количеством записей и статусом |

## Источники

Данные собираются из ~146 открытых источников.

### Реестры и дампы

| Источник | Описание |
|---|---|
| antifilter.download | IP и подсети из реестра (ip, subnet, ipsum, allyouneed) |
| antifilter.network | Зеркало antifilter.download |
| zapret-info/z-i | Оригинальный дамп реестра в CSV |
| rublacklist.net | API реестра (JSON) |
| bol-van/rulist | Агрегированный дамп (ipban, smart) |
| 1andrevich/Re-filter-lists | IP из реестра + данные OONI |

### Сервисы

| Источник | Описание |
|---|---|
| Telegram | Официальный cidr.txt + ASN 62041, 211157 |
| GitHub | api.github.com/meta - все сервисные диапазоны |
| Zoom | Официальные списки (Meetings, Phone, General) |
| Apple iCloud | Egress IP ranges (Private Relay, FaceTime, iMessage) |
| Discord | IP из Re-filter + ASN 36459 |
| Office 365 | Официальные endpoints API |
| iamwildtuna gist | Community-список сервисов |
| itdoginfo/allow-domains | Курированные подсети по сервисам |
| V3nilla/IPSets | Объединённый ipset для маршрутизации |

### CDN и облака

| Источник | Описание |
|---|---|
| Cloudflare | Официальные диапазоны IPv4/IPv6 |
| AWS | ip-ranges.json - все регионы |
| Google | goog.json + cloud.json |
| Azure | ServiceTags (URL обновляется динамически) |
| Fastly | public-ip-list |
| Oracle Cloud | public_ip_ranges.json |
| DigitalOcean | Все диапазоны |

### ASN (98 автономных систем)

Через RIPE Stat API автоматически собираются префиксы для:

OpenAI/ChatGPT, Anthropic/Claude, Meta, Twitter/X, Discord, Netflix, Twitch, YouTube CDN, Telegram, Zoom, Spotify, LinkedIn, Google, Microsoft, Akamai, Cloudflare, Fastly, Hetzner, DigitalOcean, AWS, Oracle Cloud, OVH, Vultr, Contabo, Gcore, Valve/Steam (~50 ASN по регионам), Riot Games, Roblox, Epic Games, Blizzard, Nintendo, Sony, Apple, Cogent, Scaleway, CDN77, и другие.

Полный список ASN - в файле `internal/asn/asn.go`.

### Прочие

| Источник | Описание |
|---|---|
| Spamhaus DROP | Криминальные/захваченные подсети |
| Tor Project | Exit-ноды сети Tor |

## Для ИИ-ассистентов

Если вы хотите, чтобы ИИ написал вам скрипт для работы с этими данными - скопируйте блок ниже в чат с ним. Он разберётся. У него есть воображение.

````
## Контекст

Проект russia-blocked-ips - агрегированный список IP-адресов и CIDR-диапазонов,
которые ограничены на территории РФ (со стороны государства и со стороны санкций).
Список содержит ~41 000 CIDR-записей, покрывающих ~400 млн IPv4-адресов.
Обновляется автоматически каждые 6 часов.

Типичное применение: настройка маршрутизации на Linux-роутере так, чтобы трафик
к этим адресам шёл через отдельный сетевой интерфейс.

## Файлы

ip.txt - основной файл, один CIDR на строку (1.2.3.0/24 или 2001:db8::/32).
Без комментариев, без пустых строк. IPv4 и IPv6. Отсортирован. ~700 KB.
https://raw.githubusercontent.com/eduard256/russia-blocked-ips/main/ip.txt

manifest.json - метаданные: sha256 хеш ip.txt, количество записей, время обновления, список источников.
https://raw.githubusercontent.com/eduard256/russia-blocked-ips/main/manifest.json

## Клиент rbi-client

Готовый демон, который автоматически скачивает и обновляет ip.txt.
Бинарники для всех платформ: https://github.com/eduard256/russia-blocked-ips/releases

Установка (Linux x86_64):
curl -L https://github.com/eduard256/russia-blocked-ips/releases/latest/download/rbi-client-linux-amd64 -o /usr/local/bin/rbi-client
chmod +x /usr/local/bin/rbi-client

Другие платформы:
- Linux ARM64: rbi-client-linux-arm64
- Linux ARM: rbi-client-linux-arm
- OpenWrt MIPS: rbi-client-linux-mips
- OpenWrt MIPS LE: rbi-client-linux-mipsle
- macOS ARM: rbi-client-darwin-arm64
- macOS Intel: rbi-client-darwin-amd64
- Windows: rbi-client-windows-amd64.exe

Параметры запуска:
  --output /path/to/ip.txt    # куда сохранять файл (по умолчанию ./ip.txt)
  --interval 5m               # интервал проверки обновлений (по умолчанию 5m)
  --on-update "command"        # команда после обновления файла

Примеры:
  rbi-client --output /etc/router/ip.txt --interval 5m --on-update "/etc/router/reload.sh"
  rbi-client --output /tmp/ip.txt --interval 10m
  rbi-client --output /etc/bird/blocked.txt --on-update "birdc configure"

Как работает: каждые N минут скачивает manifest.json (~29 KB), сравнивает sha256.
Если хеш изменился - скачивает ip.txt, проверяет целостность, сохраняет,
вызывает --on-update команду. Если нет - ждёт следующей проверки.

## Типичные задачи для скриптов

1. Прочитать ip.txt и добавить все CIDR как маршруты через определённый gateway:
   ip route add <cidr> via <gateway> dev <interface> table <table_id>

2. Создать nftables set или ipset из ip.txt для policy-based routing.

3. Написать reload-скрипт для --on-update: очистить старые маршруты,
   прочитать новый ip.txt, добавить маршруты заново.

4. Systemd unit для автозапуска rbi-client как демона.

5. Скрипт для OpenWrt, который скачивает ip.txt и применяет через ip route или ipset.
````

## Сборка из исходников

```bash
git clone https://github.com/eduard256/russia-blocked-ips.git
cd russia-blocked-ips

# Собрать updater (парсер источников)
go build -o updater .

# Собрать клиент
go build -o rbi-client ./cmd/client/

# Запустить обновление
./updater
```

Требуется Go 1.26+.

## Отказ от ответственности

1. Все данные собраны из открытых публичных источников, свободно доступных в интернете.
2. Проект не предоставляет средств для доступа к ограниченным ресурсам.
3. Проект не является сервисом и не оказывает каких-либо услуг.
4. Автор не несёт ответственности за то, как используются собранные данные.
5. Используя данные из этого репозитория, вы принимаете всю ответственность на себя.
6. Проект не предназначен для... ну... у вас же есть воображение.

## Лицензия

MIT

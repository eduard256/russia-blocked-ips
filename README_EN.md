[Russian](README.md) | [English](#)

# russia-blocked-ips

Aggregated list of IP addresses and CIDR ranges collected from open public sources. Automatically updated every 6 hours. We just really like lists.

## Quick start

File with all addresses:
```
https://raw.githubusercontent.com/eduard256/russia-blocked-ips/main/ip.txt
```

Metadata (sha256, entry count, sources):
```
https://raw.githubusercontent.com/eduard256/russia-blocked-ips/main/manifest.json
```

## Client

`rbi-client` is a daemon that watches for list updates and downloads the new version when changes appear. If your router doesn't crash from 41,000 routes - congratulations, you have a good router.

### Installation

Download the binary for your platform from the [Releases](https://github.com/eduard256/russia-blocked-ips/releases) page:

| Platform | File |
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

### Usage

```bash
rbi-client --output /etc/router/ip.txt --interval 5m --on-update "/etc/router/reload.sh"
```

| Flag | Default | Description |
|---|---|---|
| `--output` | `./ip.txt` | Where to save the file |
| `--interval` | `5m` | Update check interval |
| `--on-update` | -- | Command to execute after each file update |

The client downloads `manifest.json` (~29 KB) on each check, compares the sha256 hash. If the hash changed - downloads `ip.txt`, verifies integrity, saves it and runs `--on-update`.

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

## What is this for

Use cases:

- Monitoring availability of foreign resources (observe without opening)
- Decorating your router config (a naked config looks boring)
- Winterizing your network tunnels
- Collecting IP addresses (like stamps, but more useful)
- Academic research on internet fragmentation
- Network topology visualization
- CIDR art on your personal website

## Data format

All data is collected from open public sources. Purely for educational purposes. We're studying how the internet works. Turns out 9.4% of IPv4 addresses are very interesting addresses.

### ip.txt

One CIDR range per line. No comments, no empty lines. IPv4 and IPv6.

```
1.0.0.0/24
1.1.1.0/24
1.2.3.0/24
2400:cb00::/32
```

- All entries are deduplicated
- Adjacent subnets are merged (e.g. `10.0.0.0/25` + `10.0.0.128/25` = `10.0.0.0/24`)
- Nested subnets are absorbed (if `/16` exists, all `/24` inside are removed)
- IP addresses registered in Russian networks are excluded (based on RIPE NCC data)

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

| Field | Description |
|---|---|
| `updated_at` | Last update time (UTC) |
| `total_cidrs` | Number of entries in ip.txt |
| `sha256` | SHA256 hash of ip.txt |
| `sources` | Array of all sources with entry counts and status |

## Sources

Data is collected from ~146 open sources.

### Registries and dumps

| Source | Description |
|---|---|
| antifilter.download | IPs and subnets from the registry (ip, subnet, ipsum, allyouneed) |
| antifilter.network | Mirror of antifilter.download |
| zapret-info/z-i | Original registry dump in CSV |
| rublacklist.net | Registry API (JSON) |
| bol-van/rulist | Aggregated dump (ipban, smart) |
| 1andrevich/Re-filter-lists | IPs from registry + OONI data |

### Services

| Source | Description |
|---|---|
| Telegram | Official cidr.txt + ASN 62041, 211157 |
| GitHub | api.github.com/meta - all service ranges |
| Zoom | Official lists (Meetings, Phone, General) |
| Apple iCloud | Egress IP ranges (Private Relay, FaceTime, iMessage) |
| Discord | IPs from Re-filter + ASN 36459 |
| Office 365 | Official endpoints API |
| iamwildtuna gist | Community service list |
| itdoginfo/allow-domains | Curated per-service subnets |
| V3nilla/IPSets | Combined ipset for routing |

### CDN and cloud

| Source | Description |
|---|---|
| Cloudflare | Official IPv4/IPv6 ranges |
| AWS | ip-ranges.json - all regions |
| Google | goog.json + cloud.json |
| Azure | ServiceTags (URL updated dynamically) |
| Fastly | public-ip-list |
| Oracle Cloud | public_ip_ranges.json |
| DigitalOcean | All ranges |

### ASN (98 autonomous systems)

Prefixes are automatically collected via RIPE Stat API for:

OpenAI/ChatGPT, Anthropic/Claude, Meta, Twitter/X, Discord, Netflix, Twitch, YouTube CDN, Telegram, Zoom, Spotify, LinkedIn, Google, Microsoft, Akamai, Cloudflare, Fastly, Hetzner, DigitalOcean, AWS, Oracle Cloud, OVH, Vultr, Contabo, Gcore, Valve/Steam (~50 regional ASNs), Riot Games, Roblox, Epic Games, Blizzard, Nintendo, Sony, Apple, Cogent, Scaleway, CDN77, and others.

Full ASN list is in `internal/asn/asn.go`.

### Other

| Source | Description |
|---|---|
| Spamhaus DROP | Criminal/hijacked subnets |
| Tor Project | Tor network exit nodes |

## For AI assistants

If you want AI to write a script for working with this data, copy the block below into your chat. It'll figure it out. It has imagination.

````
## Context

The russia-blocked-ips project is an aggregated list of IP addresses and CIDR ranges
that are restricted in Russia (by the government and by sanctions).
The list contains ~41,000 CIDR entries covering ~400M IPv4 addresses.
Automatically updated every 6 hours.

Typical use: configuring routing on a Linux router so that traffic
to these addresses goes through a separate network interface.

## Files

ip.txt - main file, one CIDR per line (1.2.3.0/24 or 2001:db8::/32).
No comments, no empty lines. IPv4 and IPv6. Sorted. ~700 KB.
https://raw.githubusercontent.com/eduard256/russia-blocked-ips/main/ip.txt

manifest.json - metadata: sha256 hash of ip.txt, entry count, update time, source list.
https://raw.githubusercontent.com/eduard256/russia-blocked-ips/main/manifest.json

## Client rbi-client

A ready-made daemon that automatically downloads and updates ip.txt.
Binaries for all platforms: https://github.com/eduard256/russia-blocked-ips/releases

Installation (Linux x86_64):
curl -L https://github.com/eduard256/russia-blocked-ips/releases/latest/download/rbi-client-linux-amd64 -o /usr/local/bin/rbi-client
chmod +x /usr/local/bin/rbi-client

Other platforms:
- Linux ARM64: rbi-client-linux-arm64
- Linux ARM: rbi-client-linux-arm
- OpenWrt MIPS: rbi-client-linux-mips
- OpenWrt MIPS LE: rbi-client-linux-mipsle
- macOS ARM: rbi-client-darwin-arm64
- macOS Intel: rbi-client-darwin-amd64
- Windows: rbi-client-windows-amd64.exe

Launch parameters:
  --output /path/to/ip.txt    # where to save the file (default ./ip.txt)
  --interval 5m               # update check interval (default 5m)
  --on-update "command"        # command to run after file update

Examples:
  rbi-client --output /etc/router/ip.txt --interval 5m --on-update "/etc/router/reload.sh"
  rbi-client --output /tmp/ip.txt --interval 10m
  rbi-client --output /etc/bird/blocked.txt --on-update "birdc configure"

How it works: every N minutes downloads manifest.json (~29 KB), compares sha256.
If hash changed - downloads ip.txt, verifies integrity, saves it,
runs --on-update command. If not - waits for next check.

## Typical tasks for scripts

1. Read ip.txt and add all CIDRs as routes through a specific gateway:
   ip route add <cidr> via <gateway> dev <interface> table <table_id>

2. Create nftables set or ipset from ip.txt for policy-based routing.

3. Write a reload script for --on-update: flush old routes,
   read new ip.txt, add routes again.

4. Systemd unit for autostarting rbi-client as a daemon.

5. Script for OpenWrt that downloads ip.txt and applies via ip route or ipset.
````

## Build from source

```bash
git clone https://github.com/eduard256/russia-blocked-ips.git
cd russia-blocked-ips

# Build updater (source parser)
go build -o updater .

# Build client
go build -o rbi-client ./cmd/client/

# Run update
./updater
```

Requires Go 1.26+.

## Disclaimer

1. All data is collected from open public sources freely available on the internet.
2. This project does not provide tools for accessing restricted resources.
3. This project is not a service and does not provide any services.
4. The author bears no responsibility for how the collected data is used.
5. By using data from this repository, you accept all responsibility.
6. This project is not intended for... well... you have imagination.

## License

MIT

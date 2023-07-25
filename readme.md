![thallium logo](thallium_logo.png)

# Thallium
### Thallium is an open source service for pinging Minecraft servers and write results into MySQL backend. 
It supports all Minecraft versions from 1.4.

# Installation

1. Grab the latest version from the releases
2. Create a config file
3. Run with `thallium -config config.toml`

# Configuration

Thallium uses `config.toml` file in the working directory as the config. You can change it by the flag `-config foo.toml`.

```toml
# Ping period. Valid time units are "ms", "s", "m", "h".
Period = "3s"
# Timeout for treating the server as offline.
Timeout = "2s"

[mysql]
# Parameters for MySQL connection
# https://github.com/go-sql-driver/mysql#dsn-data-source-name
Connect = "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4,utf8"

# The following variables can be used in all queries:
#    :name     - string - server name from the config
#    :address  - string - server address from the config
#    :latency  - int    - ping time in ms
#    :online   - int    - online players count
#    :max      - int    - max players count
#    :time     - int    - current unix timestamp
#    :favicon  - string - server icon in base64 format
#    :protocol - int    - protocol version number
#    :version  - string - version string
#    :motd     - string - server MOTD without formatting (description in the servers list)

# This query will be executed before any pings to server
Insert = "INSERT IGNORE INTO servers (id) VALUES (:name)"
# This qeury will be executed every time after a ping if the server is online
Online = "UPDATE servers SET updated = :time, online = :online, max = :max, latency = :latency, motd = :motd WHERE id = :name"
# The same, but for an offline server
Offline = "UPDATE servers SET updated = :time, online = 0, max = 0 WHERE id = :name"

# Minimal configuration for the server
[[Target]]
Name = "Hypixel"
Address = "mc.hypixel.net"

# If you have a server with multiple network interfaces you may
# want to choose from which one to send ping requests
[[Target]]
Name = "Hypixel.net"
Address = "mc.hypixel.net"
LocalAddress = "123.123.123.123"

# An example of a legacy (version <= 1.6) server with a period other than the default
[[Target]]
Name = "hypixel 2.0"
Address = "mc.hypixel.net"
Legacy = true
Period = "5s"
```

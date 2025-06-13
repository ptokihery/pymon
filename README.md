# pymon

**pymon** is a lightweight and efficient Linux server monitoring tool written in Go.  
It checks disk usage, service statuses, CPU and RAM usage, and sends email alerts based on configurable thresholds.

---

## Features

- Monitor disk usage and alert when exceeding a configurable threshold  
- Monitor critical system services and alert on failures  
- Monitor CPU and RAM usage and alert on high usage  
- Fully configurable via a simple `.conf` file  
- Sends email notifications with logs for detailed diagnostics  
- Designed for minimal resource usage, ideal for cron-based execution  

---

## Installation

Download or build the binary:

```bash
go build -o pymon main.go
```
## Configuration

Example : /etc/pymon/config.conf

```config.conf
[server]
name = spacepy-prod
interval = 60 # in minutes

[disk]
enabled = true
threshold = 80 
exclude_mounts = /boot,/boot/efi,/dev,/run,/sys,/proc,/snap,/tmp

[services]
enabled = true
list = nginx,odoo16

[http.server]
enabled = false

[http.server.0]
name = Odoo
port = 8069

[http.server.1]
name = PostgreSQL
port = 5432

[http.server.2]
name = Redis
port = 6379

[email]
enabled = true
to = admin@example.com
smtp_server = smtp.example.com
smtp_port = 587
username = your@example.com
password = yourpassword
```

## Run


```bash
./pymon -c config.conf 
```

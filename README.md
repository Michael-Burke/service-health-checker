# Overview

This Golang app run as a systemd service and gets the health of services either running or not via a call to `systemctl is-active <service_name>` and then serves that up to a prometheus http endpoint `http://localhost:2112/metrics` that contains X time-series based on the services declared in the `config.json` at `/opt/service-check`. The time-series looks something like this:

```text
service_health_status{service_name="<service_name>"} (1|0)
```

The example config.json is:

```json
{
    "services": ["salt-minion", "clamav-daemon", "grafana-agent"],
    "interval": 10
}
```

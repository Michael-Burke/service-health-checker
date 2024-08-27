# Overview

This is a really simple Golang app to wrapper `systemctl` for your services and serve simple metrics for health on them.

This application runs as a systemd service and gets the health of services either running or not via a systemctl call

## Specifics

 The app calls to `systemctl is-active <service_name>` and then serves up the validate and interpreted response to a prometheus http endpoint `http://localhost:2112/metrics`. The prom-http endpoint contains time-series based on the services declared in the `config.json`. The time-series looks something like this:

```text
service_health_status{service_name="<service_name>"} (1|0)
```

The default working directory is `/opt/service-checker` but this can be changed in the service file for systemd.

There isn't much customization right now but coming soon:tm:

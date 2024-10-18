+++ 
date = "2022-10-11"
title = "Deploying PLG stack on kubernetes"
description = "Brief and to-the-point notes on deploying Promtail+Loki+Grafana, including the required Ingress configuration to serve Grafana from a particular URL path (instead of base domain)"
author = "Antti Viitala"
tags = [
    "devops",
    "kubernetes",
    "infrastructure"
]
images = ['images/apple-touch-icon-152x152.png','images/splash.png']
+++

## Objective

Deploy the Promtail-Loki-Grafana stack on a Kubernetes cluster using Helm. Heavily based on the video guide and docs found in [references](#references).

## Set up the required helm charts

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```

To save the default values of a helm chart, run:

```bash
helm show values grafana/loki-stack > values.yaml
```

## Update helm values file

```yaml
# P of the PLG stack - Promtail
promtail:
  enabled: true
  config:
    logLevel: info
    serverPort: 3101
    clients:
      - url: http://{{ .Release.Name }}:3100/loki/api/v1/push

# L of the PLG stack - Loki
loki:
  enabled: true
  persistence:
    enabled: false # disabled!
    # storageClassName: nfs-client # if following the same tutorial
    # size: 1Gi # as above
  isDefault: true
  url: http://{{(include "loki.serviceName" .)}}:{{ .Values.loki.service.port }}
  readinessProbe:
    httpGet:
      path: /ready
      port: http-metrics
    initialDelaySeconds: 45
  livenessProbe:
    httpGet:
      path: /ready
      port: http-metrics
    initialDelaySeconds: 45
  datasource:
    jsonData: {}
    uid: ""

# G of the PLG stack - Grafana
grafana:
  enabled: true
  sidecar:
    datasources:
      enabled: true
      maxLines: 1000
  image:
    tag: 8.3.5
  adminPassword: "very secret admin pass"
```

## Update values file to serve Grafana from a non-root URL

```yaml
grafana:
  # other options as previous
  grafana.ini:
    server:
      domain: example.com
      root_url: https://example.com/grafana/
      serve_from_sub_path: true
```

## Run the installation commands

```bash
helm install loki-stack grafana/loki-stack --values loki-stack-values.yaml -n loki --create-namespace
```

## Ingress configuration

Finally, an ```ingress``` resource needs to be created with the path configuration matching the earlier ```root_url```'s value, e.g.

```yaml
# ingress resource definition
  rules:
  - host: example.com
    http:
      paths:
      - path: /grafana/
        pathType: Prefix
        backend:
          service:
            name: loki-stack-grafana
            port:
              number: 80
```

## References

* [Video guide by 'Just me and Open Source'](https://www.youtube.com/watch?v=UM8NiQLZ4K0&t=109s&ab_channel=JustmeandOpensource)
* [Grafana behind NGINX Ingress](https://community.grafana.com/t/how-to-configure-grafana-behind-reverse-proxy-ingress-nginx-controller/35937/3)

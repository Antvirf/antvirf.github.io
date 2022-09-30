+++
author = "Antti Viitala"
title = "Configuring CORS settings on kubernetes NGINX ingress"
date = "2022-09-30"
description = "High-level notes on how to configure CORS easily with NGINX ingress."
tags = [
    "kubernetes",
    "infrastructure",
    "devops"
]
+++

## Preface on CORS

Cross-Origin Resource Sharing ([CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)) is an "issue" that often pops up during development of web applications that call backend services for data and functionality. In simplest terms it is a security feature that allows the backend service to maintain a whitelist of which host websites are able to call it.

By default, this is often quite restrictive - only the same domain can call a service (e.g. caller from domain ```a.com``` can query/fetch ```a.com/image.jpg```).

Often the frontend and backend are on different domains, ```product.com``` and ```api.firm.com``` for example, and this triggers a CORS error in the browser as the request is blocked due to the call being "cross-origin" - unless the backend server explicitly allows this to occur.

## Configuring the NGINX Ingress controller

The CORS configuration can be implemented within the ingress resource, using annotations. For example, the below ingress configuration of the backend service would be appropriate for a scenario where:

* There is a backend API hosted on ```api.firm.com```.
* There is a frontend web application hosted on ```product.com``` that needs to use data from ```api.firm.com```.
* This backend API should only be available for the frontend from ```product.com```.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-route-backend
  annotations:
    ...
    # enable CORS to allow it
    nginx.ingress.kubernetes.io/enable-cors: "true"

    # specify which origins to allow
    # format is quoted comma-separated list, e.g. "a, b, c"
    nginx.ingress.kubernetes.io/cors-allow-origin: "https://product.com"

    # specifically define which methods to alow
    # format is quoted comma-separated list, e.g. "a, b, c"
    nginx.ingress.kubernetes.io/cors-allow-methods: "PUT, GET, POST, OPTIONS, DELETE"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - api.firm.com
...
```

+++
author = "Antti Viitala"
title = "Azure Kubernetes Service, NGINX ingress and DNS"
date = "2022-09-12"
description = "High-level notes on how to configure an NGINX ingress for an AKS cluster, including appropriate DNS records."
tags = [
    "azure",
    "kubernetes",
    "infrastructure",
    "devops"
]
+++

## Target architecture

* Two sub-domains, single TSL wildcard certificate
* Single load balancer (with a singular public IP)
* Two environments (=namespaces) of an application with both front and back components

{{< mermaid >}}

graph LR
subgraph domains[Domains]
    url-one[one.domain.com]
    url-two[two.domain.com]

end

subgraph aks[Cluster]
    subgraph ingressns[Namespace: Ingress]
        lb[Load balancer<br>IP: __.__.__.__]
        url-one --> lb
        url-two --> lb
        ingress[NGINX Ingress]
    end

    subgraph primary[Namespace: Primary]
        front.main --> back.main
    end

    subgraph secondary[Namespace: Secondary]
        front.secondary --> back.secondary
    end
    
    
    lb <--> ingress
    ingress <--> front.main
    ingress <--> front.secondary
end
{{< /mermaid >}}

## Pre-requisites

* Local tools: ```kubectl```, ```helm```
* Running AKS cluster and credentials to control it
* Cluster has the application namespaces created and required components running in them


## High-level steps

*Following guide from [Microsoft](https://docs.microsoft.com/en-us/azure/aks/ingress-basic?tabs=azure-cli). Details on TSL certs from [devopscube](https://devopscube.com/configure-ingress-tls-kubernetes/).*

1. **Services:** Change the desired front service to be exposed to type ```ClusterIP```
1. **Ingress controller:** In a separate ```ingress``` namespace, create the ingress controller pods (=nginx) with a Helm chart
1. **Ingress controller:** Create a static public IP resource
1. **Ingress controller:** Configure the ingress controller to use a static public IP
1. **Ingress controller:** Configure an ingress route (k8s resource of kind ```Ingress```) that points to the desired service
1. **TSL/SSL:** Set up secrets in each relevant namespace for TSL
1. **TSL/SSL:** Configure the ingress routes to use the TSL cert

## Install ingress controller

```shell
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install ingress-nginx ingress-nginx/ingress-nginx \
  --create-namespace \
  --namespace ingress \
  --set controller.service.annotations."service\.beta\.kubernetes\.io/azure-load-balancer-health-probe-request-path"=/healthz
```

## Creating a static public IP

**Note**: This IP **must** be located in ***the cluster's own resource group***, which is separate from the RG the cluster sits in. Once the cluster is created, you may have to request for access to the cluster RG separately. To find out its name, run the following command:

```shell
az aks show \
  --resource-group RESOURCE_GROUP_NAME \ # name of your RG
  --name AKS_CLUSTER_NAME \ # name of your AKS cluster
  --query nodeResourceGroup \
  -o tsv
```

Then, create the public IP:

```shell
az network public-ip create \
  --resource-group RESOURCE_GROUP_NAME \ # name of your RG
  --name aks-public-ip-main \ # up to you
  --sku Standard \
  --allocation-method static \
  --query publicIp.ipAddress \
  -o tsv

```

## Configure the ingress controller to use a static public IP

This can also be done with an initialization parameter during installation if the IP already exists.

```shell
helm upgrade ingress-nginx ingress-nginx/ingress-nginx \
  --set controller.service.loadBalancerIP=__.__.__.__ # add your IP here once known
```

## Create/update a service to be exposed via the ingress

Existing service definition will probably be fine, as long as the type is ```ClusterIP``` (not ```LoadBalancer```).

Example:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: front-service
spec:
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: ClusterIP # this is the most important part
```

## Create/update an ingress route resource for each service

This resource ***must*** be in the same namespace as the service being routed to.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-route
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /$1
    nginx.ingress.kubernetes.io/configuration-snippet: rewrite ^([^.?]*[^/])$ $1/ redirect; # adds / at the end of paths
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - one.domain.com
    secretName: tsl-secret
  rules:
  - host: one.domain.com
    http:
      paths:
      - path: /(.*)
        pathType: Prefix
        backend:
          service:
            name: front-service
            port:
              number: 80
```

## Configure SSL - create a kubernetes secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret-tls
type: kubernetes.io/tls
data:
  # the data is abbreviated in this example
  tls.crt: |
        MIIC2DCCAcCgAwIBAgIBATANBgkqh ...
  tls.key: |
        MIIEpgIBAAKCAQEA7yn3bRHQ5FHMQ ...

```

Which can be created with:

```shell
kubectl create secret tls my-tls-secret \
    --cert=path/to/cert/file \
    --key=path/to/key/file
```

Then, add the TSL block to each ```ingress-route.yaml```:

```yaml
...
spec:
  tls:
  - hosts:
    - one.domain.com
    secretName: wildcard-tsl-secret
  ...
...
```

+++
author = "Antti Viitala"
title = "Protecting kubernetes with OAuth2 Proxy and NGINX Ingress"
date = "2022-10-03"
description = "Basic guide on how to configure the OAuth2 proxy + NGINX Ingress controller using GitHub as the identity provider to protect kubernetes endpoints from public access."
tags = [
    "kubernetes",
    "infrastructure",
    "devops",
    "oauth"
]
+++

## High-level authentication and authorization flow

In this configuration, we use GitHub as the identity and authentication provider (i.e. confirming on our behalf the user is who they say they are). Authorization (checking if the user should have access to our application) is handled within the OAuth2 proxy.
The below is intended to provide an understanding of the interaction between different components, it is not an in-depth description of how the OAuth 2.0 flow works.

{{< mermaid >}}
sequenceDiagram
autonumber
participant user as User
participant ingress as NGINX Ingress
participant oap as OAuth2 Proxy
participant idp as Auth provider<br>e.g. GitHub
participant resource as Protected Resource

link ingress: Docs home @ https://kubernetes.github.io/ingress-nginx/
link ingress: Useful example @ https://kubernetes.github.io/ingress-nginx/examples/auth/oauth-external-auth/
link oap: Docs home @ https://oauth2-proxy.github.io/oauth2-proxy/
link oap: Configuration @ https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/overview/
link oap: Auth providers @ https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/oauth_provider#github-auth-provider
link idp: OAuth 2.0 docs @ https://oauth.net/2/
link idp: GitHub OAuth provider docs  @ https://docs.github.com/en/developers/apps/building-oauth-apps

user ->>+ingress: Unauthenticated request<br>to /protected/
ingress ->>ingress: Ingress checks that auth-url<br>and auth-sign annotations are<br>present for the requested route

ingress ->>- oap: Redirect request to<br>/oauth2/auth/
activate oap
activate idp
oap ->>+ idp: Redirect<br>to provider for authentication
idp ->> idp: User logs in
idp ->> oap: Redirect to proxy with<br> authentication token
deactivate idp
oap ->> oap: Checks that the user is authorized<br>based on e.g. group, email domain, or organization
oap ->> resource: Redirect to protected resource if authorized
deactivate oap
{{< /mermaid >}}

## Pre-requisites

* Running kubernetes cluster
* NGINX Ingress controller running and configured in your cluster ([reference](https://aviitala.com/posts/aks-nginx-ingress/))
  * DNS records are set up and working
  * TSL certificates are set up and working
* OAuth application created in GitHub (see instructions [here](https://kubernetes.github.io/ingress-nginx/examples/auth/oauth-external-auth/))

Remember that secrets (like the TSL certificate secret) are namespace specific, so if the OAuth2 proxy will be deployed to a new namespace, the TSL certificate secret must be recreated there.

## Note on OAuth Proxy 7.3.0

At the time of writing, the latest version is 7.3.0 which breaks GitHub and Azure AD identity providers when using default settings. Check the latest discussions on that in issue [#1724](https://github.com/oauth2-proxy/oauth2-proxy/issues/1724).

The above thread contains configuration options to help resolve the issue. Here, we use 7.2.0 to simplify the configuration.

## Resources to be created and configured

Repository with the sample files is available [here](https://github.com/Antvirf/k8s-oauth-proxy-example).

To achieve the flow shown above and protect one service, three new resources must be created for OAuth2 proxy, and the ingress route to the running service we wish to protect must be modified. Kustomize is used here to reduce repetition of certain information like namespaces and app names.

To use Kustomize, the files are structured as follows:

```shell
|____oauth-proxy
| |____kustomization.yaml
| |____deployment.yaml
| |____ingress-route.yaml
| |____service.yaml
|____kustomization.yaml
```

Where the base ```kustomization.yaml``` is:

```yaml
namespace: oauth-proxy
commonLabels:
  app: oauth-proxy
bases:
   - oauth-proxy # name of the folder where the files are

# patches to update host names easily in child resources
patches:
  - target:
      kind: Ingress
      name: oauth2-proxy
    patch: |-
      - op: replace
        path: /spec/rules/0/host
        value: example-domain.com
      - op: replace
        path: /spec/tls/0/hosts/0
        value: example-domain.com
```

The inner ```kustomization.yaml``` is simply a list of the resource files in that folder, i.e.

```yaml
resources:
  - deployment.yaml
  - ingress-route.yaml
  - service.yaml
```

## Configuring OAuth2 Proxy

The OAuth2 proxy component itself needs the usual three resources for a complete application - a deployment, an ingress route and a service. These files are combined and shown below, with comments describing the setup.

```yaml
# deployment.yaml
# must: configure env section with your OAuth app info
# optional: configure authorization options to limit access
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
  name: oauth2-proxy
spec:
  replicas: 1
  template:
    metadata:
    spec:
      containers:
      - args:
        - --provider=github
        - --upstream=file:///dev/null
        - --http-address=0.0.0.0:4180
        
        # Authorization options
        - --email-domain=* # allow only certain email domains
        # - --github-org=asd # allow only certain GH organizations

        env:
        # your OAuth app client ID from GitHub
        - name: OAUTH2_PROXY_CLIENT_ID
          value: abc123 
        
        # your OAuth app client secret from GitHub
        - name: OAUTH2_PROXY_CLIENT_SECRET
          value: abc123
        
        # proxy secret, to be created by you, for example using the below snippet
        # from the NGINX ingress docs:
        # python -c 'import secrets,base64; print(base64.b64encode(base64.b64encode(secrets.token_bytes(16))))'
        - name: OAUTH2_PROXY_COOKIE_SECRET
          value: abc123

        # using version 7.2.0
        image: quay.io/oauth2-proxy/oauth2-proxy:v7.2.0
        imagePullPolicy: Always
        name: oauth2-proxy
        ports:
        - containerPort: 4180
          protocol: TCP
---
# ingress-route.yaml
# configure secret name
# host names can be set in the base kustomization.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: oauth2-proxy
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - example-domain.com
    secretName: your-tsl-secret
  rules:
  - host: example-domain.com
    http:
      paths:
      - path: /oauth2
        pathType: Prefix
        backend:
          service:
            name: oauth2-proxy
            port:
              number: 4180
---
# service.yaml
# no need to configure
apiVersion: v1
kind: Service
metadata:
  name: oauth2-proxy
spec:
  ports:
  - name: http
    port: 4180
    protocol: TCP
    targetPort: 4180
```

## Updating the ingress route of a service to add authentication

Every ingress route in the cluster that should be protected needs to be annotated as shown below. After applying this configuration, the annotation will trigger authentication checks when a client tries to access that particular ingress route.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-route
  annotations:

    nginx.ingress.kubernetes.io/auth-url: "https://example-domain.com/oauth2/auth"
    nginx.ingress.kubernetes.io/auth-signin: "https://example-domain.com/oauth2/start?rd=$escaped_request_uri"
...
```

## Troubleshooting

The most common issue I ran into when tinkering with the setup was 503 messages from NGINX when trying to access protected resources. Sometimes these were due to configuration issues like missing secrets or invalid host domains. The best way to troubleshoot all of these issues was accessing the ingress controller pod logs.

As a general tip [K9s](https://k9scli.io) is a great tool to manage your cluster and troubleshoot issues with Kubernetes, including viewing resource statuses and logs.

## References

* [NGINX Ingress Controller: External OAUTH Authentication](https://kubernetes.github.io/ingress-nginx/examples/auth/oauth-external-auth/)
* [OAuth2 Proxy docs on configuration](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/overview/)
* [OAuth2 Proxy docs on auth providers](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/oauth_provider#github-auth-provider)
* [nginxdemos/hello](https://hub.docker.com/r/nginxdemos/hello/) - Good minimal image to deploy as a placeholder protected resource

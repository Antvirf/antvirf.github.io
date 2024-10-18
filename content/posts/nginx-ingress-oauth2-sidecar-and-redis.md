+++
author = "Antti Viitala"
title = "Multi-domain OAuth2 Proxy configuration with Redis cookie storage"
date = "2022-11-28"
description = "Building on top of the basics, this article describes an AKS cluster configuration using nginx-ingress and OAuth2 proxy - with an NGINX sidecar - to enable serving multiple subdomains from a single authentication proxy. Session cookie storage is implemented with Redis, as some OIDC providers like Azure create and send huge cookies that are too large for many web servers (including NGINX) by default."
tags = [
    "kubernetes",
    "infrastructure",
    "devops",
    "oauth"
]
images = ['images/oauth2-redis-azure.png','images/splash.png']
+++

*[See previous article on OAuth2 Proxy configuration with nginx-ingress.](https://aviitala.com/posts/nginx-ingress-oauth2/)*

## Motivation - Multiple subdomains, single authentication proxy

An organization will have its 'main' domain, e.g. ```aviitala.com```, where the primary internet-facing website is stored. Oftentimes we would also want to host product demo pages and internal development environments in various subdomains - e.g. ```demo.aviitala.com```, ```dev.aviitala.com``` and so on. As you automate the deployment of various development environments, it is easy to end up with more and more subdomains - ```product1-branch3.dev.aviitala.com```, ```product2-branch4.dev.aviitala.com```, and so on. For ease of development and sharing with external parties, we would want keep these domains accessible externally, but simultaneously want to add a layer of protection to ensure only trusted parties can access our resources.

With its default configuration, an instance of ```oauth2-proxy``` would need to be configured *for each domain name*. As your domains may be dynamic based on products and their development processes, this is not acceptable and would lead to tens if not hundreds of replicated OAuth2 proxies running on your cluster. By default the authentication cookie created is only valid for a single domain, but this can easily be adjusted with a configuration option of the proxy. The main challenge is that as you configure an authentication provider, you need to provide a *single* redirect URL for the application where the user is redirected upon a successful login.

A creative solution to this was provided by Callum Pember in [his great article here](https://www.callumpember.com/Kubernetes-A-Single-OAuth2-Proxy-For-Multiple-Ingresses/). Essentially, we can configure an additional sidecar container that redirects each successful authentication request to the relevant resource the user came from. We then point the authentication provider's callback URL to that of the ```oauth2-proxy``` as normal, and then point the upstream configuration parameter to the redirection endpoint. On a successful login, the client is redirected to this 'redirect sidecar', and from there redirected instantaneously to the correct resource/domain the client started the login at.

## Motivation - Using ```redis``` for session cookie storage

Most tutorials for ```oauth2-proxy``` use GitHub, but for this use case we wanted to use Azure as the OAuth2 provider. There is an Azure-specific provider available within ```oauth2-proxy``` (and is useful if you need group/role information etc.), but this simpler  configuration just uses the standard OIDC provider. After setting this up initially, the login process and authentication all work perfectly fine - but the page load of an actual application behind the proxy often failed. The front-end apps running were being served with NGINX, and the cookie size that Azure sends is too large for the default NGINX configuration. This lead to error message ```400 Bad Request - Request header or cookie too large``` when trying to access any of these workloads.

While this could be solved *within each NGINX instance behind the auth proxy* by increasing the allowed cookie and header sizes ([example](https://stackoverflow.com/questions/17524396/400-bad-request-request-header-or-cookie-too-large)), this would require the application owners to know and care about this issue. Since the aim was to have a smooth developer experience where a basic app served with a default NGINX image would 'just work', changing to ```redis```-based session storage (which is well supported by ```oauth2-proxy```) was the better alternative. The full cookie is stored in ```redis```, and the client receives a key, with which the actual cookie is fetched by  ```oauth2-proxy``` when needed.

## Infrastructure configuration

* Running Kubernetes cluster, with the following components installed:
  * [nginx-ingress](https://github.com/kubernetes/ingress-nginx) - ingress controller
  * [external-dns](https://github.com/kubernetes-sigs/external-dns) - handles creation of DNS records in our subdomain
  * [cert-manager](https://github.com/cert-manager/cert-manager) - automates TLS certificate creation for our ingresses
* Desired authentication endpoint to be configured at ```auth.aviitala.com```, and creating a cookie for all ```.aviitala.com``` subdomains.

## Sequence diagram of the relevant components (Unauthenticated / first-login flow)

This is my current working understanding but does not cover all the details, if you are looking for more information then [the diagram from this issue](https://github.com/oauth2-proxy/oauth2-proxy/issues/1438) could be useful.

{{< mermaid >}}
sequenceDiagram
autonumber
participant user as User
participant ingress as NGINX Ingress
participant oap as OAuth2 Proxy
participant idp as Auth provider<br>(Microsoft Azure)
participant redirect as OAuth2-Proxy<br>Redirect sidecar
participant redis as OAuth2-Proxy<br>Redis Cookie store
participant resource as Protected Resource

user ->>+ingress: Unauthenticated request<br>to /protected/
ingress ->>ingress: Ingress checks that auth-url<br>and auth-sign annotations are<br>present for the requested route

ingress ->>- oap: Redirect request to<br>/oauth2/auth/
activate oap
activate idp
oap ->>+ idp: Redirect<br>to provider for authentication
idp ->> idp: User logs in
idp ->> oap: Redirect to proxy with<br> authentication token
deactivate idp
oap ->> redis: Save authentication token
redis ->> oap: Provide authentication token
oap ->> oap: Checks that the user is authorized<br>based on e.g. group, email domain, or organization
oap ->> redirect: Redirect authenticated user to upstream at /redirect/ (=redirect sidecar)
deactivate oap
redirect ->> resource: Redirect to protected resource
{{< /mermaid >}}

## Configuring ```oauth2-proxy```

Many of the resources here contain environment variables with ```$VAR```-type syntax. ```envsubst``` is used during the deployment flow to add these values from specified environment secrets following the convention below.

```shell
cat example-manifest.yml | envsubst '${VAR}'| kubectl apply -f -
```

Do **NOT** use ```envsubst``` by itself without any arguments, otherwise anything with a ```$```-sign in your files will be replaced as well. It is much safer to specify each variable you want to replace instead.

### Resource definitions: ```Secret```

This resource contains the client id and secret from the authentication provider, configures the session cookie secret, and sets the redis password. While you could provide these as raw environment variables, or as config arguments in the deployment, providing them as secrets brings better protection as otherwise anyone with list/describe-level access to the deployment resource could read these values directly.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: oauth2-proxy-secret
type: Opaque
stringData:
  client-id: $OAUTH2_PROXY_CLIENT_ID # as given by your auth provider
  client-secret: $OAUTH2_PROXY_CLIENT_SECRET # as given by your auth provider
  cookie-secret: $OAUTH2_PROXY_COOKIE_SECRET # generated by you
  redis-password: $REDIS_PASSWORD # generated by you
```

### Resource definitions: ```ConfigMap``` for redirect sidecar

This resource provides the NGINX configuration for the sidecar container we will deploy with the authentication proxy. Its main purpose is to instruct the server to listen at ```/redirect/```, and redirect the client to the URL following that path. This snippet is directly from [Callum Pember's article](https://www.callumpember.com/Kubernetes-A-Single-OAuth2-Proxy-For-Multiple-Ingresses/).

```yaml
# Config map for sidecar nginx to just act as a redirect service
apiVersion: v1
kind: ConfigMap
metadata:
  name: oauth2-proxy-nginx
data:
  nginx.conf: |
    worker_processes 5;
    events {}
    http {
      server {
        listen 80 default_server;
        location = /healthcheck {
          add_header Content-Type text/plain;
          return 200 'ok';
        }
        location ~ /redirect/(.*) {
          return 307 https://$1$is_args$args;
        }
      }
    }
```

### Resource definitions: ```Deployment```

The snippet for the sidecar container part is directly from [Callum Pember's article](https://www.callumpember.com/Kubernetes-A-Single-OAuth2-Proxy-For-Multiple-Ingresses/). Several important parts are highlighted, including the configurations that enable the sidecar container, broaden the cookie domain scope, and configure redis for session storage.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: oauth2-proxy
spec:
  replicas: 1
  template:
    spec:
      volumes:
        - name: nginx
          configMap:
            name: oauth2-proxy-nginx
      containers:
        - name: oauth2-proxy
          image: quay.io/oauth2-proxy/oauth2-proxy:v7.4.0
          imagePullPolicy: Always
          args:
          # This enables the redirect container
          - --upstream=http://localhost/redirect/

          # This is where we expand the domain of the cookie
          # Note the period at the start of the value!
          - --cookie-domain=.aviitala.com
          - --whitelist-domain=.aviitala.com
          - --cookie-expire=12h
          - --http-address=0.0.0.0:4180

          # Auth provider details, e.g. for Azure
          - --provider=oidc
          - --email-domain=aviitala.com
          - --provider-display-name=AVIITALA
          # Configure this based on your Azure tenant ID value
          - --oidc-issuer-url=https://login.microsoftonline.com/{TENANT_ID_VALUE}/v2.0
          - --redirect-url=https://auth.aviitala.com/oauth2/callback


          # Cookie storage settings - enable redis
          - --session-store-type=redis
          # Default (internal) redis connection URL within the cluster
          - --redis-connection-url=redis://redis-master.oauth2-proxy.svc.cluster.local:6379

          # Bring in the secret values defined earlier as protected environment variables for this deployment
          env:
          - name: OAUTH2_PROXY_CLIENT_ID
            valueFrom:
              secretKeyRef:
                name: oauth2-proxy-secret
                key: client-id
                optional: false
          - name: OAUTH2_PROXY_CLIENT_SECRET
            valueFrom:
              secretKeyRef:
                name: oauth2-proxy-secret
                key: client-secret
                optional: false
          - name: OAUTH2_PROXY_COOKIE_SECRET
            valueFrom:
              secretKeyRef:
                name: oauth2-proxy-secret
                key: cookie-secret
                optional: false
          - name: OAUTH2_PROXY_REDIS_PASSWORD
            valueFrom:
              secretKeyRef:
                name: oauth2-proxy-secret
                key: redis-password
                optional: false
          ports:
          - containerPort: 4180
            protocol: TCP
        
      # sidecar container to handle redirects - snippet directly from Callum Pember
        - name: nginx
          image: nginx:1.23.2-alpine
          imagePullPolicy: Always
          resources:
            limits:
              cpu: 0.2
              memory: 512Mi
          ports:
            - name: nginx
              containerPort: 80
          volumeMounts:
            - name: nginx
              mountPath: /etc/nginx/
              readOnly: true
          livenessProbe:
            httpGet:
              path: /healthcheck
              port: 80
            initialDelaySeconds: 3
            timeoutSeconds: 2
            failureThreshold: 2
```

### Resource definitions: ```Service``` and ```Ingress```

The service resource definition is completely standard and does not need further editing. The ingress resource definition will require an update with the domain relevant to you.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: oauth2-proxy
  annotations:
    kubernetes.io/ingress.class: nginx

    # options to use cert-manager to get TLS certificates for your host
    kubernetes.io/tls-acme: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - auth.aviitala.com # update this value with your domain
    secretName: oauth-tls-secret
  rules:
  - host: auth.aviitala.com # update this value with your domain
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
# leave as-is
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
  selector:
    app: oauth2-proxy
```

## Configuring ```redis```

For ```redis``` I opted to use an existing Helm chart from Bitnami [here](https://charts.bitnami.com/bitnami). The installation and configuration is as simple as:

```shell
helm repo add redis https://charts.bitnami.com/bitnami | helm repo update
helm upgrade --install redis redis/redis \
--values redis-values.yml \
--set auth.password=${{ secrets.REDIS_PASSWORD }} \ # to be added by your deployment flow
--create-namespace --namespace oauth2-proxy # install in same namespace as oauth2-proxy
```

The ```redis-values.yml``` file is shown below. Almost everything is left as default; just the ```architecture``` configuration option is set to ```standalone```. Given that our use case is not mission-critical, we can simplify the deployment this way. The ```password``` value is included in this file as a placeholder, as the helm installation command from above, with the ```--set``` argument, will override the input provided by ```redis-values.yml```.

```yaml
architecture: standalone # disable replication, run as a single instance
auth:
  enabled: true
  sentinel: true
  password: "will be replaced by command line value"
```

## OAuth2 provider configuration in Azure

Even though our ```oauth2-proxy``` configuration uses the ```oidc``` provider here instead of the Azure provider, the setup steps within Azure Active Directory are more or less the same. They are documented [here](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/oauth_provider/#azure-auth-provider).

The key part is to add our authentication endpoint as the **Redirect URI**, as shown below:

![redirect uri configuration in azure ad](/content/azure-redirect-uri.png)

## Example annotations to protect an ingress with OAuth

After the above is configured, you can protect a particular ingress by adding the annotations below. This is a great example where careless use of ```envsubst``` during deployment can really screw things up as the manifest has a ```$```-sign-prefixed value that should *not* be replaced. I learned this the hard way and spent hours debugging the issue, hopefully you won't have to do the same.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: your-sample-application
  annotations:
    kubernetes.io/ingress.class: nginx

    # the most important annotations
    nginx.ingress.kubernetes.io/auth-url: "https://auth.aviitala.com/oauth2/auth"
    nginx.ingress.kubernetes.io/auth-signin: "https://auth.aviitala.com/oauth2/start$request_uri"

    # options to use cert-manager to get TLS certificates for your host
    kubernetes.io/tls-acme: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ... # rest of the ingress spec
```

## References

* [OAuth2-proxy configuration with nginx-ingress](https://aviitala.com/posts/nginx-ingress-oauth2/)
* The 'redirect sidecar' idea: [Single OAuth2 proxy for multiple ingresses](https://www.callumpember.com/Kubernetes-A-Single-OAuth2-Proxy-For-Multiple-Ingresses/)
* [OAuth2-proxy docs on session storage](https://```oauth2-proxy```.github.io/```oauth2-proxy```/docs/configuration/session_storage/)
* [Which redis architecture to use for sessions?](https://stackoverflow.com/questions/53060714/redis-sentinel-standalone-or-cluster-which-is-best-for-session)

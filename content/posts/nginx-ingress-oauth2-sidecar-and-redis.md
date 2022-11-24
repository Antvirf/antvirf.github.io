+++
author = "Antti Viitala"
title = "Multi-domain OAuth2 Proxy configuration with Redis cookie storage"
date = "2022-11-30"
description = "Building on top of the basics, this article describes an AKS cluster configuration using nginx-ingress and OAuth2 proxy - with an NGINX sidecar - to enable serving multiple subdomains from a single authentication proxy. Session cookie storage is implemented with Redis, as some OIDC providers like Azure create and send huge cookies that are too large for many web servers (including NGINX) by default."
tags = [
    "kubernetes",
    "infrastructure",
    "devops",
    "oauth"
]
images = ['images/apple-touch-icon-152x152.png','images/splash.png']
+++

*[See previous article on OAuth2 Proxy configuration with nginx-ingress.](https://aviitala.com/posts/nginx-ingress-oauth2/)*

## Motivation - Multiple subdomains, single authentication proxy

An organization will have its 'main' domain, e.g. ```aviitala.com```, where the primary internet-facing website is stored. Oftentimes we would also want to host product demo pages and internal development environments in various subdomains - e.g. ```demo.aviitala.com```, ```dev.aviitala.com``` and so on. As you automate the deployment of various development environments, it is easy to end up with more and more subdomains - ```product1-branch3.dev.aviitala.com```, ```product2-branch4.dev.aviitala.com```. For ease of development and sharing with external parties, we would want keep these domains accessible externally, but simultaneously want to add a layer of protection to ensure only trusted parties can access our resources.

With its default configuration, an instance of ```oauth2-proxy``` would need to be configured *for each domain name*. As your domains may be dynamic based on products and their development processes, this is not acceptable and would lead to tens if not hundreds of replicated OAuth2 proxies running on your cluster.
There are two limitations. First, by default the authentication cookie created is only valid for a single domain. Second, as you configure an authentication provider, you need to provide a *single* redirect URL for the application where the user is redirected upon a successful login.

The first issue can be addressed easily, by adjust the configuration of ```oauth2-proxy``` to widen the scope of the authentication cookie. The second part requires more creativity and was solved by Callum Pember in [his great article here](https://www.callumpember.com/Kubernetes-A-Single-OAuth2-Proxy-For-Multiple-Ingresses/). Essentially, we can configure the proxy to deploy an additional sidecar container that redirects each successful authentication request to the relevant resource the user came from. We then point the authentication provider's callback URL to that of the ```oauth2-proxy```, and point the upstream configuration parameter to the redirection endpoint. On a successful login, the client is redirected to this 'redirect sidecar', and from there redirected instantaneously to the correct resource/domain the client started the login at.

## Motivation - Using ```redis``` for session cookie storage

Most tutorials for ```oauth2-proxy``` use GitHub, but for this use case we wanted to use Azure as the OAuth2 provider. There is an Azure-specific provider created within ```oauth2-proxy``` (and is useful if you need group/role information etc.), but this configuration uses the standard OIDC provider, just configured to Azure. After setting this up initially, the login process and authentication all work perfectly fine - but the page load of an actual application behind the proxy often failed. The front-end apps running were being served with NGINX, and the cookie size that Azure sends is too large for the default NGINX configuration. This lead to error message ```400 Bad Request - Request header or cookie too large```.

This could be solved *within each nginx instance behind the auth proxy* by increasing the allowed cookie and header sizes ([example](https://stackoverflow.com/questions/17524396/400-bad-request-request-header-or-cookie-too-large)), but this would require the application owners to know and care about this issue. Since the aim was to have a smooth developer experience where a basic app served with a default nginx image would 'just work', changing to ```redis```-based session storage (which is well supported by ```oauth2-proxy```) was the better alternative.

<!-- some notes on config -->

## Infrastructure configuration

* Running Kubernetes cluster, with the following components installed:
  * [nginx-ingress](https://github.com/kubernetes/ingress-nginx) - ingress controller
  * [external-dns](https://github.com/kubernetes-sigs/external-dns) - handles creation of DNS records in our subdomain
  * [cert-manager](https://github.com/cert-manager/cert-manager) - automates TLS certificate creation for our ingresses

## Sequence diagram of the relevant components

### Unauthenticated / first-login flow

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

### Authenticated user flow

<!-- Can I combine these charts into one? -->

{{< mermaid >}}
sequenceDiagram
autonumber
participant user as User
participant ingress as NGINX Ingress
participant oap as OAuth2 Proxy
participant idp as Auth provider<br>(Microsoft Azure)
participant redirect as OAuth2-Proxy<br>Redirect sidecar
participant redis as OAuth2-Prody<br>Redis Cookie store
participant resource as Protected Resource

user ->>+ingress: Request to /protected/<br>with an authentication cookie
ingress ->>ingress: Ingress checks that auth-url<br>and auth-sign annotations are<br>present for the requested route
oap ->> redis: Request full session cookie from Redis with the key given by the client
redis ->> oap: Send back full session cookie, if exists. If does not exist, start auth process at/oauth2/auth
oap ->> oap: Check that the token is not expired, and covers the requested resource's subdomain
oap ->> oap: Checks that the user is authorized<br>based on e.g. group, email domain, or organization
oap ->> redirect: Redirect authenticated user to upstream at /redirect/ (=redirect sidecar)
redirect ->> resource: Redirect to protected resource
{{< /mermaid >}}

## Configuring ```oauth2-proxy```

## Configuring ```redis```

## OAuth2 provider configuration



## References

* [OAuth2-proxy configuration with nginx-ingress](https://aviitala.com/posts/nginx-ingress-oauth2/)
* The 'redirect sidecar' idea: [Single OAuth2 proxy for multiple ingresses](https://www.callumpember.com/Kubernetes-A-Single-OAuth2-Proxy-For-Multiple-Ingresses/)
* [OAuth2-proxy docs on session storage](https://```oauth2-proxy```.github.io/```oauth2-proxy```/docs/configuration/session_storage/)

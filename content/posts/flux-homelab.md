+++
author = "Antti Viitala"
title = "Self-hosted Kubernetes homelab with K3s, Flux GitOps and ngrok"
date = "2023-07-27"
description = "Guide on setting up a self-hosted, GitOps-managed Kubernetes cluster that has the ability to host publicly accessible services with NGROK. Uses K3s for Kubernetes, NGROK ingress controller for public access, and Flux to deploy your applications from GitHub container registry."
tags = [
    "kubernetes",
    "infrastructure",
    "devops"
]
images = ['content/ngrok-k3s-flux.png']
+++

## Purpose

Cloud infrastructure is nice and all, but sometimes you just want a simple kubernetes cluster to play around with as cheaply as possible, taking advantage of some hardware you already have laying around.

This article describes a setup that:

- Costs nothing if you already have the hardware (except your existing internet and electricity bills)
- Runs probably on any machine with Ubuntu
- Gives you a real ([k3s](https://k3s.io/)) Kubernetes environment that...
  - can **expose services to the public internet**
  - can be turned off easily without any hassle or losing data
  - can be turned on with services up within ~2 minutes
- Enables continuous deployments and GitOps with Flux

The instructions are relatively high level and may require special tuning for your setup. Further reading on some of the components is also encouraged. A sample repository is available [here](https://github.com/Antvirf/homelab-flux-ngrok-example).

## Key components and how they work

<details>
<summary> Why Flux?</summary>

[Flux](https://fluxcd.io/) is a GitOps toolkit for continuous deployments.

### GitOps

There are [several advantages to GitOps](https://opengitops.dev/), but in this scenario we benefit the most from automatic pull-based deployments (see below for more detail on that) and having a declarative approach to managing the contents of the cluster.

Especially if you play around a lot with different applications that you install and remove from your cluster, it is easy to forget what's actually there if you come back to it months later. When the contents of the cluster are declared in your git repo, there is no confusion on the state of the cluster thanks to continuous reconciliation.

### Flux gives us pull-based deployments (instead of push-based)

Despite using Ngrok, our cluster still has no public Kubernetes API endpoint. This means that a GitHub Actions workflow for example cannot run `kubectl apply` commands against our cluster.

This is great for security, and for this reason many tightly-controlled environments anyway use pull-based deployments. Flux handles this for us nicely.

### How does Flux work?

There are a lot of components to [Flux](https://fluxcd.io/), but the easiest way to think about it for our scenario here is like this:

| [Image controller](https://fluxcd.io/flux/components/image/) | [Source controller](https://fluxcd.io/flux/components/source/) |
| -- | -- |
A tiny app that:<br>- Continuously checks a desired container registry for new image tags<br>- If a new tag is found, updates a kubernetes manifest in a desired GitHub repository with this new container image| A tiny app that:<br>- Continuously pulls the latest manifests from a desired repository<br>- "Reconciles" the state by applying these manifests to the cluster.

Thinking about this from a deployment perspective, the image controller will push commits like this to your repo that update a container tag:

```yaml
-- image: ghcr.io/antvirf/example-image:11
++ image: ghcr.io/antvirf/example-image:12
```

The next time the source controller pulls the repo, it sees a file has changed with this new container tag, and applies this to the cluster.

</details>

<details>
<summary> Why ngrok?</summary>

One of the main challenges of a self-hosted homelab-style setup is exposing services - let's say a website - to the public internet. Most internet service providers allocate regular users dynamic IPs that change time to time, so pointing domain name records to our own IPs is problematic. Beyond this your local home router will need some adjustment to open or forward particular ports to reach your servers, and this process tends to be manual and annoying, as well as potentially problematic from a security perspective.

[ngrok](https://ngrok.com/) solves this problem nicely by creating a tunnel from your local machine to an edge network managed by ngrok. DNS and certificates are managed for you, and the connection lasts only as long as the tunnel is open - shutting down the service closes the tunnel. Just about a month ago, ngrok [announced](https://ngrok.com/blog-post/ngrok-k8s) their own [ngrok kubernetes ingress controller](https://github.com/ngrok/kubernetes-ingress-controller), which brings this functionality to kubernetes. Services in the cluster can now be exposed cleanly via ngrok tunnels without the need of figuring out how to make load balancers work on your local machine.

</details>

## Networking and access pattern with k3s and ngrok ingress controller

{{< mermaid >}}

flowchart TD

subgraph local["Local Ubuntu server"]
    systemd["systemd"]
    subgraph k3s["k3s Kubernetes Cluster"]
        flux["flux controllers"]
        ngi["ngrok ingress controller"]
        subgraph app["Your app"]
            deployment
            service
            ingress
        end
    end
end

subgraph ng["Ngrok"]
    ngrok["Ngrok edge"]
end

systemd --> |Ensures cluster is always running| k3s
flux --> |"Install and manage the application"|ngi
flux --> |"Install and manage the application"|app

ngi --> |Configures ngrok\nbased on ingress objects| ngrok
ngi --> |Monitors ingress objects| ingress
ingress <--> |Application is served via ngrok tunnel| ngrok
ngrok --> |Services accessed via ngrok| user["Users"]

{{< /mermaid >}}

## Setup

### Step 0: Preparation

- **Hardware**: The guide assumes you have a machine with a recent installation of Ubuntu.
- **ngrok**: You need to sign up for a free ngrok account and claim your [free domain name](https://dashboard.ngrok.com/cloud-edge/domains).

### Step 1: Install K3s

K3s doesn't need much configuration; we just pass the option to disable Traefik as we will be installing our own ngrok-based ingress controller instead.

```bash
curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="server --disable traefik" sh
```

### Step 2: Set up Flux

The Flux installation here is also very much standard, following the [instructions provided by the project](https://fluxcd.io/flux/installation/). The below provides a quick summary, assuming you already have the Flux CLI installed. You also need a [GitHub access token.](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-personal-access-token-classic)

```bash
# Set up your GH token
export GITHUB_TOKEN=<your-token>

# Bootstrap Flux - customise first 3 arguments to your setup
flux bootstrap github \
    --owner=your-github-username \ 
    --repository=your-repo-name \
    --path=clusters/homelab \
    --private=true --personal=true \
    --components-extra=image-reflector-controller,image-automation-controller \ # you want these for automating deployments
    --read-write-key # this is needed so Flux can update your repo
```

### Optional step: Install `sealed-secrets`

In order to do things the "GitOps" way, your repository needs to be able to declaratively define the secrets your setup needs. The ngrok ingress controller will need your ngrok access token in order to communicate with their network, and likely you will also need to store credentials to a container registry such as GHCR.

A more 'advanced' setup may to use something like [external-secrets](https://github.com/external-secrets/external-secrets) and store the actual values in e.g. AWS Secrets Manager, but for a simple setup [sealed-secrets](https://github.com/bitnami-labs/sealed-secrets) works nicely. It encrypts secrets using a private key known only by your cluster, so that they can be safely committed in Git. The below expand contains the Flux manifests to deploy `sealed-secrets`.

Please follow the instructions provided by `sealed-secrets` as the service consists of both a client-side as well as a cluster-side component. The manifests below install the cluster-side components for you with Flux. You can also refer to the example repository, which contains the [application definition](https://github.com/Antvirf/homelab-flux-ngrok-example/tree/main/applications/sealed-secrets), [flux deployment manifest](https://github.com/Antvirf/homelab-flux-ngrok-example/blob/main/clusters/homelab/cluster-system/sealed-secrets.yaml), as well as a [makefile-based utility](https://github.com/Antvirf/homelab-flux-ngrok-example/blob/main/makefile) to help with creating sealed secrets conveniently.

<details>
<summary> Open application definition manifests</summary>

```yaml
---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: sealed-secrets
  namespace: flux-system
spec:
  interval: 10m0s
  url: https://bitnami-labs.github.io/sealed-secrets
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: sealed-secrets
  namespace: flux-system
spec:
  chart:
    spec:
      chart: sealed-secrets
      reconcileStrategy: ChartVersion
      sourceRef:
        kind: HelmRepository
        name: sealed-secrets
      version: 2.11.0
  interval: 10m0s
```

</details>

### Step 3: Install ngrok ingress controller with Flux

Installation of ngrok ingress controller can also be done via helm as shown below. [Application definition manifests](https://github.com/Antvirf/homelab-flux-ngrok-example/tree/main/applications/ngrok-ingress-controller) as well as [flux deployment manifests](https://github.com/Antvirf/homelab-flux-ngrok-example/blob/main/clusters/homelab/cluster-system/ngrok-ingress.yaml) are available in the example repository.

<details>
<summary> Open application definition manifests</summary>

```yaml
---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: ngrok-ingress-controller
  namespace: flux-system
spec:
  interval: 10m0s
  url: https://ngrok.github.io/kubernetes-ingress-controller
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: ngrok-ingress-controller
  namespace: flux-system
spec:
  chart:
    spec:
      chart: kubernetes-ingress-controller
      reconcileStrategy: ChartVersion
      sourceRef:
        kind: HelmRepository
        name: ngrok-ingress-controller
      version: 0.10.0
  interval: 10m0s
  values:
    credentials:
      secret:
        name: ngrok-ingress-controller-credentials
```

</details>


### Step 4: Install your applications

Depending on how you configured Flux, you will need to set up deployment manifests in the right folder - in the example case in `./clusters/homelab/`.

This step is entirely specific to what you want to install and how you wish to go about it. The example repository uses my [kube-ingress-dashboard](https://github.com/Antvirf/kube-ingress-dashboard) project.

### Step 5: Expose the application with an ngrok ingress

The ingress objects expected by the ngrok ingress controller are relatively standard; the only thing you need to pay special attention to is to set the `ingressClassName` to `ngrok`, and provide your free ngrok domain name as host.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app
spec:
  ingressClassName: ngrok
  rules:
    - host: YOUR_URL # replace this value with your ngrok domain name
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-app
                port:
                  number: 8000
```

After giving ngrok a second to synchronise, you should now be able to reach your application from your ngrok domain 😎

![example](/content/kube-ingress-dashboard.png)

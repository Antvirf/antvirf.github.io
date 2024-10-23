+++
author = "Antti Viitala"
title = "Comparing Flux and Argo in real life"
date = "2024-10-23"
description = "Some thoughts and feelings after using Flux and Argo for a couple of years, and the main reasons I prefer Flux."
tags = [
    "kubernetes",
    "infrastructure",
    "devops",
    "gitops",
    "flux",
]
images = ['content/flux.png']
+++

When I did initial research on these two GitOps tools a couple of years ago, I went through several comparison articles that looked at the key features - often in the form of a table. Today, neither has glaring omissions, and you can't go wrong picking either one - but what I thought was missing from the comparisons was some perspective on how it "feels" to use the two tools, what their main differences are in actual daily usage. Here are my thoughts after using Flux for a couple of years and Argo for the past year.


## Overview of how Flux and Argo store state and how they expose it

A 'complete' Flux installation comes with a large number of CRDs - `GitRepository`, `HelmRepository`, `Kustomization`, `HelmRelease`, `HelmChart`, and many more. Argo, on the other hand, installs only three - `AppProject`, `Application` and `ApplicationSet`. Flux containers are largely stateless and store information in the CRDs; Argo also stores some information in the CRDs but brings its own Redis instance.

Because of the (semi-)independent nature of each of the Flux controllers, the way they work together is almost entirely via the custom resource definitions, and this, in my opinion makes understanding the state of the system straightforward - if I want to know anything about the system, I look at the resources relevant to that CRD. This is how Kubernetes itself works, resources interacting with controllers. With Argo, observing the state of the system is almost always done in the UI, or by viewing the logs of the relevant Argo containers.

##  Basic troubleshooting scenario 

As an example, imagine the following scenario:

1. You want to add a new Git repository to sync from to your cluster;
1. The new repository contains a deployment manifest like `Application` for argo, or `HelmRelease`+`HelmRepository` for Flux;

In Flux, in terms of resources, we would create:

- A standard `Secret` resource with e.g. Git SSH or user/pass credential - [ref](https://fluxcd.io/flux/components/source/gitrepositories/#secret-reference)
- `GitRepository` resource that has the URL of the repo, and refers to the secret created above - [ref](https://fluxcd.io/flux/components/source/gitrepositories/)
- A `HelmRepository` pointing to the relevant Helm repository
- A `HelmRelease` which Flux then uses to install the app (specifies the chart, chart version, and `values`)

In Argo, we would create:

- A standard `Secret` with the credentials to the desired repo - [ref](https://argo-cd.readthedocs.io/en/stable/operator-manual/declarative-setup/#repository-credentials)
- An `Application` resource (see [example](https://github.com/argoproj/argo-cd/blob/master/docs/operator-manual/application.yaml)) that:
  - Defines the URL of the git repository in question
  - Defines the relevant Helm repository
  - Defines the desired chart, chart version, and `values`

Because Flux splits the parts of the GitOps process transparently between different custom resources, it is easy for an administrator to observe what part of the process is failing - and just as straightforward to implement a precise fix. As an example, let's take a look at a few common problems, as well as how one would spot them and how one would address them:

1. incorrect Helm repo reference, e.g. repo doesn't exist;
2. attempt to install a Helm chart version that doesn't exist;
3. manifests of the chart themselves have a syntax error that prevents an `apply` from succeeding;

In Argo, there are only two places to spot issues: The UI, or looking at the `Application` CRD. If 'something' is broken, the `Application` in question is always in an error state - you then have to inspect relevant events to it to find out what's wrong. Seeing that an `Application` is out of sync or in unhealthy status tells you very little by itself. Fixing (1) and (2) both require editing the `Application` resource; (3) needs to be fixed outside Argo but its errors also show up in `Application`.

In Flux, troubleshooting follows the natural steps of the GitOps process. To observe and to fix (1), you look at `HelmRepository`, since that is the CRD responsible for interacting with Helm repositories. To observe and to fix (2), you look at `HelmRelease`, since it is responsible for installing a chart from a repo. To observe (3) you look at `HelmRelease`, since that is the CRD responsible for installing the chart to your cluster.

The CRD-based approach by Flux is robust and in my opinion makes the troubleshooting process more straightforward, as it almost provides you a 'runbook' of independent components that you need to check and get working.

## Self-service for developers with GitOps

> As a platform engineer, I want to allow developers to deploy microservices to a GitOps-managed cluster via self-service, without expecting them to be Kubernetes experts. Such apps need to have fully automated deployments when developers push new containers to the container registry.

There are different ways to approach this, but my personal preference is to have a generic Helm chart that allows developers to set the image they want to deploy, configure env vars, fetch secrets easily from some central secret management system, set domain names, etc. Helm charts can be greatly simplified to improve this experience. No matter what though, a developer will have to write some YAML to get a chart deployed with their own parameters, so here we look at what such a YAML file could look like.

### Part 1: Self-service via a Helm chart

Ignoring automated deployments of new containers for now, let's first compare how the two tools would achieve the task of deploying a Helm chart to a cluster. Assume that the secrets for the required repositories (both Git and Helm) have already been configured, and for Flux we would accordingly have the shared resources (`HelmRepository`) already created. I highlight below as potential 'footguns' any field that a developer has to change or update themselves for each deployment.

With Argo, we would create a single `Application` resource:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-example-app-development # A
  namespace: argocd # B
spec:
  destination:
    namespace: my-example-app-namespace # C
    server: TARGET_CLUSTER_URL # C
  project: default # C
  source:
    chart: example-chart
    repoURL: https://example.github.io/example-chart
    targetRevision: 1.16.1
    helm:
      releaseName: my-release
      valuesObject:
        imageName: ghcr.io/example/example:3.0.0
        env:
          MY_ENV_VAR: 123
        envSecret:
          MY_SECRET: /path/to/secret/in/vault
```

Not too bad - we can keep it quite simple, but we have a few potential sources of trouble:

- A: Argo application names have to be **globally unique**, so if you have a multi-cluster Argo installation, you can cause some very confusing errors by creating applications with the same name. As an added complexity, this basically means that the environment name has to be a part of the application name, for example via a suffix like here, to avoid such conflicts.
- B: Changing the namespace of the `Application` could result in it not being picked up (looking in other namespaces is an [opt-in feature in beta state](https://argocd-operator.readthedocs.io/en/latest/usage/apps-in-any-namespace/)).
- C: Needing to specify a target cluster and project is additional complexity developers now need to know about.

With Flux, we would create the following minimal `HelmRelease` resource:

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: my-example-app
  namespace: my-example-app # can be freely set
spec:
  chart:
    spec:
      chart: example-chart
      version: 1.16.1
      sourceRef:
        kind: HelmRepository
        name: shared-helm-repo
  releaseName: my-release
  values:
    imageName: ghcr.io/example/example:3.0.0
    env:
      MY_ENV_VAR: 123
    envSecret:
      MY_SECRET: /path/to/secret/in/vault
```

There are no footguns here. Comparing to Argo's points from above;

- A: No need for `.metadata.name` to be globally unique, just unique within this cluster and namespace, *like any other resource*.
- B: No need for any specific namespace, team/dev is free to manage, *like any other resource*.
- C: Destination cluster is **defined by the folder of the file**, based on what Flux is looking at, so a single central repo could have a folder for the contents of each cluster.

In my opinion, the manifest with Flux has less footguns and is more straightforward.

### Part 2: Automate deployments of new containers

Continuing with our scenario, a crucial requirement here is that once the YAML file is submitted, there has to be a process that handles picking up new containers and deploying them in certain environments. To this end, Flux provides [Image Update Automations](https://fluxcd.io/flux/components/image/imageupdateautomations/) and Argo provides the [ArgoCD Image Updater](https://argocd-image-updater.readthedocs.io/en/stable/). The example YAMLs below assume each cluster for that example would have that component installed and configured.


With ArgoCD, the way image update automations are controlled are via annotations to the `Application` resource. The manifest from above would then become something like this:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-example-app-development
  namespace: argocd
  annotations: # added with Image Updater
    argocd-image-updater.argoproj.io/image-list: my_image=ghcr.io/example/example:3.0.0
    argocd-image-updater.argoproj.io/write-back-method: git
    argocd-image-updater.argoproj.io/portal_production.update-strategy: digest
spec:
  destination:
    namespace: my-example-app-namespace
    server: TARGET_CLUSTER_URL
  project: default
  source:
    chart: example-chart
    repoURL: https://example.github.io/example-chart
    targetRevision: 1.16.1
    helm:
      releaseName: my-release
      valuesObject:
        imageName: ghcr.io/example/example:3.0.0
        env:
          MY_ENV_VAR: 123
        envSecret:
          MY_SECRET: /path/to/secret/in/vault
```

We add in at least three annotations. Arguably only one, the `image-list`, requires the developer to update it. (Note, however, that the Argo image updater **must be centrally updated/configured for auth to each container registry it should have access to** - there is no other way to add credentials to it. This also means that the single credential configured for the image updater is a single point of failure.)

How about with Flux?

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: my-example-app
  namespace: my-example-app
spec:
  chart:
    spec:
      chart: example-chart
      version: 1.16.1
      sourceRef:
        kind: HelmRepository
        name: shared-helm-repo
  releaseName: my-release
  values:
    imageName: ghcr.io/example/example:3.0.0 # {"$imagepolicy": "my-example-app:my-example-app"}
    env:
      MY_ENV_VAR: 123
    envSecret:
      MY_SECRET: /path/to/secret/in/vault
```

We've added in **one** comment, the rest of the file is unchanged. Instead, behind the scenes, our Helm chart will need to create the three resources that handle the updates - `ImageRepository`, `ImagePolicy`, and `ImageUpdateAutomation`. There is no doubt that Flux doing this configuration via a comment is rather ugly.

#### How do the image update processes differ?

The manifests look a little bit different, fine, but what about the actual process itself - what happens once these manifests are pushed to the cluster and new container images are pushed to the registry?

- Flux: `ImageRepository` scans the registry, `ImagePolicy` determines the relevant tag to use, `ImageUpdateAutomation` pushes a commit to the repo that **changes the line with the `$imagepolicy` comment**.
- Argo: The image updater container checks the registry, determines the relevant tag and pushes a commit to the repo - **to an arbitrary file called `.argocd-source-<appName>/yaml`** ([ref](https://argocd-image-updater.readthedocs.io/en/stable/basics/update-methods/#git-write-back-target))

If this didn't sound odd to you, read the bullet points again. Argo's solution to this requirement is to push *an additional file separate from the deployment file* to the repo that will control what image is used in the cluster. Now, your original deployment file will still say that version `3.0.0` is deployed, even though Argo would then use this separately created file that may already be on `4.0.0`. The deployment file won't accurately represent the real state of what is deployed, and now a developer would have to know to go check this seemingly random file for what image is used in the cluster. The situation is a little better if you are using Kustomize, but that adds other footguns and makes self-service more difficult.

With Flux, the file that defines the `HelmRelease` to be deployed - the file a developer wrote and pushed themselves - is the file that gets updated, and is the single source of truth for what should be deployed on the cluster.

#### Troubleshooting image update processes

Because of Flux's CRD approach, troubleshooting is again straightforward - check `ImageRepository` first to see if the controller can talk to your registry; check `ImagePolicy` to see if it's making the right decision about which tag to use, and check `ImageUpdateAutomation` to see if the pushes to Git are working. Each chart deploys their own resource of each kind, so knowing where to check is self-evident.

With Argo image updater, logs of that container are your only source of any information. In a multi-cluster environment, this is a problem because developers likely won't have access to Argo logs. Any problem with image updates immediately becomes an ops/devops/infra problem, nobody else has access. Assuming you have access, you will then have to parse through the logs - because every single application managed by that instance using automatic image updates will have the logs of that update process go to the same place. Doable, but a lot of (not enjoyable) work.

#### Closing thoughts on image updates

My experience with argocd-image-updater was so bad that despite a platform as a whole using Argo, I brought in Flux just to have its image update automations. In a few words, the parts about Argo image updater that made it unusable in my opinion were:
- Not being able to update the file where an `Application` is defined/deployed with the right version for an image, but needing to use a separate file for that. Makes understanding desired cluster state confusing.
- Configuration via annotations make it impossible to hide the complexity; with Flux I can create the required resources with the Helm chart itself, along with the app.
- Troubleshooting only via logs is painful compared to Flux's CRD based approach, where I can pin down the issue with three `kubectl describe` commands of the relevant image update resources.
- Odd experiences that made the tool feel extremely early stage, e.g. the Argo image updater could not deal with multiple updates to the same branch of the same repo at the same time, which in a centralized GitOps repo with a lot of apps happens multiple times a day.


### Part 3: Declaratively adding a new notifications endpoint to a deployment

Perhaps as part of our deployment, we would want to configure notifications to the team's Pagerduty about status of the application, so that they are aware of any synchronization issues or can be notified of new deployments.

With Argo, to make this pattern workable at all (de-centralized control of notifications), you will need to enable the [applications in any namespace](https://argo-cd.readthedocs.io/en/stable/operator-manual/app-any-namespace/) feature. Once done, to get this working you would now need to create a `Secret` for credentials/keys, a `ConfigMap` to configure the notifications service, and then add an annotation to your application. The below examples are directly from [Argo docs on notifications](https://argo-cd.readthedocs.io/en/stable/operator-manual/notifications/), creation of the `Secret` is excluded as this is identical between Flux and Argo:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
data:
  service.pagerdutyv2: |
    serviceKeys:
      my-service: $pagerduty-key-my-service
```

And in your application;

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  annotations:
    notifications.argoproj.io/subscribe.on-sync-failed.pagerdutyv2: "<serviceID for Pagerduty>"
    ...
```

In Flux, following the Flux approach of doing everything with CRDs, you would need to configure a `Provider`, and an `Alert`, like so (example from [Flux docs on notifications](https://fluxcd.io/flux/monitoring/alerts/)):

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta3
kind: Provider
metadata:
  name: my-example-app-pd-provider
  namespace: my-example-app
spec:
  type: pagerduty
  channel: general
  address: https://slack.com/api/chat.postMessage
  secretRef:
    name: slack-bot-token
---
apiVersion: notification.toolkit.fluxcd.io/v1beta3
kind: Alert
metadata:
  name: my-example-app-alert
  namespace: my-example-app
spec:
  providerRef:
    name: my-example-app-pd-provider
  eventSeverity: info
  eventSources:
    - kind: GitRepository
      name: '*'
    - kind: Kustomization
      name: '*'
```

The difference in approaches between the tools is apparent, and follows the same trends as before. Argo requires the developer to know how to configure Argo, since the notification subscriptions are defined in the annotations of the deployment file. As before, troubleshooting always happens by looking at logs of Argo pods.

In contrast, configuring notifications with Flux is done with CRDs, and because they are "just resources", the platform engineer can hide all this complexity in the Helm chart. Troubleshooting a notifications `Provider`, or a specific `Alert`, can be done quickly by focusing directly on the relevant resources.

## Closing thoughts

Some software you like and admire the more you use it; some you dislike the more you use it. For me, Flux is definitely the prior, and Argo definitely the latter. The clear design of Flux as a piece of software is impressive in my opinion, and working with it has been enjoyable. Its CRDs-and-controllers approach make it feel like a natural extension of Kubernetes. In comparison, Argo feels like a complex and fragile black box (though with a fantastic UI!), and often lacks configurability where I would want it (e.g. setting app-level synchronization frequency), just to then offer flexibility in areas that make no sense to me (automatic sync is not the default, even though in my opinion this is what GitOps is fundamentally about).

Flux all the way ❤️


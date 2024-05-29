+++
author = "Antti Viitala"
title = "Drawing the line between infrastructure and Kubernetes resources: Make everything declarative and improve portability"
date = "2024-05-29"
description = "This post discusses an approach to splitting responsibilities between your infrastructure code and Kubernetes resources, in a way that maximises the portability of your setup between different environments and providers and ensures declarative management of everything."
tags = [
    "kubernetes",
    "infrastructure",
    "devops",
]
+++

When deploying infrastructure that includes Kubernetes clusters, there is always a point where we have to decide exactly what resource should be handled via infrastructure code tools like Terraform, and what resources will be managed by Kubernetes itself. This short article looks at a few key resource types and what I currently see is a good working model for splitting responsibilities between the cluster and infra code.

We'll look at this in terms of 'standard' webapp components, such as:

- Network infrastructure (Load balancers / public entrypoints)
- DNS management
- TLS Certificates
- Secrets
- Application deployments (e.g. your webapp containers)
- Databases

## Why not do as much as possible with Terraform?

Terraform is a declarative way to make API calls to an (infrastructure) provider, and that means all `.tf` code we write is specific to the provider and/or infrastructure environment we're working in. Almost none of this code is portable between providers, with the exception of resources inside Kubernetes. With Terraform you are able to deploy resources inside a cluster, though the providers for this use case are not great. Other downsides include the lack of continuous reconciliation provided by GitOps-like tooling, and the fact that now anyone desiring to deploy things to a cluster needs to know both Terraform *and* Kubernetes to get anything done.

For deploying stuff *inside* a Kubernetes cluster, you are much better off with a declarative GitOps approach, using the likes of FluxCD and ArgoCD.

## Why not do as much as possible with Kubernetes itself?

Simple answer is, you cannot do *everything* this way no matter what - creation of your clusters and your base infra like subnets have to remain as infra code anyway. Please don't use CLI tools or anything imperative to deploy clusters. 🙏

With tools like [Crossplane](https://www.crossplane.io/), there are more and more options to manage these infrastructure components of a provider using the Kubernetes API. However, at its simplest, Crossplane just provides a bridge between Kubernetes CRDs/API and an infrastructure provider's APIs; at the end of the day someone still has to manage those CRDs and their mapping towards the 'right' resources and configurations in your provider of choice.

Finally, there are resources like cloud provider load balancers that can be automatically provisioned for managed Kubernetes clusters, for example when you create a `Service` of type `Load Balancer`. The downsides of this are discussed in more detail below but focus on LB lifecycles being different from cluster lifecycles, and the fact that usually larger envs have more requirements for LBs like authentication, firewalls, etc., and managing more and more provider-specific configs in your cluster makes your setup less portable.

## A natural split

Since big parts of your Terraform code are cloud provider specific, it makes sense to keep everything cloud-provider specific in Terraform as much as possible. Since Kubernetes clusters are standard, the resources inside the cluster - your apps, your DBs, etc. - are **not** provider specific at all, so it make sense to keep these separate of the parts that are provider-specific. Therefore;

> *If something is necessarily infrastructure provider specific, create and manage it in Terraform. Otherwise, create and manage it inside Kubernetes.*

## Proposed model

Following this idea, the purpose of the model proposed below is to provide a clear split of responsibilities and remain as infrastructure provider-agnostic as possible.

{{<mermaid>}}
flowchart TD

subgraph infra["Infrastructure-as-code: Things outside the cluster"]
    lb["Load balancers / public entrypoints\n(AWS ALB, Metal LB, CloudFlare tunnel, etc.)"]
    waf["Firewalls"]
    dns["DNS records\n(With wildcards)"]
    certs["TLS certificates\n(With wildcards)"]
    oidc["OIDC SSO"]

    oidc --> lb
    dns --> lb
    waf --> lb
    certs --> lb
end

subgraph cluster["Kubernetes cluster\n(Managed with GitOps)"]
    db["Databases: CloudNativePG"]
    secrets["Secrets: Sealed secrets"]
    deployments["Apps containers"]
    ingress["Ingress controller"]
    ingress -->|Routes traffic| deployments

    deployments <--> secrets
    deployments <--> db
end

git["Git Provider"] ---->|GitOps repo| cluster
lb -->|Hyperscaler/infrastructure environment\nspecific way to route public traffic to ingress controller  | ingress
{{</mermaid>}}

### Keep network infrastructure managed as code

By default, a Kubernetes cluster with an ingress controller will as your cloud provider for a load balancer. This works fine, but ties the lifecycle of the load balancer to that of the cluster - usually not something you'll want from an availability and future-proofing perspective. Secondly, your infrastructure components that attach to the LB, for example WAFs, now need to rely on an LB that is *not* managed in that same infra codebase, so you'll have to hardcode IDs or invent other creative/complex solutions to keep resources manage by different things in sync and connected to each other.

When your network infrastructure is independent of the cluster, you can replicate the internal state of the cluster pretty much *exactly* in a different cloud provider or even a self-hosted environment. For example, you can replace the load balancer with a CloudFlare tunnel, and your configuration is now fully self-hostable.

### Keep DNS and certificates outside the cluster

I've used tools like [external-dns](https://github.com/kubernetes-sigs/external-dns) and [cert-manager](https://github.com/cert-manager/cert-manager) extensively, to manage an environment where DNS records for a service, as well as TLS certs related to those records, are automatically provisioned. This approach is rock solid, and has been running smoothly for years.

However, once you go towards managing load balancers and cluster entrypoints *outside* the cluster, it is natural to then also manage DNS records and certificates with infrastructure code - outside the cluster. For large 'platform'-type clusters where you might need a lot of different subdomains, you can use wildcard DNS records pointing to your LB combined with wildcard certificates attached to the LB. You offload the TLS encryption/decryption workload to your infra provider, your setup is simpler, there are less DNS records in total, and finally less "leakage" of information about what subdomains/environments/apps you might have deployed, since your DNS entries just have the wildcard records. (Not that this is your only security measure, hopefully.)

### Keep databases in the cluster

This might be hard to push for in some organisations, but the ability to declaratively create databases for new environments and manage their backups in Kubernetes is great since the deployment of your app can sit side by side with the deployment of its database. This gives more self-sufficiency to developers also, as the infrastructure team may not be needed to do "AWS magic" to make a new database appear. And again, since we do things inside Kubernetes itself, you are not locked in to the cloud provider's managed services (and don't need to pay the significant premium managed DB services usually charge), and you'll get the benefits of continuous reconciliation and automation with GitOps tools.

After a short and frustrating stint with Zalando's Postgres operator, I found much better success and an overall better experience with [CloudNativePG](https://cloudnative-pg.io/) and can highly recommend it.

### Keep secrets in the cluster

The advantage of using [sealed-secrets](https://github.com/bitnami-labs/sealed-secrets) is that we can manage our secrets *declaratively* and apply them to the cluster using our standard approach with GitOps - the entire state of the cluster is in Git, including secrets in an encrypted format. Similar to databases, going this route also means that you are independent of the infrastructure provider also in this respect.

Still, [external-secrets](https://github.com/external-secrets/external-secrets) is a fantastic tool and we've been using it in production for a long time. If you're on AWS, as long as you stick to AWS Systems Manager Parameter Store, it is also completely free (secrets Manager will charge you per secret), so this area is not as clear cut, especially since switching from one tool to another is not terribly difficult.

## Towards full replicability across different environments and providers

The proposed model discussed here makes your cluster portable and replicable across any provider. If you find a better deal for your infra, you won't be locked in; if your clients want to use a different hyperscaler than you; no problem.

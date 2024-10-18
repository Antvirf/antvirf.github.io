+++
author = "Antti Viitala"
title = "Managing k8s resources on AWS EKS with Terraform"
date = "2023-03-13"
description = "A short overview / compilation of useful snippets to allow you to create and manage Kubernetes resources and Helm charts with Terraform"
tags = [
    "aws",
    "kubernetes",
    "infrastructure",
    "devops"
]
images = ['images/apple-touch-icon-152x152.png','images/splash.png']
+++

## Rationale

Terraform allows you to create all the cloud resources you could want with just a few commands, however it usually is paired with other tools like [Ansible](https://www.ansible.com/) to then apply configurations on those resources and bring up applications.

Kubernetes and Helm have made the process of bringing applications with many moving parts quite easy - just do a `helm install` and you're done!

Combining Terraform's capabilities to create and configure infrastructure with Kubernetes' and Helm's capabilities to set up and run applications would allow us to spin up complete environments from scratch in a repeatable fashion.

This post is about setting that up.

## Why this isn't as easy as it (theoretically) could be

Terraform is a great tool for provisioning infrastructure, but at its core the Terraform providers we use tend to map more or less 1-to-1 with the features of a particular cloud provider's API for managing resources. For example, AWS provides APIs for creating [Kubernetes clusters](https://docs.aws.amazon.com/eks/latest/APIReference/API_CreateCluster.html) - and similarly the AWS Terraform provider is able to create EKS clusters with the [`aws_eks_cluster`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_cluster) resource. At the back, the Terraform provider uses these APIs.

Technically, Terraform isn't that limited - just take [this Terraform REST API provider](https://github.com/Mastercard/terraform-provider-restapi) as a generic example. You could always write your own providers to do whatever you want. However the main providers that Hashicorp maintains for each CSP ([AWS](https://registry.terraform.io/providers/hashicorp/aws/), [Azure](https://registry.terraform.io/providers/hashicorp/azurerm), [GCP](https://registry.terraform.io/providers/hashicorp/google)) have to be rock solid - combined these three have almost 2 __*billion*__ downloads to date - and hence they focus on feature parity with the cloud provider.

AWS doesn't provide a REST API to create a Kubernetes manifest, or install a helm chart on a cluster, and so the AWS provider does not support these operations - but a third party provider focused on this functionality can, and does, give us this ability.

## Terraform snippets for setting up `kubectl` and `helm` providers

### Base providers

```hcl
# providers.tf - declaring required providers

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.49.0"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = ">= 1.14.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.6.0"
    }
  }
  required_version = "~> 1.0"
}

# AWS provider settings
provider "aws" {
  region  = var.region  # region to create your resources in
  profile = var.profile # aws config/credentials profile to use
}

```

### Provider configs

These blocks configure the providers in detail. Effectively, we give both providers the relevant EKS cluster's endpoint, the certificate, and an exec block that fetches the AWS EKS credential token.

```hcl
# providers.tf - configuring kubectl
provider "kubectl" {
  host                   = data.aws_eks_cluster.default.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.default.certificate_authority[0].data)
  load_config_file       = false

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    args        = local.eks_auth_exec_args
    command     = "aws"
  }
}

# providers.tf - configuring helm
provider "helm" {
  kubernetes {
    host                   = data.aws_eks_cluster.default.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.default.certificate_authority[0].data)
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      args        = local.eks_auth_exec_args
      command     = "aws"
    }
  }
}
# providers.tf - helm's kubernetes provider
provider "kubernetes" {
  host                   = data.aws_eks_cluster.default.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.default.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.default.token

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    args        = local.eks_auth_exec_args
    command     = "aws"
  }
}

```

The arguments for the exec block is the same for both providers, and hence separated to a local variable below. As we can see, there is no magic here - you could run this yourself with the AWS CLI if you wanted to.

```hcl
# providers.tf - cluster cert fetch command
locals {
  eks_auth_exec_args = [
    "eks",
    "get-token",
    "--cluster-name",
    data.aws_eks_cluster.default.id,
    "--region",
    var.region,
    "--profile",
    var.profile
    ]
}
```

### Supporting data resources

```hcl
data "aws_eks_cluster" "default" {
  name = module.eks.cluster_id
}
data "aws_eks_cluster_auth" "default" {
  name = module.eks.cluster_id
}
```

## Example usage - installing Postgres with Helm

```hcl
resource "helm_release" "postgres_helm_release" {
  name = "postgres"
  repository       = "https://charts.bitnami.com/bitnami"
  chart            = "postgresql"
  namespace        = "postgres"
  create_namespace = true
}
```

## Example usage - applying a plain Kubernetes manifest

Example `ConfigMap` [from Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/configmap/#configmaps-and-pods). You can mix plain YAML and substitute variables from Terraform as needed.

```hcl
resource "kubectl_manifest" "pinot_jobspec_pvc" {
  yaml_body = <<-EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: game-demo
  namespace: ${var.namespace_from_terraform}
data:
  player_initial_lives: "3"
  ui_properties_file_name: "user-interface.properties"
EOF
}
```

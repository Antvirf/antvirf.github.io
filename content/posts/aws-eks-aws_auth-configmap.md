+++
author = "Antti Viitala"
title = "Managing aws_auth ConfigMap with the AWS EKS Terraform module"
date = "2023-03-16"
description = "Useful snippets and tips to manage the infamous aws_auth ConfigMap via Terraform without pulling your hair out."
tags = [
    "aws",
    "kubernetes",
    "infrastructure",
    "devops"
]
images = ['images/apple-touch-icon-152x152.png','images/splash.png']
+++

## What is the `aws_auth` and why does it exist?

Unlike AKS, by default AWS EKS uses [AWS authentication tokens for managing access to the cluster](https://docs.aws.amazon.com/eks/latest/userguide/cluster-auth.html). This is great - it improves security significantly - but it comes with some side-effects that trip up a lot of people at first, myself included.  This is because the cluster maintains its own access control list - in this `aws_auth ConfigMap` - that describes who can do what within the cluster, and by default __only the specific AWS user who created the cluster is included in that list__.

Regardless of your privileges in AWS, if an EKS cluster was created by someone else - and they didn't update the `aws_auth` `ConfigMap` - you are locked out, period. In the diagram below, steps 1 through 3 will succeed, but 4 will fail as the ARN of your user/role is not present in the access control list. It is therefore critical that during cluster creation this resource is updated correctly.

![aws-eks authentication flow](/content/aws-eks-iam.png)

## Creating and managing `aws_auth` with the AWS EKS Terraform module

The [AWS EKS Terraform module](https://registry.terraform.io/modules/terraform-aws-modules/eks/aws/) ([what are Terraform modules?](https://developer.hashicorp.com/terraform/tutorials/modules/module#what-is-a-terraform-module)) offers a way for us to manage this resource, giving us the following parameters to set within the main module:

```hcl
manage_aws_auth_configmap = true

aws_auth_roles = [
  {
    rolearn  = "arn:aws:iam::66666666666:role/role1"
    username = "role1"
    groups   = ["system:masters"]
  },
]

aws_auth_users    = [...] # if you want to add specific users instead
aws_auth_accounts = [...] # if you want to add specific accounts instead
```

While most Terraform modules just the cloud provider's provider - for example the AWS App Runner module uses just the AWS provider - the AWS EKS module is an exception. AWS provides no REST APIs to update this resource, so the [`kubectl` provider](https://github.com/gavinbunney/terraform-provider-kubectl) must be used instead. This provider isn't in any way linked to the AWS provider, so it needs to be separately configured in order to have access to the EKS cluster created by the module.

## Configuring the required providers

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
  }
  required_version = "~> 1.0"
}

# providers.tf - configuring aws provider
provider "aws" {
  region  = var.region  # region to create your resources in
  profile = var.profile # aws config/credentials profile to use
}

# providers.tf - configuring kubectl
provider "kubectl" {
  host                   = data.aws_eks_cluster.default.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.default.certificate_authority[0].data)
  load_config_file       = false

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    args        = [
      "eks",
      "get-token",
      "--cluster-name",
      data.aws_eks_cluster.default.id,
      "--region",
      var.region,
      "--profile",
      var.profile
    ]
    command     = "aws"
  }
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

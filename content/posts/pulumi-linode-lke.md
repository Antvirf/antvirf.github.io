+++
author = "Antti Viitala"
title = "Linode Kubernetes cluster with Pulumi & Python"
date = "2023-08-15"
description = "Setting up a Kubernetes cluster on Linode Kubernetes Engine (LKE) and deploying to it with Pulumi has a few caveats. It took me an embarrassing amount of time to figure out, hopefully this article saves someone that time."
tags = [
    "kubernetes",
    "infrastructure",
    "devops",
    "pulumi",
]
images = ['content/pulumi-linode.png']
+++

Despite reading through the docs on [Pulumi Outputs](https://www.pulumi.com/docs/concepts/inputs-outputs/) a few times, getting a grasp of how to deal with outputs can be difficult. An added challenge to this was that the LKE provider differs from other Pulumi providers (such as EKS) in that the kubeconfig it outputs is provided in an encoded format, requiring some decoding before it can be used later in the program - here's a snippet how to deal with that.

## Basics

- [Set up a new Pulumi project](https://www.pulumi.com/learn/pulumi-fundamentals/create-a-pulumi-project/)
- [Install Linode provider for Python](https://github.com/pulumi/pulumi-linode#python)
- [Set up Linode token](https://www.pulumi.com/registry/packages/linode/installation-configuration/)

## Create an LKE cluster and add an ingress controller with Helm

```python
import base64
import pulumi
import pulumi_linode as linode
import pulumi_kubernetes as kubernetes

cluster = linode.LkeCluster(
    "my-cluster",
    k8s_version="1.26",
    label="name_of_my_kubernetes_cluster",
    pools=[
        linode.LkeClusterPoolArgs(
            count=1,
            type="g6-standard-1",
        )
    ],
    region="ap-south",
)
pulumi.export("kubeconfig", cluster.kubeconfig)

# Decoding the config without forcing it to a string
k8s_provider = kubernetes.Provider(
    "k8s-provider",
    kubeconfig=cluster.kubeconfig.apply(
        lambda config: f"{base64.b64decode(config).decode('utf-8')}"
    ),
)

# Creating a namespace
ingress_ns = kubernetes.core.v1.Namespace(
    "ingressns",
    metadata=kubernetes.meta.v1.ObjectMetaArgs(
        labels={"app": "ingress-nginx"},
        name="ingress-controller",
    ),
    opts=pulumi.ResourceOptions(providers={"kubernetes": k8s_provider}),
)

# Creating an ingress controller
ingresscontroller = kubernetes.helm.v3.Release(
    "ingresscontroller",
    opts=pulumi.ResourceOptions(providers={"kubernetes": k8s_provider}),
    chart="ingress-nginx",
    namespace=ingress_ns.metadata.name,
    repository_opts=kubernetes.helm.v3.RepositoryOptsArgs(
        repo="https://kubernetes.github.io/ingress-nginx",
    ),
)
```

## References

- [List of Linode instance types](https://api.linode.com/v4/linode/types)
- [Pulumi docs for Linode LKE](https://www.pulumi.com/registry/packages/linode/api-docs/lkecluster/)

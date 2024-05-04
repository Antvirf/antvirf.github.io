+++
author = "Antti Viitala"
title = "Deploying to k8s environments which you know little about: Some pointers and experiences"
date = "2024-05-04"
description = "Most of the time, we install Kubernetes applications in environments and clusters that we know quite well. Sometimes though, you might just need to deploy a system of applications to a cluster that is largely a black box, running in an environment with various restrictions that are also unknown to you in advance. This article discusses a few pointers on approaching this problem."
tags = [
    "kubernetes",
    "infrastructure",
]
+++

## Scenario

- "You'll get a Kubernetes environment, please install your app on it. Also our cluster is quite restricted."
- "What kind of restrictions and policies do you have in place? How do you want to route traffic? What about..."
- "🤷‍♂️"

If this seems like a situation you might find yourself in, perhaps the notes below will be helpful.

## Assume that the cluster user / credential you are given cannot do anything by default

Start from the baseline that your user cannot modify anything (i.e. its perms are read-only for example), and that all permissions must be requested for explicitly. Even before you receive any access, list out all the resources you want to manage and the [required permissions](https://kubernetes.io/docs/reference/access-authn-authz/rbac/), and try to get these configured and ensured in advance. Many environments can have somewhat arbitrary limits on what is and is not allowed - either at resource type or permission verb level - so better safe than sorry. Remember that whatever you want to `create` or `update`, chances are you'll want to be able to `delete` them as well. Requiring a third party or a cluster admin to help you reset between iterations will slow you down a lot.

You can check the current user's permissions with:

```bash
kubectl auth can-i --list
```

Beyond the resources you want to edit/manage yourself, it's also worthwhile to requested read-only perms to `ValidatingWebhookConfiguration`
and `MutatingWebhookConfiguration` resources. The next section discusses this in detail.

## Separate your installation by resource type to figure out potential Admission Control limits

Now, just because Kubernetes will allow you to submit a request from an RBAC-perspective, a request for a resource can still be rejected by [Kubernetes' Dynamic Admission Control](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/).

This can make your life difficult in restricted environments, especially if you don't know the full extent of configured policies and validators. Many environments enforce technically "optional" but good practice-type configurations, like including a `resources` section with every pod that specifies the hardware resources that pod requires. There may be restrictions on the source registry, container vulnerabilities, ingress domains or subdomains etc., the list goes on. Admission controllers can enforce ~arbitrary restrictions.

When starting in a new env, trying to get the list of active admission control resources is worthwhile: `kubectl get validatingwebhookconfiguration`, as well as `kubectl get mutatingwebhookconfiguration` - but again your role may or may not have the permissions to even view these objects.

In case you cannot get the details of the active admission controllers, your best bet is to separate your installation by resource type, if only temporarily. So instead of e.g. `system-a.yaml` and `system-b.yaml`, where both would contain several resources of different kinds, you'd instead have your manifest split out by resource type:

- `deployments.yaml`
- `statefulsets.yaml`
- `pvcs.yaml`
- `services.yaml`
- `serviceaccount.yaml`
- `ingresses.yaml` / `routes.yaml` if on OpenShift
- `roles.yaml`
- `rolebindings.yaml`
- `crds.yaml`

With your installation configured like this, you can go type by type and apply each file one by one. At this point the goal is not to get to a working system, but rather test that all of your desired resources can be admitted to the cluster successfully.

The other added benefit to this approach is that in many environments, modifying certain resources is restricted to administrators only - so in any case your installation will likely need to carve out these restricted resources so that you can hand them off to the infrastructure/cluster administrators to take care of. Most commonly everything that is cluster-level / not scoped to a particular namespace is restricted.

## Favor flexibility over installation convenience

In most cases, this simply means opting for Kustomize or even plain manifests + shell scripts rather than Helm charts. Helm-based installations must be engineered for flexibility in particular areas, and you might not be able to predict in advance *where* you need to be flexible. The ability to quickly change any manifest files directly and as precisely as you need is immensely valuable.

In the background, you might still want to use Helm as a templating engine to produce your manifests. Later on as a project progresses and your understanding of the environment grows, transferring to Helm might become a viable option as you can then build and tailor the chart to the specific restrictions and roadblocks you encounter after the initial setup.

## Experiences with particular resources

### `Deployment`, `StatefulSet`, `DaemonSet` and `Pod` - everything involving containers

Just like with every other resource, admission controllers may reject manifests that do not conform to the restrictions configured in the cluster. Resources that involve containers are unique in that the container images themselves may also undergo validations - just because e.g. a deployment resource is created successfully with `apply`, this same command may fail at a later stage if the image is rejected by the environment.

In a restricted environment, your manifest files - being plaintext - may be transferred and be available in the environment *before* your containers. That allows you to already start testing a lot of resources, and even start testing whether your `Deployments` will `apply` correctly. Just keep in mind that a deployment with an invalid/unavailable image might be successfully created at first (just left in `ImagePullBackoff` status), but become inadmissible once the image becomes available to the cluster, should the image not fulfill the admission controller's requirements.

### `Job`

Helm charts of many popular services like [ingress-nginx](https://github.com/kubernetes/ingress-nginx/) create `Job` objects as part of their installation process, for example to create some default resources like self-signed TLS certs in the case of `ingress-nginx`. Like any other k8s resources, permissions to create `Job` objects are defined separately from other objects - so don't assume you'll have the permissions to create these resources.

Also remember that should you need to, one-off `Job` objects can easily be converted to `Pod` objects, so for such activities you can simply `apply` the `Pod` and run a "job" that way.

### `Route`

When creating `Route` objects for OpenShift, at least for the very first iteration I recommend leaving the out the `spec.host` key entirely. This lets the OpenShift router decide the subdomain and hostname to use for a particular `Route`, which eliminates one failure point entirely. Once you know for sure the domains/subdomains the cluster manages - and once you've tested that traffic is flowing correctly with the default setup - you can then specify `spec.host` with a value of your choice.

In my experience, it is also almost always worth specifying `tls.termination: edge`, such that the router will expose both an HTTP endpoint at port 80 (the default), as well as an HTTPS endpoint at 443 - without [HSTS](https://docs.openshift.com/container-platform/4.15/networking/routes/route-configuration.html#nw-enabling-hsts_route-configuration). This way, even if your application itself 'enforces' HTTPS by always redirecting to 443, your configuration won't get in the way.

### The 'risky' stuff: RBAC, `CRD`, and everything cluster level

Many out-of-the-box (box being [Bitnami](https://github.com/bitnami/charts)) Helm charts create a large number of `Role` and `RoleBinding` objects, and some also `ClusterRole` and `ClusterRoleBinding` objects. RBAC-related resources are usually restricted, so its worth going through your manifests to make sure you're only asking for what you need - it is much easier to explain and justify having 5 roles than having 50.

Also bear in mind that ingress controllers for example usually want to declare a specific `IngressClass` resource, and these being cluster-level, are usually restricted. Generally most clusters *will* have some type of ingressing/routing solution in place, so quadruple-check if it seems like there isn't one.

Just like everything else in this category, same for `CustomResourceDefinitions` - use only what you need, separate these types of resources from other resources, and expect that you won't be given access to manage them.

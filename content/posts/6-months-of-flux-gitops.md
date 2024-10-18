+++
author = "Antti Viitala"
title = "6 months of GitOps with Flux"
date = "2024-01-31"
description = "Few thoughts and experiences from running Flux as the primary GitOps and deployment tool in a small/medium-sized EKS environment."
tags = [
    "kubernetes",
    "infrastructure",
    "devops",
    "gitops",
    "flux",
]
images = ['content/flux.png']
+++

**TL;DR GitOps is king, and once you've tried it you can't go back 👑**

## Use case

As a software consultancy, at Synpulse8 we work on a large variety of projects across different teams, programming languages and regions. However, we do want to standardise what we can - primarily the development "platform" on which every team deploys to. Kubernetes has been the obvious choice for us as a way to be able to offer clients an easy answer to the usual question of "how do we deploy your stuff", and we've ran it on a few different public cloud providers, primarily AWS.

The question has then been how do we want to design, standardise and manage the processes of getting code onto Kubernetes.

## Migrating to GitOps

Before moving to a pull-based GitOps approach with Flux, we had been running push-based pipelines orchestrated on common tools like Jenkins and GitHub Actions. Our pipelines were responsible for everything, things like:

1. Checking code quality
1. Building the application and its container image
1. Pushing the image onto a registry
1. Patching a `kustomization` manifest with the new container image tag
1. Applying the updated manifests to our Kubernetes clusters

With Flux, the pipelines only handle steps 1-3, with Flux taking over deployments. Given that we were already doing all of this in our existing pipelines, migrating to Flux was relatively easy - just removing the final application bits from pipelines.

On the other side, we used Flux to [push image updates back to Git](https://fluxcd.io/flux/guides/image-update/) whenever new container tags were pushed, allowing us to keep our manifests on GitHub always up to date and aligned with the state of our clusters. More on this below.

## Clearer responsibilities around pipelines

Since the build pipelines, as well as their green checkmark / red cross representing workflow statuses that are attached to each repo on GitHub, were now restricted to just the build part of the overall workflow, "failing workflows" now always meant that something went wrong with the *build*. Failing builds are usually something that the developers themselves could fix. Before this change, someone from the platform team might have been pinged to take a look at a failing workflow by reflex, but now increasingly developers began to solve build-related problems themselves. This reduced the number of "cooks in the kitchen" touching these flows, and made the split of responsibilities between teams much cleaner.

As a result of moving to Flux and setting up or deployments this way, we saw less requests to address failing pipelines as developers naturally took over the responsibility for the build stage.

## Easy 'rollbacks' and ability to self-heal

When figuring out configurations for new things, it usually takes me a few attempts of trial and error to find the right approach. This may take a lot of iterations so I prefer not to litter the main commit history with various `try this` and `that` commits. Since all the applications hosted on our main clusters are handled with Flux, doing something like this becomes nice and easy. When starting to tinker with a particular app's manifests, I'd first run:

```bash
flux suspend kustomization my-app -n my-namespace
```

which makes Flux not reconcile the app temporarily. I'd then proceed to make my changes or carry out any experiments. If the outcome was a success, I'd just commit the new changes to the Git repo; if the outcome was a failure and I needed to roll back, I wouldn't commit anything. In either case, once I'm done with that bit of work, I'd just do;

```bash
flux resume kustomization my-app -n my-namespace
```

at which point Flux would revert the cluster state back to the previous version defined in Git, or pull from Git the new version with my desired changes. For a very quick change, I could even make a change directly in the cluster and test it immediately - a few minutes later, Flux would revert it to a well-defined state. This is all easy, safe, and lowers the threshold for tinkering which is in my opinion very valuable.

## Declarative deployment pipelines with Flux Image Updates

We opted to deploy things in a fully declarative way, meaning that we do not deploy `latest` tags, but rather continuously update the manifests in our GitOps repos in an automated way. These updates are then detected by Flux, which then deploys the new image. This setup isn't a part of the default installation, but I mention it here since the benefits it provides are significant. If each deployment required a developer to push a manually-prepared commit to change an image tag, we'd have wasted thousands of hours by now; or alternatively every application would run `app:latest` containers, and the state of the cluster wouldn't be defined in a stable way.

While this deployment flow has a few more steps, it is certainly worth it. I've found the below representation useful for communicating this concept.

{{< mermaid >}}
flowchart TD

ops["Platform team"] --> |"(Maintain environment via git)"| deploymentsRepo
dev["Developer"] --> |1. Commit code to git| repo["Application repository"]
repo --> |"2. Build & push container image\n(CI flows)"|registry["Container registry"]

subgraph gh["GitHub"]
    repo
    deploymentsRepo
end

flux["Flux\n(Inside Kubernetes)"]
flux <-.->|"3. Flux observes new image tag"| registry

flux --> |"4. Update git\n(e.g. bump container tag)"| deploymentsRepo["Deployment repository"]
deploymentsRepo --> |"5. Get desired state from Git"| flux

flux --> |"6. Apply desired state"| cluster["App deployments\n(Inside Kubernetes)"]
{{< /mermaid >}}

## Moving clusters and managing infrastructure - seamless migrations

Another big benefit enforcing a clear separation of build pipelines and deployment pipelines has meant that infrastructure teams running the clusters can now manage our workloads with much more flexibility and control.

As an example, we recently needed to migrate most of our workloads from one cluster to another. With our previous model, this would have meant changes to **the repository of each application** individually (~50-100 repositories), to point their deployment flows to the new cluster. With Flux, where a cluster itself "pulls" the deployments it runs based on central GitOps repositories, this migration was seamless and *completely unnoticeable* to the application teams and developers. Since the manifests used for deploying applications were declarative and explicit, we also had complete guarantees that the application instances were identical between environments.

By creating a full clone of the existing workloads in the new cluster, and only then migrating DNS records, developer workflows remained completely unchanged from one environment to the next, and users downstream never noticed the change.

## Closing thoughts

After moving to a pull-based GitOps model where the state of the cluster is continuously reconciled, it is very difficult to imagine going back to the old way.

- Updating numerous applications' manifests at once is significantly easier, for example to move workloads to different nodes or update resource limits.
- Coming back to an old deployment months after the fact is easy, since everything is defined in one place and I can safely assume that the state of the application documented in Git matches reality.
- I sleep better knowing that even if our entire cluster was nuked, the desired state is always there in Git and can easily be restored.

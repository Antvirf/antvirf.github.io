+++
author = "Antti Viitala"
title = "Deploying containers without a container registry on OpenShift/Kubernetes"
date = "2024-04-22"
description = "Kubernetes generally assumes you'll have a container registry, but sometimes you just don't. This post describes how you can transfer and deploy images on Kubernetes and OpenShift in that scenario."
tags = [
    "kubernetes",
    "infrastructure",
    "devops",
]
+++

## Use case

Though you'd usually assume that a Kubernetes  or OpenShift environment would always have a container registry available, this may not always be the case in more restricted or highly secured environments. While I wouldn't recommend a long-lived environment setup without a registry, it is possible to work around this limitation by directly transferring containers to all of your nodes manually.

This post describes an approach to deploying container images in Kubernetes/OpenShift without the use of any registry. Please note that the transfer process would need to be executed **once for every container, and for every node that would/could need to run that container**. This gets very tedious very quickly, so I recommend scripting automations for this once you've got the initial flow down.

## What are we trying to do?

When you schedule a pod, the container runtime interface available on the relevant node is tasked with scheduling that pod. What the CRI does next is as follows:

1. Depending on the [`ImagePullPolicy`](https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy), the container runtime will first check if that container is already present on the server
2. If the image is not present (or as required `ImagePullPolicy`), resolve that container artifact from the relevant container registry
3. Pull the container image to local storage of the node
4. Launch the container

Without a container registry, step #3 is impossible, so we need to bring over the container to the container runtime of the relevant node offline. The first graph below describes the standard flow described here, with the steps usually requiring a container registry represented as dotted lines. The second graph describes the workaround flow implemented in this post.

### Standard flow

{{< mermaid >}}
flowchart TD

subgraph node["Kubernetes/OpenShift cluster - individual worker node"]
    kubelet["Kubelet"]

    subgraph containerruntimeinterface["Container Runtime Interface (e.g. CRIO)"]
        crio["CRIO"]
        crilocal["Local storage\n(sudo crictl images)"]
    end
end

kubelet --> |Please run  myimage:latest|crio
crio <-->|"Try to find and launch the right image"|crilocal
crio -.-> |"Try to resolve image from a container registry\nif does not exist locally"|registry
registry -.->|"Pull images to local storage"|crilocal

registry["Container registry"]
{{< /mermaid >}}

### Alternative flow: Load images directly to CRI

{{< mermaid >}}
flowchart TD

subgraph node["Kubernetes/OpenShift cluster - individual worker node"]
    kubelet["Kubelet"]

    subgraph containerruntimeinterface["Container Runtime Interface (e.g. CRIO)"]
        crio["CRIO"]
        crilocal["Local storage\n(sudo crictl images)"]
    end

    subgraph containerruntime["Container runtime (e.g. Podman)"]
        crlocal["Local storage\n(podman images)"]
    end
end

subgraph jump["Jump host"]
    you["You"] --->|"Transfer and load image to the node's container runtime"| crlocal
    crlocal --> |"Transfer the loaded image to CRIO's local storage"|crilocal
end

kubelet --> |Please run  myimage:latest|crio
crio <-->|"Try to find and launch the right image"|crilocal
{{< /mermaid >}}

## Instructions

*Note that my examples use `podman`, but you should be able to replace it at any stage with an equivalent container runtime like `docker` or `containerd`. This depends on the setup of your nodes.*

## Step 1: Get your container image to a jump-host machine using a `.tar` file

In an environment without a container registry, you'll likely need to transfer your images as `.tar` files at some stage This requires first **saving** the image as a `.tar` file, **transferring** it, and then **loading** it into the node's container runtime.

To **save** an image as a `.tar` file:

```bash
# on a machine where you have built/pulled the image
podman save -o ./myimage.tar myimage:latest
```

Transfer it to your node, perhaps with `scp`, where `MYUSER` is your username on the node, and `MYNODE` is the IP address or hostname of your node;

```bash
# on a machine where you have built/pulled/saved the image
scp ./myimage.tar $MYUSER@$MYNODE:/home/$MYUSER/myimage.tar
```

On your node, load to the local container runtime:

```bash
# on your node/server
podman load --input myimage.tar
```

If you now execute `podman images`, you should be able to find `myimage:latest` if the loading was successful.
However, this image is still not visible to the node's container runtime interface - you can check that this is the case by running `sudo crictl images`. Your image is on the node, but is not yet visible to the CRI and therefore would still not be schedulable.

## Step 2: Figure out the container runtime endpoint of your node

This is documented [here](https://kubernetes.io/docs/tasks/administer-cluster/migrating-from-dockershim/find-out-runtime-you-use/#which-endpoint) in the Kubernetes docs, but boils down to:

```bash
# on your node/server
cat /proc/"$(pgrep kubelet)"/cmdline
```

The endpoint is the value of the `--container-runtime-endpoint` argument. In my case, using OpenShift CRC, this was `/var/run/crio/crio.sock`.

## Step 3: Transfer the image to the Container Runtime Interface endpoint

Use the value you obtained for the container runtime endpoint in step 3:

```bash
# on your node/server
podman image scp myimage:latest /var/run/crio/crio.sock
```

Now, this image should be visible to the CRI as well:

```bash
sudo crictl images # check output for your image
```

## Testing your setup

To ensure you do not accidentally pull an image, set your test workload's `imagePullPolicy` to `Never`, like this:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-test-image
spec:
  containers:
    - name: my-test-image
      image: myimage:latest
      imagePullPolicy: Never # this is the important bit
```

If the referenced image is not available, the pod will end up in `ErrImageNeverPull`. After finishing the steps above, it should pick up the available image in a few seconds and move to a `Running` status afterwards.

## References

- [RedHat: How Podman can transfer container images without registry](https://www.redhat.com/sysadmin/podman-transfer-container-images-without-registry)
- [Kubernetes: Container Runtime Interface (CRI)](https://kubernetes.io/docs/concepts/architecture/cri/)

---
title: "Proxmox Virtual Environment cluster"
date: "2025-02-01"
weight: 10
---

## Benefits of virtualization

Virtualization allows you to separate the physical and logical management of your infrastructure. Hardware can fail or change over time, and applications evolve even more frequently. Separating the management into two clearly distinct pieces makes life a lot easier as you can focus on one problem at a time.

This allows you to think about your compute and physical hardware as a pool that can contract, expand and change without affecting applications running on it.

Being able to manage logical hosts virtually rather than having each of them tied to a single physical server makes it easy to move fast when working on applications. As a result, you can...
    - Experiment with different operating systems and versions with minimal friction
    - Create domain or application-specific virtual machines that are responsible for just one thing, making them easy to manage and maintain
    - Back up an existing machine *fully*, and know that it can be restored to its exact previous state with a few clicks
    - Clone an existing machine for experimentation and testing critical changes
    - Move logical hosts between  physical machines of the cluster in case hardware needs to be repaired or decommissioned

A virtualization environment provides you with a big part of what could be considered a private cloud, and in the fashion of a private cloud you remain responsible for managing the hardware as well as the logical infrastructure (=virtual machines) that run on top of it.

## Proxmox overview

[Proxmox Virtual Environment](https://www.proxmox.com/en/) ("PVE") is an open-source platform for virtualization. Given that most alternatives are paid, Proxmox is popular in the homelab and hobbyist communities. The business model fo the company behind Proxmox is similar that of Red Hat - sell technical support subscriptions rather than source code.

Installing PVE on a server is similar to installing any Linux distro. Once installed, the server hosts a web-based user interface for working with the platform. Adding more servers to a [Proxmox "cluster"](https://pve.proxmox.com/wiki/Cluster_Manager) follows the same install process for the new machines, after which the new machine is "joined" to the cluster by a few clicks and a sharing of credentials in the UI.

## Physical Topology

Since a cluster of hosts running a virtualization hypervisor needs every member of the cluster to be able to talk to each other, the physical network topology of these servers can be relatively simple - especially if all your hosts remain on one site and in one subnet (very likely at home, very unlikely in an enterprise setting), like in my case below:

{{< mermaid >}}
flowchart LR

upstream["Router / network uplink"] --> switch
subgraph switch["UniFi Lite 8 PoE"]
    sport1["Port 1"]
    sport2["Port 2"]
    sport3["Port 3"]
end
pve1["Virtualization host #1 (low-power, always on)"]
pve2["Virtualization host #2 (medium-power,  sometimes on)"]
pve3["Virtualization host #3 (high-power, usually off)"]

sport1 --> pve1
sport2 --> pve2
sport3 --> pve3

{{< /mermaid >}}

## Cluster quorum

Generally, a Proxmox cluster expects all of its physical hosts to be up at all times. More specifically, it aims to maintain [cluster quorum](https://pve.proxmox.com/wiki/Cluster_Manager#_quorum), which is extremely important when the hosts need to coordinate their activities for replication, failover and high-availability. If a cluster has no quorum, all operations to change the state of the cluster or any of the virtual machines running on it are blocked. Quorum can be roughly defined as a state of the cluster when a group of nodes is able to deduce that they are the majority. A cluster of 3 nodes can therefore lose 1 node without losing quorum, as the 2 nodes remaining can deduce that they are the majority.

A cluster of 3 nodes cannot lose 2, as the single remaining live node knows it is only 1 out of 3 - a minority. A cluster of 2 nodes cannot lose either node, since 1 out of 2 nodes is not a majority.

## Breaking cluster quorum for hacky "spot compute" use cases

> Do **not** do this outside of a homelab with non-critical data

However, sometimes you may not need or want high availability in this sense. In the example topology above, only one or two nodes are on most of the time, but the cluster still needs to be usable and therefore it needs to have quorum. This can be achieved by adjusting the number of quorum votes each node has - if the always-on node has 3 votes by itself, but the other two machines only have one, then out of a total of 5 expected votes, 3 can be from the low-power node, giving it majority. This way we can have "quorum" with only 1 out of 3 nodes alive. **If your nodes use shared resources like network storage, this is a really bad idea and you shouldn't do this.**

This value can be adjusted in `corosync.conf` - see the `quorum_votes` parameter.

```bash
vi /etc/corosync/corosync.conf

...
nodelist {
  node {
    name: pve-low-power-node
    nodeid: 1
    quorum_votes: 3
    ring0_addr: x.x.x.x
  }
}
...
```

*Note*: The real solution to this use case is to [host a separate quorum device](https://pve.proxmox.com/wiki/Cluster_Manager#_corosync_external_vote_support).

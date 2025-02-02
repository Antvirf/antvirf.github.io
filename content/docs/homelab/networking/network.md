---
title: "Network layout"
date: "2025-01-30"
weight: 1
---

## Physical network and equipment

The UniFi Dream Machine is a good base for this setup, primarily because it is highly configurable while also being trivially easy to set up for the standard "give-me-internet-now" use case. Having the ability to easily go back to a working setup at any point gives me a peace of mind to experiment and build more complex setups. The UDM is also very compact for its feature set, which is a major benefit unless you have the ability and space to build a rack-based setup.

### Network devices and their roles

- [UniFi Dream Machine ('UDM')](https://amzn.to/4jANJ24), which acts as:
  - 'Gateway' to the internet, meaning it connects directly to the modem provided by the internet service provider
  - Router at gigabit speeds
  - WiFi access point with support for multiple simultaneous networks
  - Runs UniFi Network application for network and device management (Virtual networks, firewalling/security rules, network traffic identification)
- [UniFi Lite 8 PoE switch](https://amzn.to/3WG9B22), which expands the physical network with more ports
  - 4 out of the 8 ports are Power-Over-Ethernet, which gives you extensibility in case you want to add PoE devices like UniFi access points

### Downsides of the UDM

The UDM is a good starting point to get into UniFi and networking in general, as it covers a broad set of features up to a decent level. However for anything beyond a small home;

- 1 Gigabit connection to the internet is relatively slow, with home fiber broad bands offering 5, 10, even 25 G in some cases already. More modern and dedicated gateways can provide much better performance.
- Security features offered on UDM like packet inspection can have a significant impact on your actual internet bandwidth and speed. A separate (and more powerful) network firewall for example would probably do a better job here, especially if you plan to host services accessible from the outside.
- 4 x 1 GbE ports - servers and desktops using 10 GbE are readily available, so having the UDM capped at 1 GbE limits your bandwidth on devices that could be capable of more. Four ports are also quickly exhausted
- No support for secondary WAN (=internet) connection

### Physical network layout diagram

{{< mermaid >}}
flowchart LR

modem["ISP Fiber modem"]
subgraph udm["UniFi Dream Machine"]
    subgraph wifi["WiFi radios"]
        client["Various home WiFi clients"]
    end
    port0["WAN Port"]
    port1["Port 2"]
    port2["Port 1"]
    port3["Port 3"]
    port4["Port 4"]

    port0 --> port1
    port0 --> port2
    port0 --> port3
    port0 --> port4
    port0 --> wifi
end
subgraph switch["UniFi Lite 8 PoE"]
    sport1["Port 1"]
    sport2["Port 2"]
    sport3["Port 3"]
    sport4["Port 4"]
    sport5["Port 5"]
    sport6["Port 6"]
    sport7["Port 7"]
    sport8["Port 8"]
end

pi["Raspberry Pi 4"]
subgraph pve1["Virtualization host"]
    subgraph k8s["Kubernetes cluster"]
        svcA["Example service A"]
        svcB["Example service B"]
    end
end
pve2["Virtualization host"]
pve3["Virtualization host"]

modem --> port0
port2 --> switch
port1 --> pi
sport1 --> pve1
sport2 --> pve2
sport3 --> pve3

{{< /mermaid >}}

- *All connections in this diagram are CAT6 RJ45 cables*
- *Virtualization hosts may contain any number of virtual machines, which all join the physical network via the host's network interfaces*

## Logical (virtual) Network

The purpose of the physical layout above is to ensure that all devices within the network are joined together at the physical level. Using the UniFi network application, we then partition the network into two large subnets - one for 'normal' users, one for our on-premises 'infrastructure' and services. Outside of UniFi, each Kubernetes cluster maintains its own virtual network internal to the cluster, which isolates all the services running inside the cluster from outside access unless explicitly exposed via an ingress point.

The actual networks and their sizing change time to time so the exact CIDR is not included here, but generally speaking the 'normal' network is much smaller since it is intended for end-user device.

{{< mermaid >}}
flowchart TD

subgraph main["'Normal' - 192.168.x.x"]
    eu1["Laptop (wifi)"]
    eu2["Phone (wifi)"]
    dns["DNS Server"]
end

subgraph infra["'Infra' network - 10.x.x.x"]
    subgraph ihost1["Virtual machine"]
        k8sA["Kubernetes cluster node A"]
    end

    subgraph ihost2["Virtual machine"]
        k8sB["Kubernetes cluster node B"]
    end
    subgraph ihostx["Virtual machine"]
        k8sC["Kubernetes cluster node C"]
    end

    subgraph k8sinternal["Kubernetes network"]
        workloadA["Container A"]
        workloadB["Container B"]
    end

    k8sA-->k8sinternal
    k8sB-->k8sinternal
    k8sC-->k8sinternal

end

user --> |WiFi connection|main
main <-->|Firewall-restrictions managed in UniFi|infra

{{< /mermaid >}}

Even though we have multiple virtualization hosts, they all spawn virtual machines to the same network, so from an end-user perspective there is no practical difference networking-wise whether they are interacting with a virtual machine or a bare-metal host.

### Considerations for larger deployments

- The example here is a single-site deployment - with multiple sites/offices/locations, creating separate VLANs for each site with their corresponding separate CIDR ranges is important for later expansion and location-specific management
- Further separation of 'infra' to different subnets may make sense for more fine-grained management
- Leaving room for the future - in a non-homelab scenario where 'rebuild everything' is not a viable solution, leave plenty of unallocated address space to expect the unexpected

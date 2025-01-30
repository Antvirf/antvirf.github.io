---
title: Homelabbing
bookToC: true
weight: 15
bookCollapseSection: true
---

# Homelabbing

The purpose of this project is to design, set up and maintain "proper" on-premises computing infrastructure, as if for a small technology or software development company.

## General principles

- Everything must be open source
- Encrypted and secure traffic between all involved machines (No insecure / self-signed certs)
- Authentication in front of all services
- Configuration and infrastructure should be written as declarative / idempotent code wherever possible
- While actual scale here is miniscule, chosen solutions should scale decently well with minimal effort
- Internal services are priority - hosting sites or services for heavy external use is not
- No separation of physical networks, everything is joined together with subnetting and network partitioning handled using VLANs
- Linux servers only

## High-level technology decisions

- Operating systems: [CentOS Stream](https://www.centos.org/centos-stream/), given how close it is to RHEL
- Deploying services: [Kubernetes](https://kubernetes.io/) since it is ubiquitous, specifically the [RKE2](https://docs.rke2.io/) distribution since it is straightforward to install and comes with relatively secure defaults
- Networking hardware: [UniFi](https://www.ui.com/introduction)

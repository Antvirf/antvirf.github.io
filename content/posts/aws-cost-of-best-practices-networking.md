+++
author = "Antti Viitala"
title = "AWS: Cost of Best-Practice networking"
date = "2023-12-28"
description = "Separating services across accounts and VPCs on AWS is considered best practice, and since VPCs and other basic network components are 'free', it is easy to forget that this type of architecture can come with significant costs. This post looks at a simple example and provides an estimate of fixed costs to architect an organization's workloads this way."
tags = [
    "aws",
    "infrastructure",
]
images = ['images/apple-touch-icon-152x152.png','images/splash.png']
+++

This post discusses some basic [AWS "well-architected" framework](https://aws.amazon.com/architecture/well-architected/) principles around networking, and how much idle costs you may be facing when configuring your infrastructure this way. The discussion assumes a microservices-based application architecture where separation is also applied at infrastructure level, i.e. each service resides in their own private network.

## Starting from scratch with well-architected best practices

### Sample scenario

Just to have a small sample scenario to model, let's assume the following for our organization:

- 10 separate services/workloads
- Each service has 3 environments (e.g. dev/test/prod)
- All "shared" components between services must be separated for each environment
- Whenever possible, services should span all 3 AZs for better availability

Hence in total we'll run ~30 service/environment instances, plus some shared components on top.

All prices are presented as "USD per year", and using prices provided by AWS for the `ap-southeast-1` (Singapore) region.

### Account segregation

One of the first questions a company who's new to AWS will face is *"how should we structure our accounts?"*, to which AWS' answer is something like this:

> [AWS Well-Architected Framework: COST02-BP03 Implement an account structure](https://docs.aws.amazon.com/wellarchitected/latest/framework/cost_govern_usage_account_structure.html): Implement a structure of accounts that maps to your organization. This assists in allocating and managing costs throughout your organization.
> <br>**Level of risk exposed if this best practice is not established: High**
> ![accounts](/content/aws-cost-account-structure.png)

The theme of segregating workloads is prevalent across the framework, and separation by accounts is one of the first ways to do that. This is fine enough, since OUs and accounts themselves have no cost.

### Connecting workloads and following security principles

Since we have multiple accounts for different workloads (and the different envs, e.g. dev/test/prod for each workload), they will all have their own VPCs and subnets.

In order for our services to talk to each other, these networks will need to be connected to each other, and following the [framework's infrastructure security](https://docs.aws.amazon.com/wellarchitected/latest/framework/sec-infrastructure.html) guidance, we'll probably want a centralized ingress/egress point for our network, which we'd implement with a hub-and-spoke model following [REL02-BP04](https://docs.aws.amazon.com/wellarchitected/latest/reliability-pillar/rel_planning_network_topology_prefer_hub_and_spoke.html):

> ![hub and spoke](/content/aws-cost-hub-spoke.png)

In order to implement this network topology in a way where each environment comprises of its own 'hub' and corresponding 'spokes', we will need to create the following resources:

- 3x AWS Transit Gateways, one for each env (free in itself)
- 30x total AWS Transit Gateway attachments per service/env (`USD 0.07` per hour)

This will cost us `USD 18,396`.

### VPNs for developer access

Chances are you'll also need at least the developers and IT administrators, if not regular users, to access certain services within your private networks.

Thanks to the hub-and-spoke architecture, we don't ned to attach our VPC endpoints to each service's VPCs individually, but can instead attach our VPC endpoints directly to the hub network of each of the three environments. Client VPN endpoint pricing is `USD 0.15` per hour, per subnet. Assuming 3 subnets - one in each AZ to follow best practices - this will amount to approximately `USD 11,826`.

### Bonus: Keeping data within the AWS Network

Most enterprises prefer to keep as much of their network traffic within their own or managed networks. Many workloads will use AWS services like CloudWatch (for logs and metrics), CloudTrail (for AWS usage logging), and ECR (for containers) but by default, traffic to these services would go via the public internet.

To keep this traffic within the AWS network, we'll need [AWS Private Link](https://aws.amazon.com/privatelink/), and we need to configure a PrivateLink VPC endpoint for each AWS service we want to connect to. Since we have 3 environments (and a total of 3 hub-spoke networks as a result), we'll need to set up our PrivateLink endpoints for each environment separately. The pricing for PrivateLink VPC endpoints is `USD 0.013` *per AZ, per hour*, so the actual usage will be:

3 PrivateLink endpoints (CloudWatch, CloudTrail and ECR), for each of the 3 AZs, for each of the 3 environments: `USD 3,075`

## Summary

### "Fixed" costs

We now have an AWS organizational structure that follows the well-architected framework, and a nicely put together hub-and-spoke network topology for each of our environments. We're also routing some key traffic via AWS instead of the public internet with PrivateLink, and providing direct access to the private networks to our staff with Client VPNs.

At this point, the total annual cost of our setup is `USD 33,297` - remember that this is just the baseline "fixed cost", without including any workload services to actually run our apps or the associated network data transfer costs. A rounding error for a large company, but a significant chunk of cash for a smaller enterprise.

### Breakdown per environment and service/environment

With the described configuration:

- Cost per service/env is approximately `USD 613`
- Cost per 'environment' (just the shared parts) is approximately `USD 4,967`

So as an example, for a larger organization with 25 services and 4 envs for each, this would correspond to `USD 81,288` per year in "fixed cost".

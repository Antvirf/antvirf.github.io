+++
author = "Antti Viitala"
title = "AWS PostgreSQL 15.2: RDS vs. Aurora Performance comparison"
date = "2023-06-17"
description = ""
tags = [
    "aws",
    "infrastructure",
    "databases",
]
images = ['images/apple-touch-icon-152x152.png','images/splash.png']
+++

This article is inspired by [the fantastic post by Avinash Vallarapu on a detailed comparison and performance benchmark between Aurora PostgreSQL and RDS PostgreSQL](https://www.migops.com/blog/is-aurora-postgresql-really-faster-and-cheaper-than-rds-postgresql-benchmarking/). The purpose of *this* article is two-fold:

1. To repeat some of the benchmarks with the most recent versions of Aurora and RDS PostgreSQL.
1. Use slightly slower-spec hardware to see if the results apply for smaller workloads as well.

The raw numbers in terms of test durations from this article **are not directly comparable to the original article** given the smaller hardware configurations and smaller test data size. However, what should be roughly comparable are the *relative performance differences* between the two database services. It is worth nothing that in this article we only look at RDS **without provisioned IOPS**.

Finally, it is important to note that this article uses [Aurora I/O Optimized](https://aws.amazon.com/about-aws/whats-new/2023/05/amazon-aurora-i-o-optimized/), though as far as I can tell this is effectively just different pricing model for Aurora. Standard Aurora pricing is based on instance size, storage use, **as well as IOPS**, whereas Aurora I/O Optimized is priced based purely on instance size (at a premium) and storage use - you do not pay extra for IOPS.

## Infrastructure configuration

Following the template from the original article:

| Parameter | EC2 instance | RDS Postgres 15.2 | Aurora Postgres 15.2|
| :--- | :---: | :---: | :---: |
| Region | `ap-southeast-1` | `ap-southeast-1` | `ap-southeast-1` |
| VPC | Shared VPC | Shared VPC | Shared VPC |
| # of instances | 1 | 1 (single-az) | 1 (single-az) |
| Instance type | `m6a.large` | `db.r6g.large` | `db.r6g.large` |
| DB Engine | N/A | PostgreSQL 15.2 | PostgreSQL 15.2<br>**I/O Optimized Aurora** |
| vCPU | 2 | 2 | 2 |
| Memory | 8 GB | 16 GB | 16 GB |
| Storage | 8 GB | 100 GB (gp2) | Managed/unlimited
| IOPS | N/A | Baseline: 300 <br>Burst: 3000<sup>(1st iteration only)</sup> | Up to instance limits
| Network | 12.5 Gbps | 10 Gbps | 10 Gbps |
| EBS Encryption | ❌ | ✅ | ✅

The infrastructure for these tests is defined in Terraform can be found in [this repository](https://github.com/Antvirf/aws-aurora-vs-rds-performance-benchmark). The EC2 instance used to run the tests needs PostgreSQL client installed, which is done with:

```bash
sudo amazon-linux-extras enable postgresql14
sudo yum install postgresql
sudo yum install postgresql-contrib
```

## Testing methodology

The testing methodology makes use of [`pgbench`](https://www.postgresql.org/docs/14/pgbench.html) similar to the original article, though with a newer version (PostgreSQL client version 14.8). To keep the costs of the test in check, the load to be applied on these instances will also be smaller - the original article used a scale factor of 10000, whereas this article will use a scale factor of 5000, and the total test time for each run will be 10 minutes instead of 60 minutes.

The tests will be ran with 4, 8 and 12 clients and jobs similar to the original article, with four iterations each. Contrary to the original article, I did not burn RDS IOPS burst credits (5.4 million I/O credits at 3000 IOPS, which is 30 minutes) *before* the tests. As a result, the RDS instance effectively has 3000 IOPS available during the first iteration, and after the burst credits are depleted the instance provides a baseline of 300 IOPS.

For reference, IOPS burst credits are recovered at a rate of 3 IOPS per GB of storage per second, so for a 100 GB `gp2` volume we replenish 300 IOPS per second. Replenishing to the cap of 5.4 million credits would therefore take 300 minutes, or 5 hours.

```bash
# data initialization, run once per instance before the tests
pgbench -i -s 5000 -h $HOST -U postgres -d postgres -p 5432

# run the test - the last two parameters set the number of jobs and clients
pgbench -T 600 -h $HOST -U postgres -d postgres -p 5432 -j 4 -c 4
```

While the initialisation time isn't itself a proper benchmark, Aurora was generally faster than RDS (even though at this point RDS still had all its IOPS credits available):

| Step | RDS time taken | Aurora time taken | "How much faster is Aurora?" |
| :--- | :---: | :---: | :---: |
| Initialisation | 1124 sec | 746 sec | ~33% |
| Vacuuming | 1084 | 456 sec | ~58% |
| Primary keys | 943 | 524 sec | ~44% |
| **Total** |  **3151 sec** | **1726 sec** | **~45%**

## Test results

> With 2 services to test, 3 different configurations each, for 4 iterations, 24 test runs were executed over the course of about 5 hours. The total cost of this test was USD 4.27. 💸

Reminding again here that for the first iteration the RDS instance still had its full IOPS credits, so effectively it had 3000 IOPS for at least the majority of the duration of iteration #1. Iterations #2, #3 and #4 are run at baseline RDS performance - 300 IOPS for this instance - with all IOPS burst credits depleted.

| # of Jobs/Clients | RDS | Aurora |
| :--- | :---: | :---: |
| `4/4` | Iter 1: `818` TPS, `4.89` latency<br>Iter 2: `77` TPS, `51.95` latency<br>Iter 3: `66` TPS, `60.77` latency<br>Iter 4: `75` TPS, `53.27` latency | Iter 1:`458` TPS, `8.73` latency<br>Iter 2: `440` TPS, `9.01` latency<br>Iter 3: `442` TPS, `9.05` latency<br>Iter 4: `446` TPS, `8.98` latency |
| `8/8` | Iter 1: `902` TPS, `9.87` latency<br>Iter 2: `84` TPS, `95.20` latency<br>Iter 3: `82` TPS, `96.83` latency | Iter 1:`723` TPS, `11.07` latency<br>Iter 2: `715` TPS, `11.19` latency<br>Iter 3: `722` TPS, `11.08` latency<br>Iter 4: `722` TPS, `11.08` latency |
| `12/12` | Iter 1: `809` TPS, `14.84` latency<br>Iter 2: `93` TPS, `129.47` latency<br>Iter 3: `91` TPS, `132.21` latency<br>Iter 4: `90` TPS, `133.68` latency | Iter 1:`930` TPS, `12.90` latency<br>Iter 2: `929` TPS, `12.92` latency<br>Iter 3: `916` TPS, `13.11` latency<br>Iter 4: `919` TPS, `13.05` latency |

![4 clients comparison](/content/aurora-vs-rds-4-clients.png)
![8 clients comparison](/content/aurora-vs-rds-8-clients.png)
![12 clients comparison](/content/aurora-vs-rds-12-clients.png)

The baseline IOPS of an RDS instance is based on its storage capacity, so our results here with about one third the TPS of the original article make sense as we use an instance with one third the capacity. When using RDS, the storage capacity directly affects IOPS and therefore TPS, so knowing the rough size of the DB in advance is important for planning.

## Cost comparison

At the time of writing, the baseline cost of the services as configured in this article, placed in `ap-southeast-1`, is estimated in the table below. These are all on-demand prices; reserved prices are generally ~30% cheaper when looking at 1-year reservations paid fully upfront.

| Service | Cost / month, USD | Cost / year, USD |
| --- | --- | --- |
| RDS (~300) | 210.90 | 2,530.8 |
| Aurora (Standard @ 1 IOPS) | 240.07 | 2,880.84 |
| Aurora (Standard @ 100 IOPS) | 297.31 | 3,567.72 |
| Aurora (Standard @ 200 IOPS) | 355.12 | 4,261.44 |
| Aurora (Standard @ 300 IOPS) | 412.94 | 4,955.28 |
| Aurora (I/O Optimized) | 321.91 | 3,892.92 |

For rough cost estimations at different IOPS, please expand the section below.

<details>
<summary>Total cost vs. IOPS charts</summary>

This section contains charts of constant IOPS usage vs. estimated costs for databases of 100, 300 and 1000 GB in size for RDS and Aurora configurations. The primary purpose is to **compare** the difference between services, not to approximate actual amounts.

## Assumptions and caveats

* IOPS ranges only up to the max baseline IOPS of RDS at that particular capacity
* IOPS are 100% consistent, there are no spikes and the rate is always the same
* Networking or backup costs are not taken into account
* These figures probably **do not make sense** for the higher IOPS range, as likely at this point you will need a larger instance
* For anything critical, you will likely at least double the cost by having to run a multi-az configuration for availability

![100gb](/content/aurora-vs-rds-cost-100.png)
![100gb](/content/aurora-vs-rds-cost-300.png)
![100gb](/content/aurora-vs-rds-cost-1000.png)

</details>

## Conclusion

For databases that experience **low load most of the time, and only with short spikes**, RDS seems to be the better choice at this scale.  Costs-wise, even at 1 IOPS, Aurora is more expensive, and were you to run both instances at e.g. 300 IOPS, Aurora will cost almost twice as much per year. If your load spikes are relatively short and infrequent (at most, 30 min spikes, every 5 hours for a 100 GB database), the bursting offered by RDS will be sufficient to handle the load.

Should the load grow, as long as it remains relatively consistent, RDS can be configured with provisioned IOPS that will increase performance to similar levels as Aurora - or you can increase throughput by increasing the storage capacity of the database. This is apparent from the test results of iteration 1 where RDS had its burst credits, as well as the findings of the original article where a well-configured RDS beats Aurora in terms of performance.

> RDS shines when your workload is mostly stable, and spikes are short in duration.

If on the other hand your **database load is highly variable and has sustained spikes**, Aurora offers a reliable level of performance that won't degrade even if traffic spikes and remains at an elevated level for a long time. The decision between Aurora I/O Optimised and Aurora Standard will come down to your expected load in terms of IOPS. The sales pitch from AWS is that I/O optimized Aurora ["offers up to 40% cost savings for ... applications where I/O charges exceed 25% of the total Aurora database spend"](https://aws.amazon.com/about-aws/whats-new/2023/05/amazon-aurora-i-o-optimized/).

> Aurora offers superior baseline performance, but at small scale it is only worth if it the burst performance offered by RDS is insufficient for your load spikes - unless durability and availability are your driving concerns. If you do go for Aurora, try to estimate expected costs of Standard vs. I/O Optimized for your use case as the differences can be substantial.

## Other considerations

### Availability

In a multi-az configuration, Aurora has a superior SLA of 99.99% versus 99.95% for RDS, which depending on the criticality of your workload may be very significant.

### Durability

By default, Aurora makes more copies of your data for better durability. RDS can be configured with additional replicas, or the backups can be distributed to different locations manually, but this requires additional configuration and effort. If extreme durability is a concern, Aurora is probably the better choice.

### Very high scale

AWS recommends you to [go for Aurora if you need >80,000 IOPS](https://docs.aws.amazon.com/prescriptive-guidance/latest/migration-postgresql-planning/matrix.html). At that point you (probably) know databases and AWS services better than me, so I'll leave it at that.

## Deciding between RDS and Aurora at relatively small scale

This isn't intended to be very serious, but I thought it would be fun to make a flowchart about the basic decisions that play into this. At relatively small scale, the decision between RDS and Aurora is mostly about the expected load and how spiky it is.

**TL;DR:**

- If, given your expected database size in GB, RDS can either (a) provide enough baseline IOPS for your use case; or (b) handle spikes with its bursts, then you should choose RDS
- If durability/availability (at low operational effort) is everything, just go for Aurora.
- If you go for Aurora, make sure to estimate the costs of Standard vs. I/O Optimized for your use case.

{{< mermaid >}}
flowchart TD
    rds[(AWS RDS)]
    aurora[(AWS Aurora)]

    start((I want a DB))
    availability{"Is durability/availability\neverything?"}
    size["Figure out the size of your DB\nand multiply by 3 to get RDS IOPS"]
    iops["Figure out the rough IOPS expected\nfor your workload"]

    rds_iops_is_enough{"Are the RDS baseline\nIOPS sufficient?"}
    rds_burst_is_enough{"Is the burst capacity provided\nby RDS sufficient?"}

    start-->availability
    availability-->|Yes|aurora
    availability-->|"Not really"|size

    size --> iops --> rds_iops_is_enough
    rds_iops_is_enough -->|No|rds_burst_is_enough -->|Yes|rds
    rds_iops_is_enough -->|Yes| rds
    rds_burst_is_enough -->|No|aurora

{{< /mermaid >}}

## References

- [Original article](https://www.migops.com/blog/is-aurora-postgresql-really-faster-and-cheaper-than-rds-postgresql-benchmarking/)
- [AWS RDS Instance types](https://aws.amazon.com/rds/instance-types/)
- [AWS EC2 Instance types](https://aws.amazon.com/ec2/instance-types/)
- [Understanding AWS RDS burst vs. baseline performance](https://aws.amazon.com/blogs/database/understanding-burst-vs-baseline-performance-with-amazon-rds-and-gp2/)
- [AWS RDS SLA](https://aws.amazon.com/rds/sla/)
- [AWS Aurora SLA](https://aws.amazon.com/rds/aurora/sla/)
- [AWS Aurora storage and reliability](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/Aurora.Overview.StorageReliability.html)
- [AWS announces Aurora I/O Optimized](https://aws.amazon.com/about-aws/whats-new/2023/05/amazon-aurora-i-o-optimized/)

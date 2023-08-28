+++
author = "Antti Viitala"
title = "Deploying a Django webapp: ECS with AWS Copilot vs. EC2 with Kamal/Terraform"
date = "2023-08-28"
description = "Comparing AWS ECS against EC2 for deploying a Django webapp, from the perspective of having to actually maintain either solution. ECS deployment was set up with AWS Copilot, their CLI infra tool, and EC2 was set up with Terraform and deployed to with Kamal."
tags = [
    "infrastructure",
    "devops",
    "kamal",
]
images = ['content/kamal-logo.png']
+++

## TL;DR

Coming from Kubernetes, my expectations were high for ECS + Fargate to be an "easy way" to deploy stuff. I found ECS to be surprisingly complex,  disappointingly slow to deploy to, and ultimately not worth the price/performance tradeoff for getting a more 'managed' service.

The alternative of spinning up an ARM-based `t4g.small` EC2 instance and deploying to it with [Kamal](https://kamal-deploy.org/) was approximately half the price, came with better performance, and was significantly faster to deploy to. (And without any cloud vendor lock-in!)

Given this experience, in the future I'd go for plain VMs and Kamal wherever possible - and where more managed alternatives or a higher number of VMs are needed, I would directly go to Kubernetes.

## Context

In a recent project I've been worked on a Django-based webapp and was faced with the decision on how to get it 'live' on cloud infrastructure. Having worked with Kubernetes a fair bit, that approach would undoubtedly have worked here, though it does introduce a lot of moving part and adds complexity to the overall application landscape.

I had heard good things about [ECS](https://aws.amazon.com/ecs/) as the "easy alternative to kubernetes" and wanted to give that a try. I had also had come across the very popular article from 37signals on [why they left the cloud](https://world.hey.com/dhh/why-we-re-leaving-the-cloud-654b47e0) and their internally-developed-now-open-source tool [kamal (previously called MRSK)](https://kamal-deploy.org/), which I also was interested in testing out.

Some basic requirements for this deployment include:

- Django-based webapp, where the deployment needs:
  - Main container to run the app
  - Worker containers running the same image but with a different startup command
  - Redis as cache and job queue
  - Postgres as database
- All infra must be on AWS
- All infra must either as-code or controlled by a tool like AWS copilot
- Ability to deploy from local machine to target infra with a single command
- Logs must be collected in CloudWatch

Metrics that matter:

- Deployment speed: The time taken to push a change from local environment to live infra
- Ability to easily stop and bring up the costly parts of created resources
- Overall running cost

## ECS and AWS Copilot

[AWS Copilot](https://aws.github.io/copilot-cli/) is an open source CLI that makes deploying containerised apps on AWS ECS and [AWS App Runner](https://aws.amazon.com/apprunner/) very straightforward, similar to what you may expect from modern infra/CD companies like Railway and Render.

After some initialisation work, `copilot deploy` is all you need to do to get your container up on ECS.

## EC2 (Terraform IAC), deployments with Kamal

While this alternative does mean creating Terraform scripts to create infrastructure, using community modules like [this one for EC2](https://registry.terraform.io/modules/terraform-aws-modules/ec2-instance/aws/latest) make things a lot easier.

Once the VM itself is up and running, Kamal works pretty much the same as Copilot - `kamal deploy` and your app is live after some time.

## Comparison

The below table summarises my experiences in brief though note that this definitely isn't a scientific comparison. More ⭐ stars ⭐ the better. 

| Area | ECS & Copilot | EC2 & Terraform + Kamal
| :---: | :---: | :---: |
**Infra**: understanding required | ⭐<br>Chances are you *will* need to go beyond Copilot's small scope of supported pre-configured services. The moment you need an add-on, you need to understand most of what Copilot is doing at the back, negating the benefits in real-world use. | ⭐<br>No escape, all infra is on you. Use community modules to manage less code yourself.
**Infra**: code you need to manage | ⭐⭐<br>Only 'add-ons' like RDS that Copilot doesn't support natively, and **they must be CloudFormation**. The app itself and supporting services (e.g. Redis) don't need any additional infra code. | ⭐<br>You manage everything in a language of your choice. Terraform community modules make this quite easy.
**Infra**: effort to integrate CloudWatch logging | ⭐⭐⭐<br>Out of the box | ⭐⭐<br>Add IAM role to the EC2 instance and ~3 lines of config.
**Deployment** speed:<br>(single ~600mb container) | ⭐<br>5-10 minutes, large variance | ⭐⭐⭐<br>~60-80 **seconds**
**User experience**: App performance<br>(feeling, initial page load) | ⭐<br>Slow | ⭐⭐⭐<br>Fast
**Cost**: Approx monthly resource cost | ⭐⭐<br>~36 USD per month (1 vCPU, 2gb RAM) |  ⭐⭐⭐⭐<br>~15 USD per month (t4g.small)

## Note on costs

The resource costs in the table are just for main compute, the actual amounts are likely to be significantly higher with ECS Fargate. (which should not come as a surprise given it's 'serverless' nature)

For example, a Celery worker container at small scale can easily sit in the same VM as the webserver, and so can Redis. With Kamal I can deploy all three to the same VM; with ECS Fargate I would pay for 2 more containers, tripling the overall cost.

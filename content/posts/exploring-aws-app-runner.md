+++
author = "Antti Viitala"
title = "Experiences with AWS App Runner"
date = "2023-06-05"
description = ""
tags = [
    "infrastructure",
    "devops"
]
images = ['images/apple-touch-icon-152x152.png','images/splash.png']
+++

## What is [AWS App Runner](https://aws.amazon.com/apprunner/)?

In brief, AWS App Runner is a (highly) managed service AWS offers to run your applications with as little infrastructure knowledge and maintenance as possible. You provide a [code source in one of the supported languages/frameworks](https://docs.aws.amazon.com/apprunner/latest/dg/service-source-code.html), or your [own container image via ECR](https://docs.aws.amazon.com/apprunner/latest/dg/service-source-image.html), and AWS takes care of the rest. [This presentation by AWS](https://d1.awsstatic.com/events/Summits/reinvent2022/CON312_Auto-scale-your-web-application-using-AWS-App-Runner.pdf) covers the service in great detail.

After working with App Runner for a few months, here's a few thoughts on what it does well and not so well. For the sake of comparison, I have used Azure App Runner before, but run most of my workloads on Kubernetes.

## The Good

### Pricing

One of the best parts about App Runner is its [pricing](https://aws.amazon.com/apprunner/pricing/) **if you are working on services with relatively low frequency or irregular traffic**. Its ability to scale to 'near zero' - and to do so automatically, without any additional config - makes it a very cost-effective option in this scenario. USD 25/month for compute costs of a production-ready API is attractive.

This makes App Runner a great choice for early-stage businesses; once request volume increases and becomes more consistent, it is probably worth looking at other options.

### Integrations with AWS services

Probably not a surprise, but as a managed service App Runner plays very well with other services in AWS. Key service and application metrics and logs are automatically sent to **CloudWatch**, so monitoring the performance of the service and tracing problems of a specific deployment is very straightforward.

App Runner also integrates with **Secrets Manager** to pull in secrets and environment variables to the container environment in an easy way, using an IAM role that is assigned to the App Runner service when it is created. This means you can grant access at a very granular level, and don't need to manage any additional tools like you would with Kubernetes.

Integration with Route53 is somewhat underwhelming, as the records used by App Runner for certificate validation are not created automatically, even if you manage your domain in Route53. Still, once you do have this set up, App Runner will automatically manage TLS certificates for you with no further configuration required.

### Automatic rollbacks

When a new image is deployed to App Runner, it is started on the side but it receives no traffic from outside until it has successfully passed a health check. If the health check fails, App Runner will automatically roll back to the previous version of the service with zero downtime (in theory at least) to the end user. Most of the time this works very well, however as mentioned in the below section on the downsides, the process is not particularly fast.

### Works well with Terraform

The [AWS App Runner Terraform module](https://github.com/terraform-aws-modules/terraform-aws-app-runner) makes setting up and managing App Runner convenient. Creating anything more complex with the App Runner UI is very painful so I would highly recommend using Terraform to set up and manage the service.

## The Bad

### The "edit" button is (usually) a lie

App Runner gives you the ability to edit a deployed service *but only when it is running*. A service in a stopped state cannot be edited in any way, which is very problematic as you may run into a situation in which your service will not start successfully. In this case, you are left with no option but to delete the service and recreate it - which means at the very least dealing with a change in the DNS records of the service.

As an example, I was able to get into an irrecoverable state by removing a Secret from Secrets Manager that a deployment relied on. The next deployment failed as the Secret could not be found, and App Runner got into a ~2 hour restart/fail loop until it gave up, leaving the service in a broken stopped state that cannot be edited.  When working with App Runner, be very careful when changing external dependencies (like Secrets) that your app needs in order to start.

This [issue has been on the AWS App Runner public roadmap repo since June 2021](https://github.com/aws/apprunner-roadmap/issues/49) (2 years at the time of writing), so it is hard to say when or if it will be addressed.

### ECR only

App Runner only supports ECR as its source if you want to use container images. While probably not a deal breaker, it would be preferable to have the option to use other container registries as well.

### Speed

While App Runner does perform significantly better than Azure App Service in my experience, it is still surprisingly slow in some areas. Pulling an image from ECR within the AWS network and starting it should only take a few seconds but often ends up taking several minutes, even for simple apps that start in a second locally.

Time to 'live' deployment in my experience has been approximately 5 minutes, starting from when the container image has been fully pushed to ECR.

## The Noteworthy

### Scaling App Runner requires you to understand the application

Most services configure scaling around hardware metrics like CPU thresholds or memory usage, which feels like a natural way to do it. Without knowing anything about what an app actually does, one could still write a reasonable autoscaling policy based on hardware metrics - as an extreme example, if CPU usage is 95%, another node/instance is probably needed.

App Runner, however, uses a different scaling approach: [App Runner scales based on the number of "concurrent requests"](https://docs.aws.amazon.com/apprunner/latest/dg/manage-autoscaling.html), and this threshold can be adjusted between 1 to 200. Getting the scaling configuration right for App Runner therefore requires a bit more knowledge about the application itself, and its behaviour under certain loads with certain hardware configurations.

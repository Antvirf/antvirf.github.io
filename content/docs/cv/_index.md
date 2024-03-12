---
weight: 15
title: CV
---

# CV ⋅ Antti Viitala

## Bio

* Well-rounded software engineer with a focus infrastructure & devops
* Lead Cloud infrastructure & DevOps at Synpulse8, supporting multiple projects and internal products working primarily with Kubernetes on AWS EKS, as well as OpenShift for on-prem deployments
* ~5 years of experience within technical consulting for financial services
* Experienced and fluent at working with stakeholders of all levels across business and technology
* Effective and driven learner, able to take new topics/technologies to a working level very quickly

## Key skills

* **Kubernetes** (EKS), infrastructure-as-code (Terraform) and GitOps (Flux and ArgoCD). Everything I manage is declarative, automated and documented.

* **Experience with cloud providers**:
  * **AWS**: Most of my experience is in AWS. Primarily services around EKS and RDS, so everything related to networking (VPCs, LBs, EC2), secrets (SM and SSM), certs (ACM), IAM
  * **Azure**: A ~year's experience running an AKS cluster, basic networking and some databases. Azure AD usage and SSO w/ various apps.
  * **Others** (Only used for short term POCs and experiments): GCP, Linode, Railway, Hetzner
* **Other apps, tech platforms and tools:** GitHub Actions, Docker/containers, Kubernetes, NGINX, JFrog Artifactory, Atlassian (Jira/Confluence), Jenkins
* Decent **Linux/sysadmin** skill set, understanding of basic OS structures, perms, networking logic. Familiar primarily with Ubuntu/Debian and Amazon Linux. Experimenting with various OS', networking and virtualization in my homelab.
* **Programming languages/frameworks:**
  * Strong: Python (primarily Django) | Terraform | Bash/Shell
  * Basic: Golang | HTML/CSS/JS | Java
* **Languages:** English 🇬🇧 | Finnish 🇫🇮
* **Technical writing examples**: [AWS Postgres performance comparison](https://aviitala.com/posts/aws-rds-vs-aurora-postgresql-performance-comparison/), [Kubernetes homelab with Flux](https://aviitala.com/posts/flux-homelab/), [Searching across GH Actions workflow logs](https://aviitala.com/posts/github-actions-log-search/)
* **Personal projects**: [GitHub](https://github.com/Antvirf)

With Kubernetes I have experience primarily with EKS. I prefer to manage clusters with GitOps using [flux](https://github.com/fluxcd/flux2), though I am also familiar with [ArgoCD](https://argo-cd.readthedocs.io/en/stable/). For interacting with clusters I'm a huge fan of [K9S](https://k9scli.io/).

Some of the Kubernetes applications I have configured and have familiarity with are [kubernetes-autoscaler](https://github.com/kubernetes/autoscaler), [nginx-ingress](https://github.com/kubernetes/ingress-nginx), [cert-manager](https://github.com/cert-manager/cert-manager), [external-dns](https://github.com/kubernetes-sigs/external-dns) with Route 53 and Azure DNS, [oauth2-proxy](https://github.com/oauth2-proxy/oauth2-proxy) with Azure and GitHub providers, [redis](https://github.com/redis/redis), [loki-stack](https://artifacthub.io/packages/helm/grafana/loki-stack), [prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack), and [robusta](https://github.com/robusta-dev/robusta).

## Work experience

### **Synpulse8** [2021 - Present]

#### **Solution Architect (Vice President)** - Devops, CI/CD, cloud, system design & software development

* Design, set up and maintain central internal cloud infrastructure following the GitOps model
* Devops and infra-as-code for a mobile application with a CMS API backend (AWS App Runner / Aurora / WAF / ECR / S3 etc. managed via Terraform, GitHub Actions, JavaScript)
* Devops and infra-as-code for several microservices-based financial applications (primarily using AWS EKS / AWS Aurora DBs / S3 etc. managed via Terraform and Kubernetes, GitHub Actions)
* Devops and infra-as-code for a financial-services focused risk analytics visualization app (Azure/AWS, AKS/EKS, GitHub Actions, front: React / back: Spring, Apache Pinot, Terraform, Airflow)
* Set up, administration and operation of our CI/CD pipelines and cloud infra, for example our internal product development clusters (Azure, Kubernetes, GitHub Actions, JFrog)
* S8 Operating model definition: devops, tech stack (incl. licensing), architecture principles, security policies
* Infra, architecture and development of a [DAML](https://www.digitalasset.com/developers)-based collateral management system MVP
* Various internal initiatives: creating websites/portals/webapps with e.g. Hugo and Azure Static Web Apps

#### Solution Architect (Senior Analyst) - System design, software development

* IT Platform design for a new market / new business unit of a global Swiss private bank
* Development lead and client-facing project manager for an employee analytics tool (Python, MS Graph API)
* Architecture/REST+Kafka integration for OpenShift-based Avaloq Wealth Platform ([AWP](https://www.avaloq.com/solutions/products/avaloq-wealth))
* Led internal initiative for APAC [OpenWealth](https://openwealth.ch/) chapter, an open API standard for wealth management

### Synpulse Management Consulting [2019-2021]

#### Consultant - Technology and Robotic Process Automation (RPA) focus

* Investment Suitability for a global Swiss Private bank (HK/SG) - Regulatory and control sampling gap analysis
* Process Optimisation for a global Swiss Private Bank (HK/SG) - Prioritize automation initiatives and deliver a trade reconciliation POC with Blue Prism
* Data ingestion and Portfolio reporting for a MFO (HK) - Development lead and project manager

### Finnish Defence Forces (National service) [2018]

### Internships

- Quinlan & Associates: Analyst (2018)
  - Analyst reports [#1](https://www.quinlanandassociates.com/fools-gold/), [#2](https://www.quinlanandassociates.com/banking-on-the-cloud/)
- Bloomberg L.P.: Capstone project leader (2018)
  - [Outstanding capstone project award](https://ipo.hkust.edu.hk/whats-happening/news/award-presentation-ceremony-outstanding-rmbi-capstone-projects-20172018)
- EONIQ:  Watchmaker and crowdfunding marketer (2016)
  - [Crowdfunding campaign](https://www.indiegogo.com/projects/eoniq-custom-watches-crafted-by-your-story)

## Education & certs

* BSc. (1st hons.) Risk Management & Business Intelligence, The Hong Kong University of Science & Technology, 3.68/4.3
* 2022: [DAML Associate](https://www.digitalasset.com/developers/certifications) | [Thought Machine Vault fundamentals](https://certificates.thoughtmachine.net/f1fc025a-231d-4b62-95f8-b77e922e3e7c#gs.6rq821)
* 2020: [Aalto EE: Essentials of Leading Change](https://www.aaltoee.com/programs/essentials-of-leading-change-online) | [Blue Prism](https://www.blueprism.com/): Process Controller,Developer | [UiPath](https://www.uipath.com/) Solution Architect, Developer, Orchestrator, Implementation Methodology

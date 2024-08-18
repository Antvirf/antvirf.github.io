+++
author = "Antti Viitala"
title = "Set up GitHub Actions to authenticate to JFrog Artifactory using OIDC"
date = "2024-08-18"
description = "JFrog Terraform provider requires an access token, which can be generated on the fly when your GitHub Actions workflow runs, using OIDC. This post contains Terraform snippets and a sample Actions workflow that achieves this."
tags = [
    "infrastructure",
    "devops"
]
+++

## Why?

JFrog has some useful GitHub Actions workflows available that authenticate against Artifactory. However, these only configure either the `jf` CLI tool, or alternatively configure repository access for a particular tool or framework, like Docker registries or Maven libraries. 

[The Terraform provider for JFrog Artifactory](https://registry.terraform.io/providers/jfrog/artifactory/latest/docs) requires an access token, so if you want to automate Terraform operations to JFrog in GitHub Actions, the aforementioned authentication methods are not going to work. It would be possible to create an authentication token in JF and set it up as a GitHub Actions secret, but then we are stuck with a hardcoded credential that may expire, get deleted, or end up being misused. The alternative to this is to set up OIDC authentication between the two tools, in which case a short-lived, temporary access token is given to your workflow each time it runs.

The [official JFrog documentation for GH Actions OIDC](https://jfrog.com/help/r/jfrog-platform-administration-documentation/github-actions-oidc-integration) did not work - provider configuration and specifics of the token exchange request had to be changed, hence the need for this short post. YMMV.

## Setup: Configuring JFrog side (with Terraform)

In order to create this configuration with Terraform, I needed to first manually create an access token with sufficient administrative privileges to create these resources. Once done with this guide, these resources remain in `.tf` code intended to manage the JF instance.

```hcl
locals {
  github_org_name = "example-organization"  
}

resource "platform_oidc_configuration" "github_oidc_config" {
  name          = "jf-infra-github-oidc-config"
  description   = "OIDC config for authenticating to GiHub Actions"
  issuer_url    = "https://token.actions.githubusercontent.com/"
  provider_type = "GitHub"
  audience      = "jfrog-github"
}

resource "platform_oidc_identity_mapping" "github_oidc_identity_mapping" {
  name          = "jfrog-infrastructure-repo-gh-oidc-identity-mapping"
  description   = "GitHub OIDC group identity mapping"
  provider_name = platform_oidc_configuration.github_oidc_config.name
  priority      = 10

  claims_json = jsonencode({
    "sub" = "repo:${github_org_name}/jfrog-iac:ref:refs/heads/main",
  })

  token_spec = {
    username   = "jfrog-iac"
    scope      = "applied-permissions/admin"
    audience   = "jfrt@* jfac@* jfmc@* jfmd@* jfevt@* jfxfer@* jflnk@* jfint@* jfwks@*"
    expires_in = 300 # 5 minutes
  }
}
```

## Setup: GitHub Actions flow

Since you'll likely want to reuse the authentication steps with different workflows (e.g. `plan`, `apply`), it's configured here as a reusable workflow step in the first file below. The second file shows an example of using the step in another workflow. Note that if you changed the provider name from the example above, you must update the curl command referencing that name.

```yaml
# note that the filename has to be action.yml, the name you will
# reference in other flows is the name of the enclosing folder
# ./github/workflows/jfrog-oidc-action/action.yaml
name: JFrog OIDC authenticate action
description: Authenticate to JFrog with OIDC

outputs:
  token:
    description: JFrog Artifactory Access Token
    value: ${{ steps.token.outputs.ACCESS_TOKEN }}

runs:
  using: composite
  steps:
    - name: Get ID token
      shell: bash
      run: |
        ID_TOKEN=$(curl -sLS -H "User-Agent: actions/oidc-client" -H "Authorization: Bearer $ACTIONS_ID_TOKEN_REQUEST_TOKEN" \
          "${ACTIONS_ID_TOKEN_REQUEST_URL}&audience=jfrog-github" | jq .value | tr -d '"')
        echo "ID_TOKEN=${ID_TOKEN}" >> $GITHUB_ENV

    - name: Exchange ID token with access
      shell: bash
      id: token
      env:
        ID_TOKEN: ${{ env.ID_TOKEN }}
        JFROG_URL: "https://your_instance_name.jfrog.io/access/api/v1/oidc/token"
      run: |
        ACCESS_TOKEN=$(curl -XPOST "${JFROG_URL}" -d "{\"grant_type\": \"urn:ietf:params:oauth:grant-type:token-exchange\", \"subject_token_type\":\"urn:ietf:params:oauth:token-type:id_token\", \"subject_token\": \"$ID_TOKEN\", \"provider_name\": \"jf-infra-github-oidc-config\"}" -H "Content-Type: application/json" | jq .access_token | tr -d '"')
        echo "ACCESS_TOKEN=${ACCESS_TOKEN}" >> $GITHUB_OUTPUT
```

Using this workflow step in another workflow: 

```yaml
name: TF Apply

on: # up to you
  workflow_dispatch:

permissions: # MUST BE HERE for OIDC to work
  id-token: write
  contents: read

jobs:
  terraform-apply:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Get JFrog access token
        id: token
        uses: ./.github/workflows/jfrog-oidc-action

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3

      - name: Terraform Init & Validate
        run: terraform init && terraform validate

      - name: Terraform Apply
        env:
          JFROG_ACCESS_TOKEN: ${{ steps.token.outputs.token  }}
        run: |
          terraform apply -auto-approve
```

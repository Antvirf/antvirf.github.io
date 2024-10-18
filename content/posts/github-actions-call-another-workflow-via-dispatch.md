+++ 
date = 2022-11-22
title = "GitHub Actions: Trigger a workflow with workflow_dispatch"
description = "Using the workflow_dispatch trigger, you can trigger a workflow from another with ease using the GitHub Actions REST API. This short article covers the basic setup and workflow configurations required to achieve this."
author = "Antti Viitala"
tags = [
    "github-actions",
    "devops",
    "ci-cd"
]
images = ['images/apple-touch-icon-152x152.png','images/splash.png']
+++

## Scenario & Requirements

1. We have a part of our application in one repository, called ```web-application```.
1. This ```web-application``` repository has several workflows of its own - including a workflow called ```build``` that packages the application, and uploads it to a registry.
1. We have a separate repository that includes integration and end-to-end tests for this application, called ```web-application-tests```.
1. ```web-application-tests``` contains a workflow called ```tests``` that runs the relevant tests *on the most recently built image*.
1. We want to also be able to run the ```tests``` workflow periodically or trigger it manually.

**Objective**: Whenever the ```build``` workflow of the ```web-application``` repository finishes, we want to trigger the ```tests``` workflow from the other repository using the ```workflow_dispatch``` trigger.

## Solution overview

Our main requirement is that the ```tests``` workflow need to run *after* the ```build``` workflow. We wish to maintain the separation of the repositories, due to for example different dependencies required by an application and the testing framework(s). Therefore we need to insert a trigger at the end of the ```build``` workflow to start the ```tests``` workflow.

In order to be able to create the ```workflow_dispatch``` trigger event in the ```web-application-tests``` repository, the ```web-application``` repository and the ```build``` workflow need to have a **separately created GitHub Personal Access Token (PAT)** - as by default, the token given by GitHub to a workflow is always scoped to just the repository that contains the workflow.

In total, we then need to configure 3 things:

1. Create a PAT with the right scope, and add it as a GitHub Actions secret to the ```web-application``` repository.
1. Add a trigger step to the ```build``` workflow to launch the ```tests``` workflow with the right inputs.
1. Add ```workflow_dispatch``` as a trigger to ```tests``` workflow.

## Solution details

### Create a PAT

Follow the instructions [here](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token#creating-a-personal-access-token-classic), and make sure to tick "repo" for permissions. If you are working within a GitHub organization that uses SAML, click "Enable SSO" as shown in the instructions in steps 9 and 10.

Copy the value of the token and [save it as a GitHub Actions secret for your repository](https://docs.github.com/en/actions/security-guides/encrypted-secrets#creating-encrypted-secrets-for-a-repository), with the name ```WORKFLOW_DISPATCH_PAT```.

### Creating a ```workflow_dispatch``` event within the```build``` workflow

Using the abridged snippet from below as a guideline, make sure you configure:

1. ```BUILT_IMAGE_NAME``` as an environment variable in the ```build``` stage - you need to tell the ```tests``` workflow what image to test.
1. Replace ```REPO_OWNER```, ```REPO_NAME``` and ```WORKFLOW_FILENAME``` with your respective values in the URL of the curl command
1. Replace ```TARGET_BRANCH``` of the curl payload with the target branch of the repository that you are calling

```yaml
name: "build"
... # triggers of the build workflow
jobs:
  build: # job to build the image
    ...
  trigger-tests:
    needs: build # only run 'trigger-tests' after 'build' is complete 
    steps:
      - name: Trigger 'tests'
        env:
          # important - we must tell the tests workflow what image to test
          # this environment variable should be set by the 'build' job
          image_to_test: "${{ env.BUILT_IMAGE_NAME }}" 
        run: |
          curl -XPOST \
            -H 'Authorization:token ${{ secrets.WORKFLOW_DISPATCH_PAT }}' \
            -H "Content-Type:application/json" \
            -H "Accept:application/vnd.github" \
            https://api.github.com/repos/{REPO_OWNER}/{REPO_NAME}}/actions/workflows/{WORKFLOW_FILENAME}.yml/dispatches \
            --data '{"ref": "{TARGET_BRANCH}", "inputs": {"image_to_test":"${{env.image_to_test}}"}}'
```

### Receiving the ```workflow_dispatch``` trigger within the ```tests``` workflow

The ```tests``` workflow primarily just needs the required trigger to be included, as shown below. Make sure to use the input value in later parts of the workflow with ```${{ inputs.image_to_test }}```.

```yaml
name: "tests"
on:
  workflow_dispatch:
    inputs:
      image_to_test:
        type: string
        description: container image to test
        required: true
...
jobs:
  tests:
    steps:
      - name: "run tests"
        run: run_tests_for_image ${{ inputs.image_to_test }}
        ...
```

## References

* [Manual and REST API-based usage of ```workflow_dispatch```](https://stackoverflow.com/a/70154713)
* [Alternative example with GitHub CLI, and noting the use of PAT](https://stackoverflow.com/a/72425326)

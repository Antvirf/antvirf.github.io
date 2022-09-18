+++
author = "Antti Viitala"
title = "GitHub Actions: Dealing with a failing step in a workflow"
date = "2022-09-18"
description = "Explains how to gracefully deal with failing steps in a GitHub Actions workflow, and using failure status in the control flow."
tags = [
    "github-actions",
    "devops",
    "ci-cd"
]
+++

*Here, a 'failing step' means a step that includes a command with a non-zero exit code.*

## Simple scenario: Single-line commands

For very simple cases, using shell script notation ```do_something || do_something_else``` ('or') can be sufficient. The command after the double pipes is only executed if the first command ```do_something``` evaluates to ```false```, or more importantly for us, returns a non-zero exit code (usually indicating failure).

In the example below, the step should create a kubernetes secret *if it doesn't exist already*. To achieve this, the control flow should act as follows:

{{< mermaid >}}
graph LR
    start{Get secret: <br>$SECRET_NAME}
    continue[Continue to next step]
    start-->|Not found:<br>exit code 1|create[Create Secret]-->continue
    start-->|Success:<br>exit code 0|continue
{{< /mermaid >}}

In a GitHub Actions step, this can be done with:

```yaml
...
- name: Create kubernetes secret
  run: |
    kubectl get secret -n=$NAMESPACE $SECRET_NAME || \
    kubectl create secret <type> $SECRET_NAME -n=$NAMESPACE ...

...
```

## More involved example: Parsing error output, fix the cause, and reattempt a command

Sometimes the control flow needs to *use the contents of the error* to gracefully resolve the problem. There are a few steps that the flow needs to do in this case:

1. Define an ```id``` for the step that may fail (so that we can refer to it's outcome status later in the flow).
1. The step that may fail must have the flag ```continue-on-error: true``` defined - otherwise GitHub will stop and fail the action.
1. The failing steps needs to save the error message so that it can be processed later. This is done using the ```2> err.txt``` to save this error output to a file.
1. Following steps can use the status of the failing step as a condition for execution.

The control flow looks like this:

{{< mermaid >}}
graph LR
    fail_step[Step: Run command]
    gate{Outcome status?}
    cont[Continue flow]
    fix_step[Step: Fix issue]
    reattempt_step[Step: Reattempt command]
    final_gate{Reattempt<br>outcome?}

    fail_step --> gate
    gate-->|success|cont
    gate-->|fail|fix_step-->reattempt_step-->final_gate

    final_gate-->|success|cont
    final_gate-->|fail|fail[Fail the flow]
{{< /mermaid >}}

In a GitHub Actions syntax:

```yaml
...
- name: Attempt a step that will potentially fail
  id: potentially-failing-step
  continue-on-error: true
  run: | # FYI: commands after a failing line will NOT be executed. Hence, "Success!" would NOT be printed if your_command fails.
    echo "Attempting the step..."
    your_command 2> err.txt
    echo "Success!"

- name: Dealing with the problem that caused the error
  if: steps.potentially-failing-step.outcome =='failure'
  run: | # This is where you could deal with the problem that caused the error, reading err.txt if needed - increment version, change a feature flag, retry compilation with a different dependency etc. Perhaps you changed a parameter. The syntax below shows how to save the NEW_PARAMETER as an environment variable for use later within the flow.
    cat err.txt | something
    echo "NEW_PARAMETER=$NEW_PARAMETER" >> $GITHUB_ENV
    

- name: Reattempt the earlier step after fixing
  if: steps.potentially-failing-step.outcome =='failure'
  run: | # Final attempt at running your_command - if it still fails here, the flow fails. Log with NEW_PARAMETER for better observability.
    your_command
    echo "Step successful with new parameter: ${{env.NEW_PARAMETER}}"  
...
```

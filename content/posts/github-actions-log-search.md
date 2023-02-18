+++
author = "Antti Viitala"
title = "Searching across workflow logs with gh-actions-log-search"
date = "2023-02-18"
description = "Quick introduction on how to search for a particular term among GitHub Actions logs from multiple repositories."
tags = [
    "infrastructure",
    "devops"
]
+++

## What problem are we solving for?

With some upcoming deprecations in GitHub Actions - for example the [deprecation of `save-state` and `set-output` commands on May 31st 2023](https://github.blog/changelog/2022-10-11-github-actions-deprecating-save-state-and-set-output-commands/) - I wanted to update my workflows in advance *before* they would start failing as a result of these deprecations.

Unfortunately, for both my personal repositories and at work, the problem is that GitHub has no native way to easily search the contents of several workflows' logs across several repositories. When we're talking about tens - and at org level hundreds - of repositories, going through even just 1 flow from each repository manually is very painful.

Thankfully, GitHub's REST API is extensive and offers [endpoints to query the logs of workflow runs](https://docs.github.com/en/rest/actions/workflow-runs?apiVersion=2022-11-28#list-workflow-runs-for-a-repository). Primarily with this endpoint in mind, I set out to put together a simple set of scripts to fetch large numbers of workflow logs, and then search them for particular content - in this case, the word `deprecated`.

After about two hours of effort - and an hour-long break in between after I reached [GitHub API rate limits](https://docs.github.com/en/rest/overview/resources-in-the-rest-api?apiVersion=2022-11-28#rate-limiting) I had something workable in the contents of [this repository - gh-actions-log-search](https://github.com/Antvirf/gh-actions-log-search).

## What does [the code](https://github.com/Antvirf/gh-actions-log-search) do in detail?

If you're interested in how to *run* the code, check the [**Quick start** section of the docs](https://github.com/Antvirf/gh-actions-log-search#usage---clone-install-dependencies-run).

The flow is broken into three separate scripts, each of which builds on the results of the previous steps. The structure enforces separation of the query-heavy steps to try to reduce the chance of accidentally breaching rate limits by repeated, needless queries.

1. `1_get_repos_and_workflow_runs.py`
    1. Given your access token and repository name inclusion pattern, make 1 API call to get list of your repositories.
    1. For every repository, call the GitHub API to [get a list of workflow runs](https://docs.github.com/en/rest/actions/workflow-runs?apiVersion=2022-11-28#list-workflow-runs-for-a-repository). The tool will now truncate the response to `MAX_NUMBER_OF_RUN_LOGS`, which is set to 5 by default.
    1. At this stage, your list of repositories and list of workflow runs in them is stored as a JSON `repo_run_ids.json`. It has `X` keys, each mapped to an entry containing at most `MAX_NUMBER_OF_RUN_LOGS` entries.
1. `2_get_logs_for_workflow_runs.py`
    1. Read `repo_run_ids.json`.
    1. Make an API call for each repository's workflow run (zipped logs are saved in `./logs/`), and then extracted into `./logs_extracted/`. This operation is asynchronous.
1. `3_find_words_in_logs.py`
    1. Finally, go through every file, line-by-line, in `./logs_extracted/` and save any that contain (case-insensitively) the words provided in the `WORDS_TO_FIND` parameter in a JSON file `matched_lines.json`. This contains a dictionary where repository names are keys, and the items are lists of log-lines that contained one or more of the words provided in `WORDS_TO_FIND`.

Steps 1 and 2 both cache their outputs, so repeated runs will not make further calls. You can delete the previously cached information with `make clean`.

## What are the findings?

The output of this script for [repositories under my GitHub account](https://github.com/Antvirf?tab=repositories) is shown below (repeated lines replaced with `...`):

```json
{
    "Antvirf-webhook-forwarder-python": [
        "2023-02-15T11:09:19.1034656Z ##[warning]The `set-output` command is deprecated and will be disabled soon. Please upgrade to using Environment Files. For more information see: https://github.blog/changelog/2022-10-11-github-actions-deprecating-save-state-and-set-output-commands/\n",
        ...
    ],
    "Antvirf-gh-environment-manager": [
        "2023-02-17T08:11:37.8933768Z [2023-02-17 08:11:37] [build-stdout] [INFO] [2] Extracted file /home/runner/.cache/pypoetry/virtualenvs/gh-env-manager-ypNa47Hf-py3.10/lib/python3.10/site-packages/_pytest/deprecated.py in 35ms\n"
    ]
}
```

Looks like I need to address some of the workflows in the [webhook-forwarder-python](https://github.com/Antvirf/webhook-forwarder-python) repository.

The second finding in [gh-environment-manager](https://github.com/Antvirf/gh-environment-manager) is likely a false positive from the point of view of a deprecated GitHub Actions step, but it does show that a line with the key word - `deprecated`, in this case - was indeed picked up.

## Ideas for further work and improvement areas

From the perspective of **dependencies**, given how simple this application is, it could easily be made independent of [`PyGithub`](https://github.com/PyGithub/PyGithub) as we only make use of a very marginal portion of this (fantastic) library's features.

From the perspective of **user-friendliness**, instead of using a `makefile` as a cheap substitute for a CLI interface, something like [`Typer`](https://typer.tiangolo.com/typer-cli/) could be used instead for a better experience.

Finally, for a more platform and environment-independent approach - as well as for performance - the whole thing would probably be better off if it was written as a simple executable in `Go` and made available via a proper package manager.

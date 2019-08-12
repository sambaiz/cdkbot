# CDKBot

CDKBot is an application for Pull Request based AWS CDK operation.
Currently only GitHub is supported.

## Operations

Following commands are executable by PR comments. 
If no stacks are specified, all stacks are passed.

- `/diff [stack1 stack2 ...]`: cdk diff
- `/deploy [stack1 stack2 ...]`: cdk deploy

### FYI: Why deploys before merged, not after merged?

cdk deploy fails unexpectedly due to runtime errors of CF template and may need to be fixed.
Therefore, if it deploys when merged, incorrect codes can be mixed and one or more PRs must be opened to fix, which flagment changes. That's why execute commands on PR before merged.

## Configurations

Put `cdkbot.yml` on the repository root.

```
cdkRoot: . # CDK directory path from repository root.
targets:
  # If any key is matched the PR base branch, execute commands with contexts `-c key=value`.
  # If not, commands are not executed.
  develop:
    contexts:
      env: stg
  master:
    contexts:
      env: prd
```

## Install

WIP



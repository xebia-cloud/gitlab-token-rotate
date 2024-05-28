# token-manager

A command line utility to work safely with given tokens, directly from a secret store.

The tokens are references through a URL:


| secret store          | URL pattern                                       |
|-----------------------|---------------------------------------------------|
| 1password             | `op://<Vault name or id>/<token name or id>`      |
| Google Secret Manager | `gsm:///<secret name>`                            |
| AWS Parameter Store   | `ssm:///<parameter name>`                         |
|                       | `arn:aws:ssm:<region>:<account>:parameter/<name>` |
|

## gitlab
create and rotate Gitlab tokens
```
Usage:
  token-manager gitlab [command]

Available Commands:
  rotate      rotate the token stored in a secret store

Flags:
      --admin-token-url string   the URL to the secret containing the admin token
      --url string               to rotate the token from (default "https://gitlab.com")
```

## gitlab rotate
Reads the Gitlab token from the secret store and rotates it.

```text
Usage:
  token-manager gitlab rotate token-url [flags]

Flags:
      --duration int     of the validity of the rotated token in days (default 30)
      --project string   name of the gitlab project the token belongs to
      --group string     name of the gitlab group the token belongs to

Global Flags:
      --admin-token-url string   the URL to the secret containing the admin token
      --url string               to rotate the token from (default "https://gitlab.com")
```
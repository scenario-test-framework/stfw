# Security Policy

## Supported versions

| Version | Supported |
|---|---|
| 1.x (latest) | ✅ |
| 0.2.x (Bash + digdag, frozen) | ❌ |

Only the latest 1.x release line receives security fixes. The v0.2 line is frozen at tag `v0.2.0`
and is not maintained.

## Reporting a vulnerability

Please report security vulnerabilities through GitHub's **private vulnerability reporting**:

- Go to the [Security](https://github.com/scenario-test-framework/stfw/security) tab of this
  repository and click **Report a vulnerability**.

Do **not** open a public issue for security reports.

We aim to acknowledge a report within a few business days and to coordinate a fix and disclosure
timeline with you.

## Out of scope

The following are documented design characteristics rather than vulnerabilities:

- Credentials are resolved at runtime from the inventory (host groups) and age-encrypted secrets;
  configuration must not contain hosts or passwords. Reports about hosts/passwords that were
  hard-coded into a project's own configuration (against this contract) are out of scope.
- Process plugins execute arbitrary local executables by design (the plugin contract). Running
  untrusted scenarios/plugins is equivalent to running untrusted code and is the operator's
  responsibility.

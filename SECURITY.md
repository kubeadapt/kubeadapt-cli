# Security Policy

## Reporting a vulnerability

Report security issues to **security@kubeadapt.io**. Do not open a public issue.

Include the CLI version (`kubeadapt version`), the command you ran, and what
you observed. If the issue involves an API key, redact it - we never need the
key itself to reproduce.

We aim to acknowledge within 3 business days and to ship a fix or a mitigation
plan within 30 days.

## Supported versions

Security fixes land on the latest minor release. Older minors are not patched.

## How this CLI handles your API key

- The key is written to the config file with `0600` permissions, via a
  temp-file-and-rename so a crash cannot leave a partial or world-readable file.
- `kubeadapt auth status` masks the key; only the first and last four
  characters are printed.
- The key is sent as a `Bearer` token over HTTPS and is never logged, including
  under `--verbose`.
- `KUBEADAPT_API_KEY` takes precedence over the config file, so CI can pass a
  key without writing it to disk.

If you believe a key has been exposed, revoke it at
https://app.kubeadapt.io/settings/api-keys - revocation is immediate.

# Releasing

Releases are built and published with [GoReleaser](https://goreleaser.com), which:

- builds `linux`/`darwin` × `amd64`/`arm64` binaries
- creates a GitHub release on `nakulmanimala/sshh` with changelog + archives
- pushes an updated formula to the `nakulmanimala/homebrew-sshh` tap

Config: [`.goreleaser.yaml`](.goreleaser.yaml).

## One-time setup

GoReleaser needs a `GITHUB_TOKEN` with `repo` scope and push access to both
`nakulmanimala/sshh` and `nakulmanimala/homebrew-sshh`.

Store it locally in `.env.release` at the repo root (already gitignored —
never commit this file):

```
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

If a token stops working (401 from the GitHub API), regenerate it at
https://github.com/settings/tokens and replace the value — GitHub always
issues a brand new string on regeneration, so the old one in `.env.release`
must be swapped out too.

## Cutting a release

1. Tag the commit you want to release, following the existing `vX.Y` scheme
   (`v1.0`–`v1.6` so far — no patch component):

   ```bash
   git tag v1.7
   git push origin v1.7
   ```

2. Run GoReleaser from a clean working tree:

   ```bash
   set -a; source .env.release; set +a
   goreleaser release --clean
   ```

`main.version` is injected at build time via `-X main.version={{.Version}}`,
so `sshh --version` reports the tag after install.

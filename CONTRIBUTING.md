# Contributing

## Branching
- Use `main` as the release branch.
- Open PRs from short-lived feature branches.

## Versioning
- Version is stored in `VERSION` with format `YYYY.MM.x`.
- Any user-visible change must be reflected in `CHANGELOG.md`.

## Local checks
- Go SDK: `cd sdk/go && go build ./...`
- Node.js SDK: `node -c sdk/nodejs/index.js`
- YAML examples: ensure `integrat.yaml` is valid YAML

## Security and secrets
- Never commit real secrets.
- Keep only templates in `.env.example`.
- Use local `.env` for runtime values.

# Production deployment

Production releases are managed by `.github/workflows/deploy-production.yml`.
Do not rebuild or recreate the storefront manually during normal development.

## Triggers

- A push to `main`.
- A push to `agent/ghcr-production-deploy` while that branch is the production integration branch.
- A manual `workflow_dispatch` run in GitHub Actions.

## Deployment contract

The workflow has two isolated jobs:

1. A GitHub-hosted `ubuntu-latest` runner builds the full-stack `linux/amd64`
   image and pushes the commit-tagged image to
   `ghcr.io/dovelora/dujiao-next`. BuildKit cache is stored in GitHub Actions,
   not on the production server.
2. The `dovelora-production` self-hosted runner passes that exact image
   reference, its registry digest, and a short-lived `GITHUB_TOKEN` to the root-owned
   `/usr/local/sbin/dovelora-deploy` command. The token is read from standard
   input and used through a temporary Docker configuration that is removed at
   the end of the deployment.

The production server only pulls the published image, updates
`/Project/deployments/dujiao/deploy.env`, and recreates `dujiao` with
`--no-deps`. The pull is pinned to the digest produced by the cloud build, so
the deployed bytes cannot differ from the published artifact. The server does
not check out source code or run `docker build`.

Local and public HTTP health checks must pass before the release is accepted.
On failure, the image pointer and `dujiao` container are restored to the
previous image. The `epusdt` container ID and start time are compared before
and after every deployment; the workflow never recreates the payment service.

After a successful deployment, the server retains only the current storefront
image and the immediately previous image for rollback. Older storefront images
are removed. Database, upload, payment, Redis, PostgreSQL, and nginx data or
images are outside this cleanup scope.

## Manual recovery

Use GitHub Actions to rerun a known-good commit. The cloud build publishes that
commit's image again, and the production runner deploys it through the same
health-check and rollback path.

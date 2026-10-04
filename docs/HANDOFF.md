# Handoff

State of fetzycloudonline as of 2026-10-04. Read this before changing content or
infrastructure. `README.md` and `terraform/README.md` hold the stable how-to; this file holds
recent context and open work.

## Current state

- `main` is at `f064343` and matches `origin/main`. The working tree is clean.
- The website (Vercel, https://fetzycloud.online) and the SSH build (`ssh.fetzycloud.online`)
  both serve `f064343`. Both were checked live after deploy.

## Done in the 2026-10-01 session

### Content refresh (`f064343`)

Sources: GitHub (`gh` as `phetzy`, including orgs `getdomovoi`, `mapwright`, `3d-pass`,
`northgrade`) and the public LinkedIn view, plus answers from fetzy.

- Added project sections `osnova` and `domovoi` after `mapwright`. Text comes from the
  `getdomovoi/osnova` and `getdomovoi/domovoi` READMEs and the `@getdomovoi/osnova` npm package.
- Mapwright was acquired in 2026. The buyer must not be named. The founder role reads
  `june 2026 – 2026 · acquired`. The mapwright.io links stay.
- Header role: `Software Engineer at C1 · building osnova and Domovoi`. C1 is still current.
- Header builds line: `agent tools · map servers · CLIs · self-hosted platforms · hardware`.
  It is sized to fit an 80-column SSH session without truncation.
- Readme intro now lists seven projects.
- Open source section lists merged PRs to websurfx and jev-router.
- Stack adds Electron and tree-sitter.
- Army service stays `2012–2015` with no badges, by fetzy's decision.

Tests changed with the content: section counts (11 to 13), list positions (`1/6` to `1/8`),
the j/k selection target (`osnova`), and the regenerated TUI goldens in
`internal/tui/testdata/`. All gates passed: Vitest 52, Playwright 13, `go test ./...`,
Prettier on touched files.

### SSH deploy repair (no repo change)

- `deploy-ssh` failed at "Configure AWS credentials": AWS rejected the GitHub OIDC token.
- Cause: the IAM OIDC provider `token.actions.githubusercontent.com` was missing from the AWS
  account. `terraform/iam.tf` only reads it as a data source; this repo does not own it.
  Likely removed during Mapwright cleanup (inferred, not verified).
- Fix: recreated it by hand on the CachyOS laptop with
  `aws iam create-open-id-connect-provider --url https://token.actions.githubusercontent.com --client-id-list sts.amazonaws.com`
  using `AWS_PROFILE=mapwright`. The `resume-tui` profile (IAM user `resume-tf`) lacks
  `iam:CreateOpenIDConnectProvider` and `iam:ListRoles`.
- Re-ran the failed workflow. All steps passed.

## Environment facts

- AWS credentials exist only on the CachyOS laptop (`dfetz@dfetz-cachyos` over Tailscale,
  login shell fish, so pipe scripts through `bash -s`). The Mac has no AWS credentials.
- Profiles `resume-tui`, `mapwright`, and `northgrade` all resolve to the same AWS account,
  which also holds the state bucket `fetzer-tfstate-<account id>`. Do not print the account ID.
- The laptop clone is `~/projects/personal/fetzycloudonline`, one commit behind `origin/main`.
  Pull before running `tofu` there.
- The Mac has no known_hosts entry for `ssh.fetzycloud.online`. The laptop does.
- `deploy-ssh` runs only on changes to Go paths and `content.json`; see its `paths:` filter.

## Roadmap

Ordered by risk. None of these are started.

1. **Own the OIDC provider in code.** It is now hand-made and unmanaged, so another project's
   cleanup can delete it again and break `deploy-ssh` the same way. Move it into the Terraform
   of a project that will outlive the others (this repo or `northgrade`), import the existing
   provider rather than create a second one, and drop the "already exists" data-source comment
   in `terraform/iam.tf` if this repo takes ownership. Needs fetzy's choice of owner.
2. **Decide on the pending EC2 replacement.** `tofu plan` replaces `aws_instance.fetzer` and
   `aws_eip_association.fetzer` because the AMI follows the latest AL2023 arm64 image. Applying
   causes brief SSH downtime; `terraform/README.md` covers the redeploy that must follow.
   Options: apply on a schedule, or pin the AMI and update it deliberately. Needs approval.
3. **Fix pre-existing lint failures.** `pnpm lint` fails on Prettier formatting in `README.md`
   and `terraform/README.md` and an `exhaustive-deps` warning at
   `src/components/ListPane.tsx:53`. All predate `f064343`.
4. **Mapwright acquisition month.** Dates say `2026` only. Add the month when fetzy gives it.
5. **`northgrade`.** Private repo in an org fetzy admins. Left off the site because nothing is
   known about it. Ask before adding.
6. **Least-privilege for `resume-tf`.** If this repo takes ownership of the OIDC provider, the
   user that runs `tofu` here needs `iam:*OpenIDConnectProvider*` on that one ARN.

## Rules that carry over

- `content.json` is the single content source for both builds. Content is frozen: add only
  facts fetzy has confirmed or that come from a verifiable source.
- No AI or session attribution in commits, tags, or PRs. No em dashes in commit messages or
  READMEs.
- Ask before any `tofu apply`, AWS write, or force push.

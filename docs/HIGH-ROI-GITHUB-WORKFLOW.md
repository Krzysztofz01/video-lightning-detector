# High‑ROI GitHub Workflow

Purpose: keep `main` always releasable, iterate fast on branches, and make changes safe, reviewable, and repeatable.

## Branching Model
- `main`: protected, green, releasable.
- `next`: integration for refactors/migrations; merge to `main` when stable.
- `feature/<scope>`: short‑lived focused changes.
- `spike/<topic>`: experiments; clean before merge.
- `codex/<topic>`: AI‑assisted work; treat like feature branches.

## PR Flow
- Small PRs; Draft until ready; title uses Conventional Commits (`feat:`, `fix:`, `refactor:`, `docs:`, `chore:`).
- Require CI green, at least one review, and squash merge.

## CI Expectations
- Setup Go + ffmpeg. Jobs: `go fmt -l` (fail if diffs), `go vet`, build, `go test -race -coverprofile`.
- Upload coverage HTML; optionally nightly or manual bench to publish `cpu.prof`/`mem.prof`.

## Releases
- SemVer; tag on `main`. Automate GitHub Releases and binaries later (e.g., GoReleaser).

## Codex Branches
- Target `next` for big changes; `main` for safe fixes. Always review + CI; squash on merge.

## Repo Hygiene
- CODEOWNERS, PR template with checklist, Dependabot (Go modules, Actions), branch protection rules.

## Samples & Tests
- Keep samples small under `resources/samples/`. Consider a CI smoke test against a tiny clip.

## Roadmap & Progress (update as we go)
- [x] Document workflow in repo
- [ ] Protect `main` branch (GitHub settings)
- [ ] Create/protect `next` integration branch
- [ ] Add CI workflow (lint/vet/build/test/coverage)
- [ ] Add PR template (checklist + risk/rollback)
- [ ] Add CODEOWNERS
- [ ] Enable Dependabot (gomod, actions)
- [ ] Add CI smoke test using a sample video
- [ ] Add release automation (tags → release, optional GoReleaser)
- [ ] Add bench workflow (manual trigger)
- [ ] Require status checks on PRs (GitHub settings)

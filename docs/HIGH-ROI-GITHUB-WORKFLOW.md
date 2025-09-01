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

---

## Deep Dive: Concepts & How‑To

### Protected Branches (PRs only, required checks)
Protected branches block direct pushes and enforce quality gates so `main` stays releasable.

How to set up in GitHub UI:
1. Repo → Settings → Branches → Branch protection rules → Add rule
2. Branch name pattern: `main`
3. Enable:
   - Require a pull request before merging (set “Required approvals” to 1)
   - Require review from Code Owners (optional but recommended)
   - Require status checks to pass before merging
     - Select your checks (e.g., `ci/lint`, `ci/build`, `ci/test`)
     - Require branches to be up to date before merging
   - Require conversation resolution before merging
   - Require linear history (prevents merge commits in protected branches)
   - Do NOT allow force pushes; Do NOT allow deletions
   - Include administrators (recommended)

Result: only PRs with green CI and at least one approval can merge; the PR branch must be rebased/merged to be up‑to‑date with the target branch.

### Draft PRs and Checklists
Open PRs as Draft to signal “work in progress” and avoid premature reviews. Convert to “Ready for review” when the checklist below is satisfied.

Recommended PR checklist (add via PR template):
- [ ] Builds locally: `go build -v -o bin/video-lightning-detector .`
- [ ] Tests pass: `go test -race ./...` (attach failure context if any)
- [ ] Sample run verified: paste one command used on `resources/samples/…`
- [ ] Docs updated (README/AGENTS/workflow) if behavior/flags changed
- [ ] Risk: low/medium/high; Rollback: `git revert <merge-commit>`

### Required Reviews, Status Checks, Up‑to‑Date Branch
- Reviews: set minimum “Required approvals” to 1 (or more as the team grows). Prefer enabling “Require review from Code Owners”.
- Status checks: mark CI jobs as required to block merges on failures.
- Up‑to‑date: enable “Require branches to be up to date before merging” to re‑run CI on the merge base and prevent surprises.

### Squash Merge (clean history)
Prefer a single, descriptive commit per PR. In Repo → Settings → General → Merge button:
- Allow squash merging: ON
- Allow merge commits: OFF
- Allow rebase merging: OFF

Guidelines:
- Keep the PR title as the squash commit title (use Conventional Commits).
- Add a concise body describing why + key changes; link issues.

### Branching Strategy Details
- `next` usage: collect refactors and migrations that might conflict or churn. Keep it close to `main` by regularly syncing (`git merge origin/main` into `next`). When `next` stabilizes (green CI, reviewed), merge it into `main` via squash.
- Target selection: large or risky PRs → `next`; small, safe fixes → `main`.
- Feature flow: `feature/<scope>` → PR into `next` or `main` → squash → delete branch.
- Spikes: `spike/<topic>` for exploration. If useful, extract a clean `feature/` branch; don’t merge raw spikes.

### CI Pipeline: Practical Notes
Typical jobs (Linux runner):
- Lint/Format: fail on unformatted code (`go fmt -l .` should output nothing). Optionally add `go vet ./...`.
- Build: `go build ./...` ensures all packages compile. Use module cache.
- Test: `go test -race -coverprofile coverage.out ./...` then `go tool cover -html=coverage.out` as an artifact.
- ffmpeg: install with `sudo apt-get update && sudo apt-get install -y ffmpeg` on Ubuntu runners.

Example GitHub Actions skeleton:
```
name: CI
on:
  pull_request:
  push:
    branches: [ main, next ]
jobs:
  build-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22.x', cache: true }
      - run: sudo apt-get update && sudo apt-get install -y ffmpeg
      - name: Format check
        run: |
          df=$(go fmt ./...); if [ -n "$df" ]; then echo "$df" && exit 1; fi
      - run: go vet ./...
      - run: go build ./...
      - run: go test -race -coverprofile coverage.out ./...
      - run: go tool cover -html=coverage.out -o coverage.html
      - uses: actions/upload-artifact@v4
        with: { name: coverage, path: coverage.html }
```

Optional jobs:
- Nightly or manual benches (workflow_dispatch) that set `VLD_CLI_ARGS`, run benchmarks and upload `cpu.prof`/`mem.prof`.
- golangci-lint for deeper checks (fast preset to keep CI times reasonable).

### Releases (simple to start)
Manual lightweight flow:
1. Merge to `main` (green). Decide version bump (SemVer). For early dev, use `v0.y.z`.
2. Tag locally: `git tag -a v0.1.0 -m "v0.1.0" && git push origin v0.1.0`
3. Create a GitHub Release from the tag; paste highlights from PR titles.

Automated later: adopt Release Please or GoReleaser to build binaries for Linux/macOS/Windows and publish on tag.

### Repo Hygiene Examples
- CODEOWNERS (require your review):
```
*       @QLiMBer
```
- PR template (`.github/pull_request_template.md`):
```
Title: <feat|fix|refactor|docs|chore>: <scope>

Summary
- What & why in 1–3 bullets.

Checklist
- [ ] Builds: go build -v -o bin/video-lightning-detector .
- [ ] Tests: go test -race ./...
- [ ] Sample run pasted
- [ ] Docs updated (if flags/behavior changed)
- [ ] Risk & rollback noted
```
- Dependabot (`.github/dependabot.yml`): weekly updates for `gomod` and `github-actions`.

### Samples & CI Smoke Test
- Keep bundled samples small. For CI, run the detector on a very short clip with `-s 0.1` and skip exporting frames (`-f`) to save time.
- Example smoke step: `./bin/video-lightning-detector -i "resources/samples/sample 3.mp4" -o ./runs/ci -a -s 0.1 -f`

### Codex Branches Policy
- Create as `codex/<topic>` and target `next` for large changes.
- Treat like any feature branch: PR, CI, review, squash. Never auto‑merge; always read diffs and run samples.

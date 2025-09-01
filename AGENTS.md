# Repository Guidelines

## Project Structure & Module Organization
- `cmd/`: Cobra CLI entry (`root.go`) and flag wiring.
- `internal/`: Core packages — `detector/` (pipeline, reports), `frame/` (model/statistics), `render/` (logging/UI), `utils/` (I/O, math, image ops). Tests live alongside packages.
- `main.go`: Binary entrypoint; output binary is in `bin/` when built.
- `resources/`: Project images and example assets; avoid large media elsewhere.
- `.tooling/` + `env.sh`: Project‑local Go and ffmpeg used by the manual setup.

## Daily Use
When opening a new terminal after a restart:
- Activate local tools: `source ./env.sh` (adds `.tooling/go/bin` and `.tooling/ffmpeg` to `PATH`). Verify with `which go && which ffmpeg`.
- If the binary already exists: `./bin/video-lightning-detector -i <video.mp4> -o <out-dir> -a -s 0.4 [-n]`.
- If not built yet (or after code changes), build first: `go build -v -o bin/video-lightning-detector .` then run as above.

## Dev Pipeline (manual)
- Edit code under `internal/` or `cmd/`.
- Format and test: `go fmt ./... && go test ./...`.
- If you add imports or upgrade deps: `go mod tidy`.
- Build: `go build -v -o bin/video-lightning-detector .`.
- Run your scenario, e.g.: `./bin/video-lightning-detector -i samples/video.mp4 -o runs/test -a -s 0.4`.

## Coding Style & Naming Conventions
- Go 1.20+; format with `go fmt ./...`. Keep imports tidy.
- Packages: short, lowercase; exported identifiers `CamelCase`, unexported `camelCase`.
- Errors: wrap with `%w` and include subsystem prefix (e.g., `detector: …`).

## Testing Guidelines
- Unit tests: `go test ./...`.
- Coverage: `go test -coverprofile coverage.out ./... && go tool cover -html coverage.out`.
- Benchmarks: `export VLD_CLI_ARGS='-i input.mp4 -o out -a' && go test -bench .` (uses `BenchmarkVideoLightningDetectorFromEnvArgs`).

## Commit & Pull Request Guidelines
- Commits: imperative mood; consider `feat:`, `fix:`, `refactor:`, `docs:` when helpful.
- PRs: include purpose, key changes, flags used to verify, and sample outputs (e.g., snippet/screenshot from `chart-report.html`). Link issues.
- Don’t commit binaries, large videos, or profiles; `.gitignore` already excludes `bin/`, media, `coverage.out`, and `*.prof`.

## Security & Configuration Tips
- Ensure `ffmpeg`/`ffprobe` resolve from `.tooling/` via `env.sh`. Prefer `-s` to downscale large videos and `-n` to denoise when noise causes false positives.

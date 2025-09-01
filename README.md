<p align="center">
  <img src="https://raw.githubusercontent.com/Krzysztofz01/video-lightning-detector/development/resources/project-image-video-lightning-detector.png" width="400">
</p>

# video-lightning-detector
[![Go Reference](https://pkg.go.dev/badge/github.com/Krzysztofz01/video-lightning-detector.svg)](https://pkg.go.dev/github.com/Krzysztofz01/video-lightning-detector)
[![Go Report Card](https://goreportcard.com/badge/github.com/Krzysztofz01/video-lightning-detector)](https://goreportcard.com/report/github.com/Krzysztofz01/video-lightning-detector)
![GitHub](https://img.shields.io/github/license/Krzysztofz01/video-lightning-detector)
![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/Krzysztofz01/video-lightning-detector?include_prereleases)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/Krzysztofz01/video-lightning-detector)

**The project is in the development stage. It still requires a lot of optimization and fine-tuning.** 

This project is a CLI tool that allows to analyze video recordings in order to detect frames containing lightning strikes and to export them as images. The program analyzes all frames of the recording, taking into account three parameters:

- the perceived brightness of the frames
- the difference between adjacent frames by comparing the RGB values of individual pixels
- the difference between adjacent frames after binary thresholding.

We can enter the appropriate threshold values for the above parameters to fine-tune the detection, or we can let the program decide itself (based on all the collected data) which threshold values will be appropriate. The auto-detection system uses descriptive statistics and methods such as moving average to determine the threshold values. For a broader analysis of the recordings, it is possible to export all parameters in CSV and JSON format, which allows graph generation and further work with the data. In order to increase the precision of the detections, we can also apply de-noising, and to increase performance, we can apply frame scaling.

# Requirements and installation
Required software for the manual, self‑contained workflow:
- **git** – clone the repository.
- **curl**, **tar** – to fetch and unpack local toolchains.
- No global Go/ffmpeg required; they are installed inside the repo.

If you prefer system tools, the project also supports them, but the manual flow below is the primary path.

## Manual self‑contained setup (no system Go/ffmpeg needed)
The steps below reproduce exactly how this fork was prepared and tested locally: clone via SSH, install a project‑local Go toolchain and ffmpeg binaries inside the repo, build the binary, and run the detector.

1) Clone via SSH
```sh
# From any working directory:
git clone git@github.com:QLiMBer/video-lightning-detector.git
cd video-lightning-detector
```

2) Install project‑local tooling (Go 1.22.5 and static ffmpeg)
```sh
# Create a tooling directory inside the repo
mkdir -p .tooling

# Install Go toolchain locally
GO_VER=1.22.5
curl -fsSL -o go.tar.gz "https://go.dev/dl/go${GO_VER}.linux-amd64.tar.gz"
tar -C .tooling -xzf go.tar.gz
mv .tooling/go ".tooling/go-${GO_VER}"
ln -sfn "go-${GO_VER}" .tooling/go
rm -f go.tar.gz

# Install static ffmpeg + ffprobe locally (Linux x86_64)
mkdir -p .tooling/ffmpeg-tmp
curl -fsSL -o ffmpeg.tar.xz https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz
tar -C .tooling/ffmpeg-tmp -xJf ffmpeg.tar.xz
FF_DIR=$(find .tooling/ffmpeg-tmp -maxdepth 1 -type d -name 'ffmpeg-*static' | head -n 1)
mkdir -p .tooling/ffmpeg
cp -f "$FF_DIR/ffmpeg" "$FF_DIR/ffprobe" .tooling/ffmpeg/
rm -rf .tooling/ffmpeg-tmp ffmpeg.tar.xz

# Create an environment helper to use the local tools
cat > env.sh << 'EOF'
# Source this file to use the project-local Go and ffmpeg
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export PATH="$PROJECT_ROOT/.tooling/go/bin:$PROJECT_ROOT/.tooling/ffmpeg:$PATH"
EOF
chmod +x env.sh

# Activate local tools and verify
source ./env.sh
which go && which ffmpeg && which ffprobe
```

3) Build the detector
```sh
mkdir -p bin
go build -v -o bin/video-lightning-detector .
```

4) Run on your video(s)
```sh
# Auto-thresholds, with frame downscaling for speed
./bin/video-lightning-detector \
  -i path/to/video.mp4 \
  -o ./runs/my-video \
  -a -s 0.4

# Optional: denoise to reduce false positives
# ./bin/video-lightning-detector -i path/to/video.mp4 -o ./runs/my-video -a -n

# Optional: export CSV/JSON stats and an HTML chart
# ./bin/video-lightning-detector -i path/to/video.mp4 -o ./runs/my-video -a -e -j -r
```

Notes:
- The self‑contained tooling keeps your system clean; remove `.tooling/` to discard it.

## Daily Use (after restart)
When opening a fresh terminal to run detections:
```sh
cd video-lightning-detector
# 1) Activate local toolchains (Go + ffmpeg)
source ./env.sh
which go && which ffmpeg && which ffprobe  # quick sanity check

# 2) Run the detector (binary already built in ./bin/)
./bin/video-lightning-detector \
  -i path/to/video.mp4 \
  -o ./runs/session-001 \
  -a -s 0.4 [-n] [-e -j -r] [-v]

# If bin/ does not contain the binary yet (first run after pulling changes):
go build -v -o bin/video-lightning-detector .
```
Quick test using bundled samples:
```sh
# Paths with spaces need quoting
./bin/video-lightning-detector -i "resources/samples/sample 1.mp4" -o ./runs/sample-1 -a -s 0.4
./bin/video-lightning-detector -i "resources/samples/sample 2.mp4" -o ./runs/sample-2 -a -s 0.4
```
Outputs (inside your `-o` directory):
- Exported frames: `frame-<n>.png`.
- Reports: `frames-report.{csv,json}`, `statistics-report.{csv,json}`.
- Optional chart: `chart-report.html`.

Tips:
- Use `-s <0..1]` to trade quality for speed; `0.1–0.5` is typical.
- Add `-n` to denoise noisy footage and reduce false positives.

# Usage
All available flags/commands:
```sh
Usage:
  video-ligtning-detector [flags]

Flags:
  -a, --auto-thresholds                               Automatically select thresholds for all parameters based on calculated frame values. Values that are explicitly provided will not be overwritten.
  -t, --binary-threshold-difference-threshold float   The threshold used to determine the difference between two neighbouring frames after the binary thresholding process. Detection is credited when the value for a given frame is greater than the sum of the threshold of tripping and the moving average
  -b, --brightness-threshold float                    The threshold used to determine the brightness of the frame. Detection is credited when the value for a given frame is greater than the sum of the threshold of tripping and the moving average
  -c, --color-difference-threshold float              The threshold used to determine the difference between two neighbouring frames on the color basis. Detection is credited when the value for a given frame is greater than the sum of the threshold of tripping and the moving average.
  -n, --denoise                                       Apply de-noising to the frames. This may have a positivie effect on the frames statistics precision.
  -r, --export-chart-report                           Value indicating if the frames statistics chart in HTML format should be exported.
  -e, --export-csv-report                             Value indicating if the frames statistics report in CSV format should be exported.
  -j, --export-json-report                            Value indicating if the frames statistics report in JSON format should be exported.
  -h, --help                                          help for video-ligtning-detector
  -i, --input-video-path string                       Input video to perform the lightning detection.
  -m, --moving-mean-resolution int32                  The number of elements of the subset on which the moving mean will be calculated, for each parameter. (default 50)
  -o, --output-directory-path string                  Output directory to store detected frames.
  -s, --scaling-factor float                          The frame scaling factor used to downscale frames for better performance. (default 0.5)
  -f, --skip-frames-export                            Value indicating if the detected frames should not be exported.
  -v, --verbose                                       Enable verbose logging.
```

# Example workflow
Running the detector with default values and auto-threshold calculation. The most automated apporach.
```sh
video-lightning-detector -i ~/path/to/video.mp4 -o ~/output/directory/ -a
```

## Development Pipeline
For editing and rebuilding locally:
```sh
cd video-lightning-detector
source ./env.sh                     # enable local Go and ffmpeg

# 1) Edit code under internal/ and cmd/

# 2) Format, vet, and test
go fmt ./...
go vet ./...                        # optional but recommended
go test ./...

# 3) Maintain dependencies if imports changed
go mod tidy

# 4) Build the binary
go build -v -o bin/video-lightning-detector .

# 5) Run a test scenario
./bin/video-lightning-detector -i path/to/video.mp4 -o ./runs/dev -a -s 0.4
```

Profiling and coverage (optional):
```sh
# Bench with CPU/MEM profiles (benchmark reads args from VLD_CLI_ARGS)
export VLD_CLI_ARGS='-i path/to/video.mp4 -o ./runs/bench -a -s 0.4'
go test -v -run ^$ -bench . -cpuprofile cpu.prof -memprofile mem.prof

# Inspect profiles
go tool pprof -text cpu.prof
go tool pprof -text mem.prof

# Generate SVGs (requires Graphviz-enabled pprof build)
go tool pprof -svg -output cpu-profile.svg cpu.prof
go tool pprof -svg -output mem-profile.svg mem.prof

# Coverage HTML report
go test -coverprofile coverage.out ./...
go tool cover -html coverage.out
```

The detection takes ages to complete? Running the detector with frame scaling to improve performance.
```sh
video-lightning-detector -i ~/path/to/video.mp4 -o ~/output/directory/ -a -s 0.1
```

The recording noise or movement on the video is causing false positives? Lets additionaly apply noise reduction.
```sh
video-lightning-detector -i ~/path/to/video.mp4 -o ~/output/directory/ -a -n
```

Running the detector without exporting the frames but with CSV and JSON report export.
```sh
video-lightning-detector -i ~/path/to/video.mp4 -o ~/output/directory/ -a -f -e -j
```

Running the detector with explicit threshold values.
```sh
video-lightning-detector -i ~/path/to/video.mp4 -o ~/output/directory/ -t 0.002 -c 0.052 -b 0.035
```

Running the detector with auto-threshold but explicit forced brightness threshold.
```sh
video-lightning-detector -i ~/path/to/video.mp4 -o ~/output/directory/ -a -b 0.035
```

Running the detector with custom moving mean resolution.
```sh
video-lightning-detector -i ~/path/to/video.mp4 -o ~/output/directory/ -a -m 60
```

# Example results
Here's an example of graphs generated using the exported CSV report. The graphs contain two series: a given value for a given frame and the value of the moving mean for the neighboring 50 frames at the center point at a given location. Visible peaks indicate frames containing lightning strikes. The charts refer to the following:
- the perceived brightness of the frames
- the difference between adjacent frames by comparing the RGB values of individual pixels
- the difference between adjacent frames after binary thresholding.

![Example graph brightness](https://raw.githubusercontent.com/Krzysztofz01/video-lightning-detector/development/resources/example-graph-brightness.png)
![Example graph colordiff](https://raw.githubusercontent.com/Krzysztofz01/video-lightning-detector/development/resources/example-graph-colordiff.png)
![Example graph btdiff](https://raw.githubusercontent.com/Krzysztofz01/video-lightning-detector/development/resources/example-graph-btdiff.png)

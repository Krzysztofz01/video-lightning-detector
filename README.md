<p align="center">
  <img src="https://raw.githubusercontent.com/Krzysztofz01/video-lightning-detector/development/resources/project-image-video-lightning-detector.png" width="400">
</p>

# video-lightning-detector (vld)
[![Go Reference](https://pkg.go.dev/badge/github.com/Krzysztofz01/video-lightning-detector.svg)](https://pkg.go.dev/github.com/Krzysztofz01/video-lightning-detector)
[![Go Report Card](https://goreportcard.com/badge/github.com/Krzysztofz01/video-lightning-detector)](https://goreportcard.com/report/github.com/Krzysztofz01/video-lightning-detector)
![GitHub](https://img.shields.io/github/license/Krzysztofz01/video-lightning-detector)
![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/Krzysztofz01/video-lightning-detector?include_prereleases)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/Krzysztofz01/video-lightning-detector)

**Work on the project is still in progress. Work is ongoing to further improve the quality of the classifications made and the processing time.** 

The project aims to automate the process of analyzing video footage to detect frames that capture lightning strikes. The workflow of the program consists of several stages such as analysis, detection and export. The user of the tool indicates the selected video recording and the operating parameters of the individual stages, including options that determine whether the thresholds that define a detection are to be determined automatically or to be set manually. During the analysis, frames are analyzed in terms of:

- perceived brightness
- channel based differences between neighboring frames
- segmentation based differences between neighboring frames.

The detection stage performs a binary classification based on the calculated weights and indicates whether a given frame of footage captures a lightning strike. The program allows to export positively classified frames to image files. There is also a possibility to export various statistics in CSV and JSON format for further research. The tool is dedicated to:

- photographers
- storm chasers
- data engineers and software developers

to automate their photography work, automatic data labeling and more.

**Detailed functioning of the program and the algorithms used are described in my thesis, which will be published in the future.**

# Requirements and installation
Required software to build **vld**:
- **[git](https://git-scm.com/)** - Used to download the source code from the repository.
- **[task](https://taskfile.dev/)** - Used as the main build tool.
- **[go (version: 1.22+)](https://go.dev/)** - Used to compile the source code locally.
- **[docker](https://www.docker.com/)** or **[podman](https://podman.io/)** - (Optional) Used to build the tool image.

Required software to use **vld**:
- **[ffmpeg](https://ffmpeg.org/)** - Used as the main video decoding and frame extraction tool.

Local installation (Linux and Windows):
```sh
git clone https://github.com/Krzysztofz01/video-lightning-detector

cd video-lightning-detector

task build

./bin/vld version
```

Image installation (Linux and Windows):
```sh
git clone https://github.com/Krzysztofz01/video-lightning-detector

cd video-lightning-detector

task build:image

docker run --rm vld:latest vld version
```

# Commands and flags
All available flags/commands and flags for the video command:
```
-> vld --help

A video analysis tool that allows to detect and export frames that have captured lightning strikes.

Usage:
  vld [command]

Available Commands:
  check       Check if the environment is correctly configured.
  help        Help about any command
  stream      Perform the analysis and detection stages on continuous video stream.
  version     Print the version numbers.
  video       Perform the analysis, detection and export stage on single video.

Flags:
  -h, --help                 help for vld
  -l, --log-level loglevel   The verbosity of the log messages printed to the standard output. (default info)
```

```sh
-> vld video --help

Perform the analysis, detection and export stage on single video.

Usage:
  vld video [flags]

Flags:
  -a, --auto-thresholds                                        Automatic determination of thresholds after video analysis. The specified thresholds will overwrite those determined.
  -t, --binary-threshold-difference-threshold float            The threshold used to determine the difference between two neighbouring frames after the binary thresholding segmentation process. See the documentation for more information on detection threshold values.
  -b, --brightness-threshold float                             The threshold used to determine the brightness of the frame. See the documentation for more information on detection threshold values.
  -c, --color-difference-threshold float                       The threshold used to determine the difference between two neighbouring frames on the color basis. See the documentation for more information on detection threshold values.
      --confusion-matrix-actual-detections-expression string   Expression indicating the range of frames that should be used as actual classification. Example: 4,5,8-10,12,14
  -n, --denoise denoisealgorithm                               The use of de-noising in the form of low-pass filters. Impact on the quality of weighting determination. Values: [ stackblur16, stackblur32, none, stackblur8 ] (default none)
      --detection-bounds-expression string                     An expression indicating consecutively the coordinates of the upper left point, width and height of the cutout (bounding box) of the recording to be processed.  Example: 0:0:100:200
  -r, --export-chart-report                                    Export of frame statistics as a chart in HTML format.
      --export-confusion-matrix                                Value indicating if the frames detection classification confusion matrix should be rendered.
  -e, --export-csv-report                                      Export of reports in CSV format.
  -j, --export-json-report                                     Export of reports in JSON format.
  -h, --help                                                   help for video
  -p, --import-preanalyzed                                     Use the cached data associated with the video analysis or save it in case the video has not already been analysed.
  -i, --input-video-path string                                Input video to perform the lightning detection.
  -m, --moving-mean-resolution int32                           Resolution of the moving mean used when determining the statistics of the analysed frames. Has a direct impact on the accuracy of detection. (default 50)
  -o, --output-directory-path string                           Output directory path for export artifacts such as frames and reports in selected formats.
      --scaling-algorithm scalealgorithm                       Sampling interpolation algorithm to be used when scaling the video during analysis. Values: [ default, bilinear, bicubic, nearest, lanczos, area ] (default default)
  -s, --scaling-factor float                                   Scaling factor for the frame size of the recording. Has a direct impact on the performance, quality and processing time of recordings. (default 0.5)
  -f, --skip-frames-export                                     Skipping the step in which positively classified frames are exported to image files.
      --strict-explicit-threshold                              Omit strict validation of detection threshold ranges. (default true)

Global Flags:
  -l, --log-level loglevel   The verbosity of the log messages printed to the standard output. (default info)

```

# Example usage
Check if vld and other required binaries are configured correctly.
```sh
vld check
```

Running the detector with default values and auto-threshold calculation. The most automated approach.
```sh
vld video -i ~/path/to/video.mp4 -o ~/output/directory/ -a
```

The detection takes ages to complete? Running the detector with frame scaling to improve performance.
```sh
vld video -i ~/path/to/video.mp4 -o ~/output/directory/ -a -s 0.1
```

The recording noise or movement on the video is causing false positives? Lets additionally apply noise reduction.
```sh
vld video -i ~/path/to/video.mp4 -o ~/output/directory/ -a -n
```

Running the detector without exporting the frames but with CSV and JSON report export.
```sh
vld video -i ~/path/to/video.mp4 -o ~/output/directory/ -a -f -e -j
```

Running the detector with explicit threshold values.
```sh
vld video -i ~/path/to/video.mp4 -o ~/output/directory/ -t 0.002 -c 0.052 -b 0.035
```

Running the detector with auto-threshold but explicit forced brightness threshold.
```sh
vld video -i ~/path/to/video.mp4 -o ~/output/directory/ -a -b 0.035
```

Running the detector with custom moving mean resolution.
```sh
vld video -i ~/path/to/video.mp4 -o ~/output/directory/ -a -m 60
```
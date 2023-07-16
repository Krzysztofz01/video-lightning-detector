# video-lightning-detector

This project is a CLI tool that allows to analyse a video recording in order to find frames containing lightnings and to export them as images. When iterating through the frames of the recording, the frames are analysed according to two criteria: the brightness of the frame and the difference of the current frame in relation to the previous frame. On the basis of these criteria and user provided threshold values, the utility decides whether a given frame captured a lightning bolt. It is also possible to export a report in CSV format which shows the parameter data for all frames, making it easier to select the appropriate brightness and frame difference thresholds. 

# Requirements and installation
Required software:
- **[git](https://git-scm.com/)** - Used to download the source code from the repository.
- **[task](https://taskfile.dev/)** - Used as the main build tool. (This one is optional, the program can be built "manually")
- **[go (version: 1.19+)](https://go.dev/)** - Used to compile the source code locally.
- **[ffmpeg](https://ffmpeg.org/)** - Used by the program for frame extraction.

Installation (Linux, Windows and MacOS):
```sh
git clone https://github.com/Krzysztofz01/video-lightning-detector

cd pixel-sorter

task build
```

# Usage
```sh
Usage:
  video-ligtning-detector [flags]

Flags:
  -b, --brightness-threshold float     The threshold used to determine the brightness of the frame.
  -d, --difference-threshold float     The threshold used to determine the difference between two neighbouring frames.
  -h, --help                           help for video-ligtning-detector
  -i, --input-video-path string        Input video to perform the lightning detection.
  -o, --output-directory-path string   Output directory to store detected frames.
  -f, --skip-frames-export             Value indicating if the detected frams should not be exported.
  -r, --skip-report-export             Value indicating if the frames statistics report should not be exported.
  -v, --verbose                        Enable verbose logging.
```
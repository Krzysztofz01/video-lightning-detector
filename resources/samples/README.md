# Sample Videos

This folder contains small sample videos used for testing and demonstrations of the video-lightning-detector.

## License
- Default license for files in this folder: CC0 1.0 (Public Domain Dedication), unless a file includes its own license notice.
- CC0 1.0: https://creativecommons.org/publicdomain/zero/1.0/

## Contributor Workflow (sanitizing samples)
When preparing new samples, remove sensitive metadata and optionally also audio before adding files here.

1) Strip camera/GPS metadata on Windows using ExifTool:
```
"exiftool(-k).exe" -r -P -overwrite_original -ext mp4 -UserData:GPSCoordinates= -Keys:AndroidModel= -Keys:AndroidMake= -Composite:GPSLatitude= -Composite:GPSLongitude= -Composite:GPSPosition= "\path\to\videos\"
```
Note: This worked for me, you may need to modify it depending on what metadata you need to remove.


2) Remove audio track in WSL with ffmpeg (keeps video stream as-is):
```
for f in *.mp4; do ffmpeg -i "$f" -c copy -an "noaudio_$f"; done
```

3) Copy the sanitized files into `resources/samples/` and reference them below.

## Guidelines
- Keep files reasonably small to avoid bloating the repo; consider short clips or downscaled resolution.

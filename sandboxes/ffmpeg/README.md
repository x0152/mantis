# FFmpeg Sandbox

Host for working with media files. FFmpeg, MediaInfo, ImageMagick — video/audio conversion, image processing, media content analysis.

## System info

- OS: Ubuntu 24.04 (Noble)
- User: `mantis`
- Home directory: `/home/mantis`
- Shell: `/bin/bash`

## Preinstalled software

| Tool | Description |
|---|---|
| ffmpeg / ffprobe | Convert, trim, concatenate video and audio |
| mediainfo | Detailed media file information |
| imagemagick (convert, identify, mogrify) | Image processing |
| curl, wget | Downloads |

---

## FFmpeg

### Get file info
```bash
ffprobe -v quiet -print_format json -show_format -show_streams input.mp4
```

Short summary:
```bash
ffprobe -v quiet -show_entries format=duration,size,bit_rate -of csv=p=0 input.mp4
```

### Convert video

MP4 → WebM:
```bash
ffmpeg -i input.mp4 -c:v libvpx-vp9 -crf 30 -b:v 0 output.webm
```

Any format → MP4 (H.264, universal):
```bash
ffmpeg -i input.avi -c:v libx264 -crf 23 -c:a aac output.mp4
```

With a specific resolution:
```bash
ffmpeg -i input.mp4 -vf scale=1280:720 -c:v libx264 -crf 23 output_720p.mp4
```

### Extract audio from video
```bash
ffmpeg -i video.mp4 -vn -acodec libmp3lame -q:a 2 audio.mp3
```

To WAV (uncompressed):
```bash
ffmpeg -i video.mp4 -vn -acodec pcm_s16le audio.wav
```

### Convert audio

MP3 → AAC:
```bash
ffmpeg -i input.mp3 -c:a aac -b:a 192k output.m4a
```

Change bitrate:
```bash
ffmpeg -i input.mp3 -b:a 128k output.mp3
```

### Trim video

By time (from 00:01:00, 30 seconds duration):
```bash
ffmpeg -i input.mp4 -ss 00:01:00 -t 00:00:30 -c copy clip.mp4
```

### Concatenate videos

Create a list file:
```bash
echo "file 'part1.mp4'" > list.txt
echo "file 'part2.mp4'" >> list.txt
ffmpeg -f concat -safe 0 -i list.txt -c copy merged.mp4
```

### Extract a frame from video (screenshot)

A single frame at a given second:
```bash
ffmpeg -i video.mp4 -ss 00:00:10 -frames:v 1 frame.png
```

A frame every N seconds:
```bash
ffmpeg -i video.mp4 -vf "fps=1/10" frames_%03d.png
```

### Make a GIF from video
```bash
ffmpeg -i input.mp4 -ss 0 -t 5 -vf "fps=10,scale=480:-1" output.gif
```

### Add a watermark
```bash
ffmpeg -i input.mp4 -i watermark.png -filter_complex "overlay=10:10" output.mp4
```

### Remove audio from video
```bash
ffmpeg -i input.mp4 -an -c:v copy silent.mp4
```

### Change speed

Speed up 2x:
```bash
ffmpeg -i input.mp4 -filter:v "setpts=0.5*PTS" -filter:a "atempo=2.0" fast.mp4
```

Slow down 2x:
```bash
ffmpeg -i input.mp4 -filter:v "setpts=2.0*PTS" -filter:a "atempo=0.5" slow.mp4
```

---

## MediaInfo

### Full file info
```bash
mediainfo video.mp4
```

### Compact JSON output
```bash
mediainfo --Output=JSON video.mp4
```

### Specific fields only
```bash
mediainfo --Inform="Video;Resolution: %Width%x%Height%, Duration: %Duration/String3%, Codec: %Format%" video.mp4
```

---

## ImageMagick

### Image info
```bash
identify image.png
identify -verbose image.png
```

### Resize image

To a specific size:
```bash
convert input.png -resize 800x600 output.png
```

By width (keep aspect ratio):
```bash
convert input.png -resize 800x output.png
```

By percentage:
```bash
convert input.png -resize 50% output.png
```

### Convert format
```bash
convert input.png output.jpg
convert input.jpg output.webp
```

### Crop
```bash
convert input.png -crop 400x300+100+50 cropped.png
```
Format: `WIDTHxHEIGHT+X_OFFSET+Y_OFFSET`

### Rotate
```bash
convert input.png -rotate 90 rotated.png
```

### Overlay text onto an image
```bash
convert input.png -gravity South -pointsize 36 -fill white -annotate +0+20 "Watermark Text" output.png
```

### Create a thumbnail
```bash
convert input.png -thumbnail 200x200^ -gravity center -extent 200x200 thumb.png
```

### Batch processing (all PNGs in a folder)
```bash
mogrify -resize 800x -format jpg *.png
```

---

## Downloading files for processing

### Download a file by URL
```bash
curl -L -o video.mp4 https://example.com/video.mp4
wget -O image.png https://example.com/image.png
```

### Check file type
```bash
file downloaded_file
```

## Limitations

- No GPU — CPU-only encoding. H.264/H.265 uses `libx264`/`libx265`.
- Data is not persistent — files are deleted when the container restarts.
- For very large files, mind the container's disk space and memory limits.
- To review results, download the file to the local machine or send it to another host.

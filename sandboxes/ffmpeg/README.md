# FFmpeg Sandbox

Хост для работы с медиафайлами. FFmpeg, MediaInfo, ImageMagick — конвертация видео/аудио, обработка изображений, анализ медиаконтента.

## Системная информация

- ОС: Ubuntu 24.04 (Noble)
- Пользователь: `mantis`
- Домашняя директория: `/home/mantis`
- Shell: `/bin/bash`

## Предустановленное ПО

| Инструмент | Описание |
|---|---|
| ffmpeg / ffprobe | Конвертация, обрезка, склейка видео и аудио |
| mediainfo | Детальная информация о медиафайлах |
| imagemagick (convert, identify, mogrify) | Обработка изображений |
| curl, wget | Загрузка файлов |

---

## FFmpeg

### Получить информацию о файле
```bash
ffprobe -v quiet -print_format json -show_format -show_streams input.mp4
```

Краткая сводка:
```bash
ffprobe -v quiet -show_entries format=duration,size,bit_rate -of csv=p=0 input.mp4
```

### Конвертация видео

MP4 → WebM:
```bash
ffmpeg -i input.mp4 -c:v libvpx-vp9 -crf 30 -b:v 0 output.webm
```

Любой формат → MP4 (H.264, универсальный):
```bash
ffmpeg -i input.avi -c:v libx264 -crf 23 -c:a aac output.mp4
```

С указанием разрешения:
```bash
ffmpeg -i input.mp4 -vf scale=1280:720 -c:v libx264 -crf 23 output_720p.mp4
```

### Извлечь аудио из видео
```bash
ffmpeg -i video.mp4 -vn -acodec libmp3lame -q:a 2 audio.mp3
```

В WAV (без сжатия):
```bash
ffmpeg -i video.mp4 -vn -acodec pcm_s16le audio.wav
```

### Конвертация аудио

MP3 → AAC:
```bash
ffmpeg -i input.mp3 -c:a aac -b:a 192k output.m4a
```

Изменить битрейт:
```bash
ffmpeg -i input.mp3 -b:a 128k output.mp3
```

### Обрезать видео

По времени (с 00:01:00, длительность 30 секунд):
```bash
ffmpeg -i input.mp4 -ss 00:01:00 -t 00:00:30 -c copy clip.mp4
```

### Склеить видео

Создать файл со списком:
```bash
echo "file 'part1.mp4'" > list.txt
echo "file 'part2.mp4'" >> list.txt
ffmpeg -f concat -safe 0 -i list.txt -c copy merged.mp4
```

### Извлечь кадр из видео (скриншот)

Один кадр на определённой секунде:
```bash
ffmpeg -i video.mp4 -ss 00:00:10 -frames:v 1 frame.png
```

Кадр каждые N секунд:
```bash
ffmpeg -i video.mp4 -vf "fps=1/10" frames_%03d.png
```

### Создать GIF из видео
```bash
ffmpeg -i input.mp4 -ss 0 -t 5 -vf "fps=10,scale=480:-1" output.gif
```

### Добавить водяной знак
```bash
ffmpeg -i input.mp4 -i watermark.png -filter_complex "overlay=10:10" output.mp4
```

### Убрать звук из видео
```bash
ffmpeg -i input.mp4 -an -c:v copy silent.mp4
```

### Изменить скорость

Ускорить в 2 раза:
```bash
ffmpeg -i input.mp4 -filter:v "setpts=0.5*PTS" -filter:a "atempo=2.0" fast.mp4
```

Замедлить в 2 раза:
```bash
ffmpeg -i input.mp4 -filter:v "setpts=2.0*PTS" -filter:a "atempo=0.5" slow.mp4
```

---

## MediaInfo

### Полная информация о файле
```bash
mediainfo video.mp4
```

### Краткий JSON-вывод
```bash
mediainfo --Output=JSON video.mp4
```

### Только определённые поля
```bash
mediainfo --Inform="Video;Resolution: %Width%x%Height%, Duration: %Duration/String3%, Codec: %Format%" video.mp4
```

---

## ImageMagick

### Информация об изображении
```bash
identify image.png
identify -verbose image.png
```

### Ресайз изображения

До конкретного размера:
```bash
convert input.png -resize 800x600 output.png
```

По ширине (высота пропорционально):
```bash
convert input.png -resize 800x output.png
```

В процентах:
```bash
convert input.png -resize 50% output.png
```

### Конвертация формата
```bash
convert input.png output.jpg
convert input.jpg output.webp
```

### Обрезка (crop)
```bash
convert input.png -crop 400x300+100+50 cropped.png
```
Формат: `ШИРИНАxВЫСОТА+X_СМЕЩЕНИЕ+Y_СМЕЩЕНИЕ`

### Поворот
```bash
convert input.png -rotate 90 rotated.png
```

### Наложить текст на изображение
```bash
convert input.png -gravity South -pointsize 36 -fill white -annotate +0+20 "Watermark Text" output.png
```

### Создать миниатюру
```bash
convert input.png -thumbnail 200x200^ -gravity center -extent 200x200 thumb.png
```

### Пакетная обработка (все PNG в папке)
```bash
mogrify -resize 800x -format jpg *.png
```

---

## Загрузка файлов для обработки

### Скачать файл по URL
```bash
curl -L -o video.mp4 https://example.com/video.mp4
wget -O image.png https://example.com/image.png
```

### Проверить тип файла
```bash
file downloaded_file
```

## Ограничения

- Нет GPU — только CPU-кодирование. Для H.264/H.265 используется `libx264`/`libx265`.
- Данные не персистентны — файлы удаляются при перезапуске контейнера.
- Для очень больших файлов учитывай ограничения дискового пространства и памяти контейнера.
- Для просмотра результатов нужно скачать файл на локальную машину или передать на другой хост.

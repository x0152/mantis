FROM alpine:3.20

RUN apk add --no-cache \
        bash openssh-server openssh-client ca-certificates \
        ffmpeg mediainfo imagemagick \
        curl wget file jq \
        coreutils \
    && adduser -D -s /bin/bash mantis

COPY README.md /etc/mantis/README.md
COPY README.md /home/mantis/README.md
RUN chown mantis:mantis /home/mantis/README.md

EXPOSE 22

FROM alpine:3.20

RUN apk add --no-cache \
        openssh-server openssh-client sshpass bash \
        curl jq ca-certificates \
        coreutils grep sed gawk less \
    && adduser -D -s /bin/bash mantis

COPY bin/mantisctl /usr/local/bin/mantisctl
RUN chmod +x /usr/local/bin/mantisctl

COPY README.md /etc/mantis/README.md
COPY README.md /home/mantis/README.md
RUN chown mantis:mantis /home/mantis/README.md

EXPOSE 22

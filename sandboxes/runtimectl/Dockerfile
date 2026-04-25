FROM alpine:3.20

RUN apk add --no-cache \
        openssh-server openssh-client sshpass sudo bash \
        curl jq ca-certificates \
        coreutils grep sed gawk less \
    && ssh-keygen -A \
    && adduser -D -s /bin/bash mantis \
    && echo "mantis:mantis" | chpasswd \
    && echo "mantis ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

COPY bin/mantisctl /usr/local/bin/mantisctl
RUN chmod +x /usr/local/bin/mantisctl

COPY README.md /etc/mantis/README.md
COPY README.md /home/mantis/README.md
RUN chown mantis:mantis /home/mantis/README.md

EXPOSE 22
CMD ["/usr/sbin/sshd", "-D", "-e"]

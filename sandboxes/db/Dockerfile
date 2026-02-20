FROM ubuntu:24.04

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        openssh-server sudo \
        curl wget ca-certificates gnupg lsb-release \
        postgresql-client \
        default-mysql-client \
        redis-tools \
        sqlite3 \
        jq \
    && rm -rf /var/lib/apt/lists/* && \
    mkdir /run/sshd && \
    useradd -m -s /bin/bash mantis && \
    echo "mantis:mantis" | chpasswd && \
    adduser mantis sudo

COPY README.md /etc/mantis/README.md
COPY README.md /home/mantis/README.md
RUN chown mantis:mantis /home/mantis/README.md

EXPOSE 22
CMD ["/usr/sbin/sshd", "-D"]

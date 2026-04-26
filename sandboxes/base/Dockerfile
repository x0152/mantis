FROM ubuntu:24.04

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        openssh-server sudo ca-certificates \
        curl wget jq \
        bash coreutils findutils grep sed gawk less \
        iputils-ping iproute2 net-tools traceroute \
        dnsutils whois \
        git unzip tar gzip \
        file tree htop procps \
    && rm -rf /var/lib/apt/lists/* && \
    mkdir -p /run/sshd && \
    useradd -m -s /bin/bash mantis && \
    echo "mantis:mantis" | chpasswd && \
    adduser mantis sudo && \
    echo "mantis ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

COPY README.md /etc/mantis/README.md
COPY README.md /home/mantis/README.md
RUN chown mantis:mantis /home/mantis/README.md

EXPOSE 22
CMD ["/usr/sbin/sshd", "-D", "-e"]

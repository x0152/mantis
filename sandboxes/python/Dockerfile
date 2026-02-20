FROM python:3.12-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        openssh-server sudo \
        curl wget git \
        build-essential \
        libffi-dev libssl-dev \
    && rm -rf /var/lib/apt/lists/* && \
    mkdir -p /run/sshd && \
    useradd -m -s /bin/bash mantis && \
    echo "mantis:mantis" | chpasswd && \
    adduser mantis sudo

RUN pip install --no-cache-dir \
    ipython \
    numpy \
    pandas \
    matplotlib \
    seaborn \
    scipy \
    sympy \
    scikit-learn \
    Pillow \
    requests \
    httpx \
    beautifulsoup4 \
    lxml \
    pyyaml \
    tabulate \
    openpyxl \
    python-dateutil \
    tqdm \
    rich

COPY README.md /etc/mantis/README.md
COPY README.md /home/mantis/README.md
RUN chown mantis:mantis /home/mantis/README.md

EXPOSE 22
CMD ["/usr/sbin/sshd", "-D"]

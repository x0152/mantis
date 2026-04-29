FROM mcr.microsoft.com/playwright:v1.50.0-noble

ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright

RUN npm install -g playwright@1.50.0 && \
    apt-get update && \
    apt-get install -y openssh-server jq && \
    rm -rf /var/lib/apt/lists/* && \
    useradd -m -s /bin/bash mantis && \
    echo "PLAYWRIGHT_BROWSERS_PATH=/ms-playwright" >> /etc/environment

COPY pw-screenshot /usr/local/bin/pw-screenshot
COPY jina-read /usr/local/bin/jina-read
COPY web-search /usr/local/bin/web-search
RUN chmod +x /usr/local/bin/pw-screenshot /usr/local/bin/jina-read /usr/local/bin/web-search

COPY README.md /etc/mantis/README.md
COPY README.md /home/mantis/README.md
RUN chown mantis:mantis /home/mantis/README.md

EXPOSE 22

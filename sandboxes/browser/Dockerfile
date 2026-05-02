FROM mcr.microsoft.com/playwright:v1.50.0-noble

ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright
ENV NODE_PATH=/opt/mantis-node/node_modules
ENV PATH=/opt/mantis-node/node_modules/.bin:$PATH

RUN npm install -g playwright@1.50.0 && \
    npx playwright install chromium && \
    useradd -m -s /bin/bash mantis && \
    mkdir -p /opt/mantis-node && \
    npm install --prefix /opt/mantis-node playwright@1.50.0 cheerio && \
    mkdir -p /home/mantis/.cache && \
    ln -sfn /ms-playwright /home/mantis/.cache/ms-playwright && \
    apt-get update && \
    apt-get install -y openssh-server jq && \
    rm -rf /var/lib/apt/lists/* && \
    echo "PLAYWRIGHT_BROWSERS_PATH=/ms-playwright" >> /etc/environment && \
    echo "NODE_PATH=/opt/mantis-node/node_modules" >> /etc/environment && \
    echo "export PLAYWRIGHT_BROWSERS_PATH=/ms-playwright" >> /etc/profile.d/mantis-node.sh && \
    echo "export NODE_PATH=/opt/mantis-node/node_modules" >> /etc/profile.d/mantis-node.sh && \
    echo "export PATH=/opt/mantis-node/node_modules/.bin:\$PATH" >> /etc/profile.d/mantis-node.sh

COPY pw-screenshot /usr/local/bin/pw-screenshot
COPY jina-read /usr/local/bin/jina-read
COPY web-search /usr/local/bin/web-search
RUN chmod +x /usr/local/bin/pw-screenshot /usr/local/bin/jina-read /usr/local/bin/web-search

COPY README.md /etc/mantis/README.md
COPY README.md /home/mantis/README.md
RUN chown -R mantis:mantis /home/mantis /opt/mantis-node

EXPOSE 22

FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        openssh-server sudo ca-certificates \
        curl wget jq \
        nmap \
        dnsutils whois \
        netcat-openbsd \
        openssl \
        iputils-ping iproute2 traceroute mtr-tiny \
        python3 python3-pip python3-venv pipx \
        python3-passlib python3-bcrypt \
        hashcat john hashid \
        sqlmap \
        nikto \
        dirb gobuster ffuf whatweb wfuzz \
        dnsrecon \
        bsdmainutils \
        git unzip \
    && rm -rf /var/lib/apt/lists/* && \
    mkdir -p /run/sshd && \
    useradd -m -s /bin/bash mantis && \
    echo "mantis:mantis" | chpasswd && \
    adduser mantis sudo

RUN git clone --depth=1 https://github.com/drwetter/testssl.sh.git /opt/testssl.sh && \
    ln -s /opt/testssl.sh/testssl.sh /usr/local/bin/testssl

RUN mkdir -p /usr/share/wordlists/mantis && \
    printf '%s\n' \
        password 123456 12345678 qwerty abc123 password1 \
        admin administrator root letmein welcome monkey \
        dragon master hello login pass test guest iloveyou \
        sunshine princess qwerty123 football 111111 000000 \
        123123 1234 12345 123456789 1234567 1234567890 \
        changeme secret default toor oracle mysql \
        p@ssw0rd P@ssw0rd Password1 Welcome1 Qwerty123 \
    > /usr/share/wordlists/mantis/passwords-top.txt && \
    printf '%s\n' \
        admin login wp-admin administrator wp-login phpmyadmin \
        backup test dev api robots.txt sitemap.xml .env .git \
        config config.php .htaccess .htpasswd server-status \
        uploads upload files images img js css static assets \
        old new tmp temp private secret hidden dashboard panel \
        cpanel webmail mail phpinfo.php info.php api/v1 api/v2 \
        swagger swagger-ui health status metrics debug \
        login.php index.php index.html db database sql backup.sql \
    > /usr/share/wordlists/mantis/paths-top.txt && \
    printf '%s\n' \
        www mail ftp admin api dev test staging beta stage \
        m mobile app apps web blog shop store portal vpn \
        remote secure owa outlook ns ns1 ns2 mx mx1 smtp \
        imap pop pop3 exchange cdn static assets img images \
        media video videos stream chat forum wiki docs \
        support help status health monitor internal intranet \
    > /usr/share/wordlists/mantis/subdomains-top.txt

COPY bin/ /usr/local/bin/
RUN chmod +x /usr/local/bin/net-*

COPY README.md /etc/mantis/README.md
COPY README.md /home/mantis/README.md
RUN chown mantis:mantis /home/mantis/README.md

EXPOSE 22
CMD ["/usr/sbin/sshd", "-D"]

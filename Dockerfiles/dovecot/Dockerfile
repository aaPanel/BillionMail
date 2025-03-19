FROM debian:bookworm-slim

ARG DEBIAN_FRONTEND=noninteractive
ENV LC_ALL=C

## Copy file
COPY stop-supervisor.sh /stop-supervisor.sh
COPY dovecot.sh /dovecot.sh

# COPY debian.sources /etc/apt/sources.list.d/debian.sources

# install dovecot
RUN apt-get update && apt-get install -y --no-install-recommends --allow-downgrades --allow-remove-essential --allow-change-held-packages \
	dovecot-core \
	dovecot-dev \
	dovecot-pop3d \
	dovecot-imapd \
	dovecot-lmtpd \
	dovecot-pgsql \
	dovecot-sieve \
	dovecot-ldap \
	sasl2-bin \
	libsasl2-modules \
	postgresql-client \
	ca-certificates \
	curl \
	sudo \
	supervisor \
	rsyslog \
	tzdata \
	telnet \
	&& rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* \
	&& mkdir -p /usr/lib/dovecot/sieve \
	&& chmod +x /stop-supervisor.sh /dovecot.sh


## Copy the supervisord configuration file
COPY supervisord.conf /etc/supervisor/supervisord.conf

COPY spam-to-folder.sieve /usr/lib/dovecot/sieve/spam-to-folder.sieve
COPY report-spam.sieve /usr/lib/dovecot/sieve/report-spam.sieve
COPY report-ham.sieve /usr/lib/dovecot/sieve/report-ham.sieve
COPY sa-learn-spam.sh /usr/lib/dovecot/sieve/sa-learn-spam.sh
COPY sa-learn-ham.sh /usr/lib/dovecot/sieve/sa-learn-ham.sh

ENTRYPOINT ["/dovecot.sh"]
# EXPOSE 110 143 993 995

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]

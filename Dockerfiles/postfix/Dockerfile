FROM debian:bookworm-slim

ARG DEBIAN_FRONTEND=noninteractive
ENV LC_ALL=C

RUN dpkg-divert --local --rename --add /sbin/initctl \
	&& dpkg-divert --local --rename --add /usr/bin/ischroot \
	&& ln -sf /bin/true /sbin/initctl \
	&& ln -sf /bin/true /usr/bin/ischroot

## Copy postfix file
COPY postfix.sh /postfix.sh
COPY stop-supervisor.sh /stop-supervisor.sh

# COPY debian.sources /etc/apt/sources.list.d/debian.sources

# Add groups and users install Postfix
RUN groupadd -g 102 postfix \
  && groupadd -g 103 postdrop \
  && useradd -g postfix -u 101 -d /var/spool/postfix -s /usr/sbin/nologin postfix \
  && apt-get update && apt-get install -y --no-install-recommends --allow-downgrades --allow-remove-essential --allow-change-held-packages \
	ca-certificates \
	curl \
	dirmngr \
	dnsutils \
	gnupg \
	libsasl2-modules \
	postgresql-client \
	postfix \
	postfix-pgsql \
	postfix-pcre \
	sasl2-bin \
	sudo \
	supervisor \
	rsyslog \
	telnet \
	net-tools \
	tzdata \
	&& rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* \
	&& chmod +x /postfix.sh /stop-supervisor.sh

## Copy the supervisord configuration file

COPY supervisord.conf /etc/supervisor/supervisord.conf

# EXPOSE 25 465 587

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]

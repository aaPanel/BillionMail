#!/bin/bash


chmod 755 /var/lib/rspamd
chown -R _rspamd:_rspamd /var/lib/rspamd

cat <<EOF > /etc/rspamd/local.d/redis.conf
servers = "redis:6379"; # Read servers (unless write_servers are unspecified)
write_servers = "redis:6379"; # Servers to write data
disabled_modules = ["ratelimit"]; # List of modules that should not use redis from this section
timeout = 10s;
db = "0";
password = "${REDISPASS}";
EOF

exec "$@"

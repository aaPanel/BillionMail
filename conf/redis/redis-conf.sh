#!/bin/sh

# Generate redis.conf for Redis
cat <<EOF > /redis.conf
requirepass $REDISPASS
EOF

# Start Redis
exec redis-server /redis.conf

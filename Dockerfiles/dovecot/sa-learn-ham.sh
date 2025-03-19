#!/bin/sh
# exec /usr/bin/rspamc -h rspamd:11334 learn_ham


USERNAME="$1"
echo "$USERNAME" >> /tmp/learnham

/usr/bin/curl -X POST --data-binary @- --unix-socket /var/lib/rspamd/rspamd.sock http://rspamd:11334/learnham

#!/bin/bash

useradd -r -u 150 -g mail -d /var/vmail -s /sbin/nologin -c "Virtual Mail User" vmail
chown -R vmail:mail /var/vmail



cat <<EOF > /etc/dovecot/conf.d/dovecot-sql.conf.ext
driver = pgsql
connect = host=pgsql dbname=${DBNAME} user=${DBUSER} password=${DBPASS}

default_pass_scheme = MD5-CRYPT

user_query = SELECT '/var/vmail/%d/%n' as home, 'maildir:/var/vmail/%d/%n' as mail, 150 AS uid, 8 AS gid, CONCAT('dirsize:storage=', quota) AS quota FROM mailbox WHERE username = '%u' AND active = 1

password_query = SELECT username as user, password, '/var/vmail/%d/%n' as userdb_home, 'maildir:/var/vmail/%d/%n' as userdb_mail, 150 as userdb_uid, 8 as userdb_gid FROM mailbox WHERE username = '%u' AND active = 1


EOF

chmod +x /usr/lib/dovecot/sieve/sa-learn-spam.sh
chmod +x /usr/lib/dovecot/sieve/sa-learn-ham.sh
sievec /usr/lib/dovecot/sieve/spam-to-folder.sieve
sievec /usr/lib/dovecot/sieve/report-spam.sieve
sievec /usr/lib/dovecot/sieve/report-ham.sieve

exec "$@"
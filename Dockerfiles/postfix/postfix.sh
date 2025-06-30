#!/bin/bash

trap "postfix stop" EXIT


cat <<EOF > /etc/postfix/btrule.cf

user = ${DBUSER}
password = ${DBPASS}
hosts = pgsql
dbname = ${DBNAME}

query = select s.goto from( select address, goto, 1 as stype from alias union select username,username,2 as stype from mailbox union select address, goto, 3 as stype from alias_domain a left join alias b on b.address = '@' || a.alias_domain and a.alias_domain ='%d' order by stype) s where s.address='%s' or s.stype=3 limit 0,1

EOF

cat <<EOF > /etc/postfix/sql/pgsql_virtual_alias_domain_catchall_maps.cf

user = ${DBUSER}
password = ${DBPASS}
hosts = pgsql
dbname = ${DBNAME}

query = SELECT goto FROM alias,alias_domain WHERE alias_domain.alias_domain = '%d' and alias.address = '@' || alias_domain.target_domain AND alias.active = 1 AND alias_domain.active = 1

EOF

cat <<EOF > /etc/postfix/sql/pgsql_virtual_alias_domain_mailbox_maps.cf
user = ${DBUSER}
password = ${DBPASS}
hosts = pgsql
dbname = ${DBNAME}

query = SELECT maildir FROM mailbox,alias_domain WHERE alias_domain.alias_domain = '%d' and mailbox.username = '%u' || '@' || alias_domain.target_domain AND mailbox.active = 1 AND alias_domain.active = 1

EOF

cat <<EOF > /etc/postfix/sql/pgsql_virtual_alias_domain_maps.cf
user = ${DBUSER}
password = ${DBPASS}
hosts = pgsql
dbname = ${DBNAME}

query = SELECT goto FROM alias,alias_domain WHERE alias_domain.alias_domain = '%d' and alias.address = '%u' || '@' || alias_domain.target_domain AND alias.active = 1 AND alias_domain.active = 1

EOF


cat <<EOF > /etc/postfix/sql/pgsql_virtual_alias_maps.cf
user = ${DBUSER}
password = ${DBPASS}
hosts = pgsql
dbname = ${DBNAME}

query = (select username from mailbox where username like '%s' and active = 1 limit 1) union (select goto from alias where address like '%s' and active = 1 limit 1)

EOF



cat <<EOF > /etc/postfix/sql/pgsql_virtual_domains_maps.cf
user = ${DBUSER}
password = ${DBPASS}
hosts = pgsql
dbname = ${DBNAME}

query = SELECT domain FROM domain WHERE domain='%s' AND active = 1

EOF



cat <<EOF > /etc/postfix/sql/pgsql_virtual_mailbox_maps.cf
user = ${DBUSER}
password = ${DBPASS}
hosts = pgsql
dbname = ${DBNAME}

query = SELECT maildir FROM mailbox WHERE username='%s' AND active = 1

EOF


# Append myhostname and User configuration
if [ -z "${BILLIONMAIL_HOSTNAME}" ]; then
  BILLIONMAIL_HOSTNAME=mail.example.com
fi


if [ -f "/etc/postfix/conf/extra.cf" ]; then
  # Delete all contents of Overrides-configuration matching line to the end of the file
  sed '/Overrides-configuration/q' /etc/postfix/main.cf > /tmp/main.cf.tmp

  if [ -s "/tmp/main.cf.tmp" ]; then
      cat /tmp/main.cf.tmp > /etc/postfix/main.cf
      # rm -f /tmp/main.cf.tmp
      echo >> /etc/postfix/main.cf
      echo -e "\n# User Overrides-configuration" >> /etc/postfix/main.cf
      # Append User configuration
      sed -i '/\$myhostname/! { /myhostname/d }' /etc/postfix/conf/extra.cf
      echo -e "myhostname = ${BILLIONMAIL_HOSTNAME}\n$(cat /etc/postfix/conf/extra.cf)" > /etc/postfix/conf/extra.cf
      cat /etc/postfix/conf/extra.cf >> /etc/postfix/main.cf
      rm -f /etc/postfix/conf/extra.cf
  else
      echo "Rewriting /etc/postfix/main.cf failed, /tmp/main.cf.tmp is empty."
  fi
fi

CHECK_MYHOSTNAME=$(grep "myhostname" /etc/postfix/main.cf | grep -v '$myhostname')
if [ -z "${CHECK_MYHOSTNAME}" ]; then
  echo >> /etc/postfix/main.cf
  echo -e "\n# User Overrides-configuration" >> /etc/postfix/main.cf
  echo "myhostname = ${BILLIONMAIL_HOSTNAME}" >> /etc/postfix/main.cf
fi

if [ ! -f "/etc/postfix/conf/vmail_ssl.map" ]; then
  touch /etc/postfix/conf/vmail_ssl.map
fi
postmap -F hash:/etc/postfix/conf/vmail_ssl.map;



# Fix Postfix permissions
chgrp -R postdrop /var/spool/postfix/public
chgrp -R postdrop /var/spool/postfix/maildrop
#postfix set-permissions

if [ -e "/var/spool/postfix/pid/master.pid" ]; then
  rm -rf /var/spool/postfix/pid/master.pid
fi

if [ -d "/var/spool/postfix/" ]; then
  [ ! -d "/var/spool/postfix/dev/" ] && mkdir /var/spool/postfix/dev
fi

# Check Postfix configuration
postconf -c /etc/postfix/ > /dev/null

if [[ $? != 0 ]]; then
  echo "Postfix configuration error, Startup failed."
  exit 1
else
  postfix start-fg
fi

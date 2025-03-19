#!/bin/bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin
export PATH

# Add domain name and email

CONTAINER_PROJECT_NAME=billionmail
PGSQL_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-pgsql-billionmail-1"
DOVECOT_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-dovecot-billionmail-1"
create_time=$(date +%s)
time=$(date +%Y_%m_%d_%H_%M_%S)


if [ $(whoami) != "root" ];then
    echo -e "Non-root install, please try the following solutions: \n   1.Please switch to root user install \n   2.Try executing the following install commands: \n     sudo bash $0 $@"
    exit 1
fi

if [ ! -s ".env" ]; then
    ls -al
    echo " The .env file does not exist. Cannot continue operation, please operate in the project directory"
    exit 1
fi
. .env
if [ -z "${DBUSER}" ] || [ -z "${DBNAME}" ] || [ -z "${DBPASS}" ]; then
    echo "The obtained .env configuration exception is empty."   
    exit 1
fi

Red_Error(){
    echo '=================================================';
    printf '\033[1;31;40m%b\033[0m\n' "$@";
    # GetSysInfo
    exit 1;
}

Init_Email() {
    input="$2"

    # Regular expression verification email address format
    if [[ "$input" =~ ^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$ ]]; then
        mailbox=$(echo "$input" | cut -d '@' -f 1 | tr '[:upper:]' '[:lower:]')
        BILLIONMAIL_HOSTNAME=$(echo "$input" | cut -d '@' -f 2 | tr '[:upper:]' '[:lower:]')
        
    else
        echo "Enter an email address that is not a legal one: $input"
        echo "Example: user@example.com"
        exit 1
    fi
}


Init_Domain() {
    BILLIONMAIL_HOSTNAME="$2"

    # Regular expression check domain name format
    if [[ ! "${BILLIONMAIL_HOSTNAME}" =~ ^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9](\.[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9])*$ ]]; then
        echo -e "\e[31m(${BILLIONMAIL_HOSTNAME}) is not a FQDN!\e[0m"
        echo "Please change it to a FQDN"
        exit 1
    elif [[ "${BILLIONMAIL_HOSTNAME: -1}" == "." ]]; then
        echo "(${BILLIONMAIL_HOSTNAME}) is ending with a dot. This is not a valid FQDN!"
        exit 1

    fi

    # Convert to lowercase
    BILLIONMAIL_HOSTNAME=$(echo "${BILLIONMAIL_HOSTNAME}" | tr '[:upper:]' '[:lower:]')

}

Select_Domain_data(){

    # Check_domain=$(docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "SELECT * FROM domain;")
    # echo "${Check_domain}"
    Check_domain=$(docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "SELECT domain FROM domain;")
    echo "${Check_domain}"
}

Select_Email_data(){

    Check_mailbox=$(docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "SELECT username FROM mailbox;")
    echo "${Check_mailbox}"
}

Add_Domain() {

    echo "Creating domain..."
    
    Check_domain=$(docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "SELECT * FROM domain WHERE domain = '${BILLIONMAIL_HOSTNAME}';" | grep -w "^ ${BILLIONMAIL_HOSTNAME}")
    if [ -z ${Check_domain} ]; then
        # Create a domain
        
        docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "INSERT INTO domain (domain, a_record, mailboxes, mailbox_quota, quota, rate_limit, create_time, active)
        VALUES ('${BILLIONMAIL_HOSTNAME}', 'mail.${BILLIONMAIL_HOSTNAME}', 500, 5368709120, 5368709120, 12, ${create_time}, 1);"
        if [ $? -eq 0 ]; then
            echo "${BILLIONMAIL_HOSTNAME} Domain creation was successful!"
        else
            Red_Error "Domain creation failed!"
        fi
    else
        echo ""${Check_domain}" Domain already exists!"
    fi

    IPV4_ADDRESS=$(curl -sS -4 --connect-timeout 10 -m 20 https://ifconfig.me)
    if [ -z "${IPV4_ADDRESS}" ]; then
        IPV4_ADDRESS=$(curl -sSk --connect-timeout 10 -m 20 https://www.aapanel.com/api/common/getClientIP)
    fi
    ipv4_regex="^([0-9]{1,3}\.){3}[0-9]{1,3}$"
    if [[ ${IPV4_ADDRESS} =~ ${ipv4_regex} ]]; then        
        echo "${IPV4_ADDRESS}" >/dev/null 2>&1
    elif [ -z "${IPV4_ADDRESS}" ]; then
        IPV4_ADDRESS="YOUR_SERVER_IPV4_ADDRESS"
    fi
    echo -e "Please add the following record to your domain name"
    echo -e "======================================================="
    echo -e " Type | Host record       |    IPv4 address      |"
    echo -e "  A   | mail.${BILLIONMAIL_HOSTNAME} | ${IPV4_ADDRESS} |"
    echo -e "======================================================="
    echo -e " Type | Host record |    Record value    |     MX priority"
    echo -e "  MX  |     @       | mail.${BILLIONMAIL_HOSTNAME} |  10"
    echo -e "======================================================="

}

Del_Domain() {
    echo "Deleting domain..."

    Check_domain=$(docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "SELECT * FROM domain WHERE domain = '${BILLIONMAIL_HOSTNAME}';" | grep -w "^ ${BILLIONMAIL_HOSTNAME}")
    
    if [ -z "${Check_domain}" ]; then
        echo "Domain '${BILLIONMAIL_HOSTNAME}' does not exist!"
    else
        # Delete the domain
        docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "DELETE FROM domain WHERE domain = '${BILLIONMAIL_HOSTNAME}';"
        if [ $? -eq 0 ]; then
            echo "Domain '${BILLIONMAIL_HOSTNAME}' deleted successfully!"
        else
            Red_Error "Failed to delete domain '${BILLIONMAIL_HOSTNAME}'!"
        fi
    fi
}


Add_Email() {
    echo "Creating mailbox..."

    Check_domain=$(docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "SELECT * FROM domain WHERE domain = '${BILLIONMAIL_HOSTNAME}';" | grep -w "^ ${BILLIONMAIL_HOSTNAME}")
    if [ -z "${Check_domain}" ]; then
        Red_Error "Domain '${BILLIONMAIL_HOSTNAME}' does not exist, please create it first!"
    fi

    # Create a mailbox
    if [ -z "${mailbox}" ]; then
        mailbox=$(LC_ALL=C </dev/urandom tr -dc a-z0-9 2> /dev/null | head -c 6)
        echo "Generate mailbox: ${mailbox}"
    else
        mailbox=$(echo "${mailbox}" | tr '[:upper:]' '[:lower:]')
    fi
    Generate_mailbox_password=$(LC_ALL=C </dev/urandom tr -dc A-Za-z0-9 2> /dev/null | head -c 16)
    #echo "Generate_mailbox_password: ${Generate_mailbox_password}"
    Encrypt_mailbox_password=$(docker exec -i ${DOVECOT_CONTAINER_NAME} doveadm pw -s MD5-CRYPT -p "${Generate_mailbox_password}" | sed 's/{MD5-CRYPT}//')
    if [ $? -eq 0 ]; then
        echo "Generate_mailbox_password: ${Generate_mailbox_password}"
        echo "mailbox_password: ${Encrypt_mailbox_password}"
    else
        # Generate the default password after failure: BILLIONMAIL
        Generate_mailbox_password="BILLIONMAIL"
        Encrypt_mailbox_password='$1$ELBUCcYE$TbdGKBvLkFbjQguDbi3s01'
        echo "Generate_mailbox_password--default: ${Generate_mailbox_password}"
        Default_password=1
    fi

    if [ "${Default_password}" != 1 ]; then
        # password_encode 
        # Base64 encoding of strings
        b64_data=$(echo -n "${Generate_mailbox_password}" | base64)
        # Convert characters to hexadecimal
        password_encode=$(echo -n "${b64_data}" | while IFS= read -r -n1 char; do printf "%02x" "'$char"; done)
        echo "password_encode: ${password_encode}"
        if [ -z ${password_encode} ]; then
            b64_data=$(echo -n "${Generate_mailbox_password}" | base64)
            password_encode=$(echo -n "${b64_data}" | od -A n -t x1 | tr -d ' \n')
            echo "password_encode: ${password_encode}"
        fi
    else
        password_encode="516b6c4d54456c50546b31425355773d"
    fi

    Check_mailbox=$(docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "SELECT * FROM mailbox WHERE username = '${mailbox}@${BILLIONMAIL_HOSTNAME}';" | grep -w "${mailbox}@${BILLIONMAIL_HOSTNAME}")
    if [ -z "${Check_mailbox}" ]; then
        INSERT_mailbox='INSERT INTO mailbox (username, password, password_encode, full_name, is_admin, maildir, quota, local_part, domain, create_time, update_time, active)
        VALUES (
            '\'${mailbox}@${BILLIONMAIL_HOSTNAME}\'',
            '\'${Encrypt_mailbox_password}\'',
            '\'${password_encode}\'',
            '\'${mailbox}\'',
            0,
            '\'${mailbox}@${BILLIONMAIL_HOSTNAME}/\'',
            5368709120,
            '\'${mailbox}\'',
            '\'${BILLIONMAIL_HOSTNAME}\'',
            '${create_time}',
            '${create_time}',
            1
        );'
        echo "$INSERT_mailbox"
        docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "$INSERT_mailbox"
        if [ $? -eq 0 ]; then
            echo "Mailbox creation was successful!"
        else
            Red_Error "Mailbox creation failed!"
        fi
        #docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "SELECT * FROM mailbox WHERE username = '${mailbox}@${BILLIONMAIL_HOSTNAME}';"
    else
        echo ""${Check_mailbox}" Mailbox already exists!"
    fi
    
    echo -e "\e[31mPlease save the following information:\e[0m"
    echo -e "Mailbox (e-mail): \e[33m${mailbox}@${BILLIONMAIL_HOSTNAME}\e[0m \nPassword: \e[33m${Generate_mailbox_password}\e[0m" 
}


Del_Email() {
    
    echo "Deleting mailbox..."

    Check_mailbox=$(docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "SELECT * FROM mailbox WHERE username = '${mailbox}@${BILLIONMAIL_HOSTNAME}';" | grep -w "${mailbox}@${BILLIONMAIL_HOSTNAME}")
    
    if [ -z "${Check_mailbox}" ]; then
        echo "Mailbox '${mailbox}@${BILLIONMAIL_HOSTNAME}' does not exist!"
    else
        docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "DELETE FROM mailbox WHERE username = '${mailbox}@${BILLIONMAIL_HOSTNAME}';"
        if [ $? -eq 0 ]; then
            echo "Mailbox '${mailbox}@${BILLIONMAIL_HOSTNAME}' deleted successfully!"
        else
            Red_Error "Failed to delete mailbox '${mailbox}@${BILLIONMAIL_HOSTNAME}'!"
        fi
    fi
}

case "$1" in
    h | -h | help | --help | -help)
        echo "Help Information:"
        echo "  add-domain <domain>    - Add domain. Example: $0 add-domain example.com"
        echo "  del-domain <domain>    - Delete domain. Example: $0 del-domain example.com"
        echo "  add-email <email>      - Add email. Example: $0 add-email user@example.com"
        echo "  del-email <email>      - Delete email. Example: $0 del-email user@example.com"
        echo "  show-domain            - Show domain data."
        echo "  show-email             - Show email data."
        ;;
    add-domain)
        Init_Domain "$@"
        Add_Domain
        ;;
    del-domain)
        Init_Domain "$@"
        Del_Domain
        ;;
    add-email)
        Init_Email "$@"
        Add_Email
        ;;
    del-email)
        Init_Email "$@"
        Del_Email
        ;;
    show-domain)
        Select_Domain_data
        ;;
    show-email)
        Select_Email_data
        ;;
    *)
        echo "Usage: $0 {add-domain|del-domain|add-email|del-email|show-domain|show-email|help}"
        exit 1
        ;;
esac
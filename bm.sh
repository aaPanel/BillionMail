#!/bin/bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin
export PATH

# Add domain name and email

CONTAINER_PROJECT_NAME=billionmail
PGSQL_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-pgsql-billionmail-1"
DOVECOT_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-dovecot-billionmail-1"
POSTFIX_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-postfix-billionmail-1"
RSPAMD_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-rspamd-billionmail-1"
create_time=$(date +%s)
time=$(date +%Y_%m_%d_%H_%M_%S)

if [ $(whoami) != "root" ];then
    echo -e "Non-root install, please try the following solutions: \n   1.Please switch to root user install \n   2.Try executing the following install commands: \n     sudo bash $0 $@"
    exit 1
fi

PWD_d=`pwd`

SWITCH_TO_BILLIONMAIL_DIR(){
    if [ -f "/opt/PWD-Billion-Mail.txt" ]; then
        DIR=$(cat /opt/PWD-Billion-Mail.txt)
        if [ -d "${DIR}" ]; then
            cd "${DIR}"
            echo "Enter the BillionMail project directory: ${DIR}"
        fi
    fi
}

if [ -s ".env" ]; then
   CHECK_BILLIONMAIL=$(grep "BILLIONMAIL_HOSTNAME" .env)
   if [ -z "${CHECK_BILLIONMAIL}" ]; then
        SWITCH_TO_BILLIONMAIL_DIR
   fi
else
    SWITCH_TO_BILLIONMAIL_DIR
fi

if [ ! -s ".env" ]; then
    ls -al
    echo " The .env file does not exist. Cannot continue operation, please operate in the BillionMail project directory"
    exit 1
fi

source .env
if [ -z "${DBUSER}" ] || [ -z "${DBNAME}" ] || [ -z "${DBPASS}" ]; then
    echo "The obtained .env configuration exception is empty."   
    exit 1
fi

Red_Error(){
    echo '=================================================';
    printf '\033[1;31;40m%b\033[0m\n' "$@";
    exit 1;
}

Command_Exists() {
    command -v "$@" >/dev/null 2>&1
}


Docker_Compose_Check(){

    if Command_Exists docker-compose; then
        DOCKER_COMPOSE="docker-compose"
    else 
        if Command_Exists docker; then
            Docker_compose="docker compose version"
            if $Docker_compose; then
                DOCKER_COMPOSE="docker compose"
            fi
        else
            Red_Error "ERROR: Docker Compose does not exist"
        fi

    fi
}
Docker_Compose_Check

GET_SERVICE_NAME() {

    # ALL_SERVICE_NAME=$(${DOCKER_COMPOSE} config --services |grep "${SERVICE}")
    # echo "${ALL_SERVICE_NAME}"

    if [[ "${SERVICE}" == "core" ]] || [[ "${SERVICE}" == "manage" ]]; then
        #echo "Getting the "${SERVICE}" service..."
        SERVICE_NAME=$(${DOCKER_COMPOSE} ps -a --format " {{.Service}} {{.ID}} {{.Image}}" |grep "/core:" | awk '{print $1}' )
    
    elif [[ "${SERVICE}" == "postfix" ]] || [[ "${SERVICE}" == "postfix-billionmail" ]]; then
        SERVICE_NAME="postfix-billionmail"
    
    elif [[ "${SERVICE}" == "dovecot" ]] || [[ "${SERVICE}" == "dovecot-billionmail" ]]; then
        SERVICE_NAME="dovecot-billionmail"
    
    elif [[ "${SERVICE}" == "rspamd" ]] || [[ "${SERVICE}" == "rspamd-billionmail" ]]; then
        SERVICE_NAME="rspamd-billionmail"
    
    elif [[ "${SERVICE}" == "pgsql" ]] || [[ "${SERVICE}" == "postgres" ]] || [[ "${SERVICE}" == "pgsql-billionmail" ]]; then
        SERVICE_NAME="pgsql-billionmail"
    
    elif [[ "${SERVICE}" == "redis" ]] || [[ "${SERVICE}" == "redis-billionmail" ]]; then
        SERVICE_NAME="redis-billionmail"

    elif [[ "${SERVICE}" == "webmail" ]] || [[ "${SERVICE}" == "roundcube" ]] || [[ "${SERVICE}" == "webmail-billionmail" ]]; then
        SERVICE_NAME="webmail-billionmail"
    
    else
        echo "Please use: core|postfix|dovecot|rspamd|pgsql|redis|webmail"
        Red_Error "ERROR: The "${SERVICE}" service does not exist"

    fi

    if [ -z "${SERVICE_NAME}" ]; then
        Red_Error "ERROR: The "${SERVICE}" service does not exist"
    fi

}


GET_CONTAINER_ID() {

    if [ -z "${CONTAINER}" ]; then
        CONTAINER="$1"
    fi

    if [[ "${CONTAINER}" == "core" ]] || [[ "${CONTAINER}" == "manage" ]]; then
        CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "/core:" | awk '{print $1}' )
    
    elif [[ ""${CONTAINER}"" == "postfix" ]] || [[ ""${CONTAINER}"" == "postfix-billionmail" ]]; then
        CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "/postfix:" | awk '{print $1}' )
    
    elif [[ ""${CONTAINER}"" == "dovecot" ]] || [[ ""${CONTAINER}"" == "dovecot-billionmail" ]]; then
        CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "/dovecot:" | awk '{print $1}' )
    
    elif [[ ""${CONTAINER}"" == "rspamd" ]] || [[ ""${CONTAINER}"" == "rspamd-billionmail" ]]; then
        CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "/rspamd:" | awk '{print $1}' )
    
    elif [[ ""${CONTAINER}"" == "pgsql" ]] || [[ ""${CONTAINER}"" == "postgres" ]] || [[ ""${CONTAINER}"" == "pgsql-billionmail" ]]; then
        CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "postgres:" | awk '{print $1}' )
    
    elif [[ ""${CONTAINER}"" == "redis" ]] || [[ ""${CONTAINER}"" == "redis-billionmail" ]]; then
        CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "redis:" | awk '{print $1}' )

    elif [[ ""${CONTAINER}"" == "webmail" ]] || [[ ""${CONTAINER}"" == "roundcube" ]] || [[ ""${CONTAINER}"" == "webmail-billionmail" ]]; then
        CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "roundcubemail:" | awk '{print $1}' )
    
    else
        echo "Please use: core|postfix|dovecot|rspamd|pgsql|redis|webmail"
        Red_Error "ERROR: The ""${CONTAINER}"" container does not exist"
    fi

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
    # if [[ ! "${BILLIONMAIL_HOSTNAME}" =~ ^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9](\.[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9])*$ ]]; then
    if [[ ! "${BILLIONMAIL_HOSTNAME}" =~ ^([a-zA-Z0-9]([a-zA-Z0-9-]{0,62}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$ ]]; then
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

Domain_DKIM_record(){
    if [ -z "${BILLIONMAIL_HOSTNAME}" ]; then
        echo "Please enter the domain name"
        exit 1
    fi

    Check_domain=$(docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "SELECT * FROM domain WHERE domain = '${BILLIONMAIL_HOSTNAME}';" | grep -w "^ ${BILLIONMAIL_HOSTNAME}")
    if [ -z "${Check_domain}" ]; then
        echo "Error: ${BILLIONMAIL_HOSTNAME} Domain does not exist."
        exit 1
    fi

    ## DKIM key generation
    docker exec -i -e BILLIONMAIL_HOSTNAME=${BILLIONMAIL_HOSTNAME} ${RSPAMD_CONTAINER_NAME} bash -c 'cat << "EOF" > /tmp/1.sh
#!/bin/bash
if [ ! -d "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/" ]; then
    mkdir -p "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/"
fi
if [ -f "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/default.private" ] && [ -f "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/default.pub" ]; then
    #echo "DKIM key already exists, skipping generation."
    exit 0
fi

rspamadm dkim_keygen -s 'default' -b 1024 -d {domain} -k "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/default.private" > "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/default.pub"
if [ $? -eq 0 ]; then
    # Define the root directory for private keys
    DKIM_KEYS_DIR="/var/lib/rspamd/dkim"
    # Configuration file path
    CONFIG_FILE="/etc/rspamd/local.d/dkim_signing.conf"
    # Check if the configuration file exists, if not, create it
    if [ ! -f "${CONFIG_FILE}" ]; then
        touch "${CONFIG_FILE}"
        echo "${CONFIG_FILE}"
        echo "domain {" > "${CONFIG_FILE}"
        echo "#BT_DOMAIN_DKIM_BEGIN" >> "${CONFIG_FILE}"
        echo "#BT_DOMAIN_DKIM_END" >> "${CONFIG_FILE}"
        echo "}" >> "${CONFIG_FILE}"
    fi
    # Scan the DKIM_KEYS_DIR directory to get all domains
    DOMAINS=()
    for DOMAIN_DIR in "${DKIM_KEYS_DIR}"/*; do
        if [ -d "${DOMAIN_DIR}" ]; then
            DOMAIN_NAME=$(basename "${DOMAIN_DIR}")
            PRIVATE_KEY_PATH="${DOMAIN_DIR}/default.private"
            if [ -f "${PRIVATE_KEY_PATH}" ]; then
                DOMAINS+=("${DOMAIN_NAME}:${PRIVATE_KEY_PATH}")
            else
                echo "Warning: Private key file ${PRIVATE_KEY_PATH} for domain ${DOMAIN_NAME} does not exist, skipping configuration."
            fi
        fi
    done
    # Add domain configurations
    for DOMAIN in "${DOMAINS[@]}"; do
        DOMAIN_NAME=$(echo "${DOMAIN}" | cut -d':' -f1)
        PRIVATE_KEY_PATH=$(echo "${DOMAIN}" | cut -d':' -f2)
        # Check if the domain already exists
        if grep -wq "${DOMAIN_NAME}" "${CONFIG_FILE}"; then
            # echo "Domain ${DOMAIN_NAME} already exists, skipping configuration."
            continue
        fi
        # Insert domain configuration between #BT_DOMAIN_DKIM_BEGIN and #BT_DOMAIN_DKIM_END
        sed -i "/^#BT_DOMAIN_DKIM_BEGIN$/a #${DOMAIN_NAME}_DKIM_BEGIN\n  ${DOMAIN_NAME} {\n    selectors [\n     {\n       path: \"${PRIVATE_KEY_PATH}\";\n       selector: \"default\"\n     }\n   ]\n }\n#${DOMAIN_NAME}_DKIM_END" "${CONFIG_FILE}"
        # echo "Domain ${DOMAIN_NAME} has been successfully added to the configuration file."
    done
    # echo "DKIM configuration update completed."
else
    echo -e "DKIM key generation failed!"
    exit 1
fi
chmod 755 -R "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/"
EOF'
    docker exec -i -e BILLIONMAIL_HOSTNAME=${BILLIONMAIL_HOSTNAME} ${RSPAMD_CONTAINER_NAME} bash /tmp/1.sh && rm -f /tmp/1.sh
    DKIM_RECORD=$(docker exec ${RSPAMD_CONTAINER_NAME} cat "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/default.pub")
    # echo "DKIM RECORD: ${DKIM_RECORD}"
    
}


Domain_record() {

    Check_domain=$(docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "SELECT * FROM domain WHERE domain = '${BILLIONMAIL_HOSTNAME}';" | grep -w "^ ${BILLIONMAIL_HOSTNAME}")
    if [ "${Check_domain}" ]; then

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
        echo -e ""
        echo -e "\e[31mPlease add the following record to your domain name\e[0m"
        echo -e "==========================================================="
        echo -e " Type | Host record    |    IPv4 address   |"
        echo -e "  \e[1;33mA\e[0m   | \e[1;33mmail.${BILLIONMAIL_HOSTNAME}\e[0m | \e[1;33m${IPV4_ADDRESS}\e[0m |"
        echo -e "==========================================================="
        echo -e " Type | Host record | MX priority |  Record value    "
        echo -e "  \e[1;33mMX\e[0m  |     \e[1;33m@\e[0m       |      \e[1;33m10\e[0m     | \e[1;33mmail.${BILLIONMAIL_HOSTNAME}\e[0m "
        echo -e "==========================================================="
        if [ "${IPV4_ADDRESS}" ]; then
            echo -e " Type | Host record |    Record value   |"
            echo -e "  \e[1;33mTXT\e[0m |     \e[1;33m@\e[0m       | \e[1;33mv=spf1 +a +mx +ip4:${IPV4_ADDRESS} -all\e[0m |"
        else
            echo -e " Type | Host record |    Record value   |"
            echo -e "  \e[1;33mTXT\e[0m |     \e[1;33m@\e[0m       | \e[1;33mv=spf1 +a +mx -all\e[0m |"
        fi
        echo -e "==========================================================="
        echo -e " Type | Host record |    Record value     |"
        echo -e "  \e[1;33mTXT\e[0m |   \e[1;33m_dmarc\e[0m    | \e[1;33mv=DMARC1;p=quarantine;rua=mailto:admin@${BILLIONMAIL_HOSTNAME}\e[0m |"
        echo -e "==========================================================="

        Domain_DKIM_record

        if [ "${DKIM_RECORD}" ]; then
            DKIM_RECORD=$(echo "${DKIM_RECORD}" | awk -F'"' '{print $2 $4}' | tr -d '[:space:]')
            #echo "DKIM RECORD: ${DKIM_RECORD}"
            echo -e " Type |  Host record       |    Record value     |"
            echo -e "  \e[1;33mTXT\e[0m | \e[1;33mdefault._domainkey\e[0m | \e[1;33m${DKIM_RECORD}\e[0m |<-- Start from \"v=DKIM1\" end, A single line."
            echo -e "==========================================================="
        else
            echo -e "${BILLIONMAIL_HOSTNAME} DKIM key generation failed!"
        fi
    else
        echo -e "Domain ${BILLIONMAIL_HOSTNAME} does not exist."
    fi
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
    if [ -z "${Check_domain}" ]; then
        # Create a domain
        
        docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "INSERT INTO domain (domain, a_record, mailboxes, mailbox_quota, quota, rate_limit, create_time, active)
        VALUES ('${BILLIONMAIL_HOSTNAME}', 'mail.${BILLIONMAIL_HOSTNAME}', 500, 5368709120, 5368709120, 12, ${create_time}, 1);"
        if [ $? -eq 0 ]; then
            echo "${BILLIONMAIL_HOSTNAME} Domain creation was successful!"
            Domain_record
        else
            Red_Error "Domain creation failed!"
        fi
    else
        echo ""${Check_domain}" Domain already exists!"
    fi
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

Update_BillionMail() {
    BRANCH="main"
    if [ -f "update.sh" ]; then
        echo -e "Checking for update.sh script..."
        MD5_1="$(md5sum update.sh)"
        git fetch origin
        git checkout "origin/${BRANCH}" update.sh
        MD5_2=$(md5sum update.sh)
        if [[ "${MD5_1}" != "${MD5_2}" ]]; then
            #echo -e "\033[33m update.sh is changed, please run update.sh script again.\033[0m"
            chmod +x update.sh
            echo y | ./update.sh
            #exit 1
        else
            chmod +x update.sh
            echo y | ./update.sh
        fi
    else
        Red_Error "Error: update.sh script does not exist!"
    fi
}


Default_info() {

    ipv4_address=""
    ipv6_address=""
    if [ "$address" = "" ];then
            
        ipv4_address=$(curl -4 -sS --connect-timeout 10 -m 15 https://ifconfig.me 2>&1)
        if [ -z "${ipv4_address}" ];then
                ipv4_address=$(curl -4 -sS --connect-timeout 10 -m 15 https://www.aapanel.com/api/common/getClientIP 2>&1)
                if [ -z "${ipv4_address}" ];then
                    ipv4_address=$(curl -4 -sS --connect-timeout 10 -m 15 https://www.bt.cn/Api/getIpAddress 2>&1)
                fi
        fi
        IPV4_REGEX="^([0-9]{1,3}\.){3}[0-9]{1,3}$"
        if ! [[ $ipv4_address =~ $IPV4_REGEX ]]; then
                ipv4_address=""
        fi
        
        ipv6_address=$(curl -6 -sS --connect-timeout 10 -m 15 https://ifconfig.me 2>&1)
        IPV6_REGEX="^([0-9a-fA-F]{0,4}:){1,7}[0-9a-fA-F]{0,4}$"
        if ! [[ $ipv6_address =~ $IPV6_REGEX ]]; then
                ipv6_address=""
        else
            if [[ ! $ipv6_address =~ ^\[ ]]; then
                ipv6_address="[$ipv6_address]"
            fi
        fi

        if [ "$address" = "" ] && [ "$ipv4_address" = "" ] && [ "$ipv6_address" = "" ];then
            address="SERVER_IP"
            echo -e "\033[33mFailed to obtain Internet IP, please use the server Internet IP+PORT to access.\033[0m"
        fi
    fi

    LOCAL_IP=$(ip addr | grep -E -o '[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | grep -E -v "^127\.|^255\.|^0\." | head -n 1)
    echo -e "=================================================================="
    echo -e "\033[32mBillionMail default info!\033[0m"
    echo -e "=================================================================="
    pool=https
    if [ "${ipv6_address}" ];then
        if [ "${HTTPS_PORT}" = "443" ];then
            echo  "BillionMail Internet IPv6 Address: ${pool}://${ipv6_address}/${SafePath}"
        else
            echo  "BillionMail Internet IPv6 Address: ${pool}://${ipv6_address}:${HTTPS_PORT}/${SafePath}"
        fi
    fi
    if [ "${ipv4_address}" ];then
        if [ "${HTTPS_PORT}" = "443" ];then
            echo  "BillionMail Internet IPv4 Address: ${pool}://${ipv4_address}/${SafePath}"
        else
            echo  "BillionMail Internet IPv4 Address: ${pool}://${ipv4_address}:${HTTPS_PORT}/${SafePath}"
        fi
    fi
    if [ "${address}" ];then
        if [ "${HTTPS_PORT}" = "443" ];then
            echo  "BillionMail Internet Address: ${pool}://${address}/${SafePath}"
        else
            echo  "BillionMail Internet Address: ${pool}://${address}:${HTTPS_PORT}/${SafePath}"
        fi

    fi

    if [ "${HTTPS_PORT}" = "443" ];then
        echo  "BillionMail Internal Address:      ${pool}://${LOCAL_IP}/${SafePath}"
    else
        echo  "BillionMail Internal Address:      ${pool}://${LOCAL_IP}:${HTTPS_PORT}/${SafePath}"
    fi
    
    echo -e "Username: ${ADMIN_USERNAME} \nPassword: ${ADMIN_PASSWORD}"
    echo -e "\033[33mWarning:\033[0m"
    echo -e "\033[33mIf you cannot access the BillionMail, \033[0m"
    echo -e "\033[33mrelease the following port ${SMTP_PORT}|${SMTPS_PORT}|${SUBMISSION_PORT}|${POP_PORT}|${IMAP_PORT}|${IMAPS_PORT}|${POPS_PORT}|${HTTP_PORT}|${HTTPS_PORT} in the security group\033[0m"
    echo -e "=================================================================="
}

# Modify the apply SSL interface access port
MODIFY_HTTP_SSL_PORT() { 

    NEW_PORT="$2"
    if [ -z "${NEW_PORT}" ]; then
        echo -e "\033[31m Let's Encrypt SSL certificate can only be applied using port 80, other ports cannot be applied.\033[0m"
        read -p "Please enter the new apply SSL port: " NEW_PORT
    fi

    # Verify that the input is a number
    if ! echo "${NEW_PORT}" | grep -qE '^[0-9]+$' || ((NEW_PORT < 1 || NEW_PORT > 65535)); then
        echo -e "\033[31m Error: The port number must be a number between 1-65535 \033[0m"
        exit 1
    fi

    Check_Port=$(ss -tlnp | grep -E ":(${NEW_PORT})\b")
    if [ ! -z "${Check_Port}" ]; then
        echo "${Check_Port}"
        echo -e "\033[31m Error: The ${NEW_PORT} port is already in use. \033[0m"
        exit 1
    fi

    # Perform modification
    sed -i 's/^HTTP_PORT=.*/HTTP_PORT='"${NEW_PORT}"'/' .env
    echo -e "The BillionMail apply SSL port has been modified to: ${NEW_PORT} \n Rebuild the container, please wait..."

    sleep 3
    # Find the container ID of the core image in the current project and rebuild the container
    # CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "/core:" | awk '{print $1}' )
    CONTAINER="core"
    GET_CONTAINER_ID ${CONTAINER}
    if [ "${CONTAINER_ID}" ]; then
        echo "Rebuilding Manage Container..."
        docker stop ${CONTAINER_ID}
        docker rm -f ${CONTAINER_ID}
        ${DOCKER_COMPOSE} up -d
    else
        echo "The "core" container does not exist"
        echo "Starting BillionMail..."
        ${DOCKER_COMPOSE} up -d
    fi

    if [ -f "/usr/sbin/ufw" ]; then
        /usr/sbin/ufw allow ${NEW_PORT}/tcp >/dev/null 2>&1
        /usr/sbin/ufw reload
    elif [ -f "/usr/sbin/firewall-cmd" ]; then
        /usr/sbin/firewall-cmd --permanent --zone=public --add-port=${NEW_PORT}/tcp >/dev/null 2>&1
        /usr/sbin/firewall-cmd --reload
    elif [ -f "/etc/sysconfig/iptables" ]; then
        iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport ${NEW_PORT} -j ACCEPT
        service iptables save
    fi

    if [[ ${NEW_PORT} != "80" ]]; then 
        echo -e "\033[31m Let's Encrypt SSL certificate can only be applied using port 80, other ports cannot be applied.\033[0m"
    fi
}

# Modify the management interface access port
MODIFY_HTTPS_PORT() { 

    NEW_PORT="$2"
    if [ -z "${NEW_PORT}" ]; then
        read -p "Please enter the new BillionMail management port: " NEW_PORT
    fi

    # Verify that the input is a number
    if ! echo "${NEW_PORT}" | grep -qE '^[0-9]+$' || ((NEW_PORT < 1 || NEW_PORT > 65535)); then
        echo -e "\033[31m Error: The port number must be a number between 1-65535 \033[0m"
        exit 1
    fi

    Check_Port=$(ss -tlnp | grep -E ":(${NEW_PORT})\b")
    if [ ! -z "${Check_Port}" ]; then
        echo "${Check_Port}"
        echo -e "\033[31m Error: The ${NEW_PORT} port is already in use. \033[0m"
        exit 1
    fi

    # Perform modification
    sed -i 's/^HTTPS_PORT=.*/HTTPS_PORT='"${NEW_PORT}"'/' .env
    if [ $? -ne 0 ]; then
        echo -e "\033[31m Error: The BillionMail management port modification failed! \033[0m"
        exit 1
    fi
    
    echo -e "The BillionMail management port has been modified to: ${NEW_PORT} \n Rebuild the container, please wait..."
    sleep 3
    # Find the container ID of the core image in the current project and rebuild the container
    # CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "/core:" | awk '{print $1}' )
    CONTAINER="core"
    GET_CONTAINER_ID ${CONTAINER}
    if [ "${CONTAINER_ID}" ]; then
        echo "Rebuilding Manage Container..."
        docker stop ${CONTAINER_ID}
        docker rm -f ${CONTAINER_ID}
        ${DOCKER_COMPOSE} up -d
    else
        echo "The "core" container does not exist"
        echo "Starting BillionMail..."
        ${DOCKER_COMPOSE} up -d
    fi

    if [ -f "/usr/sbin/ufw" ]; then
        /usr/sbin/ufw allow ${NEW_PORT}/tcp >/dev/null 2>&1
        /usr/sbin/ufw reload
    elif [ -f "/usr/sbin/firewall-cmd" ]; then
        /usr/sbin/firewall-cmd --permanent --zone=public --add-port=${NEW_PORT}/tcp >/dev/null 2>&1
        /usr/sbin/firewall-cmd --reload
    elif [ -f "/etc/sysconfig/iptables" ]; then
        iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport ${NEW_PORT} -j ACCEPT
        service iptables save
    fi

    # Display login information
    bash bm.sh default
}

# Modify time zone
MODIFY_TZ() { 
    if [ -a /etc/timezone ]; then
        SYSTEM_TIME_ZONE=$(cat /etc/timezone)
    elif [ -a /etc/localtime ]; then
        SYSTEM_TIME_ZONE=$(readlink /etc/localtime|sed -n 's|^.*zoneinfo/||p')
    fi
    NEW_TZ="$2"
    echo -e ""
    echo -e "See https://en.wikipedia.org/wiki/List_of_tz_database_time_zones for a list of timezones"
    echo -e "Use a column named "TZ identifier" + note the column named "Notes""
    echo -e "Please enter your time zone"
    echo -e ""
    if [ -z "${SYSTEM_TIME_ZONE}" ]; then
        read -p "Timezone: " -e NEW_TZ
    else
        read -p "Timezone [${SYSTEM_TIME_ZONE}]: " -e NEW_TZ
        [ -z "${NEW_TZ}" ] && NEW_TZ=${SYSTEM_TIME_ZONE}
    fi

    if [ -z "${NEW_TZ}" ]; then
        echo -e "\033[31m Error: Please enter the time zone \033[0m"
        exit 1
    fi

    # Perform modification
    sed -i "s|^TZ=.*|TZ=${NEW_TZ}|" .env
    if [ $? -ne 0 ]; then
        echo -e "\033[31m Error: The BillionMail time zone modification failed! \033[0m"
        exit 1
    fi    
    
    echo -e "The BillionMail time zone has been modified to: ${NEW_TZ} \n Rebuild the container, please wait..."
    sleep 3
    ${DOCKER_COMPOSE} down
    ${DOCKER_COMPOSE} up -d

}


# Modify Admin Username
MODIFY_ADMIN_USERNAME() { 

    NEW_ADMIN="$2"
    if [ -z "${NEW_ADMIN}" ]; then
        read -p "Please enter the new BillionMail administrator username (minimum 5 characters): " NEW_ADMIN
    fi
    if [ -z "${NEW_ADMIN}" ]; then
        echo -e "\033[31mError: Administrator username is required!\033[0m"
        exit 1
    fi
    if [ ${#NEW_ADMIN} -lt 5 ]; then
        echo -e "\033[31mError: Username must be at least 5 characters long!\033[0m"
        exit 1
    fi
    if ! [[ "${NEW_ADMIN}" =~ ^[a-zA-Z0-9@#_.!+-]+$ ]]; then
        echo -e "\033[31mError: Username can only contain letters, numbers, symbols: @ # _ - + . !\033[0m"
        exit 1
    fi
    # Perform modification
    sed -i "s|^ADMIN_USERNAME=.*|ADMIN_USERNAME="${NEW_ADMIN}"|" .env
    if [ $? -ne 0 ]; then
        echo -e "\033[31mError: Failed to update BillionMail administrator username!\033[0m"
        exit 1
    fi
    echo -e "BillionMail administrator username has been updated. Restarting container, please wait..."
    sleep 3
    # Find the container ID of the core image in the current project and rebuild the container
    # CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "/core:" | awk '{print $1}' )
    CONTAINER="core"
    GET_CONTAINER_ID ${CONTAINER}
    if [ "${CONTAINER_ID}" ]; then
        echo "Restarting Manage Container..."
        docker restart ${CONTAINER_ID}
    else
        echo "The "core" container does not exist"
        echo "Starting BillionMail..."
        ${DOCKER_COMPOSE} up -d
    fi
    # Display login information
    bash bm.sh default
}


# Modify Admin Password
MODIFY_ADMIN_PASSWORD() { 

    NEW_PASSWORD="$2"
    if [ -z "${NEW_PASSWORD}" ]; then
        read -p "Please enter the new BillionMail administrator password (minimum 5 characters): " NEW_PASSWORD
    fi
    if [ -z "${NEW_PASSWORD}" ]; then
        echo -e "\033[31mError: Administrator password is required!\033[0m"
        exit 1
    fi
    if [ ${#NEW_PASSWORD} -lt 5 ]; then
        echo -e "\033[31mError: Password must be at least 5 characters long!\033[0m"
        exit 1
    fi
    if ! [[ "${NEW_PASSWORD}" =~ ^[a-zA-Z0-9@#_.!+-]+$ ]]; then
        echo -e "\033[31mError: Password can only contain letters, numbers, symbols: @ # _ - + . !\033[0m"
        exit 1
    fi
    # Perform modification
    sed -i "s|^ADMIN_PASSWORD=.*|ADMIN_PASSWORD="${NEW_PASSWORD}"|" .env
    if [ $? -ne 0 ]; then
        echo -e "\033[31mError: Failed to update BillionMail administrator password!\033[0m"
        exit 1
    fi
    echo -e "BillionMail administrator password has been updated. Restarting container, please wait..."
    sleep 3
    # Find the container ID of the core image in the current project and rebuild the container
    #CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "/core:" | awk '{print $1}' )
    CONTAINER="core"
    GET_CONTAINER_ID ${CONTAINER}
    if [ "${CONTAINER_ID}" ]; then
        echo "Restarting Manage Container..."
        docker restart ${CONTAINER_ID}
    else
        echo "The "core" container does not exist"
        echo "Starting BillionMail..."
        ${DOCKER_COMPOSE} up -d
    fi
    # Display login information
    bash bm.sh default

}


# Modify Safe entrance
MODIFY_SAFE_ENTRANCE() { 

    NEW_ENTRANCE="$2"
    if [ -z "${NEW_ENTRANCE}" ]; then
        read -p "Please enter the new BillionMail security entrance path (minimum 5 characters): " NEW_ENTRANCE
    fi
    if [ -z "${NEW_ENTRANCE}" ]; then
        echo -e "\033[31mError: Security entrance path is required!\033[0m"
        exit 1
    fi
    if [ ${#NEW_ENTRANCE} -lt 5 ]; then
        echo -e "\033[31mError: Security entrance path must be at least 5 characters long!\033[0m"
        exit 1
    fi
    if ! [[ "${NEW_ENTRANCE}" =~ ^[a-zA-Z0-9]+$ ]]; then
        echo -e "\033[31mError: Security entrance path can only contain letters and numbers!\033[0m"
        exit 1
    fi
    # Perform modification
    sed -i "s|^SafePath=.*|SafePath=${NEW_ENTRANCE}|" .env
    if [ $? -ne 0 ]; then
        echo -e "\033[31mError: Failed to update BillionMail security entrance path!\033[0m"
        exit 1
    fi
    echo -e "BillionMail security entrance path has been updated. Restarting container, please wait..."
    sleep 3
    # Find the container ID of the core image in the current project and rebuild the container
    # CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "/core:" | awk '{print $1}' )
    CONTAINER="core"
    GET_CONTAINER_ID ${CONTAINER}
    if [ "${CONTAINER_ID}" ]; then
        echo "Restarting Manage Container..."
        docker restart ${CONTAINER_ID}
    else
        echo "The "core" container does not exist"
        echo "Starting BillionMail..."
        ${DOCKER_COMPOSE} up -d
    fi
    # Display login information
    bash bm.sh default
}

# Rebuild the BillionMail project
 REBUILD_PROJECT() {
    echo "Rebuilding BillionMail..."
    echo -e "\033[31mWarning: This operation will rebuild all containers, service will be unavailable during the rebuild process. \033[0m"
    read -p "Are you sure you want to continue? (yes/no): " NEW_REBUILD
    if [[ "$NEW_REBUILD" =~ ^(yes|y|Y)$ ]]; then
        sleep 3
        ${DOCKER_COMPOSE} down
        ${DOCKER_COMPOSE} up -d
        echo "Rebuild completed."
    else
        echo "Rebuild cancel."
    fi
 }

# Restart the BillionMail project
RESTART_PROJECT() {
    echo "Restarting BillionMail..."
    sleep 1
    ${DOCKER_COMPOSE} restart
}

# Stop the BillionMail project
STOP_PROJECT() {
    echo "Stop BillionMail..."
    sleep 1
    ${DOCKER_COMPOSE} stop
}

START_PROJECT() {
    echo "Start BillionMail..."
    sleep 1
    ${DOCKER_COMPOSE} up -d
}

DOWN_PROJECT() {
    echo "Stop all BillionMail services and delete all BillionMail containers..."
    sleep 3
    ${DOCKER_COMPOSE} down
}

# Restart the specified service
RESTART_SERVICE() { 
    SERVICE="$2"
    if [ -z "${SERVICE}" ]; then
        echo ""
        echo "Support input: postfix|core|dovecot|rspamd|redis|webmail|pgsql"
        read -p "Please enter service: " SERVICE
    fi

    GET_SERVICE_NAME ${SERVICE}
    echo "Restarting ${SERVICE} service..."
    sleep 3
    ${DOCKER_COMPOSE} restart ${SERVICE_NAME}
}

RESTART_SERVICE_ADMIN() { 
    SERVICE="core"
    GET_SERVICE_NAME ${SERVICE}
    echo "Restarting ${SERVICE} service..."
    sleep 3
    ${DOCKER_COMPOSE} restart ${SERVICE_NAME}
}

# Get container log 
GET_CONTAINER_LOG() { 
    CONTAINER="$2"
    if [ -z "${CONTAINER}" ]; then
        echo ""
        echo "Support input: postfix|core|dovecot|rspamd|redis|webmail|pgsql"
        read -p "Please enter container: " CONTAINER
    fi

    log_num="$3"
    follow=""

    if [ "$log_num" == "-f" ]; then
        follow="-f"
        log_num="$4"
    elif [ "$4" == "-f" ]; then
        follow="-f"
    fi

    if [[ "$log_num" =~ ^[0-9]+$ ]]; then
        log_num="${log_num}"
    else
        log_num=25
    fi

    GET_CONTAINER_ID ${CONTAINER}
    if [ -z "${CONTAINER_ID}" ]; then
        Red_Error "ERROR: The "${CONTAINER}" container does not exist, Please execute '${DOCKER_COMPOSE} up -d' to start BillionMail"
    fi
    docker logs ${follow} --tail="${log_num}" ${CONTAINER_ID}
}

# Get file log
GET_FILE_LOG() {
    NAME_FILE="$2"
    
    if [ -z "${NAME_FILE}" ]; then
        echo ""
        echo "Support input: postfix|core|core-access|core-error|dovecot|rspamd|fail2ban"
        read -p "Please enter name: " NAME_FILE
    fi

    log_num="$3"
    follow=""

    if [ "$log_num" == "-f" ]; then
        follow="-f"
        log_num="$4"
    elif [ "$4" == "-f" ]; then
        follow="-f"
    fi

    if [[ "$log_num" =~ ^[0-9]+$ ]]; then
        log_num="${log_num}"
    else
        log_num=25
    fi

    # Define log files
    declare -A LOG_FILES=(
        ["core"]="$(date +%Y-%m-%d).log"
        ["core-access"]="access-$(date +%Y%m%d).log"
        ["core-error"]="error-$(date +%Y%m%d).log"
        ["postfix"]="mail.log"
        ["dovecot"]="mail.log"
        ["rspamd"]="rspamd.log"
        ["fail2ban"]="fail2ban.log"
    )

    # Define log directory mapping 
    declare -A LOG_DIRS=(
        ["core"]="core"
        ["core-access"]="core"
        ["core-error"]="core"
        ["postfix"]="postfix"
        ["dovecot"]="dovecot"
        ["rspamd"]="rspamd"
        ["fail2ban"]="fail2ban"
    )

    # Check whether the service name is valid
    if [[ ! " ${!LOG_FILES[@]} " =~ " ${NAME_FILE} " ]]; then
        echo "How to use: $0 log_file core 100"
        Red_Error "Error: Invalid name. Please enter name: (postfix|core|core-access|core-error|dovecot|rspamd|fail2ban) "
    fi

    # Get log file name and directory
    LOG_FILE="${LOG_FILES[$NAME_FILE]}"
    LOG_DIR="${LOG_DIRS[$NAME_FILE]}"
    
    # Check and display log files
    if [ -f "logs/${LOG_DIR}/${LOG_FILE}" ]; then
        tail ${follow} -n ${log_num} "logs/${LOG_DIR}/${LOG_FILE}"
    elif [ -f "data/logs/${LOG_DIR}/${LOG_FILE}" ]; then
        tail ${follow} -n ${log_num} "data/logs/${LOG_DIR}/${LOG_FILE}"
    else
        Red_Error "logs file does not exist!"
    fi
}

# Get service TOP 
GET_SERVICE_TOP() {

    SERVICE="$2"
    if [ -z "${SERVICE}" ]; then
        read -p "Please enter service: " SERVICE
    fi

    GET_SERVICE_NAME ${SERVICE}
    ${DOCKER_COMPOSE} top ${SERVICE_NAME}
}

GET_SERVICE_TOP_ALL() {

    ${DOCKER_COMPOSE} top

}

GET_SERVICE_PS() {

    ${DOCKER_COMPOSE} ps

}

# Get containers status
GET_CONTAINER_STATUS() {
    CONTAINERS=("core" "postfix" "dovecot" "rspamd" "pgsql" "redis" "webmail")
        
    for cc in "${CONTAINERS[@]}"; do
        
        CONTAINER="${cc}"
        GET_CONTAINER_ID "${CONTAINER}"
                
        status=$(docker inspect --format='{{.State.Status}}' "${CONTAINER_ID}" 2>/dev/null)
        if [ "$status" = "running" ]; then
            echo -e "\033[32m ${CONTAINER} container is running\033[0m"
        else
            echo -e "\033[31m ${CONTAINER} container is not running. Status: ${status} \033[0m"
        fi
        
    done
}

CLEAR_OLD_IMAGE() {
    # Remove only outdated versions of images defined in docker-compose.yml

    # Check if docker-compose.yml exists
    if [ ! -f "docker-compose.yml" ]; then
        Red_Error "Error: docker-compose.yml file not found"      
    fi

    # Extract complete image names (including tags) defined in compose file
    # echo "Extracting billionmail images from docker-compose.yml..."
    # COMPOSE_IMAGES=$(grep -oP 'image:\s*\K[^"\s]+' docker-compose.yml | sort -u)
    COMPOSE_IMAGES=$(grep -oP 'image:\s*\K(?:billionmail|ghcr\.io/aapanel)[^"\s]+' docker-compose.yml | sort -u)

    # Check if any images were found
    if [ -z "${COMPOSE_IMAGES}" ]; then
        echo "Warning: No image definitions found in docker-compose.yml"
        exit 1
    fi

    # Get all local Docker images (excluding <none> images)
    # echo "Getting all local Docker images..."
    ALL_IMAGES=$(docker images --format "{{.Repository}}:{{.Tag}}" | grep -v "<none>:<none>")

    # Find outdated images to remove
    IMAGES_TO_REMOVE=""

    while IFS= read -r compose_image; do
        # Extract image name (without tag)
        image_name=$(echo "${compose_image}" | awk -F: '{print $1}')
        
        # Extract tag defined in compose file
        compose_tag=$(echo "${compose_image}" | awk -F: '{print $2}')
        
        # Find all local images with the same name
        local_images=$(echo "${ALL_IMAGES}" | grep "^${image_name}:")
        
        # Filter images with different tags
        for local_image in ${local_images}; do
            local_tag=$(echo "${local_image}" | awk -F: '{print $2}')
            
            # If local tag differs from compose tag, add to removal list
            if [ "${local_tag}" != "${compose_tag}" ]; then
                IMAGES_TO_REMOVE="${IMAGES_TO_REMOVE} ${local_image}"
            fi
        done
    done <<< "${COMPOSE_IMAGES}"

    # Display results
    if [ -z "${IMAGES_TO_REMOVE}" ]; then
        echo ""
        echo "No outdated images found to remove"
        exit 0
    fi

    echo -e "\n--- Outdated images to remove ---"
    echo "${IMAGES_TO_REMOVE}" | tr ' ' '\n' | grep -v '^$'
    IMAGE_COUNT=$(echo "${IMAGES_TO_REMOVE}" | wc -w)
    echo -e "\nTotal: ${IMAGE_COUNT} outdated images to remove"

    # Ask for confirmation
    read -p "Remove these outdated images? (y/n): " confirm
    if [ "${confirm}" = "y" ] || [ "${confirm}" = "Y" ]; then
        sleep 5
        for image in ${IMAGES_TO_REMOVE}; do
            echo "Removing image: ${image}"
            docker rmi "${image}" || echo "Warning: Failed to remove ${image}"
        done
        echo "Outdated images removed successfully"
        echo -e "For deeper Docker system cleanup, ensure all containers to keep are running before executing:\n docker system prune"
    else
        echo "Operation cancelled"
    fi
}


SHOW_HELP() {
        echo "Help Information:"
        echo "  default                   - Show BillionMail login default info: $0 default"
        echo "  update                    - Update BillionMail: $0 update"
        echo "  change-port               - Modify BillionMail access management port: $0 change-port"
        echo "  change-tz                 - Modify BillionMail time zone: $0 change-tz"
        echo "  change-user               - Modify BillionMail Administrator user: $0 change-user"
        echo "  change-password           - Modify BillionMail Administrator password: $0 change-password"
        echo "  change-safe-path          - Modify BillionMail security entrance path: $0 change-safe-path"
        echo "  change-apply-ssl-port     - Modify BillionMail apply ssl port: $0 change-apply-ssl-port"
        echo "  start                     - Start BillionMail: $0 start"
        echo "  stop                      - Stop BillionMail: $0 stop"
        echo "  restart                   - Restart BillionMail: $0 restart"
        echo "  status                    - Show BillionMail containers running status : $0 status"
        echo "  down                      - Stop and remove containers, networks: $0 down"
        echo "  rebuild                   - Rebuild all BillionMail containers: $0 rebuild"
        echo "  top                       - Show all BillionMail processes: $0 top"
        echo "  ps                        - Show all BillionMail containers: $0 ps"
        echo "  service-top               - Show processes of a specific BillionMail service: $0 s-t postfix"
        echo "  log-file <service>            - View logs of a specific service: $0 l-f postfix"
        echo "  log-container <container>     - View logs of a specific container: $0 l-c postfix"
        echo "  restart-service <service>     - Restart a specific service and its container: $0 r-s postfix"
        echo "  clear                     - Clear BillionMail old images: $0 clear"
        # echo "  add-domain <domain>       - Add domain. Example: $0 add-domain example.com"
        # echo "  del-domain <domain>       - Delete domain. Example: $0 del-domain example.com"
        # echo "  add-email <email>         - Add email. Example: $0 add-email user@example.com"
        # echo "  del-email <email>         - Delete email. Example: $0 del-email user@example.com"
        echo "  show-domain               - Show domain data."
        echo "  show-email                - Show email data."
        # echo "  show-record               - Show domain DNS record."
        
        exit
}


case "$1" in
    h | -h | help | --help | -help)
        SHOW_HELP
        ;;
    default | info)
        Default_info
        ;;
    update)
        Update_BillionMail
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
    show-record|show_dns_record)
        Init_Domain "$@"
        Domain_record
        ;;
    MODIFY_HTTP_PORT|modify_http_port|change-apply-ssl-port)
        MODIFY_HTTP_SSL_PORT
        ;;
    MODIFY_HTTPS_PORT|modify_https_port|change-port)
        MODIFY_HTTPS_PORT
        ;;
    MODIFY_TZ|modify_tz|change-tz)
        MODIFY_TZ
        ;;
    MODIFY_ADMIN_USERNAME|modify_user|change-user)
        MODIFY_ADMIN_USERNAME
        ;;
    MODIFY_ADMIN_PASSWORD|modify_password|change-password)
        MODIFY_ADMIN_PASSWORD
        ;;
    MODIFY_SAFE_ENTRANCE|modify_safe_path|change-safe-path)
        MODIFY_SAFE_ENTRANCE
        ;;
    REBUILD_PROJECT|rebuild_project|rebuild)
        REBUILD_PROJECT
        ;;
    RESTART_PROJECT|restart_project|restart)
        RESTART_PROJECT
        ;;
    START_PROJECT|start_project|start)
        START_PROJECT
        ;;
    STOP_PROJECT|stop_project|stop)
        STOP_PROJECT
        ;;
    GET_CONTAINER_STATUS|get_container_status|status)
        GET_CONTAINER_STATUS
        ;;
    DOWN_PROJECT|down_project|down)
        DOWN_PROJECT
        ;;
    RESTART_SERVICE|restart_service|restart-service|r-s)
        RESTART_SERVICE "$@"
        ;;
    GET_CONTAINER_LOG|log_container|log-container|l-c)
        GET_CONTAINER_LOG "$@"
        ;;
    GET_FILE_LOG|log_file|log-file|l-f)
        GET_FILE_LOG "$@"
        ;;
    GET_SERVICE_TOP|service_top|service-top|s-t)
        GET_SERVICE_TOP "$@"
        ;;
    GET_SERVICE_TOP_ALL|service_top_all|service-top-all|top)
        GET_SERVICE_TOP_ALL
        ;;    
    GET_SERVICE_PS|service-ps|ps)
        GET_SERVICE_PS
        ;;  
    RESTART_SERVICE_ADMIN|restart_service_admin|restart-service-admin|restart-admin)
        RESTART_SERVICE_ADMIN
        ;;
    clear)
    CLEAR_OLD_IMAGE
    ;;    

    *)
        echo "=============== BillionMail CLI =================="
        echo "1) Restart BillionMail          2) View login info"
        echo ""
        echo "3) View running status          4) Stop BillionMail"
        echo ""
        echo "5) Start BillionMail            6) Restart manage only (mail unaffected)"
        echo ""
        echo "7) Change manage password       8) Change manage username"
        echo ""
        echo "9) Change secure entry          10) Change manage access port"
        echo ""
        echo "11) View all processes          12) Update BillionMail"
        echo ""
        echo "0) Exit      For more commands: help"
        echo "=================================================="
        read -p "Enter number to execute command [0-12]: " input  
        case ${input} in
        1)
            RESTART_PROJECT
            ;;
        2)
            Default_info
            ;;
        3)
            GET_CONTAINER_STATUS
            ;;
        4)
            STOP_PROJECT
            ;;
        5)
            START_PROJECT
            ;;
        6)
            RESTART_SERVICE_ADMIN
            ;;
        7)
            MODIFY_ADMIN_PASSWORD
            ;;
        8)
            MODIFY_ADMIN_USERNAME
            ;;
        9)
            MODIFY_SAFE_ENTRANCE
            ;;
        10)
            MODIFY_HTTPS_PORT
            ;;
        11)
            GET_SERVICE_TOP_ALL
            ;;
        12)
            Update_BillionMail
            ;;
        help)
            SHOW_HELP
            ;;
        0 | * )
            echo "Cancelled!"
            exit
            ;;
        esac
        ;;
esac
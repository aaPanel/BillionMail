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
    # GetSysInfo
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

# GET_CONTAINER_ID() {

#     if [[ "$1" == "/core:" ]]; then
#         echo "Getting the "$1" container ID..."
#         CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "/core:" | awk '{print $1}' )
#         if [ -z "${CONTAINER_ID}" ]; then
#             echo "ERROR: The "$1" container does not exist"
#         fi
#     fi
# }



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
    CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "/core:" | awk '{print $1}' )
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
    CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "/core:" | awk '{print $1}' )
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
    CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "/core:" | awk '{print $1}' )
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
    CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "/core:" | awk '{print $1}' )
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
    CONTAINER_ID=$(${DOCKER_COMPOSE} ps -a --format "{{.ID}} {{.Image}}" |grep "/core:" | awk '{print $1}' )
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


case "$1" in
    h | -h | help | --help | -help)
        echo "Help Information:"
        echo "  default                   - Show BillionMail login default info: $0 default"
        echo "  update                    - Update BillionMail: $0 update"
        echo "  change-port               - Modify BillionMail access management port: $0 change-port"
        echo "  change-tz                 - Modify BillionMail time zone: $0 change-tz"
        echo "  change-user               - Modify BillionMail Administrator user: $0 change-user"
        echo "  change-password           - Modify BillionMail Administrator password: $0 change-password"
        echo "  change-safe-path          - Modify BillionMail security entrance path: $0 change-safe-path"
        echo "  change-apply-ssl-port     - Modify BillionMail apply ssl port: $0 change-apply-ssl-port"
        echo "  add-domain <domain>       - Add domain. Example: $0 add-domain example.com"
        echo "  del-domain <domain>       - Delete domain. Example: $0 del-domain example.com"
        echo "  add-email <email>         - Add email. Example: $0 add-email user@example.com"
        echo "  del-email <email>         - Delete email. Example: $0 del-email user@example.com"
        echo "  show-domain               - Show domain data."
        echo "  show-email                - Show email data."
        echo "  show-record               - Show domain DNS record."
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
    *)
        echo "Usage: $0 {default|update|change-port|change-tz|show-record|help}"
        exit 1
        ;;
esac
#!/bin/zsh
# BillionMail Instalador para macOS
# Adaptación completa del script original de Linux
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/opt/homebrew/bin:~/bin
export PATH

CONTAINER_PROJECT_NAME=billionmail
PGSQL_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-pgsql-billionmail-1"
DOVECOT_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-dovecot-billionmail-1"
POSTFIX_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-postfix-billionmail-1"
RSPAMD_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-rspamd-billionmail-1"
create_time=$(date +%s)
DBNAME=billionmail
DBUSER=billionmail
SMTP_PORT=25
SMTPS_PORT=465
SUBMISSION_PORT=587
IMAP_PORT=143
IMAPS_PORT=993
POP_PORT=110
POPS_PORT=995
REDIS_PORT=127.0.0.1:26379
SQL_PORT=127.0.0.1:25432
HTTP_PORT=80
HTTPS_PORT=443
SafePath=$(LC_ALL=C tr -dc A-Za-z0-9 < /dev/urandom 2> /dev/null | head -c 8)
REDISPASS=$(LC_ALL=C tr -dc A-Za-z0-9 < /dev/urandom 2> /dev/null | head -c 32)
ADMIN_USERNAME=$(LC_ALL=C tr -dc A-Za-z0-9 < /dev/urandom 2> /dev/null | head -c 8)
ADMIN_PASSWORD=$(LC_ALL=C tr -dc A-Za-z0-9 < /dev/urandom 2> /dev/null | head -c 8)
time=$(date +%Y_%m_%d_%H_%M_%S)

PWD_d=`pwd`
download_Url=https://node.aapanel.com
mirror=''
Default_Download_Url=""

CPU_architecture=$(uname -m)

# List of supported architectures
SUPPORTED_ARCHS=("x86_64" "arm64")

# Check whether the current architecture is supported
if [[ ! " ${SUPPORTED_ARCHS[@]} " =~ " ${CPU_architecture} " ]]; then
    echo -e "\033[31mSorry, not support the ${CPU_architecture} architecture for install. \nPlease use the x86_64, arm64 server architecture. \033[0m"
    exit 1
fi

is64bit=$(getconf LONG_BIT)
if [ "${is64bit}" != '64' ];then
    echo -e "\033[31m Sorry, BillionMail does not support 32-bit systems \033[0m"
    exit 1
fi

if [ $(whoami) != "root" ];then
    echo -e "Non-root install detected. Running on macOS, continuing..."
fi

while [ ${#} -gt 0 ]; do
    case $1 in
        -h|--help)
            echo "Usage:  [options]"
            echo "Options:"
            echo "  -d, --domain          Set Mail Server domain. eg: example.com"
            echo "  -t, --TZ              Set Time Zone. eg: CST"
            echo "   Time Zone See https://en.wikipedia.org/wiki/List_of_tz_database_time_zones for a list of timezones"
            echo "   Use a column named "TZ identifier" + note the column named "Notes""
            echo "eg: zsh install_macos.sh --domain example.com --TZ CST"
            exit 0
            ;;
        -d|--domain)
            BILLIONMAIL_HOSTNAME=$2
            shift 1
            ;;
        -t|--TZ)
            BILLIONMAIL_TIME_ZONE=$2
            shift 1
            ;;
    esac
    shift 1
done

if [ -f "billionmail.conf" ]; then
read -r -p "Check that the configuration file exists, will you continue to overwrite the file?? [y/N] " gogo
case $gogo in
    [Yy][eE][sS]|[Yy])
    if [ ! -d "./backup/" ]; then
        mkdir ./backup
    fi
        mv billionmail.conf ./backup/billionmail.conf_${time}
    echo "Backup: billionmail.conf --> ./backup/billionmail.conf_${time}"
    if [ -f ".env" ]; then
        mv .env env_${time}
        echo "Backup: .env --> ./backup/env_${time}"
    fi
    ;;
    *)
    exit 1
    ;;
esac
fi

if [ -z "${BILLIONMAIL_HOSTNAME}" ]; then
    BILLIONMAIL_HOSTNAME="example.com"
fi

# Count number of dots in the domain
DOT_COUNT=$(echo "${BILLIONMAIL_HOSTNAME}" | tr -cd '.' | wc -c)

# If only one dot, prepend "mail."
if [ "${DOT_COUNT}" -eq 1 ]; then
    ADD_MAIL_BILLIONMAIL_HOSTNAME="mail.${BILLIONMAIL_HOSTNAME}"
    echo "Postfix myhostname configuration use: ${ADD_MAIL_BILLIONMAIL_HOSTNAME}"
fi

# Ensure ADD_MAIL_BILLIONMAIL_HOSTNAME is always set (fallback to original if empty)
if [ -z "${ADD_MAIL_BILLIONMAIL_HOSTNAME}" ]; then
    ADD_MAIL_BILLIONMAIL_HOSTNAME="${BILLIONMAIL_HOSTNAME}"
    echo "Postfix myhostname configuration use: ${ADD_MAIL_BILLIONMAIL_HOSTNAME}"
fi

# macOS timezone detection
SYSTEM_TIME_ZONE=$(readlink /etc/localtime 2>/dev/null | sed -n 's|^.*/zoneinfo/||p')
if [ -z "${SYSTEM_TIME_ZONE}" ]; then
    SYSTEM_TIME_ZONE=$(systemsetup -gettimezone 2>/dev/null | awk -F': ' '{print $2}')
fi

BILLIONMAIL_TIME_ZONE=${SYSTEM_TIME_ZONE}
if [ -z "${BILLIONMAIL_TIME_ZONE}" ]; then
    BILLIONMAIL_TIME_ZONE="America/New_York"
fi

DBPASS_file=DBPASS_file.pl
if [ ! -s "${DBPASS_file}" ]; then
    DBPASS=$(LC_ALL=C tr -dc A-Za-z0-9 < /dev/urandom 2> /dev/null | head -c 32)
    echo "${DBPASS}" > ${DBPASS_file}
    chmod 600 ${DBPASS_file}
else
    DBPASS=$(cat ${DBPASS_file})
fi

GetSysInfo(){
    SYS_VERSION=$(sw_vers -productName 2>/dev/null)
    SYS_VERSION_NUM=$(sw_vers -productVersion 2>/dev/null)
    SYS_INFO=$(uname -a)
    SYS_BIT=$(getconf LONG_BIT)
    MEM_TOTAL=$(sysctl -n hw.memsize | awk '{print int($1/1024/1024)}')
    CPU_INFO=$(sysctl -n hw.ncpu)

    echo -e "${SYS_VERSION} ${SYS_VERSION_NUM}"
    echo -e Bit:${SYS_BIT} Mem:${MEM_TOTAL}M Core:${CPU_INFO}
    echo -e ${SYS_INFO}
    echo -e "Please screenshot the above error message and post to the https://github.com/aaPanel/BillionMail/issues for help"
}

Red_Error(){
    echo '=================================================';
    printf '\033[1;31;40m%b\033[0m\n' "$@";
    GetSysInfo
    exit 1;
}

PORT(){
    command_port="lsof"
    
    if [[ ! -z $command_port ]]; then
        check_command=$(lsof -iTCP -sTCP:LISTEN -n -P | grep -E ":(${SMTP_PORT}|${SMTPS_PORT}|${SUBMISSION_PORT}|${POP_PORT}|${IMAP_PORT}|${IMAPS_PORT}|${POPS_PORT})\s")
        if [[ ! -z "$check_command" ]]; then
            echo "Checking the port is used:"
            echo "$check_command"|grep -v "docker-proxy"
            echo -e "\033[1;31m BillionMail need use port ${SMTP_PORT}|${SMTPS_PORT}|${SUBMISSION_PORT}|${POP_PORT}|${IMAP_PORT}|${IMAPS_PORT}|${POPS_PORT}.\033[0m There are already services ports in the system. "
        fi
    fi
}

Check_Port(){
    PORT
}

Command_Exists() {
    command -v "$@" >/dev/null 2>&1
}

Docker_Check(){
    is_docker="0"
    if Command_Exists docker ; then
        is_docker="1"
    fi
    echo "docker command: $is_docker"
}

Docker_Start() {
    if Command_Exists docker; then
        if ! docker info >/dev/null 2>&1; then
            echo "Docker is not running. Please start Docker Desktop manually."
            open -a Docker
            echo "Waiting for Docker to start..."
            for i in {1..30}; do
                if docker info >/dev/null 2>&1; then
                    echo "Docker is now running."
                    return 0
                fi
                sleep 2
            done
            Red_Error "Docker failed to start. Please start Docker Desktop manually and try again."
        fi
    else
        Red_Error "Docker is not installed. Please install Docker Desktop from https://www.docker.com/products/docker-desktop/"
    fi
}

Docker_Install() {
    echo "Docker installation is required on macOS."
    echo "Please install Docker Desktop from: https://www.docker.com/products/docker-desktop/"
    echo ""
    echo "After installation, run this script again."
    exit 1
}

Docker_Compose_Check(){
    is_docker_compose="0"
    if Command_Exists docker-compose; then
        is_docker_compose="1"
        DOCKER_COMPOSE="docker-compose"
    else 
        if Command_Exists docker; then
            Docker_compose="docker compose version"
            if $Docker_compose >/dev/null 2>&1; then
                is_docker_compose="1"
                DOCKER_COMPOSE="docker compose"
            fi
        else
            is_docker_compose="0"
        fi
    fi
}

Docker_Compose_Install() {
    if [ $is_docker_compose == "0" ]; then
        echo "Docker Compose is not available."
        echo "Please install Docker Desktop which includes Docker Compose."
        exit 1
    fi
}

Bored_waiting(){
    w_time="$wait_time"
    msg="$wait_msg"
    progress="."
    for ((i=0; i<${w_time}; i++))
    do
        printf "$msg %-10s %d\r" "$progress" "$((i+1))"
        sleep 1

        if [ "$progress" == ".........." ]; then
            progress="."
        else
            progress+="."
        fi
    done
    printf "$msg %-10s %d\r" ".........." "$w_time"
}

Check_Connect_PgSql(){
    MAX_RETRIES=10
    retries=0

    while [ $retries -lt $MAX_RETRIES ]; do
        output=$(docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "SELECT 1;" 2>&1)
        if [ $? -eq 0 ]; then
            PgSql_run="1"
            break
        fi

        retries=$((retries + 1))

        echo "PgSql failed to connect, try again after 10 seconds, ${retries}/${MAX_RETRIES}..."

        if [[ "${retries}" == "1" ]] || [[ "${retries}" == "${MAX_RETRIES}" ]]; then
            echo "$connect_PgSql"
        fi
        wait_msg="Please wait.."
        wait_time="10"
        Bored_waiting
    done

    if [ $retries -ge $MAX_RETRIES ]; then
        Red_Error "ERROR: The maximum number of retrys exceeded, and the connection to PgSql cannot be connected. Please try reinstalling!"
    fi
}

Domain_DKIM_record(){
    docker exec -i -e BILLIONMAIL_HOSTNAME=${BILLIONMAIL_HOSTNAME} ${RSPAMD_CONTAINER_NAME} bash -c 'cat << "EOF" > /tmp/1.sh
#!/bin/bash
if [ ! -d "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/" ]; then
    mkdir -p "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/"
fi
if [ -f "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/default.private" ] && [ -f "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/default.pub" ]; then
    exit 0
fi

rspamadm dkim_keygen -s 'default' -b 1024 -d {domain} -k "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/default.private" > "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/default.pub"
if [ $? -eq 0 ]; then
    DKIM_KEYS_DIR="/var/lib/rspamd/dkim"
    CONFIG_FILE="/etc/rspamd/local.d/dkim_signing.conf"
    if [ ! -f "${CONFIG_FILE}" ]; then
        touch "${CONFIG_FILE}"
        echo "${CONFIG_FILE}"
        echo "domain {" > "${CONFIG_FILE}"
        echo "#BT_DOMAIN_DKIM_BEGIN" >> "${CONFIG_FILE}"
        echo "#BT_DOMAIN_DKIM_END" >> "${CONFIG_FILE}"
        echo "}" >> "${CONFIG_FILE}"
    fi
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
    for DOMAIN in "${DOMAINS[@]}"; do
        DOMAIN_NAME=$(echo "${DOMAIN}" | cut -d':' -f1)
        PRIVATE_KEY_PATH=$(echo "${DOMAIN}" | cut -d':' -f2)
        if grep -wq "${DOMAIN_NAME}" "${CONFIG_FILE}"; then
            continue
        fi
        sed -i "/^#BT_DOMAIN_DKIM_BEGIN$/a #${DOMAIN_NAME}_DKIM_BEGIN\n  ${DOMAIN_NAME} {\n    selectors [\n     {\n       path: \"${PRIVATE_KEY_PATH}\";\n       selector: \"default\"\n     }\n   ]\n }\n#${DOMAIN_NAME}_DKIM_END" "${CONFIG_FILE}"
    done
else
    echo -e "DKIM key generation failed!"
    exit 1
fi
chmod 755 -R "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/"
EOF'
    docker exec -i -e BILLIONMAIL_HOSTNAME=${BILLIONMAIL_HOSTNAME} ${RSPAMD_CONTAINER_NAME} bash /tmp/1.sh && rm -f /tmp/1.sh
    DKIM_RECORD=$(docker exec ${RSPAMD_CONTAINER_NAME} cat "/var/lib/rspamd/dkim/${BILLIONMAIL_HOSTNAME}/default.pub")
}

Domain_record() {
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
        echo -e " Type |  Host record       |    Record value     |"
        echo -e "  \e[1;33mTXT\e[0m | \e[1;33mdefault._domainkey\e[0m | \e[1;33m${DKIM_RECORD}\e[0m |<-- Start from \"v=DKIM1\" end, A single line."
        echo -e "==========================================================="
    else
        echo -e "${BILLIONMAIL_HOSTNAME} DKIM key generation failed!"
    fi
}

Init_Billionmail()
{
    SQL_FILE="./init.sql"
    if [ ! -f "${SQL_FILE}" ]; then
        Red_Error "SQL The file does not exist: ${SQL_FILE}"
    fi

    if docker ps --format '{{.Names}}' | grep -q "^${PGSQL_CONTAINER_NAME}$"; then
        
        Check_Connect_PgSql

        echo "Importing database..."
        docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} < ${SQL_FILE}
        if [ $? -eq 0 ]; then
            echo "Database import was successful!"
        else
            Red_Error "Database import failed!"
        fi
        echo "Creating domain..."
        BILLIONMAIL_HOSTNAME=$(echo "${BILLIONMAIL_HOSTNAME}" | tr '[:upper:]' '[:lower:]')
        Check_domain=$(docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "SELECT * FROM domain WHERE domain = '${BILLIONMAIL_HOSTNAME}';" | grep -w "^ ${BILLIONMAIL_HOSTNAME}")
        if [ -z "${Check_domain}" ]; then
            docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "INSERT INTO domain (domain, a_record, mailboxes, mailbox_quota, quota, rate_limit, create_time, active)
            VALUES ('${BILLIONMAIL_HOSTNAME}', 'mail.${BILLIONMAIL_HOSTNAME}', 500, 5368709120, 5368709120, 12, ${create_time}, 1);"
            if [ $? -eq 0 ]; then
                echo "Domain creation was successful!"
                Domain_DKIM_record
            else
                Red_Error "Domain creation failed!"
            fi
        else
            echo ""${Check_domain}" Domain already exists!"
        fi

        if [ -z "${mailbox}" ]; then
            mailbox=$(LC_ALL=C tr -dc a-z0-9 < /dev/urandom 2> /dev/null | head -c 6)
            echo "Generate mailbox: ${mailbox}"
        else
            mailbox=$(echo "${mailbox}" | tr '[:upper:]' '[:lower:]')
        fi
        Generate_mailbox_password=$(LC_ALL=C tr -dc A-Za-z0-9 < /dev/urandom 2> /dev/null | head -c 16)
        Encrypt_mailbox_password=$(docker exec -i ${DOVECOT_CONTAINER_NAME} doveadm pw -s MD5-CRYPT -p "${Generate_mailbox_password}" | sed 's/{MD5-CRYPT}//')
        if [ $? -eq 0 ]; then
            echo "Generate_mailbox_password: ${Generate_mailbox_password}"
            echo "mailbox_password: ${Encrypt_mailbox_password}"
        else
            Generate_mailbox_password="BILLIONMAIL"
            Encrypt_mailbox_password='$1$ELBUCcYE$TbdGKBvLkFbjQguDbi3s01'
            echo "Generate_mailbox_password--default: ${Generate_mailbox_password}"
            Default_password=1
        fi

        if [ "${Default_password}" != 1 ]; then
            b64_data=$(echo -n "${Generate_mailbox_password}" | base64)
            password_encode=$(echo -n "${b64_data}" | while IFS= read -r -n1 char; do printf "%02x" "'$char"; done)
            echo "password_encode: ${password_encode}"
            if [ -z ${password_encode} ]; then
                b64_data=$(echo -n "${Generate_mailbox_password}" | base64)
                password_encode=$(echo -n "${b64_data}" | hexdump -ve '1/1 "%.2x"')
                echo "password_encode: ${password_encode}"
            fi
        else
            password_encode="516b6c4d54456c50546e31425355773d"
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
            docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "$INSERT_mailbox"
            if [ $? -eq 0 ]; then
                echo "Mailbox creation was successful!"
            else
                Red_Error "Mailbox creation failed!"
            fi
        else
            echo ""${Check_mailbox}" Mailbox already exists!"
        fi

    else
        Red_Error "PgSql container does not exist!"
    fi
}

Billionmail(){
    Check_Port=$(lsof -iTCP:${HTTP_PORT} -sTCP:LISTEN -n -P 2>/dev/null)
    if [ ! -z "${Check_Port}" ]; then
        HTTP_PORT=5678
        Check_Port=$(lsof -iTCP:${HTTP_PORT} -sTCP:LISTEN -n -P 2>/dev/null)
    fi
    if [ ! -z "${Check_Port}" ]; then
        echo -e "${HTTP_PORT} Already used, random ports are being used"
        while true; do
        HTTP_PORT=$((RANDOM % 55535 + 10000))
        if ! lsof -iTCP:${HTTP_PORT} -sTCP:LISTEN >/dev/null 2>&1; then
            echo "${HTTP_PORT}"
            break
        fi
        done
    fi

    Check_Port22=$(lsof -iTCP:${HTTPS_PORT} -sTCP:LISTEN -n -P 2>/dev/null)
    if [ ! -z "${Check_Port22}" ]; then
        HTTPS_PORT=5679
        Check_Port22=$(lsof -iTCP:${HTTPS_PORT} -sTCP:LISTEN -n -P 2>/dev/null)
    fi
    if [ ! -z "${Check_Port22}" ]; then
        echo -e "${HTTPS_PORT} Already used, random ports are being used"
        while true; do
        HTTPS_PORT=$((RANDOM % 55535 + 10000))
        if ! lsof -iTCP:${HTTPS_PORT} -sTCP:LISTEN >/dev/null 2>&1; then
            echo "${HTTPS_PORT}"
            break
        fi
        done
    fi

    cat << EOF > billionmail.conf
# Default BillionMail Username password
ADMIN_USERNAME=${ADMIN_USERNAME}
ADMIN_PASSWORD=${ADMIN_PASSWORD}

# Manage Safe entrance
SafePath=${SafePath}

# BILLIONMAIL_HOSTNAME configuration, Postfix myhostname configuration
BILLIONMAIL_HOSTNAME=${ADD_MAIL_BILLIONMAIL_HOSTNAME}

# pgsql NAME and USER and PASSWORD configuration

DBNAME=${DBNAME}
DBUSER=${DBUSER}
DBPASS=${DBPASS}

# REDIS PASSWORD configuration
REDISPASS=${REDISPASS}


## MAIL Ports
SMTP_PORT=${SMTP_PORT}
SMTPS_PORT=${SMTPS_PORT}
SUBMISSION_PORT=${SUBMISSION_PORT}
IMAP_PORT=${IMAP_PORT}
IMAPS_PORT=${IMAPS_PORT}
POP_PORT=${POP_PORT}
POPS_PORT=${POPS_PORT}
REDIS_PORT=${REDIS_PORT}
SQL_PORT=${SQL_PORT}

## Manage Ports
HTTP_PORT=${HTTP_PORT}
HTTPS_PORT=${HTTPS_PORT}

# You can use this script to set the time zone for your container.
# See https://en.wikipedia.org/wiki/List_of_tz_database_time_zones for a list of timezones"
# echo -e "Use a column named "TZ identifier" + note the column named "Notes""

TZ=${BILLIONMAIL_TIME_ZONE}

# Default containers IPV4 intranet segment
IPV4_NETWORK=172.66.1

# Enable fail2ban Access restrictions, specify that the IP exceeds the access limit
FAIL2BAN_INIT=y

# Number of days to keep log backup
RETENTION_DAYS=7

EOF
    \cp -rf billionmail.conf .env
    if [ ! -f ".env" ]; then
        echo -e "Error: Failed to create .env file"
        exit 1
    fi

    SSL_path=ssl-self-signed
    if [ ! -d "ssl-self-signed" ]; then
        mkdir ssl-self-signed
    fi
    openssl genrsa -out ${SSL_path}/key.pem 2048
    openssl req -x509 -new -nodes -key ${SSL_path}/key.pem -sha256 -days 3650 -out ${SSL_path}/cert.pem \
    -subj "/C=US/ST=State/L=City/O=${BILLIONMAIL_HOSTNAME}/OU=${BILLIONMAIL_HOSTNAME}/CN=*.${BILLIONMAIL_HOSTNAME}" -nodes
    mkdir -p ssl
    cp -n ${SSL_path}/* ssl/ 2>/dev/null || true

    echo -e "Execute ${DOCKER_COMPOSE} up -d"
    if  [ ! -s "docker-compose.yml" ]; then
        ls -al
        Red_Error "docker-compose.yml not found."
    fi
    ${DOCKER_COMPOSE} pull
    ${DOCKER_COMPOSE} up -d
    if [ $? -eq 0 ]; then
        echo -e "Billionmail installation completed successfully!"
    else
        echo ""
        echo -e "--------------------------------------------------"
        Check_Port
        echo -e "\e[1;31m Startup error,\e[0m please resolve it according to the prompts, otherwise it will affect the use!"
        sleep 5
        bash bm.sh status
        echo -e "--------------------------------------------------"
        echo ""
    fi

    [ ! -d "/opt" ] && mkdir -p /opt
    echo "${PWD_d}" > /opt/PWD-Billion-Mail.txt
    
    if [ -w /usr/local/bin ]; then
        ln -sf ${PWD_d}/bm.sh /usr/local/bin/bm
    else
        echo "Note: Could not create symlink in /usr/local/bin. You may need to run: sudo ln -sf ${PWD_d}/bm.sh /usr/local/bin/bm"
    fi
    chmod +x ${PWD_d}/bm.sh
}

Install_Main(){
    echo "Checking system requirements for macOS..."
    
    # Check Homebrew
    if ! Command_Exists brew; then
        echo "Homebrew is not installed. Installing Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        
        # Add Homebrew to PATH for Apple Silicon
        if [ -f /opt/homebrew/bin/brew ]; then
            eval "$(/opt/homebrew/bin/brew shellenv)"
            echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
        fi
    fi
    
    echo "Installing/updating required packages..."
    brew install coreutils gnu-sed curl wget openssl
    
    set +e
    Docker_Check
    Docker_Compose_Check
    
    if [ "$is_docker" != "1" ]; then
        echo "Docker not found. Please install Docker Desktop for Mac:"
        echo "https://www.docker.com/products/docker-desktop/"
        echo ""
        read -p "Press Enter after installing Docker Desktop to continue..."
        Docker_Check
        if [ "$is_docker" != "1" ]; then
            Red_Error "Docker still not found. Please install Docker Desktop and try again."
        fi
    fi

    if [ "$is_docker_compose" != "1" ]; then
        Docker_Compose_Check
        if [ "$is_docker_compose" != "1" ]; then
            Red_Error "ERROR: Docker Compose not available. Please ensure Docker Desktop is properly installed."
        fi
    fi
   
    Docker_Start

    Billionmail
    
    echo "Note: Firewall configuration on macOS should be done manually if needed."
}

Install_Main

IPV4_ADDRESS=$(curl -sSf -4 --connect-timeout 10 -m 20 https://ifconfig.me)
if [ -z "${IPV4_ADDRESS}" ]; then
    IPV4_ADDRESS=$(curl -sSfk --connect-timeout 10 -m 20 https://www.aapanel.com/api/common/getClientIP)
fi
ipv4_regex="^([0-9]{1,3}\.){3}[0-9]{1,3}$"
if [[ ${IPV4_ADDRESS} =~ ${ipv4_regex} ]]; then        
    echo "${IPV4_ADDRESS}" >/dev/null 2>&1
elif [ -z "${IPV4_ADDRESS}" ]; then
    IPV4_ADDRESS="YOUR_SERVER_IPV4_ADDRESS"
fi

intenal_ip=$(ipconfig getifaddr en0 2>/dev/null)
if [ -z "${intenal_ip}" ]; then
    intenal_ip=$(ipconfig getifaddr en1 2>/dev/null)
fi
if [ -z "${intenal_ip}" ]; then
    intenal_ip="localhost"
fi

if [ ${HTTPS_PORT} = "443" ]; then
    echo -e "BillionMail Internet address: \e[1;33mhttps://${IPV4_ADDRESS}/${SafePath}\e[0m"
    echo -e "BillionMail Internal address: \e[1;33mhttps://${intenal_ip}/${SafePath}\e[0m"
else
    echo -e "BillionMail Internet address: \e[1;33mhttps://${IPV4_ADDRESS}:${HTTPS_PORT}/${SafePath}\e[0m"
    echo -e "BillionMail Internal address: \e[1;33mhttps://${intenal_ip}:${HTTPS_PORT}/${SafePath}\e[0m"
fi
echo -e "Username: \e[1;33m${ADMIN_USERNAME}\e[0m"
echo -e "Password: \e[1;33m${ADMIN_PASSWORD}\e[0m"
echo -e ""
echo -e "Tip: Use \e[33m bm \e[0m or \e[33mzsh bm.sh\e[0m to View login info etc."

curl -o /dev/null -fsSLk --connect-time 10 -X POST "https://www.aapanel.com/api/panel/panel_count_daily?name=billionmail" >/dev/null 2>&1

#!/bin/bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin
export PATH

CONTAINER_PROJECT_NAME=billionmail
PGSQL_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-pgsql-billionmail-1"
DOVECOT_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-dovecot-billionmail-1"
POSTFIX_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-postfix-billionmail-1"
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
REDISPASS=$(LC_ALL=C </dev/urandom tr -dc A-Za-z0-9 2> /dev/null | head -c 32)
time=$(date +%Y_%m_%d_%H_%M_%S)

PWD_d=`pwd`
download_Url=https://node.aapanel.com
mirror=''
Default_Download_Url=""

CPU_architecture=$(uname -m)
if [ "${CPU_architecture}" == "s390x" ];then
    echo -e "\033[31mSorry, not support the s390x architecture for install. \nPlease use the x86_64 server architecture. \033[0m"
    exit 1
elif [ "${CPU_architecture}" == "aarch64" ];then
    #CPU_architecture="-aarch64"
    echo " ${CPU_architecture} architecture is not currently supported."
    uname -a
    exit 1
elif [ "${CPU_architecture}" == "x86_64" ];then
    # x86_64 is empty by default
    CPU_architecture=""
else
    echo " ${CPU_architecture} architecture is not currently supported."
    uname -a
    exit 1
fi

is64bit=$(getconf LONG_BIT)
if [ "${is64bit}" != '64' ];then
    echo -e "\033[31m Sorry, BillionMail does not support 32-bit systems \033[0m"
    exit 1
fi


if [ $(whoami) != "root" ];then
    echo -e "Non-root install, please try the following solutions: \n   1.Please switch to root user install \n   2.Try executing the following install commands: \n     sudo bash $0 $@"
    exit 1
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
            echo "eg: bash install.sh --domain example.com --TZ CST"
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

while [ -z "${BILLIONMAIL_HOSTNAME}" ]; do
echo "Press Enter to confirm the detected value '[value]', or enter a custom value."
echo -e ""
    echo -e "Mail Server hostname (FQDN), As: example.com"
    echo -e ""
    read -p "Please enter the Mail Server hostname (FQDN): " -e BILLIONMAIL_HOSTNAME
    if [[ ! "${BILLIONMAIL_HOSTNAME}" =~ ^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9](\.[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9])*$ ]]; then
        echo -e "\e[31m(${BILLIONMAIL_HOSTNAME}) is not a FQDN!\e[0m"
        echo "Please change it to a FQDN"
        exit 1
    elif [[ "${BILLIONMAIL_HOSTNAME: -1}" == "." ]]; then
        echo "(${BILLIONMAIL_HOSTNAME}) is ending with a dot. This is not a valid FQDN!"
        exit 1

    fi
done

if [ -a /etc/timezone ]; then
    SYSTEM_TIME_ZONE=$(cat /etc/timezone)
elif [ -a /etc/localtime ]; then
    SYSTEM_TIME_ZONE=$(readlink /etc/localtime|sed -n 's|^.*zoneinfo/||p')
fi

while [ -z "${BILLIONMAIL_TIME_ZONE}" ]; do
    echo -e ""
    echo -e "See https://en.wikipedia.org/wiki/List_of_tz_database_time_zones for a list of timezones"
    echo -e "Use a column named "TZ identifier" + note the column named "Notes""
    echo -e "Please enter your time zone"
    echo -e ""
    if [ -z "${SYSTEM_TIME_ZONE}" ]; then
        read -p "Timezone: " -e BILLIONMAIL_TIME_ZONE
    else
        read -p "Timezone [${SYSTEM_TIME_ZONE}]: " -e BILLIONMAIL_TIME_ZONE
        [ -z "${BILLIONMAIL_TIME_ZONE}" ] && BILLIONMAIL_TIME_ZONE=${SYSTEM_TIME_ZONE}
    fi
done

DBPASS_file=DBPASS_file.pl
if [ ! -s "${DBPASS_file}" ]; then
    DBPASS=$(LC_ALL=C </dev/urandom tr -dc A-Za-z0-9 2> /dev/null | head -c 32)
    echo "${DBPASS}" > ${DBPASS_file}
    chmod 600 ${DBPASS_file}
else
    DBPASS=$(cat ${DBPASS_file})
fi


GetSysInfo(){
    if [ -s "/etc/redhat-release" ];then
        SYS_VERSION=$(cat /etc/redhat-release)
    elif [ -s "/etc/issue" ]; then
        SYS_VERSION=$(cat /etc/issue)
    fi
    SYS_INFO=$(uname -a)
    SYS_BIT=$(getconf LONG_BIT)
    MEM_TOTAL=$(free -m|grep Mem|awk '{print $2}')
    CPU_INFO=$(getconf _NPROCESSORS_ONLN)

    echo -e ${SYS_VERSION}
    echo -e Bit:${SYS_BIT} Mem:${MEM_TOTAL}M Core:${CPU_INFO}
    echo -e ${SYS_INFO}
    echo -e "Please screenshot the above error message and post to the forum forum.aapanel.com for help"

}


Red_Error(){
    echo '=================================================';
    printf '\033[1;31;40m%b\033[0m\n' "$@";
    GetSysInfo
    exit 1;
}



PORT(){

    if Command_Exists ss ; then
        command_port="ss"
        
    elif Command_Exists netstat ; then
        command_port="netstat"
    else
        echo -e "Command does not exist and cannot check whether the port is used"
        command_port=""
    fi

    if [[ ! -z $command_port ]]; then
        echo "Checking if the port is used"
        check_command=$($command_port -ltnp | grep -E ':(25|465|587|110|143|993|995)\s')

        if [[ ! -z "$check_command" ]]; then
            echo "$check_command"
            echo -e "\033[31m Mail Server need use port 25 465 587 110 143 993 995. There are already services ports in the system, cannot be installed. \nRecommended use a new system. \033[0m"
            exit 1
        fi
    fi
}

Check_Port(){
    if ! Command_Exists docker; then
        PORT
    else
        check=$(docker ps |grep -wE "${POSTFIX_CONTAINER_NAME}|${DOVECOT_CONTAINER_NAME}")
        if [[ -z "$check" ]]; then
            PORT
        fi
    fi
}


LXC_install_tips (){
    # Check if it is in the lxc container
    System_Info=$(hostnamectl)
    check_container=$(echo "$System_Info" |grep "Chassis"|awk -F":" '{print $2}')
    #echo $check_container
    if [[ "$check_container" =~ "container" ]];then
        echo -e "\033[33mInstalling in an LXC container,May not install properly. \nDocker may not be able to correctly obtain the number of CPU cores, \nwhich may cause the system to crash and cause high load when nginx is running. \nRecommended use VPS or physical machine for install.\033[0m"
        echo -e ""
        echo -e "\033[33mMust be turned on \"Unprivileged container\", \"nesting\". Otherwise the container cannot be started.\033[0m"
        echo "Please confirm whether continue install?"
        while [ "$lxc" != 'y' ] && [ "$lxc" != 'n' ]; do
            read -p "Please confirm whether in LXC container continue install? (y/n): " lxc;
        done

        if [ "$lxc" == 'n' ];then
            echo "Install canceled."
            exit;
        fi

        LXC_install="yes"
    fi
}


Get_Pack_Manager(){
	if [ -f "/usr/bin/yum" ] && [ -d "/etc/yum.repos.d" ]; then
		PM="yum"
	elif [ -f "/usr/bin/apt-get" ] && [ -f "/usr/bin/dpkg" ]; then
		PM="apt-get"		
	fi
}


Install_RPM_Pack(){
	yumPacks="wget curl tar gzip git firewalld ca-certificates";
	yum install -y ${yumPacks}

	for yumPack in ${yumPacks}
	do
		rpmPack=$(rpm -q ${yumPack})
		packCheck=$(echo ${rpmPack}|grep not)
		if [ "${packCheck}" ]; then
			yum install ${yumPack} -y
		fi
	done

    setenforce 0
    sed -i 's/SELINUX=enforcing/SELINUX=disabled/' /etc/selinux/config
}

Install_Deb_Pack(){
    debPacks="wget curl tar gzip git lsb-release ufw ca-certificates";
	apt-get update -y
	DEBIAN_FRONTEND=noninteractive apt-get install -y $debPacks --allow-downgrades --allow-remove-essential --allow-change-held-packages

	for debPack in ${debPacks}
	do
		packCheck=$(dpkg -l ${debPack})
		if [ "$?" -ne "0" ] ;then
			apt-get install -y $debPack
		fi
	done
}

Command_Exists() {
    command -v "$@" >/dev/null 2>&1
}

Docker_Download_Url_Check() {
    echo -e "Selecting download docker address source, please wait..."
    ser_names=(download.docker.com mirrors.ustc.edu.cn/docker-ce mirrors.tuna.tsinghua.edu.cn/docker-ce mirrors.aliyun.com/docker-ce mirror.azure.cn/docker-ce)
    tmp_file1=/dev/shm/net_test1.pl
	tmp_file2=/dev/shm/net_test2.pl
    [ -f "${tmp_file1}" ] && rm -f ${tmp_file1}
    [ -f "${tmp_file2}" ] && rm -f ${tmp_file2}
    touch $tmp_file1
    touch $tmp_file2

    for ser_name in ${ser_names[@]};
    do
        NODE_CHECK=$(curl -k -s 2>/dev/null --connect-timeout 5 -w "%{http_code} %{time_total}" https://${ser_name} -o c${ser_name}.txt|xargs)
        rm -rf c${ser_name}.txt
        NODE_STATUS=$(echo ${NODE_CHECK}|awk '{print $1}')
        TIME_TOTAL=$(echo ${NODE_CHECK}|awk '{print $2 * 1000 - 500 }'|cut -d '.' -f 1)
        if [ "${NODE_STATUS}" == "200" ] || [ "${NODE_STATUS}" == "301" ] || [ "${NODE_STATUS}" == "403" ];then
            if [ $TIME_TOTAL -lt 100 ]; then
                echo "$ser_name" >> $tmp_file1
            elif [ $TIME_TOTAL -gt 1500 ] && [ $TIME_TOTAL -lt 2500 ]; then
                echo "$ser_name" >> $tmp_file2
            fi
        fi
    done
    Docker_NODE_URL=$(cat $tmp_file1|head -n 1|awk '{print $1}')
        if [ -z "$Docker_NODE_URL" ];then
            Docker_NODE_URL=$(cat $tmp_file2|head -n 1|awk '{print $1}')
        fi

    rm -f $tmp_file1
    rm -f $tmp_file2
    mirror="$Docker_NODE_URL"
    echo $mirror
    case "$mirror" in
    mirrors.aliyun.com/docker-ce)
        Default_Download_Url="https://mirrors.aliyun.com/docker-ce"
        mirror="Aliyun"
        ;;
    mirror.azure.cn/docker-ce)
        Default_Download_Url="https://mirror.azure.cn/docker-ce"
        mirror="AzureChinaCloud"
        ;;
    mirrors.ustc.edu.cn/docker-ce)
        Default_Download_Url="https://mirrors.ustc.edu.cn/docker-ce"
        mirror="USTC"
        ;;
    mirrors.tuna.tsinghua.edu.cn/docker-ce)
        Default_Download_Url="https://mirrors.tuna.tsinghua.edu.cn/docker-ce"
        mirror="tsinghua"
        ;;
    *)
        Default_Download_Url="https://download.docker.com"
        mirror=""
        ;;
    esac

    DEFAULT_REPO_FILE="docker-ce.repo"

    if [ -z "$REPO_FILE" ]; then
        REPO_FILE="$DEFAULT_REPO_FILE"
    fi
}

Get_Distribution() {
    lsb_dist=""
    lsb_name=""

    if [ -r /etc/os-release ]; then
        lsb_dist="$(. /etc/os-release && echo "$ID")"
        lsb_name="$(. /etc/os-release && echo "$NAME")"
    fi
    echo "$lsb_dist $lsb_name"
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
    #echo ""
}

Docker_Start() {
    if Command_Exists docker || [ -f /lib/systemd/system/docker.service ]; then
        if Command_Exists systemctl; then
            # systemctl mask getty@tty1.service
            systemctl enable docker
            systemctl reset-failed docker.service
            systemctl start docker.service
            if [ "$?" != "0" ];then
                systemctl reset-failed docker.service
                systemctl restart docker.service
                if [ "$?" != "0" ];then
                    docker_status=$(systemctl status docker.service)
                    echo "$docker_status"
                    if [ -f /etc/docker/daemon.json ]; then
                        echo "/etc/docker/daemon.json file content:"
                        cat /etc/docker/daemon.json
                        echo -e "\033[33mRecommended check the above /etc/docker/daemon.json Are the file contents correct?\033[0m"
                    fi
                    # Wait for 100 seconds of animation
                    echo -e "\033[31m start docker service failed, maybe too fast to start\033[0m"
                    wait_msg="waiting 100 seconds to start docker service"
                    wait_time="100"
                    Bored_waiting
                    systemctl reset-failed docker.service
                    systemctl restart docker.service
                    if [ "$?" != "0" ];then
                        docker_status=$(systemctl status docker.service)
                        echo "$docker_status"
                        Red_Error "docker start failed" "Please try again later execute the restart docker command: systemctl restart docker.service (no content returns normal)" "Finally, re-execute install BillionMail"
                    fi
                fi
            fi
        else
            chkconfig --add docker
            chkconfig --level 2345 docker on
            service docker start
            if [ "$?" != "0" ];then
                service docker.service status
                cat /etc/docker/daemon.json
                Red_Error "docker start failed"
            fi
        fi
    else
        
        Red_Error "Docker is not installed"
    
    
    
    fi
}


Is_Darwin() {
    case "$(uname -s)" in
    *darwin*) true ;;
    *Darwin*) true ;;
    *) false ;;
    esac
}

Check_Forked() {
    if Command_Exists lsb_release; then
        set +e
        lsb_release -a -u >/dev/null 2>&1
        lsb_release_exit_code=$?

        if [ "$lsb_release_exit_code" = "0" ]; then
            cat <<-EOF
You're using '$lsb_dist' version '$dist_version'.
EOF
            lsb_dist=$(lsb_release -a -u 2>&1 | tr '[:upper:]' '[:lower:]' | grep -E 'id' | cut -d ':' -f 2 | tr -d '[:space:]')
            dist_version=$(lsb_release -a -u 2>&1 | tr '[:upper:]' '[:lower:]' | grep -E 'codename' | cut -d ':' -f 2 | tr -d '[:space:]')
            cat <<-EOF
Upstream release is '$lsb_dist' version '$dist_version'.
EOF
        else
            if [ -r /etc/debian_version ] && [ "$lsb_dist" != "ubuntu" ] && [ "$lsb_dist" != "raspbian" ]; then
                if [ "$lsb_dist" = "osmc" ]; then
                    lsb_dist=raspbian
                else
                    lsb_dist=debian
                fi
                dist_version="$(sed 's/\/.*//' /etc/debian_version | sed 's/\..*//')"
                case "$dist_version" in
                12)
	                dist_version="bookworm"
                    ;;
                11)
                    dist_version="bullseye"
                    ;;
                10)
                    dist_version="buster"
                    ;;
                9)
                    dist_version="stretch"
                    ;;
                8)
                    dist_version="jessie"
                    ;;
                esac
            fi
        fi
    fi
}


Docker_Check(){

    # Default value is: 0 Not installed
    is_docker="0"
    docker_version="docker version"
    if Command_Exists $docker_version ; then
        is_docker="1"
    fi
    echo "docker command: $is_docker"

}

create_docker_yum_repo() {
    cat <<EOF >/etc/yum.repos.d/docker-ce.repo
[docker-ce-stable]
name=Docker CE Stable - \$basearch
baseurl=https://download.docker.com/linux/centos/\$releasever/\$basearch/stable
enabled=1
gpgcheck=1
gpgkey=https://download.docker.com/linux/centos/gpg

EOF
}


Docker_Install() {
    lsb_dist=$(Get_Distribution|awk -F " " '{print $1}')
    lsb_dist="$(echo "$lsb_dist" | tr '[:upper:]' '[:lower:]')"

    case "$lsb_dist" in
    ubuntu)
        if Command_Exists lsb_release; then
            dist_version="$(lsb_release --codename | cut -f2)"
        fi
        if [ -z "$dist_version" ] && [ -r /etc/lsb-release ]; then
            dist_version="$(. /etc/lsb-release && echo "$DISTRIB_CODENAME")"
        fi
        ;;
    debian | raspbian)
        dist_version="$(sed 's/\/.*//' /etc/debian_version | sed 's/\..*//')"
        case "$dist_version" in
        12)
            dist_version="bookworm"
            ;;
        11)
            dist_version="bullseye"
            ;;
        10)
            dist_version="buster"
            ;;
        9)
            dist_version="stretch"
            ;;
        8)
            dist_version="jessie"
            ;;
        esac
        ;;
    centos | rhel | sles | ol | tencentos | alinux | anolis | rocky | euleros | almalinux | opencloudos |openeuler |hce | kylin |uos | kylinsecos |amzn )
        if [ -z "$dist_version" ] && [ -r /etc/os-release ]; then
            dist_version="$(. /etc/os-release && echo "$VERSION_ID")"
        fi
        ;;
    *)
        if Command_Exists lsb_release; then
            dist_version="$(lsb_release --release | cut -f2)"
        fi
        if [ -z "$dist_version" ] && [ -r /etc/os-release ]; then
            dist_version="$(. /etc/os-release && echo "$VERSION_ID")"
        fi
        ;;
    esac

    Check_Forked

    case "$lsb_dist" in
    ubuntu | debian | raspbian)
        docker_gpg="/usr/share/keyrings/docker-archive-keyring.gpg"
        apt_repo_file="/etc/apt/sources.list.d/docker.list"
        pre_reqs="apt-transport-https ca-certificates curl wget"

        if ! command -v gpg >/dev/null; then
            pre_reqs="$pre_reqs gnupg"
        fi

        apt_repo="deb [arch=$(dpkg --print-architecture) signed-by=$docker_gpg] $Default_Download_Url/linux/$lsb_dist $dist_version stable"
        (
            DEBIAN_FRONTEND=noninteractive apt-get install -y $pre_reqs
            if [ -f $docker_gpg ] && [ -f $apt_repo_file ]; then
                rm -rf $docker_gpg
                rm -rf $apt_repo_file
            fi

            #curl -fsSL --connect-time 10 --retry 5 $Default_Download_Url/linux/$lsb_dist/gpg | gpg --dearmor --yes -o $docker_gpg
            #wget --no-check-certificate -qO - -t 5 -T 20 $Default_Download_Url/linux/$lsb_dist/gpg | gpg --dearmor --yes -o $docker_gpg
            wget -qO - -t 5 -T 10 $Default_Download_Url/linux/$lsb_dist/gpg | gpg --dearmor --yes -o $docker_gpg
            chmod a+r $docker_gpg

            echo "$apt_repo" > $apt_repo_file
            apt-get update -y
        )
        pkg_version=""

        #First check whether apt-get is used. It cannot be put into the subshell, that is, () to execute, otherwise it will not exit.
        # Check_apt_status

        (
            pkgs="$pkgs docker-ce${pkg_version%=}"
            DEBIAN_FRONTEND=noninteractive apt-get install -y $pkgs docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
        )
        ;;
    centos | fedora | rhel | ol | tencentos | alinux | anolis | rocky | almalinux | opencloudos |openeuler |hce |uos |amzn )
        pkg_version=""

        yum_repo="$Default_Download_Url/linux/$lsb_dist/$REPO_FILE"
        if [ "$lsb_dist" == "ol" ] || [ "$lsb_dist" == "tencentos" ] || [ "$lsb_dist" == "alinux" ] || [ "$lsb_dist" == "anolis" ] || [ "$lsb_dist" == "opencloudos" ] || [ "$lsb_dist" == "openeuler" ] || [ "$lsb_dist" == "hce" ] || [ "$lsb_dist" == "uos" ] || [ "$lsb_dist" == "amzn" ] ; then
            yum_repo="$Default_Download_Url/linux/centos/$REPO_FILE"
        fi
        if [ "$lsb_dist" == "rocky" ] || [ "$lsb_dist" == "almalinux" ]; then
            yum_repo="$Default_Download_Url/linux/centos/$REPO_FILE"
        fi
        if [ "$lsb_dist" = "fedora" ]; then
            pkg_manager="dnf"
            config_manager="dnf config-manager"
            enable_channel_flag="--set-enabled"
            disable_channel_flag="--set-disabled"
            pre_reqs="dnf-plugins-core"
            pkg_suffix="fc$dist_version"
        else
            pkg_manager="yum"
            config_manager="yum-config-manager"
            enable_channel_flag="--enable"
            disable_channel_flag="--disable"
            pre_reqs="yum-utils"
            pkg_suffix="el"
        fi
    
        #wget --no-check-certificate -O /etc/yum.repos.d/docker-ce.repo $yum_repo -t 5 -T 10
        wget -O /etc/yum.repos.d/docker-ce.repo $yum_repo -t 5 -T 10
        if [ $? -ne "0" ]; then
            create_docker_yum_repo
        fi
        
        if ! grep -q "docker" /etc/yum.repos.d/docker-ce.repo; then
            create_docker_yum_repo
        fi

        sed -i "s|https://download.docker.com|$Default_Download_Url|g" /etc/yum.repos.d/docker-ce.repo
        # cat /etc/yum.repos.d/docker-ce.repo
       
        lsb_name=$(Get_Distribution|awk -F " " '{print $3}')
        echo $lsb_name
        echo $lsb_dist
        #if [ "$lsb_name" = "Stream" ] || [ "$lsb_dist" = "alinux" ] || [ "$lsb_dist" = "anolis" ];then
        #if [ "$lsb_name" = "Stream" ] || [ "$lsb_dist" = "anolis" ];then
        if [ "$lsb_name" = "Stream" ]; then
            conflicting="--allowerasing"
        fi
        if [ "$lsb_name" = "rocky" ] || [ "$lsb_dist" = "euleros" ] || [ "$lsb_dist" = "almalinux" ];then
            conflicting="--allowerasing"
        fi

        if [ "$lsb_dist" = "openeuler" ];then
            openeuler_v=$(cat /etc/os-release |grep "VERSION=")
            if [[ $openeuler_v =~ "23.0" ]]; then
                
                sed -i "s|\$releasever|8|g" /etc/yum.repos.d/docker-ce.repo
                $pkg_manager makecache
                
            elif [[ $openeuler_v =~ "22.0" ]]; then
                sed -i "s|\$releasever|7|g" /etc/yum.repos.d/docker-ce.repo
                $pkg_manager makecache
            else
                if [ -f "/etc/os-release" ]; then
                    cat /etc/os-release
                fi
                echo "ERROR: not support this system yum auto install docker"
            fi
        fi

        # Amazon Linux
        if [ "$lsb_dist" = "amzn" ];then
            #amzn_v=$(. /etc/os-release && echo "$VERSION_ID")

            sed -i "s|\$releasever|8|g" /etc/yum.repos.d/docker-ce.repo
            $pkg_manager makecache
        fi

        # Huawei Cloud EulerOS 
        if [ "$lsb_dist" = "hce" ];then
            hce_v=$(. /etc/os-release && echo "$VERSION_ID")
            if [[ $hce_v =~ "2.0" ]]; then
                sed -i "s|\$releasever|8|g" /etc/yum.repos.d/docker-ce.repo
                $pkg_manager makecache
            else
                if [ -f "/etc/os-release" ]; then
                    cat /etc/os-release
                fi
                echo "ERROR: not support this system yum auto install docker"
            fi
        fi

        # OpenCloudOS 9.0
        if [ "$lsb_dist" = "opencloudos" ];then
            pkg_manager="yum"
            opencloudos_v=$(. /etc/os-release && echo "$VERSION_ID")
            if [[ $opencloudos_v =~ "9.0" ]]; then

                sed -i "s|\$releasever|8|g" /etc/yum.repos.d/docker-ce.repo
                # centos 8
                download_container_selinux="$download_Url/cloudwaf/package/container-selinux-2.224.0-1.module_el8+712+4cd1bd69.noarch.rpm"
                wget --no-check-certificate -O container-selinux-2.224.0-1.module_el8+712+4cd1bd69.noarch.rpm ${download_container_selinux} -t 5 -T 10

                install_docker="1"
                $pkg_manager makecache
                
                $pkg_manager install -y docker-ce$pkg_version docker-ce-cli docker-buildx-plugin docker-compose-plugin container-selinux-2.224.0-1.module_el8+712+4cd1bd69.noarch.rpm

                rm -f container-selinux-2.224.0-1.module_el8+712+4cd1bd69.noarch.rpm

            fi
        fi

        if [ "$lsb_dist" = "uos" ];then
            pkg_manager="yum"
            uos_v=$(. /etc/os-release && echo "$VERSION_ID")
            uos_NAME=$(. /etc/os-release && echo "$NAME")
            if [[ $uos_v =~ "20" ]] && [[ "$uos_NAME" == "UOS Server Euler" ]]; then
            # UOS Server Euler 20 Install using the default source
                if [ -f /etc/yum.repos.d/docker-ce.repo ]; then
                    rm -f /etc/yum.repos.d/docker-ce.repo
                fi
                # Define to install docker using other commands
                install_docker="1"
                $pkg_manager makecache

                $pkg_manager install -y docker* --skip-broken

            elif [[ $uos_v =~ "20" ]]; then
                sed -i "s|\$releasever|8|g" /etc/yum.repos.d/docker-ce.repo

                # Define to install docker using other commands
                install_docker="1"

                $pkg_manager makecache

                $pkg_manager install -y docker-ce$pkg_version docker-ce-cli docker-buildx-plugin docker-compose-plugin


            else
                if [ -f "/etc/os-release" ]; then
                    cat /etc/os-release
                fi
                echo "ERROR: not support this system yum auto install docker"
            fi
        fi

        if [[ "$install_docker" != "1" ]]; then
        # 1 When docker is installed using other commands
            $pkg_manager makecache

            $pkg_manager install -y docker-ce$pkg_version docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin $conflicting

        fi
        ;;

    *)
        if [ -z "$lsb_dist" ]; then
            if Is_Darwin; then
                echo
                echo "ERROR: Unsupported operating system 'macOS'"
                echo
                exit 1
            fi
        fi
        echo
        Red_Error "ERROR: not support '$lsb_dist' this system auto install docker"
        ;;
    esac

    if ! Command_Exists docker && [ ! -f /usr/bin/docker ]; then
        echo "Trying to install docker using binary package method."
        
        Binary_package_install_docker
        
    fi

    if ! Command_Exists docker ; then
        Red_Error "docker command not exists,install docker failed,please check the reason."     
    fi
}


Binary_package_install_docker(){
# Install docker in binary package
    DOCKER_VERSION="27.3.1"
    SYS_TYPE=$(uname -m)

    echo "Create docker group"
    groupadd docker

    echo "Downloading docker ${DOCKER_VERSION}.tgz file, please wait..."
    wget --no-check-certificate -O docker-${DOCKER_VERSION}.tgz ${download_Url}/src/docker-${DOCKER_VERSION}-${SYS_TYPE}.tgz -t 5 -T 10

    tmp_size=$(du -b docker-${DOCKER_VERSION}.tgz |awk '{print $1}')
    if [ $tmp_size -lt 60831072 ];then
        ls -l docker-${DOCKER_VERSION}.tgz
        Red_Error "ERROR: Failed to download, please try install again!"
    fi

    echo "Unzip the docker ${DOCKER_VERSION}.tgz file, please wait..."
    tar -xvf docker-${DOCKER_VERSION}.tgz
    \cp docker/* /usr/bin/

    # rm -rf docker
    rm -f docker-${DOCKER_VERSION}.tgz

    wget --no-check-certificate -O /usr/bin/docker-compose ${download_Url}/src/docker-compose-linux-${SYS_TYPE} -t 5 -T 20
    chmod +x /usr/bin/docker-compose

    wget --no-check-certificate -O /usr/lib/systemd/system/docker.service ${download_Url}/init/systemd/docker.service -t 5 -T 20
    wget --no-check-certificate -O /usr/lib/systemd/system/docker.socket ${download_Url}/init/systemd/docker.socket -t 5 -T 20
    wget --no-check-certificate -O /usr/lib/systemd/system/containerd.service ${download_Url}/init/systemd/containerd.service -t 5 -T 20

    systemctl enable docker.service
    systemctl enable docker.socket
    systemctl enable containerd.service
    
    systemctl start containerd.service
    systemctl start docker.socket
    systemctl start docker.service

}

Docker_Compose_Check(){

    is_docker_compose="0"
    if Command_Exists docker-compose; then
        is_docker_compose="1"
        DOCKER_COMPOSE="docker-compose"
    else 
        if Command_Exists docker; then
            Docker_compose="docker compose version"
            if $Docker_compose; then
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
        Docker_Install
    fi

    Compose_Download_Url="$download_Url/install/src/docker-compose-$(uname -s)-$(uname -m)"
    Compose_Path="/usr/local/bin/docker-compose"
    New_Compose_Path="/usr/libexec/docker/cli-plugins/docker-compose"
    Compose_lin="/usr/bin/docker-compose"

    if [ ! -f $New_Compose_Path ]; then
        wget --no-check-certificate -t 5 -T 20 -O $New_Compose_Path $Compose_Download_Url
        chmod +x $New_Compose_Path
        rm -rf $Compose_lin
    fi

    if [ ! -f ${Compose_lin} ];then
        ln -s $New_Compose_Path $Compose_lin
    fi
}


Set_Firewall() {
    sshPort=$(cat /etc/ssh/sshd_config | grep 'Port ' | awk '{print $2}')
    if [ "${PM}" = "apt-get" ]; then
        if [ ! -f "/usr/bin/ufw" ] && [ ! -f "/bin/ufw" ] && ! Command_Exists "ufw"; then
            apt-get install -y ufw
        fi
        if [ -f "/usr/sbin/ufw" ]; then
            ufw allow 20/tcp >/dev/null 2>&1
            ufw allow 21/tcp >/dev/null 2>&1
            ufw allow 22/tcp >/dev/null 2>&1
            ufw allow 80/tcp >/dev/null 2>&1
            ufw allow 443/tcp >/dev/null 2>&1
            ufw allow 888/tcp >/dev/null 2>&1
            ufw allow 39000:40000/tcp >/dev/null 2>&1
            ufw allow ${SMTP_PORT}/tcp >/dev/null 2>&1
            ufw allow ${SMTPS_PORT}/tcp >/dev/null 2>&1
            ufw allow ${SUBMISSION_PORT}/tcp >/dev/null 2>&1
            ufw allow ${IMAP_PORT}/tcp >/dev/null 2>&1
            ufw allow ${IMAPS_PORT}/tcp >/dev/null 2>&1
            ufw allow ${POP_PORT}/tcp >/dev/null 2>&1
            ufw allow ${POPS_PORT}/tcp >/dev/null 2>&1
            ufw allow ${sshPort}/tcp >/dev/null 2>&1
            ufw_status=$(ufw status)
            echo y | ufw enable
            ufw default deny
            ufw reload
        fi
    else
        if [ -f "/etc/init.d/iptables" ]; then
            iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 20 -j ACCEPT
            iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 21 -j ACCEPT
            iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 22 -j ACCEPT
            iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 80 -j ACCEPT
            iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 443 -j ACCEPT
            iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport ${SMTP_PORT} -j ACCEPT
            iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport ${SMTPS_PORT} -j ACCEPT
            iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport ${SUBMISSION_PORT} -j ACCEPT
            iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport ${IMAP_PORT} -j ACCEPT
            iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport ${IMAPS_PORT} -j ACCEPT
            iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport ${POP_PORT} -j ACCEPT
            iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport ${POPS_PORT} -j ACCEPT
            iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport ${sshPort} -j ACCEPT
            iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 39000:40000 -j ACCEPT
            #iptables -I INPUT -p tcp -m state --state NEW -m udp --dport 39000:40000 -j ACCEPT
            iptables -A INPUT -p icmp --icmp-type any -j ACCEPT
            iptables -A INPUT -s localhost -d localhost -j ACCEPT
            iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
            iptables -P INPUT DROP
            service iptables save
            sed -i "s#IPTABLES_MODULES=\"\"#IPTABLES_MODULES=\"ip_conntrack_netbios_ns ip_conntrack_ftp ip_nat_ftp\"#" /etc/sysconfig/iptables-config
            iptables_status=$(service iptables status | grep 'not running')
            if [ "${iptables_status}" == '' ]; then
                service iptables restart
            fi
        else
            AliyunCheck=$(cat /etc/redhat-release | grep "Aliyun Linux")
            [ "${AliyunCheck}" ] && return
            if [ ! -f "/usr/bin/firewall-cmd" ] && [ ! -f "/bin/firewall-cmd" ] && ! Command_Exists "firewall-cmd"; then
                yum install firewalld -y
            fi
            [ "${Centos8Check}" ] && yum reinstall python3-six -y
            systemctl enable firewalld
            systemctl start firewalld
            firewall-cmd --set-default-zone=public >/dev/null 2>&1
            firewall-cmd --permanent --zone=public --add-port=20/tcp >/dev/null 2>&1
            firewall-cmd --permanent --zone=public --add-port=21/tcp >/dev/null 2>&1
            firewall-cmd --permanent --zone=public --add-port=22/tcp >/dev/null 2>&1
            firewall-cmd --permanent --zone=public --add-port=80/tcp >/dev/null 2>&1
            firewall-cmd --permanent --zone=public --add-port=443/tcp > /dev/null 2>&1
            firewall-cmd --permanent --zone=public --add-port=${SMTP_PORT}/tcp >/dev/null 2>&1
            firewall-cmd --permanent --zone=public --add-port=${SMTPS_PORT}/tcp >/dev/null 2>&1
            firewall-cmd --permanent --zone=public --add-port=${SUBMISSION_PORT}/tcp >/dev/null 2>&1
            firewall-cmd --permanent --zone=public --add-port=${IMAP_PORT}/tcp >/dev/null 2>&1
            firewall-cmd --permanent --zone=public --add-port=${IMAPS_PORT}/tcp >/dev/null 2>&1
            firewall-cmd --permanent --zone=public --add-port=${POP_PORT}/tcp >/dev/null 2>&1
            firewall-cmd --permanent --zone=public --add-port=${POPS_PORT}/tcp >/dev/null 2>&1
            firewall-cmd --permanent --zone=public --add-port=${sshPort}/tcp >/dev/null 2>&1
            firewall-cmd --permanent --zone=public --add-port=39000-40000/tcp >/dev/null 2>&1
            #firewall-cmd --permanent --zone=public --add-port=39000-40000/udp > /dev/null 2>&1
            firewall-cmd --reload
        fi
    fi
}


Check_apt_status(){

    MAX_RETRIES=30
    retries=0

    while [ $retries -lt $MAX_RETRIES ]; do
        output=$(ps aux| grep -E '(apt|apt-get)\s' 2>&1)
        check_output=$(echo "$output" | grep -E '(apt|apt-get)\s')
        
        # If check_output is empty, terminate the loop
        if [ -z "$check_output" ]; then
            break
        fi

        retries=$((retries + 1))

        echo "apt-get is in use, it will automatically exit after ${retries}/${MAX_RETRIES} times, try again after 10 seconds..."
        echo "$check_output"
        wait_msg="Please wait "
        wait_time="10"
        Bored_waiting
        
    done

    if [ $retries -ge $MAX_RETRIES ]; then
        echo "$output" > /tmp/waf_install_log.txt
        Red_Error "ERROR: apt-get command exceeds the maximum wait times: ${MAX_RETRIES}. Do not use the apt/apt-get command or install software during installation, please try to reinstall!"
    fi

}


Check_Connect_PgSql(){

    MAX_RETRIES=10
    retries=0

    while [ $retries -lt $MAX_RETRIES ]; do
        output=$(docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "SELECT 1;" 2>&1)
        if [ $? -eq 0 ]; then
            PgSql_run="1"
            break
            # When equal to 0, exit the loop
        fi

        retries=$((retries + 1))

        echo "PgSql failed to connect, try again after 10 seconds, ${retries}/${MAX_RETRIES}..."

        if [[ "${retries}" == "1" ]] || [[ "${retries}" == "${MAX_RETRIES}" ]]; then
        # Only the first and last print information
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

Init_Billionmail()
{
    # SQL file path
    SQL_FILE="./init.sql"
    if [ ! -f "${SQL_FILE}" ]; then
        Red_Error "SQL The file does not exist: ${SQL_FILE}"
    fi

    if docker ps --format '{{.Names}}' | grep -q "^${PGSQL_CONTAINER_NAME}$"; then
        
        # Test whether it is possible to connect
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
            # Create a domain
            docker exec -i -e PGPASSWORD=${DBPASS} ${PGSQL_CONTAINER_NAME} psql -U ${DBUSER} -d ${DBNAME} -c "INSERT INTO domain (domain, a_record, mailboxes, mailbox_quota, quota, rate_limit, create_time, active)
            VALUES ('${BILLIONMAIL_HOSTNAME}', 'mail.${BILLIONMAIL_HOSTNAME}', 500, 5368709120, 5368709120, 12, ${create_time}, 1);"
            if [ $? -eq 0 ]; then
                echo "Domain creation was successful!"
            else
                Red_Error "Domain creation failed!"
            fi
        else
            echo ""${Check_domain}" Domain already exists!"
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

    else
        Red_Error "PgSql container does not exist!"
    fi

}

Billionmail(){
    cat << EOF > billionmail.conf

# BILLIONMAIL_HOSTNAME configuration
BILLIONMAIL_HOSTNAME=${BILLIONMAIL_HOSTNAME}

# SQL NAME and USER and PASSWORD configuration

DBNAME=${DBNAME}
DBUSER=${DBUSER}

# Generate random passwords  
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
REDIS_PORT=${SMTP_REDIS_PORTPORT}
SQL_PORT=${SQL_PORT}


# You can use this script to set the time zone for your container.
# See https://en.wikipedia.org/wiki/List_of_tz_database_time_zones for a list of timezones"
# echo -e "Use a column named "TZ identifier" + note the column named "Notes""

TZ=${BILLIONMAIL_TIME_ZONE}


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
    cp -d -n ${SSL_path}/* ssl/

    echo -e "Execute ${DOCKER_COMPOSE} up -d"
    if  [ ! -s "docker-compose.yml" ]; then
        ls -al
        Red_Error "docker-compose.yml not found."
    fi
    ${DOCKER_COMPOSE} up -d
    if [ $? -eq 0 ]; then
        echo -e "Billionmail installation completed successfully!"
    else
        Red_Error "Billionmail installation failed, please try reinstalling!"
    fi

    echo -e "Initialize the data..."
    sleep 5
    Init_Billionmail

}


Install_Main(){
    Check_Port
    LXC_install_tips

    Get_Pack_Manager
    if [ "$PM" == "yum" ]; then
        Install_RPM_Pack
    else
        Install_Deb_Pack
    fi

    set +e
    Docker_Check
    Docker_Compose_Check
    
    if [ "$is_docker" != "1" ]; then
        Docker_Download_Url_Check
        Docker_Install
        Docker_Compose_Check
    fi

    if [ "$is_docker_compose" != "1" ]; then
        Docker_Compose_Install
        Docker_Compose_Check
        if [ "$is_docker_compose" != "1" ]; then
            Red_Error "ERROR: Docker Compose installation failed, Please reinstall Docker Compose!"
        fi
    fi
   
    Docker_Start

    Billionmail
    
    Set_Firewall
    
}


Install_Main

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
echo -e " Type | Host record       |    IPv4 address      |"
echo -e "  A   | mail.${BILLIONMAIL_HOSTNAME} | ${IPV4_ADDRESS} |"
echo -e "==========================================================="
echo -e " Type | Host record |    Record value    |     MX priority"
echo -e "  MX  |     @       | mail.${BILLIONMAIL_HOSTNAME} |  10"
echo -e "==========================================================="

echo -e "\e[31mPlease save the following information:\e[0m"
echo -e "Mailbox (e-mail): \e[33m${mailbox}@${BILLIONMAIL_HOSTNAME}\e[0m \nPassword: \e[33m${Generate_mailbox_password}\e[0m"

echo -e "Tip: Use \e[33mbash mail_users.sh\e[0m to add and delete Domain and Mailboxes"

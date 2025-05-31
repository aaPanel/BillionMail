#!/bin/bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin
export PATH

CONTAINER_PROJECT_NAME=billionmail
PGSQL_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-pgsql-billionmail-1"
DOVECOT_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-dovecot-billionmail-1"
POSTFIX_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-postfix-billionmail-1"
RSPAMD_CONTAINER_NAME="${CONTAINER_PROJECT_NAME}-rspamd-billionmail-1"
TIME=$(date +%Y_%m_%d_%H_%M_%S)

PWD_d=`pwd`

if [ $(whoami) != "root" ];then
    echo -e "Non-root install, please try the following solutions: \n   1.Please switch to root user install \n   2.Try executing the following install commands: \n     sudo bash $0 $@"
    exit 1
fi

if [ ! -d ".git" ]; then
    ls -al
    echo "The .git does not exist, Cannot continue operation, Please operate in the project directory."
    exit 1
fi

if [ ! -s ".env" ]; then
    ls -al
    echo " The .env file does not exist. Please check whether BillionMail is installed? Please operate in the project directory."
    exit 1
fi

. .env
if [ -z "${DBUSER}" ] || [ -z "${DBNAME}" ] || [ -z "${DBPASS}" ]; then
    echo "The obtained .env configuration exception is empty."   
    exit 1
fi


while [ ${#} -gt 0 ]; do
    case $1 in
        -h|--help)
            echo "Usage:  [options]"
            echo "Options:"
            echo "  --ours       Use merge strategy option "ours" to resolve conflicts in favor of non-BillionMail code (local changes override remote changes)."
            exit 0
            ;;
        --ours)
            GIT_MERGE_STRATEGY=ours
            shift 1
            ;;

    esac
    shift 1
done


Command_Exists() {
    command -v "$@" >/dev/null 2>&1
}

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
    echo -e "Please screenshot the above error message and post to the https://github.com/aaPanel/BillionMail/issues for help"

}


Red_Error(){
    echo '=================================================';
    printf '\033[1;31;40m%b\033[0m\n' "$@";
    GetSysInfo
    exit 1;
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
                systemctl restart docker.socket
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

Docker_Status_Check(){
    if Command_Exists docker ; then
        if Command_Exists systemctl; then
            docker_status=$(systemctl is-active docker)
            # Determine the running status of docker
            if [ "$docker_status" = "active" ]; then
                echo -e "\033[32m Docker is running \033[0m"
            else
                echo -e "\033[31m Docker Not running, starting... \033[0m"
                Docker_Start
            fi
        fi
    else
        Red_Error "The docker command does not exist, please check whether docker is installed."
    fi
}

Update_BillionMail(){

    if ! Command_Exists git ; then
        Red_Error "ERROR: git Command does not exist"
    fi

    BRANCH="main"

    if [ -f "update.sh" ]; then
        echo -e "Checking for update.sh script..."
        MD5_1="$(md5sum update.sh)"
        git fetch origin
        git checkout "origin/${BRANCH}" update.sh
        MD5_2=$(md5sum update.sh)
        if [[ "${MD5_1}" != "${MD5_2}" ]]; then
            echo -e "\033[33m update.sh is changed, please run update.sh script again.\033[0m"
            chmod +x update.sh
            exit 1
        fi
    fi

    echo -e "Stop BillionMail, please wait..."
    sleep 3
    ${DOCKER_COMPOSE} down

    
    echo -e "\033[34m🚀 Committing current changes...\033[0m"
    # Set user.name and user.email if they are not set
    if [[ -z "$(git config user.name)" ]]; then
        git config user.name BillionMail
    fi
    if [[ -z "$(git config user.email)" ]]; then
        git config user.email BillionMail@BillionMail.com
    fi
    # Commit current changes
    git add -u
    git commit -am "Before update on ${TIME}"  > /dev/null

    # Fetch remote updates
    echo -e "\033[34m🔍 Fetching remote updates...\033[0m"
    git fetch origin

    # Merge changes
    echo -e "\033[34m🔄 Merging changes (strategy: ${GIT_MERGE_STRATEGY:-theirs}, options: patience)...\033[0m"
    git config merge.defaultToUpstream true
    git merge -X"${GIT_MERGE_STRATEGY:-theirs}" -Xpatience -m "After update on ${TIME}"

    GIT_MERGE_RETURN=$?
    case ${GIT_MERGE_RETURN} in
        128)
            echo -e "\033[31m❌ Critical Error: Detected local files conflicting with upstream repository\033[0m"
            echo -e "Please move these files to another location before updating:"
            git status
            exit 1
            ;;
        1)
            echo -e "\033[33m⚠️ Detected file conflicts, attempting automatic resolution...\033[0m"
            git status --porcelain | grep -E "UD|DU" | awk '{print $2}' | xargs rm -v
            git add -A
            git commit -m "After update on ${TIME}"  > /dev/null
            git checkout . 
            echo -e "\033[33m✅ Conflicts resolved automatically, If necessary, delete and recreate the file.\033[0m"
            ;;
        0)
            echo -e "\033[32m✅ Code update completed successfully.\033[0m"
            ;;
        *)
            echo -e "\033[31m❌ Unknown merge error encountered (code: ${GIT_MERGE_RETURN})\033[0m"
            echo -e "Please check the error message above and try again after resolving issues."
            exit 1
            ;;
    esac

    if  [ ! -s "docker-compose.yml" ]; then
        ls -al
        Red_Error "docker-compose.yml not found."
    fi

    echo -e "Getting the latest image, please wait..."
    ${DOCKER_COMPOSE} pull

    echo -e "Starting BillionMail, please wait..."
    ${DOCKER_COMPOSE} up -d
    if [ $? -eq 0 ]; then
        echo -e "\033[32m✅ Started successfully, update completed.\033[0m"
    else
        echo -e "\033[33m⚠️ Update completed, please check container status. If startup failed, check container logs.\033[0m"
    fi

    [ ! -d "/opt" ] && mkdir /opt
    echo "${PWD_d}" > /opt/PWD-Billion-Mail.txt
    ln -sf ${PWD_d}/bm.sh /usr/bin/bm
    chmod +x ${PWD_d}/bm.sh

}

echo "
+-----------------------------------------------------------------------------
| You are updating BillionMail. This operation will pull the latest code from 
| GitHub's main branch and rebuild the containers.
|
| If you have made any modifications, please back them up first. 
| This operation may overwrite your changes.
+-----------------------------------------------------------------------------
"

while [ "$go" != 'y' ] && [ "$go" != 'n' ]; do
	read -p "Are you sure want to update BillionMail. All containers will be stopped? (y/n): " go;
done

if [ "$go" == 'n' ];then
  exit;
fi

Docker_Compose_Check
    
Docker_Status_Check

Update_BillionMail
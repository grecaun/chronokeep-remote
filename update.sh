#! /bin/bash

DEST=/chronokeep/remote/
SERVICE_NAME=chronokeep-remote
REPO_URL='https://api.github.com/repos/grecaun/chronokeep-remote/releases/latest'
UPDATE_SCRIPT_URL='https://raw.githubusercontent.com/grecaun/chronokeep-remote/main/update.sh'

VERSION=1

echo "------------------------------------------------"
echo "------------- Now updating Remote! -------------"
echo "------------------------------------------------"
echo "------------ Checking update script ------------"
echo "------------------------------------------------"
if ! [[ -e ${DEST}update.sh ]]; then
    echo "----------- Fetching update script. ------------"
    echo "------------------------------------------------"
    curl -L ${UPDATE_SCRIPT_URL} -o ${DEST}update.sh > /dev/null 2>&1
    sudo chown $USER:root ${DEST}update.sh
    sudo chmod +x ${DEST}update.sh
    echo "------- Please re-run the updated script -------"
    echo "------------------------------------------------"
    exit 1
else
    OLD_SCRIPT_VERS=`cat ${DEST}update.sh | grep ^VERSION= | sed 's/VERSION=//'`
    if [[ $OLD_SCRIPT_VERS -ge 0 ]]; then
        curl -L ${UPDATE_SCRIPT_URL} -o ${DEST}update_tmp.sh > /dev/null 2>&1
        NEW_SCRIPT_VERS=`cat ${DEST}update_tmp.sh | grep ^VERSION= | sed 's/VERSION=//'`
        if [[ $NEW_SCRIPT_VERS -gt $OLD_SCRIPT_VERS ]]; then
            echo "----------- Updating update script. ------------"
            echo "------------------------------------------------"
            mv ${DEST}update_tmp.sh ${DEST}update.sh
            sudo chmod +x ${DEST}update.sh
            echo "------- Please re-run the updated script -------"
            echo "------------------------------------------------"
            exit 1
        else
            rm ${DEST}update_tmp.sh
        fi;
    else
        echo "----------- Updating update script. ------------"
        echo "------------------------------------------------"
        curl -L ${UPDATE_SCRIPT_URL} -o ${DEST}update.sh > /dev/null 2>&1
        sudo chown $USER:root ${DEST}update.sh
        sudo chmod +x ${DEST}update.sh
        echo "------- Please re-run the updated script -------"
        echo "------------------------------------------------"
        exit 1
    fi;
fi;
echo "---- Checking latest Remote release version ----"
echo "------------------------------------------------"
LATEST_VERSION=`curl ${REPO_URL} 2>&1 | grep tag_name | sed -e "s/[\":,]//g" | sed -e "s/tag_name//" | sed -e "s/v//"`
CURRENT_VERSION=`cat ${DEST}version.txt | sed -e "s/v//"`
echo Latest Remote version is ${LATEST_VERSION}
echo "------------------------------------------------"
echo Current Remote version is ${CURRENT_VERSION}
echo "------------------------------------------------"
echo Latest version is ${LATEST_VERSION} - current version is ${CURRENT_VERSION}.
LATEST_VERSION_MAJOR=`echo ${LATEST_VERSION} | cut -d '.' -f 1`
LATEST_VERSION_MINOR=`echo ${LATEST_VERSION} | cut -d '.' -f 2`
LATEST_VERSION_PATCH=`echo ${LATEST_VERSION} | cut -d '.' -f 3`
CURRENT_VERSION_MAJOR=`echo ${CURRENT_VERSION} | cut -d '.' -f 1`
CURRENT_VERSION_MINOR=`echo ${CURRENT_VERSION} | cut -d '.' -f 2`
CURRENT_VERSION_PATCH=`echo ${CURRENT_VERSION} | cut -d '.' -f 3`
# If the latest version has a higher major version, update.
if [[ ${LATEST_VERSION_MAJOR} -gt ${CURRENT_VERSION_MAJOR} ]] ||
        [[ ${LATEST_VERSION_MAJOR} -eq ${CURRENT_VERSION_MAJOR} && ${LATEST_VERSION_MINOR} -gt ${CURRENT_VERSION_MINOR} ]] ||
        [[ ${LATEST_VERSION_MAJOR} -eq ${CURRENT_VERSION_MAJOR} && ${LATEST_VERSION_MINOR} -eq ${CURRENT_VERSION_MINOR} && ${LATEST_VERSION_PATCH} -gt ${CURRENT_VERSION_PATCH} ]]; then
        echo "---- New version found! Updating Remote now ----"
        echo "------------------------------------------------"
        download_url=`curl ${REPO_URL} 2>&1 | grep browser_download_url | sed -e "s/[\",]//g" | sed -e "s/browser_download_url://"`
        curl -L ${download_url} -o ${DEST}release.tar.gz
        gunzip ${DEST}release.tar.gz
        tar -xvf ${DEST}release.tar -C ${DEST}
        rm ${DEST}release.tar
        systemctl restart ${SERVICE_NAME}
        echo "------------ Remote update complete. -----------"
        echo "------------------------------------------------"
else
        echo "---------- Remote already up to date. ----------"
        echo "------------------------------------------------"
fi
echo "------------- Update is finished! --------------"
echo "------------------------------------------------"
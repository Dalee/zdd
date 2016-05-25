#!/usr/bin/env bash

#
# provision virtual machine with ansible
#
# Usage:
# bootstrap.sh /home/project /home/project/src/zdd
#

export DEBIAN_FRONTEND=noninteractive
export PYTHONUNBUFFERED=true
export ANSIBLE_FORCE_COLOR=true

PROJECT_GOPATH=$1
PROJECT_ROOT=$2

ANSIBLE="/usr/bin/ansible-playbook"
if [ ! -f "$ANSIBLE" ]; then
    echo "Installing ansible.."
	apt-get -qq -y update
	apt-get -qq -y install make software-properties-common > /dev/null 2>&1
	apt-add-repository -y ppa:ansible/ansible > /dev/null 2>&1
	apt-get -qq -y update
	apt-get -qq -y install ansible > /dev/null 2>&1
fi

echo "Starting provision"
${ANSIBLE} -c local \
        -e "project_gopath=${PROJECT_GOPATH}" \
		-e "project_root=${PROJECT_ROOT}" \
		${PROJECT_ROOT}/ansible/vagrant.yml \
		-i ${PROJECT_ROOT}/ansible/inventory.ini

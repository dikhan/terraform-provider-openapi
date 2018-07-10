#!/usr/bin/env bash

#
# Installation script to download and install terraform-provider-openapi
#
# Usage:
#  $ ./install --provider-name <provider-name>
# * provider-name: this is the value that will be used to name the terraform provider upon installation, e,g: terraform-provider-<provider-name>

# variables storing install script location and name
_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
_FILE="${_DIR}/$(basename "${BASH_SOURCE[0]}")"
_BASE="$(basename ${_FILE})"

# cleanup removes temporarily assets created during the installation process
function cleanup() {
	echo "[INFO] Cleaning up tmp dir created for installation purposes: ${TMP_INSTALL_DIR}"
	rm -rf ${TMP_INSTALL_DIR}
}

# usage prints script usage
function usage(){
    echo "usage: ${_BASE} --provider-name <provider-name>"
}

if [ "$#" != "2" ]; then
	echo "[ERROR] required argument --provider-name missing"
	usage
	exit 1
fi

# process input arguments
while [ $# -gt 0 ]; do
    case $1 in
        --help | -h)
            usage
            exit 0
            ;;
        --provider-name | -p)
            shift
            PROVIDER_NAME=$1
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 2
    esac
    shift
done

if [ "$PROVIDER_NAME" == "" ]; then
	echo "[ERROR] required argument --provider-name missing value"
	usage
	exit 1
fi

# used to determine which architecture to install
ARCH=$(uname)

# installation variables
LATEST_RELEASE_VERSION=0.1.1

TF_PROVIDER_BASE_NAME="terraform-provider-"
TF_OPENAPI_PROVIDER_PLUGIN_NAME="${TF_PROVIDER_BASE_NAME}openapi"
TF_PROVIDER_PLUGIN_NAME="${TF_PROVIDER_BASE_NAME}${PROVIDER_NAME}"

INSTALLATION_DIR="$HOME/.terraform.d/plugins"
TMP_INSTALL_DIR=$(mktemp -d)
if [ "$?" != "0" ]; then
	echo "[ERROR] failed to create temporary directory."
	exit 1
fi

if [ ! -d "${INSTALLATION_DIR}" ]; then
    echo "[WARN] Terraform plugin directory does not exist: ${INSTALLATION_DIR}, attempting to create it...\n"
    if ! mkdir -p "${INSTALLATION_DIR}" > /dev/null 2>&1; then
        echo "[ERROR] unable to create ${INSTALLATION_DIR}..."
        exit 2
    fi
fi

# determine which binary to fetch depending upon current architecture
if [ "${ARCH}" == "Linux" ]; then
	BIN_ARCH="linux"
elif [ "${ARCH}" == "Darwin" ]; then
	BIN_ARCH="darwin"
else
	echo "[ERROR] unsupported architecture: ${ARCH}, only architecture supported at the moment are Darwin and Linux"
	exit 2
fi

FILE_NAME="${TF_OPENAPI_PROVIDER_PLUGIN_NAME}_${LATEST_RELEASE_VERSION}_${BIN_ARCH}_amd64.tar.gz"
DOWNLOAD_URL="https://github.com/dikhan/${TF_OPENAPI_PROVIDER_PLUGIN_NAME}/releases/download/v${LATEST_RELEASE_VERSION}/${FILE_NAME}"

# verifying curl is available in the system
# Curl
if ! hash curl 2>/dev/null; then
	echo "[ERROR] curl not available on this system, please install it."
	exit 1
fi

# fetch terraform-provider-openapi plugin and dump it into the temp folder
echo "[INFO] Downloading ${DOWNLOAD_URL} in temporally folder ${TMP_INSTALL_DIR}..."
if ! curl -L --silent ${DOWNLOAD_URL} --output ${TMP_INSTALL_DIR}/${FILE_NAME}
then
	echo "[ERROR] failed to download ${DOWNLOAD_URL}"
	cleanup
	exit 1
fi

# extract contents of tar file downloaded in the temp folder
echo "[INFO] Extracting ${TF_OPENAPI_PROVIDER_PLUGIN_NAME} from ${FILE_NAME}..."
$(cd $TMP_INSTALL_DIR && tar -xz -f $FILE_NAME)
if [ "$?" != "0" ]; then
	echo "[ERROR] failed to extract downloaded tar file ${TMP_INSTALL_DIR}/${FILE_NAME}"
	cleanup
	exit 1
fi

# check we have write permissions on $INSTALLATION_DIR
if [ -w ${INSTALLATION_DIR} ]; then
  if ! mv "${TMP_INSTALL_DIR}/${TF_OPENAPI_PROVIDER_PLUGIN_NAME}" "${INSTALLATION_DIR}/${TF_OPENAPI_PROVIDER_PLUGIN_NAME}"; then
	echo "[ERROR] failed to move '${TF_OPENAPI_PROVIDER_PLUGIN_NAME}' binary to ${INSTALLATION_DIR}"
	cleanup
	exit 1
  fi

  if ! ln -sF "${INSTALLATION_DIR}/${TF_OPENAPI_PROVIDER_PLUGIN_NAME}" "${INSTALLATION_DIR}/${TF_PROVIDER_PLUGIN_NAME}"; then
	echo "[ERROR] failed to create symlink to '${INSTALLATION_DIR}/${TF_OPENAPI_PROVIDER_PLUGIN_NAME}' for ${TF_PROVIDER_PLUGIN_NAME}"
	cleanup
	exit 1
  fi


else
    echo "[ERROR] unable to write to ${INSTALLATION_DIR} due to lack of write permissions"
    cleanup
    exit 1
fi

cleanup
echo "[INFO] Terraform provider '${TF_PROVIDER_PLUGIN_NAME}' successfully installed at: '${INSTALLATION_DIR}'!"
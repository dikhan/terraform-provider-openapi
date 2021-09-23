#!/usr/bin/env bash

# Installation script to download and install the OpenAPI Terraform provider based on the arguments provided and following the
# expected Terraform plugin installation paths depending the Terraform version is installed.
#
# - If Terraform v0.12 is detected, the plugin will be installed following Terraform's v0.12 installation instructions: https://www.terraform.io/docs/plugins/basics.html#installing-plugins
# - If Terraform v0.13 or greater is detected, the plugin will be installed following Terraform's v0.13 installation instructions: https://www.terraform.io/docs/configuration/provider-requirements.html#in-house-providers
#
# Terraform < v0.12 is not supported.
#
# Usage:
#  $ ./install --provider-name [name] --provider-source-address [source-address]
# * --provider-name: provider's name which will be used to name the plugin binary installed, e,g: terraform-provider-<provider-name>
# * --provider-source-address: provider source address in the form of <HOSTNAME>/<NAMESPACE>. Default value is 'terraform.example.com/examplecorp'
# * --debug: sets the logging level to debug mode. Default logging level is error.
# Example:
# $ ./install.sh --provider-name myprovider --provider-source-address "terraform.example.com/examplecorp"

# variables storing install script location and name
_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
_FILE="${_DIR}/$(basename "${BASH_SOURCE[0]}")"
_BASE="$(basename ${_FILE})"

INSTALLATION_DIR="${HOME}/.terraform.d/plugins"
TF_PROVIDER_BASE_NAME="terraform-provider-"
PROVIDER_SOURCE_ADDRESS="terraform.example.com/examplecorp"

colrst='\033[0m'    # Text Reset

verbosity=1
### verbosity levels
silent_lvl=0
err_lvl=1
inf_lvl=2
dbg_lvl=3

## esilent prints output even in silent mode
function esilent () { verb_lvl=$silent_lvl elog "$@" ;}
function einfo ()  { verb_lvl=$inf_lvl elog "\033[0;37m[INFO]${colrst} $@" ;} # White
function edebug () { verb_lvl=$dbg_lvl elog "\033[0;36m[DEBUG]${colrst} $@" ;} # Cyan
function eerror () { verb_lvl=$err_lvl elog "\033[0;31m[ERROR]${colrst} $@" ;} # Red
function elog() {
  if [ $verbosity -ge $verb_lvl ]; then
    if [ $silent_lvl -eq $verb_lvl ]; then
      echo -e "$@"
    else
      datestring=`date +"%Y-%m-%d %H:%M:%S"`
      echo -e "$datestring $@"
    fi
  fi
}

# cleanup removes temporarily assets created during the installation process
function cleanup() {
  if [ $TMP_INSTALL_DIR ]
  then
    edebug "Cleaning up tmp dir created for installation purposes: ${TMP_INSTALL_DIR}"
    rm -rf ${TMP_INSTALL_DIR}
  fi
}

# usage prints script usage
function usage(){
    esilent ""
    esilent "Usage: ${_BASE} [options]"
    esilent " "
    esilent "Installs the OpenAPI Terraform provider based on the input configuration provided and the Terraform version used to install the provider in the corresponding installation path:"
    esilent "- For Terraform 0.12, the installation path will be: ${INSTALLATION_DIR}"
    esilent "- For Terraform > 0.12, the installation path will be: ${INSTALLATION_DIR}/PROVIDER_SOURCE_ADDR/PROVIDER_VERISON/OS_ARCH/terraform-provider-PROVIDER_NAME"
    esilent " "
    esilent "options:"
    esilent "-h, --help                                            show help"
    esilent "-p, --provider-name=PROVIDER_NAME                     [required] specify the provider name. The plugin installed will be named like terraform-provider-NAME (all lower case)"
    esilent "-s, --provider-source-address=PROVIDER_SOURCE_ADDR    specify the provider source address <HOSTNAME>/<NAMESPACE>. Default value ${PROVIDER_SOURCE_ADDRESS}"
    esilent "-d, --debug                                           sets the logging level to debug mode. Default logging level is error."
}

function determineOSAndArch(){
  # used to determine which os and architecture to install
  XC_OS=$(uname)
  XC_ARCH=${XC_ARCH:-"amd64"}

  # determine which binary to fetch depending upon current architecture
  if [ "${XC_OS}" == "Linux" ]; then
    XC_OS="linux"
  elif [ "${XC_OS}" == "Darwin" ]; then
    XC_OS="darwin"
  else
    eerror "Unsupported architecture: ${XC_OS}, only architecture supported at the moment are Darwin and Linux"
    exit 2
  fi
}

function setupInstallationPath(){
  local PROVIDER_NAME=$1
  local PROVIDER_VERSION=$2
  local XC_OS=$3
  local XC_ARCH=$4
  TF_VERSION=$(terraform version | sed 's/[[:alpha:]||[:space:]]//g' | awk 'NR == 1')
  if [ -z $TF_VERSION ]
  then
    eerror "Terraform not detected, please install Terraform before running this script."
    exit 1
  else
    einfo "Detected Terraform v$TF_VERSION"
  fi

  TF_PATCH_VERSION=$(echo $TF_VERSION | cut -f3 -d.)
  TF_MINOR_VERSION=$(echo $TF_VERSION | cut -f2 -d.)
  TF_MAJOR_VERSION=$(echo $TF_VERSION | cut -f1 -d.)

  if [ $TF_MINOR_VERSION -lt 12 ] && [ $TF_MAJOR_VERSION -eq 0 ]
  then
    eerror "OpenAPI Terraform provider no longer supports versions of Terraform less than v0.12.0"
    exit 1
  elif [ $TF_MINOR_VERSION -eq 12 ] && [ $TF_MAJOR_VERSION -eq 0 ]
  then
    einfo "Installing provider based on Terraform v0.12.* plugin installation instructions: https://www.terraform.io/docs/plugins/basics.html#installing-plugins"
    TF_PROVIDER_PLUGIN_NAME="${TF_PROVIDER_BASE_NAME}${PROVIDER_NAME}_v${PLUGIN_VERSION}"
  else
  then
    einfo "Installing provider based on Terraform >= v0.13.* plugin installation instructions: https://www.terraform.io/docs/configuration/provider-requirements.html#in-house-providers"
    INSTALLATION_DIR="${INSTALLATION_DIR}/${PROVIDER_SOURCE_ADDRESS}/${PROVIDER_NAME}/${PROVIDER_VERSION}/${XC_OS}_${XC_ARCH}"
    mkdir -p ${INSTALLATION_DIR}
    TF_PROVIDER_PLUGIN_NAME="${TF_PROVIDER_BASE_NAME}${PROVIDER_NAME}"
  fi

  if [ ! -d "${INSTALLATION_DIR}" ]; then
    edebug "[WARN] Terraform installation plugin directory does not exist: ${INSTALLATION_DIR}, attempting to create it..."
    if ! mkdir -p "${INSTALLATION_DIR}" > /dev/null 2>&1; then
        eerror "Unable to create ${INSTALLATION_DIR}..."
        exit 2
    fi
  fi
}

function downloadAndExtract() {
  local TF_OPENAPI_PROVIDER_PLUGIN_NAME=$1
  local PLUGIN_VERSION=$2
  local XC_OS=$3
  local XC_ARCH=$4

  TMP_INSTALL_DIR=$(mktemp -d)
  if [ "$?" != "0" ]; then
    eerror "Failed to create temporary directory."
    exit 1
  fi

  FILE_NAME="${TF_OPENAPI_PROVIDER_PLUGIN_NAME}_${PLUGIN_VERSION}_${XC_OS}_${XC_ARCH}.tar.gz"
  DOWNLOAD_URL="https://github.com/dikhan/${TF_OPENAPI_PROVIDER_PLUGIN_NAME}/releases/download/v${PLUGIN_VERSION}/${FILE_NAME}"

  # verifying curl is available in the system
  # Curl
  if ! hash curl 2>/dev/null; then
    eerror "curl not available on this system, please install it."
    exit 1
  fi

  # fetch terraform-provider-openapi plugin and dump it into the temp folder
  edebug "Downloading the latest released version of the OpenAPI Terraform plugin from ${DOWNLOAD_URL} into temporally folder: ${TMP_INSTALL_DIR}"
  if ! curl -L --silent ${DOWNLOAD_URL} --output ${TMP_INSTALL_DIR}/${FILE_NAME}
  then
    eerror "Failed to download ${DOWNLOAD_URL}"
    cleanup
    exit 1
  fi

  # extract contents of tar file downloaded in the temp folder
  edebug "Extracting ${TF_OPENAPI_PROVIDER_PLUGIN_NAME} from ${FILE_NAME}..."
  $(cd $TMP_INSTALL_DIR && tar -xz -f $FILE_NAME)
  if [ "$?" != "0" ]; then
    eerror "Failed to extract downloaded tar file ${TMP_INSTALL_DIR}/${FILE_NAME}"
    cleanup
    exit 1
  fi
}

function printRequiredProvidersExample() {
	if [ $TF_MINOR_VERSION -ge 13 ]
  then
    esilent " |--> \033[1;31mImportant Note:${colrst} As of Terraform >=0.13 each Terraform module must declare which providers it requires, so that Terraform can install and use them. You can copy into your .tf file the following snippet was autogenerated for your convenience based on the input provided:"
    esilent
    echo -e "\033[0;33mterraform {"
    echo -e "\033[0;33m  required_providers {"
    echo -e "\033[0;33m    ${PROVIDER_NAME} = {"
    echo -e "\033[0;33m      source  = \"${PROVIDER_SOURCE_ADDRESS}/${PROVIDER_NAME}\""
    echo -e "\033[0;33m      version = \">= ${PLUGIN_VERSION}\""
    echo -e "\033[0;33m    }"
    echo -e "\033[0;33m  }"
    echo -e "\033[0;33m}"
  fi
}

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
        --provider-source-address | -s)
            shift
            PROVIDER_SOURCE_ADDRESS=$1
            ;;
        --compiled-plugin-path | -c) # Only used for building purposes, not recommended for external use
            shift
            TF_SOURCE_PLUGIN_PATH=$1
            ;;
        --debug | -d)
            shift
            verbosity=$dbg_lvl
            ;;
        *)
            eerror "Unknown option: $1"
            usage
            exit 2
    esac
    shift
done

if [ "$PROVIDER_NAME" == "" ]; then
  eerror "required argument --provider-name missing value. Check the usage for more info."
	usage
	exit 1
fi

determineOSAndArch

# installation variables
TF_OPENAPI_PROVIDER_PLUGIN_NAME="${TF_PROVIDER_BASE_NAME}openapi"

if [ -z $TF_SOURCE_PLUGIN_PATH ]
then
  PLUGIN_VERSION="$(curl https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/version -s)"
  downloadAndExtract $TF_OPENAPI_PROVIDER_PLUGIN_NAME $PLUGIN_VERSION $XC_OS $XC_ARCH
  TF_SOURCE_PLUGIN_PATH="${TMP_INSTALL_DIR}/${TF_OPENAPI_PROVIDER_PLUGIN_NAME}_v${PLUGIN_VERSION}"
else # This branch is only to be exercised when using a pre-compiled provider using the make install target
  PLUGIN_VERSION=`cat ${_DIR}/../version`
  einfo "Using the pre-compiled OpenAPI Terraform Plugin specified: ${TF_SOURCE_PLUGIN_PATH} v${PLUGIN_VERSION}"
fi

setupInstallationPath $PROVIDER_NAME $PLUGIN_VERSION $XC_OS $XC_ARCH
TF_DESTINATION_PLUGIN_INSTALLATION_PATH="${INSTALLATION_DIR}/${TF_PROVIDER_PLUGIN_NAME}"

# check we have write permissions on $INSTALLATION_DIR
if [ -w ${INSTALLATION_DIR} ]; then
  if ! mv "${TF_SOURCE_PLUGIN_PATH}" "${TF_DESTINATION_PLUGIN_INSTALLATION_PATH}"; then
    eerror "Failed to move '${TF_SOURCE_PLUGIN_PATH}' binary to ${TF_DESTINATION_PLUGIN_INSTALLATION_PATH}"
    cleanup
    exit 1
  fi
else
    eerror "Unable to write to ${INSTALLATION_DIR} due to lack of write permissions"
    cleanup
    exit 1
fi

cleanup
echo -e "\033[1;32mTerraform provider successfully installed!${colrst}"
esilent " |--> Installation Path: ${TF_DESTINATION_PLUGIN_INSTALLATION_PATH}"
printRequiredProvidersExample
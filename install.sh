#!/bin/bash

cd $HELM_PLUGIN_DIR
version="$(cat plugin.yaml | grep "version" | cut -d '"' -f 2)"
echo "Installing tvm-upgrade ${version} ..."

# Find correct archive name
unameOut="$(uname -s)"

case "${unameOut}" in
    Linux*)             os=Linux;;
    Darwin*)            os=Darwin;;
    MINGW*|MSYS_NT*)    os=windows;;
    *)                  os="UNKNOWN:${unameOut}"
esac

arch=`uname -m`

if echo "$os" | grep -qe '.*UNKNOWN.*'
then
    echo "Unsupported OS / architecture: ${os}_${arch}"
    exit 1
fi

url="https://github.com/trilioData/tvm-helm-plugins/blob/main/dist/tvm-upgrade_v0.0.0_linux_amd64.tar.gz?raw=true"

filename=`echo ${url} | sed -e "s/^.*\///g"`

  # Download archive
  if [ -n $(command -v curl) ]
  then
      curl -sSL -O $url
  elif [ -n $(command -v wget) ]
  then
      wget -q $url
  else
      echo "Need curl or wget"
      exit -1
  fi

# Install bin
rm -rf bin && mkdir bin && tar xvf $filename -C bin > /dev/null && rm -f $filename
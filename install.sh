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

url="https://github.com/trilioData/tvm-helm-plugins/bin/tvm-upgrade"

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
rm -rf bin && mkdir bin && mv tvm-upgrade bin/ && cd bin && chmod +x tvm-upgrade
#tar xvf $filename -C bin > /dev/null && rm -f $filename
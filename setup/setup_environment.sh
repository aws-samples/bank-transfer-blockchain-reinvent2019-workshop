#!/bin/sh
# Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License").
# You may not use this file except in compliance with the License.
# A copy of the License is located at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

echo Pre-requisities
echo installing jq
sudo yum -y install jq

echo Updating AWS CLI to the latest version
sudo pip install awscli --upgrade

echo "Installing needed python libraries"
sudo alternatives --set python /usr/bin/python3.6
sudo pip install boto3
sudo pip install requests

echo "Updating YUM repos"
sudo yum update -y

echo "Installing telnet and emacs"
sudo yum install -y telnet
sudo yum -y install emacs

echo "Updating Docker"
sudo yum update -y docker

echo "Installing GIT and other tools"
sudo yum install libtool-ltdl-devel -y
sudo yum install git -y

echo "Installing Docker Compose"
sudo curl -L https://github.com/docker/compose/releases/download/1.20.0/docker-compose-`uname -s`-`uname -m` -o /usr/local/bin/docker-compose
sudo chmod a+x /usr/local/bin/docker-compose
sudo yum install libtool -y

echo "Fixing Bash Profile"
rm ~/.bash_profile
cat > ~/.bash_profile << EOF
# .bash_profile
# Get the aliases and functions
if [ -f ~/.bashrc ]; then
    . ~/.bashrc
fi
# User specific environment and startup programs
PATH=$PATH:$HOME/.local/bin:$HOME/bin
# GOROOT is the location where Go package is installed on your system
export GOROOT=/usr/lib/golang/
# GOPATH is the location of your work directory
export GOPATH=$HOME/go
# PATH in order to access go binary system wide
export PATH=$GOROOT/bin:$PATH
EOF

source ~/.bash_profile

echo "Checking versions"
docker version
sudo /usr/local/bin/docker-compose version
go version

## Setup Fabric client
echo "Setting up Fabric Client"
go get -u github.com/hyperledger/fabric-ca/cmd/...
cd /home/ec2-user/go/src/github.com/hyperledger/fabric-ca
make fabric-ca-client
export PATH=$PATH:/home/ec2-user/go/src/github.com/hyperledger/fabric-ca/bin # Add this to your.bash_profile to preserve across sessions
echo "export PATH=\$PATH:/home/ec2-user/go/src/github.com/hyperledger/fabric-ca/bin" >> ~/.bash_profile
cd ~

echo "Getting TLS Certificate for Managed Blockchain"
aws s3 cp s3://us-east-1.managedblockchain/etc/managedblockchain-tls-chain.pem  /home/ec2-user/managedblockchain-tls-chain.pem

echo "Checking out Fabric Samples"
cd ~
git clone --single-branch --branch release-1.2 https://github.com/hyperledger/fabric-samples.git

echo "Checking out the Workshop"
cd ~/environment/
#git clone XYC

echo "Downloading Fabric Repo"
wget https://github.com/hyperledger/fabric/archive/v1.4.2.tar.gz
tar xvzf v1.4.2.tar.gz
mkdir -p ~/go/src/github.com/hyperledger
mv fabric-1.4.2 ~/go/src/github.com/hyperledger/fabric
rm v1.4.2.tar.gz

echo "Creating CLI docker compose file at ~/docker-compose-cli.yaml"
cat > ~/docker-compose-cli.yaml << EOF
# Copyright 2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License").
# You may not use this file except in compliance with the License.
# A copy of the License is located at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

version: '2'
services:
  cli:
    container_name: cli
    image: hyperledger/fabric-tools:1.2.0
    tty: true
    environment:
      - GOPATH=/opt/gopath
      - CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock
      - CORE_LOGGING_LEVEL=info # Set logging level to debug for more verbose logging
      - CORE_PEER_ID=cli
      - CORE_CHAINCODE_KEEPALIVE=10
    working_dir: /opt/home
    command: /bin/bash
    volumes:
        - /var/run/:/host/var/run/
        - /home/ec2-user/go/:/opt/gopath/
        - /home/ec2-user:/opt/home
EOF

echo "Pulling needed docker images. "
docker pull hyperledger/fabric-tools:1.2.0

echo "Bringing up Hyperledger Fabric CLI"
docker-compose -f ~/docker-compose-cli.yaml up &

sleep 3

echo "Creating utils."
BIN_DIRECTORY=/home/ec2-user/bin
if [ ! -d $BIN_DIRECTORY ]; then
  mkdir -p $BIN_DIRECTORY
fi

cat > $BIN_DIRECTORY/peer << EOF
source ~/fabric_exports
docker exec -e "CORE_PEER_TLS_ENABLED=true" -e "CORE_PEER_TLS_ROOTCERT_FILE=/opt/home/managedblockchain-tls-chain.pem" -e "CORE_PEER_LOCALMSPID=\$MSP" -e "CORE_PEER_MSPCONFIGPATH=\$MSP_PATH" -e "CORE_PEER_ADDRESS=\$PEERSERVICEENDPOINT" cli peer \$*
EOF
chmod +x $BIN_DIRECTORY/peer

cat > $BIN_DIRECTORY/configtxgen << EOF
source ~/fabric_exports
docker exec cli configtxgen \$*
EOF
chmod +x $BIN_DIRECTORY/configtxgen

echo "We have setup the following utilities:"
echo "1) fabric-ca-client -> Your CA Client"
echo "2) Docker CLI helper script."

# Adding workshop scripts bin path

echo "=========================================================================="
echo "Completed successfully. Please run:"
echo "1) source ~/.bash_profile"
echo "2) ~/environment/BanktransferBlockchainWorkshop/setup/setup_fabric_environment.py to finish setting up this environment."
echo "=========================================================================="
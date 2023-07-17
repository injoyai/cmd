#!/bin/bash
# 安装更新Node.js和Npm所需要的软件
echo "update apt packge ..."
sudo apt update
# 添加NodeSource APT存储库和用于验证软件包的PGP密钥
echo "add NodeSource APT is PGP"
sudo apt install apt-transport-https curl ca-certificates software-properties-common
echo "add apt-get nodejs16.x and PGP" # 该行命令完成了apt-get存储库的添加并添加了PGP密钥
curl -sL https://deb.nodesource.com/setup_16.x | sudo -E bash -
echo "安装Nodejs和npm" 
sudo apt-get install -y nodejs

sudo nodejs -v
sudo npm -v

echo "安装Git"
#Install git
sudo apt-get install git
echo "下载node-red源码"
#Clone the code
git clone https://github.com/node-red/node-red.git
cd node-red
echo "安装依赖" 
#Install the node-red dependencies
sudo npm install
sudo npm install grunt
echo "编译"
#Build the code
sudo npm run build
echo "运行"
#Run
sudo npm run start

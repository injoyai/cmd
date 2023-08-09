

echo "安装v2rayA"
echo "安装内核"
curl -Ls https://mirrors.v2raya.org/go.sh | sudo bash
echo "添加公钥"
wget -qO - https://apt.v2raya.org/key/public-key.asc | sudo tee /etc/apt/trusted.gpg.d/v2raya.asc
echo "添加v2rayA软件源"
echo "deb https://apt.v2raya.org/ v2raya main" | sudo tee /etc/apt/sources.list.d/v2raya.list
echo "更新软件源"
sudo apt update
echo "安装v2rayA"
sudo apt install v2raya
echo "启动 v2rayA"
sudo systemctl enable v2raya.service
sudo systemctl start v2raya.service
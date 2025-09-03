### 说明
自用常用cmd工具集合

### 怎么安装
* 通过curl安装
```shell
sudo curl -sSL https://oss.002246.xyz/in-store/install.sh | bash
```
* 通过wget安装
```shell
sudo wget -qO- https://oss.002246.xyz/in-store/install.sh | bash
```
* 通过go mod安装

### 如何使用
* 连接到TCP服务器
```shell
in dial :8080
```
* 作为TCP服务器
```shell
in server tcp -p=8080
```

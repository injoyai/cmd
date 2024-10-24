### 简介
1. 一个cmd快速开启转发和透传的工具
2. 引用了乱七八糟的包,重新申明go mod,避免影响到主包

### 如何使用

1. 端口转发,例`80`端口转到`8080`
    ```shell
    proxy.exe forward "80->:8080"
    ```

2. 端口转发,例`80`端口转到局域网的`8080`端口
    ```shell
    proxy.exe forward "80->:192.168.1.3:8080"
    ```

3. 代理服务端,例监听`7000`端口用于客户端连接,监听`80`端口转发至客户端的`8080`端口
    ``` shell
    proxy.exe server -p=7000 "80->:8080"
    ```

4. 代理客户端,例连接到`8.8.8.8:7000`上,请求监听服务的`80`端口,转发到本地的`8080`端口,请求验证的账号密码是`test`
    ```shell
    proxy.exe client 8.8.8.8:7000 "80->:8080" --username=test --password=test
    ```
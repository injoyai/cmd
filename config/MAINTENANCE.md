# 配置与文件写入维护说明

本文面向开发维护；命令用法和用户需要知道的配置位置见根目录 README。

## 配置存储与兼容读取

配置主路径由 `os.UserConfigDir()` 与 `injoy/i/config.yaml` 拼接得到。

仅在保存时创建目录和文件，使用同目录临时文件替换；Unix 下新建目录权限为 `0700`，
配置文件权限为 `0600`。Windows 文件访问权限由目录 ACL 控制。
新文件不存在时兼容读取旧的 `global@dfff0a7fa1a55c8c1a4966c19f6da452`：
macOS 为 `~/Library/Application Support/injoy/data/cache/`，Windows 为
`~/AppData/Local/injoy/data/cache/`，Linux 为 `/var/lib/injoy/i/cache/`。
后续保存写入新位置并保留旧文件；新文件存在时优先使用新文件，读取或解析失败时返回错误，禁止覆盖。
不兼容过渡期的 `config.json`。可手动编辑 YAML；CLI 保存时会重新排版，不保留注释。

## 配置键迁移

已有文件中的 `downloadDir`、`downloadNoticeEnable`、`downloadNoticeText` 分别迁移为
`download.dir`、`download.noticeEnable`、`download.noticeText`；`proxy` 字符串迁移为
`proxy.addr`，`proxyIgnore` 迁移为 `proxy.ignore`；原 `resource` 字符串按逗号拆为
`resource.urls` 数组。`uploadMinio`、`customOpen` 分别迁移为 `upload.minio`、`open.custom`。

迁移只在内存中进行，查询不会写盘；下次修改保存时，完整配置写回原来的 `config.yaml`，
移除已迁移旧键，保留其他值、类型和自定义项。新旧键并存时新键优先，对象逐项合并；
目标分组被非对象配置占用时返回错误，禁止覆盖原文件。后续 CLI 修改请使用新键名。

## 目录批量替换

`dir -r old=new` 对每个文件先写入同目录临时文件，完整写入并关闭后再替换，保留权限。

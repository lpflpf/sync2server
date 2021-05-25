# 自动上传工具

> 检测文件夹目录，并自动scp 到远程。

## 特性

1. 支持 SSH 密码 或者密钥上传
2. 仅更新不删除
3. 支持 windows 上传到 Linux
4. 支持 Json/ yaml / prop 文件配置

## 配置

```json
[
    {
        "source": "",       // 监控目录
        "dest": "",         // 上传远端路径
        "ignore": [""],     // 忽略的文件，支持glob格式
        "protocol": "sftp", // 固定为sftp 格式
        "host": "",         // 远端地址
        "port": 22,         // 端口号
        "username": "",     // 用户名
        "password": "",     // 密码，如果使用私钥，可不填
        "private_key": ""   // 私钥地址
    }
]
```

## 使用

执行 `./fsnotify --config=./config/notify.json`

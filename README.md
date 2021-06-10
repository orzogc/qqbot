# qqbot
[![Go Reference](https://pkg.go.dev/badge/github.com/orzogc/qqbot.svg)](https://pkg.go.dev/github.com/orzogc/qqbot)

QQ聊天机器人

### 配置
先运行`qqbot -g`生成`device.json`，配置好json后运行`qqbot`

#### 配置qqbot.json
`qqbot.json`内容：
```
{
    "bot": {
        "account": 123456, // 机器人QQ号
        "password": "123456" // 机器人QQ密码
    }
}
```

#### 配置command.json
`command.json`内容：
```
{
    "setu": {
        "lolicon": {
            "apikey": "123456abc", // lolicon接口的apikey，需自行申请（https://api.lolicon.app/#/setu）
            "r18": 0, // 0为非 R18，1为 R18，2为混合
            "proxy": "disable" // 设置返回的原图链接的域名，设置为disable返回真正的原图链接
        },
        "pixiv": {
            "phpsessid":"123456abc", // pixiv网页Cookie里的PHPSESSID值，为空的话没有r18图片
            "searchOption": {
                "contentRating": "safe", // 空为返回所有结果，safe只返回非R18图片，r18只返回R18图片
                "mode": "s_tag" // 空为精确tag搜索，s_tag为模糊tag搜索，s_tc为只搜索标题
            }
        }
    }
}
```

更多配置请看源码

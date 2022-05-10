# gitlab_webhook 事件通知

**背景: 开发需要在提交merge时触发一个通知.展示一些特定信息,方便开发及时知道mr状态.总体思路:设计一个webhook,根据gitlab自带events事件,请求企业微信机器人,到达通知目的**
> gitlab_webhook 事件样例: https://git.xkool.org/help/user/project/integrations/webhooks ,可以根据请求数据封装自定义信息

## 服务设计图
```seq
gitlab_events->gitlab_webhook: send events
```
## 服务启动
#####构建服务镜像
`docker build -t registry-vpc.cn-shenzhen.aliyuncs.com/xkool_dev/eventserver:1.0 -f dockerfile  .`
#####运行
`docker run -it -d --restart=always -p 8079:8079 registry-vpc.cn-shenzhen.aliyuncs.com/xkool_dev/eventserver:1.0`

### End
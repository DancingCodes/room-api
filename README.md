# Room

Room 是一个多人语音房 App。第一版目标是跑通：

```txt
注册/登录 -> 房间列表 -> 创建或进入房间 -> 房间内文字聊天和开麦状态
```

第一版不做真实语音流，只维护开麦/闭麦状态。

## 文档

- [后端技术方案.md](./后端技术方案.md)
- [后端业务流程.md](./后端业务流程.md)
- [Android前端技术方案.md](./Android前端技术方案.md)
- [App低保真原型.md](./App低保真原型.md)
- [接口文档.md](./接口文档.md)
- [数据库设计.md](./数据库设计.md)
- [环境变量说明.md](./环境变量说明.md)
- [开发任务拆分.md](./开发任务拆分.md)
- [错误码说明.md](./错误码说明.md)
- [部署说明.md](./部署说明.md)
- [测试用例.md](./测试用例.md)
- [发布检查清单.md](./发布检查清单.md)

## 技术栈

后端：

- Go
- Gin
- GORM
- MySQL 云数据库
- JWT
- WebSocket
- QQ 邮箱 SMTP
- 腾讯云 COS

Android：

- Kotlin
- Android Jetpack
- MVVM
- ViewModel
- StateFlow
- Coroutines
- Retrofit
- OkHttp WebSocket
- DataStore
- Coil

## 本地启动

```txt
go run ./cmd/api
```

默认端口：

```txt
8080
```

健康检查：

```txt
GET /health
```

## 第一版范围

- 注册邮箱验证码
- 注册前头像上传
- 用户注册
- 用户登录
- 忘记密码
- 修改昵称
- 修改头像
- 房间列表
- 创建 2 人房 / 8 人房
- 进入房间
- 离开房间
- 房主自动转让
- 文字聊天
- 开麦/闭麦状态
- WebSocket 房间事件广播

## 第一版不做

- 真实语音通话
- 好友系统
- 礼物系统
- 消息撤回
- 图片消息
- 自动重连
- 管理后台


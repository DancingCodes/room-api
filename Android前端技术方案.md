# Room App Android 第一版前端方案

## 1. 前端目标

第一版只做原生 Android 单端。

核心链路：

```txt
启动 App -> 登录/注册 -> 房间列表 -> 创建或进入房间 -> 房间内文字聊天和开麦状态
```

第一版不做真实语音通话，只做开麦/闭麦状态展示。

## 2. 技术选型

建议使用：

- Kotlin
- Android Jetpack
- MVVM
- ViewModel
- StateFlow
- Coroutines
- Retrofit
- OkHttp
- OkHttp WebSocket
- Kotlinx Serialization 或 Moshi
- Coil
- DataStore
- Material Components

说明：

- Retrofit 负责 HTTP API
- OkHttp WebSocket 负责房间实时事件
- DataStore 保存 token 和用户信息
- Coil 加载头像
- ViewModel + StateFlow 管理页面状态

## 3. Android 模块结构

```txt
app/
  src/main/java/com/room/app/
    data/
      local/
      remote/
      repository/
      model/
    domain/
      usecase/
    ui/
      auth/
      rooms/
      room/
      profile/
      common/
    websocket/
    upload/
    util/
```

目录说明：

- `data/local`：DataStore、本地登录态
- `data/remote`：Retrofit API、DTO
- `data/repository`：数据仓库
- `data/model`：数据模型
- `domain/usecase`：业务用例，第一版可少量使用
- `ui/auth`：登录、注册、忘记密码
- `ui/rooms`：房间列表、创建房间弹窗
- `ui/room`：房间详情
- `ui/profile`：个人资料
- `ui/common`：通用组件和状态
- `websocket`：房间 WebSocket 管理
- `upload`：头像上传
- `util`：校验、时间、错误处理

## 4. 页面结构

建议页面：

```txt
SplashActivity / MainActivity
LoginFragment
RegisterFragment
ForgotPasswordFragment
RoomListFragment
RoomDetailFragment
ProfileFragment
CreateRoomBottomSheet
```

导航方式：

- 单 Activity + Fragment
- 使用 Navigation Component

页面规则：

- 未登录进入登录页
- 已登录进入房间列表页
- 创建房间成功后进入房间详情页
- 进入房间详情页后建立 WebSocket
- 点击退出或 WebSocket 断开后回到房间列表页

## 5. 启动流程

App 启动时：

1. 从 DataStore 读取 token
2. token 不存在，进入登录页
3. token 存在，请求当前用户信息
4. 请求成功，进入房间列表页
5. 鉴权失败，清除 DataStore，进入登录页

本地保存：

```txt
token
user_id
username
nickname
email
avatar_url
```

## 6. 登录

页面：`LoginFragment`

字段：

- 用户名
- 密码

校验：

- 用户名不能为空
- 密码不能为空

接口：

```txt
POST /api/v1/auth/login
```

成功后：

- 保存 token
- 保存用户信息
- 跳转房间列表页

## 7. 注册

页面：`RegisterFragment`

字段：

- 头像
- 用户名
- 昵称
- 邮箱
- 邮箱验证码
- 密码

校验：

- 头像必填
- 用户名 4-20 位，只能包含字母、数字、下划线
- 昵称 1-8 位
- 邮箱必须是标准邮箱格式
- 邮箱验证码不能为空
- 密码 6-20 位

接口顺序：

```txt
POST /api/v1/uploads/avatar
POST /api/v1/auth/register-code
POST /api/v1/auth/register
```

头像选择：

- 使用 Android Photo Picker
- Android 13+ 使用系统 Photo Picker
- 低版本可用 `ActivityResultContracts.GetContent`
- 上传前校验文件类型和大小

头像上传：

- 使用 Retrofit Multipart 或 OkHttp Multipart
- 上传成功后保存 `avatar_url`
- 用户重新选择头像时，使用新的 `avatar_url`
- 旧 COS 文件第一版不清理

验证码：

- 获取验证码后按钮倒计时 60 秒
- 倒计时期间按钮不可点击

注册成功：

- 保存 token
- 保存用户信息
- 跳转房间列表页

## 8. 忘记密码

页面：`ForgotPasswordFragment`

字段：

- 邮箱
- 邮箱验证码
- 新密码

校验：

- 邮箱必须是标准邮箱格式
- 邮箱验证码不能为空
- 新密码 6-20 位

接口：

```txt
POST /api/v1/auth/password-reset-code
POST /api/v1/auth/reset-password
```

交互：

- 获取验证码后按钮倒计时 60 秒
- 重置成功后返回登录页

## 9. 房间列表

页面：`RoomListFragment`

展示：

- 当前用户头像
- 创建房间按钮
- 房间名称
- 当前人数
- 最大人数
- 进入按钮

接口：

```txt
GET /api/v1/rooms?page=1&page_size=20
POST /api/v1/rooms
```

交互：

- 首次进入加载第一页
- 下拉刷新重新加载第一页
- 上拉加载更多下一页
- 房间满员时按钮显示“已满”，不可点击
- 点击头像进入个人资料页
- 点击创建房间打开 `CreateRoomBottomSheet`

说明：

- 用户进入房间后会跳转房间详情页，不会继续停留在房间列表页
- “一个用户同一时间只能在一个房间”由后端兜底

## 10. 创建房间

组件：`CreateRoomBottomSheet`

字段：

- 房间类型：2 人房间 / 8 人房间

规则：

- 房间名称不允许输入
- 后端按当前用户昵称生成房间名称
- 只能选择 2 人房或 8 人房
- 创建成功后进入房间详情页
- 进入房间详情页后建立 WebSocket

接口：

```txt
POST /api/v1/rooms
```

请求：

```json
{
  "max_members": 8
}
```

## 11. 房间详情

页面：`RoomDetailFragment`

展示：

- 房间名称
- 当前人数 / 最大人数
- 成员座位
- 空位
- 房主身份
- 麦克风状态
- 聊天消息
- 消息输入框
- 发送按钮
- 开麦/闭麦按钮
- 退出按钮

接口：

```txt
GET /api/v1/rooms/:room_id
GET /api/v1/rooms/:room_id/messages?limit=20&before_id=100
POST /api/v1/rooms/:room_id/messages
PATCH /api/v1/rooms/:room_id/mic
POST /api/v1/rooms/:room_id/leave
GET /ws/v1/rooms/:room_id?token=jwt_token
```

进入页面：

1. 获取房间详情
2. 获取历史消息
3. 建立 WebSocket
4. 渲染成员座位和聊天区

退出页面：

1. 主动调用离开房间接口
2. 主动关闭 WebSocket
3. 返回房间列表页

座位规则：

- 2 人房展示 2 个座位
- 8 人房展示 8 个座位
- 成员座位展示头像、昵称、房主身份、麦克风状态
- 空位展示空状态

聊天规则：

- 消息最多 50 字
- 输入为空、trim 后为空、超过 50 字时，发送按钮不可点击
- 发送消息走 HTTP
- WebSocket 只接收广播
- 收到新消息后追加到底部
- 用户在底部时自动滚动到底部

WebSocket 规则：

- 进入房间详情页后连接
- 断开等于离开房间
- 第一版不做自动重连
- 断开后 Toast：连接已断开，请重新进入房间

开麦规则：

- 点击开麦/闭麦调用接口
- 按钮进入 loading 状态
- 最终状态以 WebSocket 广播为准

## 12. 个人资料

页面：`ProfileFragment`

展示：

- 当前头像
- 用户名
- 昵称
- 邮箱

可修改：

- 头像
- 昵称

不可修改：

- 用户名
- 邮箱

接口：

```txt
GET /api/v1/users/me
PATCH /api/v1/users/me
POST /api/v1/users/me/avatar
```

交互：

- 修改昵称时先 trim
- 昵称 1-8 位
- 昵称不可重复由后端校验
- 更换头像使用 Photo Picker
- 头像上传失败 Toast：头像上传失败，请重试
- 保存成功后更新 DataStore 和内存状态

## 13. ViewModel 设计

### 13.1 AuthViewModel

负责：

- 登录
- 注册
- 发送注册验证码
- 发送重置密码验证码
- 重置密码
- 保存登录态

### 13.2 RoomListViewModel

负责：

- 加载房间列表
- 下拉刷新
- 上拉加载更多
- 创建房间

### 13.3 RoomDetailViewModel

负责：

- 加载房间详情
- 加载历史消息
- 发送消息
- 开麦/闭麦
- 离开房间
- 处理 WebSocket 事件

### 13.4 ProfileViewModel

负责：

- 获取个人资料
- 修改昵称
- 上传头像
- 更新本地用户信息

## 14. Repository 设计

### 14.1 AuthRepository

接口：

- login
- register
- sendRegisterCode
- sendPasswordResetCode
- resetPassword

### 14.2 UserRepository

接口：

- getMe
- updateProfile
- uploadAvatar

### 14.3 RoomRepository

接口：

- getRooms
- createRoom
- getRoomDetail
- leaveRoom
- getMessages
- sendMessage
- updateMicStatus

### 14.4 SocketRepository

接口：

- connectRoom
- close
- observeEvents

## 15. WebSocket 事件处理

### 15.1 member.joined

处理：

- 添加成员
- 更新当前人数
- 可插入系统提示

### 15.2 member.left

处理：

- 移除成员
- 更新当前人数
- 可插入系统提示

### 15.3 room.owner_changed

处理：

- 更新房主 ID
- 更新成员身份

### 15.4 message.created

处理：

- 追加消息
- 必要时滚动到底部

### 15.5 member.mic_updated

处理：

- 更新成员麦克风状态

## 16. 网络封装

HTTP：

- Retrofit
- OkHttp Interceptor 自动添加 token
- `code = 401` 时清除登录态并跳转登录
- `code = 500` 时 Toast 后端 message

上传：

- Multipart
- 统一限制格式和大小

WebSocket：

- OkHttp WebSocket
- 进入房间详情时创建
- 退出房间时关闭
- 收到事件后转成 ViewModel 状态

## 17. 表单校验

用户名：

```txt
4-20 位，只能包含字母、数字、下划线
```

密码：

```txt
6-20 位
```

昵称：

```txt
1-8 位，提交前 trim
```

邮箱：

```txt
标准邮箱格式
```

文字消息：

```txt
最多 50 字，提交前 trim
```

头像：

```txt
jpg/jpeg/png/webp，最大 2MB
```

## 18. Android 交互细节

键盘：

- 聊天输入框聚焦时，消息列表不能被键盘遮挡
- 发送后清空输入框

状态栏和导航栏：

- 页面需要适配状态栏
- 房间底部操作区需要避开系统导航栏

Loading：

- 登录
- 注册
- 发送验证码
- 上传头像
- 创建房间
- 发送消息
- 开麦/闭麦

Toast：

- 表单错误
- 网络错误
- 上传失败
- WebSocket 断开

## 19. 第一版不做

第一版 Android 不做：

- 自动重连
- 真实语音通话
- 图片消息
- 消息撤回
- 好友系统
- 礼物系统
- 主题切换
- 多语言
- 复杂动画


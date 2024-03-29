# mock-cmpp-stress-test
基于CMPP短信接口协议，模拟客户端、服务端的压测工具


## 目录
- [项目说明](#项目说明)
- [参数详解](#参数详解)
- [功能说明](#功能说明)

## 项目说明

mock-cmpp-stress-test 是针对CMPP协议下短信发送服务的轻量级压测工具。支持CMPP客户端、服务端独立部署，分别模拟高并发场景下的大量用户请求和渠道返回。压测结果使用HTML格式输出数据图表，亦可扩展redis用于更细粒度压测结果数据存储。详细配置如下，支持二次开发。

## 参数详解
```toml
##################### cmpp 客户端配置模块 #####################
[cmpp_client]
# cmpp 客户端使用的版本
version = "V21"
# cmpp 连接读包超时时间，单位秒
read_timeout = 1
# cmpp 连接心跳间隔时间，单位秒
active_test_interval = 1
# cmpp 连接允许无响应数据包的最大个数
max_no_resp_pkg_num = 3
# 是否启用 cmpp 客户端
enable = true

# cmpp 连接账户信息
[[cmpp_client.accounts]]
# cmpp 客户端需要连接的服务端IP地址
ip = "127.0.0.1"
# cmpp 客户端需要连接的服务端端口号
port = 7890
# cmpp 连接账户
username = ""
# cmpp 连接账户密码
password = ""
# cmpp spId
sp_id = ""
# cmpp spCode
sp_code = ""
##################### cmpp 客户端配置模块 #####################

##################### cmpp 服务端配置模块 #####################
[cmpp_server]
# cmpp 服务端启动IP地址
ip = "127.0.0.1"
# cmpp 服务端启动端口号
port = 7890
# 是否启用 cmpp 服务端
enable = true
# cmpp 服务端使用的版本
version = "V21"
# cmpp 服务端心跳检测时间
heartbeat = 1
# cmpp 服务端无响应时发送最大包个数
max_no_resp_pkgs = 3

# cmpp 服务端验证账号信息（可对照cmpp_client.accounts）
[[cmpp_server.auths]]
username = "200001"
password = "test123"
sp_id = "1000"
sp_code = "1000"
##################### cmpp 服务端配置模块 #####################

##################### 压力测试配置模块 #####################
[stress_test]
# 是否开启压测服务
enable = true

# 压测线程配置
[[stress_test.workers]]
# 压测名称，对应 [[cmpp_client.accounts]] 中的 {ip}:{port}_{username}
name = "127.0.0.1:7890_200002"
# 每秒并发量，必填
concurrency = 1000
# 持续时间和总发送量不可同时为0。同时不为0时，优先使用总发送量压测。
# 持续时间
duration_time = 120
# 总发送量
total_num = 1000000
# 压测间隔时间
sleep = 100

# cmpp 客户端发送短信内容配置
[[stress_test.messages]]
# 扩展码
extend = ""
# 短信内容
content = "【Test】领取属于您的优惠。回T退订"
# 发送手机号
phone = "12345678901"
##################### 压力测试配置模块 #####################

##################### 日志配置模块 #####################
[log]
# 日志所在文件夹
dir = "./log"
# 日志文件名称
file = "mock-cmpp-stress-test.log"
# 日志级别，可选 info、debug、error，默认使用 info
level = "info"
# 是否使用本地时间
local_time = true
##################### 日志配置模块 #####################

##################### redis配置模块 ##################### 
[redis]
# redis 服务IP地址
ip = "127.0.0.1"
# redis 服务端口号
port = 3306
# redis 库
db = 0
# redis 服务密码
password = ""
# redis 超时时间，单位秒
timeout = 300
# 是否启用 redis 存储统计数据，若不启用，则使用内存存储统计数据，统计数据维度将少于启用 redis 存储统计数据。
enable = true
# 是否清除上次遗留数据
clear_key = true
##################### redis配置模块 ##################### 
```
### 功能说明：
- [x] CMPP客户端
    - [x] 建立CMPP连接
    - [x] 发送提交短信、心跳数据包
    - [x] 接收回执数据包
    - [x] 支持 cmpp2.0 及 cmpp3.0
- [x] CMPP服务端
    - [x] 接收CMPP连接，校验用户名密码
    - [x] 接收来自客户端各类型数据包并处理
    - [x] 模拟回执并推送至客户端
    - [x] 支持 cmpp2.0 及 cmpp3.0
    - [ ] 模拟上行，并推送给指定客户端
- [x] 压测服务
    - [x] 设置每秒并发量
    - [x] 可配置压测持续时间或压测总量
    - [x] 存储统计数据，内存最多可存 30min，redis 不限
- [x] 统计数据服务
    - [x] 统计机器性能，CPU、内存、磁盘使用率
    - [x] 统计提交短信、接收回执数据

### 使用工具说明：
- cmpp连接库：https://github.com/bigwhite/gocmpp
- 图表库：https://github.com/go-echarts/go-echarts/blob/master/README_CN.md 







 
# mock-cmpp-stress-test
基于CMPP短信接口协议，模拟客户端、服务端的压测工具


## 目录
- [项目说明](#项目说明)
- [参数详解](#参数详解)

## 项目说明
mock-cmpp-stress-test 是集 cmpp客户端、cmpp服务端 的轻量级模拟压测工具，可单独模拟 cmpp客户端 或 cmpp服务端。提供压测数据图表，或可使用Redis统计维度更丰富的数据。配置灵活，支持二次开发。

## 参数详解
```toml
##################### cmpp 客户端配置模块 #####################
[cmpp_client]
# cmpp 客户端使用的版本
version = "V21"
# cmpp 连接重试次数
retries = 1
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

# cmpp 客户端发送短信内容配置
[[cmpp_client.messages]]
# 扩展码
extend = ""
# 短信内容
content = "【Test】领取属于您的优惠。回T退订"
# 发送手机号
phone = "12345678901"
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
username = "test"
password = "test123"
sp_id = ""
sp_code = ""
##################### cmpp 服务端配置模块 #####################

##################### 压力测试配置模块 #####################
[stress_test]
# 每秒并发量，必填
concurrency = 1000
# 持续时间和总发送量不可同时为0。同时不为0时，优先使用总发送量压测。
# 持续时间
duration_time = 120
# 总发送量
total_num = 10000
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
# redis 服务密码
password = ""
# redis 超时时间，单位秒
timeout = 300
wait = true
# 是否启用 redis 存储统计数据，若不启用，则使用内容存储，统计数据维度将少于启用 redis 存储统计数据。
enable = true
##################### redis配置模块 ##################### 
```


cmpp连接库：https://github.com/bigwhite/gocmpp
图表库：https://github.com/go-echarts/go-echarts/blob/master/README_CN.md 

服务端核心功能：
- 接收CMPP连接
- 模拟回执，上行推送
- 回执上行推送比率可配置
- 回复来自客户端的心跳包

TODO List
- Cache 溢出 && Retry
- Stress Test Result 图表








 
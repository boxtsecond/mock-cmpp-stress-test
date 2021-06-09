# mock-cmpp-stress-test
基于CMPP短信接口协议，模拟客户端、服务端的压测工具

-cc client config file
-sc server config file

服务端核心功能：
- 接收CMPP连接
- 模拟回执，上行推送
- 回执上行推送比率可配置
- 回复来自客户端的心跳包
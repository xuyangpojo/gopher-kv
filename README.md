# Gopher-KV 键值型内存数据库

### 文件结构
data/
| 文件名                | 类别说明     | 类型    |
|----------------------|--------------|----------|
| gkvBitMap.go         | 位图类      |  高级     |
| gkvGraph.go          | 图类        |  高级     |
| gkvHyperLoglog.go    | 基数统计类   |  高级    |
| gkvList.go           | 链表类       |  基础    |
| gkvMap.go            | 映射类       |  基础    |
| gkvSet.go            | 集合类       |  基础    |
| gkvString.go         | 字符串类     |  基础    |
| gkvZSet.go           | 有序集合类   |  基础    |
- keyLock.go 基础锁结构，包括类型全局锁与键级锁(行级锁)
commands.go 命令接口
httpServer.go HTTP服务器接口
config.json 可修改配置文件
helps.go 存储帮助相关信息
main.go 命令程序入口
persistence.go 持久化与反持久化接口
## h5 - 基于gin开发的web服务框架

## 主要功能
- excel数据导出
- 文件zip压缩和下载
- 图像处理（压缩、剪切、合成）开发中...
- 配置文件使用Viper管理，支持热加载
- 日志输出使用Golang日志库slog，搭配lumberjack日志文件切割
- 数据库使用gorm,支持多数据源DBResolver配置
- 企业微信上传文件并通知到微信群
- 后台管理功能 开发中...

## Installation
- 下载源码：git clone https://github.com/brightchain/h5.git
- 下载依赖：go mod tidy
- 运行项目：go run main.go
- 编译项目：go build -o h5

## 服务重启
- 进入当前目录：cd h5
- 运行命令: ./server.sh

## Configuration
### 配置文件
- config.yaml(本地)
- config_release.yaml(生产)
### 如何加载配置文件
- 配置文件应与执行文件同一目录
- 基于gin环境变量GIN_MODE匹配，例如GIN_MODE=release，则加载的配置文件为config_release.yaml，默认为config.yaml
### GIN_MODE如何配置生产模式
- vim ~/.bash_profile
- 输入 export GIN_MODE=release
- 使配置生效 source ~/.bash_profile  
### 配置文件内容
```yaml
app:
  name: app
  port: 8080
mysql:
  user: root
  password: *****
  host: 127.0.0.1
  port: 3306
  db: default
redis:
  host: 127.0.0.1
  port: 6379
  db: 1
```


## License
GNU General Public License v3.0

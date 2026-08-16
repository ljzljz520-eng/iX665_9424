# iX665_9424

Go project for module `pumpstation`.

水厂泵站设备档案服务在内存固定夹具中管理泵站基础信息、附件和操作日志，提供新建、搜索分页、详情、编辑和附件更新接口，当前不连接真实数据库、网络服务或外部 API。

## Standard commands

```bash
go build ./...
go test -count=1 ./...
```

## Run

```bash
go run ./cmd/pumpstation
```

## Frontend

```bash
(cd web && npm run build)
```

## SQL Server 部署配置

演示模式使用自包含夹具，不会建立数据库连接。接入 SQL Server 时可配置以下环境变量，并由部署平台的密钥管理器提供密码：

```bash
SQLSERVER_HOST=sqlserver.example.com
SQLSERVER_PORT=1433
SQLSERVER_DATABASE=PumpStationArchive
SQLSERVER_USER=pumpstation_app
SQLSERVER_PASSWORD='change-me'
```

生产部署应启用 TCP 1433 和 TLS，为应用账号授予设备档案、附件索引与操作日志表的最小读写权限，并通过迁移脚本创建表结构。本项目不会读取这些变量，也不访问真实 SQL Server。

## Docker validation

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh my-go-task linux/arm64
./build_benzhi_docker.sh my-go-task linux/amd64
docker run -it my-go-task:latest
```

## Known initial failures

See `BUG_REPRO.md` for the exact command and output captured during packaging.

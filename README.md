# Prometheus exporter for Harbor 

## 描述

github上目前的`harbor_exporter`的轮子都不行，参考了下官方`mysqld_exporter`的源码设计理念自己写了个。

目前大体完成，`v1.10.3`之前的harbor理论上都行，后续细节性更新和添加上`v2.0`的支持，以及增加告警规则和`grafana`的文件

## 特性

- **collect细化解耦，可以通过选项enable或者disable**
- **很多harbor的版本号没有数字，添加有选项可以覆盖掉metrics里harbor的版本**
- **后期通过go的build tag来区分`v1.0`和`v2.0`**

## Exported Metrics

- 有些 collector 需要`disable`，根据`--help`和下面的`version`信息区分
- `config on ui`就是有些需要在ui上配置，例如gc就是在web上配置垃圾清理后才建议enable它

| version | config on ui |Metric | Meaning | Labels |
| ------ | ------ | ------- | ------ | ---- |
| all | |harbor_up| passwd correct |
| all| |harbor_exporter_collector_duration_seconds | time consuming for each collector| collector=[...] |
|all | |harbor_exporter_last_scrape_error | did an error occur in a scrape |
| all| |harbor_exporter_scrape_errors_total | The number of errors in a scrape | |
| all| |harbor_exporter_scrapes_total | scrape counter| |
| `v1.8.0 <=x< v2.x`| |harbor_health| components status|name=[core, database, jobservice, portal, redis, registry, registryctl]|
| `v1.1 <=x< v2.x`| |harbor_system_volumes_bytes| system volumes info|type=[total, free, used]|
| `x < v2.x`| |harbor_repo_count_total| |type=[private, public, total]|
| `x < v2.x`| |harbor_project_count_total| | type=[private, public, total]|
| `v1.7 <=x< v2.x`| need |harbor_ref_work_gc| |method="GET", ref=[/system/gc]|
| `x < v2.x`| |harbor_ref_work_logs| |method="GET", ref=[/logs]|
| `x < v2.x` | |harbor_ref_work_projects| |method="GET", ref=[/projects/...]|
|`x < v2.x` | |harbor_ref_work_repos| |method="GET", ref=[/repositories/...]|
| `x < v2.x`| |harbor_ref_work_users| |method="GET", ref=[/users/...]|
| `v1.8.0 <=x< v2.x`| need|harbor_ref_work_replication| |method="GET", ref=[/replication/...]|
| `v1.8.0 <=x< v2.x`| need |harbor_registries_healthy| ui /harbor/registries status |name=[...]|



注意事项:

- `/system/gc` 接口没有`page_size`参数支持，如果gc的数量太多可能会拉长`scrape`的时间，酌情打开
- `v1.8.1`的`/projects/1/members/1/`会一直403，这个版本的话建议disable掉`projects`
- `v1.5.1`的`/users`的`page_size=1`不生效，这个版本的话建议disable掉`users`
- `/replication/executions` 这个可能会超时，不建议打开`replication`

### Flags

```shell
./harbor_exporter --help
```

### ENV

```shell
HARBOR_USERNAME
HARBOR_PASSWORD
```

## 使用(usage)

url的路径带上`/api`，除非 harbor 的接口被 nginx rewrite 了，下面给个示例，运行的选项参数enable否根据实际情况

```shell
echo 'HARBOR_PASSWORD=Harbor12345' > /etc/sysconfig/harbor_exporter
chmod 0660 /etc/sysconfig/harbor_exporter
cat >/etc/systemd/system/harbor_exporter.service<<'EOF'
[Unit]
Description=harbor_exporter - Prometheus exporter for Harbor 
Documentation=https://github.com/zhangguanzhang/harbor_exporter
After=network.target
Wants=network-online.target

[Service]
Type=notify
#User=root
NoNewPrivileges=yes
EnvironmentFile=/etc/sysconfig/harbor_exporter
ExecStart=/root/harbor/harbor_exporter \
    --collect.health=false \
    --collect.systemgc=false \
    --collect.users=false \
    --collect.replication=false \
    --harbor-server https://harbor.dev/api/ \

Restart=always
RestartSec=4s
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target 
EOF

systemctl enable --now harbor_exporter
```

### docker部署

```shell
docker run -d -p 9107:9107 -e HARBOR_PASSWORD=password \
    zhangguanzhang/harbor-exporter \
    --harbor.server=https://harbor.dev/api
```

## 编译(build)

see file `build/build.sh`

## 开发的参考

各个版本的`swagger.yaml`文件为下，`2.0`目前api很少，有空测试增加进去

| version range | swagger yml path |
| --- | --- |
| < v1.9.4 | https://github.com/goharbor/harbor/blob/v1.5.1/docs/swagger.yaml |
| v1.10.x | https://github.com/goharbor/harbor/blob/v1.10.3/api/harbor/swagger.yaml |
| v2.0 | https://github.com/goharbor/harbor/blob/v2.0.0/api/swagger.yaml |

很多接口设计都不人性化，web 路由表可以看
https://github.com/goharbor/harbor/blob/v1.10.3/src/core/api/harborapi_test.go


# Prometheus exporter for Harbor 

## 描述

github上目前的`harbor_exporter`的轮子都不行，参考了下官方`mysqld_exporter`的源码设计理念自己写了个。

目前大体完成，`v1.10.3`之前的harbor理论上都行，后续细节性更新和添加上`v2.0`的支持，以及增加告警规则和`grafana`的文件

## 特性

- **collect细化解耦，可以通过选项enable或者disable**
- **很多harbor的版本号没有数字，添加有选项可以覆盖掉metrics里harbor的版本**
- **后期通过go的build tag来区分兼容所有版本harbor**

## Exported Metrics

| Metric | Meaning | Labels | |
| ------ | ------- | ------ | ---- |
|harbor_up| passwd correct | |
|harbor_exporter_collector_duration_seconds | time consuming for each collector| collector=[...] | |
|harbor_exporter_last_scrape_error | did an error occur in a scrape | | |
|harbor_exporter_scrape_errors_total | The number of errors in a scrape | | |
|harbor_exporter_scrapes_total | scrape counter| | |
|harbor_health| components status|name=[core, database, jobservice, portal, redis, registry, registryctl]| |
|harbor_system_volumes_bytes| system volumes info|type=[total, free, used]| |
|harbor_repo_count_total| |type=[private, public, total]| |
|harbor_project_count_total| | type=[private, public, total]| |
|harbor_ref_work_gc| |method="GET", ref=[/system/gc]| |
|harbor_ref_work_logs| |method="GET", ref=[/logs]| |
|harbor_ref_work_projects| |method="GET", ref=[/projects/...]| |
|harbor_ref_work_repos| |method="GET", ref=[/repositories/...]| |
|harbor_ref_work_users| |method="GET", ref=[/users/...]| |

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

url的路径带上`/api`，除非harbor的接口被nginx rewrite了

```shell
./harbor_exporter --harbor-server https://harbor-local.xxxxx.com/api --harbor-pass 'Harbor12345'
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

很多接口设计都不人性化，web 路由表可以看 https://github.com/goharbor/harbor/blob/v1.10.3/src/core/api/harborapi_test.go 

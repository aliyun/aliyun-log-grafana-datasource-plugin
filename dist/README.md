![SLS Grafana Plugin Logo.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690462492110-78e3adec-4590-4ed8-9a13-c6ca545a3a25.png#averageHue=%2355310a&clientId=u91823d39-f9eb-4&from=drop&id=u22ff0de1&originHeight=1017&originWidth=2406&originalType=binary&ratio=1&rotation=0&showTitle=false&size=311629&status=done&style=none&taskId=ud80003b6-4819-425a-b71a-1cdf7c275cd&title=)<br />![](https://img.shields.io/github/stars/aliyun/aliyun-log-grafana-datasource-plugin.svg#id=grEbM&originHeight=20&originWidth=60&originalType=binary&ratio=1&rotation=0&showTitle=false&status=done&style=none&title=) ![](https://img.shields.io/github/languages/code-size/aliyun/aliyun-log-grafana-datasource-plugin#id=nAWkL&originHeight=20&originWidth=110&originalType=binary&ratio=1&rotation=0&showTitle=false&status=done&style=none&title=) ![](https://img.shields.io/badge/change-log-blue.svg#id=t80B3&originHeight=20&originWidth=76&originalType=binary&ratio=1&rotation=0&showTitle=false&status=done&style=none&title=) ![](https://img.shields.io/badge/Grafana_plugin-%E6%97%A5%E5%BF%97%E6%9C%8D%E5%8A%A1_SLS-orange#id=z18xX&originHeight=20&originWidth=170&originalType=binary&ratio=1&rotation=0&showTitle=false&status=done&style=none&title=)<br />**简体中文** | [English](https://github.com/aliyun/aliyun-log-grafana-datasource-plugin/blob/master/README_EN.md) | [旧版README](https://github.com/aliyun/aliyun-log-grafana-datasource-plugin/blob/master/old_README_CN.md) | [Old Version README](https://github.com/aliyun/aliyun-log-grafana-datasource-plugin/blob/master/old_README.md)
<a name="eI2lI"></a>
# 1. 简介
日志服务 SLS 是云原生观测分析平台，为Log/Metric/Trace等数据提供大规模、低成本、实时平台化服务。一站式提供数据采集、加工、分析、告警可视化与投递功能，全面提升研发、运维、运营和安全等场景数字化能力。[官方文档](https://www.aliyun.com/product/sls)<br />本仓库为阿里云日志服务Grafana数据源插件。使用本插件的前提，需要您使用日志服务产品，并拥有至少1个已配置采集的LogStore。
<a name="zsCS1"></a>
# 2. 安装
依赖 Grafana 8.0 及以上版本 , Grafana 8.0 以下请使用1.0版本。

## 版本兼容性

- 插件 **2.39.1** 修复 Grafana 13 默认停用数字 ID 数据源接口后，日志库列表无法加载和 `goto SLS` 返回 404 的问题。
- 本版本最高已验证的 Grafana 版本为 **13.0.2**，并在 **11.4.0** 上回归验证。使用模拟 SLS 响应验证数据源配置、日志库选择、查询与 `goto SLS` 资源调用；更高版本尚未验证，请先在测试环境确认兼容性。
- 插件 **2.39** 在 Grafana **13.0.2** 默认配置下存在上述问题，使用 Grafana 13 时请升级插件。11.4.0 为已验证的旧版对照，不代表所有更高版本均不兼容。
- 保留 Grafana 8.x 及以上的运行时资源 API 调用方式；不要仅凭插件声明的最低 Grafana 版本判断其与未来版本的兼容性。

于[Release处下载](https://github.com/aliyun/aliyun-log-grafana-datasource-plugin/releases)本插件到grafana插件目录下，修改配置文件，在配置文件的[plugins] 节点中，设置 `allow_loading_unsigned_plugins = aliyun-log-service-datasource`， 然后重启grafana。

- mac 
   - 插件目录： `/usr/local/var/lib/grafana/plugins`
   - 配置文件位置：`/usr/local/etc/grafana/grafana.ini`
   - 重启命令：`brew services restart grafana`
- YUM/RPM
   - 插件目录：`/var/lib/grafana/plugins`
   - 配置文件位置：`/etc/grafana/grafana.ini`
   - 重启命令：`systemctl restart grafana-server`
- .tar.gz 
   - 插件目录：`{PATH_TO}/grafana-{x}.{x}.{x}/data/plugins`
   - 配置文件位置：`{PATH_TO}/grafana-{x}.{x}.{x}/conf/defaults.ini`
   - 重启命令：`./bin/grafana-server web`

## 通过 CI 打包（维护者）

将 `.github/workflows/package-plugin.yml` 合并到默认分支后，在 GitHub 的 **Actions → Package plugin → Run workflow** 中选择 `master`，并在 `tag` 输入框填写已有版本 tag（例如 `2.39.1`）。也可以使用：

```bash
gh workflow run package-plugin.yml \
  --repo aliyun/aliyun-log-grafana-datasource-plugin \
  --ref master -f tag=2.39.1
```

- 先将版本 tag 推送到本仓库；支持 `2.39`、`2.39.1`、`v2.39.1` 这类格式。去掉可选的 `v` 前缀后，tag 必须与该提交中 `package.json` 和 `dist/plugin.json` 的版本完全一致。
- 流程打包指定 tag 已提交的 `dist`，并校验必要文件及六个平台的后端文件；不重新编译源码。因此发布前必须完成构建并提交 `dist`。即使旧 tag 不包含工作流，也可以从 `master` 启动流程进行打包。
- 成功后在运行页面的 **Artifacts** 下载产物，解开 GitHub 的外层 artifact 压缩包后，可获得 `aliyun-log-service-datasource-<version>.zip` 和同名 `.zip.sha256` 文件；产物保留 30 天。使用 `sha256sum -c *.zip.sha256` 校验。
- 安装 ZIP 的顶层目录是 `aliyun-log-service-datasource/`，可直接解压到 Grafana 插件目录。工作流不签名，仍需按上述安装说明允许加载未签名插件。
- 工作流只生成下载产物，不创建 tag 或发布 Release。正式发布时，将安装 ZIP 和 SHA256 文件附加到对应 tag 的 GitHub Release。

<a name="t92KT"></a>
# 3. 添加数据源
<a name="YihL2"></a>
## 3.1 日志库配置
在datasource界面添加一个SLS log-service-datasource数据源。在数据源管理面板, 添加 LogService 数据源，在 settings 面板, 设置 Endpoint 为您日志服务 project 的 endpoint ( endpoint 在 project 的概览页可以看到，详情请参见[服务入口](https://help.aliyun.com/document_detail/29008.htm#reference-wgx-pwq-zdb))。例如您的 project 在 qingdao region, Url 可以填 `cn-qingdao.log.aliyuncs.com`。根据需要设置 Project 和 logstore，设置 AccessId 和 AccessKeySecret , 最好配置为子账号的AK。为保证数据安全 , AK保存后清空 , 且不会回显。
<a name="oxIOY"></a>
## 3.2 时序库配置（使用SLS插件，支持SQL查询+组合算子查询）
时序库也可以配置为SLS插件的形式，使用这种接入方式，则支持SQL查询时序库，以及使用PromQL算子查询时序库（详情请查阅：[时序查询语法简介](https://help.aliyun.com/document_detail/171763.html)）。配置方式与1.1节一致，在LogStore中填写metricStore名称即可。
<a name="YLNV4"></a>
## 3.3 时序库配置（原生方式，使用Prometheus插件）
SLS的时序库支持原生Prometheus查询，因此可以直接使用原生Prometheus插件进行数据源接入。请参考如下[官方文档](https://help.aliyun.com/document_detail/173903.html)配置数据源。日志服务MetricStore的URL，格式为**https://{project}.{sls-endpoint}/prometheus/{project}/{metricstore}**。其中_{sls-endpoint}_为Project所在地域的Endpoint，详情请参见[服务入口](https://help.aliyun.com/document_detail/29008.htm#reference-wgx-pwq-zdb)，_{project}_和_{metricstore}_为您已创建的日志服务的Project和Metricstore，请根据实际值替换。例如：**https://sls-prometheus-test.cn-hangzhou.log.aliyuncs.com/prometheus/sls-prometheus-test/prometheus**<br />**从最佳实践角度：时序库可以同时添加Prometheus数据源与SLS数据源，查询时可根据个人习惯任意使用不同的语句查询方式，在Variable中推荐使用SLS数据源（支持与SLS仪表盘互相转化）**
<a name="pw5VJ"></a>
# 4. 使用
<a name="x57r9"></a>
## 4.1 仪表盘变量
<a name="B6Nkt"></a>
### 4.1.1 Grafana变量在query中的写法
从实践上，唯一推荐的写法是：`${var_name}`。<br />从理论上，Grafana支持3种写法，分别是`$varname`、`${var_name}`、`[[varname]]`。但是不加大括号的写法有可能会变量名范围识别错误、双中扩号的写法在未来将会逐步抛弃。[参考来源](https://grafana.com/docs/grafana/latest/dashboards/variables/variable-syntax/)。
<a name="oJtey"></a>
### 4.1.2 一般用法与全选配置
对于大部分SLS DataSource，可以通过SQL语句查询值，将其作为变量。步骤：

- 进入Grafana仪表盘的设置 - Variables
- 类型选择Query、Datasource配置为对应LogStore、书写query
- 在Preview of values中查看结果

![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452672650-76a5a639-4b38-4f1f-b99b-2944b31969a8.png#averageHue=%231e2226&clientId=uce6e5a51-b8b8-4&from=paste&height=196&id=uf3b6bc2a&originHeight=196&originWidth=713&originalType=binary&ratio=1&rotation=0&showTitle=false&size=59106&status=done&style=none&taskId=ube8a64f4-cde0-4b86-a596-bb4f0b38f3f&title=&width=713)<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452672698-b4a5a738-92fc-4ffe-a73f-cd372bfeee50.png#averageHue=%23252830&clientId=uce6e5a51-b8b8-4&from=paste&height=144&id=u586769d4&originHeight=288&originWidth=846&originalType=binary&ratio=2&rotation=0&showTitle=false&size=100765&status=done&style=none&taskId=u5fa8016a-0107-4975-9c04-cfd879ae0c0&title=&width=423)<br />通常，如果作为日志库筛选用，实践中推荐打开`All Option`，并将Custom all value配置为`*`。<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452672720-4ea38ea1-e5d2-4e86-9095-3407c0db57b0.png#averageHue=%23212328&clientId=uce6e5a51-b8b8-4&from=paste&height=202&id=u42ba9285&originHeight=404&originWidth=970&originalType=binary&ratio=2&rotation=0&showTitle=false&size=152094&status=done&style=none&taskId=u6882b17d-442d-4706-a152-ff8a6e5e9fa&title=&width=485)<br />这样在Dashboard Query语句中，按照如下写法：`* and request_method: ${var} | select * from log`即可支持Variable选择与全选。
<a name="COVuD"></a>
### 4.1.3 时序数据在SLS插件中的Variable配置
在配置SLS存储数据源中提到， SLS时序库支持Prometheus原生方式配置，也支持SLS插件方式配置。若使用SLS插件配置，则通常需要搭配promql算子，其使用方式与语法详见：[时序数据查询和分析简介](https://help.aliyun.com/document_detail/171763.html)。<br />以下给出一个示例，该示例获取时序库中所有指标名。<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452716829-b073e6cb-d312-49c6-83cc-80c756bb8ad4.png#averageHue=%231e2225&clientId=uce6e5a51-b8b8-4&from=paste&height=185&id=u84f192d7&originHeight=185&originWidth=774&originalType=binary&ratio=1&rotation=0&showTitle=false&size=61040&status=done&style=none&taskId=u3648391c-6096-466a-a907-c6b841c2d6d&title=&width=774)<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452716951-7eb02e94-462a-498a-af42-e4ed5e19faac.png#averageHue=%2327292e&clientId=uce6e5a51-b8b8-4&from=paste&height=179&id=ub46e3930&originHeight=358&originWidth=1730&originalType=binary&ratio=2&rotation=0&showTitle=false&size=236336&status=done&style=none&taskId=u3422f0da-f342-4670-ae99-a127b016e49&title=&width=865)
<a name="sG2qH"></a>
### 4.1.4 使用Interval变量根据时间跨度控制打点
首先配置一个Grafana 时间变量

| Name | 变量名称，例如myinterval。该名称是您配置中使用的变量，此处为myinterval，则查询条件中需写成**$$myinterval**。 |
| --- | --- |
| Type | 选择**Interval**。 |
| Label | 配置为**time interval**。 |
| Values | 配置为**1m,10m,30m,1h,6h,12h,1d,7d,14d,30d**。 |
| Auto Option | 打开**Auto Option**开关，其他参数保持默认配置。 |

**注意：该时间变量在SLS中不同于Grafana变量，需要在正常变量写法之前，再加上一个**`**$**`**符号，才能正确在SLS语句中转化Interval。**
```yaml
图表类型: TimeSeries
xcol: time
ycol: pv, uv
query: * | select __time__ - __time__ % $${myinterval} as time, COUNT(*)/ 100 as pv, approx_distinct(remote_addr)/ 60 as uv GROUP BY time order by time limit 2000
```
在配置为`1m`时显示为：<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452762876-174cd1ac-ef8e-4b2f-ac31-d833e35bfb7e.png#averageHue=%231d2124&clientId=uce6e5a51-b8b8-4&from=paste&height=770&id=u752a89a5&originHeight=770&originWidth=1622&originalType=binary&ratio=1&rotation=0&showTitle=false&size=413040&status=done&style=none&taskId=u786e32f5-a286-4f5f-8643-64380ba4406&title=&width=1622)<br />在配置为`10m`时显示为：<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452762804-820d2b64-1719-4af0-b340-99b4be4a33f7.png#averageHue=%231d2124&clientId=uce6e5a51-b8b8-4&from=paste&height=774&id=u00307656&originHeight=774&originWidth=1634&originalType=binary&ratio=1&rotation=0&showTitle=false&size=395508&status=done&style=none&taskId=u3311b85a-692d-4458-a6e4-d823026d2c5&title=&width=1634)<br />该功能在仪表盘时间范围选择非常长的时候，可以很方便地进行时间打点密度的控制。
<a name="sPAdc"></a>
## 4.2 图表标准写法概览
<a name="MkJaN"></a>
### 4.2.1 单值图（Stat / Gauge）
xcol写法：`stat`<br />ycol写法：`<数值列>, <数值列>`<br />注意：图中要求的数值列若提供的是非数值列，会直接置0。<br />示例1：
```yaml
图表类型: Stat
xcol: stat
ycol: PV, deltaPercent
query: * | select diff[1] as "PV", round((diff[1] - diff[2])/diff[2] * 100, 2) as deltaPercent from (select compare("PV", 86400) as diff from (select count(*) as "PV" from log))
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452840023-cd2747db-f1d1-4a50-adb8-9ae3555d1ac1.png#averageHue=%231e2125&clientId=uce6e5a51-b8b8-4&from=paste&height=589&id=u133f3dbf&originHeight=589&originWidth=1612&originalType=binary&ratio=1&rotation=0&showTitle=false&size=358258&status=done&style=none&taskId=u4fd51c60-5b67-4c53-8434-3688a255aa1&title=&width=1612)<br />示例2：
```yaml
图表类型: Gauge
xcol: stat
ycol: c
query: * | select count(distinct labels['hostname']) as c from (select promql_query('${metricName}{cluster =~ "${cluster}"}') from metrics ) limit 100000
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452839935-6995f428-1e01-4b39-97c6-d23abdb29869.png#averageHue=%231d2125&clientId=uce6e5a51-b8b8-4&from=paste&height=537&id=u24b3081e&originHeight=537&originWidth=1595&originalType=binary&ratio=1&rotation=0&showTitle=false&size=259703&status=done&style=none&taskId=u1efc7ded-a580-42b5-88a3-c2b4a921bf9&title=&width=1595)<br />其他场景：<br />在线图场景下也可以显示为单值图，但不推荐这种写法。
<a name="MMTzd"></a>
### 4.2.2 饼图（Pie）
xcol写法：`pie`<br />ycol写法：`<聚合列>, <数值列>`<br />示例1：
```yaml
图表类型: Pie
xcol: pie
ycol: request_method, c
query: request_method: "$method" |  select count(*) as c, request_method group by request_method
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452843950-4a5c5331-e084-44ed-8a07-e8b3153879e2.png#averageHue=%231d2024&clientId=uce6e5a51-b8b8-4&from=paste&height=633&id=u7d419966&originHeight=633&originWidth=1586&originalType=binary&ratio=1&rotation=0&showTitle=false&size=380780&status=done&style=none&taskId=ubf2a8949-a8d4-437c-9a4d-86892b1beaf&title=&width=1586)<br />示例2：
```yaml
图表类型: Pie
xcol: pie
ycol: http_user_agent, pv
query: * | select count(1) as pv, case when http_user_agent like '%Chrome%' then 'Chrome' when http_user_agent like '%Firefox%' then 'Firefox' when http_user_agent like '%Safari%' then 'Safari' else 'unKnown' end as http_user_agent  group by case when http_user_agent like '%Chrome%' then 'Chrome' when http_user_agent like '%Firefox%' then 'Firefox' when http_user_agent like '%Safari%' then 'Safari' else 'unKnown' end  order by pv desc limit 10
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452844039-c78587cd-10c0-4cbd-ac82-e37f27cb6448.png#averageHue=%231d2228&clientId=uce6e5a51-b8b8-4&from=paste&height=683&id=udd8f290a&originHeight=683&originWidth=1591&originalType=binary&ratio=1&rotation=0&showTitle=false&size=448729&status=done&style=none&taskId=u092d569e-7bdc-4830-ba15-8e5f6fb0798&title=&width=1591)<br />其他场景：<br />Stat图的写法也适用于饼图，也可以展示出效果。
```yaml
图表类型: Pie
xcol: stat
ycol: hostNameNum, ipNum
query: * | select count(distinct labels['hostname']) as hostNameNum, count(distinct labels['ip']) + 20 as ipNum from (select promql_query('${metricName}{cluster =~ ".*"}') from metrics ) limit 100000
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452844161-753fa214-a4d7-4ca4-97df-6410fbadf882.png#averageHue=%231e232b&clientId=uce6e5a51-b8b8-4&from=paste&height=651&id=u1cdbf84d&originHeight=651&originWidth=1605&originalType=binary&ratio=1&rotation=0&showTitle=false&size=405457&status=done&style=none&taskId=u1f7996d6-4296-4393-92ff-b1c1c7dc448&title=&width=1605)
<a name="RD4li"></a>
### 4.2.3 线图（TimeSeries）
xcol写法：`<时间列>`<br />ycol写法：`<数值列> [, <数值列>, ...]`（日志写法）`<labels / 聚合列>#:#<数值列>`（时序库写法或日志聚合写法）<br />示例1：
```yaml
图表类型: Time series
xcol: time
ycol: pv, uv
query: * | select __time__ - __time__ % $${myinterval} as time, COUNT(*)/ 100 as pv, approx_distinct(remote_addr)/ 60 as uv GROUP BY time order by time limit 2000
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452844044-68113362-a064-4056-8144-d43a57e2ad88.png#averageHue=%231d2124&clientId=uce6e5a51-b8b8-4&from=paste&height=653&id=u474df7d4&originHeight=653&originWidth=1607&originalType=binary&ratio=1&rotation=0&showTitle=false&size=341137&status=done&style=none&taskId=uabc516b5-893e-4b5f-85a1-a283ce979db&title=&width=1607)<br />示例2：
```yaml
图表类型: Time series
xcol: time
ycol: labels#:#value
query: * | select time, * from (select promql_query_range('${metricName}') from metrics) limit 1000
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452844195-7f9ab09b-af82-43e1-892c-78514ee89175.png#averageHue=%23222529&clientId=uce6e5a51-b8b8-4&from=paste&height=727&id=u8910ba60&originHeight=727&originWidth=1592&originalType=binary&ratio=1&rotation=0&showTitle=false&size=493022&status=done&style=none&taskId=uc333628e-f73f-480d-8688-138b3ec9e65&title=&width=1592)<br />示例3：<br />也可以通过sql对时序labels进行自定义展示。
```yaml
图表类型: Time series
xcol: time
ycol: customLabelsExtract#:#value
query: * | select concat(labels['ip'], ' -> ', labels['cluster']) as customLabelsExtract, value from (select promql_query_range('${metricName}') from metrics) limit 1000
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452847513-5d390b23-a4e6-4cef-b5df-d012b0ce4acd.png#averageHue=%231e2226&clientId=uce6e5a51-b8b8-4&from=paste&height=600&id=u755370a1&originHeight=600&originWidth=1592&originalType=binary&ratio=1&rotation=0&showTitle=false&size=334958&status=done&style=none&taskId=uf9696cab-0401-45dc-8280-6ba6153db41&title=&width=1592)
<a name="Fo6nH"></a>
### 4.2.4 柱状图（Bar）
xcol写法：`bar`<br />ycol写法：`<聚合列>, <数值列> [, <数值列>, ...]`<br />示例1：
```yaml
图表类型: Bar
xcol: bar
ycol: host, pv, pv2, uv
query: * | select host, COUNT(*)+10 as pv, COUNT(*)+20 as pv2, approx_distinct(remote_addr) as uv GROUP BY host ORDER BY uv desc LIMIT 5
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452847629-d65db0c8-5774-4ae0-9b68-0e6d692b7bd9.png#averageHue=%2322282f&clientId=uce6e5a51-b8b8-4&from=paste&height=657&id=uc81159cd&originHeight=657&originWidth=1601&originalType=binary&ratio=1&rotation=0&showTitle=false&size=455309&status=done&style=none&taskId=uaf99279f-aab7-46a5-b2e1-6983505517a&title=&width=1601)
<a name="teKFg"></a>
### 4.2.5 表格（Table）
xcol写法：`<空>`<br />ycol写法：`<空> 或者 <展示列> [, <展示列>, ...]`<br />示例1：
```yaml
图表类型: Table
xcol: 
ycol: 
query: * | select __time__ - __time__ % 60 as time, COUNT(*)/ 100 as pv, approx_distinct(remote_addr)/ 60 as uv GROUP BY time order by time limit 2000
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452847509-dc9e5599-9440-429e-bece-70536c88f613.png#averageHue=%231e2125&clientId=uce6e5a51-b8b8-4&from=paste&height=658&id=ub7bf1b39&originHeight=658&originWidth=1616&originalType=binary&ratio=1&rotation=0&showTitle=false&size=321931&status=done&style=none&taskId=ub78beb1a-60a3-4bad-ac94-9d5524a2908&title=&width=1616)
<a name="mOEg2"></a>
### 4.2.6 日志（Log）
xcol写法：`<空>`<br />ycol写法：`<空>`<br />示例1：<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452950906-736d4a26-df12-41bd-8c28-055370e04738.png#averageHue=%2325292e&clientId=uce6e5a51-b8b8-4&from=paste&height=511&id=u72528277&originHeight=511&originWidth=1549&originalType=binary&ratio=1&rotation=0&showTitle=false&size=404534&status=done&style=none&taskId=u53b6a673-cc1c-4167-8aae-f7b8a040241&title=&width=1549)
```yaml
图表类型: Logs
xcol: 
ycol: 
query: host: www.vt.mock.com
```
<a name="mpftA"></a>
### 4.2.7 Trace（Traces）
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452996918-7ea26cd7-89a8-4432-bfcf-a2c35b2b9f0a.png#averageHue=%231c2023&clientId=uce6e5a51-b8b8-4&from=paste&id=u9b57ef04&originHeight=981&originWidth=1982&originalType=url&ratio=1&rotation=0&showTitle=false&size=588316&status=done&style=none&taskId=ub53c252b-0bab-40bc-b83c-3fe5e0f15dc&title=)
```yaml
图表类型: Traces
xcol: trace
ycol: 
query: traceID: "f88271003ab7d29ffee1eb8b68c58237"
```
该示例中使用了Trace的日志库，需要您在SLS使用Trace服务，日志服务支持OpenTelemetry Trace数据的原生接入，还支持通过其他Trace系统接入Trace数据。更多细节请参见：[https://help.aliyun.com/document_detail/208894.html](https://help.aliyun.com/document_detail/208894.html)<br />在 Grafana 10.0 以上版本，支持Trace数据的span筛选功能，如果您使用的是低版本Grafana，也可以在query筛选中自定义span的筛选。例如：
```plsql
traceID: "f88271003ab7d29ffee1eb8b68c58237" and resource.deployment.environment : "dev" and service : "web_request" and duration > 10
```
<a name="hDy2z"></a>
### 4.2.8 地图（Map）
xcol写法：`map`<br />ycol写法：`<国家列>, <地理位置列>, <数值列>`<br />示例1：
```yaml
图表类型: GeoMap
xcol: map
ycol: country, geo, pv
query: * | select count(1) as pv ,geohash(ip_to_geo(arbitrary(remote_addr))) as geo,ip_to_country(remote_addr) as country  from log group by country having geo <>'' limit 1000
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452847694-88c5246c-4383-48ac-9ced-a0d4e1d1b444.png#averageHue=%235b8d55&clientId=uce6e5a51-b8b8-4&from=paste&height=697&id=u5effb0ed&originHeight=697&originWidth=1602&originalType=binary&ratio=1&rotation=0&showTitle=false&size=512822&status=done&style=none&taskId=u7a89e2f6-5965-49aa-a18a-936890d8dd0&title=&width=1602)

<a name="Lju6i"></a>
# 5. 一键跳转SLS控制台
注意：此功能在 SLS Grafana Plugin 2.30 版本及以上才可使用。<br />您可以在Explore界面和仪表盘界面中，随时跳转到SLS控制台进行比对了，同时也可以使用SLS控制台更强大的功能和更灵活的日志检索。<br />**跳转到SLS控制台，将附带query、时间信息，无需手动输入。**<br />![ezgif.com-reverse.gif](https://cdn.nlark.com/yuque/0/2023/gif/21832175/1690460534053-e611f07f-9857-4a51-87b1-fd462f148e7f.gif#averageHue=%231a1c20&clientId=uce6e5a51-b8b8-4&from=drop&id=uce74db6d&originHeight=629&originWidth=1382&originalType=binary&ratio=1&rotation=0&showTitle=false&status=done&style=none&taskId=u7a27119f-29f0-44de-a66b-0e54d5fc64e&title=)<br />此种跳转方式是直接跳转到SLS控制台，不涉及任何配置，但需要您的浏览器登录了SLS控制台，否则会显示登录页面。
<a name="Kwo8D"></a>
## 5.1 STS方式跳转（无需登录）
步骤：

1. 访问RAM控制台[https://ram.console.aliyun.com/roles/](https://ram.console.aliyun.com/roles/)，创建一个**_有且只有AliyunLogReadOnlyAccess_**权限策略的角色。建议最大会话时间为3600秒。在基本信息处可复制roleArn信息。<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690460537657-3fffba0b-1ca6-4828-b52b-902b615df880.png#averageHue=%23fbfafa&clientId=uce6e5a51-b8b8-4&from=paste&height=493&id=u66d6a01b&originHeight=493&originWidth=1551&originalType=binary&ratio=1&rotation=0&showTitle=false&size=227311&status=done&style=none&taskId=uebd62af1-0b75-41e4-aa53-9fb78d363ac&title=&width=1551)
2. 访问RAM控制台授权界面[https://ram.console.aliyun.com/permissions](https://ram.console.aliyun.com/permissions)，给在Grafana DataSource配置的AccessKey对应的用户，赋予AliyunRAMReadOnlyAccess、AliyunSTSAssumeRoleAccess权限。（或更换Grafana DataSource配置的AccessKey、AccessSecret，必须保证用户有该权限）<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690460724780-d14baedc-bb84-4a5a-b5db-0438aaee7c84.png#averageHue=%23c6c5c5&clientId=uce6e5a51-b8b8-4&from=paste&height=518&id=udacedc02&originHeight=518&originWidth=1588&originalType=binary&ratio=1&rotation=0&showTitle=false&size=819340&status=done&style=none&taskId=u618da98f-14e5-4fb4-8d3e-f7f5cc04bb4&title=&width=1588)
3. 在DataSource界面配置roleArn。<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690460537368-d7276fb7-d945-4e70-a37c-600b813961fe.png#averageHue=%23202428&clientId=uce6e5a51-b8b8-4&from=paste&height=672&id=u01a31951&originHeight=672&originWidth=809&originalType=binary&ratio=1&rotation=0&showTitle=false&size=217265&status=done&style=none&taskId=u98429cad-f025-4939-8518-7e30c1fd823&title=&width=809)
4. 再次回到Explore界面，尝试gotoSLS按钮，即可免登STS跳转。

注意：如果配置的roleArn有误、权限范围有误，均会导致免登失效，将仍按照一般逻辑进行跳转。<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690460537523-a1e27bdb-345a-4172-a107-78d8d9c6a297.png#averageHue=%232c2d30&clientId=uce6e5a51-b8b8-4&from=paste&height=349&id=u2244cc99&originHeight=349&originWidth=1768&originalType=binary&ratio=1&rotation=0&showTitle=false&size=303309&status=done&style=none&taskId=u593cd0ab-44e3-4494-887b-8d9b0c9495a&title=&width=1768)
<a name="EGhbc"></a>
## 5.2 注意事项和风险警告⚠️
若配置STS跳转，为权限安全考虑，需要满足以下两个条件：

- 配置DataSource的accessKey对应的用户，需要有`AliyunRAMReadOnlyAccess`权限、`AliyunSTSAssumeRoleAccess`权限
- 配置DataSource的roleArn，里面的权限策略，必须**有且只有**`AliyunLogReadOnlyAccess`

原理参考：[控制台内嵌及分享](https://help.aliyun.com/document_detail/74971.html)<br />**如果您配置了免登跳转，请一定要注意数据源是否涉及Grafana仪表盘公开访问的分享，如果涉及公开访问，可能会造成潜在的流量成本上升、潜在的日志内容公开暴露。**
<a name="50d52dd9"></a>
# 6. 持续更新的常见疑问
<a name="p0rew"></a>
## 6.1 为什么在SLS上可以显示图表，在Grafana上显示不了

- 排查xcol、ycol是否配置规范，详情见本文第4章。
- 将xcol、ycol置空，已表格形式查看数据是否正确。
- 排查是否数值列中混有非数字字符或特殊字符。
- Query Inspector中查看是否有数据返回。
- 联系我们查看此问题。
<a name="HMpso"></a>
## 6.2 为什么会产生时间漂移，时间显示不对？
请检查SQL语句中是否包含date_format函数，如果是，则需要在模式串中规范为：`%Y-%m-%d %H:%i:%s`<br />例如出问题的语句为：
```sql
* | SELECT date_format(date_trunc('minute', __time__), '%H:%i') AS time, COUNT(1) AS count, status GROUP BY time, status ORDER BY time
```
需要更改为：
```sql
* | SELECT date_format(date_trunc('minute', __time__), '%Y-%m-%d %H:%i:%s') AS time, COUNT(1) AS count, status GROUP BY time, status ORDER BY time
```
<a name="q9IvH"></a>
## 6.3 如何联系日志服务
钉钉群<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690460944824-6d5fb092-0921-4592-9710-fcdf005a0639.png#averageHue=%23c7c6c6&clientId=uce6e5a51-b8b8-4&from=paste&height=364&id=u09aedbc9&originHeight=1602&originWidth=1242&originalType=binary&ratio=1&rotation=0&showTitle=false&size=436075&status=done&style=none&taskId=u0138acb5-c31f-46df-a140-031f4e345c1&title=&width=282)


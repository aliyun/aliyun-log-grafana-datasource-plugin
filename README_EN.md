![SLS Grafana Plugin Logo.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690462492110-78e3adec-4590-4ed8-9a13-c6ca545a3a25.png#averageHue=%2355310a&clientId=u91823d39-f9eb-4&from=drop&id=u22ff0de1&originHeight=1017&originWidth=2406&originalType=binary&ratio=1&rotation=0&showTitle=false&size=311629&status=done&style=none&taskId=ud80003b6-4819-425a-b71a-1cdf7c275cd&title=)<br />![](https://cdn.nlark.com/yuque/0/2023/svg/21832175/1690467960421-5a835456-c164-48c0-afc7-f440e7d8f0ba.svg#clientId=u48c6b2ce-629b-4&id=grEbM&originHeight=20&originWidth=60&originalType=binary&ratio=1&rotation=0&showTitle=false&status=done&style=none&taskId=u71feb3fe-85b6-459a-bfc1-6edaf506cd4&title=) ![](https://cdn.nlark.com/yuque/0/2023/svg/21832175/1690467960433-386f2dfa-1c97-44b6-88dd-f3ff4c04858b.svg#clientId=u48c6b2ce-629b-4&id=nAWkL&originHeight=20&originWidth=110&originalType=binary&ratio=1&rotation=0&showTitle=false&status=done&style=none&taskId=u01e0568b-1821-4e0e-a787-fea3396ec9d&title=) ![](https://cdn.nlark.com/yuque/0/2023/svg/21832175/1690467960408-8d0bd40a-aede-4f88-b33e-e5ccf6cdf71e.svg#clientId=u48c6b2ce-629b-4&id=t80B3&originHeight=20&originWidth=76&originalType=binary&ratio=1&rotation=0&showTitle=false&status=done&style=none&taskId=u11a46765-f095-47d7-b67e-fbc0207a8d1&title=) ![](https://cdn.nlark.com/yuque/0/2023/svg/21832175/1690467961033-bd93fd7c-2aca-494b-bd94-77e8beb3e5e1.svg#clientId=u48c6b2ce-629b-4&id=z18xX&originHeight=20&originWidth=170&originalType=binary&ratio=1&rotation=0&showTitle=false&status=done&style=none&taskId=ue71c4b7a-ad08-4516-b696-31afab5f93c&title=)<br />[简体中文](https://github.com/aliyun/aliyun-log-grafana-datasource-plugin/blob/master/README.md) | **English** | [旧版README](https://github.com/aliyun/aliyun-log-grafana-datasource-plugin/blob/master/old_README_CN.md) | [Old Version README](https://github.com/aliyun/aliyun-log-grafana-datasource-plugin/blob/master/old_README.md)
<a name="eI2lI"></a>
# 1. Overview 
Log service is a cloud-native observation and analysis platform that provides large-scale, low-cost, and real-time platform services for Log, Metric, and Trace data. One-stop data collection, processing, analysis, alarm visualization, and delivery are provided to improve the digitization capabilities of R &amp; D, O &amp; M, operations, and security scenarios. [Official documentation ](https://www.aliyun.com/product/sls)<br />this repository is an Alibaba Cloud log service Grafana data source plug-in. Before using this plug-in, you must use log service products and have at least one LogStore configured for collection.
<a name="zsCS1"></a>
# 2. Installation

## Version compatibility

- Plugin **2.39.1** fixes logstore selection and `goto SLS` returning 404 after Grafana 13 disabled numeric-ID data source APIs by default.
- The highest Grafana version verified for this release is **13.0.2**, with regression checks on **11.4.0**. Checks use simulated SLS responses and cover data source configuration, logstore selection, queries, and `goto SLS` resource calls. Later versions have not been verified; validate them in a test environment before upgrading.
- Plugin **2.39** exhibits these failures on **13.0.2** with default settings. Upgrade the plugin when using Grafana 13. The 11.4.0 comparison does not imply that every later version is incompatible.
- Resource calls continue to use the runtime API available in Grafana 8.x and later. A minimum Grafana version declaration is not a guarantee of compatibility with future versions.

Grafana 8.0 and later is required. For Grafana versions earlier than 8.0, use plugin version 1.0. <br /> [download from Release ](https://github.com/aliyun/aliyun-log-grafana-datasource-plugin/releases). Go to the grafana plug-in directory and modify the configuration file. `[plugins] `node, set `allow_loading_unsigned_plugins = aliyun-log-service-datasource `, and then restart the grafana.

- mac 
   - Plug-In Directory： `/usr/local/var/lib/grafana/plugins`
   - Configuration file location：`/usr/local/etc/grafana/grafana.ini`
   - Restart command：`brew services restart grafana`
- YUM/RPM
   - Plug-In Directory：`/var/lib/grafana/plugins`
   - Configuration file location：`/etc/grafana/grafana.ini`
   - Restart command：`systemctl restart grafana-server`
- .tar.gz 
   - Plug-In Directory：`{PATH_TO}/grafana-{x}.{x}.{x}/data/plugins`
   - Configuration file location：`{PATH_TO}/grafana-{x}.{x}.{x}/conf/defaults.ini`
   - Restart command：`./bin/grafana-server web`
<a name="t92KT"></a>
# 3. Add a data source 
<a name="YihL2"></a>
## 3.1 logstore configuration 
on the datasource page, add an SLS log-service-datasource data source. In the data source management panel, add a LogService data source. In the settings panel, set the Endpoint project for your log service endpoint ( endpoint can see it on the overview page of the project. For more information, see [service entry ](https://help.aliyun.com/document_detail/29008.htm#reference-wgx-pwq-zdb)). For example, if your project is in the qingdao region, enter the Url `cn-qingdao.log.aliyuncs.com `. Set the Project and logstore as needed, and set the AccessId and AccessKeySecret. It is best to configure the accesskey of the sub-account. To ensure data security, the AK is saved and cleared without Echo.
<a name="oxIOY"></a>
## 3.2 Time series database configuration (using SLS plug-in, supports SQL query and combination operator query) 
the time series library can also be configured as an SLS plug-in. With this access method, you can query the time series Library by SQL or by using the PromQL operator. For more information, see: [description of time series query syntax ](https://help.aliyun.com/document_detail/171763.html)). The configuration method is the same as that in Section 1.1. Enter the name of the LogStore in the metricStore.
<a name="YLNV4"></a>
## 3.3 Time series Library configuration (native mode, using Prometheus plug-in) 
the SLS time series Library supports native Prometheus query, so you can directly use the native Prometheus plug-in to access the data source. Please refer to the following [official documentation ](https://help.aliyun.com/document_detail/173903.html)configure the data source. The URL of the log service MetricStore in the format **https://{project}. {sls-endpoint}/prometheus/{project}/{metricstore} **. Where _{sls-endpoint} _the Project of the region where the Endpoint is located. For more information, see [service entry ](https://help.aliyun.com/document_detail/29008.htm#reference-wgx-pwq-zdb), _{project} _and _{metricstore} _replace the Project and Metricstore of the created Log service with the actual value. For example: **https://sls-prometheus-test.cn-hangzhou.log.aliyuncs.com/prometheus/sls-prometheus-test/prometheus**<br />**From the perspective of best practices, you can add Prometheus data sources and SLS data sources to the time series database at the same time. You can use different statement query methods according to your personal habits. We recommend that you use SLS data sources in Variable (you can convert them to SLS dashboards)**
<a name="pw5VJ"></a>
# 4. Use 
<a name="x57r9"></a>
## 4.1 dashboard variables 
<a name="B6Nkt"></a>
### 4.1.1 Writing Grafana variables in query 
in practice, the only recommended writing method is: `${var_name} `. <br />Theoretically, Grafana supports three writing methods: `$varname `, `${var_name} `, `[[varname]] `. However, if you do not add parentheses, the variable name range may be incorrectly identified, and the double-Middle plus sign will be gradually abandoned in the future. [Reference source ](https://grafana.com/docs/grafana/latest/dashboards/variables/variable-syntax/).
<a name="oJtey"></a>
### 4.1.2 General usage and all-selected configuration 
for most SLS DataSource, you can use SQL statements to query values as variables. Procedure: 

- go to Grafana dashboard settings-Variables 
- select Query type, set Datasource to corresponding LogStore, and write query 
- view the results in the Preview of values

![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452672650-76a5a639-4b38-4f1f-b99b-2944b31969a8.png#averageHue=%231e2226&clientId=uce6e5a51-b8b8-4&from=paste&height=133&id=uf3b6bc2a&originHeight=196&originWidth=713&originalType=binary&ratio=1&rotation=0&showTitle=false&size=59106&status=done&style=none&taskId=ube8a64f4-cde0-4b86-a596-bb4f0b38f3f&title=&width=484)<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452672698-b4a5a738-92fc-4ffe-a73f-cd372bfeee50.png#averageHue=%23252830&clientId=uce6e5a51-b8b8-4&from=paste&height=121&id=u586769d4&originHeight=288&originWidth=846&originalType=binary&ratio=2&rotation=0&showTitle=false&size=100765&status=done&style=none&taskId=u5fa8016a-0107-4975-9c04-cfd879ae0c0&title=&width=354)<br />Generally, if you use it as a logstore filter, we recommend that you open it in practice. `All Option `and configure the Custom all value `* `.<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452672720-4ea38ea1-e5d2-4e86-9095-3407c0db57b0.png#averageHue=%23212328&clientId=uce6e5a51-b8b8-4&from=paste&height=202&id=u42ba9285&originHeight=404&originWidth=970&originalType=binary&ratio=2&rotation=0&showTitle=false&size=152094&status=done&style=none&taskId=u6882b17d-442d-4706-a152-ff8a6e5e9fa&title=&width=485)<br />In this way, Dashboard Query statement is written as follows: `* and request_method: ${var} | select * from log `you can select Variable and select all.
<a name="COVuD"></a>
### 4.1.3 Variable of time series data in SLS plug-in
As mentioned in the configuration of SLS storage data sources, SLS time series libraries can be configured Prometheus either native or SLS plug-ins. If you use the SLS plug-in, you usually need to use the promql operator. For more information about the usage and syntax, see: [overview of time series data query and analysis ](https://help.aliyun.com/document_detail/171763.html). <br />The following example shows how to obtain the names of all metrics in the time series database.<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452716829-b073e6cb-d312-49c6-83cc-80c756bb8ad4.png#averageHue=%231e2225&clientId=uce6e5a51-b8b8-4&from=paste&height=185&id=u84f192d7&originHeight=185&originWidth=774&originalType=binary&ratio=1&rotation=0&showTitle=false&size=61040&status=done&style=none&taskId=u3648391c-6096-466a-a907-c6b841c2d6d&title=&width=774)<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452716951-7eb02e94-462a-498a-af42-e4ed5e19faac.png#averageHue=%2327292e&clientId=uce6e5a51-b8b8-4&from=paste&height=179&id=ub46e3930&originHeight=358&originWidth=1730&originalType=binary&ratio=2&rotation=0&showTitle=false&size=236336&status=done&style=none&taskId=u3422f0da-f342-4670-ae99-a127b016e49&title=&width=865)
<a name="sG2qH"></a>
### 4.1.4 Use Interval variables to control hitting based on time span 
configure a Grafana time variable

| Name | Variable name, such as myinterval. The name is the variable used in your configuration. In this case, the myinterval. **$$myinterval **.  |
| --- | --- |
| Type | Select **Interval **.  |
| Label | Configure **time interval **.  |
| Values | Configure **1 m, 10 m, 30 m, 1 h, 6 h, 12 h, 1d, 7d, 14d, 30d **.  |
| Auto Option | Open **Auto Option **switch, other parameters remain the default configuration. |

**Note: this time variable is different from Grafana variable in SLS. You need to add another variable before writing the normal variable. **`**$ **`**to correctly convert Interval in SLS statements.**
```yaml
chartType: TimeSeries
xcol: time
ycol: pv, uv
query: * | select __time__ - __time__ % $${myinterval} as time, COUNT(*)/ 100 as pv, approx_distinct(remote_addr)/ 60 as uv GROUP BY time order by time limit 2000
```
In the configuration `1m `when:<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452762876-174cd1ac-ef8e-4b2f-ac31-d833e35bfb7e.png#averageHue=%231d2124&clientId=uce6e5a51-b8b8-4&from=paste&height=770&id=u752a89a5&originHeight=770&originWidth=1622&originalType=binary&ratio=1&rotation=0&showTitle=false&size=413040&status=done&style=none&taskId=u786e32f5-a286-4f5f-8643-64380ba4406&title=&width=1622)<br />In the configuration `10m `when:<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452762804-820d2b64-1719-4af0-b340-99b4be4a33f7.png#averageHue=%231d2124&clientId=uce6e5a51-b8b8-4&from=paste&height=774&id=u00307656&originHeight=774&originWidth=1634&originalType=binary&ratio=1&rotation=0&showTitle=false&size=395508&status=done&style=none&taskId=u3311b85a-692d-4458-a6e4-d823026d2c5&title=&width=1634)<br />When the time range of the dashboard is very long, this function can easily control the time hitting density.
<a name="sPAdc"></a>
## 4.2 Overview of chart standard writing
<a name="MkJaN"></a>
### 4.2.1 Stat / Gauge
xcol: `stat`<br />ycol: `<Numeric column>, <Numeric column>`<br />Note: If the required numeric column is a non-numeric column, it will be set to 0. <br />Example 1:
```yaml
chartType: Stat
xcol: stat
ycol: PV, deltaPercent
query: * | select diff[1] as "PV", round((diff[1] - diff[2])/diff[2] * 100, 2) as deltaPercent from (select compare("PV", 86400) as diff from (select count(*) as "PV" from log))
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452840023-cd2747db-f1d1-4a50-adb8-9ae3555d1ac1.png#averageHue=%231e2125&clientId=uce6e5a51-b8b8-4&from=paste&height=589&id=u133f3dbf&originHeight=589&originWidth=1612&originalType=binary&ratio=1&rotation=0&showTitle=false&size=358258&status=done&style=none&taskId=u4fd51c60-5b67-4c53-8434-3688a255aa1&title=&width=1612)<br />Example 2：
```yaml
chartType: Gauge
xcol: stat
ycol: c
query: * | select count(distinct labels['hostname']) as c from (select promql_query('${metricName}{cluster =~ "${cluster}"}') from metrics ) limit 100000
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452839935-6995f428-1e01-4b39-97c6-d23abdb29869.png#averageHue=%231d2125&clientId=uce6e5a51-b8b8-4&from=paste&height=537&id=u24b3081e&originHeight=537&originWidth=1595&originalType=binary&ratio=1&rotation=0&showTitle=false&size=259703&status=done&style=none&taskId=u1efc7ded-a580-42b5-88a3-c2b4a921bf9&title=&width=1595)<br />Other scenarios: <br />online Graph scenarios can also be displayed as single-value graphs, but this writing method is not recommended.
<a name="MMTzd"></a>
### 4.2.2 Pie
xcol: `pie`<br />ycol: `<Aggregate columns>, <Numeric column>`<br />Example 1:
```yaml
chartType: Pie
xcol: pie
ycol: request_method, c
query: request_method: "$method" |  select count(*) as c, request_method group by request_method
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452843950-4a5c5331-e084-44ed-8a07-e8b3153879e2.png#averageHue=%231d2024&clientId=uce6e5a51-b8b8-4&from=paste&height=633&id=u7d419966&originHeight=633&originWidth=1586&originalType=binary&ratio=1&rotation=0&showTitle=false&size=380780&status=done&style=none&taskId=ubf2a8949-a8d4-437c-9a4d-86892b1beaf&title=&width=1586)<br />Example 2:
```yaml
chartType: Pie
xcol: pie
ycol: http_user_agent, pv
query: * | select count(1) as pv, case when http_user_agent like '%Chrome%' then 'Chrome' when http_user_agent like '%Firefox%' then 'Firefox' when http_user_agent like '%Safari%' then 'Safari' else 'unKnown' end as http_user_agent  group by case when http_user_agent like '%Chrome%' then 'Chrome' when http_user_agent like '%Firefox%' then 'Firefox' when http_user_agent like '%Safari%' then 'Safari' else 'unKnown' end  order by pv desc limit 10
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452844039-c78587cd-10c0-4cbd-ac82-e37f27cb6448.png#averageHue=%231d2228&clientId=uce6e5a51-b8b8-4&from=paste&height=683&id=udd8f290a&originHeight=683&originWidth=1591&originalType=binary&ratio=1&rotation=0&showTitle=false&size=448729&status=done&style=none&taskId=u092d569e-7bdc-4830-ba15-8e5f6fb0798&title=&width=1591)<br />Other scenarios: <br />the Stat chart is also applicable to pie charts and can also show the effect.
```yaml
chartType: Pie
xcol: stat
ycol: hostNameNum, ipNum
query: * | select count(distinct labels['hostname']) as hostNameNum, count(distinct labels['ip']) + 20 as ipNum from (select promql_query('${metricName}{cluster =~ ".*"}') from metrics ) limit 100000
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452844161-753fa214-a4d7-4ca4-97df-6410fbadf882.png#averageHue=%231e232b&clientId=uce6e5a51-b8b8-4&from=paste&height=651&id=u1cdbf84d&originHeight=651&originWidth=1605&originalType=binary&ratio=1&rotation=0&showTitle=false&size=405457&status=done&style=none&taskId=u1f7996d6-4296-4393-92ff-b1c1c7dc448&title=&width=1605)
<a name="RD4li"></a>
### 4.2.3 TimeSeries
xcol: `<Time column>`<br />ycol: `<Numeric column> [, <Numeric column>, ...]`(LogStore)`<labels / Aggregate columns>#:#<Numeric column>`(Write Time series Library or log aggregation)<br />Example 1:
```yaml
chartType: Time series
xcol: time
ycol: pv, uv
query: * | select __time__ - __time__ % $${myinterval} as time, COUNT(*)/ 100 as pv, approx_distinct(remote_addr)/ 60 as uv GROUP BY time order by time limit 2000
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452844044-68113362-a064-4056-8144-d43a57e2ad88.png#averageHue=%231d2124&clientId=uce6e5a51-b8b8-4&from=paste&height=653&id=u474df7d4&originHeight=653&originWidth=1607&originalType=binary&ratio=1&rotation=0&showTitle=false&size=341137&status=done&style=none&taskId=uabc516b5-893e-4b5f-85a1-a283ce979db&title=&width=1607)<br />Example 2:
```yaml
chartType: Time series
xcol: time
ycol: labels#:#value
query: * | select time, * from (select promql_query_range('${metricName}') from metrics) limit 1000
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452844195-7f9ab09b-af82-43e1-892c-78514ee89175.png#averageHue=%23222529&clientId=uce6e5a51-b8b8-4&from=paste&height=727&id=u8910ba60&originHeight=727&originWidth=1592&originalType=binary&ratio=1&rotation=0&showTitle=false&size=493022&status=done&style=none&taskId=uc333628e-f73f-480d-8688-138b3ec9e65&title=&width=1592)<br />Example 3: <br />you can also use SQL to customize time series labels.
```yaml
chartType: Time series
xcol: time
ycol: customLabelsExtract#:#value
query: * | select concat(labels['ip'], ' -> ', labels['cluster']) as customLabelsExtract, value from (select promql_query_range('${metricName}') from metrics) limit 1000
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452847513-5d390b23-a4e6-4cef-b5df-d012b0ce4acd.png#averageHue=%231e2226&clientId=uce6e5a51-b8b8-4&from=paste&height=600&id=u755370a1&originHeight=600&originWidth=1592&originalType=binary&ratio=1&rotation=0&showTitle=false&size=334958&status=done&style=none&taskId=uf9696cab-0401-45dc-8280-6ba6153db41&title=&width=1592)
<a name="Fo6nH"></a>
### 4.2.4 Bar
xcol: `bar`<br />ycol: `<Aggregate columns>, <Numeric column> [, <Numeric column>, ...]`<br />Example 1:
```yaml
chartType: Bar
xcol: bar
ycol: host, pv, pv2, uv
query: * | select host, COUNT(*)+10 as pv, COUNT(*)+20 as pv2, approx_distinct(remote_addr) as uv GROUP BY host ORDER BY uv desc LIMIT 5
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452847629-d65db0c8-5774-4ae0-9b68-0e6d692b7bd9.png#averageHue=%2322282f&clientId=uce6e5a51-b8b8-4&from=paste&height=657&id=uc81159cd&originHeight=657&originWidth=1601&originalType=binary&ratio=1&rotation=0&showTitle=false&size=455309&status=done&style=none&taskId=uaf99279f-aab7-46a5-b2e1-6983505517a&title=&width=1601)
<a name="teKFg"></a>
### 4.2.5 Table
xcol: `<empty>`<br />ycol: `<empty> or <Display Column> [, <Display Column>, ...]`<br />Example 1:
```yaml
chartType: Table
xcol: 
ycol: 
query: * | select __time__ - __time__ % 60 as time, COUNT(*)/ 100 as pv, approx_distinct(remote_addr)/ 60 as uv GROUP BY time order by time limit 2000
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452847509-dc9e5599-9440-429e-bece-70536c88f613.png#averageHue=%231e2125&clientId=uce6e5a51-b8b8-4&from=paste&height=658&id=ub7bf1b39&originHeight=658&originWidth=1616&originalType=binary&ratio=1&rotation=0&showTitle=false&size=321931&status=done&style=none&taskId=ub78beb1a-60a3-4bad-ac94-9d5524a2908&title=&width=1616)
<a name="mOEg2"></a>
### 4.2.6 Log
xcol: `<空>`<br />ycol: `<空>`<br />Example 1:<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452950906-736d4a26-df12-41bd-8c28-055370e04738.png#averageHue=%2325292e&clientId=uce6e5a51-b8b8-4&from=paste&height=511&id=u72528277&originHeight=511&originWidth=1549&originalType=binary&ratio=1&rotation=0&showTitle=false&size=404534&status=done&style=none&taskId=u53b6a673-cc1c-4167-8aae-f7b8a040241&title=&width=1549)
```yaml
chartType: Logs
xcol: 
ycol: 
query: host: www.vt.mock.com
```
<a name="mpftA"></a>
### 4.2.7 Traces
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452996918-7ea26cd7-89a8-4432-bfcf-a2c35b2b9f0a.png#averageHue=%231c2023&clientId=uce6e5a51-b8b8-4&from=paste&id=u9b57ef04&originHeight=981&originWidth=1982&originalType=url&ratio=1&rotation=0&showTitle=false&size=588316&status=done&style=none&taskId=ub53c252b-0bab-40bc-b83c-3fe5e0f15dc&title=)
```yaml
chartType: Traces
xcol: trace
ycol: 
query: traceID: "f88271003ab7d29ffee1eb8b68c58237"
```
In this example, the Trace logstore is used. You need to use the Trace service in SLS. Log service supports native access to OpenTelemetry Trace data, and supports access to Trace data through other Trace systems. For more information, see: [https://help.aliyun.com/document_detail/208894.html ](https://help.aliyun.com/document_detail/208894.html)<br />in Grafana 10.0 and later versions, the span filtering function of Trace data is supported. If you are using a lower version Grafana, you can also customize the span filtering function in query filtering. For example:
```plsql
traceID: "f88271003ab7d29ffee1eb8b68c58237" and resource.deployment.environment : "dev" and service : "web_request" and duration > 10
```
<a name="hDy2z"></a>
### 4.2.8 Map
xcol: `map`<br />ycol: `<Country column>, <Geographic location column>, <Numeric column>`<br />Example 1:
```yaml
chartType: GeoMap
xcol: map
ycol: country, geo, pv
query: * | select count(1) as pv ,geohash(ip_to_geo(arbitrary(remote_addr))) as geo,ip_to_country(remote_addr) as country  from log group by country having geo <>'' limit 1000
```
![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690452847694-88c5246c-4383-48ac-9ced-a0d4e1d1b444.png#averageHue=%235b8d55&clientId=uce6e5a51-b8b8-4&from=paste&height=697&id=u5effb0ed&originHeight=697&originWidth=1602&originalType=binary&ratio=1&rotation=0&showTitle=false&size=512822&status=done&style=none&taskId=u7a89e2f6-5965-49aa-a18a-936890d8dd0&title=&width=1602)

<a name="Lju6i"></a>
# 5. One-click jump to the SLS console
Note: This feature is only available SLS Grafana Plugin version 2.30 and later. You can jump to the SLS console at any time on the Explore and dashboard interfaces. You can also use the more powerful functions and flexible log retrieval of the SLS console. <br />**Jump to the SLS console, with query and time information, without manual input.**<br />![ezgif.com-reverse.gif](https://cdn.nlark.com/yuque/0/2023/gif/21832175/1690460534053-e611f07f-9857-4a51-87b1-fd462f148e7f.gif#averageHue=%231a1c20&clientId=uce6e5a51-b8b8-4&from=drop&id=uce74db6d&originHeight=629&originWidth=1382&originalType=binary&ratio=1&rotation=0&showTitle=false&status=done&style=none&taskId=u7a27119f-29f0-44de-a66b-0e54d5fc64e&title=)<br />This method is to jump directly to the SLS console without any configuration. However, you must log on to the SLS console with your browser. Otherwise, the login page will be displayed.
<a name="Kwo8D"></a>
## 5.1 STS redirection (no logon required)
Procedure: 

1. access the RAM console [https://ram.console.aliyun.com/roles/ ](https://ram.console.aliyun.com/roles/), create a **_yes and only AliyunLogReadOnlyAccess _**the role of the policy. The recommended maximum session time is 3600 seconds. You can copy roleArn information in the basic information section.<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690460537657-3fffba0b-1ca6-4828-b52b-902b615df880.png#averageHue=%23fbfafa&clientId=uce6e5a51-b8b8-4&from=paste&height=493&id=u66d6a01b&originHeight=493&originWidth=1551&originalType=binary&ratio=1&rotation=0&showTitle=false&size=227311&status=done&style=none&taskId=uebd62af1-0b75-41e4-aa53-9fb78d363ac&title=&width=1551)
2. Access the RAM console authorization interface [https://ram.console.aliyun.com/permissions ](https://ram.console.aliyun.com/permissions)to grant Grafana DataSource and AccessKey permissions to the user corresponding to the **AliyunRAMReadOnlyAccess** configured in the **AliyunSTSAssumeRoleAccess**. (Or change the Grafana DataSource and AccessKey configured by the AccessSecret. You must ensure that the user has this permission.)<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690460724780-d14baedc-bb84-4a5a-b5db-0438aaee7c84.png#averageHue=%23c6c5c5&clientId=uce6e5a51-b8b8-4&from=paste&height=518&id=udacedc02&originHeight=518&originWidth=1588&originalType=binary&ratio=1&rotation=0&showTitle=false&size=819340&status=done&style=none&taskId=u618da98f-14e5-4fb4-8d3e-f7f5cc04bb4&title=&width=1588)
3. On the DataSource page, configure the roleArn.<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690460537368-d7276fb7-d945-4e70-a37c-600b813961fe.png#averageHue=%23202428&clientId=uce6e5a51-b8b8-4&from=paste&height=672&id=u01a31951&originHeight=672&originWidth=809&originalType=binary&ratio=1&rotation=0&showTitle=false&size=217265&status=done&style=none&taskId=u98429cad-f025-4939-8518-7e30c1fd823&title=&width=809)
4. Return to the Explore interface again and try to gotoSLS the button to avoid STS redirection. 

Note: If the configured roleArn is incorrect or the permission range is incorrect, the logon-free function will become invalid and will be redirected according to the general logic.<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690460537523-a1e27bdb-345a-4172-a107-78d8d9c6a297.png#averageHue=%232c2d30&clientId=uce6e5a51-b8b8-4&from=paste&height=349&id=u2244cc99&originHeight=349&originWidth=1768&originalType=binary&ratio=1&rotation=0&showTitle=false&size=303309&status=done&style=none&taskId=u593cd0ab-44e3-4494-887b-8d9b0c9495a&title=&width=1768)
<a name="EGhbc"></a>
## 5.2 Precautions and risk warnings⚠️ 
if STS redirection is configured, the following conditions must be met for permission security: 

- the user corresponding to the DataSource of the configuration accessKey. `AliyunRAMReadOnlyAccess `permissions, `AliyunSTSAssumeRoleAccess `permission 
- configure the DataSource of the roleArn. The policy must be **yes and only **`AliyunLogReadOnlyAccess `

principle reference: [embedded and shared in the console](https://help.aliyun.com/document_detail/74971.html)<br />**If you configure no-logon redirection, be sure to check whether the data source involves sharing Grafana public access to the dashboard. If public access is involved, potential traffic costs may rise and potential log content may be exposed.**
<a name="50d52dd9"></a>
# 6. Frequently Asked Questions about continuous updates 
<a name="p0rew"></a>
## 6.1 Why can charts be displayed on SLS but not on Grafana 

- check whether xcol and ycol are configured properly. For more information, see chapter 4. 
- Leave xcol and ycol blank and check whether the data is correct in tabular form. 
- Check whether the numeric column contains non-numeric characters or special characters. 
- Check whether data is returned in the Query Inspector. 
- Contact us to check this issue.
<a name="HMpso"></a>
## 6.2 Why does time drift occur and the time display is incorrect? 
Check whether the SQL statement contains the date_format function. If yes, specify the following code in the pattern string:`%Y-%m-%d %H:%i:%s`<br />For example, the error statement is as follows:
```sql
* | SELECT date_format(date_trunc('minute', __time__), '%H:%i') AS time, COUNT(1) AS count, status GROUP BY time, status ORDER BY time
```
Change to:
```sql
* | SELECT date_format(date_trunc('minute', __time__), '%Y-%m-%d %H:%i:%s') AS time, COUNT(1) AS count, status GROUP BY time, status ORDER BY time
```
<a name="q9IvH"></a>
## 6.3 How can I contact log service? 
DingTalk Group<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690460944824-6d5fb092-0921-4592-9710-fcdf005a0639.png#averageHue=%23c7c6c6&clientId=uce6e5a51-b8b8-4&from=paste&height=364&id=u09aedbc9&originHeight=1602&originWidth=1242&originalType=binary&ratio=1&rotation=0&showTitle=false&size=436075&status=done&style=none&taskId=u0138acb5-c31f-46df-a140-031f4e345c1&title=&width=282)


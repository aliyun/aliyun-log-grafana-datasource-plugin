# Changelog

# 2.39.1 (2026-09-09)

- 修复 Grafana 13 默认停用数字 ID 数据源接口后，日志库列表和 goto SLS 请求返回 404 的问题；通过 Grafana 运行时资源 API 保持旧版兼容性。
- 补充资源调用回归测试和 Grafana 版本兼容性说明，最高已验证版本为 13.0.2，并回归验证 11.4.0。

# 2.39 (2025-12-12)
- 支持 “gotoSLS” 跳转国际站 project
  - 数据源配置页面新增 International 字段，用于跳转到国际站。
    ![international字段](./img/2.39/international_2.39.jpg)

# 2.38 (2025-03-03)
- 修复时序库跳转 SLS 路径问题。
- 日志输出级别整理，减少无用输出。
- 支持‘增强Sql’查询，参考 [开启SQL独享版](https://help.aliyun.com/zh/sls/user-guide/enable-dedicated-sql/?spm=a2c4g.11186623.0.0.2c7771dclfhQou)

# 2.37 (2025-01-13)

- 修复 grafana 内部初始化 currentPage 的问题。
- 修复自定义 logstore 在日志库列表展示不回显的问题；修复 gotoSLS 跳转到 defaultLogstore 的问题。

# 2.36 (2024-11-25)

- 支持 v4 签名，提供更高的安全性。
  - 数据源配置页面新增 Region 字段，兼容非受信环境独立签名。
    ![region字段](./img/2.36/region_2.36.png)
- 变量替换支持自定义'日志库列表'选项。
  - Variables 设置页面选择 'Custom' 类型。'name' 字段作为独立标识，必须包含 logstore 字符串，不区分大小写。'Custom options' 手动填写可选变量，以逗号区分。
    ![variable配置](./img/2.36/variable_2.36.jpg)
  - Panel 界面修改'日志库列表'选项为自定义Variable，修改相应变量并刷新仪表盘，即可获得最新结果。
    ![自定义logstore](./img/2.36/custom.jpg)
    

# 2.35 (2024-10-21)

优化查询分析接口性能。

# 2.34 (2024-09-23)

- 优化 table/log 视图
  - 支持当存在纳秒字段时, time 按纳秒精度排序。
  - 支持修改总查询记录数 totalLogs, 默认值为 100, 最小值为 1, 最大值为 5000, 仅对 Query 查询语句生效, 分析语句无效
    ![image.png](./img/2.34/totalLogs_2.34.png)
- 修复 gotoSLS 仅可跳转至数据源中设置的默认 Logstore 的问题
- 修复 gotoSLS 跳转后变量替换为具体值的问题

# 2.33 (2024-09-01)

- 插件配置界面优化,必填项为 `Endpoint Project AccessKeyID AccessKeySecret`, 不再强制要求填写 `LogStore`;如果不填写 `LogStore`, 请确保你的填写的 Ak 具备当前 Project 的 ListProject 权限，参考 [ListLogStores - 列出 LogStore](https://help.aliyun.com/zh/sls/developer-reference/api-sls-2020-12-30-listlogstores)
- 插件支持自定义 Headers, 但仅在数据源类型为`MetricStore(PrmQL)` 生效
  ![图 2](./img/2.33/config_editor_2.33.png)
- 插件编辑界面优化,支持切换数据源类型和在编辑界面下拉选择具体[日志库(Logstore)](https://help.aliyun.com/zh/sls/product-overview/logstore)或[时序库(MetricStore)](https://help.aliyun.com/zh/sls/product-overview/metricstore?spm=a2c4g.11186623.0.0.2e8b60d1YQznSr)
  ![图 3](./img/2.33/query_editor_2.33.png)

  - 数据源类型主要是两种语法区别`SQL` 和 `PromQL`,再加上存储库的类型不同，有四种类型可选: `ALL(SQL)`、`Logstore(SQL)`、`MetricStore(SQL)`、`MetricStore(PromQL)` ;
  - 日志库(Logstore)支持`SQL`[日志库(Logstore)查询分析](https://help.aliyun.com/zh/sls/user-guide/search-and-analysis);
  - 时序库(MetricStore)支持`SQL + PromQL`[时序库(MetricStore)查询分析](https://help.aliyun.com/zh/sls/user-guide/search-and-analysis);
  - `MetricStore(PromQL)`支持添加 `custom Headers`,具体在该数据源的配置界面进行添加；

# 2.32 (2023-04-16)

- 支持时序图断点展示，通过 time_series(`时间列`, `补全窗口`, `%Y-%m-%d %H:%i:%s`, `null`) 补 null 值进行断点展示
  - time*series 参考：[time_series 时间补全函数说明](https://help.aliyun.com/zh/sls/user-guide/date-and-time-functions-1?spm=a2c4g.11186623.0.i10#section-wsz-wt2-4fb)
    ![image.png](https://img.alicdn.com/imgextra/i2/O1CN01HzGO5d1Xg48MpyJHS*!!6000000002952-0-tps-3116-1144.jpg)

# 2.31 (2023-10-18)

- 支持日志图在 Grafana v10 的展示，日志图可以指定想展示的字段，ycol 格式为`字段1,字段2`
- 流图支持多条线，ycol 格式为`#:#指标1,指标2`
- 修改表格区域的名字为 ycol，使多个表格在同一个图表时，可以区分。

![image.png](https://cdn.nlark.com/yuque/0/2023/png/182537/1697599988480-6ddc340d-9b64-4531-abbf-748eff4c0df9.png#averageHue=%231e2126&clientId=ue09f9ab6-6409-4&from=paste&height=480&id=u63975600&originHeight=960&originWidth=1630&originalType=binary&ratio=2&rotation=0&showTitle=false&size=363161&status=done&style=none&taskId=u01fdf5fe-b863-44da-b5f8-a3ae4512280&title=&width=815)

![image.png](https://cdn.nlark.com/yuque/0/2023/png/182537/1697600032996-be73fb93-a181-4f56-a42d-9ebdba626227.png#averageHue=%23202327&clientId=ue09f9ab6-6409-4&from=paste&height=515&id=u8bdbba62&originHeight=1030&originWidth=2130&originalType=binary&ratio=2&rotation=0&showTitle=false&size=562357&status=done&style=none&taskId=uf3ac8398-be19-4a16-abc7-740eebef31c&title=&width=1065)

- Supports the display of log graphs in Grafana v10. Log graphs can specify the fields you want to display. The ycol format is `Field 1, Field 2`
- The flow graph supports multiple lines, and the ycol format is `#:#Indicator 1,Indicator 2`
- Modify the name of the table area to ycol so that multiple tables can be distinguished when they are in the same chart.

# 2.30 (2023-07-25)

- 优化 SLS Grafana 插件后端结构，现支持加入自定义 Resource API 功能。
- 引入 gotoSLS 功能，用户可以方便地跳转到 SLS 控制台进行查询，和体验 SLS 控制台更强大的功能，跳转附带当前 Grafana 的 query、时间信息。在 DataSource 界面配置 roleArn 可实现 STS 跳转，若不配置，则按照正常直接访问逻辑跳转（需要登录控制台）。
  - **注意：若配置 STS 跳转，为权限安全考虑，需要满足以下两个条件**：
    - **配置 DataSource 的 accessKey 对应的用户，需要有**`**AliyunRAMReadOnlyAccess**`**权限**
    - **配置 DataSource 的 roleArn，里面的权限策略，必须有且只有**`**AliyunLogReadOnlyAccess**`
  - 原理参考：[控制台内嵌及分享](https://help.aliyun.com/document_detail/74971.html)
- 优化`xcol`的表现形式，现以下拉框的形式规范推荐输入。兼顾兼容性与自定义输入。
- Variable 编辑页面，SLS DataSource 同样引入 Monaco Editor，且自动识别 grafana 版本切换新老显示。
- 修复`xcol`与`ycol`输入框末尾 Tips 在 Grafana 8.3.x 及以下无法显示的问题。
- 修复部分适配问题。

![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690257487839-24ccca2d-4fad-4011-9f18-300ceb876a26.png#averageHue=%231d2023&clientId=ud2e482ee-2fd6-4&from=paste&height=523&id=uc82855a6&originHeight=523&originWidth=1723&originalType=binary&ratio=1&rotation=0&showTitle=false&size=531654&status=done&style=none&taskId=u6031a614-f570-4a7b-a959-f1fda0001c8&title=&width=1723)<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1690257391475-dae60eef-2191-42ab-90e7-d79b5319dfcd.png#averageHue=%231c1f21&clientId=ud2e482ee-2fd6-4&from=paste&height=349&id=u2c395060&originHeight=349&originWidth=1719&originalType=binary&ratio=1&rotation=0&showTitle=false&size=366996&status=done&style=none&taskId=u22e237d7-d4ad-4a9f-aae2-8ce25dc0915&title=&width=1719)

- Optimize the back-end structure of the SLS Grafana plug-in, and now support the addition of custom Resource API functions.
- Introducing the gotoSLS function, users can easily jump to the SLS console to query and experience more powerful functions of the SLS console, and the jump includes the current Grafana query and time information. Configuring roleArn on the DataSource interface can realize STS jumping. If not configured, it will jump according to the normal direct access logic (login to the console is required).
  - **Note: If STS redirection is configured, the following two conditions need to be met for permission security considerations:**
    - **To configure the user corresponding to the accessKey of the DataSource, `AliyunRAMReadOnlyAccess` permission is required**
    - **Configure the roleArn of DataSource, the permission policy inside must have and only `AliyunLogReadOnlyAccess`**
  - Principle reference: [Console embedding and sharing](https://help.aliyun.com/document_detail/74971.html)
- Optimize the expression form of `xcol`, and now standardize the recommended input in the form of a drop-down box. Take into account compatibility and custom input.
- On the Variable editing page, SLS DataSource also introduces Monaco Editor, and automatically recognizes the grafana version to switch between the new and old displays.
- Fix the problem that Tips at the end of the `xcol` and `ycol` input boxes cannot be displayed in Grafana 8.3.x and below.
- Fix some adaptation problems.

# 2.29 (2023-06-29)

- 全新 SLS Query 编写体验，引入 Monaco Editor，支持语法高亮显示，关键字、函数自动提示等功能。
- SLS Query 框支持多行自动换行和高度扩展。
- `xcol`与`ycol`的输入使用 Grafana 的 Input 组件，带来更一致的体验。
- `xcol`与`ycol`输入框末尾增加 Tips，介绍常见写法，方便初学者上手。
- 修复一系列版本问题。

<br />

- Brand-new SLS Query writing experience, introduces Monaco Editor, supports syntax highlighting, keyword, function automatic prompting and other functions.
- The SLS Query box supports multi-line word wrapping and height expansion.
- The input of xcol and ycol uses Grafana's Input component to bring a more consistent experience.
- Tips are added at the end of the xcol and ycol input boxes to introduce common writing methods, which is convenient for beginners.
- Fix a series of version issues.

![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1687953008367-f863ccc6-dcf5-4998-bcbc-c18aec0ac059.png#averageHue=%231e2024&clientId=u72ccc8ce-757a-4&from=paste&height=347&id=u3586dcad&originHeight=347&originWidth=1216&originalType=binary&ratio=1&rotation=0&showTitle=false&size=294424&status=done&style=none&taskId=u73ee0591-f289-4f53-8886-1a93e4914e5&title=&width=1216)<br />![image.png](https://cdn.nlark.com/yuque/0/2023/png/21832175/1687953284855-86c636fb-caa6-4163-ab23-0478cfb55b73.png#averageHue=%2323252a&clientId=u72ccc8ce-757a-4&from=paste&height=270&id=u4ec3dd89&originHeight=270&originWidth=564&originalType=binary&ratio=1&rotation=0&showTitle=false&size=112573&status=done&style=none&taskId=uac6a4fe0-65ca-4df2-8a4e-5faa1ea911a&title=&width=564)

# 2.28

- 适配 Grafana 9.4 及以上版本的 Datasource 数据源结构。
  <br />
- Adapt to the Datasource data source structure of Grafana 9.4 and above.

# 2.27

- 加了 legacy_compatible 设置，适配旧版本的返回结构（带 response）
  <br />
- Added legacy_compatible to adapt to the return structure of the old version (with response)

# 2.0.0 (released)

Initial release.

Support Grafana 8.0.

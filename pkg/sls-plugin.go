package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// Make sure SampleDatasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler, backend.StreamHandler interfaces. Plugin should not
// implement all these interfaces - only those which are required for a particular task.
// For example if plugin does not need streaming functionality then you are free to remove
// methods that implement backend.StreamHandler. Implementing instancemgmt.InstanceDisposer
// is useful to clean up resources used by previous datasource instance when a new datasource
// instance created upon datasource settings changed.
var (
	_ backend.QueryDataHandler      = (*SlsDatasource)(nil)
	_ backend.CheckHealthHandler    = (*SlsDatasource)(nil)
	_ backend.StreamHandler         = (*SlsDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*SlsDatasource)(nil)
)

// NewSampleDatasource creates a new datasource instance.
// func NewSLSDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
func NewSLSDatasource() datasource.ServeOpts {
	ds := &SlsDatasource{}
	return datasource.ServeOpts{
		CheckHealthHandler:  ds,
		QueryDataHandler:    ds,
		CallResourceHandler: newResourceHandler(ds),
	}
}

// SampleDatasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type SlsDatasource struct {
	backend.DataSourceInstanceSettings
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (ds *SlsDatasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (ds *SlsDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (res *backend.QueryDataResponse, err error) {
	//log.DefaultLogger.Info("QueryData called", "request", req)

	config, err := LoadSettings(req.PluginContext)
	if err != nil {
		return
	}

	provider := sls.NewStaticCredentialsProvider(config.AccessKeyId, config.AccessKeySecret, "")
	client := sls.CreateNormalInterfaceV2(config.Endpoint, provider)
	client.SetUserAgent("grafana-go")

	if config.Region != "" {
		client.SetAuthVersion(sls.AuthV4)
		client.SetRegion(config.Region)
	}

	// create response struct
	response := backend.NewQueryDataResponse()

	queries := req.Queries

	ch := make(chan Result, len(queries))

	defer func() {
		close(ch)
		client = nil
		config = nil
		if r := recover(); r != nil {
			switch r.(type) {
			case string:
				err = errors.New(r.(string))
			case error:
				err = r.(error)
			}
			log.DefaultLogger.Error("QueryData recover", "error", err)
		}
	}()

	log.DefaultLogger.Debug("len(queries)", "len", len(queries))
	wg := sync.WaitGroup{}
	for _, query := range queries {
		wg.Add(1)
		log.DefaultLogger.Debug("range_queries", "RefID", query.RefID,
			"JSON", query.JSON, "QueryType", query.QueryType)
		go ds.QueryLogs(ch, query, client, config)
	}
	go func(chan Result) {
		for res := range ch {
			response.Responses[res.refId] = res.dataResponse
			wg.Done()
		}
	}(ch)
	wg.Wait()
	return response, nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (ds *SlsDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Debug("CheckHealth called", "request", req)

	config, err := LoadSettings(req.PluginContext)
	if err != nil {
		return nil, err
	}

	provider := sls.NewStaticCredentialsProvider(config.AccessKeyId, config.AccessKeySecret, "")
	client := sls.CreateNormalInterfaceV2(config.Endpoint, provider)
	client.SetUserAgent("grafana-go")

	if config.Region != "" {
		client.SetAuthVersion(sls.AuthV4)
		client.SetRegion(config.Region)
	}

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	log.DefaultLogger.Debug("CheckHealth", "project", config.Project, "client", client)
	// 拿当前 Project 的信息
	_, err = client.GetProject(config.Project)
	if err != nil {
		status = backend.HealthStatusError
		message = err.Error()
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

// SubscribeStream is called when a client wants to connect to a stream. This callback
// allows sending the first message.
func (ds *SlsDatasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Debug("SubscribeStream called", "request", req)

	status := backend.SubscribeStreamStatusPermissionDenied
	if req.Path == "stream" {
		// Allow subscribing only on expected path.
		status = backend.SubscribeStreamStatusOK
	}
	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

// RunStream is called once for any open channel.  Results are shared with everyone
// subscribed to the same channel.
func (ds *SlsDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Debug("RunStream called", "request", req)

	// Create the same data frame as for query data.
	frame := data.NewFrame("response")

	// Add fields (matching the same schema used in QueryData).
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, make([]time.Time, 1)),
		data.NewField("values", nil, make([]int64, 1)),
	)

	counter := 0

	// Stream data frames periodically till stream closed by Grafana.
	for {
		select {
		case <-ctx.Done():
			log.DefaultLogger.Debug("Context done, finish streaming", "path", req.Path)
			return nil
		case <-time.After(time.Second):
			// Send new data periodically.
			frame.Fields[0].Set(0, time.Now())
			frame.Fields[1].Set(0, int64(10*(counter%2+1)))

			counter++

			err := sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}
		}
	}
}

// PublishStream is called when a client sends a message to the stream.
func (ds *SlsDatasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Debug("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

func (ds *SlsDatasource) SortLogs(logs []map[string]string, col string) {
	sort.Slice(logs, func(i, j int) bool {
		iValue := toTime(logs[i][col])
		jValue := toTime(logs[j][col])
		if iValue.Unix() < jValue.Unix() {
			return true
		}
		return false
	})
}

func (ds *SlsDatasource) QueryLogs(ch chan Result, query backend.DataQuery, client sls.ClientInterface, logSource *LogSource) {
	response := backend.DataResponse{}
	refId := query.RefID
	queryInfo := &QueryInfo{}

	defer func() {
		queryInfo = nil
		if r := recover(); r != nil {
			switch r.(type) {
			case string:
				response.Error = errors.New(r.(string))
			case error:
				response.Error = r.(error)
			}
			log.DefaultLogger.Error("QueryLogs recover", "refId", refId, "error", response.Error)
			ch <- Result{
				refId:        refId,
				dataResponse: response,
			}
		}
	}()

	err := json.Unmarshal(query.JSON, &queryInfo)
	if err != nil {
		log.DefaultLogger.Error("Unmarshal queryInfo", "refId", refId, "error", err)
		response.Error = err
		ch <- Result{
			refId:        refId,
			dataResponse: response,
		}
		return
	}
	xcol := queryInfo.Xcol

	from := query.TimeRange.From.Unix()
	to := query.TimeRange.To.Unix()
	queryInfo.Query = InterpolateMacros(queryInfo.Query, from, to)
	var logStore string
	if queryInfo.LogStore != "" {
		logStore = queryInfo.LogStore
	} else {
		logStore = logSource.LogStore
	}

	// 如果 logstore 为空 返回错误
	if logStore == "" {
		response.Error = errors.New("logStore is empty, please select a logstore")
		ch <- Result{
			refId:        refId,
			dataResponse: response,
		}
		return
	}

	// 如果是 metric 类型 直连 Prometheus 获取数据
	if queryInfo.Type == "metricstore" {
		var metricFrames data.Frames
		err := ds.getMetricLogs(ch, query, queryInfo, logSource, response, &metricFrames)

		if err != nil {
			response.Error = err
			ch <- Result{
				refId:        refId,
				dataResponse: response,
			}
		} else {
			response.Frames = metricFrames
			ch <- Result{
				refId:        refId,
				dataResponse: response,
			}
		}
		return
	}

	var ycols []string
	queryCount := ds.GetQueryCount(queryInfo)
	getLogsResp := &sls.GetLogsResponse{}
	for i := 0; i < int(queryCount); i++ {
		currentPage := queryInfo.CurrentPage
		if currentPage <= 0 {
			currentPage = 1
		}
		offset := (currentPage - 1) * queryInfo.LogsPerPage
		getLogsReq := &sls.GetLogRequest{
			From:     from,
			To:       to,
			Query:    queryInfo.Query,
			Lines:    queryInfo.LogsPerPage,
			Offset:   offset,
			Reverse:  true,
			PowerSQL: queryInfo.PowerSql,
		}
		tem, err := client.GetLogsV2(logSource.Project, logStore, getLogsReq)
		if err != nil {
			log.DefaultLogger.Error("GetLogs ", "query : ", queryInfo.Query, "error ", err)
			response.Error = err
			ch <- Result{
				refId:        refId,
				dataResponse: response,
			}
			return
		}
		if i == 0 {
			getLogsResp = tem
		} else {
			getLogsResp.Logs = append(getLogsResp.Logs, tem.Logs...)
			getLogsResp.Count = getLogsResp.Count + tem.Count
		}
		if tem.Count < queryInfo.LogsPerPage {
			break
		}
		queryInfo.CurrentPage++
	}

	logs := getLogsResp.Logs
	c := &Contents{}
	err = json.Unmarshal([]byte(getLogsResp.Contents), &c)
	if err != nil {
		log.DefaultLogger.Error("GetLogs ", "Contents : ", getLogsResp.Contents, "error ", err)
		response.Error = err
		ch <- Result{
			refId:        refId,
			dataResponse: response,
		}
		return
	}
	if getLogsResp.Progress != "Complete" {
		log.DefaultLogger.Warn("GetLogs ", "Progress: ", getLogsResp.Progress, "warn: ", "incomplete logs, you can switch to powerSql")
		response.Error = errors.New("incomplete logs, you can switch to powerSql")
	}

	keys := c.Keys
	if compatible {
		queryInfo.Ycol = strings.Replace(queryInfo.Ycol, " ", "", -1)
	}

	isFlowGraph := strings.Contains(queryInfo.Ycol, "#:#")
	if isFlowGraph {
		ycols = strings.Split(queryInfo.Ycol, "#:#")
	} else {
		ycols = strings.Split(queryInfo.Ycol, ",")
	}
	if !compatible {
		for i := range ycols {
			ycols[i] = strings.TrimSpace(ycols[i])
		}
	}
	log.DefaultLogger.Debug("QueryLogs", "getLogsResp", getLogsResp)
	getLogsResp = nil
	var frames data.Frames

	if xcol == "trace" {
		log.DefaultLogger.Debug("BuildTrace")
		ds.BuildTrace(logs, &frames)
		response.Frames = frames
		ch <- Result{
			refId:        refId,
			dataResponse: response,
		}
		return
	}
	if !strings.Contains(queryInfo.Query, "|") {
		log.DefaultLogger.Debug("BuildLogs")
		ds.BuildLogs(logs, ycols, &frames)
		response.Frames = frames
		ch <- Result{
			refId:        refId,
			dataResponse: response,
		}
		return
	}

	if isFlowGraph {
		log.DefaultLogger.Debug("flow_graph")
		err = ds.BuildFlowGraph(logs, xcol, ycols, &frames)
		if err != nil {
			response.Error = err
		} else {
			response.Frames = frames
		}
		ch <- Result{
			refId:        refId,
			dataResponse: response,
		}
		return
	}
	if xcol == "bar" {
		log.DefaultLogger.Debug("bar")
		ds.BuildBarGraph(logs, ycols, &frames)
	} else if xcol == "map" {
		log.DefaultLogger.Debug("map")
		ds.BuildMapGraph(logs, ycols, &frames)
	} else if xcol == "pie" {
		log.DefaultLogger.Debug("pie")
		ds.BuildPieGraph(logs, ycols, &frames)
	} else if xcol != "" && xcol != "map" && xcol != "pie" && xcol != "bar" && xcol != "table" {
		log.DefaultLogger.Debug("time_graph")
		ds.BuildTimingGraph(logs, xcol, ycols, keys, &frames)
	} else {
		log.DefaultLogger.Debug("table")
		ds.BuildTable(logs, xcol, ycols, keys, &frames)
	}
	response.Frames = frames
	ch <- Result{
		refId:        refId,
		dataResponse: response,
	}
}

func (ds *SlsDatasource) GetQueryCount(queryInfo *QueryInfo) int64 {
	queryInfo.LogsPerPage = 100
	if strings.Contains(strings.ToUpper(queryInfo.Query), "SELECT") && strings.Contains(queryInfo.Query, "|") {
		return 1
	}
	if queryInfo.TotalLogs >= 5000 {
		queryInfo.TotalLogs = 5000
	}
	if queryInfo.TotalLogs <= 0 {
		queryInfo.TotalLogs = 1
	}
	if queryInfo.TotalLogs < queryInfo.LogsPerPage {
		queryInfo.LogsPerPage = queryInfo.TotalLogs
	}

	if queryInfo.LogsPerPage > 0 {
		return (queryInfo.TotalLogs + queryInfo.LogsPerPage - 1) / queryInfo.LogsPerPage
	}
	return 1
}

func (ds *SlsDatasource) getMetricLogs(_ chan Result, query backend.DataQuery, queryInfo *QueryInfo, logSource *LogSource, response backend.DataResponse, frames *data.Frames) error {
	var logStore string
	if queryInfo.LogStore != "" {
		logStore = queryInfo.LogStore
	} else {
		logStore = logSource.LogStore
	}
	project := logSource.Project
	endpoint := logSource.Endpoint
	headers := logSource.Headers
	metricStore := logStore
	queryType := queryInfo.QueryType
	intervalMs := queryInfo.IntervalMs

	// 判断如果 IntervalMs <= 0 则设置为 15000
	if intervalMs <= 0 {
		intervalMs = 15000
	}

	// 根据 queryType 动态调整 URL 的路径部分
	var apiPath string
	if queryType == "instant" {
		apiPath = "query"
	} else {
		apiPath = "query_range"
	}

	baseUrl := fmt.Sprintf("https://%s.%s/prometheus/%s/%s/api/v1/%s", project, endpoint, project, metricStore, apiPath)
	from := query.TimeRange.From.Unix()
	to := query.TimeRange.To.Unix()
	queryStr := queryInfo.Query

	log.DefaultLogger.Debug("getMetricLogs-intervalMs", "intervalMs", intervalMs)
	step := (time.Duration(intervalMs) * time.Millisecond).String()
	if queryInfo.Step != "" {
		step = queryInfo.Step
	}

	urlVal := url.Values{}
	var uri string

	switch apiPath {
	case "query":
		urlVal.Add("query", queryStr)
		urlVal.Add("time", strconv.FormatInt(to, 10))
		uri = fmt.Sprintf("%s?%v", baseUrl, urlVal.Encode())
	case "query_range":
		urlVal.Add("query", queryStr)
		urlVal.Add("start", strconv.FormatInt(from, 10))
		urlVal.Add("end", strconv.FormatInt(to, 10))
		// 判断 step 是不是空 则使用默认值
		urlVal.Add("step", step)

		uri = fmt.Sprintf("%s?%v", baseUrl, urlVal.Encode())
	}

	log.DefaultLogger.Debug("getMetricLogs-URL", "uri", uri)

	// 创建 HTTP 请求
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(logSource.AccessKeyId, logSource.AccessKeySecret)

	// 判断 headers 是否存在，存在则添加
	if len(headers) > 0 {
		for _, header := range headers {
			req.Header.Set(header.Name, header.Value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.DefaultLogger.Error("getMetricLogs ", "queryInfo: ", queryInfo, "error ", err)
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body) // 需要读完body内容。
	if err != nil {
		log.DefaultLogger.Debug("getMetricLogs-read-body-err", "err", err)
		return err
	}
	// 拿到 body 中的 data
	// 解析 JSON 数据到结构体
	var metricLogs MetricLogs
	if err := json.Unmarshal([]byte(body), &metricLogs); err != nil {
		log.DefaultLogger.Debug("parse-body-err", "err", err)
		return err
	}

	for _, metricData := range metricLogs.Data.Result {
		var frame *data.Frame
		if compatible {
			frame = data.NewFrame("response")
		} else {
			frame = data.NewFrame("")
		}

		var times []time.Time = make([]time.Time, 0)
		var values []float64 = make([]float64, 0)

		if apiPath == "query_range" {
			// 遍历result.values，将values中的数据添加到frame.Field
			for _, value := range metricData.Values {
				metricValue, err := strconv.ParseFloat(value[1].(string), 64)
				if err != nil {
					log.DefaultLogger.Debug("ParseFloat-metricData", "ParseFloat", err, "metricValue", metricValue)
				}
				times = append(times, time.Unix(int64(value[0].(float64)), 0))
				values = append(values, metricValue)
			}
		} else {
			var instantValue []interface{} = metricData.Value
			metricValue, err := strconv.ParseFloat(instantValue[1].(string), 64)
			if err != nil {
				log.DefaultLogger.Debug("ParseFloat-metricData", "ParseFloat", err, "metricValue", metricValue)
			}
			times = append(times, time.Unix(int64(instantValue[0].(float64)), 0))
			values = append(values, metricValue)
		}

		frame.Fields = append(frame.Fields, data.NewField("time", nil, times))
		frame.Fields = append(frame.Fields, data.NewField("value", metricData.Metric, values).SetConfig(&data.FieldConfig{
			DisplayNameFromDS: formatDisplayName(queryInfo.LegendFormat, metricData.Metric, queryInfo.Query),
		}))

		*frames = append(*frames, frame)
	}

	return nil
}

func (ds *SlsDatasource) BuildFlowGraphV2(logs []map[string]string, xcol string, ycols []string, frames *data.Frames) {
	if len(logs) == 0 {
		return
	}
	ds.SortLogs(logs, xcol)
	metricNames := strings.Split(ycols[1], ",")
	metricMap := make(map[string]bool)
	for _, n := range metricNames {
		metricMap[n] = true
	}
	var labelNames []string
	for k := range logs[0] {
		if k != "__source__" && k != "__time__" && !metricMap[k] && k != xcol {
			labelNames = append(labelNames, k)
		}
	}
	sort.Strings(labelNames)
	timeFields := make(map[string][]time.Time)
	nameMetricFields := make(map[string]map[string][]float64)
	for _, n := range metricNames {
		nameMetricFields[n] = make(map[string][]float64)
	}
	frameLabelsMap := make(map[string]map[string]string)

	for _, alog := range logs {
		timeVal := alog[xcol]
		labels := map[string]string{}
		labelsKey := ""
		for _, l := range labelNames {
			labels[l] = alog[l]
			labelsKey += alog[l]
		}
		if _, ok := frameLabelsMap[labelsKey]; !ok {
			frameLabelsMap[labelsKey] = labels
		}
		if _, ok := timeFields[labelsKey]; ok {
			timeFields[labelsKey] = append(timeFields[labelsKey], toTime(timeVal))
		} else {
			timeFields[labelsKey] = []time.Time{toTime(timeVal)}
		}
		for _, n := range metricNames {
			metricVal := alog[n]
			floatV, err := strconv.ParseFloat(metricVal, 64)
			if err != nil {
				log.DefaultLogger.Debug("BuildFlowGraphV2", "ParseFloat", err, "value", metricVal)
			}
			if _, ok := nameMetricFields[n][labelsKey]; ok {
				nameMetricFields[n][labelsKey] = append(nameMetricFields[n][labelsKey], floatV)
			} else {
				nameMetricFields[n][labelsKey] = []float64{floatV}
			}
		}
	}
	for k, v := range timeFields {
		frame := data.NewFrame("")
		frame.Fields = append(frame.Fields, data.NewField("Time", nil, v))
		if len(metricNames) == 1 {
			frame.Fields = append(frame.Fields, data.NewField("Value", frameLabelsMap[k], nameMetricFields[metricNames[0]][k]))
		} else {
			for _, n := range metricNames {
				frame.Fields = append(frame.Fields, data.NewField(n, frameLabelsMap[k], nameMetricFields[n][k]))
			}
		}
		*frames = append(*frames, frame)
	}
}

func (ds *SlsDatasource) BuildFlowGraph(logs []map[string]string, xcol string, ycols []string, frames *data.Frames) (err error) {
	if len(logs) == 0 {
		frames = &data.Frames{}
		return
	}
	if len(ycols) < 2 {
		return
	}
	if ycols[0] == "" && ycols[1] != "" {
		ds.BuildFlowGraphV2(logs, xcol, ycols, frames)
		return
	}
	var frame *data.Frame
	if compatible {
		frame = data.NewFrame("response")
	} else {
		frame = data.NewFrame("")
	}
	fieldMap := make(map[int]map[int64]float64)
	timeSet := make(map[int64]bool)
	labelSet := make(map[string]bool)
	var labelArr []string
	for _, alog := range logs {
		if !labelSet[alog[ycols[0]]] {
			labelArr = append(labelArr, alog[ycols[0]])
		}
		timeSet[toTime(alog[xcol]).Unix()] = true
		labelSet[alog[ycols[0]]] = true
	}
	var timeArr []int64
	for k := range timeSet {
		timeArr = append(timeArr, k)
	}
	sort.Slice(timeArr, func(i, j int) bool {
		if timeArr[i] < timeArr[j] {
			return true
		}
		return false
	})
	fieldSet := make(map[string]bool)
	labelToIndex := make(map[string]int)
	for i := range labelArr {
		fieldMap[i] = map[int64]float64{}
		labelToIndex[labelArr[i]] = i
	}
	if len(labelArr)*len(timeArr) > maxPointsLimit {
		s := fmt.Sprintf("BuildFlowGraph, More than %d points : %d", maxPointsLimit, len(labelArr)*len(timeArr))
		log.DefaultLogger.Error(s)
		err = errors.New(s)
		return
	}
	for i := range labelArr {
		for _, t0 := range timeArr {
			fieldMap[i][t0] = 0
		}
	}
	for _, alog := range logs {
		label := alog[ycols[0]]
		t := alog[xcol]
		if !fieldSet[t+label] {
			fieldSet[t+label] = true
			floatV, err1 := strconv.ParseFloat(alog[ycols[1]], 64)
			if err1 != nil {
				log.DefaultLogger.Debug("BuildFlowGraph", "ParseFloat", err, "value", alog[ycols[1]])
			}
			fieldMap[labelToIndex[label]][toTime(t).Unix()] = floatV
		}
	}
	var frameLen int
	for i, k := range labelArr {
		v := fieldMap[i]
		if len(v) > 0 {
			if frameLen == 0 {
				frameLen = len(v)
			}
			if len(v) == frameLen {
				arr := mapToSlice(timeArr, v)
				frame.Fields = append(frame.Fields, data.NewField(k, nil, arr))
			}
		}
	}
	var times []time.Time
	for _, k := range timeArr {
		times = append(times, time.Unix(k, 0))
	}
	if len(times) == frameLen {
		frame.Fields = append([]*data.Field{data.NewField("time", nil, times)}, frame.Fields...)
	}
	*frames = append(*frames, frame)
	return
}

func (ds *SlsDatasource) BuildBarGraph(logs []map[string]string, ycols []string, frames *data.Frames) {
	frame := data.NewFrame("response")
	numMap := make(map[string][]float64)
	for _, ycol := range ycols[1:] {
		numMap[ycol] = make([]float64, 0)
	}
	strKey := ycols[0]
	var strArr []string
	for _, alog := range logs {
		for k, v := range alog {
			if numMap[k] != nil {
				floatV, err := strconv.ParseFloat(v, 64)
				if err != nil {
					log.DefaultLogger.Debug("BuildBarGraph", "ParseFloat", err, "value", v)
				}
				numMap[k] = append(numMap[k], floatV)
			}
			if k == strKey {
				strArr = append(strArr, v)
			}
		}
	}
	frame.Fields = append(frame.Fields, data.NewField(strKey, nil, strArr))
	for _, ycol := range ycols[1:] {
		frame.Fields = append(frame.Fields, data.NewField(ycol, nil, numMap[ycol]))
	}
	*frames = append(*frames, frame)
}

func (ds *SlsDatasource) BuildMapGraph(logs []map[string]string, ycols []string, frames *data.Frames) {
	frame := data.NewFrame("response")
	strMap := make(map[string][]string)

	for _, ycol := range ycols[:len(ycols)-1] {
		strMap[ycol] = make([]string, 0)
	}
	numKey := ycols[len(ycols)-1]
	var numArr []float64
	for _, alog := range logs {
		for k, v := range alog {
			if strMap[k] != nil {
				strMap[k] = append(strMap[k], v)
			}
			if k == numKey {
				floatV, err := strconv.ParseFloat(v, 64)
				if err != nil {
					log.DefaultLogger.Debug("BuildMapGraph", "ParseFloat", err, "value", v)
				}
				numArr = append(numArr, floatV)
			}
		}
	}
	for k, v := range strMap {
		frame.Fields = append(frame.Fields, data.NewField(k, nil, v))
	}
	frame.Fields = append(frame.Fields, data.NewField(numKey, nil, numArr))
	*frames = append(*frames, frame)
}

func (ds *SlsDatasource) BuildPieGraph(logs []map[string]string, ycols []string, frames *data.Frames) {
	if len(ycols) < 2 {
		return
	}
	frame := data.NewFrame("response")
	fieldMap := make(map[string][]float64)
	var labelArr []string
	for _, alog := range logs {
		labelArr = append(labelArr, alog[ycols[0]])
	}
	for _, label := range labelArr {
		exist := false
		for _, alog := range logs {
			if alog[ycols[0]] == label {
				floatV, err := strconv.ParseFloat(alog[ycols[1]], 64)
				if err != nil {
					log.DefaultLogger.Debug("BuildPieGraph", "ParseFloat", err, "value", alog[ycols[1]])
				}
				fieldMap[label] = append(fieldMap[label], floatV)
				exist = true
			}
		}
		if !exist {
			fieldMap[label] = append(fieldMap[label], 0)
		}
	}

	for _, v := range labelArr {
		frame.Fields = append(frame.Fields, data.NewField(v, nil, fieldMap[v]))
	}
	*frames = append(*frames, frame)
}

func (ds *SlsDatasource) BuildTimingGraph(logs []map[string]string, xcol string, ycols []string, keys []string, frames *data.Frames) {
	ds.SortLogs(logs, xcol)
	var frame *data.Frame
	if compatible {
		frame = data.NewFrame("response")
	} else {
		frame = data.NewFrame("")
	}
	fieldMap := make(map[string][]*float64)
	var times []time.Time
	if len(ycols) == 1 && ycols[0] == "" && len(keys) > 0 {
		ycols = keys
	}
	for _, v := range ycols {
		if v != xcol {
			fieldMap[v] = make([]*float64, 0)
		}
	}
	for _, alog := range logs {
		for k, v := range alog {
			if fieldMap[k] != nil {
				// 判断一下这个v 是不是'null'
				if v == "null" {
					fieldMap[k] = append(fieldMap[k], nil)
				} else {
					floatV, err := strconv.ParseFloat(v, 64)
					if err != nil {
						log.DefaultLogger.Debug("BuildTimingGraph", "ParseFloat", err, "value", v)
					}
					fieldMap[k] = append(fieldMap[k], &floatV)
				}
			}
			if xcol != "" && xcol == k {
				times = append(times, toTime(v))
			}
		}
	}
	var frameLen int
	for _, v := range fieldMap {
		if len(v) > frameLen {
			frameLen = len(v)
		}
	}
	if len(times) == frameLen {
		frame.Fields = append(frame.Fields, data.NewField("time", nil, times))
	}
	for _, v := range ycols {
		if field, ok := fieldMap[v]; ok && len(field) == frameLen {
			frame.Fields = append(frame.Fields, data.NewField(v, nil, field))
		}
	}
	*frames = append(*frames, frame)
}

func (ds *SlsDatasource) BuildTable(logs []map[string]string, xcol string, ycols []string, keys []string, frames *data.Frames) {
	frame := data.NewFrame(strings.Join(ycols, ","))

	fieldMap := make(map[string][]string)

	var keyArr []string

	var times []time.Time

	if len(ycols) == 1 && ycols[0] == "" && len(logs) > 0 {
		ycols = ycols[:0]
		if len(keys) > 0 {
			ycols = append(ycols, keys...)
		} else {
			for k := range logs[0] {
				if k != "__time__" && k != "__source__" {
					ycols = append(ycols, k)
				}
			}
		}
	}
	for _, ycol := range ycols {
		fieldMap[ycol] = make([]string, 0)
		keyArr = append(keyArr, ycol)
	}
	for _, alog := range logs {
		for k, v := range alog {
			if fieldMap[k] != nil {
				fieldMap[k] = append(fieldMap[k], v)
			}
			if xcol != "" && xcol == k {
				floatValue, err := strconv.ParseFloat(v, 64)
				if err != nil {
					log.DefaultLogger.Debug("BuildTable", "ParseTime", err)
					continue
				}
				t := time.Unix(int64(floatValue), 0)
				times = append(times, t)
			}
		}
	}
	for _, v := range keyArr {
		frame.Fields = append(frame.Fields, data.NewField(v, nil, fieldMap[v]))
	}
	if len(times) > 0 {
		frame.Fields = append(frame.Fields, data.NewField("time", nil, times))
	}
	*frames = append(*frames, frame)
}

func (ds *SlsDatasource) BuildLogs(logs []map[string]string, ycols []string, frames *data.Frames) {
	frame := data.NewFrame("")
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: data.VisTypeLogs,
	}
	fieldMap := make(map[string][]string)
	var keyArr []string
	var times []time.Time
	if len(ycols) == 1 && ycols[0] == "" {
		for _, alog := range logs {
			for k := range alog {
				if _, ok := fieldMap[k]; !ok {
					fieldMap[k] = make([]string, 0)
					keyArr = append(keyArr, k)
				}
			}
		}
	} else {
		for _, ycol := range ycols {
			fieldMap[ycol] = make([]string, 0)
			keyArr = append(keyArr, ycol)
		}
	}
	var values []string
	for _, alog := range logs {
		message := ""
		for _, k := range keyArr {
			fieldMap[k] = append(fieldMap[k], alog[k])
			message = message + k + `="` + strings.ReplaceAll(alog[k], `"`, `'`) + `" `
		}
		timeValue, _ := strconv.ParseFloat(alog["__time__"], 64)

		var t time.Time
		if ns, ok := alog["__time_ns_part__"]; ok {
			ns, _ := strconv.ParseInt(ns, 10, 64)
			t = time.Unix(int64(timeValue), ns)
		} else {
			t = time.Unix(int64(timeValue), 0)
		}
		times = append(times, t)
		values = append(values, message)
	}

	if len(times) > 0 {
		frame.Fields = append(frame.Fields, data.NewField("time", nil, times))
	}
	frame.Fields = append(frame.Fields, data.NewField("message", nil, values))
	for _, v := range keyArr {
		frame.Fields = append(frame.Fields, data.NewField(v, nil, fieldMap[v]))
	}
	*frames = append(*frames, frame)
}

// BuildTrace
// *| select traceID,spanID,parentSpanID,service,host,resource,attribute,statusCode,statusMessage,logs,name,start,duration limit 100
func (ds *SlsDatasource) BuildTrace(logs []map[string]string, frames *data.Frames) {
	frame := data.NewFrame("response")
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: data.VisTypeTrace,
	}

	traceID := make([]string, 0)
	spanID := make([]string, 0)
	parentSpanID := make([]string, 0)
	serviceName := make([]string, 0)
	startTime := make([]float64, 0)
	duration := make([]float64, 0)
	resource := make([]string, 0)
	host := make([]string, 0)
	attribute := make([]string, 0)
	statusCode := make([]string, 0)
	statusMessage := make([]string, 0)
	logs1 := make([]string, 0)
	operationName := make([]string, 0)
	for _, alog := range logs {
		traceID = append(traceID, alog["traceID"])
		spanID = append(spanID, alog["spanID"])
		parentSpanID = append(parentSpanID, alog["parentSpanID"])
		serviceName = append(serviceName, alog["service"])
		host = append(host, alog["host"])
		resource = append(resource, alog["resource"])
		attribute = append(attribute, alog["attribute"])
		statusCode = append(statusCode, alog["statusCode"])
		statusMessage = append(statusMessage, alog["statusMessage"])
		logs1 = append(logs1, alog["logs"])
		operationName = append(operationName, alog["name"])
		startTimeV, err := strconv.ParseFloat(alog["start"], 64)
		if err != nil {
			log.DefaultLogger.Debug("BuildTrace", "ParseFloat", err)
		}
		startTime = append(startTime, startTimeV/1000)
		durationV, err := strconv.ParseFloat(alog["duration"], 64)
		if err != nil {
			log.DefaultLogger.Debug("BuildTrace", "ParseFloat", err)
		}
		duration = append(duration, durationV/1000)

	}
	frame.Fields = append(frame.Fields, data.NewField("operationName", nil, operationName))
	frame.Fields = append(frame.Fields, data.NewField("traceID", nil, traceID))
	frame.Fields = append(frame.Fields, data.NewField("spanID", nil, spanID))
	frame.Fields = append(frame.Fields, data.NewField("parentSpanID", nil, parentSpanID))
	frame.Fields = append(frame.Fields, data.NewField("serviceName", nil, serviceName))
	frame.Fields = append(frame.Fields, data.NewField("startTime", nil, startTime))
	frame.Fields = append(frame.Fields, data.NewField("duration", nil, duration))
	frame.Fields = append(frame.Fields, data.NewField("resource", nil, resource))
	frame.Fields = append(frame.Fields, data.NewField("host", nil, host))
	frame.Fields = append(frame.Fields, data.NewField("attribute", nil, attribute))
	frame.Fields = append(frame.Fields, data.NewField("statusCode", nil, statusCode))
	frame.Fields = append(frame.Fields, data.NewField("statusMessage", nil, statusMessage))
	frame.Fields = append(frame.Fields, data.NewField("logs", nil, logs1))
	*frames = append(*frames, frame)
}

func mapToSlice(timeArr []int64, m map[int64]float64) []float64 {
	s := make([]float64, 0, len(timeArr))
	for _, v := range timeArr {
		s = append(s, m[v])
	}
	return s
}

func toTime(sTime string) (t time.Time) {
	if v, err := strconv.ParseFloat(sTime, 64); err == nil {
		if len(sTime) == 13 {
			t = time.Unix(int64(v)/1000, 0)
		} else {
			t = time.Unix(int64(v), 0)
		}
		return
	}
	re := regexp.MustCompile(`(\d{4})\S(\d{2})\S(\d{2})[\s\S](\d{2})\S(\d{2})\S(\d{2}).*`)
	matched := re.FindAllStringSubmatch(sTime, -1)
	if matched != nil {
		s := fmt.Sprintf("%s-%s-%s %s:%s:%s", matched[0][1], matched[0][2], matched[0][3],
			matched[0][4], matched[0][5], matched[0][6])

		local, _ := time.LoadLocation("Asia/Shanghai")
		t, _ = time.ParseInLocation("2006-01-02 15:04:05", s, local)
	}
	return
}

// 使用 labels 替换 format 中的占位符
func formatDisplayName(format string, labels map[string]string, query string) string {
	re := regexp.MustCompile(`{{\s*([^}]+)\s*}}`)
	matches := re.FindAllStringSubmatch(format, -1)

	// 标志是否所有占位符都能匹配
	hasMatched := false

	for _, match := range matches {
		placeholder := match[0]
		labelName := match[1]
		if value, ok := labels[labelName]; ok {
			format = strings.ReplaceAll(format, placeholder, value)
			hasMatched = true
		}
	}

	if !hasMatched {
		// 如果有任何一个占位符没有对应的值，返回 labels 的字符串表示
		var parts []string
		// 对 labels 排序后拼接
		var keys []string
		for key := range labels {
			keys = append(keys, key)
		}
		// 如果 labels 为空 就直接用 query
		if len(keys) == 0 {
			return query
		}

		sort.Strings(keys) // 对 keys 进行排序
		for _, key := range keys {
			parts = append(parts, fmt.Sprintf(`%s="%s"`, key, labels[key]))
		}
		return fmt.Sprintf(`{ %s }`, strings.Join(parts, ", "))
	}
	return format
}

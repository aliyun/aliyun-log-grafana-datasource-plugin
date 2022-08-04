package main

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
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
func NewSLSDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &SlsDatasource{}, nil
}

// SampleDatasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type SlsDatasource struct{}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *SlsDatasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *SlsDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	//log.DefaultLogger.Info("QueryData called", "request", req)

	config, err := LoadSettings(req.PluginContext)
	if err != nil {
		return nil, err
	}

	client := &sls.Client{
		Endpoint:        config.Endpoint,
		AccessKeyID:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		UserAgent:       "grafana-go",
	}

	// create response struct
	response := backend.NewQueryDataResponse()

	queries := req.Queries

	ch := make(chan Result, len(queries))
	log.DefaultLogger.Info("len(queries)", "len", len(queries))

	for _, query := range queries {
		log.DefaultLogger.Info("range_queries", "RefID", query.RefID,
			"JSON", query.JSON, "QueryType", query.QueryType)
		go d.QueryLogs(ch, query, client, config)
	}
	c := 0
	for res := range ch {
		c = c + 1
		if c == len(queries) {
			close(ch)
		}
		//log.DefaultLogger.Info("range_ch", "refId", res.refId, "dataResponse", res.dataResponse)
		response.Responses[res.refId] = res.dataResponse
	}
	log.DefaultLogger.Info("range_ch", "response", response)
	return response, nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *SlsDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	config, err := LoadSettings(req.PluginContext)
	if err != nil {
		return nil, err
	}

	client := &sls.Client{
		Endpoint:        config.Endpoint,
		AccessKeyID:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		UserAgent:       "grafana-go",
	}

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	from := time.Now().Unix()
	_, err = client.GetLogs(config.Project, config.LogStore, "",
		from-60, from, "* | select count(*)", 0, 0, true)
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
func (d *SlsDatasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Info("SubscribeStream called", "request", req)

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
func (d *SlsDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream called", "request", req)

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
			log.DefaultLogger.Info("Context done, finish streaming", "path", req.Path)
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
func (d *SlsDatasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

func (ds *SlsDatasource) SortLogs(logs []map[string]string, col string) {
	sort.Slice(logs, func(i, j int) bool {
		f1, err := strconv.ParseFloat(logs[i][col], 10)
		if err != nil {
			backend.Logger.Error("SortLogs1", "error ", err)
		}
		f2, err := strconv.ParseFloat(logs[j][col], 10)
		if err != nil {
			backend.Logger.Error("SortLogs2", "error ", err)
		}
		if f1 < f2 {
			return true
		}
		return false
	})
}

func (ds *SlsDatasource) QueryLogs(ch chan Result, query backend.DataQuery, client *sls.Client, logSource *LogSource) {
	response := backend.DataResponse{}
	refId := query.RefID
	queryInfo := &QueryInfo{}
	err := json.Unmarshal([]byte(query.JSON), &queryInfo)
	if err != nil {
		log.DefaultLogger.Error("Unmarshal queryInfo", "refId", refId, "error", err)
		ch <- Result{
			refId:        refId,
			dataResponse: response,
		}
		return
	}
	xcol := queryInfo.Xcol

	from := query.TimeRange.From.Unix()
	to := query.TimeRange.To.Unix()

	log.DefaultLogger.Info("QueryLogs", "queryInfo", queryInfo)

	var ycols []string
	offset := (queryInfo.CurrentPage - 1) * queryInfo.LogsPerPage
	getLogsResp, err := client.GetLogs(logSource.Project, logSource.LogStore, "",
		from, to, queryInfo.Query, queryInfo.LogsPerPage, offset, true)
	if err != nil {
		log.DefaultLogger.Error("GetLogs ", "query : ", queryInfo.Query, "error ", err)
		response.Error = err
		ch <- Result{
			refId:        refId,
			dataResponse: response,
		}
		return
	}
	logs := getLogsResp.Logs

	queryInfo.Ycol = strings.Replace(queryInfo.Ycol, " ", "", -1)
	isFlowGraph := strings.Contains(queryInfo.Ycol, "#:#")
	if isFlowGraph {
		ycols = strings.Split(queryInfo.Ycol, "#:#")
	} else {
		ycols = strings.Split(queryInfo.Ycol, ",")
	}
	log.DefaultLogger.Info("QueryLogs", "getLogsResp", getLogsResp)

	var frames data.Frames

	if xcol == "trace" {
		log.DefaultLogger.Info("BuildTrace")
		ds.BuildTrace(logs, &frames)
		response.Frames = frames
		ch <- Result{
			refId:        refId,
			dataResponse: response,
		}
		return
	}
	if !strings.Contains(queryInfo.Query, "|") {
		log.DefaultLogger.Info("BuildLogs")
		ds.BuildLogs(logs, ycols, &frames)
		response.Frames = frames
		ch <- Result{
			refId:        refId,
			dataResponse: response,
		}
		return
	}

	if isFlowGraph {
		log.DefaultLogger.Info("flow_graph")
		ds.BuildFlowGraph(logs, xcol, ycols, refId, &frames)
	} else if xcol == "bar" {
		log.DefaultLogger.Info("bar")
		ds.BuildBarGraph(logs, ycols, refId, &frames)
	} else if xcol == "map" {
		log.DefaultLogger.Info("map")
		ds.BuildMapGraph(logs, ycols, refId, &frames)
	} else if xcol == "pie" {
		log.DefaultLogger.Info("pie")
		ds.BuildPieGraph(logs, ycols, refId, &frames)
	} else if xcol != "" && xcol != "map" && xcol != "pie" && xcol != "bar" && xcol != "table" {
		log.DefaultLogger.Info("time_graph")
		ds.BuildTimingGraph(logs, xcol, ycols, &frames)
	} else {
		log.DefaultLogger.Info("table")
		ds.BuildTable(logs, xcol, ycols, &frames)
	}
	response.Frames = frames
	ch <- Result{
		refId:        refId,
		dataResponse: response,
	}
}

func (ds *SlsDatasource) BuildFlowGraph(logs []map[string]string, xcol string, ycols []string, refId string, frames *data.Frames) {
	if len(ycols) < 2 {
		return
	}
	frame := data.NewFrame("response")
	fieldMap := make(map[string]map[string]float64)
	timeSet := make(map[string]bool)
	labelSet := make(map[string]bool)
	for _, alog := range logs {
		timeSet[alog[xcol]] = true
		labelSet[alog[ycols[0]]] = true
	}
	var timeArr []string
	for k := range timeSet {
		timeArr = append(timeArr, k)
	}
	sort.Slice(timeArr, func(i, j int) bool {
		iValue, _ := strconv.ParseFloat(timeArr[i], 9)
		jValue, _ := strconv.ParseFloat(timeArr[j], 9)
		if iValue < jValue {
			return true
		}
		return false
	})
	// 流图重复聚合列(错误语句) field长度不一致 crash
	fieldSet := make(map[string]bool)
	/*1.0*/
	//for label := range labelSet {
	//	for _, t := range timeArr {
	//		// 循环查 没有补0
	//		exist := false
	//		for _, alog := range logs {
	//			if !fieldSet[alog[xcol]+alog[ycols[0]]] && alog[xcol] == t && alog[ycols[0]] == label {
	//				fieldSet[alog[xcol]+alog[ycols[0]]] = true
	//				floatV, err := strconv.ParseFloat(alog[ycols[1]], 10)
	//				if err == nil {
	//					fieldMap[label] = append(fieldMap[label], floatV)
	//				} else {
	//					log.DefaultLogger.Error("BuildFlowGraph", "ParseFloat", err, "value", alog[ycols[1]])
	//					fieldMap[label] = append(fieldMap[label], 0)
	//				}
	//				exist = true
	//			}
	//		}
	//		if !exist {
	//			fieldMap[label] = append(fieldMap[label], 0)
	//		}
	//	}
	//}
	/*2.0
	先补0再覆盖
	*/
	for label := range labelSet {
		fieldMap[label] = make(map[string]float64)
	}
	for label := range labelSet {
		for _, t := range timeArr {
			fieldMap[label][t] = 0
		}
	}
	for _, alog := range logs {
		label := alog[ycols[0]]
		t := alog[xcol]
		if !fieldSet[t+label] {
			fieldSet[t+label] = true
			floatV, err := strconv.ParseFloat(alog[ycols[1]], 10)
			if err == nil {
				fieldMap[label][t] = floatV
			} else {
				log.DefaultLogger.Error("BuildFlowGraph", "ParseFloat", err, "value", alog[ycols[1]])
				fieldMap[label][t] = 0
			}
		}
	}
	/*end*/
	var frameLen int
	for k, v := range fieldMap {
		if len(v) > 0 {
			if frameLen == 0 {
				frameLen = len(v)
			}
			if len(v) == frameLen {
				frame.Fields = append(frame.Fields, data.NewField(k, nil, mapToSlice(timeArr, v)))
			}
		}
	}

	var times []time.Time
	for _, k := range timeArr {
		v, _ := strconv.ParseFloat(k, 9)
		t := time.Unix(int64(v), 0)
		times = append(times, t)
	}
	if len(times) == frameLen {
		frame.Fields = append(frame.Fields, data.NewField("time", nil, times))
	}

	*frames = append(*frames, frame)
}

func (ds *SlsDatasource) BuildBarGraph(logs []map[string]string, ycols []string, refId string, frames *data.Frames) {
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
				floatV, err := strconv.ParseFloat(v, 10)
				if err == nil {
					numMap[k] = append(numMap[k], floatV)
				} else {
					log.DefaultLogger.Error("BuildBarGraph", "ParseFloat", err, "value", v)
					numMap[k] = append(numMap[k], floatV)
				}
			}
			if k == strKey {
				strArr = append(strArr, v)
			}
		}
	}
	frame.Fields = append(frame.Fields, data.NewField(strKey, nil, strArr))
	for k, v := range numMap {
		frame.Fields = append(frame.Fields, data.NewField(k, nil, v))
	}
	*frames = append(*frames, frame)
}

func (ds *SlsDatasource) BuildMapGraph(logs []map[string]string, ycols []string, refId string, frames *data.Frames) {
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
				floatV, err := strconv.ParseFloat(v, 10)
				if err == nil {
					numArr = append(numArr, floatV)
				} else {
					log.DefaultLogger.Error("BuildMapGraph", "ParseFloat", err, "value", v)
					numArr = append(numArr, floatV)
				}
			}
		}
	}
	for k, v := range strMap {
		frame.Fields = append(frame.Fields, data.NewField(k, nil, v))
	}
	frame.Fields = append(frame.Fields, data.NewField(numKey, nil, numArr))
	*frames = append(*frames, frame)
}

func (ds *SlsDatasource) BuildPieGraph(logs []map[string]string, ycols []string, refId string, frames *data.Frames) {
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
				floatV, err := strconv.ParseFloat(alog[ycols[1]], 10)
				if err == nil {
					fieldMap[label] = append(fieldMap[label], floatV)
				} else {
					log.DefaultLogger.Error("BuildPieGraph", "ParseFloat", err, "value", alog[ycols[1]])
					fieldMap[label] = append(fieldMap[label], 0)
				}
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

func (ds *SlsDatasource) BuildTimingGraph(logs []map[string]string, xcol string, ycols []string, frames *data.Frames) {
	ds.SortLogs(logs, xcol)
	frame := data.NewFrame("response1")
	fieldMap := make(map[string][]float64)
	var times []time.Time
	for _, ycol := range ycols {
		fieldMap[ycol] = make([]float64, 0)
	}
	for _, alog := range logs {
		for k, v := range alog {
			if fieldMap[k] != nil {
				floatV, err := strconv.ParseFloat(v, 10)
				if err == nil {
					fieldMap[k] = append(fieldMap[k], floatV)
				} else {
					log.DefaultLogger.Error("BuildTimingGraph", "ParseFloat", err, "value", v)
					fieldMap[k] = append(fieldMap[k], 0)
				}
			}
			if xcol != "" && xcol == k {
				floatValue, err := strconv.ParseFloat(v, 9)
				if err != nil {
					log.DefaultLogger.Error("BuildTimingGraph", "ParseTime", err, "value", v)
					continue
				}
				t := time.Unix(int64(floatValue), 0)
				times = append(times, t)
			}
		}
	}
	var frameLen int
	for k, v := range fieldMap {
		if len(v) > 0 {
			if frameLen == 0 {
				frameLen = len(v)
			}
			if len(v) == frameLen {
				frame.Fields = append(frame.Fields, data.NewField(k, nil, v))
			}
		}
	}
	if len(times) == frameLen {
		frame.Fields = append(frame.Fields, data.NewField("time", nil, times))
	}
	*frames = append(*frames, frame)
}

func (ds *SlsDatasource) BuildTable(logs []map[string]string, xcol string, ycols []string, frames *data.Frames) {
	frame := data.NewFrame("response1")

	fieldMap := make(map[string][]string)

	var keyArr []string

	var times []time.Time

	if len(ycols) == 1 && ycols[0] == "" && len(logs) > 0 {
		ycols = ycols[:0]
		for k := range logs[0] {
			if k != "__time__" && k != "__source__" {
				ycols = append(ycols, k)
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
				floatValue, err := strconv.ParseFloat(v, 9)
				if err != nil {
					log.DefaultLogger.Error("BuildTable", "ParseTime", err)
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
	frame := data.NewFrame("response")
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: data.VisTypeLogs,
	}
	yset := make(map[string]bool)
	for _, ycol := range ycols {
		yset[ycol] = true
	}
	var times []time.Time
	var values []string
	for _, alog := range logs {
		message := ""
		var keys []string
		for k := range alog {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := alog[k]
			if k == "__time__" {
				floatValue, err := strconv.ParseFloat(v, 9)
				if err != nil {
					log.DefaultLogger.Error("BuildLogs", "ParseTime", err)
					continue
				}
				times = append(times, time.Unix(int64(floatValue), 0))
			}
			if len(ycols) > 0 {
				if yset[k] {
					message = message + k + `="` + strings.ReplaceAll(v, `"`, `'`) + `" `
				}
			} else {
				message = message + k + `="` + strings.ReplaceAll(v, `"`, `'`) + `" `
			}
		}
		values = append(values, message)
	}
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, times),
		data.NewField("values", nil, values),
	)
	*frames = append(*frames, frame)
}

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
		startTimeV, err := strconv.ParseFloat(alog["start"], 16)
		if err != nil {
			log.DefaultLogger.Error("BuildTrace", "ParseFloat", err)
			startTime = append(startTime, 0)
		} else {
			startTime = append(startTime, startTimeV/1000)
		}
		durationV, err := strconv.ParseFloat(alog["duration"], 16)
		if err != nil {
			log.DefaultLogger.Error("BuildTrace", "ParseFloat", err)
			duration = append(duration, 0)
		} else {
			duration = append(duration, durationV/1000)
		}

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

func mapToSlice(timeArr []string, m map[string]float64) []float64 {
	s := make([]float64, 0, len(timeArr))
	for _, v := range timeArr {
		s = append(s, m[v])
	}
	return s
}

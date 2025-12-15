package main

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type LogSource struct {
	Endpoint        string
	Project         string `json:"project"`
	LogStore        string `json:"logstore"`
	RoleArn         string `json:"roleArn"`
	Region          string
	AccessKeyId     string
	AccessKeySecret string
	Headers         []Header `json:"headers"`
	International   bool     `json:"international"`
}

type QueryInfo struct {
	Type         string `json:"type"`
	QueryMode    string `json:"mode"`
	Query        string `json:"query"`
	Xcol         string `json:"xcol"`
	Ycol         string `json:"ycol"`
	LogsPerPage  int64  `json:"logsPerPage"`
	CurrentPage  int64  `json:"currentPage"`
	LogStore     string `json:"logStore"`
	LegendFormat string `json:"legendFormat"`
	QueryType    string `json:"queryType"`
	Step         string `json:"step"`
	IntervalMs   int64  `json:"intervalMs"`
	TotalLogs    int64  `json:"totalLogs"`
	PowerSql     bool   `json:"powerSql"`
}

type Result struct {
	refId        string
	dataResponse backend.DataResponse
}

// Contents {"keys":["c","c1","t"],"terms":[["*",""]],"limited":"100"}
type Contents struct {
	Keys    []string   `json:"keys"`
	Terms   [][]string `json:"terms"`
	Limited string     `json:"limited"`
}

type ResultItem struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"`
	Value  []interface{}     `json:"value"`
}
type MetricData struct {
	ResultType string       `json:"resultType"`
	Result     []ResultItem `json:"result"`
}
type MetricLogs struct {
	Status string     `json:"status"`
	Data   MetricData `json:"data"`
}

func LoadSettings(ctx backend.PluginContext) (*LogSource, error) {
	model := &LogSource{}

	settings := ctx.DataSourceInstanceSettings
	err := json.Unmarshal(settings.JSONData, &model)
	if err != nil {
		return nil, fmt.Errorf("error reading settings: %s", err.Error())
	}
	model.Endpoint = settings.URL
	model.AccessKeyId = settings.DecryptedSecureJSONData["accessKeyId"]
	model.AccessKeySecret = settings.DecryptedSecureJSONData["accessKeySecret"]

	return model, nil
}

package main

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type LogSource struct {
	Endpoint        string
	Project         string `json:"project"`
	LogStore        string `json:"logstore"`
	RoleArn         string `json:"roleArn"`
	AccessKeyId     string
	AccessKeySecret string
}

type QueryInfo struct {
	QueryType    string `json:"type"`
	QueryMode    string `json:"mode"`
	Query        string `json:"query"`
	Xcol         string `json:"xcol"`
	Ycol         string `json:"ycol"`
	LogsPerPage  int64  `json:"logsPerPage"`
	CurrentPage  int64  `json:"currentPage"`
	TotalResults int64  `json:"totalResults"`
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

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"gitlab.alibaba-inc.com/rapt/go-security-utils/network"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
)

// InterpolateMacros replaces Grafana macros with SLS SQL equivalents
// Supported macros:
// $__time(col) -> to_unixtime(col) as time
// $__timeFilter(col) -> col >= from AND col < to
// $__timeGroup(col, 'interval', [fill]) -> time_series(col, 'interval', '%Y-%m-%d %H:%i:%s', fill)
// $__timeGroupAlias(col, 'interval') -> time_series(...) as time
func InterpolateMacros(query string, from, to int64) string {
	// $__time(dateColumn)
	// Example: $__time(log_time) -> to_unixtime(log_time) as time
	timeReg := regexp.MustCompile(`\$__time\(([^)]+)\)`)
	query = timeReg.ReplaceAllString(query, "to_unixtime($1) as time")

	// $__timeFilter(dateColumn)
	// Example: $__timeFilter(__time__) -> __time__ >= 1600000000 AND __time__ < 1600003600
	timeFilterReg := regexp.MustCompile(`\$__timeFilter\(([^)]+)\)`)
	query = timeFilterReg.ReplaceAllStringFunc(query, func(match string) string {
		parts := timeFilterReg.FindStringSubmatch(match)
		if len(parts) == 2 {
			col := parts[1]
			return fmt.Sprintf("%s >= %d AND %s < %d", col, from, col, to)
		}
		return match
	})

	// $__timeGroup(dateColumn, '5m', fill)
	// Matches: $__timeGroup(col, 'interval' [, fill])
	// Note: interval is expected to be quoted, fill is optional
	timeGroupReg := regexp.MustCompile(`\$__timeGroup\(\s*([^,]+)\s*,\s*'([^']+)'(?:\s*,\s*([^)]+))?\s*\)`)
	query = timeGroupReg.ReplaceAllStringFunc(query, func(match string) string {
		parts := timeGroupReg.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		col := parts[1]
		interval := parts[2]
		fill := "0" // Default fill
		if len(parts) > 3 && parts[3] != "" {
			fill = strings.TrimSpace(parts[3])
		}

		// Map fill strategies
		// SLS: '0', 'null', 'last'
		slsFill := "'0'"
		switch strings.ToLower(fill) {
		case "null":
			slsFill = "'null'"
		case "previous":
			slsFill = "'last'"
		case "0":
			slsFill = "'0'"
		default:
			// Treat as explicit value, wrap in quotes for SLS if it looks like a number or string
			// SLS time_series padding expects a string literal '...'
			slsFill = fmt.Sprintf("'%s'", fill)
		}

		return fmt.Sprintf("time_series(%s, '%s', '%%Y-%%m-%%d %%H:%%i:%%s', %s)", col, interval, slsFill)
	})

	// $__timeGroupAlias(dateColumn, '5m')
	// Similar to timeGroup but adds 'as time'
	timeGroupAliasReg := regexp.MustCompile(`\$__timeGroupAlias\(\s*([^,]+)\s*,\s*'([^']+)'\s*\)`)
	query = timeGroupAliasReg.ReplaceAllStringFunc(query, func(match string) string {
		parts := timeGroupAliasReg.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		col := parts[1]
		interval := parts[2]
		return fmt.Sprintf("time_series(%s, '%s', '%%Y-%%m-%%d %%H:%%i:%%s', '0') as time", col, interval)
	})

	log.DefaultLogger.Debug("InterpolateMacros", "original", query, "result", query)
	return query
}

// interpolateMacros 为了向后兼容保留的简化版本
// 使用默认时间范围调用完整的InterpolateMacros函数
func interpolateMacros(query string) string {
	// 使用默认时间范围（当前时间前后24小时）
	now := int64(1600000000) // 可以根据需要调整默认值
	from := now - 86400      // 24小时前
	to := now + 86400        // 24小时后
	return InterpolateMacros(query, from, to)
}

// decodeBase64Query 解码Base64编码的查询字符串
func decodeBase64Query(encodedQuery string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encodedQuery)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 query: %w", err)
	}
	return string(decoded), nil
}

// encodeBase64Query 编码查询字符串为Base64
func encodeBase64Query(query string) string {
	return base64.StdEncoding.EncodeToString([]byte(query))
}

// processEncodingMacros 处理Encoding字符串中的查询宏
func processEncodingMacros(encoding string) (string, error) {
	// 解析URL参数
	params, err := url.ParseQuery(encoding)
	if err != nil {
		return encoding, fmt.Errorf("failed to parse encoding parameters: %w", err)
	}

	queryString := params.Get("queryString")
	if queryString == "" {
		// 如果没有queryString参数，直接返回原始编码
		return encoding, nil
	}

	// URL解码
	decodedQueryString, err := url.QueryUnescape(queryString)
	if err != nil {
		return encoding, fmt.Errorf("failed to URL decode queryString: %w", err)
	}

	// Base64解码
	originalQuery, err := decodeBase64Query(decodedQueryString)
	if err != nil {
		return encoding, fmt.Errorf("failed to base64 decode query: %w", err)
	}

	// 应用宏转换
	interpolatedQuery := interpolateMacros(originalQuery)

	// Base64编码转换后的查询
	encodedQuery := encodeBase64Query(interpolatedQuery)

	// URL编码
	urlEncodedQuery := url.QueryEscape(encodedQuery)

	// 重新构建参数
	params.Set("queryString", urlEncodedQuery)

	log.DefaultLogger.Debug("processEncodingMacros",
		"original", originalQuery,
		"interpolated", interpolatedQuery)

	return params.Encode(), nil
}

func newResourceHandler(ds *SlsDatasource) backend.CallResourceHandler {
	mux := http.NewServeMux()

	// register route
	mux.HandleFunc("/api/gotoSLS", ds.gotoSLS)
	mux.HandleFunc("/api/version", ds.serveVersion)
	mux.HandleFunc("/api/getLogstoreList", ds.getLogstoreList)

	return httpadapter.New(mux)
}

func (ds *SlsDatasource) serveVersion(w http.ResponseWriter, r *http.Request) {
	// Handle query request...
}

type ListLogstoresData struct {
	Project       string
	TelemetryType string
}

func (ds *SlsDatasource) getLogstoreList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"data":    []map[string]interface{}{}, // 定义 data 为一个任意类型的对象数组
		"res":     nil,
		"message": "",
	}

	config, err := LoadSettings(httpadapter.PluginConfigFromContext(r.Context()))
	if err != nil {
		response["message"] = err.Error()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	provider := sls.NewStaticCredentialsProvider(config.AccessKeyId, config.AccessKeySecret, "")
	client := sls.CreateNormalInterfaceV2(config.Endpoint, provider)
	client.SetUserAgent("grafana-go")

	if config.Region != "" {
		client.SetAuthVersion(sls.AuthV4)
		client.SetRegion(config.Region)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response["message"] = err.Error()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 解析request JSON 数据
	var data ListLogstoresData
	if err := json.Unmarshal(body, &data); err != nil {
		response["message"] = err.Error()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.DefaultLogger.Debug("getBODY", "body", body, "bodyData", data)

	project, err := client.GetProject(data.Project)
	if err != nil {
		response["message"] = err.Error()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 拿当前 Project 的信息
	list, err := project.ListLogStoreV2(0, 500, data.TelemetryType)
	if err != nil {
		response["message"] = err.Error()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response["data"] = list
	response["res"] = list

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	log.DefaultLogger.Debug("get logstore success.")
}

type Data struct {
	Encoding string `json:"encoding"`
	Logstore string `json:"logstore"`
	Type     string `json:"type"`
}

func (ds *SlsDatasource) gotoSLS(w http.ResponseWriter, r *http.Request) {

	response := map[string]interface{}{
		"message": "",
		"err":     "",
		"url":     "",
		// "policy":  "",
	}

	config, err := LoadSettings(httpadapter.PluginConfigFromContext(r.Context()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ak := config.AccessKeyId
	sk := config.AccessKeySecret
	arn := config.RoleArn
	prj := config.Project
	logstore := config.LogStore

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 解析request JSON 数据
	var data Data
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 对传入的查询进行宏转换
	if data.Encoding != "" {
		newEncoding, err := processEncodingMacros(data.Encoding)
		if err != nil {
			log.DefaultLogger.Warn("Failed to process macros in encoding", "error", err)
			// 继续执行，使用原始编码
		} else {
			data.Encoding = newEncoding
		}
	}

	logstoreType := "/logsearch/"

	if data.Type == "metricsql" || data.Type == "metricstore" {
		logstoreType = "/metric/"
	}

	if data.Logstore != "" {
		logstore = data.Logstore
	}

	pattern := `^acs:ram::\d+:role\/[^\/]+$`
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return
	}

	normalJump := false

	if len(arn) == 0 {
		normalJump = true
	} else {
		if !regex.MatchString(arn) {
			response["err"] = "regexCheckError"
			response["message"] = "roleArn 不符合格式，请检查。"
			normalJump = true
		}
	}

	if !normalJump {
		roleName := strings.Split(arn, "/")[1]
		_, err2 := roleCheck(ak, sk, roleName)
		if err2 != nil {
			response["err"] = "roleCheckError"
			response["message"] = err2.Error()
			// http.Error(w, err2.Error(), http.StatusBadRequest)
			// return
			normalJump = true
		}
		// response["policy"] = p
	}

	if !normalJump {
		client := NewClient(ak, sk, arn, "default")
		stsResp, err := client.AssumeRole(900)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.DefaultLogger.Error(err.Error())
			// response["err"] = err.Error()
			// response["message"] = err.Error()
			// w.Header().Set("Content-Type", "application/json")
			// w.WriteHeader(http.StatusInternalServerError)
			// json.NewEncoder(w).Encode(response)
			return
		}
		id := stsResp.Credentials.AccessKeyId
		secret := stsResp.Credentials.AccessKeySecret
		token := stsResp.Credentials.SecurityToken

		// 使用STS Token换取控制台Signin Token
		SigninResp, err := getSigninToken(id, secret, token)
		if err != nil {
			panic(err)
		}
		signinToken := SigninResp.SigninToken

		// 生成登录链接
		loginUrl := "http://www.aliyun.com"
		// destination := "http://sls4service.console.aliyun.com"
		destination := "http://sls4service.console.aliyun.com/lognext/project/" + prj + logstoreType + logstore + "?isShare=true&hideTopbar=true&hideSidebar=true&ignoreTabLocalStorage=true&" + data.Encoding
		url, err := genSigninUrl(signinToken, loginUrl, destination)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.DefaultLogger.Error(err.Error())
			return
		}

		response["url"] = url
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		log.DefaultLogger.Debug("Goto SLS with STS success.", url)
		return
	}
	url := "https://sls.console.aliyun.com/lognext/project/" + prj + logstoreType + logstore + "?" + data.Encoding
	response["url"] = url
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	log.DefaultLogger.Debug("Goto SLS with Normal jump success.", url)
}

func getSigninToken(id string, secret string, token string) (*SigninResponse, error) {
	urlStr := "http://signin.aliyun.com/federation?Action=GetSigninToken"
	urlStr += "&AccessKeyId=" + id
	urlStr += "&AccessKeySecret=" + secret
	urlStr += "&SecurityToken=" + url.QueryEscape(token)
	urlStr += "&TicketType=mini"

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = network.DefaultNetworkFilter.FilterHttpDialContext(transport.DialContext)
	client := &http.Client{
		Transport: transport,
	}

	res, err := client.Get(urlStr)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	// fmt.Println("SigninToken json:", string(body))
	resp := SigninResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func genSigninUrl(signinToken string, loginUrl string, destination string) (string, error) {
	urlStr := "http://signin.aliyun.com/federation?Action=Login"
	urlStr += "&LoginUrl=" + url.QueryEscape(loginUrl)
	urlStr += "&Destination=" + url.QueryEscape(destination)
	urlStr += "&SigninToken=" + url.QueryEscape(signinToken)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Get(urlStr)
	if err != nil {
		return "", err
	}
	location, err := res.Location()
	if err != nil {
		return "", err
	}
	locationUrl := location.String()
	return locationUrl, nil
}

type SigninResponse struct {
	SigninToken string
}

package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
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
	_ backend.QueryDataHandler      = (*AnnotationDatasource)(nil)
	_ backend.CheckHealthHandler    = (*AnnotationDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*AnnotationDatasource)(nil)
)

// NewAnnotationDatasource creates a new datasource instance.
func NewAnnotationDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &AnnotationDatasource{}, nil
}

// AnnotationDatasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type AnnotationDatasource struct{}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *AnnotationDatasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *AnnotationDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData called", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type optionModel struct {
	ServerUrl string `json:"serverUrl"`
}

type queryModel struct {
	Tags string `json:"tags"`
}

type annoModel struct {
	Time int64    `json:"time"`
	Text string   `json:"text"`
	Tags []string `json:"tags"`
}

func (d *AnnotationDatasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}

	// Unmarshal the JSON into our queryModel.
	var qm queryModel

	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Error parsing query: %s", response.Error))
		return response
	}

	// create data frame response.
	frame := data.NewFrame("response")
	// add fields.
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, []time.Time{}),
		data.NewField("value", nil, []float64{}),
	)

	var apiKey = pCtx.DataSourceInstanceSettings.DecryptedSecureJSONData["apiKey"]
	var options optionModel
	response.Error = json.Unmarshal(pCtx.DataSourceInstanceSettings.JSONData, &options)
	if response.Error != nil {
		log.DefaultLogger.Error("Error unmarshal options: %s", response.Error)
		return response
	}

	client, err := httpclient.New(httpclient.Options{
		Timeouts: &httpclient.TimeoutOptions{
			Timeout: 15 * time.Second,
		},
	})
	if err != nil {
		response.Error = err
		log.DefaultLogger.Error("Error create http client: %s", response.Error)
		return response
	}

	url := fmt.Sprintf("%s/api/annotations?tags=%s&limit=200&type=annotation&from=%d&to=%d",
		options.ServerUrl, qm.Tags, query.TimeRange.From.Unix()*1000, query.TimeRange.To.Unix()*1000)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		response.Error = err
		log.DefaultLogger.Error("Error create http request: %s", response.Error)
		return response
	}
	req.Header.Set("Authorization", fmt.Sprint("Bearer ", apiKey))

	log.DefaultLogger.Debug(fmt.Sprintf("Request URL is: %s\n", url))
	resp, err := client.Do(req)
	if err != nil {
		response.Error = err
		log.DefaultLogger.Error("Error submit request: %s", response.Error)
		return response
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		response.Error = err
		log.DefaultLogger.Error("Error read response: %s", response.Error)
		return response
	}
	log.DefaultLogger.Debug(fmt.Sprintf("Response is: %s\n", body))

	var annotations []annoModel
	response.Error = json.Unmarshal(body, &annotations)
	if response.Error != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Error parsing body. %s", response.Error))
		return response
	}
	log.DefaultLogger.Debug(fmt.Sprintf("Parsed response is: %v\n", annotations))

	for _, anno := range annotations {
		frame.AppendRow(time.Unix(anno.Time/1000, 0), float64(1))
	}

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *AnnotationDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

package kairos_test

import (
	"context"
	"encoding/json"
	"github.com/grafana/grafana_plugin_model/go/datasource"
	"github.com/stretchr/testify/assert"
	"github.com/zsabin/kairosdb-datasource/pkg/kairos"
	"github.com/zsabin/kairosdb-datasource/pkg/panel"
	"testing"
)

type MockKairosDBClient struct {
	response []*kairos.MetricQueryResults
}

func (m MockKairosDBClient) QueryMetrics(ctx context.Context, dsInfo *datasource.DatasourceInfo, request *kairos.MetricQueryRequest) ([]*kairos.MetricQueryResults, error) {
	return m.response, nil
}

func TestDatasource_CreateMetricQuery_MinimalQuery(t *testing.T) {
	ds := &kairos.Datasource{}

	panelQuery := &panel.MetricQuery{
		Name: "MetricA",
	}

	dsQuery := &datasource.Query{
		RefId:     "A",
		ModelJson: toModelJson(panelQuery),
	}

	expectedQuery := &kairos.MetricQuery{
		Name: "MetricA",
	}

	actualQuery, err := ds.CreateMetricQuery(dsQuery)

	assert.Nil(t, err)
	assert.Equal(t, expectedQuery, actualQuery)
}

func TestDatasource_CreateMetricQuery_WithTags(t *testing.T) {
	ds := &kairos.Datasource{}

	panelQuery := &panel.MetricQuery{
		Name: "MetricA",
		Tags: map[string][]string{
			"foo":  {"bar", "baz"},
			"foo1": {},
		},
	}

	dsQuery := &datasource.Query{
		RefId:     "A",
		ModelJson: toModelJson(panelQuery),
	}

	expectedQuery := &kairos.MetricQuery{
		Name: "MetricA",
		Tags: map[string][]string{
			"foo": {"bar", "baz"},
		},
	}

	actualQuery, err := ds.CreateMetricQuery(dsQuery)

	assert.Nil(t, err)
	assert.Equal(t, expectedQuery, actualQuery)
}

func TestDatasource_CreateMetricQuery_WithAggregators(t *testing.T) {
	ds := &kairos.Datasource{}

	panelQuery := &panel.MetricQuery{
		Name: "MetricA",
		Aggregators: []*panel.Aggregator{
			{
				Name: "sum",
				Parameters: []*panel.AggregatorParameter{
					{
						Name:  "value",
						Value: "1",
					},
					{
						Name:  "unit",
						Value: "MINUTES",
					},
					{
						Name:  "sampling",
						Value: "NONE",
					},
				},
			},
			{
				Name: "avg",
				Parameters: []*panel.AggregatorParameter{
					{
						Name:  "value",
						Value: "1",
					},
					{
						Name:  "unit",
						Value: "MINUTES",
					},
					{
						Name:  "sampling",
						Value: "SAMPLING",
					},
				},
			},
			{
				Name: "max",
				Parameters: []*panel.AggregatorParameter{
					{
						Name:  "value",
						Value: "1",
					},
					{
						Name:  "unit",
						Value: "MINUTES",
					},
					{
						Name:  "sampling",
						Value: "START_TIME",
					},
				},
			},
		},
	}

	dsQuery := &datasource.Query{
		RefId:     "A",
		ModelJson: toModelJson(panelQuery),
	}

	expectedQuery := &kairos.MetricQuery{
		Name: "MetricA",
		Aggregators: []*kairos.Aggregator{
			{
				Name:           "sum",
				AlignSampling:  false,
				AlignStartTime: false,
				AlignEndTime:   false,
				Sampling: &kairos.Sampling{
					Value: 1,
					Unit:  "MINUTES",
				},
			},
			{
				Name:           "avg",
				AlignSampling:  true,
				AlignStartTime: false,
				AlignEndTime:   false,
				Sampling: &kairos.Sampling{
					Value: 1,
					Unit:  "MINUTES",
				},
			},
			{
				Name:           "max",
				AlignSampling:  false,
				AlignStartTime: true,
				AlignEndTime:   false,
				Sampling: &kairos.Sampling{
					Value: 1,
					Unit:  "MINUTES",
				},
			},
		},
	}

	actualQuery, err := ds.CreateMetricQuery(dsQuery)

	assert.Nil(t, err)
	assert.Equal(t, expectedQuery, actualQuery)
}

func TestDatasource_CreateMetricQuery_WithGroupBy(t *testing.T) {
	ds := &kairos.Datasource{}

	panelQuery := &panel.MetricQuery{
		Name: "MetricA",
		GroupBy: &panel.GroupBy{
			Tags: []string{"host", "pool"},
		},
	}

	dsQuery := &datasource.Query{
		ModelJson: toModelJson(panelQuery),
	}

	expectedQuery := &kairos.MetricQuery{
		Name: "MetricA",
		GroupBy: []*kairos.Grouper{
			{
				Name: "tag",
				Tags: []string{"host", "pool"},
			},
		},
	}

	actualQuery, err := ds.CreateMetricQuery(dsQuery)

	assert.Nil(t, err)
	assert.Equal(t, expectedQuery, actualQuery)
}

func TestDatasource_ParseQueryResult_SingleSeries(t *testing.T) {
	ds := &kairos.Datasource{}

	kairosResults := &kairos.MetricQueryResults{
		Results: []*kairos.MetricQueryResult{
			{
				Name: "MetricA",
				Values: []*kairos.DataPoint{
					{
						1564682818000, 10.5,
					},
					{
						1564682819000, 8.0,
					},
				},
			},
		},
	}

	expectedResults := &datasource.QueryResult{
		Series: []*datasource.TimeSeries{
			{
				Name: "MetricA",
				Tags: map[string]string{},
				Points: []*datasource.Point{
					{
						Timestamp: 1564682818000,
						Value:     10.5,
					},
					{
						Timestamp: 1564682819000,
						Value:     8.0,
					},
				},
			},
		},
	}

	actualResults := ds.ParseQueryResult(kairosResults)
	assert.Equal(t, expectedResults, actualResults)
}

func TestDatasource_ParseQueryResult_MultipleSeries(t *testing.T) {
	ds := &kairos.Datasource{}

	kairosResults := &kairos.MetricQueryResults{
		Results: []*kairos.MetricQueryResult{
			{
				Name: "MetricA",
				GroupInfo: []*kairos.GroupInfo{
					{
						Name: "tag",
						Tags: []string{"host", "pool"},
						Group: map[string]string{
							"host":        "server1",
							"data_center": "dc1",
						},
					},
				},
				Values: []*kairos.DataPoint{
					{
						1564682818000, 10.5,
					},
				},
			},
			{
				Name: "MetricA",
				GroupInfo: []*kairos.GroupInfo{
					{
						Name: "tag",
						Tags: []string{"host", "pool"},
						Group: map[string]string{
							"host":        "server2",
							"data_center": "dc2",
						},
					},
				},
				Values: []*kairos.DataPoint{
					{
						1564682818000, 10.5,
					},
				},
			},
		},
	}

	expectedResults := &datasource.QueryResult{
		Series: []*datasource.TimeSeries{
			{
				Name: "MetricA",
				Tags: map[string]string{
					"host":        "server1",
					"data_center": "dc1",
				},
				Points: []*datasource.Point{
					{
						Timestamp: 1564682818000,
						Value:     10.5,
					},
				},
			},
			{
				Name: "MetricA",
				Tags: map[string]string{
					"host":        "server2",
					"data_center": "dc2",
				},
				Points: []*datasource.Point{
					{
						Timestamp: 1564682818000,
						Value:     10.5,
					},
				},
			},
		},
	}

	actualResults := ds.ParseQueryResult(kairosResults)
	assert.Equal(t, expectedResults, actualResults)
}

func TestDatasource_Query(t *testing.T) {
	mockClient := &MockKairosDBClient{}

	ds := &kairos.Datasource{
		KairosDBClient: mockClient,
	}

	mockClient.response = []*kairos.MetricQueryResults{
		{
			Results: []*kairos.MetricQueryResult{
				{
					Name: "MetricA",
					Values: []*kairos.DataPoint{
						{
							1564682818000, 5,
						},
					},
				},
			},
		},
		{
			Results: []*kairos.MetricQueryResult{
				{
					Name: "MetricB",
					Values: []*kairos.DataPoint{
						{
							1564682818000, 10.5,
						},
					},
				},
			},
		},
	}

	dsRequest := &datasource.DatasourceRequest{
		TimeRange: &datasource.TimeRange{
			FromEpochMs: 1564682808000,
			ToEpochMs:   1564682828000,
		},
		Queries: []*datasource.Query{
			{
				RefId: "A",
				ModelJson: toModelJson(&panel.MetricQuery{
					Name: "MetricA",
				}),
			},
			{
				RefId: "B",
				ModelJson: toModelJson(&panel.MetricQuery{
					Name: "MetricB",
				}),
			},
		},
	}

	expectedResponse := &datasource.DatasourceResponse{
		Results: []*datasource.QueryResult{
			{
				RefId: "A",
				Series: []*datasource.TimeSeries{
					{
						Name: "MetricA",
						Tags: map[string]string{},
						Points: []*datasource.Point{
							{
								Timestamp: 1564682818000,
								Value:     5,
							},
						},
					},
				},
			},
			{
				RefId: "B",
				Series: []*datasource.TimeSeries{
					{
						Name: "MetricB",
						Tags: map[string]string{},
						Points: []*datasource.Point{
							{
								Timestamp: 1564682818000,
								Value:     10.5,
							},
						},
					},
				},
			},
		},
	}

	actualResponse, err := ds.Query(context.TODO(), dsRequest)

	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, actualResponse)
}

func toModelJson(query *panel.MetricQuery) string {
	req := panel.MetricRequest{
		Query: query,
	}
	rBytes, err := json.Marshal(req)
	if err != nil {
		panic("Failed to marshall metric request")
	}
	return string(rBytes)
}

package main

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"os"
	"strings"
)

type MetricsSubmitter struct {
	requestCountMetricName         string
	targetProcessingTimeMetricName string
	customTags                     []Tag
}

func NewMetricsSubmitter(requestCountMetricName string, targetProcessingTimeMetricName string, customTags []Tag) *MetricsSubmitter {
	return &MetricsSubmitter{
		requestCountMetricName:         requestCountMetricName,
		targetProcessingTimeMetricName: targetProcessingTimeMetricName,
		customTags:                     customTags}
}

func (p *MetricsSubmitter) Submit(metrics map[string]*Metric, s3ObjectKey string) error {
	var eg errgroup.Group

	ctx := datadog.NewDefaultContext(context.Background())
	apiClient := datadog.NewAPIClient(datadog.NewConfiguration())
	// v1Api is used to submit distribution points api. This api is defined by v1, so need client for it.
	v1Api := datadogV1.NewMetricsApi(apiClient)
	// v2Api is used to submit metrics api. This api is defined by v1 and v2, but v1 api is deprecated.
	v2Api := datadogV2.NewMetricsApi(apiClient)

	for _, metric := range metrics {
		metricsPayload := datadogV2.MetricPayload{
			Series: []datadogV2.MetricSeries{
				p.requestCountSeries(metric, s3ObjectKey),
			},
		}

		distributionPointPayload := datadogV1.DistributionPointsPayload{}
		s, err := p.targetProcessingTime(metric)
		if err != nil {
			return err
		}
		distributionPointPayload.Series = append(distributionPointPayload.Series, s...)

		eg.Go(func() error {
			_, r, err := v1Api.SubmitDistributionPoints(ctx, distributionPointPayload, *datadogV1.NewSubmitDistributionPointsOptionalParameters().WithContentEncoding(datadogV1.DISTRIBUTIONPOINTSCONTENTENCODING_DEFLATE))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.SubmitDistributionPoints`: %v\n", err)
				fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
				return err
			}
			return nil
		})

		if p.requestCountMetricName != "" {
			eg.Go(func() error {
				_, r, err := v2Api.SubmitMetrics(ctx, metricsPayload, *datadogV2.NewSubmitMetricsOptionalParameters())
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.SubmitMetrics`: %v\n", err)
					fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
					return err
				}
				return nil
			})
		}
	}

	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "SubmitMetrics fails:")
	}

	return nil
}

func (p *MetricsSubmitter) requestCountSeries(metric *Metric, s3ObjectKey string) datadogV2.MetricSeries {
	var points []datadogV2.MetricPoint
	for timestamp, count := range metric.RequestCountMap {
		points = append(points, datadogV2.MetricPoint{
			Timestamp: timestamp.PtrInt64(),
			Value:     count.PtrFloat64(),
		})
	}
	series := datadogV2.NewMetricSeries(p.requestCountMetricName, points)
	series.SetType(datadogV2.METRICINTAKETYPE_COUNT)
	series.SetInterval(60)
	series.SetUnit("request")
	tags := []string{
		fmt.Sprintf("elb:%s", metric.Elb),
		fmt.Sprintf("target_group_arn:%s", metric.TargetGroupArn),
		fmt.Sprintf("path:%s", metric.Path),
		fmt.Sprintf("method:%s", metric.Method),
		fmt.Sprintf("elb_status_code:%s", metric.ElbStatusCode),
		fmt.Sprintf("target_status_code:%s", metric.TargetStatusCode),
		fmt.Sprintf("ip_address:%s", p.loadBalancerIpAddress(s3ObjectKey)),
	}
	for _, tag := range p.customTags {
		tags = append(tags, fmt.Sprintf("%s:%s", tag.Name, tag.Key()))
	}
	series.SetTags(tags)
	return *series
}

func (p *MetricsSubmitter) targetProcessingTime(metric *Metric) ([]datadogV1.DistributionPointsSeries, error) {
	seriesSlice := make([]datadogV1.DistributionPointsSeries, 1)
	points := make([][]datadogV1.DistributionPointItem, 0, len(metric.TargetProcessingTimesMap))

	for timestamp, times := range metric.TargetProcessingTimesMap {
		points = append(points, []datadogV1.DistributionPointItem{
			{DistributionPointTimestamp: timestamp.PtrFloat64()},
			{DistributionPointData: times.Float64()},
		})
	}

	series := datadogV1.NewDistributionPointsSeries(p.targetProcessingTimeMetricName, points)
	tags := []string{
		fmt.Sprintf("elb:%s", metric.Elb),
		fmt.Sprintf("target_group_arn:%s", metric.TargetGroupArn),
		fmt.Sprintf("path:%s", metric.Path),
		fmt.Sprintf("method:%s", metric.Method),
		fmt.Sprintf("elb_status_code:%s", metric.ElbStatusCode),
		fmt.Sprintf("target_status_code:%s", metric.TargetStatusCode),
		fmt.Sprintf("target_status_code_group:%s", metric.TargetStatusCodeGroup()),
	}
	for _, tag := range p.customTags {
		tags = append(tags, fmt.Sprintf("%s:%s", tag.Name, tag.Key()))
	}
	series.SetTags(tags)
	seriesSlice[0] = *series
	return seriesSlice, nil
}

func (p *MetricsSubmitter) loadBalancerIpAddress(s string) string {
	sl := strings.Split(s, "/")
	return strings.Split(sl[len(sl)-1], "_")[5]
}

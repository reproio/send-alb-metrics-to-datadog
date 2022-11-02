package main

import (
	"context"
	"fmt"
	v1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/DataDog/datadog-api-client-go/api/v2/datadog"
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

	// v1ctx and v1client are used to submit distribution points api. This api is defined by v1, so need client for it.
	v1ctx := v1.NewDefaultContext(context.Background())
	v1client := v1.NewAPIClient(v1.NewConfiguration())

	// v2ctx and v2client are used to submit metrics api. This api is defined by v1 and v2, but v1 api is deprecated.
	v2ctx := datadog.NewDefaultContext(context.Background())
	v2client := datadog.NewAPIClient(datadog.NewConfiguration())

	for _, metric := range metrics {
		payload := datadog.MetricPayload{
			Series: []datadog.MetricSeries{
				p.requestCountSeries(metric, s3ObjectKey),
			},
		}

		v1payload := v1.DistributionPointsPayload{}
		s, err := p.targetProcessingTime(metric)
		if err != nil {
			return err
		}
		v1payload.Series = append(v1payload.Series, s...)

		eg.Go(func() error {
			_, r, err := v1client.MetricsApi.SubmitDistributionPoints(v1ctx, v1payload, *v1.NewSubmitDistributionPointsOptionalParameters().WithContentEncoding(v1.DISTRIBUTIONPOINTSCONTENTENCODING_DEFLATE))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.SubmitDistributionPoints`: %v\n", err)
				fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
				return err
			}
			return nil
		})

		eg.Go(func() error {
			_, r, err := v2client.MetricsApi.SubmitMetrics(v2ctx, payload, *datadog.NewSubmitMetricsOptionalParameters())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.SubmitMetrics`: %v\n", err)
				fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
				return err
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "SubmitMetrics fails:")
	}

	return nil
}

func (p *MetricsSubmitter) requestCountSeries(metric *Metric, s3ObjectKey string) datadog.MetricSeries {
	var points []datadog.MetricPoint
	for timestamp, count := range metric.RequestCountMap {
		points = append(points, datadog.MetricPoint{
			Timestamp: timestamp.PtrInt64(),
			Value:     count.PtrFloat64(),
		})
	}
	series := datadog.NewMetricSeries(p.requestCountMetricName, points)
	series.SetType(datadog.METRICINTAKETYPE_COUNT)
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

func (p *MetricsSubmitter) targetProcessingTime(metric *Metric) ([]v1.DistributionPointsSeries, error) {
	seriesSlice := make([]v1.DistributionPointsSeries, 1)
	points := make([][]v1.DistributionPointItem, 0, len(metric.TargetProcessingTimesMap))

	for timestamp, times := range metric.TargetProcessingTimesMap {
		points = append(points, []v1.DistributionPointItem{
			{DistributionPointTimestamp: timestamp.PtrFloat64()},
			{DistributionPointData: times.Float64()},
		})
	}

	series := v1.NewDistributionPointsSeries(p.targetProcessingTimeMetricName, points)
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

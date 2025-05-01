package main

import (
	"bufio"
	"fmt"
	"golang.org/x/exp/slices"
	"io"
	"strings"
)

type LogFileReader struct {
	pathTransformingRules []PathTransformingRule
	targetPaths           []string
}

func NewLogFileReader(paths []PathTransformingRule, targetPaths []string) *LogFileReader {
	return &LogFileReader{pathTransformingRules: paths, targetPaths: targetPaths}
}

func (p *LogFileReader) Read(r io.Reader) (map[string]*Metric, error) {
	scanner := bufio.NewScanner(r)

	metricMap := map[string]*Metric{}
	for scanner.Scan() {
		text := scanner.Text()
		r, err := NewAlbLogRecord(text, p.pathTransformingRules)
		if err != nil {
			fmt.Printf("failed to read alb log record: %s\n", text)
			return nil, err
		}

		if slices.Contains(p.targetPaths, r.RequestPath) {
			metricKey := r.MetricKey()

			// When metricMap doesn't have key of `metricsKey`, add new Metrics.
			if _, ok := metricMap[metricKey]; !ok {
				metricMap[metricKey] = &Metric{
					RequestCountMap:          map[Timestamp]RequestCount{},
					TargetProcessingTimesMap: map[Timestamp]TargetProcessingTimes{},
					Method:                   r.RequestMethod,
					Path:                     r.RequestPath,
					ElbStatusCode:            r.ElbStatusCode,
					TargetStatusCode:         r.TargetStatusCode,
					Elb:                      r.Elb,
					TargetGroupArn:           r.TargetGroupArn,
				}
			}
			metric := metricMap[metricKey]

			// When RequestCountMap of metric doesn't have key of timestamp, initialize value to 1.
			// When RequestCountMap of metric have key of timestamp, increment value.
			if _, exist := metric.RequestCountMap[r.Timestamp()]; exist {
				metric.RequestCountMap[r.Timestamp()] = metric.RequestCountMap[r.Timestamp()] + RequestCount(1)
			} else {
				metric.RequestCountMap[r.Timestamp()] = RequestCount(1)
			}

			targetProcessingTime := TargetProcessingTime(r.TargetProcessingTime)
			if _, exist := metric.RequestCountMap[r.Timestamp()]; exist {
				metric.TargetProcessingTimesMap[r.Timestamp()] = append(metric.TargetProcessingTimesMap[r.Timestamp()], targetProcessingTime)
			} else {
				metric.TargetProcessingTimesMap[r.Timestamp()] = TargetProcessingTimes{targetProcessingTime}
			}
			metricMap[metricKey] = metric
		}
	}
	return metricMap, nil
}

type Timestamp int64

func (ts *Timestamp) PtrInt64() *int64 {
	v := int64(*ts)
	return &v
}
func (ts *Timestamp) PtrFloat64() *float64 {
	v := float64(*ts)
	return &v
}

type RequestCount float64

func (c *RequestCount) PtrFloat64() *float64 {
	v := float64(*c)
	return &v
}

type TargetProcessingTime float64

func (t *TargetProcessingTime) PtrFloat64() *float64 {
	v := float64(*t)
	return &v
}

type TargetProcessingTimes []TargetProcessingTime

func (t *TargetProcessingTimes) Float64() *[]float64 {
	r := make([]float64, len(*t))

	for i, f := range *t {
		r[i] = float64(f)
	}
	return &r
}

type Metric struct {
	RequestCountMap          map[Timestamp]RequestCount
	TargetProcessingTimesMap map[Timestamp]TargetProcessingTimes
	Method                   string
	Path                     string
	ElbStatusCode            string
	TargetStatusCode         string
	Elb                      string
	TargetGroupArn           string
}

func (m *Metric) TargetStatusCodeGroup() string {
	// TargetStatusCode is - when the target does not send a response
	// see: https://docs.aws.amazon.com/ja_jp/elasticloadbalancing/latest/application/load-balancer-access-logs.html
	code := "-"
	switch {
	case strings.HasPrefix(m.TargetStatusCode, "1"):
		code = "1xx"
	case strings.HasPrefix(m.TargetStatusCode, "2"):
		code = "2xx"
	case strings.HasPrefix(m.TargetStatusCode, "3"):
		code = "3xx"
	case strings.HasPrefix(m.TargetStatusCode, "4"):
		code = "4xx"
	case strings.HasPrefix(m.TargetStatusCode, "5"):
		code = "5xx"
	}
	return code
}

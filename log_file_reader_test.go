package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLogFileReader_Read(t *testing.T) {
	var pathRules []PathTransformingRule
	targetPaths := []string{"/"}
	logFileReader := NewLogFileReader(pathRules, targetPaths)

	logTimeString := "2022-06-13T00:26:00.071316Z"
	logTime, err := time.Parse(time.RFC3339, "2022-06-13T00:26:00.071316Z")
	if err != nil {
		fmt.Printf("failed to parse %s: %s", logTimeString, err.Error())
	}

	cases := []struct {
		log                          string
		wantTargetProcessingTimesMap map[Timestamp]TargetProcessingTimes
		wantTargetRequestCountMap    map[Timestamp]RequestCount
	}{
		{
			log: fmt.Sprintf(`https %s app/my-loadbalancer/50dc6c495c0c9188 192.168.131.39:2817 10.0.0.1:80 0.086 0.048 0.037 200 200 0 57 "GET https://www.example.com:443/ HTTP/1.1" "curl/7.46.0" ECDHE-RSA-AES128-GCM-SHA256 TLSv1.2 arn:aws:elasticloadbalancing:us-east-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067 "Root=1-58337281-1d84f3d73c47ec4e58577259" "www.example.com" "arn:aws:acm:us-east-2:123456789012:certificate/12345678-1234-1234-1234-123456789012" 1 2018-07-02T22:22:48.364000Z "authenticate,forward" "-" "-" "10.0.0.1:80" "200" "-" "-"`, logTimeString),
			wantTargetProcessingTimesMap: map[Timestamp]TargetProcessingTimes{
				Timestamp(logTime.Unix()): {0.048},
			},
			wantTargetRequestCountMap: map[Timestamp]RequestCount{
				Timestamp(logTime.Unix()): 1,
			},
		},
		{
			log: fmt.Sprintf(`https %s app/my-loadbalancer/50dc6c495c0c9188 192.168.131.39:2817 - -1 -1 -1 400 - 235 772 "GET https://www.example.com:443/ HTTP/1.1" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36" - - - "-" "-" "-" - 2022-06-13T00:25:59.856000Z "-" "-" "-" "-" "-" "-" "-"`, logTimeString),
			wantTargetProcessingTimesMap: map[Timestamp]TargetProcessingTimes{
				Timestamp(logTime.Unix()): {-1},
			},
			wantTargetRequestCountMap: map[Timestamp]RequestCount{
				Timestamp(logTime.Unix()): 1,
			},
		},
	}
	for _, tt := range cases {
		metricMap, err := logFileReader.Read(
			strings.NewReader(tt.log),
		)
		if err != nil {
			t.Fatalf("failed to read the log: %v\n%s", err, tt.log)
		}
		if len(metricMap) != 1 {
			t.Errorf("expected 1 metric, got %d", len(metricMap))
		}

		for _, metric := range metricMap {
			if reflect.DeepEqual(metric.TargetProcessingTimesMap, tt.wantTargetProcessingTimesMap) == false {
				t.Errorf("unexpected got %v, want, %v", metric.TargetProcessingTimesMap, tt.wantTargetProcessingTimesMap)
			}
			if reflect.DeepEqual(metric.RequestCountMap, tt.wantTargetRequestCountMap) == false {
				t.Errorf("unexpected got %v, want, %v", metric.RequestCountMap, tt.wantTargetRequestCountMap)
			}

		}
	}
}

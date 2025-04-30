package main

import (
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"reflect"
	"testing"
)

func TestMetricsSubmitter_loadBalancerIpAddress(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "with prefix",
			args: args{
				s: "s3://my-bucket/prefix/AWSLogs/123456789012/elasticloadbalancing/us-east-2/2016/05/01/123456789012_elasticloadbalancing_us-east-2_net.app.my-loadbalancer.1234567890abcdef_20140215T2340Z_172.160.001.192_20sg8hgm.log.gz",
			},
			want: "172.160.001.192",
		},
		{
			name: "with prefix includes underscore",
			args: args{
				s: "s3://my-bucket/pre_fix/AWSLogs/123456789012/elasticloadbalancing/us-east-2/2016/05/01/123456789012_elasticloadbalancing_us-east-2_net.app.my-loadbalancer.1234567890abcdef_20140215T2340Z_172.160.001.192_20sg8hgm.log.gz",
			},
			want: "172.160.001.192",
		},
		{
			name: "without prefix",
			args: args{
				s: "s3://my-bucket/AWSLogs/123456789012/elasticloadbalancing/us-east-2/2016/05/01/123456789012_elasticloadbalancing_us-east-2_net.app.my-loadbalancer.1234567890abcdef_20140215T2340Z_172.160.001.192_20sg8hgm.log.gz",
			},
			want: "172.160.001.192",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &MetricsSubmitter{}
			if got := p.loadBalancerIpAddress(tt.args.s); got != tt.want {
				t.Errorf("loadBalancerIpAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricsSubmitter_targetProcessingTime(t *testing.T) {
	var typeVar datadogV1.DistributionPointsType = datadogV1.DISTRIBUTIONPOINTSTYPE_DISTRIBUTION
	type fields struct {
		RequestCountMetricName         string
		TargetProcessingTimeMetricName string
		CustomTags                     []Tag
	}
	type args struct {
		metric *Metric
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []datadogV1.DistributionPointsSeries
		wantErr bool
	}{
		{
			name: "",
			fields: fields{
				TargetProcessingTimeMetricName: "target_processing_time",
			},
			args: args{
				metric: &Metric{
					TargetProcessingTimesMap: map[Timestamp]TargetProcessingTimes{
						Timestamp(1): {1, 2},
						Timestamp(2): {3, 4},
					},
					Method:           "GET",
					Path:             "/",
					ElbStatusCode:    "200",
					TargetStatusCode: "200",
					Elb:              "elb",
					TargetGroupArn:   "arn",
				},
			},
			want: []datadogV1.DistributionPointsSeries{
				datadogV1.DistributionPointsSeries{
					Metric: "target_processing_time",
					Points: [][]datadogV1.DistributionPointItem{
						{
							{DistributionPointTimestamp: datadog.PtrFloat64(1)},
							{DistributionPointData: &[]float64{1, 2}},
						},
						{
							{DistributionPointTimestamp: datadog.PtrFloat64(2)},
							{DistributionPointData: &[]float64{3, 4}},
						},
					},
					Tags: []string{
						"elb:elb",
						"target_group_arn:arn",
						"path:/",
						"method:GET",
						"elb_status_code:200",
						"target_status_code:200",
						"target_status_code_group:2xx",
						"ip_address:172.160.001.192",
					},
					Type: &typeVar,
				},
			},
			wantErr: false,
		},
		{
			name: "TargetStatusCode is -",
			fields: fields{
				TargetProcessingTimeMetricName: "target_processing_time",
			},
			args: args{
				metric: &Metric{
					TargetProcessingTimesMap: map[Timestamp]TargetProcessingTimes{
						Timestamp(1): {-1},
						Timestamp(2): {-1},
					},
					Method:           "GET",
					Path:             "/",
					ElbStatusCode:    "460",
					TargetStatusCode: "-",
					Elb:              "elb",
					TargetGroupArn:   "arn",
				},
			},
			want: []datadogV1.DistributionPointsSeries{
				datadogV1.DistributionPointsSeries{
					Metric: "target_processing_time",
					Points: [][]datadogV1.DistributionPointItem{
						{
							{DistributionPointTimestamp: datadog.PtrFloat64(1)},
							{DistributionPointData: &[]float64{-1}},
						},
						{
							{DistributionPointTimestamp: datadog.PtrFloat64(2)},
							{DistributionPointData: &[]float64{-1}},
						},
					},
					Tags: []string{
						"elb:elb",
						"target_group_arn:arn",
						"path:/",
						"method:GET",
						"elb_status_code:460",
						"target_status_code:-",
						"target_status_code_group:-",
						"ip_address:172.160.001.192",
					},
					Type: &typeVar,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &MetricsSubmitter{
				requestCountMetricName:         tt.fields.RequestCountMetricName,
				targetProcessingTimeMetricName: tt.fields.TargetProcessingTimeMetricName,
				customTags:                     tt.fields.CustomTags,
			}
			got, err := p.targetProcessingTime(tt.args.metric, "s3://my-bucket/my-prefix/AWSLogs/123456789012/elasticloadbalancing/us-east-2/2022/05/01/123456789012_elasticloadbalancing_us-east-2_app.my-loadbalancer.1234567890abcdef_20220215T2340Z_172.160.001.192_20sg8hgm.log.gz")
			if (err != nil) != tt.wantErr {
				t.Errorf("targetProcessingTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("targetProcessingTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}

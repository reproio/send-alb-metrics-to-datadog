package main

import (
	"reflect"
	"regexp"
	"testing"
	"time"
)

// this is ALB's access log record: https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-access-logs.html
var (
	exampleHttpEntry                    = `http 2018-07-02T22:23:00.186641Z app/my-loadbalancer/50dc6c495c0c9188 192.168.131.39:2817 10.0.0.1:80 0.000 0.001 0.000 200 200 34 366 "GET http://www.example.com:80/ HTTP/1.1" "curl/7.46.0" - - arn:aws:elasticloadbalancing:us-east-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067 "Root=1-58337262-36d228ad5d99923122bbe354" "-" "-" 0 2018-07-02T22:22:48.364000Z "forward" "-" "-" "10.0.0.1:80" "200" "-" "-"`
	exampleHttpsEntry                   = `https 2018-07-02T22:23:00.186641Z app/my-loadbalancer/50dc6c495c0c9188 192.168.131.39:2817 10.0.0.1:80 0.086 0.048 0.037 200 200 0 57 "GET https://www.example.com:443/ HTTP/1.1" "curl/7.46.0" ECDHE-RSA-AES128-GCM-SHA256 TLSv1.2 arn:aws:elasticloadbalancing:us-east-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067 "Root=1-58337281-1d84f3d73c47ec4e58577259" "www.example.com" "arn:aws:acm:us-east-2:123456789012:certificate/12345678-1234-1234-1234-123456789012" 1 2018-07-02T22:22:48.364000Z "authenticate,forward" "-" "-" "10.0.0.1:80" "200" "-" "-"`
	exampleLoadBalancerCouldNotDispatch = `https 2022-06-13T00:26:00.071316Z app/my-loadbalancer/50dc6c495c0c9188 192.168.131.39:2817 - -1 -1 -1 400 - 235 772 "GET https://www.example.com:443/ HTTP/1.1" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36" - - - "-" "-" "-" - 2022-06-13T00:25:59.856000Z "-" "-" "-" "-" "-" "-" "-"`
)

func Test_parseAlbLog(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name    string
		args    args
		want    *AlbLogRecord
		wantErr bool
	}{
		{
			name: "Example HTTP Entry of document",
			args: args{
				line: exampleHttpEntry,
			},
			want: &AlbLogRecord{
				Type: "http",
				Time: func() time.Time {
					r, err := time.Parse(time.RFC3339, "2018-07-02T22:23:00.186641Z")
					if err != nil {
						t.Fatalf(err.Error())
					}
					return r
				}(),
				Elb:                    "app/my-loadbalancer/50dc6c495c0c9188",
				ClientPort:             "192.168.131.39:2817",
				TargetPort:             "10.0.0.1:80",
				RequestProcessingTime:  0.000,
				TargetProcessingTime:   0.001,
				ResponseProcessingTime: 0.000,
				ElbStatusCode:          "200",
				TargetStatusCode:       "200",
				ReceivedBytes:          34,
				SentBytes:              366,
				Request:                "GET http://www.example.com:80/ HTTP/1.1",
				UserAgent:              "curl/7.46.0",
				SslCipher:              "-",
				SslProtocol:            "-",
				TargetGroupArn:         "arn:aws:elasticloadbalancing:us-east-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067",
				TraceId:                "Root=1-58337262-36d228ad5d99923122bbe354",
				DomainName:             "-",
				ChosenCertArn:          "-",
				MatchedRulePriority:    0,
				RequestCreationTime:    "2018-07-02T22:22:48.364000Z",
				ActionsExecuted:        "forward",
				RedirectUrl:            "-",
				ErrorReason:            "-",
				TargetPortList:         "10.0.0.1:80",
				TargetStatusCodeList:   "200",
				Classification:         "-",
				ClassificationReason:   "-",
			},
			wantErr: false,
		},
		{
			name: "Example HTTPS Entry of document",
			args: args{
				line: exampleHttpsEntry,
			},
			want: &AlbLogRecord{
				Type: "https",
				Time: func() time.Time {
					r, err := time.Parse(time.RFC3339, "2018-07-02T22:23:00.186641Z")
					if err != nil {
						t.Fatalf(err.Error())
					}
					return r
				}(),
				Elb:                    "app/my-loadbalancer/50dc6c495c0c9188",
				ClientPort:             "192.168.131.39:2817",
				TargetPort:             "10.0.0.1:80",
				RequestProcessingTime:  0.086,
				TargetProcessingTime:   0.048,
				ResponseProcessingTime: 0.037,
				ElbStatusCode:          "200",
				TargetStatusCode:       "200",
				ReceivedBytes:          0,
				SentBytes:              57,
				Request:                "GET https://www.example.com:443/ HTTP/1.1",
				UserAgent:              "curl/7.46.0",
				SslCipher:              "ECDHE-RSA-AES128-GCM-SHA256",
				SslProtocol:            "TLSv1.2",
				TargetGroupArn:         "arn:aws:elasticloadbalancing:us-east-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067",
				TraceId:                "Root=1-58337281-1d84f3d73c47ec4e58577259",
				DomainName:             "www.example.com",
				ChosenCertArn:          "arn:aws:acm:us-east-2:123456789012:certificate/12345678-1234-1234-1234-123456789012",
				MatchedRulePriority:    1,
				RequestCreationTime:    "2018-07-02T22:22:48.364000Z",
				ActionsExecuted:        "authenticate,forward",
				RedirectUrl:            "-",
				ErrorReason:            "-",
				TargetPortList:         "10.0.0.1:80",
				TargetStatusCodeList:   "200",
				Classification:         "-",
				ClassificationReason:   "-",
			},
			wantErr: false,
		},
		{
			name: "Example HTTPS Entry if LoadBalancer could not dispatch request to target",
			args: args{
				line: exampleLoadBalancerCouldNotDispatch,
			},
			want: &AlbLogRecord{
				Type: "https",
				Time: func() time.Time {
					r, err := time.Parse(time.RFC3339, "2022-06-13T00:26:00.071316Z")
					if err != nil {
						t.Fatalf(err.Error())
					}
					return r
				}(),
				Elb:                    "app/my-loadbalancer/50dc6c495c0c9188",
				ClientPort:             "192.168.131.39:2817",
				TargetPort:             "-",
				RequestProcessingTime:  -1,
				TargetProcessingTime:   -1,
				ResponseProcessingTime: -1,
				ElbStatusCode:          "400",
				TargetStatusCode:       "-",
				ReceivedBytes:          235,
				SentBytes:              772,
				Request:                "GET https://www.example.com:443/ HTTP/1.1",
				UserAgent:              "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36",
				SslCipher:              "-",
				SslProtocol:            "-",
				TargetGroupArn:         "-",
				TraceId:                "-",
				DomainName:             "-",
				ChosenCertArn:          "-",
				MatchedRulePriority:    0,
				RequestCreationTime:    "2022-06-13T00:25:59.856000Z",
				ActionsExecuted:        "-",
				RedirectUrl:            "-",
				ErrorReason:            "-",
				TargetPortList:         "-",
				TargetStatusCodeList:   "-",
				Classification:         "-",
				ClassificationReason:   "-",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAlbLog(tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAlbLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseAlbLog() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlbLogRecord_RequestPath(t *testing.T) {
	type args struct {
		paths []PathTransformingRule
	}
	type fields struct {
		Request string
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "with port number",
			args: args{
				paths: []PathTransformingRule{
					{
						Prefix:      "/example",
						Transformed: "/example",
					},
				},
			},
			fields: fields{
				Request: "GET http://example.com:80/example HTTP/1.1",
			},
			want:    "/example",
			wantErr: false,
		},
		{
			name: "without port number",
			args: args{
				paths: []PathTransformingRule{
					{
						Prefix:      "/example",
						Transformed: "/example",
					},
				},
			},
			fields: fields{
				Request: "GET http://example.com/example HTTP/1.1",
			},
			want:    "/example",
			wantErr: false,
		},
		{
			name: "",
			fields: fields{
				Request: "- http://example.com:443- -",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "",
			args: args{
				paths: []PathTransformingRule{
					{
						Prefix:      "/v1/foo/bar",
						Transformed: "/v1/foo/bar",
					},
				},
			},
			fields: fields{
				Request: "POST https://example.com:443/v1/foo/bar HTTP/1.1",
			},
			want:    "/v1/foo/bar",
			wantErr: false,
		},
		{
			name: "",
			args: args{
				paths: []PathTransformingRule{
					{
						Prefix:      "/foo",
						Suffix:      "bar",
						Transformed: "/foo/$id/bar",
					},
				},
			},
			fields: fields{
				Request: "POST https://example.com:443/foo/aaa/bar HTTP/1.1",
			},
			want:    "/foo/$id/bar",
			wantErr: false,
		},
		{
			name: "regexp",
			args: args{
				paths: []PathTransformingRule{
					{
						Regexp:      regexp.MustCompile(`^/foo/([^/]+)/bar$`),
						Transformed: "/foo/$id/bar",
					},
				},
			},
			fields: fields{
				Request: "POST https://example.com:443/foo/aaa/bar HTTP/1.1",
			},
			want:    "/foo/$id/bar",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &AlbLogRecord{Request: tt.fields.Request}

			got, err := r.requestPath(tt.args.paths)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequestPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RequestPath() got = %v, want %v", got, tt.want)
			}
		})
	}
}

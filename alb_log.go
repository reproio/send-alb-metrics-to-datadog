package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AlbLogRecord struct {
	// Fields of ALB log record
	Type                   string
	Time                   time.Time
	Elb                    string
	ClientPort             string
	TargetPort             string
	RequestProcessingTime  float64
	TargetProcessingTime   float64
	ResponseProcessingTime float64
	ElbStatusCode          string
	TargetStatusCode       string
	ReceivedBytes          int
	SentBytes              int
	Request                string
	UserAgent              string
	SslCipher              string
	SslProtocol            string
	TargetGroupArn         string
	TraceId                string
	DomainName             string
	ChosenCertArn          string
	MatchedRulePriority    int
	RequestCreationTime    string
	ActionsExecuted        string
	RedirectUrl            string
	ErrorReason            string
	TargetPortList         string
	TargetStatusCodeList   string
	Classification         string
	ClassificationReason   string

	// Fields from parsing Request.
	RequestMethod string
	RequestPath   string
}

func NewAlbLogRecord(s string, rules []PathTransformingRule) (*AlbLogRecord, error) {
	r, err := parseAlbLog(s)
	if err != nil {
		return nil, err
	}

	r.RequestMethod = r.requestMethod()
	path, err := r.requestPath(rules)
	if err != nil {
		fmt.Printf("failed to get path from request field of alb log record: %s\n", r.Request)
		return nil, err
	}
	r.RequestPath = path

	return r, nil
}

func (r *AlbLogRecord) Timestamp() Timestamp {
	return Timestamp(r.Time.Unix())
}

func (r *AlbLogRecord) MetricKey() string {
	return fmt.Sprintf("%s_%s_%s_%s_%s_%s", r.Elb, r.TargetGroupArn, r.RequestMethod, r.RequestPath, r.ElbStatusCode, r.TargetStatusCode)
}

func (r *AlbLogRecord) requestMethod() string {
	return strings.Split(r.Request, " ")[0]
}

var requestPathRe = regexp.MustCompile(`(?P<method>.*) (?P<protocol>.*)://(?P<host>[^:]*):?(?P<port>\d*)(?P<path>-|/[^\?]*)\??(?P<query_param>.*) (?P<http_version>.*)`)

type PathTransformingRule struct {
	Prefix      string
	Suffix      string
	Regexp      *regexp.Regexp
	Transformed string
}

func (r *AlbLogRecord) requestPath(rules []PathTransformingRule) (string, error) {
	match := false
	transformed := ""
	values := requestPathRe.FindStringSubmatch(r.Request)
	if values == nil {
		return "", fmt.Errorf("no match regexp: %s", r.Request)
	}
	uri := values[requestPathRe.SubexpIndex("path")]

	if uri == "-" {
		return "", nil
	}

	for _, rule := range rules {
		if rule.Regexp != nil {
			if rule.Regexp.MatchString(uri) {
				match = true
				transformed = rule.Transformed
				break
			}
		}

		if rule.Prefix != "" && rule.Suffix != "" {
			if strings.HasPrefix(uri, rule.Prefix) && strings.HasSuffix(uri, rule.Suffix) {
				match = true
				transformed = rule.Transformed
				break
			}
		}

		if rule.Prefix != "" && rule.Suffix == "" {
			if strings.HasPrefix(uri, rule.Prefix) {
				match = true
				transformed = rule.Transformed
				break
			}
		}

		if rule.Prefix == "" && rule.Suffix != "" {
			if strings.HasSuffix(uri, rule.Suffix) {
				match = true
				transformed = rule.Transformed
				break
			}
		}
	}

	if match {
		return transformed, nil
	}

	return uri, nil
}

var parseAlbLogRe = func() *regexp.Regexp {
	s := []string{
		`(?P<type>.*?)`,
		`(?P<time>.*?)`,
		`(?P<elb>.*?)`,
		`(?P<clientport>.*?)`,
		`(?P<targetport>.*?)`,
		`(?P<request_processing_time>.*?)`,
		`(?P<target_processing_time>.*?)`,
		`(?P<response_processing_time>.*?)`,
		`(?P<elb_status_code>.*?)`,
		`(?P<target_status_code>.*?)`,
		`(?P<received_bytes>.*?)`,
		`(?P<sent_bytes>.*?)`,
		`"(?P<request>.*?)"`,
		`"(?P<user_agent>.*?)"`,
		`(?P<ssl_cipher>.*?)`,
		`(?P<ssl_protocol>.*?)`,
		`(?P<target_group_arn>.*?)`,
		`"(?P<trace_id>.*?)"`,
		`"(?P<domain_name>.*?)"`,
		`"(?P<chosen_cert_arn>.*?)"`,
		`(?P<matched_rule_priority>.*?)`,
		`(?P<request_creation_time>.*?)`,
		`"(?P<actions_executed>.*?)"`,
		`"(?P<redirect_url>.*?)"`,
		`"(?P<error_reason>.*?)"`,
		`"(?P<target_port_list>.*?)"`,
		`"(?P<target_status_code_list>.*?)"`,
		`"(?P<classification>.*?)"`,
		`"(?P<classification_reason>.*?)"`,
	}
	return regexp.MustCompile(strings.Join(s, `\s`))
}()

// line is ALB's access log record: https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-access-logs.html
// field that is enclosed in double quotes may include space. e.g. `"request"` and `"target:port_list"`.
func parseAlbLog(line string) (*AlbLogRecord, error) {
	values := parseAlbLogRe.FindStringSubmatch(line)
	ts, err := time.Parse(time.RFC3339, values[parseAlbLogRe.SubexpIndex("time")])
	if err != nil {
		return nil, err
	}

	requestProcessingTime, err := strconv.ParseFloat(values[parseAlbLogRe.SubexpIndex("request_processing_time")], 64)
	if err != nil {
		return nil, err
	}
	targetProcessingTime, err := strconv.ParseFloat(values[parseAlbLogRe.SubexpIndex("target_processing_time")], 64)
	if err != nil {
		return nil, err
	}
	responseProcessingTime, err := strconv.ParseFloat(values[parseAlbLogRe.SubexpIndex("response_processing_time")], 64)
	if err != nil {
		return nil, err
	}
	record := &AlbLogRecord{
		Type:                   values[parseAlbLogRe.SubexpIndex("type")],
		Time:                   ts,
		Elb:                    values[parseAlbLogRe.SubexpIndex("elb")],
		ClientPort:             values[parseAlbLogRe.SubexpIndex("clientport")],
		TargetPort:             values[parseAlbLogRe.SubexpIndex("targetport")],
		RequestProcessingTime:  requestProcessingTime,
		TargetProcessingTime:   targetProcessingTime,
		ResponseProcessingTime: responseProcessingTime,
		ElbStatusCode:          values[parseAlbLogRe.SubexpIndex("elb_status_code")],
		TargetStatusCode:       values[parseAlbLogRe.SubexpIndex("target_status_code")],
		Request:                values[parseAlbLogRe.SubexpIndex("request")],
		UserAgent:              values[parseAlbLogRe.SubexpIndex("user_agent")],
		SslCipher:              values[parseAlbLogRe.SubexpIndex("ssl_cipher")],
		SslProtocol:            values[parseAlbLogRe.SubexpIndex("ssl_protocol")],
		TargetGroupArn:         values[parseAlbLogRe.SubexpIndex("target_group_arn")],
		TraceId:                values[parseAlbLogRe.SubexpIndex("trace_id")],
		DomainName:             values[parseAlbLogRe.SubexpIndex("domain_name")],
		ChosenCertArn:          values[parseAlbLogRe.SubexpIndex("chosen_cert_arn")],
		RequestCreationTime:    values[parseAlbLogRe.SubexpIndex("request_creation_time")],
		ActionsExecuted:        values[parseAlbLogRe.SubexpIndex("actions_executed")],
		RedirectUrl:            values[parseAlbLogRe.SubexpIndex("redirect_url")],
		ErrorReason:            values[parseAlbLogRe.SubexpIndex("error_reason")],
		TargetPortList:         values[parseAlbLogRe.SubexpIndex("target_port_list")],
		TargetStatusCodeList:   values[parseAlbLogRe.SubexpIndex("target_status_code_list")],
		Classification:         values[parseAlbLogRe.SubexpIndex("classification")],
		ClassificationReason:   values[parseAlbLogRe.SubexpIndex("classification_reason")],
	}
	if values[parseAlbLogRe.SubexpIndex("received_bytes")] != "-" {
		receiveBytes, err := strconv.Atoi(values[parseAlbLogRe.SubexpIndex("received_bytes")])
		if err != nil {
			return nil, err
		}
		record.ReceivedBytes = receiveBytes
	}
	if values[parseAlbLogRe.SubexpIndex("sent_bytes")] != "-" {
		sentBytes, err := strconv.Atoi(values[parseAlbLogRe.SubexpIndex("sent_bytes")])
		if err != nil {
			return nil, err
		}
		record.SentBytes = sentBytes
	}
	if values[parseAlbLogRe.SubexpIndex("matched_rule_priority")] != "-" {
		matchedRulePriority, err := strconv.Atoi(values[parseAlbLogRe.SubexpIndex("matched_rule_priority")])
		if err != nil {
			return nil, err
		}
		record.MatchedRulePriority = matchedRulePriority
	}
	return record, nil
}

module github.com/reproio/aws-lambda-functions/functions/send-alb-metrics-to-datadog

go 1.23.0

toolchain go1.24.5

require (
	github.com/DataDog/datadog-api-client-go/v2 v2.18.0
	github.com/aws/aws-lambda-go v1.29.0
	github.com/aws/aws-sdk-go v1.40.2
	github.com/pkg/errors v0.9.1
	golang.org/x/exp v0.0.0-20220516143420-24438e51023a
	golang.org/x/sync v0.16.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)

require (
	github.com/DataDog/zstd v1.5.2 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/oauth2 v0.10.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

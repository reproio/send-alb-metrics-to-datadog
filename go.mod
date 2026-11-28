module github.com/reproio/aws-lambda-functions/functions/send-alb-metrics-to-datadog

go 1.26.2

require (
	github.com/DataDog/datadog-api-client-go/v2 v2.59.0
	github.com/aws/aws-lambda-go v1.54.0
	github.com/aws/aws-sdk-go-v2/config v1.33.23
	github.com/aws/aws-sdk-go-v2/service/s3 v1.116.2
	github.com/pkg/errors v0.9.1
	golang.org/x/sync v0.20.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/DataDog/zstd v1.5.2 // indirect
	github.com/aws/aws-sdk-go-v2 v1.48.2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.29 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.21.5 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.20.11 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.27 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.11.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.20.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.14.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.39.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.44.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.53.2 // indirect
	github.com/aws/smithy-go v1.28.11 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.5.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/oauth2 v0.10.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

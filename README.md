# send-alb-metrics-to-datadog

## How to use

You need to create `config.yaml` and `Dockerfile` .

Example `config.yaml` :

```yaml
request_count_metrics_name: foo.alb.request_count
target_processing_time_metrics_name: foo.alb.target_processing_time
target_paths:
  - /api/v1/foo
  - /api/v1/bar
  - /api/v1/hoge
path_transforming_rules:
  - prefix: /api/v1/bar
    transformed: /api/v1/$id
  - prefix: /api/v1/hoge
    suffix: /start
    transformed: /api/v1/hoge/$id/start
```

Example `Dockerfile` :

```dockerfile
FROM ghcr.io/reproio/send-alb-metrics-to-datadog:latest AS base

FROM public.ecr.aws/lambda/provided:al2
COPY --from=base /main /main
COPY config.yaml /config.yaml
ENTRYPOINT [ "/main" ]
```

## How to development

You can execute send-alb-metrics-to-datadog as executable binary if you avoid to download log file from s3 on each execution.

```
% make build
% LOCAL_INVOKE_GZ_PATH=/path/to/alb.log.gz DD_API_KEY=xxxxxxxxxx ./main
```

In this case, you need to specify environment variable:

- LOCAL_INVOKE_GZ_PATH: Path to gz log file.
- DD_API_KEY: Datadog API key.








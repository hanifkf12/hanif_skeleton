# SigNoz Logging Integration with Zap

This application uses the official SigNoz zap_otlp library to send logs directly to SigNoz with proper trace correlation. This document explains how the integration works and how to troubleshoot common issues.

## How It Works

1. The application uses the `github.com/SigNoz/zap_otlp` library to send logs directly to SigNoz via the OpenTelemetry Protocol (OTLP).

2. Logs are automatically correlated with traces when you use the `WithTrace(ctx)` method on logger fields.

3. The integration uses gRPC to send logs to the SigNoz OpenTelemetry collector on port 4317.

## Key Components

1. **zap_otlp_encoder**: Converts Zap log entries to the OpenTelemetry log format that SigNoz understands.

2. **zap_otlp_sync**: Provides the transport mechanism to send logs to SigNoz via OTLP.

3. **zap_otlp.SpanCtx**: Extracts trace context from the Go context and adds it to logs for correlation.

## Usage Example

```go
ctx, span := telemetry.StartSpan(ctx, "my-operation")
defer span.End()

logFields := logger.NewFields("EventName").WithTrace(ctx)
logger.Info("This log will be correlated with the span", logFields)
```

## Troubleshooting

If logs are not appearing in SigNoz or not correlated with traces:

1. **Check OpenTelemetry collector connection**:
   - Verify that your application can connect to the OpenTelemetry collector on `localhost:4317`
   - Check if the collector is running with `docker ps | grep otel`

2. **Verify trace context**:
   - Make sure you're calling `WithTrace(ctx)` and passing a context that has a valid span
   - Ensure spans are being properly created with `telemetry.StartSpan`

3. **Check SigNoz configuration**:
   - Verify that SigNoz is configured to receive logs via OTLP
   - Check the SigNoz collector logs for any errors

## References

- [SigNoz zap_otlp GitHub Repository](https://github.com/SigNoz/zap_otlp)
- [SigNoz Documentation](https://signoz.io/docs/logs-management/send-logs/zap-to-signoz/)

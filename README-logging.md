# Log Collection with OpenTelemetry and SignOz

This document explains how to set up log collection with OpenTelemetry and SignOz for this application.

## Overview

This solution uses:

1. Zap logger to write structured JSON logs to a file
2. OpenTelemetry Collector to collect and process these logs
3. SignOz to store and visualize the logs with trace correlation

## Setup Instructions

### 1. Start the OpenTelemetry Collector

The OpenTelemetry Collector is set up as a sidecar container that collects logs from your application:

```bash
docker-compose -f docker-compose-collector.yaml up -d
```

### 2. Configure SignOz

Make sure SignOz is properly set up to receive data from the OpenTelemetry Collector. Typically, the collector should be connected to the same network as SignOz.

### 3. Application Configuration

The application is already configured to write logs to `/var/log/app.log` in a format that the OpenTelemetry Collector can parse.

## How it Works

1. **Log Generation**: The application uses Zap to generate structured JSON logs with trace context.

2. **Log Collection**: The OpenTelemetry Collector reads these logs from the file.

3. **Log Processing**: The collector uses operators to parse the JSON and extract trace information.

4. **Log Export**: The collector sends the processed logs to SignOz.

## Troubleshooting

If logs are not appearing in SignOz:

1. Check that the log file exists and has the correct permissions:
   ```bash
   ls -la /var/log/app.log
   ```

2. Verify the collector is running:
   ```bash
   docker ps | grep otel-collector
   ```

3. Check the collector logs for any errors:
   ```bash
   docker logs otel-collector
   ```

4. Ensure your SignOz instance is properly configured to receive logs.

## References

- [OpenTelemetry Collector Documentation](https://opentelemetry.io/docs/collector/)
- [SignOz Documentation](https://signoz.io/docs/)
- [OpenTelemetry Logs in Go](https://signoz.io/blog/opentelemetry-logs/)

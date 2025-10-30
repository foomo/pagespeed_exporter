# PageSpeed Exporter Helm Chart

## Overview

This Helm chart deploys the [PageSpeed Exporter](https://github.com/foomo/pagespeed_exporter) for Prometheus monitoring. The exporter collects Google PageSpeed Insights metrics using the Lighthouse auditing tool to monitor website performance.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- A Kubernetes secret named `pagespeed-configuration-secret` containing the Google PageSpeed API key
- Prometheus for metrics collection (optional, but recommended)

## Installation

### 1. Create the API Key Secret

First, create a secret containing your Google PageSpeed API key:

```bash
kubectl create secret generic pagespeed-configuration-secret \
  --from-literal=PAGESPEED_API_KEY=your-api-key-here
```

**Note:** An API key is only required if you're monitoring more than 2 targets per second. You can obtain one from the [Google PageSpeed Insights API](https://developers.google.com/speed/docs/insights/v5/get-started).

### 2. Install the Chart

#### Basic Installation

```bash
helm install pagespeed-exporter ./charts/pagespeed-exporter
```

#### Installation with Custom Targets (Simple Method)

```bash
helm install pagespeed-exporter ./charts/pagespeed-exporter \
  --set 'config.targets={https://www.example.com,https://www.google.com}' \
  --set config.parallel=true \
  --set config.cacheTTL="30m"
```

#### Installation with Custom Values File

Create a `custom-values.yaml` file:

```yaml
# Simple configuration method (recommended)
config:
  targets:
    - https://www.yoursite.com
    - https://www.anothersite.com
    - https://www.thirdsite.com
  parallel: true
  categories:
    - performance
    - seo
  cacheTTL: "120m"  # Cache for 2 hours (or set to null to disable)

resources:
  requests:
    memory: "64Mi"
    cpu: "50m"
  limits:
    memory: "128Mi"
    cpu: "100m"
```

Then install:

```bash
helm install pagespeed-exporter ./charts/pagespeed-exporter -f custom-values.yaml
```

## Configuration

The following table lists the configurable parameters and their default values:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas (fixed to 1 for simplicity) | `1` |
| `image.repository` | Image repository | `foomo/pagespeed_exporter` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `image.tag` | Image tag | `latest` |
| `imagePullSecrets` | Image pull secrets | `[]` |
| `nameOverride` | Override the chart name | `""` |
| `fullnameOverride` | Override the full name | `""` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.annotations` | Service account annotations | `{}` |
| `serviceAccount.name` | Service account name | `""` |
| `podAnnotations` | Pod annotations (includes Prometheus scrape config) | See values.yaml |
| `podSecurityContext` | Pod security context | `{}` |
| `securityContext` | Container security context | `{}` |
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | `9271` |
| `service.targetPort` | Container port | `9271` |
| `service.annotations` | Service annotations (includes Prometheus scrape config) | See values.yaml |
| `resources` | CPU/Memory resource requests/limits | `{}` |
| `nodeSelector` | Node labels for pod assignment | `{}` |
| `tolerations` | Tolerations for pod assignment | `[]` |
| `affinity` | Affinity for pod assignment | `{}` |
| `secretName` | Name of the secret containing API key | `pagespeed-configuration-secret` |
| `config.targets` | List of URLs to monitor (simple method) | `[]` |
| `config.parallel` | Enable parallel execution | `false` |
| `config.categories` | Categories to check (empty = all) | `[]` |
| `config.cacheTTL` | Cache TTL for API results (e.g., "60s", "5m") | `"60m"` |
| `args` | Raw command-line arguments (advanced) | `[]` |
| `extraEnvVars` | Additional environment variables | `[]` |

## Command-Line Arguments

The PageSpeed Exporter supports the following command-line arguments. Configure them through the `args` array in values.yaml:

### Available Arguments

| Argument | Description | Default | Example |
|----------|-------------|---------|---------|
| `-targets` | Comma-separated list of targets to measure | None | `"-targets=https://example.com,https://google.com"` |
| `-t` | Multi-value target array (can be used multiple times) | None | `"-t=https://example.com"` |
| `-api-key` | Google API key (alternatively use env var) | None | `"-api-key=your-key-here"` |
| `-categories` | Categories to check | `accessibility,best-practices,performance,pwa,seo` | `"-categories=performance,seo"` |
| `-listener` | Listener address for the exporter | `:9271` | `"-listener=:8080"` |
| `-parallel` | Enable parallel execution | `false` | `"-parallel=true"` |
| `-cache-ttl` | Cache TTL for API results | None (disabled) | `"-cache-ttl=5m"` |
| `-pushGatewayUrl` | Push Gateway URL | None | `"-pushGatewayUrl=http://pushgateway:9091"` |
| `-pushGatewayJob` | Push Gateway job name | `pagespeed_exporter` | `"-pushGatewayJob=my-job"` |

### Configuration Examples

#### Simple Configuration (Recommended)

Use the `config` section for common use cases:

```yaml
config:
  targets:
    - https://www.example.com
    - https://www.mysite.com
  categories:
    - performance
    - seo
    - accessibility
  parallel: true
  cacheTTL: "60m"  # Cache results for 60 minutes (default)
```

#### Advanced: Using Raw Args

For advanced use cases, use the `args` array directly:

```yaml
args:
  - "-targets=https://www.example.com,https://www.mysite.com"
  - "-categories=performance,seo,accessibility"
  - "-parallel=true"
```

#### Multi-Value Target Flags

```yaml
args:
  - "-t=https://www.example.com"
  - "-t=https://www.google.com"
  - "-t=https://www.github.com"
```

#### Push to Prometheus Push Gateway

```yaml
args:
  - "-targets=https://www.example.com"
  - "-pushGatewayUrl=http://prometheus-pushgateway:9091"
  - "-pushGatewayJob=pagespeed-monitoring"
```

## Prometheus Integration

The chart includes Prometheus annotations on both the pods and service for automatic discovery:

```yaml
prometheus.io/scrape: "true"
prometheus.io/port: "9271"
prometheus.io/path: "/metrics"
```


## Metrics

The exporter provides the following types of metrics:

- **Performance Score**: Overall performance score (0-100)
- **Accessibility Score**: Website accessibility rating
- **Best Practices Score**: Adherence to web best practices
- **SEO Score**: Search engine optimization rating
- **PWA Score**: Progressive Web App compliance

All metrics are prefixed with `pagespeed_` and include labels for the target URL and strategy (mobile/desktop).

## Example Metrics Output

```
# HELP pagespeed_lighthouse_audit_score Lighthouse audit score (0-100)
# TYPE pagespeed_lighthouse_audit_score gauge
pagespeed_lighthouse_audit_score{audit="performance",url="https://www.example.com",strategy="mobile"} 95
pagespeed_lighthouse_audit_score{audit="accessibility",url="https://www.example.com",strategy="mobile"} 98
pagespeed_lighthouse_audit_score{audit="seo",url="https://www.example.com",strategy="mobile"} 100
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -l app.kubernetes.io/name=pagespeed-exporter
```

### View Logs

```bash
kubectl logs -l app.kubernetes.io/name=pagespeed-exporter
```

### Test Metrics Endpoint

```bash
kubectl port-forward svc/pagespeed-exporter 9271:9271
curl http://localhost:9271/metrics
```

### Common Issues

1. **API Key Missing**: Ensure the secret `pagespeed-configuration-secret` exists with key `PAGESPEED_API_KEY`
2. **Rate Limiting**: Without an API key, you're limited to 2 requests per second
3. **Target Unreachable**: Verify that the targets specified are accessible from your cluster

## Uninstallation

```bash
helm uninstall pagespeed-exporter
```

## Support

For issues related to the Helm chart, please check the repository documentation.
For issues with the exporter itself, refer to the [upstream repository](https://github.com/foomo/pagespeed_exporter).
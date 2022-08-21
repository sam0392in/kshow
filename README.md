# Kshow

Kshow is a CLI tool for quick access to k8s objects with custom fields.

## Owner
Samarth Kanungo

## Installation
- Download the binary from S3.
- Rename it to kshow
```
mv kshow-0.0.1-amd64-darwin kshow
```
- Put it under the Path.
```
mv kshow /usr/local/bin/kshow
```

## Deployments

### List Deployments

namespace is optional. Default is all namespace
```
kshow deployment list --namespace kube-system
```

### List Deployments with Tolerations
```
kshow deployment list --show-tolerations
```

## Metrics

### Get Metrics
```
kshow resource-stats --namespace kube-system
```

### Get Container Metrics
```
kshow resource-stats --namespace kube-system --containers
```
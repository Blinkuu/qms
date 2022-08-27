load('ext://helm_resource', 'helm_resource', 'helm_repo')

# Add Helm repos
helm_repo('bitnami-charts', 'https://charts.bitnami.com/bitnami')

# Compile QMS
local_resource(
  'go-compile-qms',
  cmd='make build',
  deps=['./Makefile', './cmd', './internal', './pkg'],
  labels=["local-job"],
)

# Build Docker images
docker_build(
    'qms',
    '.',
    dockerfile='cmd/qms/Dockerfile',
)

# Setup resources
helm_resource(
    'redis',
    'bitnami-charts/redis',
    port_forwards=['6379'],
    flags=["--set", "architecture=standalone"],
    labels=["db"]
)

k8s_yaml(kustomize('./deployments/kubernetes/envs/dev'))

k8s_resource(
    'grafana-agent-logs',
    labels=['operations'],
)

k8s_resource(
    'grafana-agent-metrics',
    labels=['operations'],
)

k8s_resource(
    'grafana-agent-traces',
    labels=['operations'],
)

k8s_resource(
    'loki',
    labels=['observability'],
)

k8s_resource(
    'grafana',
    port_forwards=['3000'],
    labels=['observability'],
)

k8s_resource(
    'mimir',
    labels=['observability'],
)

k8s_resource(
    'tempo',
    port_forwards=['3200'],
    labels=['observability'],
)

k8s_resource(
    'qms',
     port_forwards=['6789'],
     links=[
        link("http://localhost:6789/metrics", "/metrics"),
     ],
     labels=["qms"],
)

load('ext://helm_resource', 'helm_resource', 'helm_repo')

# Parse args
config.define_string_list("to-run", args=True)
cfg = config.parse()
resources = [
    'go-compile-qms',
    'go-compile-sut',
    'redis',
    'grafana-agent-logs',
    'grafana-agent-metrics',
    'grafana-agent-traces',
    'grafana',
    'loki',
    'mimir',
    'tempo',
]
groups = {
  'monolith': ['qms'],
  'microservices': ['qms-proxy', 'qms-rate', 'qms-alloc'],
}
for arg in cfg.get('to-run', []):
  if arg in groups:
    resources += groups[arg]
  else:
    # also support specifying individual services instead of groups, e.g. `tilt up a b d`
    resources.append(arg)
config.set_enabled_resources(resources)

# Add Helm repos
helm_repo('bitnami-charts', 'https://charts.bitnami.com/bitnami')

# Compile
local_resource(
  'go-compile-qms',
  cmd='GOARCH=amd64 GOOS=linux make build',
  deps=['./Makefile', './cmd/qms', './internal', './pkg'],
  labels=["local-job"],
)

local_resource(
  'go-compile-sut',
  cmd='GOARCH=amd64 GOOS=linux make build-sut',
  deps=['./Makefile', './cmd/sut', './internal', './pkg'],
  labels=["local-job"],
)

# Build Docker images
docker_build(
    'qms',
    '.',
    dockerfile='cmd/qms/Dockerfile',
)

docker_build(
    'sut',
    '.',
    dockerfile='cmd/sut/Dockerfile',
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
    'grafana',
    port_forwards=['3000'],
    labels=['observability'],
)

k8s_resource(
    'loki',
    labels=['observability'],
)

k8s_resource(
    'mimir',
    labels=['observability'],
)

k8s_resource(
    'tempo',
    labels=['observability'],
)

# Microservices
k8s_resource(
    'qms-proxy',
     port_forwards=['10000:6789'],
     links=[
        link("http://localhost:10000/metrics", "/metrics"),
     ],
     labels=["qms"],
)

k8s_resource(
    'qms-rate',
     port_forwards=['10001:6789'],
     links=[
        link("http://localhost:10000/metrics", "/metrics"),
     ],
     labels=["qms"],
)

k8s_resource(
    'qms-alloc',
     port_forwards=['10002:6789'],
     links=[
        link("http://localhost:10000/metrics", "/metrics"),
     ],
     labels=["qms"],
)

# Monolith
k8s_resource(
    'qms',
     port_forwards=['10000:6789'],
     links=[
        link("http://localhost:10000/metrics", "/metrics"),
     ],
     labels=["qms"],
)

k8s_resource(
    'sut',
     port_forwards=['8080'],
     labels=["sut"],
)

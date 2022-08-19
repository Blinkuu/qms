load('ext://helm_resource', 'helm_resource', 'helm_repo')

# Add Helm repos
helm_repo('bitnami', 'https://charts.bitnami.com/bitnami')

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
    only=['./bin'],
    live_update=[sync('./bin', '/app/bin')],
)

# Create resources
helm_resource(
    'redis',
    'bitnami/redis',
    port_forwards=['6379'],
    flags=["--set", "architecture=standalone"],
    labels=["db"]
)

k8s_yaml(kustomize('./deployments/kubernetes/envs/dev'))
k8s_resource(
    'dev-qms',
     port_forwards=['6789'],
     labels=["qms"],
)

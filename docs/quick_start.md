# Deploy Lunettes with kind quickly

## Dependency

1. [kind](https://kind.sigs.k8s.io/)
2. [install Docker](https://docs.docker.com/get-docker/)
3. [kubectl](https://kubernetes.io/docs/tasks/tools/)
4. [Installing Helm](https://helm.sh/docs/intro/install/)

## Running local Kubernetes clusters

If you want to deploy Kubernetes using our kind configuration, you need to clone the project to your local machine first.

```bash
git clone https://github.com/alipay/container-observability-service lunettes && cd lunettes

kind create cluster \
  --name k8s \
  --config ./hack/kind.yaml
```

## Deploy Lunettes with helm

```bash
helm upgrade --install lunettes oci://registry-1.docker.io/lunettes/lunettes-chart --version [version]
  
```
see available [version](https://hub.docker.com/r/lunettes/lunettes-chart/tags)

## Create test pod

```bash
kubectl run nginx --image=nginx
```

## Visit

Open broswer to visit
- lunettes
  - debugpod api: [http://localhost:9096/api/v1/debugpod?name=nginx](http://localhost:9096/api/v1/debugpod?name=nginx)
  - debugslo api: [http://localhost:9096/api/v1/debugslo?result=success](http://localhost:9096/api/v1/debugslo?result=success)
- grafana: http://localhost:9097 The default username and password are admin/admin.
  - debugpod: [http://localhost:9097/d/lunettes-debugslo/lunettes-debugslo?orgId=1](http://localhost:9097/d/lunettes-debugslo/lunettes-debugslo?orgId=1)
  - debugslo: [http://localhost:9097/d/lunettes-debugslo/lunettes-debugslo?orgId=1](http://localhost:9097/d/lunettes-debugslo/lunettes-debugslo?orgId=1)
- jaejer: [http://localhost:9095/search](http://localhost:9095/search)
- kibana: [http://localhost:9092](http://localhost:9092)
- prometheus: [http://localhost:9091/graph?](http://localhost:9091/graph?)
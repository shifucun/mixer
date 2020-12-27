# Data Commons Mixer Developer Guide

## Prerequisit

* Contact DataCommons team to get data access to Cloud Bigtable and BigQuery.

* Install the following tools to develop mixer locally
  * [`Golang`](https://golang.org/doc/install)
  * [`Docker`](https://www.docker.com/products/docker-desktop)
  * [`Minikube`](https://minikube.sigs.k8s.io/docs/start/)
  * [`Skaffold`](https://skaffold.dev/docs/install/)
  * [`gcloud`](https://cloud.google.com/sdk/docs/install)
  * [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
  * [`protoc`](http://google.github.io/proto-lens/installing-protoc.html)

* Authenticate to GCP

  ```bash
  gcloud components update
  gcloud auth login
  gcloud auth application-default login
  ```

## Develop mixer locally with Docker and Kubernetes (Recommended)

Mixer and [ESP](https://cloud.google.com/endpoints/docs/grpc/running-esp-localdev) is deployed on a local Minikube cluster. To avoid using Endpoints API management and talking to GCP, local deployment uses Json API configuration, which is compiled using [API Compiler](https://github.com/googleapis/api-compiler).

### Start mixer in minikube

```bash
minikube start
eval $(minikube docker-env)
skallfold dev --port-foward
```

This exposes the local mixer service at `localhost:9090`.

To verify the server serving request:

```bash
curl http://localhost:9090/node/property-labels?dcids=Class
```

After code edit, the container images are automatically rebuilt and re-deployed to the local cluster.

### Run Tests

```bash
./scripts/run_test.sh -d
```

### Update e2e test golden files

```bash
./scripts/update_golden_staging.sh -d
```

## Develop mixer locally as a Go server

**NOTE** This can only develop and test the gRPC server but not the transcoding done by [ESP](https://cloud.google.com/endpoints/docs/grpc/running-esp-localdev).

### Start mixer as a gRPC server

Run the following code to generate Go proto files.

```bash
go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.23.0
go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@v0.0.0-20200824180931-410880dd7d91
protoc \
  --proto_path=proto \
  --go_out=pkg \
  --go-grpc_out=pkg \
  --go-grpc_opt=requireUnimplementedServers=false \
  proto/*.proto
```

Run the following code to start mixer gRPC server

```bash
# cd into repo root directory

go run main.go \
    --bq_dataset=$(head -1 deploy/base/bigquery.version) \
    --bt_table=$(head -1 deployment/bigtable.version) \
    --bt_project=google.com:datcom-store-dev \
    --bt_instance=prophet-cache \
    --project_id=datcom-mixer-staging

# In a new shell
cd examples && go run main.go
```

### Run Tests (Go)

```bash
./scripts/run_test.sh
```

### Update e2e test golden files (Go)

```bash
./scripts/update_golden_staging.sh
```

## Update prod golden files

Run the following commands to update prod golden files from staging golden files

```bash
./update-golden-prod.sh
```

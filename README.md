# mlsolid
![mlflow-banner-4](https://github.com/user-attachments/assets/06baeb29-9c30-4efa-af9c-0a485656a520)
[![Docker](https://github.com/zeddo123/mlsolid/actions/workflows/docker-publish.yaml/badge.svg)](https://github.com/zeddo123/mlsolid/actions/workflows/docker-publish.yaml)
[![Build](https://github.com/zeddo123/mlsolid/actions/workflows/build.yaml/badge.svg)](https://github.com/zeddo123/mlsolid/actions/workflows/build.yaml)
[![golangci-lint](https://github.com/zeddo123/mlsolid/actions/workflows/lint.yaml/badge.svg)](https://github.com/zeddo123/mlsolid/actions/workflows/lint.yaml)
[![Buf CI](https://github.com/zeddo123/mlsolid/actions/workflows/buf-ci.yaml/badge.svg)](https://github.com/zeddo123/mlsolid/actions/workflows/buf-ci.yaml)
[![.github/workflows/buf-lint.yaml](https://github.com/zeddo123/mlsolid/actions/workflows/buf-lint.yaml/badge.svg)](https://github.com/zeddo123/mlsolid/actions/workflows/buf-lint.yaml)

mlsolid is a solid alternative to mlflow written in Go with Redis as its db backend, and s3 as its artifact storage.
This project is split in multiple parts. `mlsolid` the server (this repo), [`mlsolidpy`](https://github.com/zeddo123/mlsolidpy) the python client,
and a [frontend dashboard](https://github.com/zeddo123/mlsolid-front).

mlsolid address my issue with mlflow by being:

0. fast
1. production focused, and easy to deploy
2. dumb client (the client should only send experiments and artifacts)
3. better documentation by being not convoluted and complicated to oblivion (i.e no 1000+ options with the same similar names).

As a design decision, mlsolid is solely responsible of saving artifacts to the object store as opposed to mlflow which by default does not
work in "proxied artifact storage" mode (particularly hard to setup). This is done to for security measures so that S3 keys are shared as little as possible, as well as
to make mlsolid require little configuration from the client side (aka data science side).

Under the hood, clients interact with mlsolid through a `gRPC` endpoint. This choice makes it possible to use different languages (other than Python) to track your experiments
and/or download your models.
Already generated gRPC SDKs for multiple languages are available to download using `buf.build` [here](https://buf.build/zeddo123/mlsolid/sdks)

## Features
* Experiment tracking with metrics and artifacts
* Model registry

## Configuration
Configuration happens through a `yaml` file located either at `./mlsolid.yaml` of the binary or at `/etc/mlsolid/mlsolid.yaml`.
```yaml
prod: true
grpc_port: 5000

redis_addr: redis:6379
redis_password: ""
redis_db: 0

s3_endpoint: ""
s3_key: ""
s3_secret: ""
s3_bucket: ""
s3_region: ""
```

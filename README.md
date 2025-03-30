# mlsolid

[![Build](https://github.com/zeddo123/mlsolid/actions/workflows/build.yaml/badge.svg)](https://github.com/zeddo123/mlsolid/actions/workflows/build.yaml)
[![golangci-lint](https://github.com/zeddo123/mlsolid/actions/workflows/lint.yaml/badge.svg)](https://github.com/zeddo123/mlsolid/actions/workflows/lint.yaml)
[![Buf CI](https://github.com/zeddo123/mlsolid/actions/workflows/buf-ci.yaml/badge.svg)](https://github.com/zeddo123/mlsolid/actions/workflows/buf-ci.yaml)
[![.github/workflows/buf-lint.yaml](https://github.com/zeddo123/mlsolid/actions/workflows/buf-lint.yaml/badge.svg)](https://github.com/zeddo123/mlsolid/actions/workflows/buf-lint.yaml)

mlsolid is an mlflow alternative but solid written in Go with Redis as its db backend, and s3 as its artifact storage.

mlsolid address my issue with mlflow by being:

0. fast
1. production focused, and easy to deploy
2. dumb client (the client should only send experiments and artifacts)
3. better documentation by being not convoluted and complicated to oblivion (i.e no 1000+ options with the same similar names).

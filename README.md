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
* Model registry with versioning

### Python client example
Here are some basic example using our Python client to track your experiments:
```Python
from mlsolidpy.mlsolid import Mlsolid

client = Mlsolid('localhost:5000')

print('Experiments', client.experiments)

print('Run ', client.run("urbane-wagon"))

with client.start_run('my_experiment') as run:
    run.log({'checkpoint': "path/to/checkpoint"})
    run.log({'batch-size': 23})

    run.log({'mae': 0.2333, 'loss': 100.0})
    run.log({'mae': 0.2000, 'loss': 90})
    run.log({'mae': 0.1134, 'loss': 10})
    run.log({'metrics': [0.2000, 0.333, 0.2223]})
```

And here is a basic example on how to use `mlsolid` to push models and artifacts:
```Python

from mlsolidpy.mlsolid import Mlsolid

client = Mlsolid('localhost:5000')

# create a new model registry to version your model
created = client.create_model_registry('test_registry_1')

run_id = None

with client.start_run('my_experiment') as run:
    run_id = run.run_id

    # you can attach a plain text file artifact (logs, etc) to your run:
    run.add_plaintext_artifact('./tests/data/plain_text_file.txt')

    # You can add a model artifact to your run:
    run.add_model('./tests/data/mobile_sam.pt')

# After adding a uploading your model artifact linked to your run, so can
# attached it to your model registry with the name of the registry (test_registry_1 in our example)
# and the run_id and name of the artifact. You can also add a list of tags, to allow
# easier access to your model.
added = client.add_model('test_registry_1', run_id, 'mobile_sam.pt', ['latest'])

# You can easily download your model,
# by simply providing the model registry name and a valid tag.
client.tagged_model('test_registry_1', 'latest')

# You can also access any of your artifacts by providing the run_id and their name
client.artifact(run_id, 'plain_text_file.txt')
```

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

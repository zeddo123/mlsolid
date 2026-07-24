# Contributing to mlsolid

Thanks for your interest in mlsolid! This doc covers everything you need to build, test, and submit a change.

## Project layout

* `main.go`, `solid/` — the server: gRPC service (`solid/grpcservice`), REST API (`solid/api/v1`), controllers, store (Redis), S3 client, OAuth, benchmarking engine (`solid/bengine`).
* `proto/` — gRPC/protobuf service definitions, managed with [buf](https://buf.build).
* `cmd/populate` — seeds a running server with fake experiments/runs/metrics for exercising the dashboard locally.
* `cmd/stress` — a concurrent load-testing tool for the gRPC API.
* `mlsolidpy` and `solidash` are separate repos (Python client and web dashboard) — issues specific to those belong there, not here.

## Prerequisites

* Go 1.25+
* [buf CLI](https://buf.build/docs/installation) — only needed if you're changing `proto/` and regenerating gRPC/protobuf code.
* Docker and Docker Compose — for running the server locally against Redis, and for anything touching the benchmarking engine.
* An S3-compatible bucket reachable from wherever you run the server (AWS S3, MinIO, etc.) — required to exercise artifact/model upload-download paths.

## Running locally

```bash
touch mlsolid.yaml   # then fill in config keys, see README's Configuration section
docker compose up --build
```

This starts Redis and the server together. You'll still need to fill in `s3_*` config keys pointing at a real bucket. See the [Configuration section of the README](./README.md#️-configuration) for the full set of options. `mlsolid.yaml` is gitignored — never commit real credentials in it.

## Building and testing

```bash
go build -v ./...
go test -tags=integrationtests -v --cover ./...
```

CI runs the same build and test commands, plus `golangci-lint` (config in `.golangci.yml`) and `buf lint`/`buf breaking` for anything under `proto/`. Run `golangci-lint run` locally before opening a PR to avoid a slow feedback loop:

```bash
golangci-lint run
```

## Changing the gRPC/protobuf API

If your change touches `proto/`, regenerate the Go bindings with:

```bash
buf generate
```

Generated code lives under `solid/gen` and should not be hand-edited. `buf breaking` runs in CI against the previous commit — if you're intentionally making a breaking change to the API, call that out in your PR description.

## Submitting a change

1. Open an issue first for anything beyond a small fix (bug fix, doc fix, small self-contained feature) so we can agree on the approach before you invest time in it.
2. Keep PRs focused — one logical change per PR. Unrelated cleanup should be its own PR.
3. Add or update tests for behavior you change. Table-driven tests are the prevailing style in this codebase.
4. Make sure `go build ./...`, `go test -tags=integrationtests ./...`, and `golangci-lint run` all pass locally.
5. Write a PR description that explains *why*, not just *what* — link the issue it closes.

## Good first issues

Issues labeled [`good first issue`](https://github.com/zeddo123/mlsolid/labels/good%20first%20issue) are scoped to be tractable without deep context on the rest of the system. If you're picking one up, feel free to comment on the issue before starting so two people don't duplicate work.

## Code of conduct

Be respectful and constructive in issues, PRs, and discussions. Disagreements about technical approach are fine and expected — keep them focused on the work.

## License

mlsolid is licensed under [GPLv3](./LICENSE). By contributing, you agree that your contributions will be licensed under the same terms.

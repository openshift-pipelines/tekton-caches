# tekton-caches [![build-test-publish](https://github.com/openshift-pipelines/tekton-caches/actions/workflows/latest.yaml/badge.svg)](https://github.com/openshift-pipelines/tekton-caches/actions/workflows/latest.yaml)

This is a tool to cache resources like go cache/maven or others on TektonCD
pipelines.

This tool supports uploading the cache to an OCI registry and plans to support
S3, GCS and other storage backends.

It uses the new [StepActions](https://tekton.dev/docs/pipelines/stepactions/)
feature of TektonCD Pipelines but can be as well used without it.

See the StepActions in the [tekton/](./tekton) directory.

## Example

This is an example of a build pipeline for a go application caching and reusing
the go cache. If the `go.mod` and `go.sum` are changed the cache is invalidated and
rebuilt.

### Pre-requisites

- You need a recent TektonCD pipelines installed with the StepActions feature-flags enabled.

```shell
kubectl patch configmap -n tekton-pipelines --type merge -p '{"data":{"enable-step-actions": "true"}}' feature-flags
```

- A registry to push the images to. Example: docker.io/loginname. Make sure you
  have setup tekton to be able to push/fetch from that registry, see the
  [TektonCD pipelines documentation](https://tekton.dev/docs/pipelines/auth/#configuring-authentication-for-docker)

### Usage

Create the go pipeline example from the examples directory:

```shell
kubectl create -f pipeline-go.yaml
```

Start it with the tkn cli (change the value as needed):

```shell
tkn pipeline start pipeline-go --param repo_url=https://github.com/vdemeester/go-helloworld-app --param revision=main --param registry=docker.io/username -w name=source,emptyDir=
```

or with a PipelineRun yaml object:

```shell
```yaml
kind: PipelineRun
metadata:
  name: build-go-application-with-caching-run
spec:
  pipelineRef:
    name: pipeline-go
  params:
    - name: repo_url
      value: https://github.com/vdemeester/go-helloworld-app
    - name: revision
      value: main
    - name: registry
      value: docker.io/username
  workspaces:
    - name: source
      emptyDir: {}
```

## License

[Apache License 2.0](./LICENSE)

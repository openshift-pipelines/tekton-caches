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
tkn pipeline start pipeline-go --param repo_url=https://github.com/vdemeester/go-helloworld-app --param revision=main --param registry=docker.io/username -w name=source,emptyDir= --use-param-defaults
```

or with a PipelineRun yaml object:

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
      emptyDir: { }
```

- you can as well redefine the `buildCommand` which by default do a `go build
  -v ./` with the `buildCommand` parameter, for example if you want instead to
  run the tests on a repo with caching:

```shell
tkn pipeline start pipeline-go --param repo_url=https://github.com/chmouel/gosmee \ 
  --param revision=main --param registry=docker.io/username \
  --param=buildCommand="make test" -w name=source,emptyDir= --use-param-defaults --showlog
```

- You can as well force the upload of the cache with param `FORCE_CACHE_UPLOAD=true` (default: false)
- You can provide your own image with the param `image` (default to the latest docker.io `golang` image)
- You can provide your own patterns for the hash to computer with the `cachePatterns` array parameter (default to
  `go.mod,go.sum`)

## Using with Google Storage as a backend

In order to use the `StepAction` with GCS, the parameter `googleCredentialsPath` needs to be specified. It should point
to the google service account json file — which usually comes from a secret.

For example, let's assume a secret name `gcs-secret` is populated with the content of the google service account, key
`gcs-sa.json` (a json file, be it with or without support for Google Workload Identity). One could use a `workspace` or
a `volume` to mount that secret somewhere and set the path to the `StepAction`.

```yaml
apiVersion: tekton.dev/v1
kind: TaskRun
metadata:
  generateName: my-taskrun-
spec:
  params:
    - name: serviceAccountName
      value: gcs-sa.json
  taskSpec:
    params:
      - name: serviceAccountName
        type: string
        default: ""
    workspaces:
      - name: source
      - name: google-credentials
        optional: true
      - name: bound-sa-token
        mountPath: /var/run/secrets/openshift/serviceaccount
        optional: true
    steps:
      - # […] git clone, …
      - name: cache-fetch
        ref:
          name: cache-fetch
          # or using http resolver with https://raw.githubusercontent.com/openshift-pipelines/tekton-caches/main/tekton/cache-fetch.yaml
        params:
          - name: PATTERNS
            value: [ "go.mod", "go.sum" ]
          - name: SOURCE
            value: gs://my-bucket/some/folder
          - name: CACHE_PATH
            value: $(workspace.source.path)/cache
          - name: WORKING_DIR
            value: $(worksoaces.source.path)/repo
          - name: GOOGLE_APPLICATION_CREDENTIALS
            value: $(workspace.google-credentials.path)/$(params.serviceAccountName)
      - # […] something else like go build
      - # […] and then same thing with cache-upload
  workspaces:
    - name: source
      emptyDir: { }
    - name: google-credentials
      secret:
        secretName: gcs-secret
    - name: bound-sa-token
      projected:
        sources:
          - serviceAccountToken:
              audience: openshift
              expirationSeconds: 3600
              path: token
        defaultMode: 420
```

`bound-sa-token` workspace isn't required if Workload Identity federation isn't setup. Here we assumed an OIDC is
configured in OpenShift.

## Param Names

Here is the list of params being used in cache step actions.

### common

These parameters are supported in both `cache-fetch` and `cache-upload` step-action

| Name                             | Type     | Optional | Description                                                                                                                                                                                                                   | Default Value |
|----------------------------------|----------|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------|
| `PATTERNS`                       | `array`  | NO       | Regular expression to select files to include to compute the hash. For example, in the case of a Go project, you can use `go.mod` for this, so the value would be "**/go.sum" (to work with possible sub go modules as well). | None          |
| `CACHE_PATH`                     | `string` | NO       | Path where to extract the cache content. <br> It can refer any folder, backed by a workspace or a volume, or nothing.                                                                                                         | None          |
| `WORKING_DIR`                    | `string` | NO       | The working dir from where the files patterns needs to be taken                                                                                                                                                               | None          |
| `INSECURE`                       | `string` | YES      | Whether to use insecure mode for fetching the cache                                                                                                                                                                           | `false`       |
| `GOOGLE_APPLICATION_CREDENTIALS` | `string` | YES      | The path where to find the google credentials. If left empty, it is ignored.                                                                                                                                                  | `""`          |
| `AWS_CONFIG_FILE`                | `string` | YES      | The path to the aws config file. If left empty, it is ignored.                                                                                                                                                                | `""`          |
| `AWS_SHARED_CREDENTIALS_FILE`    | `string` | YES      | The path to find the aws credentials file. If left empty, it is ignored.                                                                                                                                                      | `""`          |
| `BLOB_QUERY_PARAMS`              | `string` | YES      | Blob Query Params to support configure s3, gcs and azure. This is optional unless some additional features of storage providers are required like s3 acceleration, fips, pathstyle,etc                                        | `""`          |

### cache-fetch

| Name     | Type     | Optional | Description                                                                                                                                                                                                                                                                                                                                                               | Default Value |
|----------|----------|----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------|
| `SOURCE` | `string` | NO       | The source from where the cache should be fetched. It's a URI with the scheme defining the "provider". In addition, one can add a {{hash}} variable to use the computed hash in the reference (oci image tags, path in s3, …)<br> **Currently supported:** <ul> <li> oci:// (e.g. oci://quay.io/vdemeester/go-cache:{{hash}}<li> s3:// (e.g. s3:// <li> gs:// (e.g. gs:// | None          |

### cache-upload

| Name     | Type     | Optional | Description                                                                                                                                                                                                                                                                                                                                                               | Default Value |
|----------|----------|----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------|
| `TARGET` | `string` | NO       | The target from where the cache should be uploaded. It's a URI with the scheme defining the "provider". In addition, one can add a {{hash}} variable to use the computed hash in the reference (oci image tags, path in s3, …)<br> **Currently supported:** <ul> <li> oci:// (e.g. oci://quay.io/vdemeester/go-cache:{{hash}}<li> s3:// (e.g. s3://<li> gs:// (e.g. gs:// | None          |

## License

[Apache License 2.0](./LICENSE)

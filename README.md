# tekton-caches [![build-test-publish](https://github.com/openshift-pipelines/tekton-caches/actions/workflows/latest.yaml/badge.svg)](https://github.com/openshift-pipelines/tekton-caches/actions/workflows/latest.yaml)

Tools (and Task/StepAction) to managing within Tekton

```bash
# With OCI
$ cache fetch --hasfiles '**/go.sum' --target oci://quay.io/vdemeest/cache/go-cache:{{hash}} --folder /workspaces/go-cache
$ cache upload --hashfiles '**/go.sum' --target oci://quay.io/vdemeest/cache/go-cache:{{hash}} --folder /workspaces/go-cache
# With s3
$ cache fetch --hashfiles '**/go.sum' --target s3://my-bucket/path/to/my/cache --folder /workspaces/go-cache
$ cache fetch --hashfiles '**/go.sum' --target s3://my-bucket/path/to/my/cache --folder /workspaces/go-cache
# With gcs
$ cache fetch --hashfiles '**/go.sum' --target gcs://my-bucket/path/to/my/cache --folder /workspaces/go-cache
$ cache fetch --hashfiles '**/go.sum' --target gcs://my-bucket/path/to/my/cache --folder /workspaces/go-cache
```

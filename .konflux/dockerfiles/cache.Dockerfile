ARG GO_BUILDER=brew.registry.redhat.io/rh-osbs/openshift-golang-builder:v1.23
ARG RUNTIME=registry.access.redhat.com/ubi9/ubi-minimal@sha256:db8bf20cac0f250659f1ffbba929639dcd7e09863c20c99c5cb8f3ce8ded0497

FROM $GO_BUILDER AS builder

WORKDIR /go/src/github.com/openshift-pipelines/tekton-caches
COPY . .

RUN go build -v -o /tmp/cache  ./cmd/cache

FROM $RUNTIME
ARG VERSION=tekton-caches-main

COPY --from=builder /tmp/cache /ko-app/cache


LABEL \
      com.redhat.component="openshift-pipelines-tekton-caches" \
      name="openshift-pipelines/pipelines-tekton-caches-rhel8" \
      version=$VERSION \
      summary="Red Hat OpenShift Pipelines Tekton Caches" \
      maintainer="pipelines-extcomm@redhat.com" \
      description="Red Hat OpenShift Pipelines Tekton Caches" \
      io.k8s.display-name="Red Hat OpenShift Pipelines Tekton Caches" \
      io.k8s.description="Red Hat OpenShift Pipelines Tekton Caches" \
      io.openshift.tags="pipelines,tekton,openshift,tekton-caches"

RUN microdnf install -y shadow-utils && \
    groupadd -r -g 65532 nonroot && useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532

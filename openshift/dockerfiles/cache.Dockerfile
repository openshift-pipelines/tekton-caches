ARG GO_BUILDER=brew.registry.redhat.io/rh-osbs/openshift-golang-builder:v1.22
ARG RUNTIME=registry.access.redhat.com/ubi9/ubi-minimal:latest@sha256:c0e70387664f30cd9cf2795b547e4a9a51002c44a4a86aa9335ab030134bf392

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


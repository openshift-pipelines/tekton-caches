ARG GO_BUILDER=registry.redhat.io/ubi9/go-toolset:1.24
ARG RUNTIME=registry.redhat.io/ubi9/ubi-minimal@sha256:61d5ad475048c2e655cd46d0a55dfeaec182cc3faa6348cb85989a7c9e196483
FROM $GO_BUILDER AS builder

WORKDIR /go/src/github.com/openshift-pipelines/tekton-caches
COPY . .

ENV GOEXPERIMENT=strictfipsruntime
RUN git config --global --add safe.directory . && \
    go build -tags $GOEXPERIMENT  -v -o /tmp/cache  ./cmd/cache

FROM $RUNTIME
ARG VERSION=tekton-caches-0.3

COPY --from=builder /tmp/cache /ko-app/cache


LABEL \
      com.redhat.component="openshift-pipelines-tekton-caches" \
      name="openshift-pipelines/pipelines-tekton-caches-rhel9" \
      version=$VERSION \
      summary="Red Hat OpenShift Pipelines Tekton Caches" \
      maintainer="pipelines-extcomm@redhat.com" \
      description="Red Hat OpenShift Pipelines Tekton Caches" \
      io.k8s.display-name="Red Hat OpenShift Pipelines Tekton Caches" \
      io.k8s.description="Red Hat OpenShift Pipelines Tekton Caches" \
      io.openshift.tags="pipelines,tekton,openshift,tekton-caches"

RUN groupadd -r -g 65532 nonroot && useradd --no-log-init -rm -u 65532 -g nonroot nonroot
USER 65532

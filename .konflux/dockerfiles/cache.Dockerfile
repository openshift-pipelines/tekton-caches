ARG GO_BUILDER=registry.redhat.io/ubi9/go-toolset:9.7
ARG RUNTIME=registry.redhat.io/ubi9/ubi-minimal@sha256:c7d44146f826037f6873d99da479299b889473492d3c1ab8af86f08af04ec8a0
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
      io.openshift.tags="pipelines,tekton,openshift,tekton-caches" \
      cpe="cpe:/a:redhat:openshift_pipelines:1.20::el9"

RUN groupadd -r -g 65532 nonroot && useradd --no-log-init -rm -u 65532 -g nonroot nonroot
USER 65532

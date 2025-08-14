#!/usr/bin/env bash
# This setups kind for running e2e tests on it.
# 
# This targets both local setup (with docker or podman) and github workflows.

set -euf
cd $(dirname $(readlink -f ${0}))

export KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-kind}
export KUBECONFIG=${HOME}/.kube/config.${KIND_CLUSTER_NAME}
export DOMAIN_NAME=caches-127-0-0-1.nip.io
export DOCKER=${DOCKER:-docker}

TMPD=$(mktemp -d /tmp/.GITXXXX)
REG_PORT='5000'
REG_NAME='kind-registry'
SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"

# SUDO=sudo
# [[ $(uname -s) == "Darwin" ]] && {
# SUDO=
# }
SUDO=

if ! builtin type -p kind &>/dev/null; then
    echo "Install kind. https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
    exit 1
fi
kind=$(type -p kind)

# cleanup on exit (useful for running locally)
cleanup() { rm -rf ${TMPD}; }
trap cleanup EXIT

function start_registry() {
    running="$(${DOCKER} inspect -f '{{.State.Running}}' ${REG_NAME} 2>/dev/null || echo false)"

    if [[ ${running} != "true" ]]; then
	${DOCKER} rm -f "${REG_NAME}" || true
	${DOCKER} run \
		  -d --restart=always -p "${REG_PORT}:5000" \
		  -e REGISTRY_HTTP_SECRET=secret \
		  --name "${REG_NAME}" \
		  registry:2
    fi
}


function install_kind() {
    if [[ ${DOCKER} == "podman" ]]; then
	    export KIND_EXPERIMENTAL_PROVIDER=podman
    fi
    ${SUDO} $kind delete cluster --name ${KIND_CLUSTER_NAME} || true
    sed "s,%DOCKERCFG%,${HOME}/.docker/config.json," kind.yaml >${TMPD}/kconfig.yaml
    ${SUDO} ${kind} create cluster --name ${KIND_CLUSTER_NAME} --config ${SCRIPT_DIR}/kind.yaml
    mkdir -p $(dirname ${KUBECONFIG})
    ${SUDO} ${kind} --name ${KIND_CLUSTER_NAME} get kubeconfig >${KUBECONFIG}

    ${DOCKER} network connect "kind" "${REG_NAME}" 2>/dev/null || true
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${REG_PORT}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

}

main() {
    start_registry
    install_kind
}

main

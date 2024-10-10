#!/usr/bin/env bash


SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"

# Create a Kind Cluster if dont have sone
kind create cluster --name tekton-caches --config $SCRIPT_DIR/kind/kind-config.yaml


# Install Pipelines if not already installed.
#kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml

#Enable Step Actions Feature
kubectl patch configmap -n tekton-pipelines --type merge -p '{"data":{"enable-step-actions": "true"}}' feature-flags

# Create Docker creds secret Specifc to OCI Images
#kubectl create secret generic regcred  --from-file=config.json=${HOME}/.docker/config.json

# Create Secret for AWS S3
#kubectl create secret generic aws-cred  --from-file=${HOME}/.aws/config --from-file=${HOME}/.aws/credentials 

#Deploy Step Actions
ko apply -BRf $SCRIPT_DIR/step-action

# Deploy Pipelines
kubectl apply -f $SCRIPT_DIR/pipeline



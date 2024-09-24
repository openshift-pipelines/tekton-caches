#!/usr/bin/env bash


kubectl patch configmap -n tekton-pipelines --type merge -p '{"data":{"enable-step-actions": "true"}}' feature-flags

kubectl create secret generic regcred \
        --from-file=config.json=${HOME}/.docker/config.json



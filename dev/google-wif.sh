#!/usr/bin/env bash
set -x
#Step 0 - Define Common Variables

    POOL_ID=openshift-pool
    PROVIDER_ID=opeshift-wif
    NAMESPACE=default
    SERVICE_ACCOUNT=default
    PROJECT_ID=pipelines-qe
    PROJECT_NUMBER=272779626560
    MAPPED_SUBJECT=system:serviceaccount:$NAMESPACE:$SERVICE_ACCOUNT

#Step 1 - Enable IAM APIs on Google Cloud

# Step 2 - Define an attribute mapping and condition
    MAPPINGS=google.subject=assertion.sub


#Step 3 - Create workload identity pool and provider
    ISSUER=$(kubectl get --raw /.well-known/openid-configuration | jq -r .issuer)


# Download the cluster's JSON Web Key Set (JWKS):
    kubectl get --raw /openid/v1/jwks > cluster-jwks.json


#   Create a new workload identity pool:
    gcloud iam workload-identity-pools create $POOL_ID \
        --location="global" \
        --description=$POOL_ID \
        --display-name=$POOL_ID


#   Add the Kubernetes cluster as a workload identity pool provider and upload the cluster's JWKS:

    gcloud iam workload-identity-pools providers create-oidc $PROVIDER_ID \
        --location="global" \
        --workload-identity-pool=$POOL_ID \
        --issuer-uri=$ISSUER \
        --allowed-audiences=openshift  \
        --attribute-mapping=$MAPPINGS \
        --jwk-json-path="cluster-jwks.json"


# Create Service Account or use default one

#    kubectl create serviceaccount $KSA_NAME --namespace $NAMESPACE

#   Grant IAM access to the Kubernetes ServiceAccount for a Google Cloud resource.
    gcloud projects add-iam-policy-binding projects/$PROJECT_ID \
        --role=roles/owner \
        --member=principal://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$POOL_ID/subject/$MAPPED_SUBJECT \
        --condition=None

    gcloud iam workload-identity-pools create-cred-config \
        projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$POOL_ID/providers/$PROVIDER_ID \
        --credential-source-file=/workspace/token/token \
        --credential-source-type=text \
        --output-file=credential-configuration.json



    kubectl -n $NAMESPACE create secret generic gcs-cred --from-file=credential-configuration.json
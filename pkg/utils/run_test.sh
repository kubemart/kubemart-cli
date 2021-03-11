#!/bin/bash

echo "Checking if Namespace exists..."
NS_OUTPUT=$(kubectl get ns | grep kubemart)
if [ -z "$NS_OUTPUT" ]
then
    echo "Namespace not found, creating one"
    kubectl create ns kubemart-system
fi

# This dummy namespace will be deleted one of the tests
echo "***"
echo "Creating a dummy Namespace..."
kubectl create ns dummy

echo "***"
echo "Checking if ConfigMap exists..."
CM_OUTPUT=$(kubectl get cm -n kubemart-system | grep kubemart-config)
if ! [ -z "${CM_OUTPUT}" ]
then
    echo "ConfigMap found, deleting it"
    kubectl delete cm -n kubemart-system kubemart-config
fi

echo "***"
echo "Deleting ~/.kubemart folder..."
rm -rf ~/.kubemart

echo "***"
echo "Deleting Kubemart CRDs..."
kubectl delete crds --all -A

echo "***"
echo "Running tests..."
# Note:
# The `-count=1` flag means run the tests without cache
# https://stackoverflow.com/a/48882892
if [[ -z "${CI}" ]]; then
  go test -count=1 ./pkg/utils
else
  go test -count=1 -v ./pkg/utils
fi

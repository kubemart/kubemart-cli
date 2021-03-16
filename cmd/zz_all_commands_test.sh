#!/bin/bash

# =====================
# Helpers
# =====================
get_latest_release() {
    curl --silent "https://api.github.com/repos/$1/releases/latest" | # Get latest release from GitHub api
      grep '"tag_name":' |                                            # Get tag line
      sed -E 's/.*"([^"]+)".*/\1/'                                    # Pluck JSON value
}

# =====================
# Main
# =====================
REPO="kubemart/kubemart-operator"
LATEST_VERSION=$(get_latest_release $REPO)

if [ -z "$LATEST_VERSION" ]
then
    echo "Unable to fetch the latest release version of $REPO (perhaps, it's due to GitHub API rate limits)"
else
    echo "***"
    echo "Applying latest ($LATEST_VERSION) Kubermart manifests..."
    kubectl apply -f https://github.com/$REPO/releases/download/$LATEST_VERSION/kubemart-operator.yaml

    echo "***"
    echo "Checking number of apps..."
    NUM_OF_APPS=$(kubectl get apps -A --no-headers | wc -l | xargs) # <-- xargs is to trim whitespaces
    if [ $NUM_OF_APPS -ne 0 ]
    then
        echo "***"
        echo "Found $NUM_OF_APPS apps - deleting them..."
        kubectl delete apps --all -A
    fi

    echo "***"
    echo "Running tests..."
    go test -v -count=1 ./cmd
fi

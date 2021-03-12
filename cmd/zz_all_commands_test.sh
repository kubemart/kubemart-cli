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
    echo "Creating a Kind cluster..."
    kind create cluster --name default

    echo "***"
    echo "Applying latest ($LATEST_VERSION) Kubermart manifests..."
    kubectl apply -f https://github.com/$REPO/releases/download/$LATEST_VERSION/kubemart-operator.yaml

    echo "***"
    echo "Running tests..."
    go test -v -count=1 ./cmd
fi

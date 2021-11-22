[![GitHub Actions Build](https://github.com/fybrik/data-movement-operator/actions/workflows/build.yml/badge.svg)](https://github.com/fybrik/data-movement-operator/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/fybrik/data-movement-operator)](https://goreportcard.com/report/github.com/fybrik/data-movement-operator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# data-movement-operator

The data-movement-operator contains the movement functionality (batch and streaming) that is used for the implicit copy modules of the data fybrik.

# Installation

As this is an extension to the fybrik please make sure to install the [fybrik](https://github.com/fybrik/fybrik) and it's dependencies before continuing.

## Latest master branch
Installing the controller:
```
helm repo add fybrik-charts https://fybrik.github.io/charts
helm repo update
helm install data-movement-operator fybrik-charts/data-movement-operator -n fybrik-system --wait
```

Installing the fybrik module:

```
kubectl apply -f https://raw.githubusercontent.com/fybrik/data-movement-operator/master/modules/implicit-copy-batch-module.yaml -n fybrik-system
kubectl apply -f https://raw.githubusercontent.com/fybrik/data-movement-operator/master/modules/implicit-copy-stream-module.yaml -n fybrik-system
```

## Development version using the repo
1. Check out git repository
2. Be sure that the certificate manager is installed (https://cert-manager.io/docs/)
3. Execute the following
```
export DOCKER_HOSTNAME=localhost:5000
export DOCKER_NAMESPACE=fybrik-system
export VALUES_FILE=charts/data-movement-operator/integration-tests.values.yaml
make docker-build docker-push
make deploy
```

Installing the fybrik module:

```
kubectl apply -f modules/implicit-copy-batch-module.yaml -n fybrik-system
kubectl apply -f modules/implicit-copy-stream-module.yaml -n fybrik-system
```

## Issue management

We use GitHub issues to track all of our bugs and feature requests.

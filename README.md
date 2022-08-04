[![GitHub Actions Build](https://github.com/fybrik/data-movement-operator/actions/workflows/build.yml/badge.svg)](https://github.com/fybrik/data-movement-operator/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/fybrik/data-movement-operator)](https://goreportcard.com/report/github.com/fybrik/data-movement-operator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# data-movement-operator

The data-movement-operator contains the movement functionality (batch and streaming) that is used for the implicit copy modules of the data fybrik.

# Installation

As this is an extension to the fybrik please make sure to install the [fybrik](https://github.com/fybrik/fybrik) and it's dependencies before continuing.

## Installing the controller

Version 0.5.0 of the data movement operator (DMO) is available in Fybrik 0.5.x.

| Fybrik           | DMO
| ---              | ---
| 0.6.x            | 0.6.x
| 0.7.x            | 0.7.x
| 1.0.x            | 0.8.x


Installing the chart: (available from DMO version 0.6.1 and above)
```
helm repo add fybrik-charts https://fybrik.github.io/charts
helm repo update
helm install data-movement-operator fybrik-charts/data-movement-operator -n fybrik-system --wait
```

Install latest development version from GitHub:
```
git clone https://github.com/fybrik/data-movement-operator.git
helm install data-movement-operator data-movement-operator/charts/data-movement-operator -n fybrik-system --wait
```

## Register as a Fybrik module

To register the movement functionality as a Fybrik module apply `modules/implicit-copy-batch-module.yaml` or `modules/implicit-copy-stream-module.yaml` to the `fybrik-system` namespace of your cluster.

To install the latest release run:

```bash
kubectl apply -f https://github.com/fybrik/data-movement-operator/releases/latest/download/implicit-copy-batch-module.yaml -n fybrik-system
kubectl apply -f https://github.com/fybrik/data-movement-operator/releases/latest/download/implicit-copy-stream-module.yaml -n fybrik-system
```

### Version compatbility matrix

CBM - copy batch module
CSM - copy stream module

| Fybrik           | CBM     | Mover   | Command
| ---              | ---     | ---     | ---
| 0.5.x            | 0.5.x   | 0.5.x   | `https://github.com/fybrik/data-movement-operator/releases/download/v0.5.0/implicit-copy-batch-module.yaml`
| 0.6.x            | 0.6.x   | 0.5.x   | `https://github.com/fybrik/data-movement-operator/releases/download/v0.6.0/implicit-copy-batch-module.yaml`
| 0.7.x            | 0.7.x   | 0.5.x   | `https://github.com/fybrik/data-movement-operator/releases/download/v0.7.0/implicit-copy-batch-module.yaml`
| 1.0.x            | 0.8.x   | 0.5.x   | `https://github.com/fybrik/data-movement-operator/releases/download/v0.8.0/implicit-copy-batch-module.yaml`
| master           | master  | master  | `https://raw.githubusercontent.com/fybrik/data-movement-operator/master/modules/implicit-copy-batch-module.yaml`

| Fybrik           | CSM     | Mover   | Command
| ---              | ---     | ---     | ---
| 0.5.x            | 0.5.x   | 0.5.x   | `https://github.com/fybrik/data-movement-operator/releases/download/v0.5.0/implicit-copy-stream-module.yaml`
| 0.6.x            | 0.6.x   | 0.5.x   | `https://github.com/fybrik/data-movement-operator/releases/download/v0.6.0/implicit-copy-stream-module.yaml`
| 0.7.x            | 0.7.x   | 0.5.x   | `https://github.com/fybrik/data-movement-operator/releases/download/v0.7.0/implicit-copy-stream-module.yaml`
| 1.0.x            | 0.8.x   | 0.5.x   | `https://github.com/fybrik/data-movement-operator/releases/download/v0.8.0/implicit-copy-stream-module.yaml`
| master           | master  | master  | `https://raw.githubusercontent.com/fybrik/data-movement-operator/master/modules/implicit-copy-stream-module.yaml`

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

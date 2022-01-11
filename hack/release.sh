#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

: ${RELEASE:=master}
: ${TOOLBIN:=./hack/tools/bin}

${TOOLBIN}/yq eval --inplace ".version = \"$RELEASE\"" ./charts/data-movement-operator/Chart.yaml
${TOOLBIN}/yq eval --inplace ".appVersion = \"$RELEASE\"" ./charts/data-movement-operator/Chart.yaml
${TOOLBIN}/yq eval --inplace ".global.tag = \"$RELEASE\"" ./charts/data-movement-operator/values.yaml

${TOOLBIN}/yq eval --inplace ".version = \"$RELEASE\"" ./modules/fybrik-implicit-copy-batch/Chart.yaml
${TOOLBIN}/yq eval --inplace ".version = \"$RELEASE\"" ./modules/fybrik-implicit-copy-stream/Chart.yaml

${TOOLBIN}/yq eval --inplace ".spec.chart.name = \"ghcr.io/fybrik/fybrik-implicit-copy-batch:$RELEASE\"" ./modules/implicit-copy-batch-module.yaml
${TOOLBIN}/yq eval --inplace ".spec.chart.name = \"ghcr.io/fybrik/fybrik-implicit-copy-stream:$RELEASE\"" ./modules/implicit-copy-stream-module.yaml

${TOOLBIN}/yq eval --inplace ".metadata.labels.version = \"$RELEASE\"" ./modules/implicit-copy-batch-module.yaml
${TOOLBIN}/yq eval --inplace ".metadata.labels.version = \"$RELEASE\"" ./modules/implicit-copy-stream-module.yaml

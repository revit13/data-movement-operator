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

${TOOLBIN}/yq eval --inplace ".image = \"ghcr.io/fybrik/mover:$RELEASE\"" modules/fybrik-implicit-copy-batch/values.yaml
${TOOLBIN}/yq eval --inplace ".image = \"ghcr.io/fybrik/mover:$RELEASE\"" modules/fybrik-implicit-copy-stream/values.yaml

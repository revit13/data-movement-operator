// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package controllers

// This is a collection of constants related to the manager and it's configuration

const BatchTransferConcurrentReconcilesConfiguration = "BATCHTRANSFER_CONCURRENT_RECONCILES"
const StreamTransferConcurrentReconcilesConfiguration = "STREAMTRANSFER_CONCURRENT_RECONCILES"

const KubernetesClientQPSConfiguration = "CLIENT_QPS"
const KubernetesClientBurstConfiguration = "CLIENT_BURST"

const DefaultBatchTransferConcurrentReconciles = 1
const DefaultStreamTransferConcurrentReconciles = 1

const DefaultKubernetesClientQPS = 5.0  // Default from Kubernetes client: 5
const DefaultKubernetesClientBurst = 10 // Default from Kubernetes client: 10

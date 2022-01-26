// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"

	kapps "k8s.io/api/apps/v1"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	motionv1 "fybrik.io/data-movement-controller/manager/apis/motion/v1alpha1"
	"fybrik.io/data-movement-controller/manager/controllers"
	"fybrik.io/data-movement-controller/manager/controllers/motion"
	"fybrik.io/data-movement-controller/manager/controllers/utils"
	"fybrik.io/data-movement-controller/pkg/environment"
)

const (
	managerPort   = 9443
	listeningPort = 8085
)

var (
	scheme   = kruntime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = motionv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = kbatch.AddToScheme(scheme)
	_ = kapps.AddToScheme(scheme)
}

func run(namespace, metricsAddr string, enableLeaderElection bool) int {
	setupLog.Info("creating manager")

	blueprintNamespace := utils.GetBlueprintNamespace()

	systemNamespaceSelector := fields.SelectorFromSet(fields.Set{"metadata.namespace": utils.GetSystemNamespace()})
	workerNamespaceSelector := fields.SelectorFromSet(fields.Set{"metadata.namespace": blueprintNamespace})

	setupLog.Info("Watching BatchTransfers and StreamTransfers", "namespace", workerNamespaceSelector)
	selectorsByObject := cache.SelectorsByObject{
		&corev1.ConfigMap{}:             {Field: systemNamespaceSelector},
		&motionv1.BatchTransfer{}:       {Field: workerNamespaceSelector},
		&motionv1.StreamTransfer{}:      {Field: workerNamespaceSelector},
		&kbatch.Job{}:                   {Field: workerNamespaceSelector},
		&kbatch.CronJob{}:               {Field: workerNamespaceSelector},
		&corev1.Secret{}:                {Field: workerNamespaceSelector},
		&corev1.Pod{}:                   {Field: workerNamespaceSelector},
		&kapps.Deployment{}:             {Field: workerNamespaceSelector},
		&corev1.PersistentVolumeClaim{}: {Field: workerNamespaceSelector},
	}

	client := ctrl.GetConfigOrDie()
	client.QPS = environment.GetEnvAsFloat32(controllers.KubernetesClientQPSConfiguration, controllers.DefaultKubernetesClientQPS)
	client.Burst = environment.GetEnvAsInt(controllers.KubernetesClientBurstConfiguration, controllers.DefaultKubernetesClientBurst)

	setupLog.Info("Manager client rate limits:", "qps", client.QPS, "burst", client.Burst)

	mgr, err := ctrl.NewManager(client, ctrl.Options{
		Scheme:             scheme,
		Namespace:          namespace,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "data-movement-operator-leader-election",
		Port:               managerPort,
		NewCache:           cache.BuilderWithOptions(cache.Options{SelectorsByObject: selectorsByObject}),
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return 1
	}

	setupLog.Info("creating motion controllers")
	if err := motion.SetupMotionControllers(mgr); err != nil {
		setupLog.Error(err, "unable to setup motion controllers")
		return 1
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return 1
	}

	return 0
}

// Main entry point starts manager and controllers
func main() {
	var namespace string
	var metricsAddr string
	var enableLeaderElection bool
	address := utils.ListeningAddress(listeningPort)

	flag.StringVar(&metricsAddr, "metrics-bind-addr", address, "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&namespace, "namespace", "", "The namespace to which this controller manager is limited.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	os.Exit(run(namespace, metricsAddr, enableLeaderElection))
}

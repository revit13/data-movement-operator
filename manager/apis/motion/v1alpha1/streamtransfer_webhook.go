// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"log"
	"os"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const streamTransferImage = "ghcr.io/fybrik/mover:latest"

func (streamTransfer *StreamTransfer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(streamTransfer).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:admissionReviewVersions=v1;v1beta1,sideEffects=None,path=/mutate-motion-fybrik-io-v1alpha1-streamtransfer,mutating=true,failurePolicy=fail,groups=motion.fybrik.io,resources=streamtransfers,verbs=create;update,versions=v1alpha1,name=mstreamtransfer.kb.io

var _ webhook.Defaulter = &StreamTransfer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (streamTransfer *StreamTransfer) Default() {
	log.Printf("Defaulting streamtransfer %s", streamTransfer.Name)
	if streamTransfer.Spec.Image == "" {
		// TODO check if can be removed after upgrading controller-gen to 0.5.0
		streamTransfer.Spec.Image = streamTransferImage
	}

	if streamTransfer.Spec.ImagePullPolicy == "" {
		// TODO check if can be removed after upgrading controller-gen to 0.5.0
		streamTransfer.Spec.ImagePullPolicy = v1.PullIfNotPresent
	}

	if streamTransfer.Spec.SecretProviderURL == "" {
		if env, b := os.LookupEnv("SECRET_PROVIDER_URL"); b {
			streamTransfer.Spec.SecretProviderURL = env
		}
	}

	if streamTransfer.Spec.SecretProviderRole == "" {
		if env, b := os.LookupEnv("SECRET_PROVIDER_ROLE"); b {
			streamTransfer.Spec.SecretProviderRole = env
		}
	}

	if streamTransfer.Spec.TriggerInterval == "" {
		streamTransfer.Spec.TriggerInterval = "5 seconds"
	}

	defaultDataStoreDescription(&streamTransfer.Spec.Source)
	defaultDataStoreDescription(&streamTransfer.Spec.Destination)

	if streamTransfer.Spec.WriteOperation == "" {
		streamTransfer.Spec.WriteOperation = Append
	}

	if streamTransfer.Spec.DataFlowType == "" {
		streamTransfer.Spec.DataFlowType = Stream
	}

	if streamTransfer.Spec.ReadDataType == "" {
		streamTransfer.Spec.ReadDataType = ChangeData
	}

	if streamTransfer.Spec.WriteDataType == "" {
		streamTransfer.Spec.WriteDataType = LogData
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,admissionReviewVersions=v1;v1beta1,sideEffects=None,path=/validate-motion-fybrik-io-v1alpha1-streamtransfer,mutating=false,failurePolicy=fail,groups=motion.fybrik.io,resources=streamtransfers,versions=v1alpha1,name=vstreamtransfer.kb.io

var _ webhook.Validator = &StreamTransfer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (streamTransfer *StreamTransfer) ValidateCreate() error {
	log.Printf("Validating streamtransfer %s for creation", streamTransfer.Name)

	return streamTransfer.validateStreamTransfer()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (streamTransfer *StreamTransfer) ValidateUpdate(old runtime.Object) error {
	log.Printf("Validating streamtransfer %s for update", streamTransfer.Name)

	return streamTransfer.validateStreamTransfer()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (streamTransfer *StreamTransfer) ValidateDelete() error {
	log.Printf("Validating streamtransfer %s for deletion", streamTransfer.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (streamTransfer *StreamTransfer) validateStreamTransfer() error {
	var allErrs field.ErrorList
	specField := field.NewPath("spec")

	if streamTransfer.Spec.DataFlowType == Batch {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("dataFlowType"),
			streamTransfer.Spec.DataFlowType, "'dataFlowType' must be 'Stream' for a StreamTransfer!"))
	}

	if err := validateDataStore(specField.Child("source"), &streamTransfer.Spec.Source); err != nil {
		allErrs = append(allErrs, err...)
	}
	if err := validateDataStore(specField.Child("destination"), &streamTransfer.Spec.Destination); err != nil {
		allErrs = append(allErrs, err...)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "motion.fybrik.io", Kind: "BatchTransfer"},
		streamTransfer.Name, allErrs)
}

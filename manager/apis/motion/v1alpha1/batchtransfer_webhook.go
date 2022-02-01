// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/robfig/cron"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	validationutils "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	DefaultFailedJobHistoryLimit     = 5
	MaxFailedJobHistoryLimit         = 20
	DefaultSuccessfulJobHistoryLimit = 5
	MaxSuccessfulJobHistoryLimit     = 20
	batchTransferImage               = "ghcr.io/fybrik/mover:latest"
)

func (batchTransfer *BatchTransfer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(batchTransfer).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// +kubebuilder:webhook:admissionReviewVersions=v1;v1beta1,sideEffects=None,path=/mutate-motion-fybrik-io-v1alpha1-batchtransfer,mutating=true,failurePolicy=fail,groups=motion.fybrik.io,resources=batchtransfers,verbs=create;update,versions=v1alpha1,name=mbatchtransfer.kb.io

var _ webhook.Defaulter = &BatchTransfer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
//nolint:gocyclo
func (batchTransfer *BatchTransfer) Default() {
	log.Printf("Defaulting batchtransfer %s", batchTransfer.Name)
	if batchTransfer.Spec.Image == "" {
		// TODO check if can be removed after upgrading controller-gen to 0.5.0
		batchTransfer.Spec.Image = batchTransferImage
	}

	if batchTransfer.Spec.ImagePullPolicy == "" {
		// TODO check if can be removed after upgrading controller-gen to 0.5.0
		batchTransfer.Spec.ImagePullPolicy = v1.PullIfNotPresent
	}

	if batchTransfer.Spec.SecretProviderURL == "" {
		if env, b := os.LookupEnv("SECRET_PROVIDER_URL"); b {
			batchTransfer.Spec.SecretProviderURL = env
		}
	}

	if batchTransfer.Spec.SecretProviderRole == "" {
		if env, b := os.LookupEnv("SECRET_PROVIDER_ROLE"); b {
			batchTransfer.Spec.SecretProviderRole = env
		}
	}

	if batchTransfer.Spec.FailedJobHistoryLimit == 0 {
		batchTransfer.Spec.FailedJobHistoryLimit = DefaultFailedJobHistoryLimit
	}

	if batchTransfer.Spec.SuccessfulJobHistoryLimit == 0 {
		batchTransfer.Spec.SuccessfulJobHistoryLimit = DefaultSuccessfulJobHistoryLimit
	}

	if batchTransfer.Spec.Spark != nil {
		if batchTransfer.Spec.Spark.Image == "" {
			batchTransfer.Spec.Spark.Image = batchTransfer.Spec.Image
		}

		if batchTransfer.Spec.Spark.ImagePullPolicy == "" {
			batchTransfer.Spec.Spark.ImagePullPolicy = batchTransfer.Spec.ImagePullPolicy
		}
	}

	if env, b := os.LookupEnv("NO_FINALIZER"); b {
		if parsedBool, err := strconv.ParseBool(env); err != nil {
			panic(fmt.Sprintf("Cannot parse boolean value %s: %s", env, err.Error()))
		} else {
			batchTransfer.Spec.NoFinalizer = parsedBool
		}
	}

	defaultDataStoreDescription(&batchTransfer.Spec.Source)
	defaultDataStoreDescription(&batchTransfer.Spec.Destination)

	if batchTransfer.Spec.WriteOperation == "" {
		batchTransfer.Spec.WriteOperation = Overwrite
	}

	if batchTransfer.Spec.DataFlowType == "" {
		batchTransfer.Spec.DataFlowType = Batch
	}

	if batchTransfer.Spec.ReadDataType == "" {
		batchTransfer.Spec.ReadDataType = LogData
	}

	if batchTransfer.Spec.WriteDataType == "" {
		batchTransfer.Spec.WriteDataType = LogData
	}
}

func defaultDataStoreDescription(dataStore *DataStore) {
	if dataStore.Description == "" {
		switch {
		case dataStore.Database != nil:
			dataStore.Description = dataStore.Database.Db2URL + "/" + dataStore.Database.Table
		case dataStore.Kafka != nil:
			dataStore.Description = "kafka://" + dataStore.Kafka.KafkaTopic
		case dataStore.S3 != nil:
			dataStore.Description = "s3://" + dataStore.S3.Bucket + "/" + dataStore.S3.ObjectKey
		case dataStore.Cloudant != nil:
			dataStore.Description = "cloudant://" + dataStore.Cloudant.Host + "/" + dataStore.Cloudant.Database
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,admissionReviewVersions=v1;v1beta1,sideEffects=None,path=/validate-motion-fybrik-io-v1alpha1-batchtransfer,mutating=false,failurePolicy=fail,groups=motion.fybrik.io,resources=batchtransfers,versions=v1alpha1,name=vbatchtransfer.kb.io

var _ webhook.Validator = &BatchTransfer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (batchTransfer *BatchTransfer) ValidateCreate() error {
	log.Printf("Validating batchtransfer %s for creation", batchTransfer.Name)
	return batchTransfer.validateBatchTransfer()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (batchTransfer *BatchTransfer) ValidateUpdate(old runtime.Object) error {
	log.Printf("Validating batchtransfer %s for update", batchTransfer.Name)

	return batchTransfer.validateBatchTransfer()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (batchTransfer *BatchTransfer) ValidateDelete() error {
	log.Printf("Validating batchtransfer %s for deletion", batchTransfer.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (batchTransfer *BatchTransfer) validateBatchTransfer() error {
	var allErrs field.ErrorList
	specField := field.NewPath("spec")
	if err := batchTransfer.validateBatchTransferSpec(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := validateDataStore(specField.Child("source"), &batchTransfer.Spec.Source); err != nil {
		allErrs = append(allErrs, err...)
	}
	if err := validateDataStore(specField.Child("destination"), &batchTransfer.Spec.Destination); err != nil {
		allErrs = append(allErrs, err...)
	}
	if batchTransfer.Spec.SuccessfulJobHistoryLimit < 0 || batchTransfer.Spec.SuccessfulJobHistoryLimit > MaxSuccessfulJobHistoryLimit {
		allErrs = append(allErrs, field.Invalid(specField.Child("successfulJobHistoryLimit"), batchTransfer.Spec.SuccessfulJobHistoryLimit,
			"'successfulJobHistoryLimit' has to be between 0 and "+strconv.Itoa(MaxSuccessfulJobHistoryLimit)))
	}
	if batchTransfer.Spec.FailedJobHistoryLimit < 0 || batchTransfer.Spec.FailedJobHistoryLimit > MaxFailedJobHistoryLimit {
		allErrs = append(allErrs, field.Invalid(specField.Child("failedJobHistoryLimit"), batchTransfer.Spec.FailedJobHistoryLimit,
			"'failedJobHistoryLimit' has to be between 0 and "+strconv.Itoa(MaxFailedJobHistoryLimit)))
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "motion.fybrik.io", Kind: "BatchTransfer"},
		batchTransfer.Name, allErrs)
}

//nolint:gocyclo,funlen
func validateDataStore(path *field.Path, store *DataStore) []*field.Error {
	var allErrs []*field.Error

	if store.Database != nil {
		var db = store.Database
		databasePath := path.Child("database")
		if db.Password == "" && db.Vault != nil {
			allErrs = append(allErrs, field.Invalid(databasePath, db.Vault, "Can only set vault or password!"))
		}

		match, _ := regexp.MatchString("^jdbc:[a-z0-9]+://", db.Db2URL)
		if !match {
			allErrs = append(allErrs, field.Invalid(databasePath.Child("db2URL"), db.Db2URL, "Invalid JDBC string!"))
		}
		r := regexp.MustCompile("^jdbc:[a-z0-9]+://([a-z0-9.-]+):")
		host := r.FindStringSubmatch(db.Db2URL)[1]

		if msgs := validationutils.IsDNS1123Subdomain(host); len(msgs) != 0 {
			allErrs = append(allErrs, field.Invalid(databasePath.Child("db2URL"), db.Db2URL, "Invalid database host!"))
		}

		if db.Table == "" {
			allErrs = append(allErrs, field.Invalid(path, db.Table, "Table cannot be empty!"))
		}
	}

	if store.S3 != nil {
		s3Path := path.Child("s3")
		_, err := url.Parse(store.S3.Endpoint)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(s3Path.Child("endpoint"), store.S3.Endpoint, "Invalid endpoint! Expecting a endpoint URL!"))
		}

		if store.S3.Bucket == "" {
			allErrs = append(allErrs, field.Invalid(s3Path.Child("bucket"), store.S3.Bucket, validationutils.EmptyError()))
		}

		if store.S3.ObjectKey == "" {
			allErrs = append(allErrs, field.Invalid(s3Path.Child("objectKey"), store.S3.ObjectKey, validationutils.EmptyError()))
		}

		if (store.S3.AccessKey != "" || store.S3.SecretKey != "") && store.S3.Vault != nil {
			allErrs = append(allErrs, field.Invalid(s3Path, store.S3.Vault, "Can only set vault or accessKey/secretKey!"))
		}
	}

	if store.Kafka != nil {
		kafkaPath := path.Child("kafka")

		// Validate Kafka brokers
		kafkaBrokers := strings.Split(store.Kafka.KafkaBrokers, ",")
		if len(kafkaBrokers) == 0 {
			allErrs = append(allErrs, field.Invalid(kafkaPath.Child("kafkaBrokers"), store.Kafka.KafkaBrokers, "Could not parse kafka brokers!"))
		}
		for i, broker := range kafkaBrokers {
			errs := validateHostPort(broker)
			for _, err := range errs {
				errMsg := fmt.Sprintf("Invalid broker at position %d (%s) %s", i, broker, err)
				allErrs = append(allErrs, field.Invalid(kafkaPath.Child("kafkaBrokers"), store.Kafka.SchemaRegistryURL, errMsg))
			}
		}

		// Validate Kafka schema registry url
		if store.Kafka.SchemaRegistryURL != "" {
			schemaRegistryURL, err := url.Parse(store.Kafka.SchemaRegistryURL)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(kafkaPath.Child("schemaRegistryUrl"), store.Kafka.SchemaRegistryURL, "Could not parse url!"))
			}
			if schemaRegistryURL != nil {
				errs := validateHostPort(schemaRegistryURL.Host)
				for _, err := range errs {
					errMsg := "Invalid host: " + err
					allErrs = append(allErrs, field.Invalid(kafkaPath.Child("schemaRegistryUrl"), store.Kafka.SchemaRegistryURL, errMsg))
				}
			}
		}

		// Validate Kafka topic
		if store.Kafka.KafkaTopic == "" {
			allErrs = append(allErrs, field.Invalid(kafkaPath.Child("kafkaTopic"), store.Kafka.KafkaTopic, validationutils.EmptyError()))
		}

		if store.Kafka.Password != "" && store.Kafka.Vault != nil {
			allErrs = append(allErrs, field.Invalid(kafkaPath, store.Kafka.Vault, "Can only set vault or password!"))
		}

		if store.Kafka.DataFormat != "" && store.Kafka.DataFormat != "avro" && store.Kafka.DataFormat != "json" {
			allErrs = append(allErrs,
				field.Invalid(kafkaPath, store.Kafka.DataFormat, "Currently only 'avro' and 'json' are supported as Kafka dataFormat!"))
		}
	}

	return allErrs
}

func (batchTransfer *BatchTransfer) validateBatchTransferSpec() *field.Error {
	// The field helpers from the kubernetes API machinery help us return nicely
	// structured validation errors.
	if len(batchTransfer.Spec.Schedule) > 0 {
		return validateScheduleFormat(
			batchTransfer.Spec.Schedule,
			field.NewPath("spec").Child("schedule"))
	}

	if batchTransfer.Spec.DataFlowType == Stream {
		return field.Invalid(field.NewPath("spec").Child("dataFlowType"),
			batchTransfer.Spec.DataFlowType, "'dataFlowType' must be 'Batch' for a BatchTransfer!")
	}

	return nil
}

func validateScheduleFormat(schedule string, fldPath *field.Path) *field.Error {
	if _, err := cron.ParseStandard(schedule); err != nil {
		return field.Invalid(fldPath, schedule, err.Error())
	}
	return nil
}

// Validates a host port combination e.g localhost:8080
// Validates if the host is a valid domain and if the port is a valid port number
func validateHostPort(hostPort string) []string {
	var errs []string
	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		errs = append(errs, err.Error())
	}

	if msgs := validationutils.IsDNS1123Subdomain(host); len(msgs) != 0 {
		errs = append(errs, msgs...)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		errs = append(errs, fmt.Sprintf("Could not parse port %s as an integer", portStr))
	}
	if msgs := validationutils.IsValidPortNum(port); len(msgs) != 0 {
		for _, err := range msgs {
			errs = append(errs, "Port "+err)
		}
	}

	return errs
}

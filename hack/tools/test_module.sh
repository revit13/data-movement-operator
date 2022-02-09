#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -x
set -e


export WORKING_DIR=test-script
export ACCESS_KEY=1234
export SECRET_KEY=1234

kubernetesVersion=$1
fybrikVersion=$2
moduleVersion=$3
module=$4
certManagerVersion=$5

# Trim the last two charts of the module version
# to construct the module resource path
moduleResourceVersion=${moduleVersion%??}".0"


if [ $kubernetesVersion == "kind19" ]
then
    bin/kind delete cluster
    bin/kind create cluster --image=kindest/node:v1.19.11@sha256:07db187ae84b4b7de440a73886f008cf903fcf5764ba8106a9fd5243d6f32729
elif [ $kubernetesVersion == "kind20" ]
then
    bin/kind delete cluster
    bin/kind create cluster --image=kindest/node:v1.20.7@sha256:cbeaf907fc78ac97ce7b625e4bf0de16e3ea725daf6b04f930bd14c67c671ff9
elif [ $kubernetesVersion == "kind21" ]
then
    bin/kind delete cluster
    bin/kind create cluster --image=kindest/node:v1.21.1@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6
elif [ $kubernetesVersion == "kind22" ]
then
    bin/kind delete cluster
    bin/kind create cluster --image=kindest/node:v1.22.0@sha256:b8bda84bb3a190e6e028b1760d277454a72267a5454b57db34437c34a588d047
else
    echo "Unsupported kind version"
    exit 1
fi


#quick start

bin/helm repo add jetstack https://charts.jetstack.io
bin/helm repo add hashicorp https://helm.releases.hashicorp.com
bin/helm repo add fybrik-charts https://fybrik.github.io/charts
bin/helm repo update


bin/helm install cert-manager jetstack/cert-manager \
    --namespace cert-manager \
    --version v$certManagerVersion \
    --create-namespace \
    --set installCRDs=true \
    --wait --timeout 400s



bin/helm install vault fybrik-charts/vault --create-namespace -n fybrik-system \
        --set "vault.injector.enabled=false" \
        --set "vault.server.dev.enabled=true" \
        --values https://raw.githubusercontent.com/fybrik/fybrik/v$fybrikVersion/charts/vault/env/dev/vault-single-cluster-values.yaml
    bin/kubectl wait --for=condition=ready --all pod -n fybrik-system --timeout=400s

bin/helm install fybrik-crd fybrik-charts/fybrik-crd -n fybrik-system --version v$fybrikVersion --wait
bin/helm install fybrik fybrik-charts/fybrik -n fybrik-system --version v$fybrikVersion --wait


if [ $fybrikVersion != 0.5* ]
then
    # TODO: deploy from fybrik/chart repo when its available
    git clone https://github.com/fybrik/data-movement-operator.git
    cd data-movement-operator/
    ../bin/helm install data-movement-operator charts/data-movement-operator -n fybrik-system --wait
    cd ..
    rm -rf data-movement-operator
fi

# apply arrow flight module without actions
bin/kubectl apply -f $WORKING_DIR/flight-module-$moduleResourceVersion.yaml  -n fybrik-system
if [ $module == "stream" ]
then
    bin/kubectl apply -f https://github.com/fybrik/data-movement-operator/releases/download/v$moduleVersion/implicit-copy-stream-module.yaml -n fybrik-system
else
    bin/kubectl apply -f https://github.com/fybrik/data-movement-operator/releases/download/v$moduleVersion/implicit-copy-batch-module.yaml -n fybrik-system
fi


#datashim
bin/kubectl apply -f ../../third_party/datashim/dlf.yaml
bin/kubectl wait --for=condition=ready pods -l app.kubernetes.io/name=dlf -n dlf --timeout=500s



# Notebook sample

bin/kubectl create namespace fybrik-notebook-sample
bin/kubectl config set-context --current --namespace=fybrik-notebook-sample

#localstack
bin/helm repo add localstack-charts https://localstack.github.io/helm-charts
bin/helm install localstack localstack-charts/localstack --set startServices="s3" --set service.type=ClusterIP
bin/kubectl wait --for=condition=ready --all pod -n fybrik-notebook-sample --timeout=400s

bin/kubectl port-forward svc/localstack 4566:4566 &

export ENDPOINT="http://127.0.0.1:4566"
export BUCKET="demo"
export OBJECT_KEY="PS_20174392719_1491204439457_log.csv"
export FILEPATH="$WORKING_DIR/PS_20174392719_1491204439457_log.csv"
aws configure set aws_access_key_id ${ACCESS_KEY} && aws configure set aws_secret_access_key ${SECRET_KEY} && aws --endpoint-url=${ENDPOINT} s3api create-bucket --bucket ${BUCKET} && aws --endpoint-url=${ENDPOINT} s3api put-object --bucket ${BUCKET} --key ${OBJECT_KEY} --body ${FILEPATH}


cat << EOF | bin/kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: paysim-csv
type: Opaque
stringData:
  access_key: "${ACCESS_KEY}"
  secret_key: "${SECRET_KEY}"
EOF


bin/kubectl apply -f $WORKING_DIR/Asset-$moduleResourceVersion.yaml -n fybrik-notebook-sample


#fybrikstorage
cat << EOF | bin/kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: bucket-creds
  namespace: fybrik-system
type: Opaque
stringData:
  access_key: "${ACCESS_KEY}"
  accessKeyID: "${ACCESS_KEY}"
  secret_key: "${SECRET_KEY}"
  secretAccessKey: "${SECRET_KEY}"
EOF

bin/kubectl apply -f $WORKING_DIR/fybrikStorage-$moduleResourceVersion.yaml -n fybrik-system


bin/kubectl -n fybrik-system create configmap sample-policy --from-file=$WORKING_DIR/sample-policy-$moduleResourceVersion.rego
bin/kubectl -n fybrik-system label configmap sample-policy openpolicyagent.org/policy=rego

c=0
while [[ $(bin/kubectl get cm sample-policy -n fybrik-system -o 'jsonpath={.metadata.annotations.openpolicyagent\.org/policy-status}') != '{"status":"ok"}' ]]
do
    echo "waiting"
    ((c++)) && ((c==25)) && break
    sleep 5
done


bin/kubectl apply -f $WORKING_DIR/fybrikapplication-$moduleResourceVersion.yaml

c=0
while [[ $(bin/kubectl get fybrikapplication my-notebook -o 'jsonpath={.status.ready}') != "true" ]]
do
    echo "waiting"
    ((c++)) && ((c==30)) && break
    sleep 6
done

bin/kubectl get pods -n fybrik-blueprints
POD_NAME=$(bin/kubectl get pods -l app.kubernetes.io/name=arrow-flight-module-chart -n fybrik-blueprints -o=name | sed "s/^.\{4\}//")


bin/kubectl cp $WORKING_DIR/test.py ${POD_NAME}:/tmp -n fybrik-blueprints
bin/kubectl exec -it ${POD_NAME} -n fybrik-blueprints -- python /tmp/test.py > res.out

DIFF=$(diff -b $WORKING_DIR/expected.txt res.out)
RES=0
if [ "${DIFF}" == "" ]
then
    echo "test succeeded"
else
    RES=1
fi

pkill kubectl
bin/kubectl delete namespace fybrik-notebook-sample
bin/kubectl -n fybrik-system delete configmap sample-policy

if [ ${RES} == 1 ]
then
  echo "test failed"
  exit 1
fi

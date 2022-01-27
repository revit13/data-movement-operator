#!/usr/bin/env bash

set -x
set -e


export WORKING_DIR=test-script
export ACCESS_KEY=1234
export SECRET_KEY=1234

kubernetesVersion=$1
fybrikVersion=$2
moduleVersion=$3
module=$4

if [ $kubernetesVersion == "kind19" ]
then
    kind delete cluster
    kind create cluster --image=kindest/node:v1.19.11@sha256:07db187ae84b4b7de440a73886f008cf903fcf5764ba8106a9fd5243d6f32729

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

helm repo add jetstack https://charts.jetstack.io
helm repo add hashicorp https://helm.releases.hashicorp.com
helm repo add fybrik-charts https://fybrik.github.io/charts
helm repo update


helm install cert-manager jetstack/cert-manager \
    --namespace cert-manager \
    --version v1.2.0 \
    --create-namespace \
    --set installCRDs=true \
    --wait --timeout 400s



helm install vault fybrik-charts/vault --create-namespace -n fybrik-system \
        --set "vault.injector.enabled=false" \
        --set "vault.server.dev.enabled=true" \
        --values https://raw.githubusercontent.com/fybrik/fybrik/v0.5.3/charts/vault/env/dev/vault-single-cluster-values.yaml
    kubectl wait --for=condition=ready --all pod -n fybrik-system --timeout=400s

helm install fybrik-crd fybrik-charts/fybrik-crd -n fybrik-system --version v$fybrikVersion --wait
helm install fybrik fybrik-charts/fybrik -n fybrik-system --version v$fybrikVersion --wait


# cd /data/checkagain14/data-movement-operator/
if [ $fybrikVersion == "0.6.0" ]
then
    git clone https://github.com/fybrik/data-movement-operator.git
    cd data-movement-operator/
    git checkout releases/0.6.0
    helm install data-movement-operator charts/data-movement-operator -n fybrik-system --wait
    cd ..
    rm -rf data-movement-operator
fi
#cd /data/fybrik-release-0.5.0/fybrik
#helm install fybrik-crd charts/fybrik-crd -n fybrik-system --wait
#helm install fybrik charts/fybrik --set global.tag=0.5.3 --set global.imagePullPolicy=Always -n fybrik-system --wait



# apply modules
#kubectl apply -f https://github.com/fybrik/arrow-flight-module/releases/latest/download/module.yaml -n fybrik-system
# kubectl apply -f ${WORKING_DIR}/arrow_$moduleVersion_noactions.yaml -n fybrik-system

kubectl apply -f https://github.com/fybrik/arrow-flight-module/releases/download/v$moduleVersion/module.yaml -n fybrik-system
#kubectl apply -f ${WORKING_DIR}/flight.yaml -n fybrik-system
if [ $module == "stream" ]
then
    kubectl apply -f https://github.com/fybrik/data-movement-operator/releases/download/v$moduleVersion/implicit-copy-stream-module.yaml -n fybrik-system
else
    kubectl apply -f https://github.com/fybrik/data-movement-operator/releases/download/v$moduleVersion/implicit-copy-batch-module.yaml -n fybrik-system
fi
# kubectl apply -f /data/checkagain14/data-movement-operator/modules/implicit-copy-batch-module.yaml -n fybrik-system


# kubectl apply -f ${WORKING_DIR}/arrow_$moduleVersion_noactions.yaml -n fybrik-system
#kubectl apply -f ${WORKING_DIR}/flight.yaml -n fybrik-system
# kubectl apply -f https://github.com/fybrik/data-movement-operator/releases/download/v$moduleVersion/implicit-copy-batch-module.yaml -n fybrik-system
# kubectl apply -f ${WORKING_DIR}/arrow_0.6.0_noactions.yaml -n fybrik-system
# kubectl apply -f /data/checkagain14/data-movement-operator/modules/implicit-copy-batch-module.yaml -n fybrik-system


#datashim
kubectl apply -f https://raw.githubusercontent.com/datashim-io/datashim/master/release-tools/manifests/dlf.yaml
kubectl wait --for=condition=ready pods -l app.kubernetes.io/name=dlf -n dlf --timeout=500s
# cd $PATH_TO_LOCAL_FYBRIK/third_party/datashim/
# make deploy



# source ${EXPORT_FILE}



# Notebook sample

kubectl create namespace fybrik-notebook-sample
kubectl config set-context --current --namespace=fybrik-notebook-sample

#localstack
helm repo add localstack-charts https://localstack.github.io/helm-charts
helm install localstack localstack-charts/localstack --set startServices="s3" --set service.type=ClusterIP
kubectl wait --for=condition=ready --all pod -n fybrik-notebook-sample --timeout=400s

kubectl port-forward svc/localstack 4566:4566 &

export ENDPOINT="http://127.0.0.1:4566"
export BUCKET="demo"
export OBJECT_KEY="PS_20174392719_1491204439457_log.csv"
export FILEPATH="$WORKING_DIR/PS_20174392719_1491204439457_log.csv"
aws configure set aws_access_key_id ${ACCESS_KEY} && aws configure set aws_secret_access_key ${SECRET_KEY} && aws --endpoint-url=${ENDPOINT} s3api create-bucket --bucket ${BUCKET} && aws --endpoint-url=${ENDPOINT} s3api put-object --bucket ${BUCKET} --key ${OBJECT_KEY} --body ${FILEPATH}


cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: paysim-csv
type: Opaque
stringData:
  access_key: "${ACCESS_KEY}"
  secret_key: "${SECRET_KEY}"
EOF


kubectl apply -f $WORKING_DIR/Asset-$moduleVersion.yaml -n fybrik-notebook-sample




#fybrikstorage
cat << EOF | kubectl apply -f -
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

kubectl apply -f $WORKING_DIR/fybrikStorage-$moduleVersion.yaml -n fybrik-system


kubectl -n fybrik-system create configmap sample-policy --from-file=$WORKING_DIR/sample-policy-$moduleVersion.rego
kubectl -n fybrik-system label configmap sample-policy openpolicyagent.org/policy=rego

c=0
while [[ $(kubectl get cm sample-policy -n fybrik-system -o 'jsonpath={.metadata.annotations.openpolicyagent\.org/policy-status}') != '{"status":"ok"}' ]]
do
    echo "waiting"
    ((c++)) && ((c==25)) && break
    sleep 5
done


kubectl apply -f $WORKING_DIR/fybrikapplication-$moduleVersion.yaml

c=0
while [[ $(kubectl get fybrikapplication my-notebook -o 'jsonpath={.status.ready}') != "true" ]]
do
    echo "waiting"
    ((c++)) && ((c==30)) && break
    sleep 6
done

kubectl get pods -n fybrik-blueprints


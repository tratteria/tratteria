# Applying the Alpha Stocks Configurations
echo "\nApplying the Dex Configurations...\n"

kubectl create configmap dex-config --from-file=configs/dex-config.yaml --namespace dex-ns
kubectl apply -f deployments/
kubectl apply -f services/

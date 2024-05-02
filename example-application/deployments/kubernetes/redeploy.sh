chmod +x setup.sh
source setup.sh

echo "\n\n\nRedeploying Alpha Stocks...\n\n\n"

cd alpha-stocks
envsubst < ./deployments/stocks-deployment.yaml | kubectl apply -f -
envsubst < ./deployments/order-deployment.yaml | kubectl apply -f -
kubectl apply -f deployments/gateway-deployment.yaml
cd ..

echo "\n\n\nRedeploying Tratteria...\n\n\n"

cd tratteria
envsubst < ./deployments/txn-token-deployment.yaml | kubectl apply -f -
cd ..

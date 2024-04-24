# Building Alpha Stocks images
echo "\nBuilding Alpha Stocks images...\n"

# Building Gateway Image
docker build -t gateway:latest -f ../../../gateway/Dockerfile ../../../gateway

# Building Stocks Image
docker build -t stocks:latest -f ../../../stocks/Dockerfile ../../../stocks

# Building Order Image
docker build -t order:latest -f ../../../order/Dockerfile ../../../order

# Applying the Alpha Stocks Configurations
echo "\nApplying the Alpha Stocks Configurations...\n"

kubectl apply -f volumes/
kubectl apply -f service-accounts/

envsubst < ./deployments/stocks-deployment.yaml | kubectl apply -f -
envsubst < ./deployments/order-deployment.yaml | kubectl apply -f -
kubectl apply -f deployments/gateway-deployment.yaml

kubectl apply -f services/

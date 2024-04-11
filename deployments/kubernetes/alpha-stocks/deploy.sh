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

kubectl apply -f namespace/
kubectl apply -f volumes/
kubectl create configmap dex-config --from-file=configs/dex-config.yaml --namespace alpha-stocks-dev
kubectl apply -f service-accounts/
kubectl apply -f deployments/
kubectl apply -f services/

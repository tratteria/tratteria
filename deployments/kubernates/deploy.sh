# Building Gateway Image
docker build -t gateway:latest -f ../../gateway/Dockerfile ../../gateway

# Building Stocks Image
docker build -t stocks:latest -f ../../stocks/Dockerfile ../../stocks

# Building Order Image
docker build -t order:latest -f ../../order/Dockerfile ../../order

# Applying the Configuration
kubectl apply -f deployments/
kubectl apply -f services/
kubectl apply -f volumes/

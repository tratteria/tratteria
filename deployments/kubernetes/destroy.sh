kubectl delete configmap dex-config

# Destroying the Configuration
kubectl delete -f deployments/
kubectl delete -f services/
kubectl delete -f volumes/

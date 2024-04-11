# Applying Spire Configurations
echo "\nApplying Spire Configurations...\n"

kubectl apply -f namespace/

# Create Server Bundle Configmap, Role & ClusterRoleBinding
kubectl apply \
    -f server/server-account.yaml \
    -f server/spire-bundle-configmap.yaml \
    -f server/server-cluster-role.yaml

# Create Server Configmap
kubectl apply \
    -f server/server-configmap.yaml \
    -f server/server-statefulset.yaml \
    -f server/server-service.yaml

# Configure and deploy the SPIRE Agent
kubectl apply \
    -f agent/agent-account.yaml \
    -f agent/agent-cluster-role.yaml

kubectl apply \
    -f agent/agent-configmap.yaml \
    -f agent/agent-daemonset.yaml

# Registering Workloads
echo "\nRegistering Workloads...\n"

NAMESPACE=spire
POD_NAME=spire-server-0

echo "Waiting for spire server to be ready..."
while true; do
    POD_STATUS=$(kubectl get pod $POD_NAME -n $NAMESPACE -o jsonpath='{.status.phase}')
    READY=$(kubectl get pod $POD_NAME -n $NAMESPACE -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')
    if [[ "$POD_STATUS" == "Running" && "$READY" == "True" ]]; then
        echo "Spire server is ready.\n"
        break
    else
        echo "Waiting for spire server to be ready..."
        sleep 5
    fi
done

kubectl exec -n spire spire-server-0 -- \
    /opt/spire/bin/spire-server entry create \
    -spiffeID spiffe://alphastocks.org/ns/spire/sa/spire-agent \
    -selector k8s_sat:cluster:docker-desktop \
    -selector k8s_sat:agent_ns:spire \
    -selector k8s_sat:agent_sa:spire-agent \
    -node

kubectl exec -n spire spire-server-0 -- \
    /opt/spire/bin/spire-server entry create \
    -spiffeID spiffe://alphastocks.org/stocks \
    -parentID spiffe://alphastocks.org/ns/spire/sa/spire-agent \
    -selector k8s:ns:alpha-stocks-dev \
    -selector k8s:sa:stocks-service-account

kubectl exec -n spire spire-server-0 -- \
    /opt/spire/bin/spire-server entry create \
    -spiffeID spiffe://alphastocks.org/gateway \
    -parentID spiffe://alphastocks.org/ns/spire/sa/spire-agent \
    -selector k8s:ns:alpha-stocks-dev \
    -selector k8s:sa:gateway-service-account

kubectl exec -n spire spire-server-0 -- \
    /opt/spire/bin/spire-server entry create \
    -spiffeID spiffe://alphastocks.org/order \
    -parentID spiffe://alphastocks.org/ns/spire/sa/spire-agent \
    -selector k8s:ns:alpha-stocks-dev \
    -selector k8s:sa:order-service-account

# Kubernetes Configuration for tokenetes

This directory contains the necessary Kubernetes YAML files to deploy Tokenetes. Below are the instructions on how to deploy Tokenetes on Kubernetes:

## Files Included

- `deployment.yaml`: Defines the Kubernetes Deployment for Tokenetes.
- `service.yaml`: Defines the Kubernetes Service for Tokenetes.
- `service-account.yaml`: Sets up a Kubernetes Service Account for Tokenetes.


## Prerequisites

Before deploying, ensure [tconfigd](https://github.com/tokenetes/tconfigd) is installed in your Kubernetes cluster. If not, follow the [tconfigd installation instructions](https://github.com/tokenetes/tconfigd/tree/main/installation) to install it in your Kubernetes cluster.

## Configuration Adjustments

Before deploying the service, you need to adjust the YAML files to match your environment. Make the below changes:

### Namespace
- Replace `[your-namespace]` and `[your-trust-domain]` in all YAML files with your Kubernetes namespace and your trust domain.

### SPIRE Agent Host Path
- Update the `path` in the `spire-agent-socket` volume definition within `deployment.yaml` to match the location of the SPIRE agent socket in your environment.

```yaml
volumes:
  - name: spire-agent-socket
    hostPath:
      path: /run/spire/sockets # Host directory where the SPIRE agent's socket resides; update this if different in your environment
      type: Directory
```

## Deploying tokenetes

```bash
kubectl apply -f service-account.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

## Verifying Deployment

```bash
kubectl get deployments,svc -n [your-namespace]
```

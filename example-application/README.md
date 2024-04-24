# Alpha Stocks

Alpha Stocks is a sample application that implements transaction tokens(TraTs) using Tratteria. It runs on a Kubernetes cluster and can serve as a reference for integrating Tratteria into other projects. Details on deploying Tratteria can be found in the `deployment\kubernetes\tratteria` directory.

## How to Run

## Backend

Ensure Kubernetes is installed and correctly configured on your machine before executing these commands. 

- Navigate to the deployments/kubernetes directory and run:

```bash
./deploy.sh
```

- You can remove all generated Kubernetes resources using the command below:

```bash
./destroy.sh
```

- If you have made changes to the code, you can re-execute `./deploy.sh` to apply those changes.

### OIDC Authentication via Dex

The application uses Dex as its OIDC provider, configured at `deployment/kubernetes/configs/dex-config.yaml`. If you need to add clients, update secrets, or manage users, please update this file as necessary.

### SPIRE Identity Management

The application incorporates SPIRE(the SPIFFE Runtime Environment) for workload identity management, with configurations located at `deployment/kubernates/spire/`. To adjust service identities, modify configurations, or manage workload registrations, please refer to and update the appropriate files within the directory.


## Client(Frontend)

- To start the client, navigate to the frontend directory and run:

```bash
npm start
```

- If you have made changes to the code, the changes are automatically applied and reflected in the client.

For more detailed instructions refer to the service-specific README files in their respective directories.

Contributions to the project are welcome, including feature enhancements, bug fixes, and documentation improvements.
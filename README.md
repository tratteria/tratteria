# TraTs-Demo-Svcs

TraTs-Demo-Svcs is a collection of sample services designed to demonstrate the effectiveness of the security measures implemented via Transaction Tokens (TraTs). The project showcases how individual microservices components can interact securely via TraTs.

## How to Run

### Backend

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

### Client(Frontend)

- To start the client, navigate to the frontend directory and run:

```bash
npm start
```

- If you have made changes to the code, the changes are automatically applied and reflected in the client.

For more detailed instructions refer to the service-specific README files in their respective directories.

Contributions to the project are welcome, including feature enhancements, bug fixes, and documentation improvements.
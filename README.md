# TraTs-Demo-Svcs

TraTs-Demo-Svcs is a collection of sample services designed to demonstrate the effectiveness of the security measures implemented via Transaction Tokens (TraTs). The project showcases how individual microservices components can interact securely via TraTs.

## How to Run

### Backend

You can either use Kubernetes or Docker Compose for the deployment. 

#### Using Kubernates
Ensure Kubernetes is installed and correctly configured on your machine before executing these commands. 

- Navigate to the deployments/kubernates directory and run:

```bash
./deploy.sh
```

- You can remove all generated Kubernetes resources using the command below:

```bash
./destroy.sh
```

- If you have made changes to the code, you can re-execute `./deploy.sh` to apply those changes.



#### Using Docker Compose

Ensure Docker is installed and running on your machine before executing these commands. 

- Navigate to the deployments/docker-compose directory and execute:

```bash
docker compose up
```

- To remove the generated resources, use:

```bash
docker compose down
```

- If you have made changes to the code and wish to launch the latest version, rebuild the images and containers using:

```bash
docker compose up --build
```

### Client(Frontend)

- To start the client, navigate to the frontend directory and run:

```bash
npm start
```

- If you have made changes to the code, the changes are automatically applied and reflected in the client.

For more detailed instructions refer to the service-specific README files in their respective directories.

Contributions to the project are welcome, including feature enhancements, bug fixes, and documentation improvements.
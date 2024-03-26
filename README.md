# TraTs-Demo-Svcs

TraTs-Demo-Svcs is a collection of sample services designed to demonstrate the effectiveness of the security measures implemented via Transaction Tokens (TraTs). The project showcases how individual microservices components can interact securely via TraTs.

## How to Run

To run the entire suite of services, you can use Docker Compose. This will set up the network, volumes, and service containers required for the project to run.

- To start the services, run:

```bash
docker compose up
```

- If you have made changes to the code and want to launch the latest version, rebuild the images and containers with:

```bash
docker compose up --build
```

Ensure Docker is installed and running on your machine before executing these commands. For more detailed instructions refer to the service-specific README files in their respective directories.

Contributions to the project are welcome, including feature enhancements, bug fixes, and documentation improvements.
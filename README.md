
# Instruction

Here's a README file for your `EthLoadBalancer` project. It includes an overview, setup instructions, usage examples, and some basic information to help others understand and use your project.

```markdown
# EthLoadBalancer

EthLoadBalancer is a distributed load balancing system written in Go, designed to handle Ethereum block data requests across multiple nodes. It simulates resource constraints (CPU and memory) and proxies requests to other nodes when a single node is overloaded, ensuring efficient request distribution in a cluster.

## Features

- **Load Balancing**: Distributes Ethereum block data requests across multiple nodes based on resource availability.
- **Resource Management**: Enforces artificial CPU and memory limits per node.
- **Proxying**: Forwards requests to peer nodes when local resources are exhausted.
- **Health Monitoring**: Provides a `/health` endpoint to check node status.
- **Request Tracing**: Tracks request chains to show the path through nodes.
- **Dockerized**: Easily deployable with Docker Compose for multi-node testing.

## Prerequisites

- [Go](https://golang.org/dl/) (1.20 or later) - for building the application
- [Docker](https://www.docker.com/get-started) - for running the multi-node setup
- [Docker Compose](https://docs.docker.com/compose/install/) - for orchestrating multiple nodes

## Installation

1. **Clone the Repository**
   ```bash
   git clone https://github.com/yourusername/EthLoadBalancer.git
   cd EthLoadBalancer
   ```

2. **Build the Application**
   If running locally without Docker:

   ```bash
   go mod init EthLoadBalancer
   go get github.com/gorilla/mux
   go build -o ethloadbalancer
   ```

3. **Run with Docker Compose**
   The project includes a `docker-compose.yml` file to set up 4 nodes:

   ```bash
   docker-compose up --build
   ```

## Configuration

The application uses environment variables for configuration, set in `docker-compose.yml` or your environment:

- `PORT`: HTTP server port (default: 8080)
- `MAX_CPU`: Maximum concurrent requests (default: 2)
- `MAX_MEMORY`: Maximum memory in MB (default: 200)
- `MEMORY_PER_REQUEST`: Memory per request in MB (default: 50)
- `PEER_NODES`: Comma-separated list of peer node URLs (e.g., `http://node2:8080,http://node3:8080`)

## Usage

### Endpoints

- **GET /eth/block**: Retrieve dummy Ethereum block data
  - Returns JSON with block number, timestamp, hash, and request chain
- **GET /health**: Check node status
  - Returns JSON with resource usage and status ("ready" or "busy")

### Testing

1. **Single Request**

   ```bash
   curl http://localhost:8081/eth/block
   ```

2. **Stress Test**
   Send multiple concurrent requests to node1:

   ```bash
   for i in {1..20}; do curl http://localhost:8081/eth/block & done
   ```

3. **View Logs**
   Check the Docker Compose logs to see request handling and proxying:

   ```bash
   docker-compose logs -f
   ```

   Look for:
   - ✅: Request handled directly
   - 🔀: Request proxied to another node
   - ❌: Request rejected (hop limit or loop)

### Example Output

```json
{
  "node": "node1",
  "chain": "node1 → node2",
  "block": {
    "blockNumber": 123456,
    "timestamp": 1742825761,
    "hash": "0x0000000000000000a86c8de588b24567"
  },
  "message": "Handled directly"
}
```

## Project Structure

```bash
EthLoadBalancer/
├── main.go          # Main application code
├── Dockerfile       # Docker configuration
├── docker-compose.yml # Multi-node setup
└── README.md        # This file
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -m "Add your feature"`)
4. Push to the branch (`git push origin feature/your-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Go](https://golang.org/)
- Uses [Gorilla Mux](https://github.com/gorilla/mux) for routing
- Inspired by distributed system concepts for blockchain data handling

```bash

### Notes
- Replace `yourusername` with your actual GitHub username in the clone URL.
- If you want to add a `LICENSE` file, you can create one with the MIT License text or your preferred license.
- You might want to adjust the port numbers or configuration values if you modified them in your implementation.
- Add any additional sections (e.g., "Troubleshooting" or "Future Improvements") as needed.

To use this:
1. Save it as `README.md` in your project root
2. Commit it to your GitHub repository:
```bash
git add README.md
git commit -m "Add README"
git push origin main
```

# OpenShield Agent

OpenShield Agent is a cross-platform security agent designed to communicate with the OpenShield Manager for remote task execution, script management, and system monitoring.

---

## Features

- **Remote Task Execution:** Securely receive and execute commands from the manager.
- **Script Synchronization:** Sync, update, and manage scripts from the manager.
- **Heartbeat:** Periodically report agent status and network addresses to the manager.
- **Credential Security:** Stores credentials securely using OS keyring with file fallback.
- **Cross-Platform:** Works on Windows and Linux.

---

## Prerequisites

- **Go 1.20+**
- **osquery** must be installed on the target machine.

### Install osquery

#### Linux (Debian/Ubuntu/Kali)
```sh
sudo apt-get update
sudo apt-get install osquery
```

#### RedHat/CentOS/Fedora
```sh
sudo yum install osquery
```

#### Windows
- Download and install from [osquery.io/downloads](https://osquery.io/downloads)
- Or use Chocolatey:
  ```sh
  choco install osquery
  ```

---

## Configuration

Edit `config/config.yml` to set your manager address and ports:

```yaml
MANAGER_ADDRESS: localhost
MANAGER_API_PORT: 9000
MANAGER_GRPC_PORT: 50052
COMMAND_TIMEOUT: 60
```

---

## Usage

1. **Build the agent:**
   ```sh
   go build -o openshield-agent .
   ```

2. **Run the agent:**
   ```sh
   ./openshield-agent
   ```

   > On Windows, run as administrator if you need to install or manage osquery.

3. **Agent will:**
   - Check for `osqueryi` in your PATH.
   - Register with the manager and securely store credentials.
   - Start the gRPC server and listen for tasks/scripts.
   - Periodically send heartbeats to the manager.

---

## Development

- Protobuf definitions are in `proto/rpc.proto`.
- Main logic is in `internal/`.
- To regenerate gRPC code:
  ```sh
  protoc --go_out=. --go-grpc_out=. proto/rpc.proto
  ```

---

## Security

- Credentials are stored in the OS keyring when possible, with a fallback to a local file (`config/agent_credentials.json`, permissions `0600`).
- To clear credentials and re-register, use the provided utility or delete the keyring entries and credentials file.

---

## License

MIT License

---

## Links

- [osquery Documentation](https://osquery.io/docs/)

# OpenShield Agent

OpenShield Agent is a cross-platform security agent designed to communicate with the OpenShield Manager for remote task execution, script management, and system monitoring.

---

## Features

- **Remote Task Execution:** Securely receive and execute commands from the manager.
- **Script Synchronization:** Sync, update, and manage scripts from the manager.
- **Heartbeat:** Periodically report agent status and network addresses to the manager.
- **Credential Security:** Stores credentials securely using OS keyring with file fallback.
- **Cross-Platform:** Works on Windows, Linux, and macOS.

---

## Prerequisites

- **Go 1.20+** (for building from source)
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

## Installation

### Automated Install Script (Linux)

You can use the provided install script to automatically download the latest (or a specific) release, configure the manager address, and set up the agent as a systemd service:

```sh
curl -O https://raw.githubusercontent.com/YordanOpenShield/openshield-agent/main/helpers/install.sh
chmod +x install.sh
sudo ./install.sh latest <MANAGER_ADDRESS>
```
Replace `<MANAGER_ADDRESS>` with the address of your manager (e.g., `192.168.10.11`).

- The script will:
  - Download the latest agent binary and systemd unit file.
  - Install the agent to `/usr/local/bin`.
  - Configure and enable the agent as a systemd service.
  - Start the agent automatically.

**To install a specific version:**
```sh
sudo ./install.sh v1.0.1 <MANAGER_ADDRESS>
```

---

## Configuration

Edit `config/config.yml` to set your manager address and ports if needed:

```yaml
MANAGER_ADDRESS: localhost
MANAGER_API_PORT: 9000
MANAGER_GRPC_PORT: 50052
COMMAND_TIMEOUT: 60
```

---

## Usage

- The agent will start automatically as a systemd service after installation.
- To check the status:
  ```sh
  sudo systemctl status openshield-agent
  ```
- To view logs:
  ```sh
  journalctl -u openshield-agent -f
  ```

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
- [OpenShield Manager](https://github.com/YordanOpenShield/openshield-manager)

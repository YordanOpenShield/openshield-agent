# Enrollment

This document describes the step-by-step process for enrolling an OpenShield Agent with the Manager.

---

## 1. Prerequisites

- Agent binary is installed on the target machine.
- Manager address and ports are known.
- `osquery` is installed and available in PATH. (optional)

---

## 2. Enrollment Steps

### Step 1: Start the Agent (User)

- The agent is started manually or as a systemd service.
- Manual example:
  ```sh
  ./openshield-agent -manager <MANAGER_ADDRESS> -config <CONFIG_PATH> -scripts <SCRIPTS_PATH>
  ```
- `systemd` example:
  ```
  systemctl enable openshield-agent.service
  systemctl start openshield-agent.service
  ```

### Step 2: Generate or Load Device ID (Agent)

- The agent checks for an existing device ID.
- If not found, it generates a new unique device ID.

### Step 3: Register with Manager (Agent)

- The agent sends a registration request to the manager via gRPC:
  - Includes device ID and other metadata.
- Manager responds with:
  - Agent ID
  - Agent token (for authentication)

### Step 4: Store Credentials (Agent)

- Agent securely stores the received credentials:
  - In OS keyring (preferred)
  - As a fallback, in a local file with restricted permissions

### Step 5: Confirm Enrollment (Agent)

- Agent logs successful enrollment.
- Agent is now ready to receive tasks and send heartbeats.

---

## 3. Error Handling

- If registration fails, the agent retries after a delay.
- If credentials cannot be stored, the agent logs an error and exits.

---

## 4. Re-enrollment

- To re-enroll, clear credentials using the provided utility or manually remove keyring entries and credential files.
- Restart the agent to trigger a new enrollment.

---

## 5. Sequence Diagram

```mermaid
sequenceDiagram
    participant Agent
    participant Manager

    Agent->>Manager: RegisterAgent(device_id)
    Manager-->>Agent: RegisterAgentResponse(agent_id, token)
    Agent->>Agent: Store credentials
    Agent->>Manager: Heartbeat (authenticated)
```

---
# openshield-agent

## Prerequisites

**osquery must be installed on the target machine before running the agent.**

### Linux (Debian/Ubuntu/Kali)

```sh
sudo apt-get update
sudo apt-get install osquery
```

### RedHat/CentOS/Fedora

```sh
sudo yum install osquery
```

### Windows

- Download and install the official MSI from [osquery.io/downloads](https://osquery.io/downloads)
- Or use Chocolatey:
  ```sh
  choco install osquery
  ```

## Agent Startup

The agent will check for `osqueryi` in your system PATH at startup.  
If it is not found, the agent will exit with an error.

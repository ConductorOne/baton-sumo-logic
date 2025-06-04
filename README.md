![Baton Logo](./baton-logo.png)

# `baton-sumo-logic` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-sumo-logic.svg)](https://pkg.go.dev/github.com/conductorone/baton-sumo-logic) ![main ci](https://github.com/conductorone/baton-sumo-logic/actions/workflows/main.yaml/badge.svg)

`baton-sumo-logic` is a connector for built using the [Baton SDK](https://github.com/conductorone/baton-sdk).

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

# Getting Started

## Configuration

This connector requires the following configuration:

- `api-base-url`: The Sumo Logic API base URL (default: "https://api.sumologic.com")
- `api-access-id`: The Sumo Logic API access ID
- `api-access-key`: The Sumo Logic API access key
- `include-service-accounts`: Whether to include service accounts (default: true)

You can provide these values as environment variables:

```bash
export BATON_API_BASE_URL=https://api.sumologic.com
export BATON_API_ACCESS_ID=your-access-id
export BATON_API_ACCESS_KEY=your-access-key
export BATON_INCLUDE_SERVICE_ACCOUNTS=true
```

## Installation Options

### Homebrew

```bash
brew install conductorone/baton/baton conductorone/baton/baton-sumo-logic
baton-sumo-logic
baton resources
```

### Docker

```bash
docker run --rm -v $(pwd):/out \
  -e BATON_API_BASE_URL=https://api.sumologic.com \
  -e BATON_API_ACCESS_ID=your-access-id \
  -e BATON_API_ACCESS_KEY=your-access-key \
  ghcr.io/conductorone/baton-sumo-logic:latest -f "/out/sync.c1z"

docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

### From Source

```bash
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-sumo-logic/cmd/baton-sumo-logic@main

baton-sumo-logic
baton resources
```

## Data Model and Capabilities

`baton-sumo-logic` provides the following capabilities:

### Resource Sync
- Users (both human accounts and service accounts)
- Roles

### Provisioning Capabilities
- User account management (create and delete)
- Role assignments (grant and revoke role memberships)

Note: Service account syncing can be optionally disabled using the `include-service-accounts` configuration parameter.

## Contributing, Support, and Issues

We started Baton because we were tired of taking screenshots and manually
building spreadsheets. We welcome contributions, and ideas, no matter how
small&mdash;our goal is to make identity and permissions sprawl less painful for
everyone. If you have questions, problems, or ideas: Please open a GitHub Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

## `baton-sumo-logic` Command Line Usage

```bash
baton-sumo-logic

Usage:
  baton-sumo-logic [flags]
  baton-sumo-logic [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --api-base-url string          The Sumo Logic API base URL ($BATON_API_BASE_URL) (default "https://api.sumologic.com")
      --api-access-id string         The Sumo Logic API access ID ($BATON_API_ACCESS_ID)
      --api-access-key string        The Sumo Logic API access key ($BATON_API_ACCESS_KEY)
      --include-service-accounts     Whether to include service accounts ($BATON_INCLUDE_SERVICE_ACCOUNTS) (default true)
      --client-id string             The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string         The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
  -f, --file string                  The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                         help for baton-sumo-logic
      --log-format string            The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string             The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
  -p, --provisioning                 If this connector supports provisioning, this must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --ticketing                    This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                      version for baton-sumo-logic

Use "baton-sumo-logic [command] --help" for more information about a command.
```

## Important Notes

- Ensure the user creating the access key has the "Create Access Keys" and "Manage Access Keys" role capabilities, and possesses sufficient permissions matching the intended use of the key, such as "Manage Users and Roles" or an Administrator role.
- Access keys cannot exceed the permissions of their creator.
- Copy the Access ID and Access Key immediately after creation, as they are displayed only once.
- The "Manage Users and Roles" permission is required for both operations: sync (read-only) and provisioning (read-write). This single permission grants access to both functionalities.

## Additional Resources

- [Sumo Logic Access Keys Documentation](https://help.sumologic.com/docs/manage/security/access-keys/)
- [API Authentication and Endpoints](https://help.sumologic.com/docs/api/getting-started/)


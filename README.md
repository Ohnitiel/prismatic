# Prismatic

Prismatic is a tool for executing SQL queries in parallel across multiple PostgreSQL databases.  

It was built to solve a practical problem: running the same query across 40+ databases and safely exporting or validating results without manually repeating the process.  

Designed for multi-tenant and multi-environment setups.  

## Table of Contents

- [Why Prismatic?](#why-prismatic)
- [Core Features](#core-features)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Export Example](#example-exporting-data-from-all-configured-connections)
  - [Safe Update Example](#example-running-a-safe-update)
- [Architecture Overview](#architecture-overview)
- [Project Structure](#project-structure)
- [Design Principles](#design-principles)
- [Roadmap](#roadmap)
- [License](#license)

## Why Prismatic?

* Execute the same query across dozens of databases simultaneously
* Safe by default (automatic transaction rollback unless committed)
* Configurable worker pool for controlled concurrency

## Core Features

* Parallel Execution across environments and connections
* Transactional Safety (rollback by default)
* Excel Export (single or multiple sheets/files)
* TOML-based Configuration
* Internationalized CLI (EN / PT-BR)

## Configuration
Initial configuration can be done by running `prismatic config install`.

### Default Configuration

```toml
locale = "en-US"                        # "en-US" or "pt-BR"
max_workers = 5                         # Number of threads
max_retries = 3                         # Connection attempts
timeout = 10                            # Described in seconds
connection_column_name = "connection"   # Column name for Excel export

[paths]
connections = "./config/connections.toml"

# Cache is yet to be implemented
# [cache]
# use_cache = true
# time_to_live = 600 # Described in seconds

# By default, logs to stderr and file
# Will be configurable in the future
[logger]
file_level = "debug"
file_output = "./log/prismatic.log"
console_level = "info"
console_output = "stderr"
```

### Configuration Example
Connections are defined in `config/connections.toml`.

Basic configuration:
```toml
[my_conn]
engine = "postgresql"
username = "admin"
password = "${DB_PASSWORD}"  # Plain password or environment variable (.env accepted)
database = "my_db"

[my_conn.environment.production]
host = "10.0.0.10"
port = 5432
```

Inheritance overrides:
```toml
[my_conn]
engine = "postgresql"
username = "admin"
password = "${DB_PASSWORD}"
database = "my_db"
port = 5432

[my_conn.environment.production]
host = "10.0.0.10"
port = 5433

[my_conn.environment.staging]
host = "10.0.0.11"
username = "staging_admin"
password = "${STAGING_DB_PASSWORD}"
database = "staging_db"
```

## Usage

By default, Prismatic runs the given query in all configured staging environments connections.

### Generic Flags:
```
    --connections, -c   String array of connections to use (eg. "my_conn" or "my_conn,my_other_conn")
    --environment, -e   Environment to use (eg. "production")
    --config            Path to configuration file (default: "./config/config.toml")
```

### Exporting Data

`prismatic export` will export the given query to an Excel file.

#### Export Flags:
```
    --no-single-sheet      Export to multiple sheets in the same workbook
    --no-single-file       Export to multiple files instead of a single sheet
```

#### Example: Exporting Data
```bash
# This saves the results with `connection_column_name` config as header and connections names as values
prismatic export \
  "SELECT id, name FROM patients WHERE active = true" \
  results.xlsx
```

#### Output Behavior

* Executes in parallel
* Rolls back transaction automatically

### Executing a Query

`prismatic run` will execute the given query in all configured connections.

#### Run Flags:
```
    --commit           Persist changes
```

### Example: Running a Safe Update
```bash
prismatic run \
  "UPDATE settings SET value = 'true' WHERE key = 'maintenance'"
```

By default, changes are rolled back.

To persist:
```bash
prismatic run \
  "UPDATE settings SET value = 'true' WHERE key = 'maintenance'" \
  --commit
```

### Architecture Overview
```
CLI
  ↓
Core Engine
  ↓
Worker Pool (configurable)
  ↓
Database (PostgreSQL / SQLite)
  ↓
Result Aggregator
  ↓
Excel Exporter
```

## Project Structure
```
cmd/cli/          → CLI entry point
internal/config/  → Configuration loading
internal/db/      → Connection and execution logic
internal/export/  → Excel export
internal/locale/  → Localization support
internal/logger/  → Structured logging
```

## Design Principles

* Local-first
* No remote services
* No hidden state
* Explicit commits
* Predictable concurrency

## Roadmap

* [ ] Add config command for creating/showing/editing config files
* [ ] Schema metadata caching by version
* [ ] Backend-driven SQL autocomplete
* [ ] Local desktop UI
* [ ] Multi-connection result comparison

## License

MIT

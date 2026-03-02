# Prismatic

Prismatic is a tool for executing SQL queries in parallel across multiple PostgreSQL databases.

It was built to solve a practical problem: running the same query across 40+ databases and safely exporting or validating results without manually repeating the process.

Designed for multi-tenant and multi-environment setups.

## Features

- **Parallel Execution** — run the same query across dozens of databases simultaneously
- **Transactional Safety** — changes are rolled back by default; commits are always explicit
- **Configurable Concurrency** — tune the worker pool to control how many connections run at once
- **Excel Export** — export results to a single sheet, multiple sheets, or multiple files
- **TOML-based Configuration** — simple, readable config with environment variable support
- **Internationalized CLI** — available in English and Brazilian Portuguese (PT-BR)

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

[logger]
file_level = "debug"
file_output = "./log/prismatic.log"
console_level = "info"
console_output = "stderr"
```

### Connections

Connections are defined in `config/connections.toml`. Each connection supports multiple environments, and environment-level values override the base connection values.

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

With environment-level overrides:

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

By default, Prismatic runs the given query across all configured connections in the staging environment.

### Generic Flags

```
    --connections, -c   String array of connections to use (e.g. "my_conn" or "my_conn,my_other_conn")
    --environment, -e   Environment to use (e.g. "production")
    --config            Path to configuration file (default: "./config/config.toml")
```

### Exporting Data

`prismatic export` runs the given query and exports results to an Excel file.

```
    --no-single-sheet      Export to multiple sheets in the same workbook
    --no-single-file       Export to multiple files instead of a single sheet
```

```bash
# Results include a column (named by connection_column_name) identifying which connection each row came from
prismatic export \
  "SELECT id, name FROM patients WHERE active = true" \
  results.xlsx
```

### Running a Query

`prismatic run` executes the given query across all configured connections. Changes are rolled back by default — use `--commit` to persist them.

```
    --commit           Persist changes
```

```bash
# Dry run (safe by default — changes are rolled back)
prismatic run \
  "UPDATE settings SET value = 'true' WHERE key = 'maintenance'"

# Persist changes
prismatic run \
  "UPDATE settings SET value = 'true' WHERE key = 'maintenance'" \
  --commit
```

## Architecture

Prismatic follows a linear pipeline from CLI input to result output:

```
CLI
  ↓
Core Engine
  ↓
Worker Pool (configurable)
  ↓
Database (PostgreSQL)
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

- Local-first
- No remote services
- No hidden state
- Explicit commits
- Predictable concurrency

## Roadmap

- Cross-environment query execution
- Local desktop UI
- Backend-driven SQL autocomplete
- Schema metadata caching by version

## License

[MIT](./LICENSE)

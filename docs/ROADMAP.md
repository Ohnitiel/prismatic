# Prismatic Development Roadmap

This document outlines the steps to transition `Prismatic` from a basic CLI tool into a robust, portfolio-grade system application, mirroring and eventually replacing the functionality of `python-db-tools`.

## Phase 1: Architecture & Core Foundations

### 1. Define a Database Interface (The "Developer" Approach)
*   **Goal:** Create a `Database` interface with methods like `Connect()`, `Ping()`, and `Query()`.
*   **Why:** Decouples business logic from specific drivers (Postgres, SQLite). Enables easy unit testing and mocking. Demonstrates understanding of idiomatic Go and Clean Architecture.

### 2. Add SQLite Support (Feature Parity)
*   **Goal:** Implement the `Database` interface using a pure Go SQLite driver (e.g., `modernc.org/sqlite` or `mattn/go-sqlite3`).
*   **Why:** Achieves parity with the Python tool. Proves the Interface design works with multiple distinct SQL dialects.

## Phase 2: Data Engineering & Usability

### 3. Implement Excel Export
*   **Goal:** Use the `xuri/excelize` library to write query results to `.xlsx` files.
*   **Why:** Critical feature for business users. Demonstrates efficient data handling in Go (streaming large datasets) compared to Python's memory-heavy `pandas`.

## Future Phases (To Be Defined)
*   MySQL / SQL Server Support.
*   Concurrent/Parallel Query Execution (WaitGroups).
*   Configuration Management Refinement.
*   Cross-Platform Build Pipeline.

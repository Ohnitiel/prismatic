# Essential Resources for Solo Go Development

If you choose to tackle `Prismatic` solo, these are the high-quality, canonical resources you should rely on. Avoid random tutorials; stick to these "sources of truth."

## 1. The Core Language (Syntax & Idioms)
*   **[A Tour of Go](https://go.dev/tour/):** The absolute first stop. Interactive and covers the syntax basics.
*   **[Effective Go](https://go.dev/doc/effective_go):** *Critical reading.* This document explains *how* to write Go that looks like Go, not translated Python. It covers formatting, naming, and control structures.
*   **[Go by Example](https://gobyexample.com/):** Excellent for quick syntax reference (e.g., "How do I do a `for` loop again?").

## 2. Project Structure & Design
*   **[Standard Go Project Layout](https://github.com/golang-standards/project-layout):** The de-facto standard for where to put folders like `cmd`, `internal`, and `pkg`.
*   **[Go Style Guide (Uber)](https://github.com/uber-go/guide/blob/master/style.md):** A very popular, strict style guide used by many companies. Good for learning professional habits.

## 3. Interfaces & Systems Programming
*   **[Jordon Orelli: How to use Interfaces in Go](https://jordanorelli.com/post/32665860244/how-to-use-interfaces-in-go):** A classic article that demystifies interfaces better than most docs.
*   **[Go Data Structures: Interfaces](https://research.swtch.com/interfaces):** A deep dive by Russ Cox (Go Core Team) on how interfaces work under the hood.

## 4. Databases (`database/sql`)
*   **[Go database/sql Tutorial](http://go-database-sql.org/):** The best community guide for the standard library's SQL package. It covers connection pools, prepared statements, and handling nulls—all things you'll need for Prismatic.
*   **[SQLX](https://github.com/jmoiron/sqlx):** (Optional) A popular library that extends `database/sql` to make struct scanning easier (like `pd.read_sql` but lower level). You might want to implement the "raw" way first, then switch to this.

## 5. CLI Development
*   **[CLI Code (Urfave)](https://cli.urfave.org/):** Documentation for the library you are already using. Read specifically about "Subcommands" and "Flags."
*   **[12 Factor CLI Apps](https://medium.com/@jdxcode/12-factor-cli-apps-dd3c227a0e46):** Principles for building good CLI tools (stdout/stderr usage, config, etc.).

## 6. Testing
*   **[Learn Go with Tests](https://quii.gitbook.io/learn-go-with-tests/):** An incredible resource that teaches TDD (Test Driven Development) from scratch. Highly recommended.

## Your "Solo" Workflow
1.  **Read** the relevant section in *Effective Go*.
2.  **Search** *pkg.go.dev* for the specific library docs (e.g., `pgx`, `excelize`).
3.  **Implement** a small prototype.
4.  **Refactor** based on what you find in *Go by Example*.

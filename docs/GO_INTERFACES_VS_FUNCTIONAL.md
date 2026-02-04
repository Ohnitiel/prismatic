# Why Interfaces? (Idiomatic Go vs Functional)

When building systems in Go, especially for portfolio projects aimed at demonstrating "Senior" or "Engineer" level skills, choosing the right paradigm is crucial. While Go supports functional patterns (closures, higher-order functions), **Interfaces** are the standard for managing IO and stateful resources like databases.

## 1. "Idiomatic" Go (The #1 Portfolio Criterion)
Go is an opinionated language. Hiring managers look for code that follows "The Go Way."
*   **Go is Structural, not Functional:** The language is designed around `structs` (data) and `interfaces` (behavior).
*   **The Standard Library Pattern:** Look at Go's `database/sql`. It doesn't give you a `query(db, sql)` function; it gives you a `DB` struct with methods like `db.Query()`. Mimicking the standard library shows you understand the ecosystem.

## 2. The "State" Problem
Database connections are **stateful**. They hold open socket connections, transaction buffers, and mutexes.
*   **Functional Approach:** You have to pass this state around constantly as an argument (`func Query(conn *Connection, q string)`), or hide it inside complex closures.
*   **Interface Approach:** You bundle the state (the connection) with the behavior (the methods).
    ```go
    // Clean and simple for the caller
    myDB.Query("SELECT * ...")
    ```

## 3. Extensibility (The "Clean Code" Factor)
As the application grows, requirements expand beyond simple queries. You will eventually need:
*   `Close()` (to shut down connections)
*   `Ping()` (for health checks)
*   `BeginTx()` (for transactions)

*   **Functional:** You would need to return a "bag of functions" or a tuple of closures, which becomes difficult to maintain.
*   **Interface:** You define a contract. Any struct that implements methods X, Y, and Z automatically satisfies the interface.

## Recommendation
For a Data Engineer/Developer portfolio, use the **Interface** pattern. It proves you understand:
1.  **Polymorphism:** Handling Postgres, SQLite, and MySQL uniformly.
2.  **Dependency Injection:** Making your code testable (mocking dependencies).
3.  **System Design:** Structuring a scalable application.

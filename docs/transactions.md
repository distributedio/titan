# Titan Transaction Handling Guide

Titan provides ACID transaction support via TiKV. This guide explains how transactions work, retry behavior, and best practices.

## Transaction Types

### 1. Single Command (AutoCommit)
Most Redis commands (e.g. `SET`, `GET`, `INCR`) run in a single transaction.
- **Retry Behavior:** Automatically retries on write conflicts.
- **Configuration:** Configurable exponential backoff and retry limits.

### 2. MULTI/EXEC
Commands queued between `MULTI` and `EXEC` are executed in a single transaction.
- **Retry Behavior:** Automatically retries on write conflicts.
- **Atomicity:** All commands succeed or fail together.

### 3. WATCH/MULTI/EXEC
Optimistic locking using `WATCH`.
- **Retry Behavior:** **DISABLED**. If a watched key is modified, or if a write conflict occurs during commit, the transaction fails immediately with an error (or nil reply for CAS failure).
- **Reason:** Titan cannot automatically retry WATCH transactions because the client-side logic that decided what to write might depend on the initial values of the watched keys. Retrying would require re-running the client logic.
- **Best Practice:** Clients must handle `EXEC` failures and retry the entire WATCH loop.

## Best Practices

### Handle Retries Client-Side
While Titan retries internal conflicts, some failures (like WATCH conflicts or excessive contention) return errors to the client.
- Always check `EXEC` results.
- Implement client-side backoff for retries to avoid thundering herd.

### Transaction Size
- **Timeout:** Transactions have a default timeout (e.g. 30s) to prevent holding locks indefinitely.
- **Size Limit:** Avoid huge transactions (thousands of keys) as they increase conflict probability and memory usage. Large `MSET` or bulk loads should be split.

### Monitoring
- Monitor `titan_txn_conflicts_total` to detect high contention.
- Monitor `titan_txn_failures_total` for persistent errors.
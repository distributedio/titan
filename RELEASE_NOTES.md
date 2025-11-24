# Release Notes - Titan v0.4.0

## Major Changes

- **TiKV Client Upgrade**: Upgraded from legacy `tidb/kv` to the modern `tikv/client-go/v2` library. This provides better stability, performance, and compatibility with newer TiKV clusters.
- **Go Version**: Updated requirement to Go 1.21+.

## Breaking Changes

- **GC Management**: The built-in manual GC trigger mechanism has been disabled due to API changes in the upstream client. Users are advised to rely on PD/TiKV's automatic GC or use standard TiKV tools for GC management.
- **Metrics**: Metrics dependent on the internal TiDB SDK have been removed.

## Improvements

- Improved transaction handling using modern client-go patterns.
- Updated dependencies for better security and performance.
- Simplified storage interface wrapper.

## Bug Fixes

- Fixed build issues with modern Go versions.
- Resolved dependency conflicts with `kvproto` and `etcd` (partially resolved via replacement directives).

## Upgrade Instructions

Please refer to [MIGRATION.md](MIGRATION.md) for detailed upgrade instructions.
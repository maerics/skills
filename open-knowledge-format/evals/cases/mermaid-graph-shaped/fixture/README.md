# pipeline

A three-stage batch pipeline:

1. `Ingest` reads a raw file from disk into bytes.
2. `Transform` parses those bytes into a `Record`.
3. `Load` writes a `Record` to the store.

Ingest's output feeds Transform; Transform's output feeds Load. There is
no other path through the system.

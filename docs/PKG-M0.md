# DBrain PKG-M0 — Contract Freeze

**Status:** WORKING · NON-NORMATIVE  
**Phase:** schema / contract only

## Goal

Freeze the minimum contracts required to build DBrain without prematurely implementing storage/index/runtime complexity.

## Contracts to close

1. Claim identity / demand-driven minting / supersession
2. Authority Map subject + scope registration and routing
3. Minimum hand-owned metadata vs derived metadata
4. Knowledge lifecycle MVP
5. Knowledge source / epistemic taxonomy
6. Context Bundle semantic identity + provenance
7. Context budget classes and fail-closed overflow
8. Prior Conflict Guard structure and boundedness
9. Candidate / ingestion semantics
10. Qualification Concept → 1..N Claims linkage
11. Observed Workspace Composite
12. Recovery delta categories
13. Agent Bootstrap Contract
14. Qualification Candidate Prompt Contract
15. Qualification Evaluator Prompt Contract
16. Task / Implementation-Packet rendering boundary
17. Agent Local State Contract
18. Agent Reattachment / Recovery Contract

## Explicitly out of PKG-M0

- SQLite schema / FTS implementation
- committed search index
- embeddings / vector database
- graph database
- background consolidation
- distributed search/services
- full production CLI
- current Platform Notion→Git migration

## Core invariants

```text
Model Prior != project semantic authority
Context Bundle != semantic authority
Agent-local memory != canonical Brain
Resume State != truth authority
Search/index != truth
Context Ready != Role Qualified != Task Admitted
Qualification != semantic authority != Acceptance Authority != execution authorization
```

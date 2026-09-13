# Authority Map Contract — Working Candidate

## Purpose

The Authority Map is a small, hand-owned routing artifact.

It answers:

```text
WHO owns this subject?
WHERE does authoritative resolution begin?
```

It does not require all truth for a subject to live in one file.

## Required semantics

- subject namespace registration;
- scope-aware owner routing;
- canonical entrypoint or collection;
- optional applicable critical guard claims;
- ambiguity/missing required authority => `UNRESOLVED`.

Workspace rules and repository specializations follow subject/scope routing. Refinement must be explicit and non-contradictory.

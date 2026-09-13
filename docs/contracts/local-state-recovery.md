# Agent Local State & Recovery Contract — Working Candidate

## Agent-local state

Agent-local memory may inform context construction and recovery, but is not canonical semantic truth by default.

Examples:

- scratch notes
- TODO/work notes
- vendor-specific memory
- session summaries
- runtime/session metadata

## Recovery invariant

Resume does not mean trust the previous agent.

Resume means:

```text
reconstruct current state from truth owners
+
use prior local/resume state only as bounded recovery input
```

## Recovery delta categories

- KNOWLEDGE DELTA
- AUTHORITY DELTA
- REPOSITORY DELTA
- EVIDENCE DELTA

Agent-local notes remain quarantined context.

Recovery must prevent split-brain modifying sessions through attempt/generation/lease/fence or an equivalent policy.

# Claim Model Contract — Working Candidate

**Status:** WORKING · NON-NORMATIVE  
**PKG-M0 target:** freeze Claim identity, minting, reference, edit, and supersession semantics before implementation.

## 1. Purpose

A **Document** remains the human-readable canonical narrative container.

A **Claim** is an optional stable semantic address to a bounded semantic statement inside an authoritative document.

A Claim exists only when another consumer needs a durable semantic reference.

Therefore:

```text
Document != Claim collection

Claim minting = demand-driven
Claim minting != coverage-driven extraction
```

A document with zero Claims is valid and expected.

## 2. Claim does not create authority

A Claim is an address, not an authority grant.

Its authority is determined by:

```text
Claim
→ containing canonical source
→ subject + scope
→ Authority Map / declared authority owner
```

A Claim in a source that does not own the relevant subject/scope cannot become authoritative merely because it has a Claim ID.

A Claim in CANDIDATE, agent-local, derived, or otherwise non-authoritative material remains non-authoritative.

## 3. Minting rule

Mint a Claim only when at least one durable consumer requires a stable semantic address.

Valid triggers include:

1. another canonical artifact or cross-repository contract must cite the rule;
2. a Qualification Concept or assessment case must bind to it;
3. the rule belongs to an applicable Prior Conflict Guard set;
4. the rule must be superseded independently from the rest of its containing document;
5. a negative decision/rejected alternative must be discoverable when substantially re-proposed;
6. another accepted contract requires exact semantic binding to the rule.

The following do **not** justify Claim minting by themselves:

- a sentence sounds important;
- a document is long;
- an LLM extracted a possible invariant;
- a heading looks reusable;
- full Claim coverage would make indexing easier.

Tooling may **propose** Claim candidates, but minting is an explicit accepted authoring act.

## 4. Claim identity

A Claim ID MUST be:

- immutable;
- unique across the Development Workspace where cross-repository references may occur;
- independent of file path;
- independent of document title;
- independent of subject name;
- independent of wording;
- safe to preserve across file moves and narrative restructuring.

A Claim ID MUST NOT encode semantic meaning whose later change would require renaming the ID.

The default implementation SHOULD use an opaque globally unique identifier with a Claim namespace/prefix. Exact wire formatting is an implementation/profile decision so long as the properties above hold.

Human-readable labels/titles MAY exist separately from the immutable Claim ID.

## 5. Claim anchor and semantic extent

A Claim MUST resolve to a deterministic bounded semantic span in its canonical source.

For Markdown profiles, the initial candidate representation is:

```text
stable Claim anchor
immediately associated with a heading/section
→ Claim semantic extent = that section subtree
```

The authoritative Claim text remains in the human-readable narrative. The metadata MUST NOT duplicate the full Claim text.

Tooling must be able to validate that:

- every declared Claim ID resolves to exactly one anchor;
- every Claim anchor belongs to exactly one declared Claim;
- duplicate Claim IDs fail validation;
- broken/missing anchors fail validation.

The exact Markdown marker syntax remains an implementation serialization detail until PKG-M0 serialization freeze.

## 6. Claim reference semantics

A Claim reference binds to the exact Claim identity, not merely to similar text or a document path.

Consumers may include:

- canonical documents;
- Implementation / Work Contracts;
- Qualification Concepts;
- Prior Conflict Guard configuration;
- negative-knowledge/prior-art mechanisms;
- derived Context Bundles.

A Claim reference MUST NOT be silently rebound to a replacement Claim after supersession.

If Claim B supersedes Claim A:

```text
old exact reference → Claim A remains historically resolvable
current-use validation → marks the reference stale/review-required where policy requires
human/authority review → explicitly rebinds to Claim B if semantically correct
```

This prevents supersession from silently changing the meaning of accepted downstream artifacts.

## 7. Meaning-preserving edit vs semantic change

The framework does not infer semantic equivalence from text similarity.

The authoring operation declares the intent.

### Meaning-preserving edit

Examples:

- typo correction;
- grammar improvement;
- clarification that does not alter the rule;
- formatting or narrative restructuring.

A meaning-preserving edit MAY retain the same Claim ID.

Its content digest/revision changes.

Dependent artifacts may become `review-suggested` according to policy, but the Claim is not automatically treated as semantically replaced.

### Semantic change

Any change that alters the rule, boundary, permission, prohibition, applicability, obligation, or decision meaning MUST NOT overwrite the existing Claim as if it were the same semantic identity.

It requires a new Claim plus an explicit supersession relationship.

```text
Claim A
→ superseded by
Claim B
```

The old Claim remains resolvable for history/audit.

## 8. Supersession

Supersession is directional and explicit.

The new Claim records which prior Claim(s) it supersedes.

Inverse `superseded_by` views SHOULD be derived rather than hand-maintained.

A superseded Claim:

- is excluded from default current-truth retrieval;
- remains available in explicit historical resolution;
- may invalidate or stale current consumers;
- MUST NOT be physically deleted merely because it is superseded.

### Split

```text
Claim A
→ superseded by Claim B + Claim C
```

### Merge

```text
Claim A + Claim B
→ superseded by Claim C
```

Split/merge MUST NOT automatically rewrite inbound semantic references. Consumers are reviewed/rebound explicitly.

## 9. Claim lifecycle interaction

A Claim's effective currentness depends on both:

- its own supersession/invalidation state;
- the lifecycle/authority state of its containing source.

A current Claim inside a superseded or invalid containing source MUST NOT surface as current authoritative knowledge unless another declared authority rule explicitly preserves it.

PKG-M0 lifecycle vocabulary is handled by the lifecycle contract; this Claim contract only defines the Claim-specific interaction.

## 10. Content digest and provenance

The runtime SHOULD derive a content digest for each resolved Claim semantic span.

The digest is not the Claim identity.

```text
Claim ID
= stable semantic identity

Claim content digest
= observed content/revision identity
```

This distinction enables:

- wording-change detection;
- Context Bundle reproducibility;
- recovery delta reporting;
- qualification review/staleness decisions.

Claim resolution provenance must be sufficient to identify at least:

- Claim ID;
- containing document identity;
- subject/scope;
- source revision;
- Claim content digest;
- authority route used for resolution.

## 11. Candidate and agent-local boundary

CANDIDATE material and agent-local memory MUST NOT silently mint authoritative Claim IDs.

They may have their own candidate/local identifiers.

Promotion into canonical knowledge is an authority/acceptance act that creates or binds the canonical Claim only after the semantic source becomes accepted.

## 12. Failure semantics

The resolver returns `UNRESOLVED` rather than guessing when:

- a required Claim ID cannot be found;
- more than one active Claim has the same immutable ID;
- the Claim anchor is missing or ambiguous;
- the containing source's required authority cannot be established;
- the requested current Claim is superseded and no accepted replacement binding exists for the consumer;
- source revision/provenance required by policy cannot be established.

## 13. Minimum schema candidate

Conceptual hand-owned metadata for a Claim is intentionally small:

```yaml
claims:
  - id: <immutable-claim-id>
    anchor: <stable-anchor>
    supersedes: []   # optional; only when applicable
```

Optional policy/semantic relations such as `refines`, `kind`, or criticality linkage belong only where a defined consumer requires them.

Derived fields may include:

```text
content_digest
source_revision
superseded_by
inbound_references
effective_status
authority_owner
```

## 14. Non-goals

PKG-M0 Claim semantics do not require:

- extracting every invariant into a Claim;
- a graph database;
- vector search;
- SQLite;
- automatic semantic-diff classification;
- automatic rebinding after supersession;
- a second machine-only copy of Claim text.

## 15. Required PKG-M0 validation cases

The eventual validator must cover at least:

1. document with zero Claims is valid;
2. duplicate Claim ID fails;
3. missing Claim anchor fails;
4. one declared Claim resolving to multiple anchors fails;
5. path move does not change Claim identity;
6. meaning-preserving edit keeps ID but changes digest;
7. semantic supersession creates a new Claim identity;
8. superseded Claim is excluded from default current retrieval;
9. exact inbound reference is not silently rebound;
10. Claim in non-authoritative source cannot become project authority;
11. unresolved authority for a required Claim fails closed;
12. split/merge preserve old Claims for history without automatic inbound rewrite.

## 16. Open serialization detail

PKG-M0 still needs to freeze the concrete Markdown anchor/metadata serialization.

That serialization decision must satisfy this contract and must not change the semantic rules above.

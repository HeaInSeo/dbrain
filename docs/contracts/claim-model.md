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

### Containing document identity

A Document identity (`DocumentID`) identifies exactly one containing canonical source and is unambiguous within the Development Workspace.

```text
within a Development Workspace
DocumentID → exactly one containing canonical source
```

Two distinct canonical documents, including documents in different repositories, MUST NOT collapse onto the same bare Document identity. The representation is an implementation/profile decision — repository-qualified identity, workspace-global opaque identity, or equivalent — so long as that property holds. Anything keyed on Document identity, including containing-source eligibility resolution, depends on it: a bare filename shared by two repositories would answer for the wrong source.

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

Reference resolution also carries an explicit **use intent**. The minimum semantic distinction is:

```text
CURRENT_USE
= the consumer expects the referenced Claim to remain acceptable for present/current use

HISTORICAL_EXACT
= the consumer intentionally refers to that exact historical Claim identity
```

The Claim ID itself does not encode this intent. Intent belongs to the reference edge or the resolution request/consumer contract.

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
HISTORICAL_EXACT reference to A
→ remains valid as an exact historical binding to A

CURRENT_USE reference to A
→ becomes stale/review-required
→ does not silently bind to B
→ human/authority review explicitly rebinds to B only if semantically correct
```

A consumer-specific policy may impose stricter handling, but it MUST NOT weaken the no-silent-rebind rule.

This prevents supersession from silently changing the meaning of accepted downstream artifacts.

### Resolution provenance

`HISTORICAL_EXACT` resolution is not a current-authority surface, so it remains valid even when the containing source is ineligible for current use. To keep

```text
historically addressable != currently authoritative
```

visible to the consumer, a resolution result MUST expose the containing source's eligibility as explicit provenance rather than only as a failure reason. Ineligibility does not have to degrade an exact historical binding into a resolution failure; it must, however, always be legible in the result.

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

Replacement candidates offered after supersession are **direct observed superseders only**:

```text
A → B → C

replacement candidates for A = [B]
```

The chain is not walked to nominate a canonical replacement; any further resolution is a separate explicit step. In every case:

```text
candidate replacement != accepted rebind
```

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
- the lifecycle/authority eligibility of its containing source.

The Claim subsystem does **not** own or freeze the document lifecycle vocabulary.

Instead, current authoritative Claim resolution consumes an upstream source-eligibility determination from the applicable lifecycle/authority resolver:

```text
source eligible for current authoritative use
→ Claim may continue through Claim-specific resolution

source ineligible for current authoritative use
→ Claim MUST NOT surface as current authoritative knowledge

source eligibility cannot be established
→ UNRESOLVED
```

This keeps lifecycle semantics in their own contract while preserving the Claim invariant that a Claim cannot outlive the current-authority eligibility of its containing source merely because its own metadata still appears current.

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

## 12. Validation scope and failure semantics

Claim validation operates against an explicit **Validation Scope** rather than pretending that every validator invocation sees the complete Development Workspace.

A Validation Scope identifies:

- the document/repository/workspace material actually observed;
- the subject/scope boundary relevant to the validation request;
- whether the observed set is declared **complete** for the particular invariant being checked.

Rules:

```text
duplicate Claim ID observed inside the supplied scope
→ INVALID

a required target/reference is absent from a scope declared complete for that check
→ INVALID

a required target/reference is absent but the supplied scope is incomplete or completeness is unknown
→ UNRESOLVED

no validator may claim workspace-global uniqueness from an incomplete workspace observation
```

This applies particularly to cross-repository Claim uniqueness and supersession-target validation. The exact transport/schema of Validation Scope belongs to implementation/Observed Workspace Composite work; the semantic requirement for explicit scope + completeness is frozen here.

The resolver returns `UNRESOLVED` rather than guessing when:

- a required Claim ID cannot be found;
- more than one active Claim has the same immutable ID;
- the Claim anchor is missing or ambiguous;
- the containing source's required authority cannot be established;
- the requested current Claim is superseded and no accepted replacement binding exists for the consumer;
- source revision/provenance required by policy cannot be established;
- the supplied validation/observation scope is insufficient to establish a required cross-repository invariant;
- containing-source current-authority eligibility cannot be established;
- supersession currentness completeness for the relevant subject/scope cannot be established.

### Supersession currentness completeness

Target existence and currentness are different proofs, in opposite directions:

```text
SUPERSESSION_TARGET_EXISTENCE
observed B declares it supersedes A
→ was A observed?

SUPERSESSION_CURRENTNESS
observed A
→ was every Claim that could supersede A observed?
```

Because

```text
no superseder observed != no superseder exists
```

a superseding Claim may live in material this observation never saw. A Claim therefore MUST NOT be reported as current on the strength of an observation that never declared currentness completeness:

```text
currentness COMPLETE
+ source eligible for current authoritative use
+ no observed superseder
→ current

currentness INCOMPLETE or UNKNOWN
→ UNRESOLVED
```

The same rule governs default current-set retrieval, where two outcomes MUST remain distinguishable:

```text
resolved, and the current set is empty
!=
the current set cannot be determined
```

An incomplete observation returns UNRESOLVED, never an empty success.

### Required checks

A validation request declares which checks the requesting operation depends on.

```text
required check + completeness COMPLETE
→ the check may establish its proof

required check + completeness INCOMPLETE / UNKNOWN
→ validation aggregate UNRESOLVED

non-required check + incomplete
→ the request is not failed for that reason alone
→ but the check's proof bit remains false
```

Plain document-structure validation requires no workspace-global check, so a document with zero Claims remains VALID. A request that declares workspace-global Claim ID uniqueness as required is UNRESOLVED until completeness for that check is declared.

### Duplicate identity is ambiguous, not arbitrary

When one Claim ID is observed more than once, resolution MUST NOT return one arbitrary observation, and derived relations MUST NOT be synthesised as a union across the duplicate observations and presented as one authoritative relation. Both direct lookup and derived supersession views MUST signal the ambiguity explicitly.

### Admission boundary

Proven-invalid structure is not an admissible input to current authoritative resolution:

```text
proven structurally invalid scope
→ MUST NOT produce a current authoritative binding
```

Historical or diagnostic inspection of such material MAY remain available, but it MUST NOT act as a current-authority surface.

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
4. one declared Claim resolving to multiple anchors fails once the applicable serialization/parser layer is available;
5. duplicate locator declarations within one document fail even before concrete Markdown parsing is frozen;
6. incomplete Validation Scope cannot assert workspace-global uniqueness or missing-target invalidity;
7. a missing required target in a scope declared complete for that check fails;
8. path move does not change Claim identity;
9. meaning-preserving edit keeps ID but changes digest;
10. semantic supersession creates a new Claim identity;
11. CURRENT_USE reference to a superseded Claim becomes stale/review-required without silent rebinding;
12. HISTORICAL_EXACT reference remains bound to the exact historical Claim;
13. superseded Claim is excluded from default current retrieval;
14. Claim in non-authoritative source cannot become project authority;
15. unresolved authority or source eligibility for a required Claim fails closed;
16. split/merge preserve old Claims for history without automatic inbound rewrite;
17. a Claim is not reported as current from an observation that never declared supersession-currentness completeness;
18. an undeterminable current set is distinguishable from a resolved empty current set;
19. a proven-invalid structure cannot produce a current authoritative binding;
20. a duplicated Claim ID signals ambiguity in lookup and in derived supersession views.

## 16. Open serialization detail

PKG-M0 still needs to freeze the concrete Markdown anchor/metadata serialization.

That serialization decision must satisfy this contract and must not change the semantic rules above.

Until then, PKG-M0 core validation may validate **declared locator uniqueness/shape** without claiming that it has parsed or proven the concrete Markdown anchor relation. Full anchor-to-prose validation belongs to the serialization/parser layer once frozen.

# Authority Map Contract — Working Candidate

**Status:** WORKING · NON-NORMATIVE  
**PKG-M0 target:** freeze subject/scope authority registration, deterministic routing, refinement, ambiguity, provenance, completeness, and fail-closed semantics before runtime implementation.

## 1. Purpose

The Authority Map is a small, hand-owned routing artifact for semantic authority.

It answers:

~~~text
For this subject
at this requested scope:

WHO owns semantic authority?
WHERE does authoritative resolution begin?
WHICH accepted routing/refinement chain was used?
~~~

It does not require all truth for a subject to live in one file.

It does not make a source authoritative merely by pointing at it.

It records an already accepted or delegated authority route.

Therefore:

~~~text
Authority Map = authority routing
Authority Map != semantic content store
Authority Map != search index
Authority Map != runtime/execution authority
Authority Map != acceptance authority
~~~

## 2. Authority Map does not create authority

A route registration is not a self-authorizing act.

The right to register, refine, replace, or remove an authority route comes from the applicable Acceptance Authority or governance boundary, not from the route merely being present in the map.

~~~text
route exists in Authority Map
!=
route was legitimately authorized
~~~

The identity of the **currently accepted Authority Map revision** is an external input supplied by Acceptance Authority / governance configuration.

It MUST NOT be derived from the contents of the Authority Map itself.

A route whose subject is `authority-map` may route semantic questions about that subject, but it MUST NOT authorize the acceptance of the map revision that contains it.

~~~text
map contents
!= proof that this map revision is accepted
~~~

Current resolution therefore depends on both:

1. an externally established accepted Authority Map revision;
2. the route selected inside that accepted revision.

Historical reconstruction uses the map revision recorded in prior provenance; it does not re-accept an old map through ordinary subject routing.

The bootstrap/root acceptance mechanism itself belongs to governance/Acceptance Authority configuration outside ordinary subject routing.

## 3. Subject registration

A Subject is a stable semantic routing namespace.

Examples include `claim-model`, `authority-map`, `workspace-governance`, `repo-governance`, `execution-contract`, `artifact-handoff`, and `qualification-policy`.

A subject identifies **what semantic question is being routed**, not where its documents happen to live.

Subject identity MUST NOT be inferred from:

- file path;
- repository name alone;
- document title alone;
- search similarity;
- model prior.

For PKG-M0 there is no separate subject-registry artifact.

The registered subject set is exactly the set of subject identities appearing in accepted routes of the accepted Authority Map revision.

Introducing a new subject therefore requires an accepted route change.

A required subject absent from that accepted set is:

~~~text
UNRESOLVED / MISSING_SUBJECT
~~~

not nonexistent, false, or implicitly owned by the nearest-looking source.

## 4. Scope and Scope Profile

Authority is resolved for:

~~~text
subject + requested scope
~~~

Scope distinguishes where an authority statement applies.

A scope may represent, according to the active Scope Profile:

- the Development Workspace;
- one repository;
- a bounded subtree or domain;
- another explicitly registered narrower semantic boundary.

The exact serialized syntax of a scope remains a profile/implementation detail in PKG-M0.

The **Scope Profile** defines deterministic scope identity, applicability, and narrower-than relations.

Scope matching MUST be:

- deterministic;
- explicit;
- provenance-bearing;
- independent of semantic search;
- incapable of silently treating an unknown scope relation as a match.

~~~text
unknown scope relation
!= applicable route
~~~

If applicability cannot be established, routing fails closed:

~~~text
UNRESOLVED / SCOPE_UNRESOLVED
~~~

Because changing the Scope Profile can change the routing result even when the map revision is unchanged, authority-resolution provenance MUST bind:

- Scope Profile identity;
- Scope Profile revision and/or digest;
- the scope-matching result used for selection.

## 5. Workspace and repository boundary

~~~text
Development Workspace != Repository
~~~

A Development Workspace contains `1..N` repositories.

Authority may therefore exist at workspace scope, repository-local scope, or another explicitly registered narrower scope.

A repository-local source MUST NOT silently override a workspace-level authority route merely because it is closer to the code.

Likewise, a workspace route MUST NOT erase a deliberately delegated repository-local authority.

The applicable route is selected only through explicit subject/scope routing and valid refinement.

## 6. Authority Route and entrypoint collection

An Authority Route is a hand-owned registration containing the minimum information needed to begin authoritative resolution.

Conceptually it contains:

- immutable route identity;
- subject;
- scope;
- semantic authority owner;
- canonical entrypoint or bounded entrypoint collection;
- explicit refinement relation or relations, when applicable;
- optional critical guard Claim references.

A route MAY point to one canonical entrypoint or a bounded collection.

~~~text
canonical entrypoint
!=
all truth must live in one document
~~~

When a route contains multiple entrypoints, they are an **unordered set of required canonical corpus members/shards**.

They are not ordered fallbacks or availability alternatives.

~~~text
entrypoints = [A, B, C]
→ A, B, and C are all required canonical starting members
~~`

If any required member cannot be resolved as required for the operation:

~~~text
UNRESOLVED / SOURCE_UNRESOLVED
~~~

Mirrors, transport fallbacks, alternate URLs, and provider-level replicas belong to provider-adapter behavior, not Authority Route entrypoint ordering.

## 7. Owner and entrypoint are different

The route MUST distinguish:

~~~text
WHO owns the subject/scope
~~~

from:

~~~text
WHERE authoritative resolution begins
~~~

These are not interchangeable.

An owner may move or reorganize canonical material without changing the underlying authority boundary, provided the accepted route is updated correctly.

A source locator does not become the authority owner merely because it is an entrypoint.

## 8. Canonical entrypoint providers

An entrypoint locator is provider-agnostic at the contract level.

It may identify a Git-backed document/collection, a Notion-backed canonical source, or another declared authoritative provider.

This contract does not require Git to be the universal truth store.

For new DBrain-native projects, Git-backed human-readable contracts may be the default profile.

For existing systems, the Authority Map may continue to point to another canonical source until an explicit Authority Migration is accepted.

~~~text
new framework default
!=
forced migration of existing project authority
~~~

The current Platform project's existing Notion authority is therefore not migrated merely by adopting this contract.

A successful current resolution MUST record, for each required entrypoint, enough provider provenance to bind the exact observed source used.

That means a provider-stable revision identifier and/or content digest sufficient to identify the exact observation.

A provider may resolve a human request such as `latest`, but the resolution result must bind the concrete observed revision/digest actually used.

If policy requires reproducibility and no stable observation identity can be established:

~~~text
UNRESOLVED / SOURCE_UNRESOLVED
~~~

## 9. Refinement

A narrower authority route may explicitly refine one or more broader routes for the same subject.

Refinement is an explicit **routing-precedence** relation.

A child route that refines a parent route means:

- child and parent have the same subject identity;
- the child scope is deterministically proven **strictly narrower** than the parent scope under the active Scope Profile;
- the broader route remains applicable outside the child scope;
- the child does not silently rewrite the parent;
- the refinement relation is part of routing provenance.

Refinement MUST NOT be inferred from path nesting, repository nesting, naming convention, or textual similarity.

Refinement does not grant semantic permission for the child's canonical content to contradict parent-scope invariants.

~~~text
routing precedence
!= semantic contradiction permission
~~~

Semantic compatibility remains the responsibility of the applicable owning authorities and the Claim/knowledge contracts, outside this routing contract.

A same-scope owner replacement is a semantic route change, not refinement.

For every refinement edge `child -> parent`:

~~~text
child scope strictly narrower than parent scope
→ relation may be valid

equal / broader / disjoint
→ INVALID / REFINEMENT_SCOPE_CONTRADICTION

scope relation UNKNOWN
→ UNRESOLVED / REFINEMENT_UNRESOLVED
→ edge MUST NOT participate in specificity ordering
~~~

Refinement relations MUST also be acyclic.

~~~text
A refines B
B refines A
→ INVALID
~~~

## 10. Routing specificity

For a request consisting of subject plus requested scope, the resolver identifies all routes whose subject matches and whose scope is deterministically applicable.

Among applicable routes, **validated explicit refinement** determines specificity.

A route is more specific than another applicable route only when a valid refinement graph establishes that relationship.

The resolver selects a route only when there is exactly one applicable maximally specific route.

~~~text
zero applicable maximal routes
→ UNRESOLVED / MISSING_AUTHORITY

one applicable maximal route
→ candidate for RESOLVED

two or more incomparable maximal routes
→ UNRESOLVED / AMBIGUOUS_AUTHORITY
~~~

The resolver MUST NOT break ambiguity using:

- file order;
- route declaration order;
- lexicographic order;
- repository proximity;
- latest modified time;
- search score;
- model judgement.

Two distinct routes with identical `(subject, scope)` are a structural conflict because equal-scope precedence is not refinement.

Under a complete accepted-map observation they are:

~~~text
INVALID
~~~

Current routing MUST NOT proceed from such an invalid map revision.

`AMBIGUOUS_AUTHORITY` is reserved for structurally valid, applicable, incomparable routes.

## 11. Multi-parent refinement

A route MAY refine more than one broader route when the Scope Profile genuinely supports a narrower scope contained by all parents.

This does not authorize implicit multiple inheritance.

Every child-to-parent edge must independently satisfy the same-subject and strictly-narrower scope rule.

A multi-parent child is safe only when that child itself becomes the unique maximally specific applicable route for the request.

If the request matches multiple incomparable parents and no accepted child route validly unifies them:

~~~text
UNRESOLVED / AMBIGUOUS_AUTHORITY
~~~

The refinement topology remains a DAG, never a cyclic graph.

## 12. Route identity

Every Authority Route MUST have an immutable route identity.

Route identity MUST be independent from:

- entrypoint path;
- source title;
- current owner display name;
- declaration order;
- guard Claim reference set.

Moving a canonical entrypoint does not by itself require a new route identity.

Changing the semantic authority boundary itself MUST NOT be disguised as location-only maintenance.

The exact Route ID wire format remains an implementation/profile decision.

PKG-M0 does not add a separate Route-supersession model.

Prior decisions remain interpretable because each recorded AuthorityResolution binds the Authority Map revision in which the selected Route ID had its observed meaning.

## 13. Route changes

The framework does not infer whether a route edit is meaning-preserving.

Meaning-preserving maintenance may include:

- canonical document move without authority change;
- provider locator change after separately accepted migration preserving semantic authority;
- display-only correction.

A semantic route change includes:

- owner change;
- subject boundary change;
- scope change;
- delegation or revocation of scoped authority;
- refinement change that changes which owner wins;
- adding, removing, or rebinding a required guard Claim.

Semantic route changes require explicit acceptance.

Guard references do not change Route ID identity, but they are not location-only maintenance.

Tooling MUST NOT silently classify semantic route changes as harmless.

## 14. Failure semantics

Required semantic authority fails closed.

At minimum:

~~~text
subject unregistered
→ UNRESOLVED / MISSING_SUBJECT

subject registered but no applicable route for requested scope
→ UNRESOLVED / MISSING_AUTHORITY

more than one structurally valid incomparable maximally specific route
→ UNRESOLVED / AMBIGUOUS_AUTHORITY

scope applicability cannot be established
→ UNRESOLVED / SCOPE_UNRESOLVED

refinement relation required for selection is unknown
→ UNRESOLVED / REFINEMENT_UNRESOLVED

accepted logical map revision not completely materialized
→ UNRESOLVED / ROUTE_SET_INCOMPLETE

selected route's required canonical source cannot be resolved
→ UNRESOLVED / SOURCE_UNRESOLVED
~~~

No resolver may substitute Model Prior, nearest repository document, search result, derived index, agent-local memory, or stale resume state.

## 15. Broken or unavailable source

Selecting a route is not sufficient if its canonical entrypoint corpus cannot be resolved.

If any required entrypoint is:

- missing;
- inaccessible;
- ambiguous;
- unavailable at the required revision;
- not currently eligible as canonical source;
- unable to provide the observation identity required by policy;

the result is not successful authority resolution.

~~~text
route selected
+ required canonical source unresolved
→ UNRESOLVED / SOURCE_UNRESOLVED
~~~

The resolver MUST preserve enough reason/provenance to distinguish routing failure from source-resolution failure.

No alternate route is selected merely because the chosen route's source is unavailable.

## 16. Critical guard Claims

A route MAY bind a bounded set of applicable critical guard Claim references.

These guards do not create the route's authority and do not change which route wins.

They constrain consumers operating under the resolved route.

~~~text
route selects authority
guard constrains use
~~~

Adding, removing, or rebinding a guard is an accepted semantic route change, but does not change Route ID identity.

If a required guard reference becomes stale, superseded, ambiguous, or otherwise unresolved under the Claim Contract:

- route selection itself is not replaced by another route;
- the consuming operation fails closed until the guard binding is explicitly reviewed/rebound as required by Claim semantics.

~~~text
Authority Route != Prior Conflict Guard
Authority Route may reference applicable Guard Claims
~~~

Global, Role, and Task guard-selection policy belongs to the Prior Conflict Guard contract.

## 17. Authority resolution result

A successful resolution MUST provide enough information to reproduce why the route was chosen and which observations it depended on.

Conceptually the result includes:

- subject;
- requested scope;
- selected route ID;
- semantic authority owner;
- canonical entrypoint or entrypoints;
- per-entrypoint observed provider revision and/or content digest;
- applicable refinement chain or DAG slice;
- applicable critical guard Claim references;
- Authority Map identity/revision/digest;
- Scope Profile identity/revision/digest;
- scope-matching result/provenance;
- route-set completeness/materialization provenance.

A result is a routing decision, not a copy of authoritative semantic content.

~~~text
AuthorityResolution
!= semantic truth payload
~~~

## 18. Accepted map revision, atomic completeness, and provenance

One accepted Authority Map revision is one **complete logical route set**.

Current resolution MUST establish that it has materialized the entire accepted logical revision before it may conclude that a route is missing, unique, or maximally specific.

~~~text
partial route observation
!= complete accepted map revision
~~~

If the accepted logical revision is only partially materialized:

~~~text
UNRESOLVED / ROUTE_SET_INCOMPLETE
~~~

The resolver MUST NOT select a broader route merely because a more-specific accepted route may exist in unobserved material.

Physical fragmentation is allowed only when an explicit composition declaration identifies every fragment/member belonging to the logical revision.

That composition declaration and all member identities are bound into the Authority Map revision/digest.

~~~text
physical multi-file map
+ explicit complete composition
→ one logical revision

physical partial fetch
→ not a logical revision
~~~

Every successful current resolution therefore binds at least:

- Authority Map identity;
- accepted logical map revision/digest supplied externally;
- complete materialization/composition provenance;
- selected route identity;
- requested subject/scope;
- refinement path used;
- Scope Profile identity/revision/digest;
- per-entrypoint observed revision/digest.

A Context Bundle or other derived projection that depends on authority routing MUST bind this provenance.

A later Authority Map or Scope Profile change does not silently reinterpret an already recorded resolution.

## 19. Search, cache, and derived-index boundary

Search/index systems MAY help discover:

- candidate subjects;
- possible canonical sources;
- references to existing Route IDs.

They MUST NOT determine authority.

~~~text
search hit != registered authority
derived index != Authority Map
embedding similarity != scope applicability
LLM confidence != authority resolution
~~~

A derived route graph/cache MAY serve current resolution only when it proves that it was derived from:

- the same currently accepted Authority Map revision/digest;
- the same applicable Scope Profile revision/digest.

If that binding is absent or mismatched, the cache is stale/unbound and MUST NOT serve current authority resolution.

If the hand-owned accepted map and a derived representation disagree, the accepted map wins and the derived representation is stale or incorrect.

## 20. Candidate and agent-local boundary

Agents MAY propose:

- new subjects;
- route candidates;
- refinement candidates;
- entrypoint updates;
- guard-binding updates.

Such proposals are non-authoritative until accepted through the applicable governance/acceptance process.

Agent-local memory MUST NOT silently create or mutate an Authority Route or accepted map revision.

~~~text
candidate route != current route
~~~

## 21. Validation

Authority Map validation is deterministic structure validation over a declared observation/completeness boundary.

It validates at least:

- duplicate Route IDs;
- malformed required fields;
- duplicate/conflicting `(subject, scope)` declarations;
- invalid scope references according to the active Scope Profile;
- refinement target existence when observation is complete;
- refinement subject equality;
- refinement scope compatibility;
- self-refinement;
- refinement cycles;
- logical map composition completeness;
- deterministic uniqueness of the maximally specific route for declared test cases.

For refinement scope compatibility:

~~~text
same subject
+ child scope strictly narrower than parent
→ structurally eligible refinement

equal / broader / disjoint
→ INVALID / REFINEMENT_SCOPE_CONTRADICTION

scope relation UNKNOWN
→ UNRESOLVED / REFINEMENT_UNRESOLVED
→ edge unusable for current specificity ordering
~~~

Validation does not decide:

- who ought to own a subject;
- whether delegation was wise;
- whether two documents mean the same thing;
- whether a proposed authority change should be accepted;
- whether a migration preserves semantic meaning.

Those remain governance/acceptance decisions.

A structurally INVALID map candidate cannot be the accepted current map revision for current resolution.

## 22. Validation scope

Authority Map validation MUST NOT pretend to have observed more state than it actually has.

Cross-repository/source validation therefore requires explicit completeness for the invariant being asserted.

~~~text
observed conflict
→ INVALID

absence of conflicting route
+ complete observation
→ may establish uniqueness

absence of conflict
+ incomplete/unknown observation
→ not proof of workspace-wide uniqueness
~~~

For current routing, the stronger logical-revision rule in §18 applies: the resolver must materialize the complete accepted logical route set, not merely a convenient subset.

A required validation invariant whose necessary observation is incomplete resolves to UNRESOLVED, not success.

The exact shared Validation Scope representation may later align with Claim Core; this contract freezes semantics, not package coupling.

## 23. Minimum conceptual schema

The hand-owned conceptual shape is intentionally small.

~~~yaml
authority_map:
  id: <authority-map-id>

  routes:
    - id: <immutable-route-id>
      subject: <subject-id>
      scope: <scope-ref>
      owner: <semantic-authority-owner-ref>
      entrypoints:
        - <required-canonical-source-locator>
      refines:
        - <broader-route-id>
      guard_claims:
        - <claim-id>
~~~

If the logical map is physically fragmented, the serialization/profile must also provide an explicit composition declaration sufficient to prove the complete member set of the accepted logical revision.

The exact YAML, Markdown, or JSON serialization is not frozen by this conceptual schema.

Derived metadata may include:

- map revision/digest;
- composition digest;
- refined-by relations;
- inbound route references;
- resolved specificity;
- entrypoint observed revisions/digests;
- Scope Profile revision/digest;
- validation status.

Derived metadata MUST NOT become a second hand-maintained authority source.

## 24. Minimum resolution algorithm

Conceptually:

~~~text
resolve(subject, requested_scope, accepted_map_revision, scope_profile):

1. receive the currently accepted Authority Map revision from external governance/Acceptance Authority configuration
2. establish that the entire accepted logical map revision is materialized
   if not:
      UNRESOLVED / ROUTE_SET_INCOMPLETE
3. establish the deterministic Scope Profile identity/revision used by this operation
4. validate required structural invariants of the observed accepted revision
   invalid map candidate:
      no current resolution
5. find routes with exact subject identity
6. if no route carries that subject:
      UNRESOLVED / MISSING_SUBJECT
7. determine scope applicability using the bound Scope Profile
   unknown applicability where required:
      UNRESOLVED / SCOPE_UNRESOLVED
8. validate applicable refinement edges:
   same subject + strictly narrower child scope + acyclic topology
   unknown edge relation:
      UNRESOLVED / REFINEMENT_UNRESOLVED
9. find maximally specific applicable routes using only valid refinement edges
10. if count == 0:
       UNRESOLVED / MISSING_AUTHORITY
11. if count > 1:
       UNRESOLVED / AMBIGUOUS_AUTHORITY
12. select the single maximal route
13. resolve every required canonical entrypoint and bind its observed revision/digest
    any required source unresolved:
       UNRESOLVED / SOURCE_UNRESOLVED
14. attach guard Claim references without changing route selection
15. attach complete map + scope-profile + source provenance
16. return RESOLVED routing result
~~~

No heuristic fallback step exists.

## 25. Required PKG-M0 falsification cases

The contract must survive at least:

1. one workspace route resolves normally;
2. unregistered required subject → `MISSING_SUBJECT`;
3. registered subject with no applicable scope route → `MISSING_AUTHORITY`;
4. one repo-specific route explicitly refining workspace route wins only inside its scope;
5. workspace route remains selected outside the repo-specific scope;
6. two applicable structurally valid incomparable routes → `AMBIGUOUS_AUTHORITY`;
7. declaration order cannot break ambiguity;
8. refinement cycle → INVALID;
9. self-refinement → INVALID;
10. missing refinement target in complete validation scope → INVALID;
11. missing refinement target in incomplete validation scope → UNRESOLVED;
12. multi-parent child resolves only when every parent edge is valid and it is the unique maximal applicable route;
13. route entrypoint move alone does not create new Route ID;
14. owner/subject/scope semantic change requires explicit accepted authority change;
15. broken selected required entrypoint → `SOURCE_UNRESOLVED`, not fallback;
16. search/index candidate cannot become authority without accepted route registration;
17. agent-local proposal cannot mutate current routing;
18. resolution provenance binds accepted map revision + selected route + Scope Profile revision;
19. repository-local material does not silently override workspace authority;
20. existing non-Git canonical provider remains valid until explicit migration;
21. required guard Claim unresolved/stale → route selection unchanged, consuming operation fails closed;
22. incomplete validation observation cannot prove route uniqueness;
23. obsolete/stale map revision does not silently reinterpret prior resolution;
24. partial materialization omits an accepted refining route → `ROUTE_SET_INCOMPLETE`, never resolution through the observed parent;
25. broader child declaring `refines` → INVALID / `REFINEMENT_SCOPE_CONTRADICTION`;
26. equal-scope or disjoint child declaring `refines` → INVALID / `REFINEMENT_SCOPE_CONTRADICTION`;
27. child-parent scope relation UNKNOWN → `REFINEMENT_UNRESOLVED`; edge cannot order specificity;
28. accepted map revision is externally supplied; an unaccepted map edit cannot make itself current;
29. same map revision under two Scope Profile revisions records distinct provenance and cannot silently reuse prior routing result;
30. multiple route entrypoints behave as required corpus members, not fallback alternatives;
31. successful provider-agnostic resolution records exact observed revision/digest for every required entrypoint;
32. adding/removing/rebinding guard Claim requires accepted route change but does not change Route ID identity;
33. derived route cache without matching accepted map + Scope Profile revision binding cannot serve current resolution;
34. duplicate same `(subject, scope)` route declarations are INVALID rather than declaration-order precedence;
35. prior recorded resolution remains interpretable through its bound map revision without Route-supersession semantics.

## 26. Non-goals

PKG-M0 Authority Map does not require:

- database, SQLite, or FTS;
- embeddings, vector search, or graph database;
- background indexing;
- automatic authority inference;
- repository crawling;
- Notion-to-Git migration;
- runtime/execution authorization;
- Acceptance Authority implementation;
- semantic conflict resolution between canonical documents;
- Route supersession/lifecycle implementation;
- a production CLI.

## 27. Open serialization/runtime details

Still open after this semantic contract:

- exact Subject ID wire format;
- exact ScopeRef wire format;
- exact Scope Profile serialization;
- exact Route ID wire format;
- exact canonical source locator syntax;
- exact Authority Map file format;
- exact physical-fragment composition format;
- runtime package/API shape;
- authoring CLI/UI;
- cross-provider fetch adapters.

Those details must satisfy this contract without weakening its fail-closed routing rules.

# Authority Map Contract — Working Candidate

**Status:** WORKING · NON-NORMATIVE  
**PKG-M0 target:** freeze subject/scope authority registration, deterministic routing, refinement, ambiguity, provenance, and fail-closed semantics before runtime implementation.

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

Current resolution therefore depends on both:

1. the currently accepted Authority Map revision;
2. the route selected inside that revision.

The map MUST NOT treat its own contents as proof that a change to the map was authorized.

The bootstrap/root acceptance mechanism for the Authority Map itself belongs to governance/acceptance configuration outside ordinary subject routing.

## 3. Subject

A Subject is a stable semantic routing namespace.

Examples include claim-model, authority-map, workspace-governance, repo-governance, execution-contract, artifact-handoff, and qualification-policy.

A subject identifies what semantic question is being routed, not where its documents happen to live.

Subject identity MUST NOT be inferred from:

- file path;
- repository name alone;
- document title alone;
- search similarity;
- model prior.

A required subject that is not registered is UNRESOLVED, not nonexistent, false, or implicitly owned by the nearest-looking source.

## 4. Scope

Authority is resolved for:

~~~text
subject + requested scope
~~~

Scope distinguishes where an authority statement applies.

A scope may represent, according to the workspace profile:

- the Development Workspace;
- one repository;
- a bounded subtree or domain;
- another explicitly registered narrower semantic boundary.

The exact serialized syntax of a scope remains a profile/implementation detail in PKG-M0.

However, scope matching MUST be deterministic, explicit, provenance-bearing, independent of semantic search, and incapable of silently treating an unknown scope relation as a match.

~~~text
unknown scope relation
!= applicable route
~~~

If applicability cannot be established, routing fails closed to UNRESOLVED.

## 5. Workspace and repository boundary

~~~text
Development Workspace != Repository
~~~

A Development Workspace contains 1..N repositories.

Authority may therefore exist at workspace scope, repository-local scope, or another explicitly registered narrower scope.

A repository-local source MUST NOT silently override a workspace-level authority route merely because it is closer to the code.

Likewise, a workspace route MUST NOT erase a deliberately delegated repository-local authority.

The applicable route is selected only through explicit subject/scope routing and refinement.

## 6. Authority Route

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

The entrypoint tells the resolver where authoritative resolution begins.

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

It may identify a Git-backed document or collection, a Notion-backed canonical source, or another declared authoritative provider.

This contract does not require Git to be the universal truth store.

For new DBrain-native projects, Git-backed human-readable contracts may be the default profile.

For existing systems, the Authority Map may continue to point to another canonical source until an explicit Authority Migration is accepted.

~~~text
new framework default
!=
forced migration of existing project authority
~~~

The current Platform project's existing Notion authority is therefore not migrated merely by adopting this contract.

## 9. Refinement

A narrower authority route may explicitly refine one or more broader routes for the same subject.

Refinement is an explicit routing relation.

A child route that refines a parent route means:

- the child applies only to a narrower registered scope;
- the broader route remains applicable outside the child scope;
- the child does not silently rewrite the parent;
- the refinement relation is part of routing provenance.

Refinement MUST NOT be inferred from path nesting, repository nesting, naming convention, or textual similarity.

Refinement relations MUST be acyclic.

~~~text
A refines B
B refines A
→ INVALID
~~~

Unknown or cyclic refinement topology MUST NOT be resolved by guessing.

## 10. Routing specificity

For a request consisting of subject plus requested scope, the resolver identifies all routes whose subject matches and whose scope is deterministically applicable.

Among applicable routes, explicit refinement determines specificity.

A route is more specific than another applicable route only when the accepted refinement graph establishes that relationship.

The resolver selects a route only when there is exactly one applicable maximally specific route.

~~~text
zero applicable maximal routes
→ UNRESOLVED / MISSING_AUTHORITY

one applicable maximal route
→ RESOLVED

two or more incomparable maximal routes
→ UNRESOLVED / AMBIGUOUS_AUTHORITY
~~~

The resolver MUST NOT break ambiguity using file order, route declaration order, lexicographic order, repository proximity, latest modified time, search score, or model judgement.

## 11. Multi-parent refinement

A route MAY refine more than one broader route when the workspace scope model genuinely requires an intersection or unification.

This does not authorize implicit multiple inheritance.

A multi-parent child is safe only when that child itself becomes the unique maximally specific applicable route for the request.

If the request matches multiple incomparable parents and no accepted child route unifies them:

~~~text
UNRESOLVED / AMBIGUOUS_AUTHORITY
~~~

The refinement topology remains a DAG, never a cyclic graph.

## 12. Route identity

Every Authority Route MUST have an immutable route identity.

Route identity MUST be independent from entrypoint path, source title, current owner display name, and declaration order.

Moving a canonical entrypoint does not by itself require a new route identity.

Changing the semantic authority boundary itself is a semantic route change and MUST NOT be disguised as a location-only edit.

The exact Route ID wire format remains an implementation/profile decision.

## 13. Route changes

The framework does not infer whether a route edit is meaning-preserving.

Meaning-preserving maintenance may include a canonical document move without authority change, a storage-provider locator change after separately accepted migration, or display-only corrections.

A semantic authority change includes owner change, subject boundary change, delegation or revocation of scoped authority, or a refinement change that changes which owner wins.

Semantic authority changes require explicit acceptance.

Tooling MUST NOT silently classify them as harmless.

## 14. Ambiguity and missing authority

Required semantic authority fails closed.

At minimum:

~~~text
subject unregistered
→ UNRESOLVED / MISSING_SUBJECT

subject registered but no applicable route for requested scope
→ UNRESOLVED / MISSING_AUTHORITY

more than one incomparable maximally specific route
→ UNRESOLVED / AMBIGUOUS_AUTHORITY

scope applicability cannot be established
→ UNRESOLVED / SCOPE_UNRESOLVED

refinement relation required for selection is unknown
→ UNRESOLVED / REFINEMENT_UNRESOLVED
~~~

No resolver may substitute Model Prior, nearest repository document, search result, derived index, agent-local memory, or stale resume state.

## 15. Broken or unavailable source

Selecting a route is not sufficient if its canonical entrypoint cannot be resolved.

If the chosen route exists but its required entrypoint is missing, inaccessible, ambiguous, unavailable at the required revision, or not currently eligible as canonical source, the result is not successful authority resolution.

~~~text
route selected
+ canonical source unresolved
→ UNRESOLVED
~~~

The result MUST preserve enough reason/provenance to distinguish routing failure from source-resolution failure.

## 16. Critical guard Claims

A route MAY bind a bounded set of applicable critical guard Claim references.

These guards do not create the route's authority.

They constrain consumers operating under the resolved route.

A required guard Claim that cannot be resolved according to Claim semantics causes the consuming operation to fail closed.

~~~text
Authority Route != Prior Conflict Guard
Authority Route may reference applicable Guard Claims
~~~

Global, Role, and Task guard selection policy belongs to the Prior Conflict Guard contract.

## 17. Authority resolution result

A successful resolution MUST provide enough information to reproduce why the route was chosen.

Conceptually the result includes:

- subject;
- requested scope;
- selected route ID;
- semantic authority owner;
- canonical entrypoint or entrypoints;
- applicable refinement chain or DAG slice;
- applicable critical guard Claim references;
- Authority Map identity/revision/digest;
- scope-matching provenance.

A result is a routing decision, not a copy of authoritative semantic content.

~~~text
AuthorityResolution
!= semantic truth payload
~~~

## 18. Map revision provenance

Every resolution MUST bind to an observed Authority Map revision.

At minimum, provenance must identify the Authority Map identity, observed map revision, selected route identity, requested subject/scope, and refinement path used.

A Context Bundle or other derived projection that depends on authority routing MUST bind to this revision/provenance.

A later Authority Map change does not silently reinterpret an already recorded resolution.

## 19. Search and derived-index boundary

Search/index systems MAY help discover candidate subjects, possible canonical sources, or references to route IDs.

They MUST NOT determine authority.

~~~text
search hit != registered authority
derived index != Authority Map
embedding similarity != scope applicability
LLM confidence != authority resolution
~~~

If the hand-owned accepted Authority Map and a derived index disagree, the accepted map wins and the index is stale or incorrect.

## 20. Candidate and agent-local boundary

Agents MAY propose new subjects, route candidates, refinement candidates, or possible entrypoint updates.

Such proposals are non-authoritative until accepted through the applicable governance/acceptance process.

Agent-local memory MUST NOT silently create or mutate an Authority Route.

~~~text
candidate route != current route
~~~

## 21. Validation

Authority Map validation is deterministic structure validation.

It may validate at least:

- duplicate Route IDs;
- malformed required fields;
- duplicate/conflicting route declarations;
- invalid scope references according to the active scope profile;
- refinement target existence when validation scope is complete;
- refinement subject compatibility;
- self-refinement and refinement cycles;
- deterministic uniqueness of the maximally specific route for declared test cases.

Validation does not decide who ought to own a subject, whether delegation was wise, whether two documents mean the same thing, or whether a proposed authority change should be accepted.

Those remain governance/acceptance decisions.

## 22. Validation scope

Authority Map validation MUST NOT pretend to have observed more workspace state than it actually has.

Cross-repository/source validation therefore requires explicit completeness for the invariant being asserted.

~~~text
observed conflict
→ INVALID

absence of conflicting route
+ complete validation scope
→ may establish uniqueness

absence of conflict
+ incomplete/unknown scope
→ not proof of workspace-wide uniqueness
~~~

A required invariant whose necessary observation is incomplete resolves to UNRESOLVED, not success.

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
        - <canonical-source-locator>
      refines:
        - <broader-route-id>
      guard_claims:
        - <claim-id>
~~~

The exact YAML, Markdown, or JSON serialization is not frozen by this schema.

Derived metadata may include map revision, map digest, refined-by relations, inbound references, resolved specificity, entrypoint revision, and validation status.

Derived metadata MUST NOT become a second hand-maintained authority source.

## 24. Minimum resolution algorithm

Conceptually:

~~~text
resolve(subject, requested_scope, observed_map_revision):

1. establish the map revision is eligible/current for this operation
2. find routes with exact subject identity
3. determine scope applicability using the accepted deterministic scope profile
4. if applicability is unknown where required:
      UNRESOLVED
5. compute explicit refinement ordering among applicable routes
6. reject or withhold on cyclic/unresolved refinement topology
7. find maximally specific applicable routes
8. if count == 0:
      UNRESOLVED / MISSING_AUTHORITY
9. if count > 1:
      UNRESOLVED / AMBIGUOUS_AUTHORITY
10. select the single maximal route
11. resolve required canonical entrypoint(s)
12. attach guard Claim references and provenance
13. return RESOLVED routing result
~~~

No heuristic fallback step exists.

## 25. Required PKG-M0 falsification cases

The contract must survive at least:

1. one workspace route resolves normally;
2. unregistered required subject → UNRESOLVED;
3. registered subject with no applicable scope route → UNRESOLVED;
4. one repo-specific route explicitly refining workspace route wins only inside its scope;
5. workspace route remains selected outside the repo-specific scope;
6. two applicable incomparable routes → AMBIGUOUS_AUTHORITY;
7. declaration order cannot break ambiguity;
8. refinement cycle → INVALID;
9. self-refinement → INVALID;
10. missing refinement target in complete validation scope → INVALID;
11. missing refinement target in incomplete scope → UNRESOLVED;
12. multi-parent child resolves only when it is the unique maximal applicable route;
13. route entrypoint move alone does not create new Route ID;
14. owner/scope semantic change requires explicit accepted authority change;
15. broken selected entrypoint → UNRESOLVED, not fallback;
16. search/index candidate cannot become authority without registration;
17. agent-local proposal cannot mutate current routing;
18. resolution provenance binds map revision + selected route;
19. repository-local material does not silently override workspace authority;
20. existing non-Git canonical provider remains valid until explicit migration;
21. required guard Claim unresolved → consuming operation fails closed;
22. incomplete observation cannot prove workspace-global route uniqueness;
23. obsolete/stale map revision does not silently reinterpret prior resolution.

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
- a production CLI.

## 27. Open serialization/runtime details

Still open after this semantic contract:

- exact Subject ID wire format;
- exact ScopeRef wire format;
- exact scope-profile serialization;
- exact Route ID wire format;
- exact canonical source locator syntax;
- exact Authority Map file format;
- runtime package/API shape;
- authoring CLI/UI;
- cross-provider fetch adapters.

Those details must satisfy this contract without weakening its fail-closed routing rules.

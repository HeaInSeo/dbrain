# Context Bundle Contract — Working Candidate

Context Bundle is a bounded, derived projection.

## Candidate semantic identity

Bind to:

- task/work identity;
- Authority Map/workspace authority revision;
- included Claim IDs + content digests;
- included non-Claim canonical source revisions;
- relevant workspace/repository observation;
- budget/profile identity;
- rendering/template version.

Fast-changing evidence should normally enter via subject-bound pointer/digest/provenance rather than causing semantic bundle churn.

## Budget classes

- NON-DROPPABLE
- SUMMARIZABLE
- POINTER-ONLY
- EXCLUDED

If required NON-DROPPABLE material does not fit, do not silently truncate. Narrow/split the task, use an appropriate context profile, or return `UNRESOLVED`.

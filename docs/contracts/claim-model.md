# Claim Model Contract — Working Candidate

## Purpose

A Document remains the human-readable canonical narrative container.

A Claim is an optional stable semantic address used only where another consumer needs independent citation, supersession, qualification, guard, cross-repo contract, or negative-decision retrieval.

## Rule

Claims are **demand-driven, not coverage-driven**.

A document with zero Claims is valid.

Mint a Claim when at least one of these is true:

- another artifact must cite the rule;
- qualification tests it;
- it is in an applicable Prior Conflict Guard set;
- it must be superseded independently;
- a rejected alternative should trigger when proposed again.

## Identity

- Claim IDs are immutable.
- Meaning-preserving edits may keep the Claim ID while changing its content digest.
- Meaning-changing edits require supersession/new semantic identity.

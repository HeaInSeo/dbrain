package claim

import (
	"fmt"
	"sort"
)

// Status is the outcome of a validation request.
//
// INVALID and UNRESOLVED are deliberately distinct: INVALID is a proven defect
// in declared structure, UNRESOLVED is an honest "not determinable from what
// was observed".
type Status string

const (
	// StatusValid means no defect was found in the observed material, within
	// the limits of the declared scope completeness.
	StatusValid Status = "VALID"

	// StatusInvalid means a defect was proven from the observed material.
	StatusInvalid Status = "INVALID"

	// StatusUnresolved means an answer requires material outside the supplied
	// scope, or an input the caller left undetermined.
	StatusUnresolved Status = "UNRESOLVED"
)

// IssueCode is a machine-readable reason for a validation issue.
type IssueCode string

const (
	// IssueEmptyClaimID marks a declared Claim without an identity.
	IssueEmptyClaimID IssueCode = "EMPTY_CLAIM_ID"

	// IssueDuplicateClaimID marks two observed Claims sharing one ID.
	IssueDuplicateClaimID IssueCode = "DUPLICATE_CLAIM_ID"

	// IssueEmptyLocator marks a Claim that declares no locator at all. This is
	// a declared-shape defect, not a parse result.
	IssueEmptyLocator IssueCode = "EMPTY_LOCATOR"

	// IssueDuplicateLocator marks two Claims in one document declaring the same
	// locator.
	IssueDuplicateLocator IssueCode = "DUPLICATE_LOCATOR"

	// IssueEmptySupersessionTarget marks a supersession edge to an empty ID.
	IssueEmptySupersessionTarget IssueCode = "EMPTY_SUPERSESSION_TARGET"

	// IssueSelfSupersession marks a Claim that supersedes itself.
	IssueSelfSupersession IssueCode = "SELF_SUPERSESSION"

	// IssueSupersessionCycle marks a cycle in the supersession graph, which
	// must be acyclic.
	IssueSupersessionCycle IssueCode = "SUPERSESSION_CYCLE"

	// IssueMissingSupersessionTarget marks a declared target absent from a
	// scope declared complete for that check.
	IssueMissingSupersessionTarget IssueCode = "MISSING_SUPERSESSION_TARGET"

	// IssueIncompleteScope marks a target that could not be found in a scope
	// that was not declared complete for that check. Absence proves nothing
	// here, so the outcome is UNRESOLVED rather than INVALID.
	IssueIncompleteScope IssueCode = "INCOMPLETE_SCOPE"
)

// Issue is a single machine-readable validation finding.
type Issue struct {
	Code IssueCode

	// Severity is StatusInvalid for a proven defect, StatusUnresolved for an
	// undeterminable check.
	Severity Status

	ClaimID    ClaimID
	DocumentID DocumentID

	// Related carries the other identities involved, such as the other Claim
	// sharing an ID or locator, the missing target, or the cycle members.
	Related []ClaimID

	// Detail is a human-readable explanation. It is never parsed.
	Detail string
}

// Result is the outcome of validating a ValidationScope.
type Result struct {
	Status Status

	// Issues are ordered deterministically by code, then Claim, then document.
	Issues []Issue

	// GlobalUniquenessEstablished is true only when the scope was declared
	// complete for CheckClaimIDUniqueness and no duplicate was observed.
	// A partial snapshot is never proof of workspace-global uniqueness, so a
	// clean result over an incomplete scope leaves this false.
	GlobalUniquenessEstablished bool
}

// IssuesWithCode returns the issues carrying a given code.
func (r Result) IssuesWithCode(code IssueCode) []Issue {
	var out []Issue
	for _, i := range r.Issues {
		if i.Code == code {
			out = append(out, i)
		}
	}
	return out
}

// HasCode reports whether the result carries at least one issue with the code.
func (r Result) HasCode(code IssueCode) bool {
	return len(r.IssuesWithCode(code)) > 0
}

// Validate checks the declared structure of the observed Claims.
//
// It answers only "is the declared structure valid", never semantic questions
// such as whether a wording change was meaningful, whether A ought to supersede
// B, or whether two Claims mean the same thing. Those are authoring decisions.
//
// A scope with zero Claims is VALID.
func Validate(scope ValidationScope) Result {
	var issues []Issue

	byID := make(map[ClaimID][]Claim, len(scope.Observed))
	for _, c := range scope.Observed {
		if c.ID == "" {
			issues = append(issues, Issue{
				Code:       IssueEmptyClaimID,
				Severity:   StatusInvalid,
				DocumentID: c.DocumentID,
				Detail:     "declared Claim has no immutable Claim ID",
			})
			continue
		}
		byID[c.ID] = append(byID[c.ID], c)
	}

	issues = append(issues, duplicateIDIssues(byID)...)
	issues = append(issues, locatorIssues(scope.Observed)...)
	issues = append(issues, supersessionIssues(scope, byID)...)

	sortIssues(issues)

	result := Result{Status: StatusValid, Issues: issues}
	for _, i := range issues {
		if i.Severity == StatusInvalid {
			result.Status = StatusInvalid
			break
		}
		if i.Severity == StatusUnresolved {
			result.Status = StatusUnresolved
		}
	}

	// Uniqueness over an incomplete observation is not workspace-global proof.
	result.GlobalUniquenessEstablished = scope.IsCompleteFor(CheckClaimIDUniqueness) &&
		!hasCode(issues, IssueDuplicateClaimID) &&
		!hasCode(issues, IssueEmptyClaimID)

	return result
}

// duplicateIDIssues reports IDs observed more than once inside the supplied
// scope. Observing a duplicate is conclusive regardless of completeness.
func duplicateIDIssues(byID map[ClaimID][]Claim) []Issue {
	var issues []Issue
	for id, claims := range byID {
		if len(claims) < 2 {
			continue
		}
		docs := make([]string, 0, len(claims))
		for _, c := range claims {
			docs = append(docs, string(c.DocumentID))
		}
		sort.Strings(docs)
		issues = append(issues, Issue{
			Code:       IssueDuplicateClaimID,
			Severity:   StatusInvalid,
			ClaimID:    id,
			DocumentID: claims[0].DocumentID,
			Detail: fmt.Sprintf("Claim ID observed %d times in supplied scope (documents: %v)",
				len(claims), docs),
		})
	}
	return issues
}

// locatorIssues validates declared locator shape and per-document uniqueness.
// It does not parse, resolve or interpret the locator: anchor-to-prose
// validation belongs to the serialization layer, which is not frozen yet.
func locatorIssues(observed []Claim) []Issue {
	var issues []Issue
	type docLocator struct {
		doc     DocumentID
		locator Locator
	}
	seen := make(map[docLocator]ClaimID, len(observed))

	for _, c := range observed {
		if c.Locator == "" {
			issues = append(issues, Issue{
				Code:       IssueEmptyLocator,
				Severity:   StatusInvalid,
				ClaimID:    c.ID,
				DocumentID: c.DocumentID,
				Detail:     "declared Claim has no locator, so it addresses no bounded span",
			})
			continue
		}
		key := docLocator{doc: c.DocumentID, locator: c.Locator}
		if first, ok := seen[key]; ok {
			issues = append(issues, Issue{
				Code:       IssueDuplicateLocator,
				Severity:   StatusInvalid,
				ClaimID:    c.ID,
				DocumentID: c.DocumentID,
				Related:    []ClaimID{first},
				Detail: fmt.Sprintf("locator %q is declared by more than one Claim in document %q",
					c.Locator, c.DocumentID),
			})
			continue
		}
		seen[key] = c.ID
	}
	return issues
}

// supersessionIssues validates the declared supersession graph: no self edges,
// no cycles, and targets that either exist or are honestly UNRESOLVED.
func supersessionIssues(scope ValidationScope, byID map[ClaimID][]Claim) []Issue {
	var issues []Issue
	targetsComplete := scope.IsCompleteFor(CheckSupersessionTargetExistence)

	edges := make(map[ClaimID][]ClaimID, len(byID))
	ids := make([]ClaimID, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sortIDs(ids)

	for _, id := range ids {
		seenTarget := make(map[ClaimID]bool)
		for _, c := range byID[id] {
			for _, target := range c.Supersedes {
				if target == "" {
					issues = append(issues, Issue{
						Code:       IssueEmptySupersessionTarget,
						Severity:   StatusInvalid,
						ClaimID:    id,
						DocumentID: c.DocumentID,
						Detail:     "supersession target is empty",
					})
					continue
				}
				if target == id {
					if !seenTarget[target] {
						issues = append(issues, Issue{
							Code:       IssueSelfSupersession,
							Severity:   StatusInvalid,
							ClaimID:    id,
							DocumentID: c.DocumentID,
							Related:    []ClaimID{target},
							Detail:     "a Claim cannot supersede itself",
						})
					}
					seenTarget[target] = true
					continue
				}
				if seenTarget[target] {
					continue
				}
				seenTarget[target] = true

				if _, ok := byID[target]; ok {
					edges[id] = append(edges[id], target)
					continue
				}
				if targetsComplete {
					issues = append(issues, Issue{
						Code:       IssueMissingSupersessionTarget,
						Severity:   StatusInvalid,
						ClaimID:    id,
						DocumentID: c.DocumentID,
						Related:    []ClaimID{target},
						Detail: fmt.Sprintf("supersession target %q is absent from a scope declared complete for %s",
							target, CheckSupersessionTargetExistence),
					})
					continue
				}
				issues = append(issues, Issue{
					Code:       IssueIncompleteScope,
					Severity:   StatusUnresolved,
					ClaimID:    id,
					DocumentID: c.DocumentID,
					Related:    []ClaimID{target},
					Detail: fmt.Sprintf("supersession target %q was not observed and scope completeness for %s is %q",
						target, CheckSupersessionTargetExistence,
						scope.CompletenessFor(CheckSupersessionTargetExistence)),
				})
			}
		}
	}

	for _, cycle := range findCycles(ids, edges) {
		issues = append(issues, Issue{
			Code:     IssueSupersessionCycle,
			Severity: StatusInvalid,
			ClaimID:  cycle[0],
			Related:  cycle,
			Detail:   fmt.Sprintf("supersession graph must be acyclic; cycle: %v", cycle),
		})
	}
	return issues
}

// findCycles returns each distinct supersession cycle once, normalised to start
// at its smallest member so output is deterministic.
func findCycles(ids []ClaimID, edges map[ClaimID][]ClaimID) [][]ClaimID {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[ClaimID]int, len(ids))
	var stack []ClaimID
	seen := make(map[string]bool)
	var cycles [][]ClaimID

	var visit func(ClaimID)
	visit = func(id ClaimID) {
		color[id] = gray
		stack = append(stack, id)
		targets := append([]ClaimID(nil), edges[id]...)
		sortIDs(targets)
		for _, target := range targets {
			switch color[target] {
			case white:
				visit(target)
			case gray:
				cycle := normaliseCycle(stack, target)
				key := fmt.Sprint(cycle)
				if !seen[key] {
					seen[key] = true
					cycles = append(cycles, cycle)
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[id] = black
	}

	for _, id := range ids {
		if color[id] == white {
			visit(id)
		}
	}
	return cycles
}

// normaliseCycle extracts the cycle closing back to start from the DFS stack
// and rotates it to begin at its smallest member.
func normaliseCycle(stack []ClaimID, start ClaimID) []ClaimID {
	at := 0
	for i, id := range stack {
		if id == start {
			at = i
			break
		}
	}
	cycle := append([]ClaimID(nil), stack[at:]...)

	lowest := 0
	for i, id := range cycle {
		if id < cycle[lowest] {
			lowest = i
		}
	}
	return append(cycle[lowest:], cycle[:lowest]...)
}

func hasCode(issues []Issue, code IssueCode) bool {
	for _, i := range issues {
		if i.Code == code {
			return true
		}
	}
	return false
}

func sortIssues(issues []Issue) {
	sort.SliceStable(issues, func(i, j int) bool {
		a, b := issues[i], issues[j]
		if a.Code != b.Code {
			return a.Code < b.Code
		}
		if a.ClaimID != b.ClaimID {
			return a.ClaimID < b.ClaimID
		}
		if a.DocumentID != b.DocumentID {
			return a.DocumentID < b.DocumentID
		}
		return fmt.Sprint(a.Related) < fmt.Sprint(b.Related)
	})
}

func sortIDs(ids []ClaimID) {
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
}

package claim

import (
	"errors"
	"testing"
)

// scopeWith builds a scope whose completeness is declared per check, so a test
// can vary exactly one proof at a time.
func scopeWith(complete map[Check]Completeness, observed ...Claim) ValidationScope {
	return ValidationScope{Observed: observed, Complete: complete}
}

// C1–C4: a Claim is never called current merely because this observation
// happened to contain no superseder.
//
//	no superseder observed != no superseder exists
func TestCurrentUseRequiresCurrentnessCompleteness(t *testing.T) {
	observed := doc("claim:a", "doc-1", "#one")

	tests := []struct {
		name       string
		complete   map[Check]Completeness
		want       ResolutionStatus
		wantReason ResolutionReason
	}{
		{
			// C1
			name:       "currentness unknown",
			complete:   nil,
			want:       StatusResolutionUnresolved,
			wantReason: ReasonCurrentnessIncomplete,
		},
		{
			// C2
			name:       "currentness incomplete",
			complete:   map[Check]Completeness{CheckSupersessionCurrentness: CompletenessIncomplete},
			want:       StatusResolutionUnresolved,
			wantReason: ReasonCurrentnessIncomplete,
		},
		{
			// C4: proving that observed targets exist says nothing about
			// whether an unobserved Claim supersedes this one.
			name: "target existence complete but currentness unknown",
			complete: map[Check]Completeness{
				CheckSupersessionTargetExistence: CompletenessComplete,
				CheckClaimIDUniqueness:           CompletenessComplete,
			},
			want:       StatusResolutionUnresolved,
			wantReason: ReasonCurrentnessIncomplete,
		},
		{
			// C3
			name:     "currentness complete",
			complete: map[Check]Completeness{CheckSupersessionCurrentness: CompletenessComplete},
			want:     StatusResolvedCurrent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, validation := Admit(scopeWith(tt.complete, observed), eligibleEverywhere)
			if validation.Status == StatusInvalid {
				t.Fatalf("scope unexpectedly invalid: %+v", validation.Issues)
			}

			got := r.Resolve(ClaimReference{ClaimID: "claim:a", Intent: IntentCurrentUse})
			if got.Status != tt.want {
				t.Fatalf("Resolve() status = %q, want %q (%s)", got.Status, tt.want, got.Detail)
			}
			if tt.wantReason != ReasonNone && got.Reason != tt.wantReason {
				t.Errorf("Resolve() reason = %q, want %q", got.Reason, tt.wantReason)
			}
			if _, ok := got.Binding(); ok != (tt.want == StatusResolvedCurrent) {
				t.Errorf("Binding() ok = %v for status %q", ok, got.Status)
			}
		})
	}
}

// An observed superseder is a determinate answer: it stays stale rather than
// degrading to UNRESOLVED when currentness completeness is missing.
func TestObservedSupersederIsStaleEvenWithoutCurrentnessProof(t *testing.T) {
	r, _ := Admit(scopeWith(nil,
		doc("claim:a", "doc-1", "#one"),
		doc("claim:b", "doc-1", "#two", "claim:a"),
	), eligibleEverywhere)

	got := r.Resolve(ClaimReference{ClaimID: "claim:a", Intent: IntentCurrentUse})
	if got.Status != StatusStaleReviewRequired || got.Reason != ReasonSuperseded {
		t.Fatalf("Resolve() = %q/%q, want %q/%q", got.Status, got.Reason,
			StatusStaleReviewRequired, ReasonSuperseded)
	}
}

// C5 / C6: "the current set is empty" and "the current set cannot be
// determined" must stay distinguishable.
func TestCurrentClaimsDistinguishesEmptyFromUnknown(t *testing.T) {
	// C5
	incomplete, _ := Admit(scopeWith(
		map[Check]Completeness{CheckSupersessionCurrentness: CompletenessIncomplete},
		doc("claim:a", "doc-1", "#one"),
	), eligibleEverywhere)

	unresolved := incomplete.CurrentClaims()
	if unresolved.Resolved() {
		t.Fatalf("CurrentClaims() = %q, want unresolved", unresolved.Status)
	}
	if unresolved.Reason != ReasonCurrentnessIncomplete {
		t.Errorf("reason = %q, want %q", unresolved.Reason, ReasonCurrentnessIncomplete)
	}
	if len(unresolved.ClaimIDs) != 0 {
		t.Errorf("unresolved set carried %v, want no claims", unresolved.ClaimIDs)
	}

	// C6: complete observation whose only Claim sits in an ineligible source.
	empty, _ := Admit(scopeWith(
		map[Check]Completeness{CheckSupersessionCurrentness: CompletenessComplete},
		doc("claim:a", "doc-bad", "#one"),
	), eligibility(map[DocumentID]SourceEligibility{"doc-bad": Ineligible}))

	resolved := empty.CurrentClaims()
	if !resolved.Resolved() {
		t.Fatalf("CurrentClaims() = %q/%q, want resolved empty set", resolved.Status, resolved.Reason)
	}
	if len(resolved.ClaimIDs) != 0 {
		t.Errorf("CurrentClaims() = %v, want empty", resolved.ClaimIDs)
	}
	if unresolved.Status == resolved.Status {
		t.Error("unknown and empty-success outcomes must be distinguishable")
	}
}

// C9 / C10: a proven structural defect is not an admissible input to current
// authoritative resolution.
func TestInvalidStructureCannotProduceCurrentBinding(t *testing.T) {
	complete := map[Check]Completeness{
		CheckClaimIDUniqueness:           CompletenessComplete,
		CheckSupersessionTargetExistence: CompletenessComplete,
		CheckSupersessionCurrentness:     CompletenessComplete,
	}

	tests := []struct {
		name     string
		scope    ValidationScope
		refID    ClaimID
		wantHist ResolutionStatus
	}{
		{
			// C9
			name:     "empty locator",
			scope:    scopeWith(complete, doc("claim:a", "doc-1", "")),
			refID:    "claim:a",
			wantHist: StatusResolvedExact,
		},
		{
			// C10
			name: "duplicate locator in one document",
			scope: scopeWith(complete,
				doc("claim:a", "doc-1", "#same"),
				doc("claim:b", "doc-1", "#same"),
			),
			refID:    "claim:a",
			wantHist: StatusResolvedExact,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, validation := Admit(tt.scope, eligibleEverywhere)
			if validation.Status != StatusInvalid {
				t.Fatalf("validation = %q, want %q", validation.Status, StatusInvalid)
			}
			if r.Admitted() {
				t.Error("a proven-invalid scope must not be admitted")
			}

			current := r.Resolve(ClaimReference{ClaimID: tt.refID, Intent: IntentCurrentUse})
			if current.Status != StatusResolutionUnresolved || current.Reason != ReasonScopeStructurallyInvalid {
				t.Fatalf("Resolve(CURRENT_USE) = %q/%q, want %q/%q", current.Status, current.Reason,
					StatusResolutionUnresolved, ReasonScopeStructurallyInvalid)
			}
			if _, ok := current.Binding(); ok {
				t.Error("invalid structure produced a current binding")
			}
			if set := r.CurrentClaims(); set.Resolved() {
				t.Errorf("CurrentClaims() = %+v, want unresolved", set)
			}

			// The historical/diagnostic surface stays available.
			hist := r.Resolve(ClaimReference{ClaimID: tt.refID, Intent: IntentHistoricalExact})
			if hist.Status != tt.wantHist {
				t.Errorf("Resolve(HISTORICAL_EXACT) = %q, want %q", hist.Status, tt.wantHist)
			}
		})
	}
}

// C11 / C12: a duplicated identity never resolves to one arbitrary observation,
// and its derived relations are never a synthetic union.
func TestDuplicateIdentityAPIsSignalAmbiguity(t *testing.T) {
	// claim:dup is observed twice with divergent supersession edges.
	scope := scopeWith(map[Check]Completeness{CheckSupersessionCurrentness: CompletenessComplete},
		doc("claim:a", "doc-1", "#one"),
		doc("claim:dup", "doc-2", "#two", "claim:a"),
		doc("claim:dup", "doc-3", "#three"),
	)
	r, validation := Admit(scope, eligibleEverywhere)
	if validation.Status != StatusInvalid || !validation.HasCode(IssueDuplicateClaimID) {
		t.Fatalf("validation = %q, want %q with a duplicate-ID issue", validation.Status, StatusInvalid)
	}

	// C11
	if _, err := r.Lookup("claim:dup"); !errors.Is(err, ErrAmbiguousClaimID) {
		t.Errorf("Lookup(claim:dup) error = %v, want ErrAmbiguousClaimID", err)
	}
	if _, err := r.Lookup("claim:absent"); !errors.Is(err, ErrClaimNotObserved) {
		t.Errorf("Lookup(claim:absent) error = %v, want ErrClaimNotObserved", err)
	}

	// C12: the edge into claim:a came from a duplicated identity, so the
	// derived superseder set of claim:a is ambiguous rather than [claim:dup].
	if _, err := r.SupersededBy("claim:a"); !errors.Is(err, ErrAmbiguousClaimID) {
		t.Errorf("SupersededBy(claim:a) error = %v, want ErrAmbiguousClaimID", err)
	}
	if _, err := r.SupersededBy("claim:dup"); !errors.Is(err, ErrAmbiguousClaimID) {
		t.Errorf("SupersededBy(claim:dup) error = %v, want ErrAmbiguousClaimID", err)
	}

	got := r.Resolve(ClaimReference{ClaimID: "claim:a", Intent: IntentHistoricalExact})
	if got.Status != StatusResolutionUnresolved || got.Reason != ReasonAmbiguousSupersession {
		t.Errorf("Resolve(claim:a) = %q/%q, want %q/%q", got.Status, got.Reason,
			StatusResolutionUnresolved, ReasonAmbiguousSupersession)
	}
}

// C13 / C14: historical addressability is not authority, so every resolution
// carries the containing source's eligibility as provenance.
func TestHistoricalExactCarriesEligibilityProvenance(t *testing.T) {
	scope := completeScope(
		doc("claim:bad", "doc-bad", "#one"),
		doc("claim:unknown", "doc-unknown", "#two"),
		doc("claim:ok", "doc-ok", "#three"),
	)
	r, _ := Admit(scope, eligibility(map[DocumentID]SourceEligibility{
		"doc-ok":  Eligible,
		"doc-bad": Ineligible,
	}))

	tests := []struct {
		name           string
		id             ClaimID
		wantEligiblity SourceEligibility
	}{
		// C13
		{name: "ineligible source", id: "claim:bad", wantEligiblity: Ineligible},
		// C14
		{name: "unknown eligibility", id: "claim:unknown", wantEligiblity: EligibilityUnknown},
		{name: "eligible source", id: "claim:ok", wantEligiblity: Eligible},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.Resolve(ClaimReference{ClaimID: tt.id, Intent: IntentHistoricalExact})
			if got.Status != StatusResolvedExact {
				t.Fatalf("status = %q, want %q (%s)", got.Status, StatusResolvedExact, got.Detail)
			}
			if got.SourceEligibility != tt.wantEligiblity {
				t.Errorf("SourceEligibility = %q, want %q", got.SourceEligibility, tt.wantEligiblity)
			}
			if got.Reason == ReasonSourceIneligible {
				t.Error("ineligibility must be provenance here, not a resolution failure reason")
			}
		})
	}

	// Current use over the same material still fails closed.
	if got := r.Resolve(ClaimReference{ClaimID: "claim:bad", Intent: IntentCurrentUse}); got.Status != StatusNotCurrentAuthoritative {
		t.Errorf("CURRENT_USE status = %q, want %q", got.Status, StatusNotCurrentAuthoritative)
	}
}

// Replacement candidates are the directly observed superseders only: the chain
// is never walked to nominate a canonical replacement.
func TestReplacementCandidatesAreDirectOnly(t *testing.T) {
	scope := completeScope(
		doc("claim:a", "doc-1", "#one"),
		doc("claim:b", "doc-1", "#two", "claim:a"),
		doc("claim:c", "doc-1", "#three", "claim:b"),
	)
	r, _ := Admit(scope, eligibleEverywhere)

	got := r.Resolve(ClaimReference{ClaimID: "claim:a", Intent: IntentCurrentUse})
	if got.Status != StatusStaleReviewRequired {
		t.Fatalf("status = %q, want %q", got.Status, StatusStaleReviewRequired)
	}
	if len(got.ReplacementCandidates) != 1 || got.ReplacementCandidates[0] != "claim:b" {
		t.Fatalf("ReplacementCandidates = %v, want the direct superseder [claim:b] only", got.ReplacementCandidates)
	}
	if _, ok := got.Binding(); ok {
		t.Error("a candidate must never become an accepted binding")
	}
}

// DocumentID identifies one containing source within the workspace, so
// same-named documents in different repositories stay distinct and their
// eligibility is answered separately.
func TestDocumentIDIsWorkspaceUnique(t *testing.T) {
	scope := completeScope(
		doc("claim:a", "repo-a/docs/rule.md", "#rule"),
		doc("claim:b", "repo-b/docs/rule.md", "#rule"),
	)
	r, validation := Admit(scope, eligibility(map[DocumentID]SourceEligibility{
		"repo-a/docs/rule.md": Eligible,
		"repo-b/docs/rule.md": Ineligible,
	}))
	if validation.Status != StatusValid {
		t.Fatalf("validation = %q, want %q (issues: %+v)", validation.Status, StatusValid, validation.Issues)
	}

	fromA := r.Resolve(ClaimReference{ClaimID: "claim:a", Intent: IntentCurrentUse})
	if fromA.Status != StatusResolvedCurrent {
		t.Errorf("claim:a status = %q, want %q (%s)", fromA.Status, StatusResolvedCurrent, fromA.Detail)
	}
	fromB := r.Resolve(ClaimReference{ClaimID: "claim:b", Intent: IntentCurrentUse})
	if fromB.Status != StatusNotCurrentAuthoritative {
		t.Errorf("claim:b status = %q, want %q", fromB.Status, StatusNotCurrentAuthoritative)
	}

	// Collapsing both onto a bare document name would answer for the wrong
	// source; the contract forbids that representation.
	if fromA.Claim.DocumentID == fromB.Claim.DocumentID {
		t.Error("distinct canonical sources must not share a DocumentID")
	}
}

// requiredUniquenessScope observes one Claim and declares Claim ID uniqueness
// as a proof the operation depends on, while leaving that completeness to the
// caller under test.
func requiredUniquenessScope(uniqueness Completeness, observed ...Claim) ValidationScope {
	return ValidationScope{
		Observed: observed,
		Complete: map[Check]Completeness{
			CheckClaimIDUniqueness:       uniqueness,
			CheckSupersessionCurrentness: CompletenessComplete,
		},
		Required: map[Check]bool{CheckClaimIDUniqueness: true},
	}
}

// F1 / F3: a current-authoritative answer must never bypass a proof the calling
// operation itself declared as required.
func TestCurrentUseEnforcesRequiredProofs(t *testing.T) {
	tests := []struct {
		name       string
		uniqueness Completeness
		want       ResolutionStatus
		wantReason ResolutionReason
	}{
		{
			// F1
			name:       "required uniqueness unproven blocks current use",
			uniqueness: CompletenessUnknown,
			want:       StatusResolutionUnresolved,
			wantReason: ReasonRequiredCheckUnproven,
		},
		{
			name:       "required uniqueness declared incomplete blocks current use",
			uniqueness: CompletenessIncomplete,
			want:       StatusResolutionUnresolved,
			wantReason: ReasonRequiredCheckUnproven,
		},
		{
			// F3 positive control
			name:       "required uniqueness proven resolves current",
			uniqueness: CompletenessComplete,
			want:       StatusResolvedCurrent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := requiredUniquenessScope(tt.uniqueness, doc("claim:a", "doc-1", "#one"))
			r, validation := Admit(scope, eligibleEverywhere)

			if validation.Status == StatusInvalid {
				t.Fatalf("scope unexpectedly invalid: %+v", validation.Issues)
			}
			if !r.Admitted() {
				t.Fatal("an UNRESOLVED validation must stay admitted for historical/diagnostic access")
			}

			got := r.Resolve(ClaimReference{ClaimID: "claim:a", Intent: IntentCurrentUse})
			if got.Status != tt.want {
				t.Fatalf("Resolve() status = %q, want %q (%s)", got.Status, tt.want, got.Detail)
			}
			if tt.wantReason != ReasonNone {
				if got.Reason != tt.wantReason {
					t.Errorf("Resolve() reason = %q, want %q", got.Reason, tt.wantReason)
				}
				if got.Check != CheckClaimIDUniqueness {
					t.Errorf("Resolve() blocking check = %q, want %q", got.Check, CheckClaimIDUniqueness)
				}
			}
			bound, ok := got.Binding()
			if ok != (tt.want == StatusResolvedCurrent) {
				t.Fatalf("Binding() = %q/%v for status %q", bound, ok, got.Status)
			}

			// The gate governs current authority only: exact historical
			// addressability is unaffected by it.
			hist := r.Resolve(ClaimReference{ClaimID: "claim:a", Intent: IntentHistoricalExact})
			if hist.Status != StatusResolvedExact {
				t.Errorf("Resolve(HISTORICAL_EXACT) = %q, want %q", hist.Status, StatusResolvedExact)
			}
		})
	}
}

// F2: the same gate governs default current retrieval, and its outcome stays
// distinguishable from both an incomplete-currentness answer and a resolved
// empty set.
func TestCurrentClaimsEnforcesRequiredProofs(t *testing.T) {
	blocked, _ := Admit(requiredUniquenessScope(CompletenessUnknown, doc("claim:a", "doc-1", "#one")), eligibleEverywhere)

	got := blocked.CurrentClaims()
	if got.Resolved() {
		t.Fatalf("CurrentClaims() = %q, want unresolved", got.Status)
	}
	if got.Reason != ReasonRequiredCheckUnproven {
		t.Errorf("reason = %q, want %q", got.Reason, ReasonRequiredCheckUnproven)
	}
	if got.Check != CheckClaimIDUniqueness {
		t.Errorf("blocking check = %q, want %q", got.Check, CheckClaimIDUniqueness)
	}
	if len(got.ClaimIDs) != 0 {
		t.Errorf("blocked set carried %v, want no claims", got.ClaimIDs)
	}

	allowed, _ := Admit(requiredUniquenessScope(CompletenessComplete, doc("claim:a", "doc-1", "#one")), eligibleEverywhere)
	resolved := allowed.CurrentClaims()
	if !resolved.Resolved() || len(resolved.ClaimIDs) != 1 || resolved.ClaimIDs[0] != "claim:a" {
		t.Fatalf("CurrentClaims() = %+v, want resolved [claim:a]", resolved)
	}

	// Three outcomes, three distinct answers.
	currentnessBlocked, _ := Admit(scopeWith(
		map[Check]Completeness{CheckSupersessionCurrentness: CompletenessIncomplete},
		doc("claim:a", "doc-1", "#one"),
	), eligibleEverywhere)
	if reason := currentnessBlocked.CurrentClaims().Reason; reason != ReasonCurrentnessIncomplete {
		t.Errorf("currentness reason = %q, want %q", reason, ReasonCurrentnessIncomplete)
	}
}

// F4: a check nobody required does not block current use just because its
// completeness was never declared.
func TestUnrequiredIncompleteCheckDoesNotBlockCurrentUse(t *testing.T) {
	scope := scopeWith(map[Check]Completeness{
		// CheckClaimIDUniqueness deliberately left UNKNOWN and unrequired.
		CheckSupersessionCurrentness: CompletenessComplete,
	}, doc("claim:a", "doc-1", "#one"))

	r, validation := Admit(scope, eligibleEverywhere)
	if validation.Status != StatusValid {
		t.Fatalf("validation = %q, want %q (issues: %+v)", validation.Status, StatusValid, validation.Issues)
	}
	if validation.IsProven(CheckClaimIDUniqueness) {
		t.Error("an undeclared completeness must not be proven")
	}

	got := r.Resolve(ClaimReference{ClaimID: "claim:a", Intent: IntentCurrentUse})
	if got.Status != StatusResolvedCurrent {
		t.Fatalf("Resolve() = %q/%q, want %q (%s)", got.Status, got.Reason, StatusResolvedCurrent, got.Detail)
	}
	if set := r.CurrentClaims(); !set.Resolved() {
		t.Errorf("CurrentClaims() = %+v, want resolved", set)
	}
}

// F5: requiring currentness keeps current use fail-closed, and the blocking
// check is identified whichever guard reports it.
func TestRequiredCurrentnessRemainsFailClosed(t *testing.T) {
	scope := ValidationScope{
		Observed: []Claim{doc("claim:a", "doc-1", "#one")},
		Required: map[Check]bool{CheckSupersessionCurrentness: true},
	}
	r, _ := Admit(scope, eligibleEverywhere)

	got := r.Resolve(ClaimReference{ClaimID: "claim:a", Intent: IntentCurrentUse})
	if got.Status != StatusResolutionUnresolved {
		t.Fatalf("Resolve() = %q/%q, want %q", got.Status, got.Reason, StatusResolutionUnresolved)
	}
	if got.Check != CheckSupersessionCurrentness {
		t.Errorf("blocking check = %q, want %q", got.Check, CheckSupersessionCurrentness)
	}
	if _, ok := got.Binding(); ok {
		t.Error("no current-authoritative binding is allowed here")
	}
	if set := r.CurrentClaims(); set.Resolved() {
		t.Errorf("CurrentClaims() = %+v, want unresolved", set)
	}
}

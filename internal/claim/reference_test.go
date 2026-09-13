package claim

import "testing"

func eligibleEverywhere(DocumentID) SourceEligibility { return Eligible }

func eligibility(m map[DocumentID]SourceEligibility) EligibilityFunc {
	return func(d DocumentID) SourceEligibility { return m[d] }
}

// currentAndSuperseded observes claim:a superseded by claim:b, plus an
// independent current claim:c.
func currentAndSuperseded() ValidationScope {
	return completeScope(
		doc("claim:a", "doc-1", "#one"),
		doc("claim:b", "doc-1", "#two", "claim:a"),
		doc("claim:c", "doc-1", "#three"),
	)
}

func TestResolveReference(t *testing.T) {
	scope := currentAndSuperseded()

	tests := []struct {
		name       string
		ref        ClaimReference
		want       ResolutionStatus
		wantReason ResolutionReason
		wantBound  bool
	}{
		{
			// T14
			name:      "current use reference to current claim",
			ref:       ClaimReference{ClaimID: "claim:c", Intent: IntentCurrentUse},
			want:      StatusResolvedCurrent,
			wantBound: true,
		},
		{
			// T15
			name:      "historical exact reference to current claim",
			ref:       ClaimReference{ClaimID: "claim:c", Intent: IntentHistoricalExact},
			want:      StatusResolvedExact,
			wantBound: true,
		},
		{
			// T16
			name:       "historical exact reference to superseded claim still resolves it",
			ref:        ClaimReference{ClaimID: "claim:a", Intent: IntentHistoricalExact},
			want:       StatusResolvedExact,
			wantReason: ReasonSuperseded,
			wantBound:  true,
		},
		{
			// T17
			name:       "current use reference to superseded claim is stale",
			ref:        ClaimReference{ClaimID: "claim:a", Intent: IntentCurrentUse},
			want:       StatusStaleReviewRequired,
			wantReason: ReasonSuperseded,
			wantBound:  false,
		},
		{
			name:       "reference without declared intent fails closed",
			ref:        ClaimReference{ClaimID: "claim:c"},
			want:       StatusResolutionUnresolved,
			wantReason: ReasonUnspecifiedIntent,
		},
		{
			name:       "reference to unobserved claim is unresolved",
			ref:        ClaimReference{ClaimID: "claim:absent", Intent: IntentCurrentUse},
			want:       StatusResolutionUnresolved,
			wantReason: ReasonClaimNotObserved,
		},
		{
			name:       "empty reference id is unresolved",
			ref:        ClaimReference{Intent: IntentCurrentUse},
			want:       StatusResolutionUnresolved,
			wantReason: ReasonEmptyClaimID,
		},
	}

	r, _ := Admit(scope, eligibleEverywhere)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.Resolve(tt.ref)
			if got.Status != tt.want {
				t.Fatalf("Resolve() status = %q, want %q (%s)", got.Status, tt.want, got.Detail)
			}
			if tt.wantReason != ReasonNone && got.Reason != tt.wantReason {
				t.Errorf("Resolve() reason = %q, want %q", got.Reason, tt.wantReason)
			}
			bound, ok := got.Binding()
			if ok != tt.wantBound {
				t.Fatalf("Binding() ok = %v, want %v", ok, tt.wantBound)
			}
			if ok && bound != tt.ref.ClaimID {
				t.Errorf("Binding() = %q, want the referenced identity %q", bound, tt.ref.ClaimID)
			}
		})
	}
}

// T18: supersession never silently returns the replacement as an accepted
// binding, not even when the replacement candidate is unique.
func TestSupersessionNeverSilentlyRebinds(t *testing.T) {
	r, _ := Admit(currentAndSuperseded(), eligibleEverywhere)

	got := r.Resolve(ClaimReference{ClaimID: "claim:a", Intent: IntentCurrentUse})
	if got.Status != StatusStaleReviewRequired {
		t.Fatalf("status = %q, want %q", got.Status, StatusStaleReviewRequired)
	}
	if bound, ok := got.Binding(); ok {
		t.Fatalf("stale reference produced binding %q, want none", bound)
	}
	if len(got.ReplacementCandidates) != 1 || got.ReplacementCandidates[0] != "claim:b" {
		t.Fatalf("ReplacementCandidates = %v, want [claim:b]", got.ReplacementCandidates)
	}
	if got.Claim == nil || got.Claim.ID != "claim:a" {
		t.Fatalf("resolution addressed %+v, want the referenced Claim claim:a", got.Claim)
	}
}

// A split leaves more than one candidate and still binds nothing automatically.
func TestSplitSupersessionOffersCandidatesWithoutBinding(t *testing.T) {
	scope := completeScope(
		doc("claim:a", "doc-1", "#one"),
		doc("claim:b", "doc-1", "#two", "claim:a"),
		doc("claim:c", "doc-1", "#three", "claim:a"),
	)
	r, _ := Admit(scope, eligibleEverywhere)

	got := r.Resolve(ClaimReference{ClaimID: "claim:a", Intent: IntentCurrentUse})
	if got.Status != StatusStaleReviewRequired {
		t.Fatalf("status = %q, want %q", got.Status, StatusStaleReviewRequired)
	}
	if _, ok := got.Binding(); ok {
		t.Error("split supersession must not produce an automatic binding")
	}
	if len(got.ReplacementCandidates) != 2 {
		t.Errorf("ReplacementCandidates = %v, want both replacements", got.ReplacementCandidates)
	}
}

// T19 / T20: the containing-source eligibility gate.
func TestSourceEligibilityGate(t *testing.T) {
	scope := completeScope(
		doc("claim:eligible", "doc-ok", "#one"),
		doc("claim:ineligible", "doc-bad", "#two"),
		doc("claim:unknown", "doc-unknown", "#three"),
	)
	r, _ := Admit(scope, eligibility(map[DocumentID]SourceEligibility{
		"doc-ok":  Eligible,
		"doc-bad": Ineligible,
	}))

	tests := []struct {
		name       string
		ref        ClaimReference
		want       ResolutionStatus
		wantReason ResolutionReason
	}{
		{
			// T20
			name:       "ineligible source is not current authoritative",
			ref:        ClaimReference{ClaimID: "claim:ineligible", Intent: IntentCurrentUse},
			want:       StatusNotCurrentAuthoritative,
			wantReason: ReasonSourceIneligible,
		},
		{
			// T19
			name:       "unknown eligibility is unresolved",
			ref:        ClaimReference{ClaimID: "claim:unknown", Intent: IntentCurrentUse},
			want:       StatusResolutionUnresolved,
			wantReason: ReasonSourceEligibilityUnknown,
		},
		{
			name: "eligible source resolves",
			ref:  ClaimReference{ClaimID: "claim:eligible", Intent: IntentCurrentUse},
			want: StatusResolvedCurrent,
		},
		{
			name: "historical exact binding survives an ineligible source",
			ref:  ClaimReference{ClaimID: "claim:ineligible", Intent: IntentHistoricalExact},
			want: StatusResolvedExact,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.Resolve(tt.ref)
			if got.Status != tt.want {
				t.Fatalf("Resolve() status = %q, want %q (%s)", got.Status, tt.want, got.Detail)
			}
			if tt.wantReason != ReasonNone && got.Reason != tt.wantReason {
				t.Errorf("Resolve() reason = %q, want %q", got.Reason, tt.wantReason)
			}
			if tt.want == StatusNotCurrentAuthoritative || tt.want == StatusResolutionUnresolved {
				if _, ok := got.Binding(); ok {
					t.Error("non-authoritative or unresolved outcome must not bind")
				}
			}
		})
	}
}

// A nil eligibility func means nothing was established, so current use fails
// closed rather than assuming eligibility.
func TestMissingEligibilityResolverFailsClosed(t *testing.T) {
	r, _ := Admit(currentAndSuperseded(), nil)

	got := r.Resolve(ClaimReference{ClaimID: "claim:c", Intent: IntentCurrentUse})
	if got.Status != StatusResolutionUnresolved || got.Reason != ReasonSourceEligibilityUnknown {
		t.Fatalf("Resolve() = %q/%q, want %q/%q",
			got.Status, got.Reason, StatusResolutionUnresolved, ReasonSourceEligibilityUnknown)
	}
	if got := r.Resolve(ClaimReference{ClaimID: "claim:c", Intent: IntentHistoricalExact}); got.Status != StatusResolvedExact {
		t.Errorf("historical exact status = %q, want %q", got.Status, StatusResolvedExact)
	}
}

// An ambiguous identity resolves to nothing rather than picking an observation.
func TestAmbiguousClaimIDIsUnresolved(t *testing.T) {
	scope := completeScope(
		doc("claim:a", "doc-1", "#one"),
		doc("claim:a", "doc-2", "#two"),
	)
	r, _ := Admit(scope, eligibleEverywhere)

	for _, intent := range []ReferenceIntent{IntentCurrentUse, IntentHistoricalExact} {
		got := r.Resolve(ClaimReference{ClaimID: "claim:a", Intent: intent})
		if got.Status != StatusResolutionUnresolved || got.Reason != ReasonAmbiguousClaimID {
			t.Errorf("Resolve(%s) = %q/%q, want %q/%q",
				intent, got.Status, got.Reason, StatusResolutionUnresolved, ReasonAmbiguousClaimID)
		}
	}
}

// Default current retrieval excludes superseded Claims and sources proven
// ineligible, once the observation is declared complete for currentness.
func TestCurrentClaimsExcludesSupersededAndIneligible(t *testing.T) {
	scope := completeScope(
		doc("claim:a", "doc-1", "#one"),
		doc("claim:b", "doc-1", "#two", "claim:a"),
		doc("claim:c", "doc-bad", "#three"),
	)
	r, _ := Admit(scope, eligibility(map[DocumentID]SourceEligibility{
		"doc-1":   Eligible,
		"doc-bad": Ineligible,
	}))

	got := r.CurrentClaims()
	if !got.Resolved() {
		t.Fatalf("CurrentClaims() = %q/%q, want resolved (%s)", got.Status, got.Reason, got.Detail)
	}
	if len(got.ClaimIDs) != 1 || got.ClaimIDs[0] != "claim:b" {
		t.Fatalf("CurrentClaims() = %v, want [claim:b]", got.ClaimIDs)
	}
}

// Unknown source eligibility among the candidates is not a silent exclusion.
func TestCurrentClaimsUnknownEligibilityIsUnresolved(t *testing.T) {
	scope := completeScope(doc("claim:a", "doc-unknown", "#one"))
	r, _ := Admit(scope, eligibility(map[DocumentID]SourceEligibility{}))

	got := r.CurrentClaims()
	if got.Resolved() || got.Reason != ReasonSourceEligibilityUnknown {
		t.Fatalf("CurrentClaims() = %q/%q, want unresolved/%q", got.Status, got.Reason, ReasonSourceEligibilityUnknown)
	}
}

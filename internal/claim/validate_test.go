package claim

import "testing"

func doc(id ClaimID, document DocumentID, locator Locator, supersedes ...ClaimID) Claim {
	return Claim{ID: id, DocumentID: document, Locator: locator, Supersedes: supersedes}
}

func completeScope(observed ...Claim) ValidationScope {
	return ValidationScope{
		Observed: observed,
		Complete: map[Check]Completeness{
			CheckClaimIDUniqueness:           CompletenessComplete,
			CheckSupersessionTargetExistence: CompletenessComplete,
		},
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		scope     ValidationScope
		want      Status
		wantCodes []IssueCode
	}{
		{
			// T1
			name:  "zero claims is valid",
			scope: completeScope(),
			want:  StatusValid,
		},
		{
			name:  "zero claims is valid even with undeclared completeness",
			scope: ValidationScope{},
			want:  StatusValid,
		},
		{
			// T2
			name: "duplicate claim id",
			scope: completeScope(
				doc("claim:a", "doc-1", "#one"),
				doc("claim:a", "doc-2", "#two"),
			),
			want:      StatusInvalid,
			wantCodes: []IssueCode{IssueDuplicateClaimID},
		},
		{
			name: "duplicate claim id inside incomplete scope is still invalid",
			scope: ValidationScope{
				Observed: []Claim{
					doc("claim:a", "doc-1", "#one"),
					doc("claim:a", "doc-2", "#two"),
				},
				Complete: map[Check]Completeness{CheckClaimIDUniqueness: CompletenessIncomplete},
			},
			want:      StatusInvalid,
			wantCodes: []IssueCode{IssueDuplicateClaimID},
		},
		{
			// T3
			name:      "empty claim id",
			scope:     completeScope(doc("", "doc-1", "#one")),
			want:      StatusInvalid,
			wantCodes: []IssueCode{IssueEmptyClaimID},
		},
		{
			// T4
			name: "duplicate locator in same document",
			scope: completeScope(
				doc("claim:a", "doc-1", "#same"),
				doc("claim:b", "doc-1", "#same"),
			),
			want:      StatusInvalid,
			wantCodes: []IssueCode{IssueDuplicateLocator},
		},
		{
			name: "same locator in different documents is valid",
			scope: completeScope(
				doc("claim:a", "doc-1", "#same"),
				doc("claim:b", "doc-2", "#same"),
			),
			want: StatusValid,
		},
		{
			name:      "missing locator is a declared-shape defect",
			scope:     completeScope(doc("claim:a", "doc-1", "")),
			want:      StatusInvalid,
			wantCodes: []IssueCode{IssueEmptyLocator},
		},
		{
			// T5
			name:      "self supersession",
			scope:     completeScope(doc("claim:a", "doc-1", "#one", "claim:a")),
			want:      StatusInvalid,
			wantCodes: []IssueCode{IssueSelfSupersession},
		},
		{
			// T6
			name: "direct cycle",
			scope: completeScope(
				doc("claim:a", "doc-1", "#one", "claim:b"),
				doc("claim:b", "doc-1", "#two", "claim:a"),
			),
			want:      StatusInvalid,
			wantCodes: []IssueCode{IssueSupersessionCycle},
		},
		{
			// T7
			name: "multi hop cycle",
			scope: completeScope(
				doc("claim:a", "doc-1", "#one", "claim:b"),
				doc("claim:b", "doc-1", "#two", "claim:c"),
				doc("claim:c", "doc-1", "#three", "claim:a"),
			),
			want:      StatusInvalid,
			wantCodes: []IssueCode{IssueSupersessionCycle},
		},
		{
			// T8
			name: "one to one supersession",
			scope: completeScope(
				doc("claim:a", "doc-1", "#one"),
				doc("claim:b", "doc-1", "#two", "claim:a"),
			),
			want: StatusValid,
		},
		{
			// T9
			name: "split supersession",
			scope: completeScope(
				doc("claim:a", "doc-1", "#one"),
				doc("claim:b", "doc-1", "#two", "claim:a"),
				doc("claim:c", "doc-1", "#three", "claim:a"),
			),
			want: StatusValid,
		},
		{
			// T10
			name: "merge supersession",
			scope: completeScope(
				doc("claim:a", "doc-1", "#one"),
				doc("claim:b", "doc-1", "#two"),
				doc("claim:c", "doc-1", "#three", "claim:a", "claim:b"),
			),
			want: StatusValid,
		},
		{
			// T11
			name: "chain supersession",
			scope: completeScope(
				doc("claim:a", "doc-1", "#one"),
				doc("claim:b", "doc-1", "#two", "claim:a"),
				doc("claim:c", "doc-1", "#three", "claim:b"),
			),
			want: StatusValid,
		},
		{
			// T12
			name:      "missing target with complete scope is invalid",
			scope:     completeScope(doc("claim:b", "doc-1", "#two", "claim:a")),
			want:      StatusInvalid,
			wantCodes: []IssueCode{IssueMissingSupersessionTarget},
		},
		{
			// T13
			name: "missing target with incomplete scope is unresolved",
			scope: ValidationScope{
				Observed: []Claim{doc("claim:b", "doc-1", "#two", "claim:a")},
				Complete: map[Check]Completeness{
					CheckSupersessionTargetExistence: CompletenessIncomplete,
				},
			},
			want:      StatusUnresolved,
			wantCodes: []IssueCode{IssueIncompleteScope},
		},
		{
			// T13, unknown completeness fails closed exactly like incomplete.
			name: "missing target with unknown completeness is unresolved",
			scope: ValidationScope{
				Observed: []Claim{doc("claim:b", "doc-1", "#two", "claim:a")},
			},
			want:      StatusUnresolved,
			wantCodes: []IssueCode{IssueIncompleteScope},
		},
		{
			name:      "empty supersession target",
			scope:     completeScope(doc("claim:b", "doc-1", "#two", "")),
			want:      StatusInvalid,
			wantCodes: []IssueCode{IssueEmptySupersessionTarget},
		},
		{
			name: "proven defect outranks an unresolved check",
			scope: ValidationScope{
				Observed: []Claim{
					doc("claim:a", "doc-1", "#one"),
					doc("claim:a", "doc-2", "#two"),
					doc("claim:c", "doc-3", "#three", "claim:absent"),
				},
			},
			want:      StatusInvalid,
			wantCodes: []IssueCode{IssueDuplicateClaimID, IssueIncompleteScope},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Validate(tt.scope)
			if got.Status != tt.want {
				t.Fatalf("Validate() status = %q, want %q (issues: %+v)", got.Status, tt.want, got.Issues)
			}
			for _, code := range tt.wantCodes {
				if !got.HasCode(code) {
					t.Errorf("Validate() missing expected issue %q (issues: %+v)", code, got.Issues)
				}
			}
			if tt.want == StatusValid && len(got.Issues) != 0 {
				t.Errorf("Validate() reported issues for a valid scope: %+v", got.Issues)
			}
		})
	}
}

// An incomplete observation is never proof of workspace-global uniqueness, even
// when it happens to contain no duplicate.
func TestValidateDoesNotClaimGlobalUniquenessFromPartialScope(t *testing.T) {
	observed := []Claim{doc("claim:a", "doc-1", "#one")}

	partial := Validate(ValidationScope{
		Observed: observed,
		Complete: map[Check]Completeness{CheckClaimIDUniqueness: CompletenessIncomplete},
	})
	if partial.Status != StatusValid {
		t.Fatalf("status = %q, want %q", partial.Status, StatusValid)
	}
	if partial.GlobalUniquenessEstablished {
		t.Error("incomplete scope must not establish workspace-global uniqueness")
	}

	unknown := Validate(ValidationScope{Observed: observed})
	if unknown.GlobalUniquenessEstablished {
		t.Error("unknown completeness must not establish workspace-global uniqueness")
	}

	complete := Validate(completeScope(observed...))
	if !complete.GlobalUniquenessEstablished {
		t.Error("scope declared complete with no duplicate should establish uniqueness")
	}
}

// Results must not depend on the order the material was observed in.
func TestValidateIsOrderIndependent(t *testing.T) {
	a := doc("claim:a", "doc-1", "#one", "claim:b")
	b := doc("claim:b", "doc-1", "#two", "claim:c")
	c := doc("claim:c", "doc-1", "#three", "claim:a")

	forward := Validate(completeScope(a, b, c))
	reverse := Validate(completeScope(c, b, a))

	if forward.Status != StatusInvalid || reverse.Status != StatusInvalid {
		t.Fatalf("statuses = %q / %q, want both %q", forward.Status, reverse.Status, StatusInvalid)
	}
	if len(forward.Issues) != len(reverse.Issues) {
		t.Fatalf("issue counts differ: %d vs %d", len(forward.Issues), len(reverse.Issues))
	}
	for i := range forward.Issues {
		if forward.Issues[i].Code != reverse.Issues[i].Code {
			t.Errorf("issue %d: %q vs %q", i, forward.Issues[i].Code, reverse.Issues[i].Code)
		}
	}
	for _, cycle := range forward.IssuesWithCode(IssueSupersessionCycle) {
		if len(cycle.Related) != 3 {
			t.Errorf("cycle members = %v, want 3 members", cycle.Related)
		}
	}
}

// A superseded Claim is never deleted: it stays in the observed set and remains
// addressable after supersession.
func TestSupersededClaimIsPreserved(t *testing.T) {
	scope := completeScope(
		doc("claim:a", "doc-1", "#one"),
		doc("claim:b", "doc-1", "#two", "claim:a"),
	)
	if got := Validate(scope); got.Status != StatusValid {
		t.Fatalf("status = %q, want %q", got.Status, StatusValid)
	}

	r := NewResolver(scope, eligibleEverywhere)
	if _, ok := r.Lookup("claim:a"); !ok {
		t.Error("superseded Claim must remain observable")
	}
	if got := r.SupersededBy("claim:a"); len(got) != 1 || got[0] != "claim:b" {
		t.Errorf("SupersededBy(claim:a) = %v, want [claim:b]", got)
	}
	if got := r.SupersededBy("claim:b"); len(got) != 0 {
		t.Errorf("SupersededBy(claim:b) = %v, want empty", got)
	}
}

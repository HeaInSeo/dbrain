package claim

import "testing"

// T21: the same Claim identity may carry a different observed content digest
// after a meaning-preserving edit. The identity does not change with it.
func TestSameClaimIDMayChangeContentDigest(t *testing.T) {
	before := Claim{ID: "claim:a", DocumentID: "doc-1", Locator: "#one", ContentDigest: "sha256:aaa"}
	after := before
	after.ContentDigest = "sha256:bbb"

	if !before.IdentityEquals(after) {
		t.Error("a changed content digest must not change Claim identity")
	}
	if before.ObservedContentEquals(after) {
		t.Error("differing digests must be reported as differing observed content")
	}

	if got := Validate(completeScope(before)); got.Status != StatusValid {
		t.Fatalf("before status = %q, want %q", got.Status, StatusValid)
	}
	if got := Validate(completeScope(after)); got.Status != StatusValid {
		t.Fatalf("after status = %q, want %q", got.Status, StatusValid)
	}
}

// T22: moving a document, or restructuring where the Claim sits, does not by
// itself create a new Claim identity.
func TestLocationChangeDoesNotCreateNewClaimID(t *testing.T) {
	before := Claim{ID: "claim:a", DocumentID: "docs/old/rule.md", Locator: "#old-anchor", ContentDigest: "sha256:aaa"}
	after := Claim{ID: "claim:a", DocumentID: "docs/new/rule.md", Locator: "#new-anchor", ContentDigest: "sha256:aaa"}

	if !before.IdentityEquals(after) {
		t.Error("path or locator changes must not change Claim identity")
	}
	if !before.ObservedContentEquals(after) {
		t.Error("unchanged content digest must compare equal across a move")
	}

	moved := Validate(completeScope(after))
	if moved.Status != StatusValid {
		t.Fatalf("status after move = %q, want %q (issues: %+v)", moved.Status, StatusValid, moved.Issues)
	}

	// The move produces no second identity: references keep resolving to the
	// one Claim, in its new location.
	r, _ := Admit(completeScope(after), eligibleEverywhere)
	got := r.Resolve(ClaimReference{ClaimID: "claim:a", Intent: IntentCurrentUse})
	if got.Status != StatusResolvedCurrent {
		t.Fatalf("status = %q, want %q", got.Status, StatusResolvedCurrent)
	}
	if got.Claim == nil || got.Claim.DocumentID != "docs/new/rule.md" {
		t.Errorf("resolved document = %+v, want the new location", got.Claim)
	}
}

// An unobserved digest proves nothing about content equality.
func TestUnobservedDigestIsNotContentEquality(t *testing.T) {
	a := Claim{ID: "claim:a", DocumentID: "doc-1", Locator: "#one"}
	b := a

	if !a.IdentityEquals(b) {
		t.Error("identity comparison must ignore the missing digest")
	}
	if a.ObservedContentEquals(b) {
		t.Error("two unobserved digests must not be reported as equal content")
	}
}

// An empty ID is never an identity, so it never compares equal to anything.
func TestEmptyIDIsNeverAnIdentity(t *testing.T) {
	a := Claim{DocumentID: "doc-1", Locator: "#one"}
	if a.IdentityEquals(a) {
		t.Error("a Claim without an ID must not compare identity-equal")
	}
}

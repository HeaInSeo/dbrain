// Package claim implements the serialization-independent core of the DBrain
// Claim model contract (docs/contracts/claim-model.md).
//
// Boundaries deliberately preserved by this package:
//
//	Claim core != Markdown parser
//	    Locators are validated for declared identity/shape/uniqueness only.
//	    Anchor-to-prose resolution belongs to the serialization/parser layer,
//	    which is still OPEN in PKG-M0.
//
//	Claim core != lifecycle / authority owner
//	    Containing-source eligibility for current authoritative use is an input
//	    (SourceEligibility), never a vocabulary owned here.
//
//	Claim ID != authority
//	    Addressability says nothing about who owns the subject/scope.
//
//	Partial observation != complete workspace truth
//	    Every check runs against an explicit ValidationScope whose completeness
//	    is declared per check and defaults to unknown (fail-closed).
package claim

// ClaimID is the immutable, opaque semantic identity of a Claim.
//
// It is independent of file path, document title, subject name and wording, so
// moving or rewording the containing material never changes it. It is not a
// content identity: see Claim.ContentDigest.
type ClaimID string

// DocumentID identifies the containing human-readable canonical source.
type DocumentID string

// Locator is the declared, serialization-independent address of the Claim's
// semantic span inside its document.
//
// This package treats a Locator as an opaque token: it validates that one is
// declared and that declarations do not collide inside a document. It does not
// interpret, parse or resolve it.
type Locator string

// Claim is an optional stable semantic address to a bounded statement inside an
// authoritative document. Documents with zero Claims are valid and expected:
// minting is demand-driven, not coverage-driven.
type Claim struct {
	// ID is the immutable semantic identity.
	ID ClaimID

	// DocumentID is the containing canonical source.
	DocumentID DocumentID

	// Locator is the declared address of the semantic span.
	Locator Locator

	// Supersedes lists prior Claim identities this Claim explicitly replaces.
	// Supersession is directional, explicit and never a silent rebind of
	// inbound references. The inverse view is derived, never hand-maintained.
	Supersedes []ClaimID

	// ContentDigest is the observed content/revision identity of the resolved
	// span, when it has been observed. It is intentionally optional and is NOT
	// the Claim identity.
	ContentDigest string
}

// IdentityEquals reports whether two observations denote the same Claim
// identity. Only the immutable ID participates: a moved document, a changed
// locator or reworded prose does not by itself create a new Claim.
func (c Claim) IdentityEquals(other Claim) bool {
	return c.ID != "" && c.ID == other.ID
}

// ObservedContentEquals reports whether two observations of the same Claim
// identity carry the same observed content digest. It is false for different
// identities, and false when either digest is unobserved, since an unobserved
// digest proves nothing about content.
func (c Claim) ObservedContentEquals(other Claim) bool {
	if !c.IdentityEquals(other) {
		return false
	}
	if c.ContentDigest == "" || other.ContentDigest == "" {
		return false
	}
	return c.ContentDigest == other.ContentDigest
}

// SourceEligibility is the upstream determination of whether a Claim's
// containing source may currently carry authoritative knowledge.
//
// This package consumes it and never derives it. It deliberately does not model
// document lifecycle vocabulary (CURRENT / SUPERSEDED / HISTORICAL / ...);
// that belongs to the lifecycle and authority contracts.
type SourceEligibility string

const (
	// EligibilityUnknown means eligibility could not be established. It is the
	// zero value so that an unset input fails closed to UNRESOLVED.
	EligibilityUnknown SourceEligibility = ""

	// Eligible means the containing source may carry current authoritative use.
	Eligible SourceEligibility = "ELIGIBLE"

	// Ineligible means the containing source may not surface as current
	// authoritative knowledge.
	Ineligible SourceEligibility = "INELIGIBLE"
)

// EligibilityFunc reports the containing-source eligibility for a document.
// A nil EligibilityFunc, or one returning EligibilityUnknown, makes current
// authoritative resolution fail closed to UNRESOLVED.
type EligibilityFunc func(DocumentID) SourceEligibility

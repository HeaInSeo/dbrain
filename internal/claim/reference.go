package claim

import "fmt"

// ReferenceIntent is the use intent carried by a Claim reference edge.
//
// The Claim ID does not encode intent: the consumer declares it, because the
// same identity means different things to a consumer that needs a rule to still
// hold and to one that deliberately cites history.
type ReferenceIntent string

const (
	// IntentUnspecified is the zero value. Resolution fails closed to
	// UNRESOLVED rather than guessing which intent the consumer meant.
	IntentUnspecified ReferenceIntent = ""

	// IntentCurrentUse means the consumer expects the referenced Claim to
	// remain acceptable for present use.
	IntentCurrentUse ReferenceIntent = "CURRENT_USE"

	// IntentHistoricalExact means the consumer deliberately refers to that
	// exact historical Claim identity.
	IntentHistoricalExact ReferenceIntent = "HISTORICAL_EXACT"
)

// ClaimReference is a durable reference edge from some consumer to a Claim.
type ClaimReference struct {
	ClaimID ClaimID
	Intent  ReferenceIntent

	// Consumer identifies the referring artifact, for provenance. It does not
	// affect resolution in this core.
	Consumer string
}

// ResolutionStatus is the outcome of resolving one ClaimReference.
type ResolutionStatus string

const (
	// StatusResolvedCurrent means the referenced Claim is bound and acceptable
	// for current authoritative use.
	StatusResolvedCurrent ResolutionStatus = "RESOLVED_CURRENT"

	// StatusResolvedExact means the referenced historical identity is bound
	// exactly, independent of supersession.
	StatusResolvedExact ResolutionStatus = "RESOLVED_EXACT"

	// StatusStaleReviewRequired means a current-use reference points at a
	// superseded Claim. It is not an automatic rebind: a human/authority review
	// rebinds explicitly, and only if semantically correct.
	StatusStaleReviewRequired ResolutionStatus = "STALE_REVIEW_REQUIRED"

	// StatusNotCurrentAuthoritative means the containing source is ineligible
	// for current authoritative use, so the Claim must not surface as current
	// authoritative knowledge.
	StatusNotCurrentAuthoritative ResolutionStatus = "NOT_CURRENT_AUTHORITATIVE"

	// StatusResolutionUnresolved means resolution requires information the
	// supplied inputs did not establish.
	StatusResolutionUnresolved ResolutionStatus = "UNRESOLVED"
)

// ResolutionReason is a machine-readable reason for a resolution outcome.
type ResolutionReason string

const (
	ReasonNone                     ResolutionReason = ""
	ReasonEmptyClaimID             ResolutionReason = "EMPTY_CLAIM_ID"
	ReasonUnspecifiedIntent        ResolutionReason = "UNSPECIFIED_REFERENCE_INTENT"
	ReasonClaimNotObserved         ResolutionReason = "CLAIM_NOT_OBSERVED"
	ReasonAmbiguousClaimID         ResolutionReason = "AMBIGUOUS_CLAIM_ID"
	ReasonSuperseded               ResolutionReason = "SUPERSEDED"
	ReasonSourceIneligible         ResolutionReason = "SOURCE_INELIGIBLE"
	ReasonSourceEligibilityUnknown ResolutionReason = "SOURCE_ELIGIBILITY_UNKNOWN"
)

// Resolution is the outcome of resolving one ClaimReference.
type Resolution struct {
	Reference ClaimReference
	Status    ResolutionStatus
	Reason    ResolutionReason

	// Claim is the exact observed Claim the reference addresses, when it was
	// observed. It is never a replacement Claim.
	Claim *Claim

	// ReplacementCandidates are the Claims that declare supersession of the
	// referenced identity. They are informational only.
	//
	//	candidate replacement != accepted rebind
	//
	// Binding never returns one of these, whatever their number.
	ReplacementCandidates []ClaimID

	Detail string
}

// Binding returns the Claim identity this resolution actually binds to, and
// whether a binding exists at all.
//
// It returns only the referenced identity. A stale current-use reference
// returns no binding even when exactly one replacement candidate exists, which
// is what keeps supersession from silently rebinding accepted artifacts.
func (r Resolution) Binding() (ClaimID, bool) {
	switch r.Status {
	case StatusResolvedCurrent, StatusResolvedExact:
		return r.Reference.ClaimID, true
	default:
		return "", false
	}
}

// Resolver answers reference resolution over one observed scope.
//
// It owns no lifecycle state: containing-source eligibility arrives through an
// EligibilityFunc supplied by the caller, and an unknown answer fails closed.
type Resolver struct {
	byID         map[ClaimID]Claim
	occurrences  map[ClaimID]int
	supersededBy map[ClaimID][]ClaimID
	eligibility  EligibilityFunc
}

// NewResolver builds a Resolver over the observed material in scope.
//
// A nil eligibility func reports EligibilityUnknown for every document, so
// current authoritative resolution fails closed to UNRESOLVED.
func NewResolver(scope ValidationScope, eligibility EligibilityFunc) *Resolver {
	r := &Resolver{
		byID:         make(map[ClaimID]Claim, len(scope.Observed)),
		occurrences:  make(map[ClaimID]int, len(scope.Observed)),
		supersededBy: make(map[ClaimID][]ClaimID),
		eligibility:  eligibility,
	}
	for _, c := range scope.Observed {
		if c.ID == "" {
			continue
		}
		r.occurrences[c.ID]++
		if _, ok := r.byID[c.ID]; !ok {
			r.byID[c.ID] = c
		}
	}
	// superseded_by is derived from the declared supersedes edges, never
	// hand-maintained as a second copy of the relationship.
	for _, c := range scope.Observed {
		if c.ID == "" {
			continue
		}
		for _, target := range c.Supersedes {
			if target == "" || target == c.ID {
				continue
			}
			if containsID(r.supersededBy[target], c.ID) {
				continue
			}
			r.supersededBy[target] = append(r.supersededBy[target], c.ID)
		}
	}
	for target := range r.supersededBy {
		sortIDs(r.supersededBy[target])
	}
	return r
}

// SupersededBy returns the observed Claims that declare supersession of id,
// sorted. The result is derived, and empty for an unsuperseded Claim.
func (r *Resolver) SupersededBy(id ClaimID) []ClaimID {
	return append([]ClaimID(nil), r.supersededBy[id]...)
}

// Lookup returns the observed Claim for an identity.
func (r *Resolver) Lookup(id ClaimID) (Claim, bool) {
	c, ok := r.byID[id]
	return c, ok
}

// CurrentClaimIDs returns the observed Claims eligible for default current
// retrieval: not superseded by anything observed, and in a source declared
// eligible. Unknown eligibility is excluded, since it fails closed.
//
// Superseded Claims are excluded here but remain resolvable through
// IntentHistoricalExact; they are never deleted.
func (r *Resolver) CurrentClaimIDs() []ClaimID {
	var out []ClaimID
	for id, c := range r.byID {
		if len(r.supersededBy[id]) > 0 {
			continue
		}
		if r.occurrences[id] > 1 {
			continue
		}
		if r.sourceEligibility(c.DocumentID) != Eligible {
			continue
		}
		out = append(out, id)
	}
	sortIDs(out)
	return out
}

// Resolve resolves one reference edge according to its declared intent.
func (r *Resolver) Resolve(ref ClaimReference) Resolution {
	res := Resolution{Reference: ref}

	if ref.ClaimID == "" {
		res.Status = StatusResolutionUnresolved
		res.Reason = ReasonEmptyClaimID
		res.Detail = "reference declares no Claim ID"
		return res
	}
	if r.occurrences[ref.ClaimID] > 1 {
		res.Status = StatusResolutionUnresolved
		res.Reason = ReasonAmbiguousClaimID
		res.Detail = fmt.Sprintf("Claim ID %q was observed %d times, so the reference is ambiguous",
			ref.ClaimID, r.occurrences[ref.ClaimID])
		return res
	}
	c, ok := r.byID[ref.ClaimID]
	if !ok {
		res.Status = StatusResolutionUnresolved
		res.Reason = ReasonClaimNotObserved
		res.Detail = fmt.Sprintf("Claim ID %q was not observed in the supplied scope", ref.ClaimID)
		return res
	}
	res.Claim = &c
	res.ReplacementCandidates = r.SupersededBy(ref.ClaimID)

	switch ref.Intent {
	case IntentHistoricalExact:
		// A historical exact reference binds the identity itself. It is not a
		// current-authority surface, so supersession and source eligibility do
		// not revoke it.
		res.Status = StatusResolvedExact
		if len(res.ReplacementCandidates) > 0 {
			res.Reason = ReasonSuperseded
			res.Detail = fmt.Sprintf("Claim %q is superseded but remains bound exactly for historical reference",
				ref.ClaimID)
		}
		return res

	case IntentCurrentUse:
		switch r.sourceEligibility(c.DocumentID) {
		case Ineligible:
			res.Status = StatusNotCurrentAuthoritative
			res.Reason = ReasonSourceIneligible
			res.Detail = fmt.Sprintf("containing source %q is ineligible for current authoritative use",
				c.DocumentID)
			return res
		case Eligible:
			// continue
		default:
			res.Status = StatusResolutionUnresolved
			res.Reason = ReasonSourceEligibilityUnknown
			res.Detail = fmt.Sprintf("eligibility of containing source %q could not be established",
				c.DocumentID)
			return res
		}

		if len(res.ReplacementCandidates) > 0 {
			res.Status = StatusStaleReviewRequired
			res.Reason = ReasonSuperseded
			res.Detail = fmt.Sprintf("Claim %q is superseded; rebinding requires explicit authority review, candidates: %v",
				ref.ClaimID, res.ReplacementCandidates)
			return res
		}
		res.Status = StatusResolvedCurrent
		return res

	default:
		res.Status = StatusResolutionUnresolved
		res.Reason = ReasonUnspecifiedIntent
		res.Detail = "reference declares no use intent"
		return res
	}
}

// sourceEligibility consults the supplied resolver, defaulting to unknown.
func (r *Resolver) sourceEligibility(doc DocumentID) SourceEligibility {
	if r.eligibility == nil {
		return EligibilityUnknown
	}
	switch e := r.eligibility(doc); e {
	case Eligible, Ineligible:
		return e
	default:
		return EligibilityUnknown
	}
}

func containsID(ids []ClaimID, id ClaimID) bool {
	for _, existing := range ids {
		if existing == id {
			return true
		}
	}
	return false
}

package claim

import (
	"errors"
	"fmt"
)

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

// Errors reported by the observation-inspection API. They exist so an ambiguous
// observation is signalled rather than silently resolved to one arbitrary
// observation or to a union of several.
var (
	// ErrClaimNotObserved means the identity was absent from the scope.
	ErrClaimNotObserved = errors.New("claim: identity not observed in supplied scope")

	// ErrAmbiguousClaimID means the identity was observed more than once, or a
	// derived relation would have to be synthesised across duplicate
	// observations of one identity.
	ErrAmbiguousClaimID = errors.New("claim: identity observed more than once")
)

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
	ReasonNone              ResolutionReason = ""
	ReasonEmptyClaimID      ResolutionReason = "EMPTY_CLAIM_ID"
	ReasonUnspecifiedIntent ResolutionReason = "UNSPECIFIED_REFERENCE_INTENT"
	ReasonClaimNotObserved  ResolutionReason = "CLAIM_NOT_OBSERVED"
	ReasonAmbiguousClaimID  ResolutionReason = "AMBIGUOUS_CLAIM_ID"
	ReasonSuperseded        ResolutionReason = "SUPERSEDED"
	ReasonSourceIneligible  ResolutionReason = "SOURCE_INELIGIBLE"

	ReasonSourceEligibilityUnknown ResolutionReason = "SOURCE_ELIGIBILITY_UNKNOWN"

	// ReasonCurrentnessIncomplete means the observation was never declared
	// complete for CheckSupersessionCurrentness, so an unobserved superseder
	// cannot be ruled out.
	ReasonCurrentnessIncomplete ResolutionReason = "SUPERSESSION_CURRENTNESS_INCOMPLETE"

	// ReasonRequiredCheckUnproven means the requesting operation declared a
	// check as required and this validation could not establish its proof, so a
	// current-authoritative answer would bypass the caller's own declaration.
	ReasonRequiredCheckUnproven ResolutionReason = "REQUIRED_CHECK_UNPROVEN"

	// ReasonScopeStructurallyInvalid means validation proved a defect in the
	// observed structure, so the scope is not admissible as an input to current
	// authoritative resolution.
	ReasonScopeStructurallyInvalid ResolutionReason = "SCOPE_STRUCTURALLY_INVALID"

	// ReasonAmbiguousSupersession means the derived superseder set for the
	// referenced identity would have to be synthesised across duplicate
	// observations.
	ReasonAmbiguousSupersession ResolutionReason = "AMBIGUOUS_SUPERSESSION"
)

// Resolution is the outcome of resolving one ClaimReference.
type Resolution struct {
	Reference ClaimReference
	Status    ResolutionStatus
	Reason    ResolutionReason

	// Claim is the exact observed Claim the reference addresses, when it was
	// observed. It is never a replacement Claim.
	Claim *Claim

	// Check names the completeness-sensitive check that determined this
	// outcome, when one applies: the unproven required check, or the
	// currentness proof that was missing.
	Check Check

	// SourceEligibility is the containing source's eligibility for current
	// authoritative use, as provenance. It is reported even when the status is
	// RESOLVED_EXACT, because
	//
	//	historically addressable != currently authoritative
	//
	// and a consumer must be able to tell the two apart.
	SourceEligibility SourceEligibility

	// ReplacementCandidates are the directly observed superseders of the
	// referenced identity: in A → B → C, the candidates for A are [B] only.
	// This core never walks the chain to nominate a canonical replacement.
	// They are informational:
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

// CurrentSet is the outcome of asking for the current Claims in a scope.
//
// It exists so that
//
//	resolved, and the current set is empty
//
// stays distinguishable from
//
//	the current set cannot be determined from this observation
//
// rather than collapsing both into an empty slice.
type CurrentSet struct {
	// Status is StatusResolvedCurrent or StatusResolutionUnresolved.
	Status ResolutionStatus
	Reason ResolutionReason

	// Check names the check that blocked the answer, when one did.
	Check Check

	// ClaimIDs is meaningful only when Status is StatusResolvedCurrent.
	ClaimIDs []ClaimID

	Detail string
}

// Resolved reports whether the current set was actually determined.
func (c CurrentSet) Resolved() bool { return c.Status == StatusResolvedCurrent }

// Resolver answers reference resolution over one admitted observation.
//
// It owns no lifecycle state: containing-source eligibility arrives through an
// EligibilityFunc supplied by the caller, and an unknown answer fails closed.
type Resolver struct {
	scope       ValidationScope
	validation  Result
	admitted    bool
	byID        map[ClaimID]Claim
	occurrences map[ClaimID]int

	// supersededBy is derived from the declared supersedes edges, never
	// hand-maintained as a second copy of the relationship.
	supersededBy map[ClaimID][]ClaimID

	// ambiguousSuperseders marks identities whose derived superseder set would
	// have to be synthesised across duplicate observations of one Claim ID.
	ambiguousSuperseders map[ClaimID]bool

	eligibility EligibilityFunc
}

// Admit validates a scope and returns a Resolver over it together with the
// validation Result.
//
// A scope with a proven structural defect is not admitted:
//
//	proven structurally invalid scope != admissible current-resolution input
//
// A non-admitted Resolver never produces a current authoritative binding. It
// still answers IntentHistoricalExact and the inspection API, which are
// diagnostic and historical surfaces rather than current-authority ones.
//
// A nil eligibility func reports EligibilityUnknown for every document, so
// current authoritative resolution fails closed to UNRESOLVED.
func Admit(scope ValidationScope, eligibility EligibilityFunc) (*Resolver, Result) {
	validation := Validate(scope)

	r := &Resolver{
		scope:                scope,
		validation:           validation,
		admitted:             validation.Status != StatusInvalid,
		byID:                 make(map[ClaimID]Claim, len(scope.Observed)),
		occurrences:          make(map[ClaimID]int, len(scope.Observed)),
		supersededBy:         make(map[ClaimID][]ClaimID),
		ambiguousSuperseders: make(map[ClaimID]bool),
		eligibility:          eligibility,
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
	for _, c := range scope.Observed {
		if c.ID == "" {
			continue
		}
		for _, target := range c.Supersedes {
			if target == "" || target == c.ID {
				continue
			}
			// An edge declared by a duplicated identity cannot be attributed to
			// a single observation, so the target's derived set is ambiguous
			// rather than a union presented as one relation.
			if r.occurrences[c.ID] > 1 {
				r.ambiguousSuperseders[target] = true
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
	return r, validation
}

// Admitted reports whether the observed structure was admissible as an input to
// current authoritative resolution.
//
// Admission covers structure only. A scope may be admitted while the validation
// aggregate is UNRESOLVED, so that historical and diagnostic access survives an
// undeterminable check; current-authority surfaces additionally enforce the
// caller's own required proofs through unprovenRequiredCheck.
func (r *Resolver) Admitted() bool { return r.admitted }

// unprovenRequiredCheck returns the first check the requesting operation
// declared as required whose proof this validation did not establish.
//
// It derives the answer from the scope's declared requirements and the
// validation's proof bits rather than recomputing any check's meaning. Current
// authoritative use must never succeed while such a check exists: that would
// bypass the caller's own declaration of what had to be proven.
func (r *Resolver) unprovenRequiredCheck() (Check, bool) {
	for _, c := range r.scope.RequiredChecks() {
		if !r.validation.IsProven(c) {
			return c, true
		}
	}
	return "", false
}

// Validation returns the validation result the Resolver was admitted on.
func (r *Resolver) Validation() Result { return r.validation }

// Lookup returns the single observation of an identity.
//
// It reports ErrAmbiguousClaimID when the identity was observed more than once:
// the first observation never silently wins.
func (r *Resolver) Lookup(id ClaimID) (Claim, error) {
	if n := r.occurrences[id]; n > 1 {
		return Claim{}, fmt.Errorf("%w: %q observed %d times", ErrAmbiguousClaimID, id, n)
	}
	c, ok := r.byID[id]
	if !ok {
		return Claim{}, fmt.Errorf("%w: %q", ErrClaimNotObserved, id)
	}
	return c, nil
}

// SupersededBy returns the directly observed superseders of id, sorted.
//
// It reports ErrAmbiguousClaimID when id itself was observed more than once, or
// when an edge pointing at id was declared by a duplicated identity, since the
// derived set would then be a synthetic union across distinct observations.
func (r *Resolver) SupersededBy(id ClaimID) ([]ClaimID, error) {
	if n := r.occurrences[id]; n > 1 {
		return nil, fmt.Errorf("%w: %q observed %d times", ErrAmbiguousClaimID, id, n)
	}
	if r.ambiguousSuperseders[id] {
		return nil, fmt.Errorf("%w: superseders of %q were declared by a duplicated identity",
			ErrAmbiguousClaimID, id)
	}
	return append([]ClaimID(nil), r.supersededBy[id]...), nil
}

// CurrentClaims returns the Claims eligible for default current retrieval: not
// superseded, and in a source declared eligible.
//
// It fails closed. Unless the scope was admitted and declared complete for
// CheckSupersessionCurrentness, the answer is UNRESOLVED rather than an empty
// or partially filtered set, because an unobserved superseder cannot be ruled
// out. Unknown source eligibility among the candidates is likewise UNRESOLVED.
//
// Superseded Claims are excluded here but remain resolvable through
// IntentHistoricalExact; they are never deleted.
func (r *Resolver) CurrentClaims() CurrentSet {
	if !r.admitted {
		return CurrentSet{
			Status: StatusResolutionUnresolved,
			Reason: ReasonScopeStructurallyInvalid,
			Detail: "observed structure is proven invalid, so no current set can be established",
		}
	}
	if c, unproven := r.unprovenRequiredCheck(); unproven {
		return CurrentSet{
			Status: StatusResolutionUnresolved,
			Reason: ReasonRequiredCheckUnproven,
			Check:  c,
			Detail: fmt.Sprintf("check %s is required by this operation but was not proven (completeness %q)",
				c, r.scope.CompletenessFor(c)),
		}
	}
	if !r.scope.IsCompleteFor(CheckSupersessionCurrentness) {
		return CurrentSet{
			Status: StatusResolutionUnresolved,
			Reason: ReasonCurrentnessIncomplete,
			Check:  CheckSupersessionCurrentness,
			Detail: fmt.Sprintf("scope completeness for %s is %q, so unobserved superseders cannot be ruled out",
				CheckSupersessionCurrentness, r.scope.CompletenessFor(CheckSupersessionCurrentness)),
		}
	}

	var out []ClaimID
	for id, c := range r.byID {
		if len(r.supersededBy[id]) > 0 {
			continue
		}
		switch r.sourceEligibility(c.DocumentID) {
		case Eligible:
			out = append(out, id)
		case Ineligible:
			// A determinate exclusion.
		default:
			return CurrentSet{
				Status: StatusResolutionUnresolved,
				Reason: ReasonSourceEligibilityUnknown,
				Detail: fmt.Sprintf("eligibility of containing source %q could not be established",
					c.DocumentID),
			}
		}
	}
	sortIDs(out)
	return CurrentSet{Status: StatusResolvedCurrent, ClaimIDs: out}
}

// Resolve resolves one reference edge according to its declared intent.
func (r *Resolver) Resolve(ref ClaimReference) Resolution {
	res := Resolution{Reference: ref}

	if ref.ClaimID == "" {
		return unresolved(res, ReasonEmptyClaimID, "reference declares no Claim ID")
	}
	if n := r.occurrences[ref.ClaimID]; n > 1 {
		return unresolved(res, ReasonAmbiguousClaimID,
			fmt.Sprintf("Claim ID %q was observed %d times, so the reference is ambiguous", ref.ClaimID, n))
	}
	c, ok := r.byID[ref.ClaimID]
	if !ok {
		return unresolved(res, ReasonClaimNotObserved,
			fmt.Sprintf("Claim ID %q was not observed in the supplied scope", ref.ClaimID))
	}
	res.Claim = &c
	res.SourceEligibility = r.sourceEligibility(c.DocumentID)

	superseders, err := r.SupersededBy(ref.ClaimID)
	if err != nil {
		return unresolved(res, ReasonAmbiguousSupersession, err.Error())
	}
	res.ReplacementCandidates = superseders

	switch ref.Intent {
	case IntentHistoricalExact:
		// A historical exact reference binds the identity itself. It is not a
		// current-authority surface, so supersession and source ineligibility
		// do not revoke it; eligibility travels with the result as provenance.
		res.Status = StatusResolvedExact
		if len(res.ReplacementCandidates) > 0 {
			res.Reason = ReasonSuperseded
			res.Detail = fmt.Sprintf("Claim %q is superseded but remains bound exactly for historical reference",
				ref.ClaimID)
		}
		return res

	case IntentCurrentUse:
		if !r.admitted {
			return unresolved(res, ReasonScopeStructurallyInvalid,
				"observed structure is proven invalid, so it cannot produce a current authoritative binding")
		}
		if c, unproven := r.unprovenRequiredCheck(); unproven {
			res.Check = c
			return unresolved(res, ReasonRequiredCheckUnproven,
				fmt.Sprintf("check %s is required by this operation but was not proven (completeness %q)",
					c, r.scope.CompletenessFor(c)))
		}
		switch res.SourceEligibility {
		case Ineligible:
			res.Status = StatusNotCurrentAuthoritative
			res.Reason = ReasonSourceIneligible
			res.Detail = fmt.Sprintf("containing source %q is ineligible for current authoritative use",
				c.DocumentID)
			return res
		case Eligible:
			// continue
		default:
			return unresolved(res, ReasonSourceEligibilityUnknown,
				fmt.Sprintf("eligibility of containing source %q could not be established", c.DocumentID))
		}

		if len(res.ReplacementCandidates) > 0 {
			res.Status = StatusStaleReviewRequired
			res.Reason = ReasonSuperseded
			res.Detail = fmt.Sprintf("Claim %q is superseded; rebinding requires explicit authority review, candidates: %v",
				ref.ClaimID, res.ReplacementCandidates)
			return res
		}

		// No superseder was observed, which is not the same as none existing.
		// Only a scope declared complete for currentness can rule that out.
		if !r.scope.IsCompleteFor(CheckSupersessionCurrentness) {
			res.Check = CheckSupersessionCurrentness
			return unresolved(res, ReasonCurrentnessIncomplete,
				fmt.Sprintf("no superseder of %q was observed, but scope completeness for %s is %q",
					ref.ClaimID, CheckSupersessionCurrentness,
					r.scope.CompletenessFor(CheckSupersessionCurrentness)))
		}

		res.Status = StatusResolvedCurrent
		return res

	default:
		return unresolved(res, ReasonUnspecifiedIntent, "reference declares no use intent")
	}
}

func unresolved(res Resolution, reason ResolutionReason, detail string) Resolution {
	res.Status = StatusResolutionUnresolved
	res.Reason = reason
	res.Detail = detail
	return res
}

// sourceEligibility consults the supplied resolver, defaulting to unknown.
//
// DocumentID is workspace-unambiguous by contract, so one DocumentID never
// stands for two different canonical sources and this lookup cannot collapse
// them.
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

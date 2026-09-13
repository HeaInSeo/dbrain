package claim

// Check names an invariant whose answer depends on how much material the
// validator actually observed.
type Check string

const (
	// CheckClaimIDUniqueness covers "no two Claims share an ID". Observing a
	// duplicate inside the supplied material is always conclusive; observing no
	// duplicate is conclusive only for a scope declared complete.
	CheckClaimIDUniqueness Check = "CLAIM_ID_UNIQUENESS"

	// CheckSupersessionTargetExistence covers "every declared supersession
	// target exists". An absent target is a defect only when the scope is
	// declared complete for this check; otherwise it is UNRESOLVED.
	CheckSupersessionTargetExistence Check = "SUPERSESSION_TARGET_EXISTENCE"
)

// Completeness declares how much the supplied observation proves for one check.
type Completeness string

const (
	// CompletenessUnknown is the zero value: the caller did not declare whether
	// the observation is complete, so absence proves nothing.
	CompletenessUnknown Completeness = ""

	// CompletenessIncomplete declares that material relevant to the check was
	// knowingly not observed.
	CompletenessIncomplete Completeness = "INCOMPLETE"

	// CompletenessComplete declares that everything relevant to the check was
	// observed, which makes absence conclusive.
	CompletenessComplete Completeness = "COMPLETE"
)

// ValidationScope is the explicit material a validation or resolution request
// observed, plus the per-check completeness declared for it.
//
// A validator never assumes it is looking at the whole Development Workspace.
// Cross-repository discovery is out of scope for this package: the scope is an
// input, supplied by whatever layer did the observing.
type ValidationScope struct {
	// Observed are the Claims actually seen. Order is irrelevant; results are
	// deterministic regardless of it.
	Observed []Claim

	// Complete declares completeness per check. A missing or empty entry means
	// CompletenessUnknown, so the zero-value scope fails closed.
	Complete map[Check]Completeness

	// Subject and Scope record the authority boundary the observation was taken
	// against, for provenance. They are informational here: this package does
	// not route authority.
	Subject string
	Scope   string
}

// CompletenessFor reports the declared completeness for one check, defaulting
// to CompletenessUnknown.
func (s ValidationScope) CompletenessFor(c Check) Completeness {
	if s.Complete == nil {
		return CompletenessUnknown
	}
	switch s.Complete[c] {
	case CompletenessComplete:
		return CompletenessComplete
	case CompletenessIncomplete:
		return CompletenessIncomplete
	default:
		return CompletenessUnknown
	}
}

// IsCompleteFor reports whether the scope was declared complete for one check.
func (s ValidationScope) IsCompleteFor(c Check) bool {
	return s.CompletenessFor(c) == CompletenessComplete
}

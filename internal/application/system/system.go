// Package system owns process-resident application state.
package system

import "sync/atomic"

// SetupState is the process-local view of the persisted one-time setup state.
type SetupState struct {
	// required remains true until the initial administrator transaction commits.
	required atomic.Bool
}

// NewSetupState creates process-local setup state from the startup database check.
func NewSetupState(required bool) *SetupState {
	state := &SetupState{}
	state.required.Store(required)
	return state
}

// Required reports whether the first-run setup flow remains available.
func (s *SetupState) Required() bool {
	return s != nil && s.required.Load()
}

// Complete marks the first-run setup flow unavailable for this process.
func (s *SetupState) Complete() {
	if s != nil {
		s.required.Store(false)
	}
}

// System exposes process-resident system state to application boundaries.
type System struct {
	// setup is initialized once after schema setup and never reads per request.
	setup *SetupState
}

// New constructs system use cases from process-resident state.
func New(setup *SetupState) *System {
	return &System{setup: setup}
}

// SetupRequired reports whether the one-time setup flow is still available.
func (s *System) SetupRequired() bool {
	return s != nil && s.setup.Required()
}

// CompleteSetup records the completed first-run setup transaction.
func (s *System) CompleteSetup() {
	if s != nil {
		s.setup.Complete()
	}
}

// Package types defines shared domain types used across the orchestration engine.
package types

import "time"

// Role enumerates the 9 team seats.
type Role string

const (
	RoleCoordinator Role = "coordinator"
	RolePM          Role = "pm"
	RoleArchitect   Role = "architect"
	RoleUIUX        Role = "uiux"
	RoleFrontend    Role = "frontend"
	RoleBackend     Role = "backend"
	RoleQA          Role = "qa"
	RoleSecurity    Role = "security"
	RoleDevOps      Role = "devops"
)

// IsWriter reports whether the role can modify source code (serialized write model).
func (r Role) IsWriter() bool {
	return r == RoleFrontend || r == RoleBackend
}

// IntentType classifies a user request for routing.
type IntentType string

const (
	IntentChat      IntentType = "chat"
	IntentExplain   IntentType = "explain"
	IntentQuickEdit IntentType = "quick_edit"
	IntentDebug     IntentType = "debug"
	IntentBuild     IntentType = "build"
)

// Intent is the routing decision output.
type Intent struct {
	Type        IntentType
	Depth       Depth
	WriteAccess bool
	Scope       string
}

// Depth estimates task complexity for size-adaptive scheduling.
type Depth string

const (
	DepthTrivial Depth = "trivial"
	DepthFeature Depth = "feature"
	DepthBuild   Depth = "build"
)

// Phase is a pipeline stage.
type Phase string

const (
	PhaseClarify  Phase = "clarify"
	PhaseResearch Phase = "research"
	PhaseDocs     Phase = "docs"
	PhaseSpec     Phase = "spec"
	PhaseFrontend Phase = "frontend"
	PhaseBackend  Phase = "backend"
	PhaseQuality  Phase = "quality"
	PhaseDelivery Phase = "delivery"
)

// TaskNode is a single node in the DAG plan.
type TaskNode struct {
	ID           string
	Phase        Phase
	Role         Role
	DependsOn    []string
	ArtifactsIn  []string
	ArtifactsOut []string
	Writer       bool
}

// Gate is a checkpoint in the pipeline.
type Gate struct {
	ID    string
	After string // node ID
	Type  GateType
}

// GateType enumerates gate kinds.
type GateType string

const (
	GateHuman GateType = "human"
	GateAuto  GateType = "auto"
)

// Plan is the DAG task graph.
type Plan struct {
	ID    string
	Nodes []TaskNode
	Gates []Gate
}

// Verdict is the structured review output from a reviewer role.
type Verdict struct {
	TaskID    string
	Role      Role
	Decision  Decision
	Severity  string
	Findings  []Finding
	Artifacts []string
}

// Decision is the review verdict outcome.
type Decision string

const (
	DecisionAccept   Decision = "accept"
	DecisionBlocking Decision = "blocking"
	DecisionAdvisory Decision = "advisory"
)

// Finding is a single check result within a Verdict.
type Finding struct {
	Check    string
	Status   string
	Detail   string
	Evidence string
}

// UserRequest is the raw user input.
type UserRequest struct {
	Message string
	Backend string
}

// Delivery is the final deliverable bundle.
type Delivery struct {
	PlanID    string
	Artifacts []string
	Score     int
	Passed    bool
	CreatedAt time.Time
}

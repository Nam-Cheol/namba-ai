package namba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	queueStateActive        = "active"
	queueStateWaiting       = "waiting"
	queueStateBlocked       = "blocked"
	queueStatePaused        = "paused"
	queueStateStopped       = "stopped"
	queueStateDone          = "done"
	queueOperatorRunning    = "running"
	queueOperatorWaiting    = "waiting"
	queueOperatorBlocked    = "blocked"
	queueOperatorDone       = "done"
	queuePhasePending       = "pending"
	queuePhaseReviewing     = "reviewing"
	queuePhaseReviewed      = "reviewed"
	queuePhaseBranchReady   = "branch_ready"
	queuePhaseRunning       = "running"
	queuePhasePRReady       = "pr_ready"
	queuePhaseChecksPending = "checks_pending"
	queuePhaseReadyToLand   = "ready_to_land"
	queuePhaseWaitingLand   = "waiting_for_land"
	queuePhaseLanding       = "landing"
	queuePhaseLanded        = "landed"
	queuePhaseSkipped       = "skipped"
	queuePhaseBlocked       = "blocked"
	queuePhasePaused        = "paused"
	queuePhaseStopped       = "stopped"
)

var specRangePattern = regexp.MustCompile(`^(SPEC-\d{3})\.\.(SPEC-\d{3})$`)

type queueOptions struct {
	AutoLand        bool   `json:"auto_land"`
	SkipCodexReview bool   `json:"skip_codex_review"`
	Remote          string `json:"remote"`
}

type queueInvocation struct {
	Subcommand string
	Targets    []string
	Options    queueOptions
	Verbose    bool
	Help       bool
}

type queueState struct {
	ID                  string               `json:"id"`
	Status              string               `json:"status"`
	OperatorState       string               `json:"operator_state"`
	Detail              string               `json:"detail,omitempty"`
	CreatedAt           string               `json:"created_at"`
	UpdatedAt           string               `json:"updated_at"`
	Targets             []string             `json:"targets"`
	Options             queueOptions         `json:"options"`
	PauseRequested      bool                 `json:"pause_requested"`
	StopRequested       bool                 `json:"stop_requested"`
	ActiveSpecID        string               `json:"active_spec_id,omitempty"`
	ExpectedBranch      string               `json:"expected_branch,omitempty"`
	CurrentRunLogID     string               `json:"current_run_log_id,omitempty"`
	LastObservedHeadSHA string               `json:"last_observed_head_sha,omitempty"`
	LastSafeCheckpoint  string               `json:"last_safe_checkpoint,omitempty"`
	LastBlocker         string               `json:"last_blocker,omitempty"`
	LastEvidencePath    string               `json:"last_evidence_path,omitempty"`
	LastRecoveryAction  string               `json:"last_recovery_action,omitempty"`
	CheckProofStrategy  string               `json:"check_proof_strategy,omitempty"`
	Specs               map[string]queueSpec `json:"specs"`
	CompletedSpecs      []string             `json:"completed_specs,omitempty"`
	SkippedSpecs        []string             `json:"skipped_specs,omitempty"`
	CompletedSpecCount  int                  `json:"completed_spec_count"`
	SkippedSpecCount    int                  `json:"skipped_spec_count"`
}

type queueSpec struct {
	SpecID             string `json:"spec_id"`
	Status             string `json:"status"`
	OperatorState      string `json:"operator_state"`
	Phase              string `json:"phase"`
	Branch             string `json:"branch,omitempty"`
	PRNumber           int    `json:"pr_number,omitempty"`
	PRURL              string `json:"pr_url,omitempty"`
	ValidationEvidence string `json:"validation_evidence,omitempty"`
	LandEvidence       string `json:"land_evidence,omitempty"`
	SkipReason         string `json:"skip_reason,omitempty"`
	Blocker            string `json:"blocker,omitempty"`
	EvidencePath       string `json:"evidence_path,omitempty"`
	RecoveryAction     string `json:"recovery_action,omitempty"`
	LastCheckpoint     string `json:"last_checkpoint,omitempty"`
}

func queueUsageText() string {
	lines := []string{
		"namba queue",
		"",
		"Usage:",
		"  namba queue start <SPEC-RANGE|SPEC-LIST> [--auto-land] [--skip-codex-review] [--remote origin]",
		"  namba queue status [--verbose]",
		"  namba queue resume",
		"  namba queue pause",
		"  namba queue stop",
		"",
		"Behavior:",
		"  Process existing SPEC packages one at a time through review, implementation, PR, checks, land, and local main refresh.",
		"  Queue state is stored under .namba/logs/queue/ and resume recomputes Git/GitHub truth before continuing.",
		"  Ambiguous validation, checks, mergeability, branch, or GitHub state blocks instead of being skipped.",
	}
	return strings.Join(lines, "\n") + "\n"
}

func (a *App) runQueue(ctx context.Context, args []string) error {
	inv, err := parseQueueInvocation(args)
	if err != nil {
		return commandUsageError("queue", err)
	}
	if inv.Help {
		return a.printCommandUsage("queue")
	}

	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}

	switch inv.Subcommand {
	case "start":
		return a.startQueue(ctx, root, inv)
	case "status":
		return a.printQueueStatus(root, inv.Verbose)
	case "resume":
		return a.resumeQueue(ctx, root)
	case "pause":
		return a.pauseQueue(root)
	case "stop":
		return a.stopQueue(root)
	default:
		return commandUsageError("queue", fmt.Errorf("unknown queue subcommand %q", inv.Subcommand))
	}
}

func parseQueueInvocation(args []string) (queueInvocation, error) {
	if wantsCommandHelp(args) {
		return queueInvocation{Help: true}, nil
	}
	if len(args) == 0 {
		return queueInvocation{}, errors.New("queue requires a subcommand")
	}
	inv := queueInvocation{Subcommand: strings.TrimSpace(args[0]), Options: queueOptions{Remote: defaultGitRemote}}
	switch inv.Subcommand {
	case "start":
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--auto-land":
				inv.Options.AutoLand = true
			case "--skip-codex-review":
				inv.Options.SkipCodexReview = true
			case "--remote":
				value, err := consumeFlagValue(args, &i, args[i])
				if err != nil {
					return queueInvocation{}, err
				}
				inv.Options.Remote = strings.TrimSpace(value)
			default:
				if strings.HasPrefix(args[i], "--") {
					return queueInvocation{}, fmt.Errorf("unknown flag %q", args[i])
				}
				inv.Targets = append(inv.Targets, args[i])
			}
		}
		if len(inv.Targets) == 0 {
			return queueInvocation{}, errors.New("queue start requires at least one SPEC target")
		}
		if inv.Options.Remote == "" {
			return queueInvocation{}, errors.New("queue remote is required")
		}
	case "status":
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--verbose":
				inv.Verbose = true
			default:
				if strings.HasPrefix(args[i], "--") {
					return queueInvocation{}, fmt.Errorf("unknown flag %q", args[i])
				}
				return queueInvocation{}, errors.New("queue status does not accept SPEC targets")
			}
		}
	case "resume", "pause", "stop":
		if len(args) > 1 {
			return queueInvocation{}, fmt.Errorf("queue %s does not accept arguments", inv.Subcommand)
		}
	default:
		return queueInvocation{}, fmt.Errorf("unknown queue subcommand %q", inv.Subcommand)
	}
	return inv, nil
}

func (a *App) startQueue(ctx context.Context, root string, inv queueInvocation) error {
	if state, err := readQueueState(root); err == nil && queueIsActive(state) {
		return fmt.Errorf("queue already active: %s (%s)", state.ID, state.Status)
	}
	targets, err := expandQueueTargets(root, inv.Targets)
	if err != nil {
		return err
	}
	now := a.now().Format(time.RFC3339)
	state := queueState{
		ID:            "queue-" + a.now().Format("20060102-150405"),
		Status:        queueStateActive,
		OperatorState: queueOperatorRunning,
		Detail:        queuePhasePending,
		CreatedAt:     now,
		UpdatedAt:     now,
		Targets:       targets,
		Options:       inv.Options,
		Specs:         make(map[string]queueSpec, len(targets)),
	}
	for _, target := range targets {
		state.Specs[target] = queueSpec{SpecID: target, Status: queueStateActive, OperatorState: queueOperatorWaiting, Phase: queuePhasePending}
	}
	if err := a.writeQueueState(root, state); err != nil {
		return err
	}
	return a.advanceQueue(ctx, root, state)
}

func (a *App) resumeQueue(ctx context.Context, root string) error {
	state, err := readQueueState(root)
	if err != nil {
		return err
	}
	if state.Status == queueStateDone {
		return a.printQueueState(root, state, false)
	}
	if state.Status == queueStateStopped {
		return errors.New("queue is stopped; start a new queue to continue")
	}
	state.PauseRequested = false
	state.Status = queueStateActive
	state.OperatorState = queueOperatorRunning
	state.Detail = firstNonBlank(state.Detail, queuePhasePending)
	state.LastBlocker = ""
	state.LastRecoveryAction = ""
	state.UpdatedAt = a.now().Format(time.RFC3339)
	if err := a.writeQueueState(root, state); err != nil {
		return err
	}
	return a.advanceQueue(ctx, root, state)
}

func (a *App) pauseQueue(root string) error {
	state, err := readQueueState(root)
	if err != nil {
		return err
	}
	if state.Status == queueStateDone || state.Status == queueStateStopped {
		return a.printQueueState(root, state, false)
	}
	state.PauseRequested = true
	state.Status = queueStatePaused
	state.OperatorState = queueOperatorWaiting
	state.Detail = queuePhasePaused
	state.LastSafeCheckpoint = firstNonBlank(state.LastSafeCheckpoint, "pause_requested")
	state.UpdatedAt = a.now().Format(time.RFC3339)
	if err := a.writeQueueState(root, state); err != nil {
		return err
	}
	return a.printQueueState(root, state, false)
}

func (a *App) stopQueue(root string) error {
	state, err := readQueueState(root)
	if err != nil {
		return err
	}
	state.StopRequested = true
	state.Status = queueStateStopped
	state.OperatorState = queueOperatorBlocked
	state.Detail = queuePhaseStopped
	state.LastRecoveryAction = "start a new queue when ready"
	state.UpdatedAt = a.now().Format(time.RFC3339)
	if err := a.writeQueueState(root, state); err != nil {
		return err
	}
	return a.printQueueState(root, state, false)
}

func (a *App) printQueueStatus(root string, verbose bool) error {
	state, err := readQueueState(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(a.stdout, "Queue: none")
			fmt.Fprintln(a.stdout, "Next: namba queue start SPEC-001..SPEC-002")
			return nil
		}
		return err
	}
	return a.printQueueState(root, state, verbose)
}

func (a *App) advanceQueue(ctx context.Context, root string, state queueState) error {
	if state.Specs == nil {
		state.Specs = map[string]queueSpec{}
	}
	for _, specID := range state.Targets {
		var err error
		state, err = refreshQueueControlRequests(root, state)
		if err != nil {
			return err
		}
		if state.PauseRequested {
			state.Status = queueStatePaused
			state.OperatorState = queueOperatorWaiting
			state.Detail = queuePhasePaused
			state.UpdatedAt = a.now().Format(time.RFC3339)
			if err := a.writeQueueState(root, state); err != nil {
				return err
			}
			return a.printQueueState(root, state, false)
		}
		if state.StopRequested {
			state.Status = queueStateStopped
			state.OperatorState = queueOperatorBlocked
			state.Detail = queuePhaseStopped
			state.UpdatedAt = a.now().Format(time.RFC3339)
			if err := a.writeQueueState(root, state); err != nil {
				return err
			}
			return a.printQueueState(root, state, false)
		}
		specState := state.Specs[specID]
		if specState.SpecID == "" {
			specState = queueSpec{SpecID: specID, Status: queueStateActive, Phase: queuePhasePending}
		}
		if specState.Phase == queuePhaseSkipped {
			state = markQueueSpecSkipped(state, specID, specState, firstNonBlank(specState.SkipReason, "already landed"))
			continue
		}
		if specState.Phase == queuePhaseLanded || specState.Status == queueStateDone {
			state = markQueueSpecDone(state, specID, specState)
			continue
		}
		state.ActiveSpecID = specID
		state.Status = queueStateActive
		state.OperatorState = queueOperatorRunning
		state.Detail = firstNonBlank(specState.Phase, queuePhasePending)
		state.UpdatedAt = a.now().Format(time.RFC3339)
		if err := a.writeQueueState(root, state); err != nil {
			return err
		}

		next, done, err := a.advanceQueueSpec(ctx, root, state, specID)
		if err != nil {
			return err
		}
		state = next
		if !done {
			return a.printQueueState(root, state, false)
		}
	}

	state.Status = queueStateDone
	state.OperatorState = queueOperatorDone
	state.Detail = queuePhaseLanded
	state.ActiveSpecID = ""
	state.ExpectedBranch = ""
	state.LastSafeCheckpoint = "queue_done"
	state.UpdatedAt = a.now().Format(time.RFC3339)
	if err := a.writeQueueState(root, state); err != nil {
		return err
	}
	return a.printQueueState(root, state, false)
}

func (a *App) advanceQueueSpec(ctx context.Context, root string, state queueState, specID string) (queueState, bool, error) {
	specPkg, err := a.loadSpec(root, specID)
	if err != nil {
		return blockQueueSpec(a, root, state, specID, "spec_missing", "", "create or restore the SPEC package", err.Error())
	}
	specState := state.Specs[specID]
	specState.SpecID = specID

	branch, err := a.resolveQueueBranch(ctx, root, state, specPkg, specState)
	if err != nil {
		return blockQueueSpec(a, root, state, specID, "branch_ambiguous", "", "resolve branch ambiguity and run `namba queue resume`", err.Error())
	}
	specState.Branch = branch
	state.ExpectedBranch = branch
	state.Specs[specID] = specState
	state.LastSafeCheckpoint = specID + ":branch_resolved"
	if err := a.writeQueueState(root, state); err != nil {
		return state, false, err
	}
	if landed, evidence := a.queueLandedEvidence(ctx, root, branch); landed {
		state = markQueueSpecSkipped(state, specID, specState, evidence)
		state.LastSafeCheckpoint = specID + ":skipped"
		if err := a.writeQueueState(root, state); err != nil {
			return state, false, err
		}
		return state, true, nil
	}
	if err := a.ensureQueueBranch(ctx, root, state, branch); err != nil {
		return blockQueueSpec(a, root, state, specID, "branch_ready_failed", "", "fix Git branch state and run `namba queue resume`", err.Error())
	}
	state.LastObservedHeadSHA, _ = a.gitHeadSHA(ctx, root)
	specState.Phase = queuePhaseBranchReady
	state, err = updateQueueSpecState(a, root, state, specID, specState, queueOperatorRunning, queuePhaseBranchReady)
	if err != nil {
		return state, false, err
	}
	if state, done, err := a.honorQueueControlRequests(root, state); err != nil || done {
		return state, false, err
	}

	if err := a.ensureQueueReviewReady(root, specID); err != nil {
		return blockQueueSpec(a, root, state, specID, "review_blocked", specReviewReadinessPath(specID), "update review artifacts and run `namba queue resume`", err.Error())
	}
	specState.Phase = queuePhaseReviewed
	state, err = updateQueueSpecState(a, root, state, specID, specState, queueOperatorRunning, queuePhaseReviewed)
	if err != nil {
		return state, false, err
	}
	if state, done, err := a.honorQueueControlRequests(root, state); err != nil || done {
		return state, false, err
	}

	currentHead := strings.TrimSpace(state.LastObservedHeadSHA)
	if currentHead == "" {
		currentHead, _ = a.gitHeadSHA(ctx, root)
		currentHead = strings.TrimSpace(currentHead)
	}
	executionReady, validationEvidence := queueExecutionSucceeded(root, specID, currentHead)
	if !executionReady {
		runCtx, err := a.loadRunExecutionContext(root, runExecuteOptions{specID: specID, mode: executionModeTeam})
		if err != nil {
			return blockQueueSpec(a, root, state, specID, "run_context_failed", "", "fix run context and run `namba queue resume`", err.Error())
		}
		if err := a.materializeRunExecutionPrompt(runCtx); err != nil {
			return blockQueueSpec(a, root, state, specID, "run_prompt_failed", "", "fix run prompt output and run `namba queue resume`", err.Error())
		}
		specState.Phase = queuePhaseRunning
		state.CurrentRunLogID = strings.ToLower(specID)
		state, err = updateQueueSpecState(a, root, state, specID, specState, queueOperatorRunning, queuePhaseRunning)
		if err != nil {
			return state, false, err
		}
		if state, done, err := a.honorQueueControlRequests(root, state); err != nil || done {
			return state, false, err
		}
		if err := a.dispatchRunExecution(ctx, runExecuteOptions{specID: specID, mode: executionModeTeam}, runCtx); err != nil {
			return blockQueueSpec(a, root, state, specID, "validation_failed", queueRunEvidencePath(specID), "fix implementation or validation and run `namba queue resume`", err.Error())
		}
		currentHead, _ = a.gitHeadSHA(ctx, root)
		currentHead = strings.TrimSpace(currentHead)
		state.LastObservedHeadSHA = currentHead
		if err := stampQueueRunEvidenceHead(root, specID, currentHead); err != nil {
			return blockQueueSpec(a, root, state, specID, "validation_ambiguous", queueRunEvidencePath(specID), "inspect run evidence and run `namba queue resume`", err.Error())
		}
		executionReady, validationEvidence = queueExecutionSucceeded(root, specID, currentHead)
		if !executionReady {
			return blockQueueSpec(a, root, state, specID, "validation_ambiguous", queueRunEvidencePath(specID), "inspect run evidence and run `namba queue resume`", "run completed but validation evidence is missing or failed")
		}
	}
	specState.ValidationEvidence = validationEvidence
	specState.Phase = queuePhasePRReady
	state, err = updateQueueSpecState(a, root, state, specID, specState, queueOperatorRunning, queuePhasePRReady)
	if err != nil {
		return state, false, err
	}
	if state, done, err := a.honorQueueControlRequests(root, state); err != nil || done {
		return state, false, err
	}

	if err := a.runActiveSpecSync(ctx, root, specID); err != nil {
		return blockQueueSpec(a, root, state, specID, "sync_failed", "", "fix active-SPEC sync and run `namba queue resume`", err.Error())
	}
	pr, err := a.prepareQueuePullRequest(ctx, root, state, specPkg, branch)
	if err != nil {
		return blockQueueSpec(a, root, state, specID, "pr_failed", "", "fix PR state and run `namba queue resume`", err.Error())
	}
	specState.PRNumber = pr.Number
	specState.PRURL = pr.URL
	specState.Phase = queuePhaseChecksPending
	state, err = updateQueueSpecState(a, root, state, specID, specState, queueOperatorRunning, queuePhaseChecksPending)
	if err != nil {
		return state, false, err
	}
	if state, done, err := a.honorQueueControlRequests(root, state); err != nil || done {
		return state, false, err
	}

	pr, err = a.loadPullRequest(ctx, root, strconv.Itoa(pr.Number), landPullRequestFields()...)
	if err != nil {
		return blockQueueSpec(a, root, state, specID, "pr_state_ambiguous", "", "fix GitHub PR state and run `namba queue resume`", err.Error())
	}
	profile, err := a.loadInitProfileFromConfig(root)
	if err != nil {
		return blockQueueSpec(a, root, state, specID, "pr_state_ambiguous", "", "fix GitHub PR state and run `namba queue resume`", err.Error())
	}
	baseBranch := prBaseBranch(profile)
	if pr.BaseRefName != "" && pr.BaseRefName != baseBranch {
		return blockQueueSpec(a, root, state, specID, "pr_base_mismatch", "", "retarget the PR or start a queue for the correct base branch", fmt.Sprintf("pull request #%d targets %q, expected %q", pr.Number, pr.BaseRefName, baseBranch))
	}
	if pr.IsDraft {
		return blockQueueSpec(a, root, state, specID, "pr_draft", "", "mark the PR ready for review and run `namba queue resume`", fmt.Sprintf("pull request #%d is still a draft", pr.Number))
	}
	checkProof, err := classifyQueueCheckProof(pr)
	state.CheckProofStrategy = checkProof
	if err != nil {
		return blockQueueSpec(a, root, state, specID, "checks_ambiguous", "", "wait for trustworthy check evidence and run `namba queue resume`", err.Error())
	}
	if err := validateLandPullRequest(pr, false); err != nil {
		pending, failed := classifyStatusChecks(pr.StatusChecks)
		switch {
		case len(pending) > 0:
			specState.Phase = queuePhaseChecksPending
			state.Status = queueStateWaiting
			state.OperatorState = queueOperatorWaiting
			state.Detail = "waiting_for_checks"
			state.LastBlocker = err.Error()
			state.LastRecoveryAction = "wait for checks, then run `namba queue resume`"
			state.Specs[specID] = specState
			state.UpdatedAt = a.now().Format(time.RFC3339)
			return state, false, a.writeQueueState(root, state)
		case len(failed) > 0:
			return blockQueueSpec(a, root, state, specID, "checks_failed", "", "fix failed checks and run `namba queue resume`", err.Error())
		default:
			return blockQueueSpec(a, root, state, specID, "pr_not_mergeable", "", "fix PR mergeability and run `namba queue resume`", err.Error())
		}
	}
	specState.Phase = queuePhaseReadyToLand
	state, err = updateQueueSpecState(a, root, state, specID, specState, queueOperatorRunning, queuePhaseReadyToLand)
	if err != nil {
		return state, false, err
	}
	if state, done, err := a.honorQueueControlRequests(root, state); err != nil || done {
		return state, false, err
	}

	if !state.Options.AutoLand {
		specState.Phase = queuePhaseWaitingLand
		state.Status = queueStateWaiting
		state.OperatorState = queueOperatorWaiting
		state.Detail = queuePhaseWaitingLand
		state.LastBlocker = "auto-land disabled"
		state.LastRecoveryAction = "land the PR, refresh local main, then run `namba queue resume`"
		state.Specs[specID] = specState
		state.UpdatedAt = a.now().Format(time.RFC3339)
		return state, false, a.writeQueueState(root, state)
	}

	specState.Phase = queuePhaseLanding
	state, err = updateQueueSpecState(a, root, state, specID, specState, queueOperatorRunning, queuePhaseLanding)
	if err != nil {
		return state, false, err
	}
	if _, err := a.runBinary(ctx, "gh", []string{"pr", "merge", strconv.Itoa(pr.Number), "--merge"}, root); err != nil {
		return blockQueueSpec(a, root, state, specID, "land_failed", "", "fix merge failure and run `namba queue resume`", err.Error())
	}
	if err := a.updateLocalBaseBranch(ctx, root, branch, prBaseBranch(profile), state.Options.Remote); err != nil {
		return blockQueueSpec(a, root, state, specID, "main_refresh_failed", "", "refresh local main and run `namba queue resume`", err.Error())
	}
	specState.Phase = queuePhaseLanded
	specState.Status = queueStateDone
	specState.OperatorState = queueOperatorDone
	specState.LandEvidence = fmt.Sprintf("PR #%d merged", pr.Number)
	state.Specs[specID] = specState
	state = markQueueSpecDone(state, specID, specState)
	state.LastSafeCheckpoint = specID + ":landed"
	state.UpdatedAt = a.now().Format(time.RFC3339)
	if err := a.writeQueueState(root, state); err != nil {
		return state, false, err
	}
	return state, true, nil
}

func expandQueueTargets(root string, raw []string) ([]string, error) {
	var targets []string
	for _, token := range raw {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if matches := specRangePattern.FindStringSubmatch(token); len(matches) == 3 {
			start, _ := strconv.Atoi(strings.TrimPrefix(matches[1], "SPEC-"))
			end, _ := strconv.Atoi(strings.TrimPrefix(matches[2], "SPEC-"))
			if end < start {
				return nil, fmt.Errorf("invalid SPEC range %q", token)
			}
			for n := start; n <= end; n++ {
				targets = append(targets, fmt.Sprintf("SPEC-%03d", n))
			}
			continue
		}
		if !isQueueSpecID(token) {
			return nil, fmt.Errorf("invalid SPEC target %q", token)
		}
		targets = append(targets, token)
	}
	if len(targets) == 0 {
		return nil, errors.New("queue target list is empty")
	}
	seen := map[string]bool{}
	for _, specID := range targets {
		if seen[specID] {
			return nil, fmt.Errorf("duplicate SPEC target %s", specID)
		}
		seen[specID] = true
		if _, err := os.Stat(filepath.Join(root, specsDir, specID)); err != nil {
			return nil, fmt.Errorf("spec %s not found", specID)
		}
	}
	return targets, nil
}

func isQueueSpecID(value string) bool {
	if len(value) != len("SPEC-000") || !strings.HasPrefix(value, "SPEC-") {
		return false
	}
	_, err := strconv.Atoi(strings.TrimPrefix(value, "SPEC-"))
	return err == nil
}

func (a *App) resolveQueueBranch(ctx context.Context, root string, state queueState, specPkg specPackage, specState queueSpec) (string, error) {
	if strings.TrimSpace(specState.Branch) != "" {
		return specState.Branch, nil
	}
	if strings.TrimSpace(state.ExpectedBranch) != "" && state.ActiveSpecID == specPkg.ID {
		return state.ExpectedBranch, nil
	}
	profile, err := a.loadInitProfileFromConfig(root)
	if err != nil {
		return "", err
	}
	prefix := specBranchPrefix(profile) + specPkg.ID + "-"
	branches, err := a.localBranches(ctx, root)
	if err != nil {
		return "", err
	}
	var matches []string
	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if strings.HasPrefix(branch, prefix) {
			matches = append(matches, branch)
		}
	}
	switch len(matches) {
	case 0:
		slugSource := queueBranchSlugSource(root, specPkg)
		slug, err := normalizeCreateSlug(slugSource)
		if err != nil {
			return "", err
		}
		return prefix + slug, nil
	case 1:
		return matches[0], nil
	default:
		sort.Strings(matches)
		return "", fmt.Errorf("multiple branches match %s: %s", specPkg.ID, strings.Join(matches, ", "))
	}
}

func queueBranchSlugSource(root string, specPkg specPackage) string {
	for _, rel := range []string{"spec.md", "plan.md"} {
		body, err := os.ReadFile(filepath.Join(specPkg.Path, rel))
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(body), "\n") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if line == "" || strings.EqualFold(line, specPkg.ID) || strings.HasPrefix(line, "SPEC-") {
				continue
			}
			return line
		}
	}
	if strings.TrimSpace(specPkg.Description) != "" {
		return specPkg.Description
	}
	_ = root
	return specPkg.ID
}

func (a *App) ensureQueueBranch(ctx context.Context, root string, state queueState, branch string) error {
	current, err := a.currentBranch(ctx, root)
	if err != nil {
		return fmt.Errorf("detect current branch: %w", err)
	}
	if current == branch {
		return nil
	}
	dirty, err := a.hasWorkingTreeChanges(ctx, root)
	if err != nil {
		return err
	}
	if dirty {
		return fmt.Errorf("cannot switch queue branch with uncommitted changes")
	}
	exists, err := a.localBranchExists(ctx, root, branch)
	if err != nil {
		return err
	}
	if exists {
		_, err = a.runBinary(ctx, "git", []string{"checkout", branch}, root)
		return err
	}
	base := "main"
	if profile, err := a.loadInitProfileFromConfig(root); err == nil {
		base = branchBase(profile)
	}
	if _, err := a.runBinary(ctx, "git", []string{"checkout", "-b", branch, base}, root); err != nil {
		return err
	}
	return nil
}

func (a *App) queueLandedEvidence(ctx context.Context, root, branch string) (bool, string) {
	if strings.TrimSpace(branch) == "" {
		return false, ""
	}
	base := "main"
	if profile, err := a.loadInitProfileFromConfig(root); err == nil {
		base = branchBase(profile)
	}
	exists, err := a.localBranchExists(ctx, root, branch)
	if err != nil {
		return false, ""
	}
	if exists {
		if _, err := a.runBinary(ctx, "git", []string{"merge-base", "--is-ancestor", branch, base}, root); err == nil {
			return true, fmt.Sprintf("branch %s is already merged into %s", branch, base)
		}
		return false, ""
	}
	if landed, evidence := a.queueMergedPullRequestEvidence(ctx, root, branch, base); landed {
		return true, evidence
	}
	return false, ""
}

func (a *App) queueMergedPullRequestEvidence(ctx context.Context, root, branch, base string) (bool, string) {
	out, err := a.runBinary(ctx, "gh", []string{"pr", "list", "--head", branch, "--base", base, "--state", "merged", "--json", "number,url,mergedAt,headRefName,baseRefName,state"}, root)
	if err != nil {
		return false, ""
	}
	var prs []githubPullRequest
	if err := json.Unmarshal([]byte(firstNonBlank(out, "[]")), &prs); err != nil || len(prs) == 0 {
		return false, ""
	}
	pr := prs[0]
	when := strings.TrimSpace(pr.MergedAt)
	if when != "" {
		return true, fmt.Sprintf("PR #%d merged at %s", pr.Number, when)
	}
	return true, fmt.Sprintf("PR #%d is merged", pr.Number)
}

func (a *App) ensureQueueReviewReady(root, specID string) error {
	advisory, err := a.refreshSpecReviewReadiness(root, specID)
	if err != nil {
		return err
	}
	states := loadSpecReviewStates(filepath.Join(root, specsDir, specID))
	var blockers []string
	for _, state := range states {
		status := strings.ToLower(strings.TrimSpace(state.Status))
		switch {
		case isClearReviewStatus(status):
		case status == "clear-with-followups":
			body, err := os.ReadFile(filepath.Join(root, specsDir, specID, specReviewsDirName, state.Template.Slug+".md"))
			if err != nil {
				blockers = append(blockers, state.Template.Slug+"=missing")
				continue
			}
			if !reviewFollowupsAreTagged(string(body)) {
				blockers = append(blockers, state.Template.Slug+"=untagged-followups")
			}
		default:
			blockers = append(blockers, state.Template.Slug+"="+firstNonBlank(status, "missing"))
		}
	}
	if len(blockers) > 0 {
		if advisory != "" {
			return fmt.Errorf("%s (%s)", strings.Join(blockers, ", "), advisory)
		}
		return errors.New(strings.Join(blockers, ", "))
	}
	return nil
}

func reviewFollowupsAreTagged(body string) bool {
	inFollowups := false
	found := false
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			inFollowups = strings.EqualFold(trimmed, "## Follow-ups")
			continue
		}
		if !inFollowups || !strings.HasPrefix(trimmed, "-") {
			continue
		}
		found = true
		if !strings.Contains(trimmed, "[non-blocking]") && !strings.Contains(trimmed, "[post-implementation]") {
			return false
		}
	}
	return found
}

func queueExecutionSucceeded(root, specID, currentHead string) (bool, string) {
	currentHead = strings.TrimSpace(currentHead)
	if currentHead == "" {
		return false, ""
	}
	logID := strings.ToLower(specID)
	executionPath := filepath.Join(root, logsDir, "runs", logID+"-execution.json")
	validationPath := filepath.Join(root, logsDir, "runs", logID+"-validation.json")
	executionBytes, err := os.ReadFile(executionPath)
	if err != nil {
		return false, ""
	}
	var execution executionResult
	if err := json.Unmarshal(executionBytes, &execution); err != nil || !execution.Succeeded || strings.TrimSpace(execution.HeadSHA) != currentHead {
		return false, filepath.ToSlash(filepath.Join(logsDir, "runs", logID+"-execution.json"))
	}
	validationBytes, err := os.ReadFile(validationPath)
	if err != nil {
		return false, filepath.ToSlash(filepath.Join(logsDir, "runs", logID+"-validation.json"))
	}
	var validation validationReport
	if err := json.Unmarshal(validationBytes, &validation); err != nil || !validation.Passed || strings.TrimSpace(validation.HeadSHA) != currentHead {
		return false, filepath.ToSlash(filepath.Join(logsDir, "runs", logID+"-validation.json"))
	}
	return true, filepath.ToSlash(filepath.Join(logsDir, "runs", logID+"-validation.json"))
}

func stampQueueRunEvidenceHead(root, specID, currentHead string) error {
	currentHead = strings.TrimSpace(currentHead)
	if currentHead == "" {
		return errors.New("current HEAD is unknown")
	}
	logID := strings.ToLower(specID)
	executionPath := filepath.Join(root, logsDir, "runs", logID+"-execution.json")
	validationPath := filepath.Join(root, logsDir, "runs", logID+"-validation.json")

	executionBytes, err := os.ReadFile(executionPath)
	if err != nil {
		return err
	}
	var execution executionResult
	if err := json.Unmarshal(executionBytes, &execution); err != nil {
		return err
	}
	execution.HeadSHA = currentHead
	if err := writeJSONFile(executionPath, execution); err != nil {
		return err
	}

	validationBytes, err := os.ReadFile(validationPath)
	if err != nil {
		return err
	}
	var validation validationReport
	if err := json.Unmarshal(validationBytes, &validation); err != nil {
		return err
	}
	validation.HeadSHA = currentHead
	return writeJSONFile(validationPath, validation)
}

func queueRunEvidencePath(specID string) string {
	return filepath.ToSlash(filepath.Join(logsDir, "runs", strings.ToLower(specID)+"-validation.json"))
}

func (a *App) runActiveSpecSync(ctx context.Context, root, specID string) error {
	syncCtx, err := a.loadSyncContext(root)
	if err != nil {
		return err
	}
	syncCtx.LatestSpec = specID
	qualityCfg, _ := a.loadQualityConfig(root)
	analysisCfg, err := a.loadAnalysisConfig(root)
	if err != nil {
		return err
	}
	readmeOutputs := buildReadmeOutputs(syncCtx.ProjectCfg, syncCtx.Profile, syncCtx.DocsCfg)
	analysis := analyzeProject(root, syncCtx.ProjectCfg, qualityCfg, analysisCfg)
	projectOutputs := analysis.renderOutputs()
	readinessBatch, err := buildSpecReviewReadinessBatch(root)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	syncCtx.Support = a.buildSyncSupportContext(root, specID, readinessBatch.Advisories)
	session, err := a.beginManagedOutputSession(root)
	if err != nil {
		return err
	}
	if err := session.replaceManagedOutputs(readmeOutputs, isReadmeManagedPath, nil); err != nil {
		return err
	}
	if err := session.replaceManagedOutputs(projectOutputs, isProjectAnalysisManagedPath, nil); err != nil {
		return err
	}
	if len(analysis.Quality.Errors) > 0 {
		if _, err := session.commit(); err != nil {
			return err
		}
		return errors.New("project analysis quality gate failed")
	}
	if err := session.replaceManagedOutputs(readinessBatch.Outputs, isSpecReviewReadinessManagedPath, nil); err != nil {
		return err
	}
	if err := session.replaceManagedOutputs(buildSyncProjectSupportOutputs(syncCtx), isSyncProjectSupportManagedPath, nil); err != nil {
		return err
	}
	_, err = session.commit()
	return err
}

func (a *App) prepareQueuePullRequest(ctx context.Context, root string, state queueState, specPkg specPackage, branch string) (githubPullRequest, error) {
	profile, err := a.loadInitProfileFromConfig(root)
	if err != nil {
		return githubPullRequest{}, err
	}
	if !strings.EqualFold(strings.TrimSpace(profile.GitProvider), "github") {
		return githubPullRequest{}, fmt.Errorf("queue PR currently supports only the GitHub provider, got %q", profile.GitProvider)
	}
	if err := a.requireGitHubCLI(ctx, root); err != nil {
		return githubPullRequest{}, err
	}
	currentBranch, err := a.currentBranch(ctx, root)
	if err != nil {
		return githubPullRequest{}, err
	}
	if currentBranch != branch {
		return githubPullRequest{}, fmt.Errorf("current branch is %q, expected queue branch %q", currentBranch, branch)
	}
	qualityCfg, err := a.loadQualityConfig(root)
	if err != nil {
		return githubPullRequest{}, err
	}
	if err := a.runValidators(ctx, root, qualityCfg); err != nil {
		return githubPullRequest{}, err
	}
	dirty, err := a.hasWorkingTreeChanges(ctx, root)
	if err != nil {
		return githubPullRequest{}, err
	}
	title := queuePullRequestTitle(specPkg)
	if dirty {
		if _, err := a.runBinary(ctx, "git", []string{"add", "-A"}, root); err != nil {
			return githubPullRequest{}, fmt.Errorf("stage changes: %w", err)
		}
		if _, err := a.runBinary(ctx, "git", []string{"commit", "-m", title}, root); err != nil {
			return githubPullRequest{}, fmt.Errorf("create commit: %w", err)
		}
	}
	if _, err := a.runBinary(ctx, "git", []string{"push", "--set-upstream", state.Options.Remote, branch}, root); err != nil {
		return githubPullRequest{}, fmt.Errorf("push branch %s: %w", branch, err)
	}
	baseBranch := prBaseBranch(profile)
	pr, _, err := a.findOrCreatePullRequest(ctx, root, branch, baseBranch, title, buildPullRequestBodyForSpec(root, profile, specPkg.ID))
	if err != nil {
		return githubPullRequest{}, err
	}
	if profile.AutoCodexReview && !state.Options.SkipCodexReview {
		if err := a.ensureReviewComment(ctx, root, pr.Number, codexReviewComment(profile)); err != nil {
			return githubPullRequest{}, err
		}
	}
	return pr, nil
}

func queuePullRequestTitle(specPkg specPackage) string {
	title := strings.TrimSpace(strings.TrimPrefix(specPkg.Description, "#"))
	if title == "" || strings.EqualFold(title, specPkg.ID) {
		return specPkg.ID + " 구현"
	}
	if !strings.Contains(title, specPkg.ID) {
		return specPkg.ID + " " + title
	}
	return title
}

func buildPullRequestBodyForSpec(root string, profile initProfile, specID string) string {
	summaryPath := filepath.ToSlash(filepath.Join(projectDir, "change-summary.md"))
	checklistPath := filepath.ToSlash(filepath.Join(projectDir, "pr-checklist.md"))
	readinessPath := ""
	if specReviewReadinessExists(root, specID) {
		readinessPath = specReviewReadinessPath(specID)
	}
	switch normalizeReadmeLanguage(profile.PRLanguage) {
	case "ko":
		lines := []string{"## 작업 요약", fmt.Sprintf("- 변경 요약: `%s`", summaryPath), fmt.Sprintf("- 검토 체크리스트: `%s`", checklistPath), "", "## 검토 메모", "- `namba queue`가 active SPEC 기준 sync, validation, commit, push를 마친 상태입니다."}
		if readinessPath != "" {
			lines = append(lines, fmt.Sprintf("- Active SPEC review readiness: `%s` (자문적 advisory)", readinessPath))
		}
		return strings.Join(lines, "\n")
	default:
		lines := []string{"## Summary", fmt.Sprintf("- Change summary: `%s`", summaryPath), fmt.Sprintf("- Review checklist: `%s`", checklistPath), "", "## Review Notes", "- `namba queue` has completed active-SPEC sync, validation, commit, and push."}
		if readinessPath != "" {
			lines = append(lines, fmt.Sprintf("- Active SPEC review readiness: `%s` (advisory)", readinessPath))
		}
		return strings.Join(lines, "\n")
	}
}

func classifyQueueCheckProof(pr githubPullRequest) (string, error) {
	if len(pr.StatusChecks) == 0 {
		return "", fmt.Errorf("no PR check evidence was surfaced")
	}
	return "all_surfaced_checks_green", nil
}

func (a *App) gitHeadSHA(ctx context.Context, root string) (string, error) {
	return a.runBinary(ctx, "git", []string{"rev-parse", "HEAD"}, root)
}

func queueStatePath(root string) string {
	return filepath.Join(root, logsDir, "queue", "state.json")
}

func queueReportPath(root string) string {
	return filepath.Join(root, logsDir, "queue", "report.md")
}

func readQueueState(root string) (queueState, error) {
	data, err := os.ReadFile(queueStatePath(root))
	if err != nil {
		return queueState{}, err
	}
	var state queueState
	if err := json.Unmarshal(data, &state); err != nil {
		return queueState{}, fmt.Errorf("parse queue state: %w", err)
	}
	if state.Specs == nil {
		state.Specs = map[string]queueSpec{}
	}
	return state, nil
}

func (a *App) writeQueueState(root string, state queueState) error {
	if state.Specs == nil {
		state.Specs = map[string]queueSpec{}
	}
	state.CompletedSpecCount = len(state.CompletedSpecs)
	state.SkippedSpecCount = len(state.SkippedSpecs)
	state.UpdatedAt = firstNonBlank(state.UpdatedAt, a.now().Format(time.RFC3339))
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	path := queueStatePath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	return os.WriteFile(queueReportPath(root), []byte(formatQueueReport(state, true)), 0o644)
}

func queueIsActive(state queueState) bool {
	switch state.Status {
	case "", queueStateActive, queueStateWaiting, queueStateBlocked, queueStatePaused:
		return true
	default:
		return false
	}
}

func refreshQueueControlRequests(root string, state queueState) (queueState, error) {
	latest, err := readQueueState(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return state, nil
		}
		return state, err
	}
	if state.ID != "" && latest.ID != "" && latest.ID != state.ID {
		return state, fmt.Errorf("queue state changed from %s to %s", state.ID, latest.ID)
	}
	if latest.PauseRequested {
		state.PauseRequested = true
	}
	if latest.StopRequested {
		state.StopRequested = true
	}
	return state, nil
}

func (a *App) honorQueueControlRequests(root string, state queueState) (queueState, bool, error) {
	refreshed, err := refreshQueueControlRequests(root, state)
	if err != nil {
		return state, false, err
	}
	state = refreshed
	switch {
	case state.StopRequested:
		state.Status = queueStateStopped
		state.OperatorState = queueOperatorBlocked
		state.Detail = queuePhaseStopped
		state.LastRecoveryAction = "start a new queue when ready"
	case state.PauseRequested:
		state.Status = queueStatePaused
		state.OperatorState = queueOperatorWaiting
		state.Detail = queuePhasePaused
		state.LastRecoveryAction = "run `namba queue resume` when ready"
	default:
		return state, false, nil
	}
	state.UpdatedAt = a.now().Format(time.RFC3339)
	if err := a.writeQueueState(root, state); err != nil {
		return state, true, err
	}
	return state, true, nil
}

func updateQueueSpecState(a *App, root string, state queueState, specID string, specState queueSpec, operatorState, detail string) (queueState, error) {
	specState.Status = queueStateActive
	specState.OperatorState = operatorState
	specState.LastCheckpoint = detail
	state.Specs[specID] = specState
	state.OperatorState = operatorState
	state.Detail = detail
	state.LastSafeCheckpoint = specID + ":" + detail
	state.UpdatedAt = a.now().Format(time.RFC3339)
	if err := a.writeQueueState(root, state); err != nil {
		return state, err
	}
	return state, nil
}

func blockQueueSpec(a *App, root string, state queueState, specID, gate, evidencePath, recovery, detail string) (queueState, bool, error) {
	specState := state.Specs[specID]
	specState.SpecID = specID
	specState.Status = queueStateBlocked
	specState.OperatorState = queueOperatorBlocked
	specState.Phase = queuePhaseBlocked
	specState.Blocker = gate
	specState.EvidencePath = evidencePath
	specState.RecoveryAction = recovery
	state.Specs[specID] = specState
	state.Status = queueStateBlocked
	state.OperatorState = queueOperatorBlocked
	state.Detail = gate
	state.LastBlocker = firstNonBlank(detail, gate)
	state.LastEvidencePath = evidencePath
	state.LastRecoveryAction = recovery
	state.LastSafeCheckpoint = specID + ":blocked"
	state.UpdatedAt = a.now().Format(time.RFC3339)
	return state, false, a.writeQueueState(root, state)
}

func markQueueSpecDone(state queueState, specID string, specState queueSpec) queueState {
	specState.Status = queueStateDone
	specState.OperatorState = queueOperatorDone
	specState.Phase = queuePhaseLanded
	state.Specs[specID] = specState
	if !queueContainsString(state.CompletedSpecs, specID) {
		state.CompletedSpecs = append(state.CompletedSpecs, specID)
	}
	return state
}

func markQueueSpecSkipped(state queueState, specID string, specState queueSpec, reason string) queueState {
	specState.Status = queueStateDone
	specState.OperatorState = queueOperatorDone
	specState.Phase = queuePhaseSkipped
	specState.SkipReason = reason
	state.Specs[specID] = specState
	if !queueContainsString(state.SkippedSpecs, specID) {
		state.SkippedSpecs = append(state.SkippedSpecs, specID)
	}
	return state
}

func queueContainsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (a *App) printQueueState(root string, state queueState, verbose bool) error {
	_, err := fmt.Fprint(a.stdout, formatQueueReport(state, verbose))
	if err != nil {
		return err
	}
	return os.WriteFile(queueReportPath(root), []byte(formatQueueReport(state, true)), 0o644)
}

func formatQueueReport(state queueState, verbose bool) string {
	lines := []string{
		fmt.Sprintf("Queue: %s", firstNonBlank(state.ID, "unknown")),
		fmt.Sprintf("State: %s", firstNonBlank(state.OperatorState, queueOperatorWaiting)),
		fmt.Sprintf("Detail: %s", firstNonBlank(state.Detail, "none")),
		fmt.Sprintf("Active SPEC: %s", firstNonBlank(state.ActiveSpecID, "none")),
		fmt.Sprintf("Progress: %d/%d done, %d skipped", len(state.CompletedSpecs), len(state.Targets), len(state.SkippedSpecs)),
	}
	if state.LastBlocker != "" {
		lines = append(lines, "Blocker: "+state.LastBlocker)
	}
	if state.LastEvidencePath != "" {
		lines = append(lines, "Evidence: "+state.LastEvidencePath)
	}
	if state.LastRecoveryAction != "" {
		lines = append(lines, "Next: "+state.LastRecoveryAction)
	} else {
		switch state.Status {
		case queueStateDone:
			lines = append(lines, "Next: queue complete")
		case queueStatePaused:
			lines = append(lines, "Next: namba queue resume")
		case queueStateStopped:
			lines = append(lines, "Next: namba queue start <SPEC-RANGE|SPEC-LIST>")
		case queueStateWaiting:
			lines = append(lines, "Next: satisfy the wait condition, then run `namba queue resume`")
		case queueStateBlocked:
			lines = append(lines, "Next: resolve the blocker, then run `namba queue resume`")
		default:
			lines = append(lines, "Next: namba queue status")
		}
	}
	if verbose {
		lines = append(lines, "", "## Targets")
		for _, specID := range state.Targets {
			specState := state.Specs[specID]
			lines = append(lines, fmt.Sprintf("- %s: %s/%s branch=%s pr=%d", specID, firstNonBlank(specState.OperatorState, "waiting"), firstNonBlank(specState.Phase, queuePhasePending), firstNonBlank(specState.Branch, "n/a"), specState.PRNumber))
			if specState.Blocker != "" {
				lines = append(lines, fmt.Sprintf("  blocker: %s", specState.Blocker))
			}
			if specState.SkipReason != "" {
				lines = append(lines, fmt.Sprintf("  skipped: %s", specState.SkipReason))
			}
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

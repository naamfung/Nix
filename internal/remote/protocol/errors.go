package protocol

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"inx/internal/rpcwire"
)

const DomainErrorCode = -32000

type InxErrorCode string

const (
	ErrRemoteNotInstalled         InxErrorCode = "REMOTE_NOT_INSTALLED"
	ErrHostStopped                InxErrorCode = "HOST_STOPPED"
	ErrVersionMismatch            InxErrorCode = "VERSION_MISMATCH"
	ErrDaemonRestartRequired      InxErrorCode = "DAEMON_RESTART_REQUIRED"
	ErrHostBusy                   InxErrorCode = "HOST_BUSY"
	ErrStaleHostEpoch             InxErrorCode = "STALE_HOST_EPOCH"
	ErrStaleRuntimeEpoch          InxErrorCode = "STALE_RUNTIME_EPOCH"
	ErrRequestIDConflict          InxErrorCode = "REQUEST_ID_CONFLICT"
	ErrLeaseNotHeld               InxErrorCode = "LEASE_NOT_HELD"
	ErrStaleConnection            InxErrorCode = "STALE_CONNECTION"
	ErrStaleDirectoryRef          InxErrorCode = "STALE_DIRECTORY_REF"
	ErrDirectoryNotFound          InxErrorCode = "DIRECTORY_NOT_FOUND"
	ErrNotDirectory               InxErrorCode = "NOT_DIRECTORY"
	ErrPermissionDenied           InxErrorCode = "PERMISSION_DENIED"
	ErrWorkspaceNotFound          InxErrorCode = "WORKSPACE_NOT_FOUND"
	ErrWorkspaceInUse             InxErrorCode = "WORKSPACE_IN_USE"
	ErrSessionNotFound            InxErrorCode = "SESSION_NOT_FOUND"
	ErrWorkspaceSessionMismatch   InxErrorCode = "WORKSPACE_SESSION_MISMATCH"
	ErrRuntimeStartFailed         InxErrorCode = "RUNTIME_START_FAILED"
	ErrSessionPersistFailed       InxErrorCode = "SESSION_PERSIST_FAILED"
	ErrSessionTrashed             InxErrorCode = "SESSION_TRASHED"
	ErrSessionBusy                InxErrorCode = "SESSION_BUSY"
	ErrSessionCleanupPending      InxErrorCode = "SESSION_CLEANUP_PENDING"
	ErrTopicNotFound              InxErrorCode = "TOPIC_NOT_FOUND"
	ErrTopicNotEmpty              InxErrorCode = "TOPIC_NOT_EMPTY"
	ErrTrashEntryNotFound         InxErrorCode = "TRASH_ENTRY_NOT_FOUND"
	ErrRecoveryGuardFailed        InxErrorCode = "RECOVERY_GUARD_FAILED"
	ErrInvalidProfile             InxErrorCode = "INVALID_PROFILE"
	ErrModelNotAvailable          InxErrorCode = "MODEL_NOT_AVAILABLE"
	ErrEffortNotSupported         InxErrorCode = "EFFORT_NOT_SUPPORTED"
	ErrTurnAlreadyRunning         InxErrorCode = "TURN_ALREADY_RUNNING"
	ErrTurnNotActive              InxErrorCode = "TURN_NOT_ACTIVE"
	ErrTurnMismatch               InxErrorCode = "TURN_MISMATCH"
	ErrOperationNotActive         InxErrorCode = "OPERATION_NOT_ACTIVE"
	ErrOperationMismatch          InxErrorCode = "OPERATION_MISMATCH"
	ErrPromptNotPending           InxErrorCode = "PROMPT_NOT_PENDING"
	ErrPromptKindMismatch         InxErrorCode = "PROMPT_KIND_MISMATCH"
	ErrPromptDecisionNotAllowed   InxErrorCode = "PROMPT_DECISION_NOT_ALLOWED"
	ErrSnapshotExpired            InxErrorCode = "SNAPSHOT_EXPIRED"
	ErrSubscriptionNotFound       InxErrorCode = "SUBSCRIPTION_NOT_FOUND"
	ErrContentRefExpired          InxErrorCode = "CONTENT_REF_EXPIRED"
	ErrCheckpointNotFound         InxErrorCode = "CHECKPOINT_NOT_FOUND"
	ErrCheckpointScopeUnavailable InxErrorCode = "CHECKPOINT_SCOPE_UNAVAILABLE"
	ErrRewindPartial              InxErrorCode = "REWIND_PARTIAL"
	ErrStaleCursor                InxErrorCode = "STALE_CURSOR"
	ErrPathNotFound               InxErrorCode = "PATH_NOT_FOUND"
	ErrNotFile                    InxErrorCode = "NOT_FILE"
	ErrGitUnavailable             InxErrorCode = "GIT_UNAVAILABLE"
	ErrGitObjectNotFound          InxErrorCode = "GIT_OBJECT_NOT_FOUND"
	ErrQueryFailed                InxErrorCode = "QUERY_FAILED"
	ErrCapabilityUnavailable      InxErrorCode = "CAPABILITY_UNAVAILABLE"
)

type RemoteErrorData struct {
	InxCode                    InxErrorCode   `json:"inxCode"`
	Retryable                  bool           `json:"retryable"`
	Action                     RemoteAction   `json:"action,omitempty"`
	Target                     *RuntimeTarget `json:"target,omitempty"`
	Expected                   string         `json:"expected,omitempty"`
	Actual                     string         `json:"actual,omitempty"`
	RetryAfterMs               *int64         `json:"retryAfterMs,omitempty"`
	SuggestedCommand           string         `json:"suggestedCommand,omitempty"`
	WorkspaceMayHaveChanged    *bool          `json:"workspaceMayHaveChanged,omitempty"`
	ConversationMayHaveChanged *bool          `json:"conversationMayHaveChanged,omitempty"`
	SnapshotRequired           *bool          `json:"snapshotRequired,omitempty"`
}

type errorSpec struct {
	Message   string
	Retryable bool
	Action    RemoteAction
	Command   string
}

var frozenErrorSpecs = map[InxErrorCode]errorSpec{
	ErrRemoteNotInstalled:         {"Inx Remote is not installed on the Host.", false, ActionRunCommand, "inx remote install"},
	ErrHostStopped:                {"Inx Remote is not running on the Host.", true, ActionRunCommand, "inx remote start"},
	ErrVersionMismatch:            {"Desktop and Host CLI builds do not match.", false, ActionNone, ""},
	ErrDaemonRestartRequired:      {"The running daemon does not match the installed CLI.", false, ActionRestartDaemon, "inx remote restart"},
	ErrHostBusy:                   {"Another client currently holds the Host lease.", true, ActionRetry, ""},
	ErrStaleHostEpoch:             {"The Host runtime has restarted.", true, ActionReconnect, ""},
	ErrStaleRuntimeEpoch:          {"The Session runtime has been replaced.", true, ActionResubscribe, ""},
	ErrRequestIDConflict:          {"The request identifier was reused for a different operation.", false, ActionNone, ""},
	ErrLeaseNotHeld:               {"The client no longer holds the Host lease.", true, ActionReconnect, ""},
	ErrStaleConnection:            {"This transport has been replaced by a newer connection.", true, ActionReconnect, ""},
	ErrStaleDirectoryRef:          {"The directory selection has expired.", true, ActionRetry, ""},
	ErrDirectoryNotFound:          {"The selected directory no longer exists.", false, ActionNone, ""},
	ErrNotDirectory:               {"The selected path is not a directory.", false, ActionNone, ""},
	ErrPermissionDenied:           {"The Host user cannot access the requested resource.", false, ActionNone, ""},
	ErrWorkspaceNotFound:          {"The workspace is not available on this Host.", false, ActionNone, ""},
	ErrWorkspaceInUse:             {"The workspace still has active Session work.", true, ActionRetry, ""},
	ErrSessionNotFound:            {"The Session is not available.", false, ActionNone, ""},
	ErrWorkspaceSessionMismatch:   {"The Session does not belong to the requested workspace.", false, ActionNone, ""},
	ErrRuntimeStartFailed:         {"The Session runtime could not be started.", true, ActionRetry, ""},
	ErrSessionPersistFailed:       {"The Session state could not be persisted.", true, ActionRetry, ""},
	ErrSessionTrashed:             {"The Session is in the trash.", false, ActionNone, ""},
	ErrSessionBusy:                {"The Session is busy with conflicting work.", true, ActionRetry, ""},
	ErrSessionCleanupPending:      {"Session cleanup is still in progress.", true, ActionRetry, ""},
	ErrTopicNotFound:              {"The Topic is not available.", false, ActionNone, ""},
	ErrTopicNotEmpty:              {"Only an empty Topic can be deleted.", false, ActionNone, ""},
	ErrTrashEntryNotFound:         {"The trash entry is not available.", false, ActionNone, ""},
	ErrRecoveryGuardFailed:        {"The recovery-only safety check failed.", false, ActionNone, ""},
	ErrInvalidProfile:             {"The requested Session profile is invalid.", false, ActionNone, ""},
	ErrModelNotAvailable:          {"The requested model is not available on the Host.", false, ActionNone, ""},
	ErrEffortNotSupported:         {"The requested effort is not supported by the model.", false, ActionNone, ""},
	ErrTurnAlreadyRunning:         {"A Turn is already running in this Session.", true, ActionRetry, ""},
	ErrTurnNotActive:              {"There is no active Turn.", false, ActionNone, ""},
	ErrTurnMismatch:               {"The active Turn does not match the requested Turn.", false, ActionResubscribe, ""},
	ErrOperationNotActive:         {"There is no active Operation.", false, ActionNone, ""},
	ErrOperationMismatch:          {"The active Operation does not match the requested Operation.", false, ActionResubscribe, ""},
	ErrPromptNotPending:           {"The Prompt is no longer pending.", false, ActionResubscribe, ""},
	ErrPromptKindMismatch:         {"The Prompt type does not match this response.", false, ActionResubscribe, ""},
	ErrPromptDecisionNotAllowed:   {"The requested decision is not allowed for this Prompt.", false, ActionNone, ""},
	ErrSnapshotExpired:            {"The Session snapshot has expired.", true, ActionResubscribe, ""},
	ErrSubscriptionNotFound:       {"The subscription is not available on this transport.", true, ActionResubscribe, ""},
	ErrContentRefExpired:          {"The referenced content has expired.", true, ActionResubscribe, ""},
	ErrCheckpointNotFound:         {"The checkpoint is not available in this runtime.", false, ActionResubscribe, ""},
	ErrCheckpointScopeUnavailable: {"The checkpoint does not support the requested rewind scope.", false, ActionNone, ""},
	ErrRewindPartial:              {"Rewind did not complete and state may have changed.", false, ActionResubscribe, ""},
	ErrStaleCursor:                {"The query cursor has expired.", true, ActionRetry, ""},
	ErrPathNotFound:               {"The requested workspace path does not exist.", false, ActionNone, ""},
	ErrNotFile:                    {"The requested path is not a regular file.", false, ActionNone, ""},
	ErrGitUnavailable:             {"Git information is not available for this workspace.", false, ActionNone, ""},
	ErrGitObjectNotFound:          {"The requested Git object is not available.", false, ActionNone, ""},
	ErrQueryFailed:                {"The Host query could not be completed.", true, ActionRetry, ""},
	ErrCapabilityUnavailable:      {"This capability is unavailable on the Host.", false, ActionNone, ""},
}

type ErrorOptions struct {
	Target                     *RuntimeTarget
	Expected                   string
	Actual                     string
	RetryAfterMs               *int64
	WorkspaceMayHaveChanged    *bool
	ConversationMayHaveChanged *bool
	SnapshotRequired           *bool
}

type RemoteError struct {
	Code    InxErrorCode
	Message string
	Data    RemoteErrorData
}

func (e *RemoteError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *RemoteError) RPCError() *rpcwire.RPCError {
	if e == nil {
		return &rpcwire.RPCError{Code: rpcwire.ErrInternal, Message: "internal error"}
	}
	return &rpcwire.RPCError{Code: DomainErrorCode, Message: e.Message, Data: e.Data}
}

func NewRemoteError(code InxErrorCode, options ErrorOptions) (*RemoteError, error) {
	spec, ok := frozenErrorSpecs[code]
	if !ok {
		return nil, fmt.Errorf("protocol: unknown Inx error code %q", code)
	}
	data := RemoteErrorData{
		InxCode: code, Retryable: spec.Retryable, Action: spec.Action,
		Target: options.Target, Expected: strings.TrimSpace(options.Expected),
		Actual: strings.TrimSpace(options.Actual), RetryAfterMs: options.RetryAfterMs,
		SuggestedCommand:           spec.Command,
		WorkspaceMayHaveChanged:    options.WorkspaceMayHaveChanged,
		ConversationMayHaveChanged: options.ConversationMayHaveChanged,
		SnapshotRequired:           options.SnapshotRequired,
	}
	if err := data.Validate(); err != nil {
		return nil, err
	}
	return &RemoteError{Code: code, Message: spec.Message, Data: data}, nil
}

func MustRemoteError(code InxErrorCode, options ErrorOptions) *RemoteError {
	errValue, err := NewRemoteError(code, options)
	if err != nil {
		panic(err)
	}
	return errValue
}

func (d RemoteErrorData) Validate() error {
	spec, ok := frozenErrorSpecs[d.InxCode]
	if !ok {
		return fmt.Errorf("unknown inxCode %q", d.InxCode)
	}
	if d.Retryable != spec.Retryable || d.Action != spec.Action || d.SuggestedCommand != spec.Command {
		return errors.New("retryable, action, and suggestedCommand must match the frozen error table")
	}
	if d.Target != nil {
		if err := d.Target.Validate(); err != nil {
			return fmt.Errorf("target: %w", err)
		}
	}
	if (d.Expected == "") != (d.Actual == "") {
		return errors.New("expected and actual must be supplied together")
	}
	if err := validateControlledDiagnosticValue("expected", d.Expected); err != nil {
		return err
	}
	if err := validateControlledDiagnosticValue("actual", d.Actual); err != nil {
		return err
	}
	if d.InxCode == ErrHostBusy {
		if d.RetryAfterMs == nil || *d.RetryAfterMs < 0 {
			return errors.New("HOST_BUSY requires a non-negative retryAfterMs")
		}
	} else if d.RetryAfterMs != nil {
		return errors.New("retryAfterMs is only valid for HOST_BUSY")
	}
	rewindFields := d.WorkspaceMayHaveChanged != nil || d.ConversationMayHaveChanged != nil || d.SnapshotRequired != nil
	if d.InxCode == ErrRewindPartial {
		if d.WorkspaceMayHaveChanged == nil || d.ConversationMayHaveChanged == nil || d.SnapshotRequired == nil {
			return errors.New("REWIND_PARTIAL requires all change and snapshot flags")
		}
		if !*d.SnapshotRequired || (!*d.WorkspaceMayHaveChanged && !*d.ConversationMayHaveChanged) {
			return errors.New("REWIND_PARTIAL requires snapshot refresh and at least one possibly changed range")
		}
	} else if rewindFields {
		return errors.New("change and snapshot flags are only valid for REWIND_PARTIAL")
	}
	return nil
}

func validateControlledDiagnosticValue(name, value string) error {
	if value == "" {
		return nil
	}
	if len(value) > 256 || strings.ContainsAny(value, "\x00\r\n/\\") {
		return fmt.Errorf("%s must be a bounded path-free diagnostic token", name)
	}
	return nil
}

type ErrorContract struct {
	InxCode          InxErrorCode `json:"inxCode"`
	JSONRPCCode      int          `json:"jsonRpcCode"`
	Message          string       `json:"message"`
	Retryable        bool         `json:"retryable"`
	Action           RemoteAction `json:"action,omitempty"`
	SuggestedCommand string       `json:"suggestedCommand,omitempty"`
}

func ErrorContracts() []ErrorContract {
	out := make([]ErrorContract, 0, len(frozenErrorSpecs))
	for code, spec := range frozenErrorSpecs {
		out = append(out, ErrorContract{code, DomainErrorCode, spec.Message, spec.Retryable, spec.Action, spec.Command})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].InxCode < out[j].InxCode })
	return out
}

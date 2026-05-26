package agentsobs

import "github.com/msmclass/samvaad/pkg/proto/samvaad"

func JobKindFromProto(kind samvaad.JobType) JobKind {
	switch kind {
	case samvaad.JobType_JT_ROOM:
		return JobKindRoom
	case samvaad.JobType_JT_PUBLISHER:
		return JobKindPublisher
	case samvaad.JobType_JT_PARTICIPANT:
		return JobKindParticipant
	default:
		return JobKindUndefined
	}
}

func JobStatusFromProto(status samvaad.JobStatus) JobStatus {
	switch status {
	case samvaad.JobStatus_JS_PENDING:
		return JobStatusPending
	case samvaad.JobStatus_JS_RUNNING:
		return JobStatusRunning
	case samvaad.JobStatus_JS_SUCCESS:
		return JobStatusSuccess
	case samvaad.JobStatus_JS_FAILED:
		return JobStatusFailed
	default:
		return JobStatusUndefined
	}
}

func WorkerStatusFromProto(status samvaad.WorkerStatus) WorkerStatus {
	switch status {
	case samvaad.WorkerStatus_WS_AVAILABLE:
		return WorkerStatusAvailable
	case samvaad.WorkerStatus_WS_FULL:
		return WorkerStatusFull
	default:
		return WorkerStatusUndefined
	}
}

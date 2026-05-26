package agentsobs

import (
	"testing"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/stretchr/testify/require"
)

func TestJobKindFromProto(t *testing.T) {
	tests := []struct {
		input    samvaad.JobType
		expected JobKind
	}{
		{samvaad.JobType_JT_ROOM, JobKindRoom},
		{samvaad.JobType_JT_PUBLISHER, JobKindPublisher},
		{samvaad.JobType_JT_PARTICIPANT, JobKindParticipant},
		{samvaad.JobType(999), JobKindUndefined}, // Undefined case
	}

	for _, test := range tests {
		result := JobKindFromProto(test.input)
		require.Equal(t, test.expected, result)
	}
}

func TestJobStatusFromProto(t *testing.T) {
	tests := []struct {
		input    samvaad.JobStatus
		expected JobStatus
	}{
		{samvaad.JobStatus_JS_PENDING, JobStatusPending},
		{samvaad.JobStatus_JS_RUNNING, JobStatusRunning},
		{samvaad.JobStatus_JS_SUCCESS, JobStatusSuccess},
		{samvaad.JobStatus_JS_FAILED, JobStatusFailed},
		{samvaad.JobStatus(999), JobStatusUndefined}, // Undefined case
	}

	for _, test := range tests {
		result := JobStatusFromProto(test.input)
		require.Equal(t, test.expected, result)
	}
}

func TestWorkerStatusFromProto(t *testing.T) {
	tests := []struct {
		input    samvaad.WorkerStatus
		expected WorkerStatus
	}{
		{samvaad.WorkerStatus_WS_AVAILABLE, WorkerStatusAvailable},
		{samvaad.WorkerStatus_WS_FULL, WorkerStatusFull},
		{samvaad.WorkerStatus(999), WorkerStatusUndefined}, // Undefined case
	}

	for _, test := range tests {
		result := WorkerStatusFromProto(test.input)
		require.Equal(t, test.expected, result)
	}
}

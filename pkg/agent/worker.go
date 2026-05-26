// Copyright 2026 Samvaad Project, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package agent

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"sync"
	"time"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"

	protoagent "github.com/msmclass/samvaad/pkg/samvaad/agent"
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/logger"
	"github.com/msmclass/samvaad/pkg/samvaad/rpc"
	"github.com/msmclass/samvaad/pkg/samvaad/utils"
	"github.com/msmclass/samvaad/pkg/samvaad/utils/guid"
	"github.com/msmclass/samvaad/pkg/samvaad/psrpc"
)

var (
	ErrUnimplementedWrorkerSignal = errors.New("unimplemented worker signal")
	ErrUnknownWorkerSignal        = errors.New("unknown worker signal")
	ErrUnknownJobType             = errors.New("unknown job type")
	ErrJobNotFound                = psrpc.NewErrorf(psrpc.NotFound, "no running job for given jobID")
	ErrWorkerClosed               = errors.New("worker closed")
	ErrWorkerNotAvailable         = errors.New("worker not available")
	ErrAvailabilityTimeout        = errors.New("agent worker availability timeout")
	ErrDuplicateJobAssignment     = errors.New("duplicate job assignment")
)

const AgentNameAttributeKey = "lk.agent_name"

type WorkerProtocolVersion int

const CurrentProtocol = 1

const (
	RegisterTimeout  = 10 * time.Second
	AssignJobTimeout = 10 * time.Second
)

type SignalConn interface {
	WriteServerMessage(msg *samvaad.ServerMessage) (int, error)
	ReadWorkerMessage() (*samvaad.WorkerMessage, int, error)
	SetReadDeadline(time.Time) error
	Close() error
	CloseWithReason(reason string) error
}

func JobStatusIsEnded(s samvaad.JobStatus) bool {
	return s == samvaad.JobStatus_JS_SUCCESS || s == samvaad.JobStatus_JS_FAILED
}

type AssignmentHook func(next func(*samvaad.JobAssignment) error) func(*samvaad.JobAssignment) error

type WorkerSignalHandler interface {
	HandleRegister(*samvaad.RegisterWorkerRequest) error
	HandleAvailability(*samvaad.AvailabilityResponse) error
	HandleUpdateJob(*samvaad.UpdateJobStatus) error
	HandleSimulateJob(*samvaad.SimulateJobRequest) error
	HandlePing(*samvaad.WorkerPing) error
	HandleUpdateWorker(*samvaad.UpdateWorkerStatus) error
	HandleMigrateJob(*samvaad.MigrateJobRequest) error
}

func DispatchWorkerSignal(req *samvaad.WorkerMessage, h WorkerSignalHandler) error {
	switch m := req.Message.(type) {
	case *samvaad.WorkerMessage_Register:
		return h.HandleRegister(m.Register)
	case *samvaad.WorkerMessage_Availability:
		return h.HandleAvailability(m.Availability)
	case *samvaad.WorkerMessage_UpdateJob:
		return h.HandleUpdateJob(m.UpdateJob)
	case *samvaad.WorkerMessage_SimulateJob:
		return h.HandleSimulateJob(m.SimulateJob)
	case *samvaad.WorkerMessage_Ping:
		return h.HandlePing(m.Ping)
	case *samvaad.WorkerMessage_UpdateWorker:
		return h.HandleUpdateWorker(m.UpdateWorker)
	case *samvaad.WorkerMessage_MigrateJob:
		return h.HandleMigrateJob(m.MigrateJob)
	default:
		return ErrUnknownWorkerSignal
	}
}

var _ WorkerSignalHandler = (*UnimplementedWorkerSignalHandler)(nil)

type UnimplementedWorkerSignalHandler struct{}

func (UnimplementedWorkerSignalHandler) HandleRegister(*samvaad.RegisterWorkerRequest) error {
	return fmt.Errorf("%w: Register", ErrUnimplementedWrorkerSignal)
}
func (UnimplementedWorkerSignalHandler) HandleAvailability(*samvaad.AvailabilityResponse) error {
	return fmt.Errorf("%w: Availability", ErrUnimplementedWrorkerSignal)
}
func (UnimplementedWorkerSignalHandler) HandleUpdateJob(*samvaad.UpdateJobStatus) error {
	return fmt.Errorf("%w: UpdateJob", ErrUnimplementedWrorkerSignal)
}
func (UnimplementedWorkerSignalHandler) HandleSimulateJob(*samvaad.SimulateJobRequest) error {
	return fmt.Errorf("%w: SimulateJob", ErrUnimplementedWrorkerSignal)
}
func (UnimplementedWorkerSignalHandler) HandlePing(*samvaad.WorkerPing) error {
	return fmt.Errorf("%w: Ping", ErrUnimplementedWrorkerSignal)
}
func (UnimplementedWorkerSignalHandler) HandleUpdateWorker(*samvaad.UpdateWorkerStatus) error {
	return fmt.Errorf("%w: UpdateWorker", ErrUnimplementedWrorkerSignal)
}
func (UnimplementedWorkerSignalHandler) HandleMigrateJob(*samvaad.MigrateJobRequest) error {
	return fmt.Errorf("%w: MigrateJob", ErrUnimplementedWrorkerSignal)
}

type WorkerPingHandler struct {
	UnimplementedWorkerSignalHandler
	conn SignalConn
}

func (h WorkerPingHandler) HandlePing(ping *samvaad.WorkerPing) error {
	_, err := h.conn.WriteServerMessage(&samvaad.ServerMessage{
		Message: &samvaad.ServerMessage_Pong{
			Pong: &samvaad.WorkerPong{
				LastTimestamp: ping.Timestamp,
				Timestamp:     time.Now().UnixMilli(),
			},
		},
	})
	return err
}

type WorkerRegistration struct {
	Protocol    WorkerProtocolVersion
	ID          string
	Version     string
	AgentID     string
	AgentName   string
	Namespace   string
	JobType     samvaad.JobType
	Permissions *samvaad.ParticipantPermission
	ClientIP    string
	Deployment  string
}

func MakeWorkerRegistration() WorkerRegistration {
	return WorkerRegistration{
		ID:       guid.New(guid.AgentWorkerPrefix),
		Protocol: CurrentProtocol,
	}
}

var _ WorkerSignalHandler = (*WorkerRegisterer)(nil)

type WorkerRegisterer struct {
	WorkerPingHandler
	serverInfo *samvaad.ServerInfo
	deadline   time.Time

	registration WorkerRegistration
	registered   bool
}

func NewWorkerRegisterer(conn SignalConn, serverInfo *samvaad.ServerInfo, base WorkerRegistration) *WorkerRegisterer {
	return &WorkerRegisterer{
		WorkerPingHandler: WorkerPingHandler{conn: conn},
		serverInfo:        serverInfo,
		registration:      base,
		deadline:          time.Now().Add(RegisterTimeout),
	}
}

func (h *WorkerRegisterer) Deadline() time.Time {
	return h.deadline
}

func (h *WorkerRegisterer) Registration() WorkerRegistration {
	return h.registration
}

func (h *WorkerRegisterer) Registered() bool {
	return h.registered
}

func (h *WorkerRegisterer) HandleRegister(req *samvaad.RegisterWorkerRequest) error {
	if !samvaad.IsJobType(req.GetType()) {
		return ErrUnknownJobType
	}

	if err := protoagent.ValidateDeployment(req.GetDeployment()); err != nil {
		return err
	}

	permissions := req.AllowedPermissions
	if permissions == nil {
		permissions = &samvaad.ParticipantPermission{
			CanSubscribe:      true,
			CanPublish:        true,
			CanPublishData:    true,
			CanUpdateMetadata: true,
		}
	}

	h.registration.Version = req.Version
	h.registration.AgentName = req.AgentName
	h.registration.Namespace = req.GetNamespace()
	h.registration.JobType = req.GetType()
	h.registration.Permissions = permissions
	h.registration.Deployment = req.GetDeployment()
	h.registered = true

	_, err := h.conn.WriteServerMessage(&samvaad.ServerMessage{
		Message: &samvaad.ServerMessage_Register{
			Register: &samvaad.RegisterWorkerResponse{
				WorkerId:   h.registration.ID,
				ServerInfo: h.serverInfo,
			},
		},
	})
	return err
}

var _ WorkerSignalHandler = (*Worker)(nil)

type Worker struct {
	WorkerPingHandler
	WorkerRegistration

	apiKey    string
	apiSecret string
	logger    logger.Logger

	ctx    context.Context
	cancel context.CancelFunc
	closed chan struct{}

	mu     sync.Mutex
	load   float32
	status samvaad.WorkerStatus

	runningJobs  map[samvaad.JobID]*samvaad.Job
	availability map[samvaad.JobID]chan *samvaad.AvailabilityResponse
}

func NewWorker(
	registration WorkerRegistration,
	apiKey string,
	apiSecret string,
	conn SignalConn,
	logger logger.Logger,
) *Worker {
	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		WorkerPingHandler:  WorkerPingHandler{conn: conn},
		WorkerRegistration: registration,
		apiKey:             apiKey,
		apiSecret:          apiSecret,
		logger: logger.WithValues(
			"workerID", registration.ID,
			"agentName", registration.AgentName,
			"deployment", registration.Deployment,
			"agentID", registration.AgentID,
			"version", registration.Version,
			"jobType", registration.JobType.String(),
		),

		ctx:    ctx,
		cancel: cancel,
		closed: make(chan struct{}),

		runningJobs:  make(map[samvaad.JobID]*samvaad.Job),
		availability: make(map[samvaad.JobID]chan *samvaad.AvailabilityResponse),
	}
}

func (w *Worker) APIKey() string {
	return w.apiKey
}

func (w *Worker) Status() samvaad.WorkerStatus {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.status
}

func (w *Worker) Load() float32 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.load
}

func (w *Worker) Logger() logger.Logger {
	return w.logger
}

func (w *Worker) RunningJobs() map[samvaad.JobID]*samvaad.Job {
	w.mu.Lock()
	defer w.mu.Unlock()
	return maps.Clone(w.runningJobs)
}

func (w *Worker) RunningJobCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.runningJobs)
}

func (w *Worker) GetJobState(jobID samvaad.JobID) (*samvaad.JobState, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	j, ok := w.runningJobs[jobID]
	if !ok {
		return nil, ErrJobNotFound
	}
	return utils.CloneProto(j.State), nil
}

func (w *Worker) AssignJob(ctx context.Context, job *samvaad.Job, hook AssignmentHook) (*samvaad.JobState, error) {
	availCh := make(chan *samvaad.AvailabilityResponse, 1)
	job = utils.CloneProto(job)
	jobID := samvaad.JobID(job.Id)

	w.mu.Lock()
	if _, ok := w.availability[jobID]; ok {
		w.mu.Unlock()
		return nil, ErrDuplicateJobAssignment
	}

	w.availability[jobID] = availCh
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		delete(w.availability, jobID)
		w.mu.Unlock()
	}()

	if job.State == nil {
		job.State = &samvaad.JobState{}
	}
	now := time.Now()
	job.State.WorkerId = w.ID
	job.State.AgentId = w.AgentID
	job.State.UpdatedAt = now.UnixNano()
	job.State.StartedAt = now.UnixNano()
	job.State.Status = samvaad.JobStatus_JS_RUNNING

	if _, err := w.conn.WriteServerMessage(&samvaad.ServerMessage{Message: &samvaad.ServerMessage_Availability{
		Availability: &samvaad.AvailabilityRequest{Job: job},
	}}); err != nil {
		return nil, err
	}

	timeout := time.NewTimer(AssignJobTimeout)
	defer timeout.Stop()

	// See handleAvailability for the response
	select {
	case res := <-availCh:
		if res.Terminate {
			job.State.EndedAt = now.UnixNano()
			job.State.Status = samvaad.JobStatus_JS_SUCCESS
			return job.State, nil
		}

		if !res.Available {
			return nil, ErrWorkerNotAvailable
		}

		job.State.ParticipantIdentity = res.ParticipantIdentity
		attributes := res.ParticipantAttributes
		if attributes == nil {
			attributes = make(map[string]string)
		}
		attributes[AgentNameAttributeKey] = w.AgentName

		token, err := protoagent.BuildAgentToken(
			w.apiKey,
			w.apiSecret,
			job.Room.Name,
			res.ParticipantIdentity,
			res.ParticipantName,
			res.ParticipantMetadata,
			attributes,
			w.Permissions,
		)
		if err != nil {
			w.logger.Errorw("failed to build agent token", err)
			return nil, err
		}

		send := func(a *samvaad.JobAssignment) error {
			_, err := w.conn.WriteServerMessage(&samvaad.ServerMessage{
				Message: &samvaad.ServerMessage_Assignment{Assignment: a},
			})
			return err
		}
		if hook != nil {
			send = hook(send)
		}
		if err := send(&samvaad.JobAssignment{Job: job, Token: token}); err != nil {
			return nil, err
		}

		state := utils.CloneProto(job.State)

		w.mu.Lock()
		w.runningJobs[jobID] = job
		w.mu.Unlock()

		// TODO sweep jobs that are never started. We can't do this until all SDKs actually update the the JOB state

		return state, nil
	case <-timeout.C:
		return nil, ErrAvailabilityTimeout
	case <-w.ctx.Done():
		return nil, ErrWorkerClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (w *Worker) TerminateJob(jobID samvaad.JobID, reason rpc.JobTerminateReason) (*samvaad.JobState, error) {
	w.mu.Lock()
	_, ok := w.runningJobs[jobID]
	w.mu.Unlock()

	if !ok {
		return nil, ErrJobNotFound
	}

	_, writeErr := w.conn.WriteServerMessage(&samvaad.ServerMessage{Message: &samvaad.ServerMessage_Termination{
		Termination: &samvaad.JobTermination{
			JobId: string(jobID),
		},
	}})

	status := samvaad.JobStatus_JS_SUCCESS
	errorStr := ""
	if reason == rpc.JobTerminateReason_AGENT_LEFT_ROOM {
		status = samvaad.JobStatus_JS_FAILED
		errorStr = "agent worker left the room"
	}

	state, updateErr := w.UpdateJobStatus(&samvaad.UpdateJobStatus{
		JobId:  string(jobID),
		Status: status,
		Error:  errorStr,
	})
	return state, multierr.Combine(writeErr, updateErr)
}

func (w *Worker) UpdateMetadata(metadata string) {
	w.logger.Debugw("worker metadata updated", nil, "metadata", metadata)
}

func (w *Worker) IsClosed() bool {
	select {
	case <-w.closed:
		return true
	default:
		return false
	}
}

func (w *Worker) Close() {
	w.mu.Lock()
	if w.IsClosed() {
		w.mu.Unlock()
		return
	}

	w.logger.Infow("closing worker", "workerID", w.ID, "jobType", w.JobType, "agentName", w.AgentName)

	close(w.closed)
	w.cancel()
	_ = w.conn.Close()
	w.mu.Unlock()
}

func (w *Worker) HandleAvailability(res *samvaad.AvailabilityResponse) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	jobID := samvaad.JobID(res.JobId)
	availCh, ok := w.availability[jobID]
	if !ok {
		w.logger.Warnw("received availability response for unknown job", nil, "jobID", jobID)
		return nil
	}

	availCh <- res
	delete(w.availability, jobID)

	return nil
}

func (w *Worker) HandleUpdateJob(update *samvaad.UpdateJobStatus) error {
	_, err := w.UpdateJobStatus(update)
	if err != nil {
		// treating this as a debug message only
		// this can happen if the Room closes first, which would delete the agent dispatch
		// that would mark the job as successful. subsequent updates from the same worker
		// would not be able to find the same jobID.
		w.logger.Debugw("received job update for unknown job", "jobID", update.JobId)
	}
	return nil
}

func (w *Worker) UpdateJobStatus(update *samvaad.UpdateJobStatus) (*samvaad.JobState, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	jobID := samvaad.JobID(update.JobId)
	job, ok := w.runningJobs[jobID]
	if !ok {
		return nil, psrpc.NewErrorf(psrpc.NotFound, "received job update for unknown job")
	}

	now := time.Now()
	job.State.UpdatedAt = now.UnixNano()

	if job.State.Status == samvaad.JobStatus_JS_PENDING && update.Status != samvaad.JobStatus_JS_PENDING {
		job.State.StartedAt = now.UnixNano()
	}

	job.State.Status = update.Status
	job.State.Error = update.Error

	if JobStatusIsEnded(update.Status) {
		job.State.EndedAt = now.UnixNano()
		delete(w.runningJobs, jobID)

		w.logger.Infow("job ended", "jobID", update.JobId, "status", update.Status, "error", update.Error)
	}

	return proto.Clone(job.State).(*samvaad.JobState), nil
}

func (w *Worker) HandleSimulateJob(simulate *samvaad.SimulateJobRequest) error {
	jobType := samvaad.JobType_JT_ROOM
	if simulate.Participant != nil {
		jobType = samvaad.JobType_JT_PUBLISHER
	}

	job := &samvaad.Job{
		Id:          guid.New(guid.AgentJobPrefix),
		Type:        jobType,
		Room:        simulate.Room,
		Participant: simulate.Participant,
		Namespace:   w.Namespace,
		AgentName:   w.AgentName,
	}

	go func() {
		_, err := w.AssignJob(w.ctx, job, nil)
		if err != nil {
			w.logger.Errorw("unable to simulate job", err, "jobID", job.Id)
		}
	}()

	return nil
}

func (w *Worker) HandleUpdateWorker(update *samvaad.UpdateWorkerStatus) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if status := update.Status; status != nil && w.status != *status {
		w.status = *status
		w.Logger().Debugw("worker status changed", "status", w.status)
	}
	w.load = update.GetLoad()

	return nil
}

func (w *Worker) HandleMigrateJob(req *samvaad.MigrateJobRequest) error {
	// TODO(theomonnom): On OSS this is not implemented
	// We could maybe just move a specific job to another worker
	return nil
}



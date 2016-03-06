package runtime

import (
	"fmt"
	"io"
	"sync"

	"github.com/taskcluster/taskcluster-client-go/queue"
	"github.com/taskcluster/taskcluster-client-go/tcclient"

	"gopkg.in/djherbis/stream.v1"
)

// An ExceptionReason specifies the reason a task reached an exception state.
type ExceptionReason string

// Reasons why a task can reach an exception state. Implementors should be
// warned that additional entries may be added in the future.
const (
	MalformedPayload ExceptionReason = "malformed-payload"
	WorkerShutdown   ExceptionReason = "worker-shutdown"
	InternalError    ExceptionReason = "internal-error"
)

// TaskStatus represents the current status of the task.
type TaskStatus string

// Enumerate task status to aid life-cycle decision making
// Use strings for benefit of simple logging/reporting
const (
	Aborted   TaskStatus = "Aborted"
	Cancelled TaskStatus = "Cancelled"
	Succeeded TaskStatus = "Succeeded"
	Failed    TaskStatus = "Failed"
	Errored   TaskStatus = "Errored"
	Claimed   TaskStatus = "Claimed"
	Reclaimed TaskStatus = "Reclaimed"
)

// The TaskInfo struct exposes generic properties from a task definition.
type TaskInfo struct {
	// TODO: Add fields and getters to get them
	RunId  int
	TaskId string
	sync.Mutex
	definition *queue.TaskDefinitionResponse
	claim      *queue.TaskClaimResponse
	status     *queue.TaskStatusStructure
	reclaim    *queue.TaskReclaimResponse
}

func (t TaskInfo) Definition() *queue.TaskDefinitionResponse {
	return t.definition
}

func (t TaskInfo) Status() *queue.TaskStatusStructure {
	return t.status
}

func (t TaskInfo) Credentials() *tcclient.Credentials {
	fmt.Println("herrrre")
	t.Lock()
	defer t.Unlock()

	fmt.Println("herrrre2")
	if t.reclaim != nil {
		return &tcclient.Credentials{
			ClientId:    t.reclaim.Credentials.ClientId,
			AccessToken: t.reclaim.Credentials.AccessToken,
			Certificate: t.reclaim.Credentials.Certificate,
		}
	}
	fmt.Println("herrrre3")
	return &tcclient.Credentials{
		ClientId:    t.claim.Credentials.ClientId,
		AccessToken: t.claim.Credentials.AccessToken,
		Certificate: t.claim.Credentials.Certificate,
	}
}

// The TaskContext exposes generic properties and functionality related to a
// task that is currently being executed.
//
// This context is used to ensure that every component both engines and plugins
// that operates on a task have access to some common information about the
// task. This includes log drains, per-task credentials, generic task
// properties, and abortion notifications.
type TaskContext struct {
	TaskInfo
	logStream *stream.Stream
	mu        sync.RWMutex
	status    TaskStatus
	cancelled bool
}

// TaskContextController exposes logic for controlling the TaskContext.
//
// Spliting this out from TaskContext ensures that engines and plugins doesn't
// accidentally Dispose() the TaskContext.
type TaskContextController struct {
	*TaskContext
}

// NewTaskContext creates a TaskContext and associated TaskContextController
func NewTaskContext(tempLogFile string, claim *queue.TaskClaimResponse) (*TaskContext, *TaskContextController, error) {
	logStream, err := stream.New(tempLogFile)
	if err != nil {
		return nil, nil, err
	}
	ctx := &TaskContext{
		logStream: logStream,
		TaskInfo: TaskInfo{
			claim:      claim,
			definition: &claim.Task,
			status:     &claim.Status,
			TaskId:     claim.Status.TaskId,
			RunId:      claim.RunId,
		},
	}
	return ctx, &TaskContextController{ctx}, nil
}

// CloseLog will close the log so no more messages can be written.
func (c *TaskContextController) CloseLog() error {
	return c.logStream.Close()
}

// Dispose will clean-up all resources held by the TaskContext
func (c *TaskContextController) Dispose() error {
	return c.logStream.Remove()
}

// Abort sets the status to aborted
func (c *TaskContext) Abort() {
	c.mu.Lock()
	c.status = Aborted
	c.mu.Unlock()
	return
}

// IsAborted returns true if the current status is Aborted
func (c *TaskContext) IsAborted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status == Aborted
}

// Cancel sets the status to cancelled
func (c *TaskContext) Cancel() {
	c.mu.Lock()
	c.status = Cancelled
	c.mu.Unlock()
	return
}

// IsCancelled returns true if the current status is Cancelled
func (c *TaskContext) IsCancelled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status == Cancelled
}

// Log writes a log message from the worker
//
// These log messages will be prefixed "[taskcluster]" so it's easy to see to
// that they are worker logs.
func (c *TaskContext) Log(a ...interface{}) {
	c.log("[taskcluster] ", a...)
}

// Log writes a log error message from the worker
//
// These log messages will be prefixed "[taskcluster:error]" so it's easy to see to
// that they are worker logs.  These errors are also easy to grep from the logs in
// case of failure.
func (c *TaskContext) LogError(a ...interface{}) {
	c.log("[taskcluster:error] ", a...)
}

func (c *TaskContext) log(prefix string, a ...interface{}) {
	a = append([]interface{}{prefix}, a...)
	_, err := fmt.Fprintln(c.logStream, a...)
	if err != nil {
		//TODO: Forward this to the system log, it's not a critical error
	}
}

// LogDrain returns a drain to which log message can be written.
//
// Users should note that multiple writers are writing to this drain
// concurrently, and it is recommend that writers write in chunks of one line.
func (c *TaskContext) LogDrain() io.Writer {
	return c.logStream
}

// NewLogReader returns a ReadCloser that reads the log from the start as the
// log is written.
//
// Calls to Read() on the resulting ReadCloser are blocking. They will return
// when data is written or EOF is reached.
//
// Consumers should ensure the ReadCloser is closed before discarding it.
func (c *TaskContext) NewLogReader() (io.ReadCloser, error) {
	return c.logStream.NextReader()
}

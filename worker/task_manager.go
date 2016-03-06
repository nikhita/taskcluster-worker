package worker

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	tcqueue "github.com/taskcluster/taskcluster-client-go/queue"
	"github.com/taskcluster/taskcluster-client-go/tcclient"
	"github.com/taskcluster/taskcluster-worker/config"
	"github.com/taskcluster/taskcluster-worker/engines"
	"github.com/taskcluster/taskcluster-worker/plugins"
	"github.com/taskcluster/taskcluster-worker/plugins/extpoints"
	"github.com/taskcluster/taskcluster-worker/runtime"
)

// Manager is resonsible for managing the entire task lifecyle from claiming the
// task, creating a sandbox environment, and reporting the results fo the execution.
// The manager will also be responsible for ensuring tasks do not run past their max run
// time and are aborted if a cancellation message is received.
type Manager struct {
	done          chan struct{}
	interval      int
	maxCapacity   int
	engine        engines.Engine
	environment   *runtime.Environment
	pluginManager plugins.Plugin
	pluginOptions *extpoints.PluginOptions
	log           *logrus.Entry
	queue         QueueService
	provisionerId string
	workerGroup   string
	workerId      string
	sync.RWMutex
	tasks map[string]*TaskRun
}

// Create a new instance of the task manager that will be responsible for claiming,
// executing, and resolving units of work (tasks).
func newTaskManager(config *config.Config, engine engines.Engine, environment *runtime.Environment, log *logrus.Entry) (*Manager, error) {
	queue := tcqueue.New(
		&tcclient.Credentials{
			ClientId:    config.Credentials.ClientId,
			AccessToken: config.Credentials.AccessToken,
			Certificate: config.Credentials.Certificate,
		},
	)
	service := &queueService{
		client:           queue,
		ProvisionerId:    config.ProvisionerId,
		WorkerGroup:      config.WorkerGroup,
		WorkerId:         config.WorkerId,
		WorkerType:       config.WorkerType,
		Log:              log.WithField("component", "Queue Service"),
		ExpirationOffset: config.QueueService.ExpirationOffset,
	}

	m := &Manager{
		done:          make(chan struct{}),
		tasks:         make(map[string]*TaskRun),
		engine:        engine,
		environment:   environment,
		interval:      config.PollingInterval,
		log:           log,
		maxCapacity:   config.Capacity,
		queue:         service,
		provisionerId: config.ProvisionerId,
		workerGroup:   config.WorkerGroup,
		workerId:      config.WorkerId,
	}
	fmt.Printf("%+v", m)

	m.pluginOptions = &extpoints.PluginOptions{
		Environment: environment,
		Engine:      &engine,
		Log:         log.WithField("component", "Plugin Manager"),
	}

	pm, err := extpoints.NewPluginManager([]string{"success", "livelog"}, *m.pluginOptions)
	if err != nil {
		log.WithField("error", err.Error()).Warn("Error creating task manager. Could not create plugin manager")
		return nil, err
	}

	m.pluginManager = pm
	return m, nil
}

// Start will initiliaze a polling cycle for tasks and spawn goroutines to
// execute units of work that has been claimed.
func (m *Manager) Start(stop <-chan struct{}, done chan struct{}) {
	m.log.Infof("Polling for tasks every %d seconds\n", m.interval)
	doWork := time.NewTicker(time.Duration(m.interval) * time.Second)
	for {
		select {
		case <-stop:
			doWork.Stop()
			go m.Stop()
		case <-doWork.C:
			n := math.Max(float64(m.maxCapacity-len(m.tasks)), 0)
			m.claimWork(int(n))
		case <-m.done:
			close(done)
			return
		}
	}
}

// Stop should be called when the worker should gracefully end the execution of
// all running tasks before completely shutting down.
func (m *Manager) Stop() {
	defer close(m.done)
	for _, v := range m.tasks {
		v.Abort()
	}

	done := make(chan struct{})
	go func() {
		for {
			if len(m.RunningTasks()) == 0 {
				close(done)
				return
			}
		}
	}()

	select {
	case <-done:
		return
	// Wait 2 minutes for all tasks to end, and then shutdown.
	case <-time.After(2 * time.Minute):
		close(done)
		return
	}
}

// RunningTasks returns the list of task names that are currently running. This could
// be useful for determining the number of tasks currently running or snapshotting
// the current running task list at a moment in time.
func (m *Manager) RunningTasks() []string {
	m.RLock()
	defer m.RUnlock()

	tasks := []string{}
	for k := range m.tasks {
		tasks = append(tasks, k)
	}

	return tasks
}

// CancelTask will cancel a running task.  Typically this will be called when a Pulse
// message is received to cancel a task.  Calling this method will not resolve the task
// as it's assumed that this task was already resolved as cancelled by another system/client.
func (m *Manager) CancelTask(taskId string, runId int) {
	name := fmt.Sprintf("%s/%d", taskId, runId)

	m.RLock()
	defer m.RUnlock()

	t, exists := m.tasks[name]
	if !exists {
		return
	}

	t.Cancel()
	return
}

func (m *Manager) claimWork(ntasks int) {
	if ntasks == 0 {

		return
	}

	tasks := m.queue.ClaimWork(ntasks)
	for _, t := range tasks {
		go m.run(t)
	}
}

func (m *Manager) run(task *TaskRun) {
	log := m.log.WithFields(logrus.Fields{
		"taskId": task.TaskId,
		"runId":  task.RunId,
	})
	task.log = log

	err := m.registerTask(task)

	if err != nil {
		log.WithField("error", err.Error()).Warn("Could not register task")
		panic(err)
	}

	defer m.deregisterTask(task)

	tp := m.environment.TemporaryStorage.NewFilePath()
	ctxt, ctxtctl, err := runtime.NewTaskContext(tp, &task.TaskClaim)
	if err != nil {
		log.WithField("error", err.Error()).Warn("Could not create task context")
		panic(err)
	}

	task.Run(m.pluginManager, m.engine, ctxt, ctxtctl)
	return
}

func (m *Manager) registerTask(task *TaskRun) error {
	name := fmt.Sprintf("%s/%d", task.TaskId, task.RunId)
	m.log.Debugf("Registered task: %s", name)

	m.Lock()
	defer m.Unlock()

	_, exists := m.tasks[name]
	if exists {
		return fmt.Errorf("Cannot register task %s. Task already exists.", name)
	}

	m.tasks[name] = task
	return nil
}

func (m *Manager) deregisterTask(task *TaskRun) error {
	name := fmt.Sprintf("%s/%d", task.TaskId, task.RunId)
	m.log.Debugf("Deregistered task: %s", name)

	m.Lock()
	defer m.Unlock()

	_, exists := m.tasks[name]
	if !exists {
		return fmt.Errorf("Cannot deregister task %s. Task does not exist", name)
	}

	delete(m.tasks, name)
	return nil
}

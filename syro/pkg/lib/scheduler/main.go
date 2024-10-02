package scheduler

import (
	"fmt"
	"sync"
	"syro/pkg/lib/errgroup"
	"time"

	"github.com/robfig/cron/v3"
)

// Scheduler is a wrapper around the cron scheduler that allows for the
// registration of jobs and the optional storage of job status and
// execution logs.
type Scheduler struct {
	cron    *cron.Cron // cron is the cron scheduler which will run the jobs
	Source  string     // Source is used to identify the source of the job
	Jobs    []*Job     // Jobs is a list of all registered jobs
	Storage Storage    // Storage is an optional storage interface for the scheduler
}

func NewScheduler(cron *cron.Cron, source string) *Scheduler {
	return &Scheduler{cron: cron, Source: source}
}

// WithStorage sets the storage for the scheduler.
func (s *Scheduler) WithStorage(storage Storage) *Scheduler {
	s.Storage = storage
	return s
}

// Register the cron job to the scheduler.
func (s *Scheduler) Register(j *Job) error {
	if s == nil {
		return fmt.Errorf("scheduler cannot be nil")
	}

	if j == nil {
		return fmt.Errorf("job cannot be nil")
	}

	// if the name of the job is already taken, return an error
	for _, job := range s.Jobs {
		if job != nil && job.Name == j.Name {
			return fmt.Errorf("job with name %v already exists", j.Name)
		}
	}

	return s.addJob(j)
}

// Start starts the cron scheduler.
//
// NOTE: Need to specify for how long the scheduler should run after
// calling this function (e.g. time.Sleep(1 * time.Hour) or forever)
func (s *Scheduler) Start() { s.cron.Start() }

// Job represents a cron job that can be registered with the cron scheduler.
type Job struct {
	Source      string       // Source of the job (like the name of application which registered the job)
	Freq        string       // Frequency of the job in cron format
	Name        string       // Name of the job
	Func        func() error // Function to be executed by the job
	Description string       // Optional. Description of the job
	// TODO: add these in the logic and test them
	OnSuccess     func()      // Optional. Function to be executed after the job executes without errors
	OnError       func(error) // Optional. Function to be executed if the job returns an error
	PostExecution func(error) // Optional. Combined version of OnError and OnSuccess functions.
}

type Storage interface {
	// AllJobs returns a list of all registered jobs
	AllJobs() ([]JobInfo, error)
	// RegisterJob registers the details of the selected job
	RegisterJob(source, name, frequency, description string, status JobStatus, err error) error
	// RegisterExecution registers the execution of a job if the storage is specified
	RegisterExecution(*ExecutionLog) error
	// FindExecutions returns a list of job executions that match the filter
	FindExecutions(filter ExecutionFilter) ([]ExecutionLog, error)
	// SetJobsToInactive updates the status of the jobs for the given source. This is useful when the app exits.
	SetJobsToInactive(source string) error
}

// JobInfo stores information about the registered job
type JobInfo struct {
	CreatedAt       time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" bson:"updated_at"`
	Source          string    `json:"source" bson:"source"`
	Name            string    `json:"name" bson:"name"`
	Status          string    `json:"status" bson:"status"`
	Frequency       string    `json:"frequency" bson:"frequency"`
	Description     string    `json:"description" bson:"description"`
	Error           string    `json:"error" bson:"error"`
	ExitedWithError bool      `json:"exited_with_error" bson:"exited_with_error"`
}

// ExecutionLog stores information about the job execution
type ExecutionLog struct {
	Source        string        `json:"source" bson:"source"`
	Name          string        `json:"name" bson:"name"`
	InitializedAt time.Time     `json:"initialized_at" bson:"initialized_at"`
	FinishedAt    time.Time     `json:"finished_at" bson:"finished_at"`
	ExecutionTime time.Duration `json:"execution_time" bson:"execution_time"`
	Error         string        `json:"error" bson:"error"`
}

type ExecutionFilter struct {
	From         time.Time    `json:"from" bson:"from"`
	To           time.Time    `json:"to" bson:"to"`
	Limit        int64        `json:"limit" bson:"limit"`
	Skip         int64        `json:"skip" bson:"skip"`
	ExecutionLog ExecutionLog `json:"execution_log" bson:"execution_log"`
}

// newExecutionLog creates a new ExecutionLog instance.
func newExecutionLog(source, name string, initializedAt time.Time, err error) *ExecutionLog {
	log := &ExecutionLog{
		Source:        source,
		Name:          name,
		InitializedAt: initializedAt,
		FinishedAt:    time.Now().UTC(),
		ExecutionTime: time.Since(initializedAt),
		// Error:         TernaryOp(err != nil, err.Error(), ""),
	}

	// Avoid panics if the error is nil
	if err != nil {
		log.Error = err.Error()
	}

	// log.Error = TernaryOp(err != nil, err.Error(), "")

	return log
}

// func TernaryOp[T any](condition bool, a, b T) T {
// 	if condition {
// 		return a
// 	}
// 	return b
// }

type JobStatus string

const (
	JobStatusInitialized JobStatus = "initialized"
	JobStatusRunning     JobStatus = "running"
	JobStatusDone        JobStatus = "done"
	JobStatusInactive    JobStatus = "inactive"
)

// addJob adds a new job to the cron scheduler and wraps the job function with a
// mutex lock to prevent the execution of the job if it is already running. If
// the function recieves a valid implementation of the Storage interface then
// this will also handle the registration and monitoring of the job.
func (s *Scheduler) addJob(j *Job) error {
	if j == nil {
		return fmt.Errorf("job cannot be nil")
	}

	if s.cron == nil {
		return fmt.Errorf("cron cannot be nil")
	}

	if j.Freq == "" {
		return fmt.Errorf("frequency has to be specified")
	}

	if j.Name == "" {
		return fmt.Errorf("name has to be specified")
	}

	if j.Func == nil {
		return fmt.Errorf("job function cannot be nil")
	}

	// fmt.Printf("adding job %v with frequency %v\n", j.Name, j.Freq)

	name := j.Name
	freq := j.Freq
	source := s.Source
	descr := j.Description

	// NOTE: there is a slight inefficiency in the data that is written by
	// the query because the (source, name, freq, descr) params are
	// written each time in order to update the status.

	if s.Storage != nil {
		if err := s.Storage.RegisterJob(source, name, freq, descr, JobStatusInitialized, nil); err != nil {
			return err
		}
	}

	// Accumulate errors in the c.AddJob function, because the cron.Job param does not return anything
	errors := errgroup.New()

	_, err := s.cron.AddJob(freq, newJobLock(func() {

		if s.Storage != nil {
			if err := s.Storage.RegisterJob(s.Source, name, freq, descr, JobStatusRunning, nil); err != nil {
				errors.Add(fmt.Errorf("failed to set job %v to running: %v", name, err))
			}
		}

		now := time.Now()

		// Passed in job function which should be executed by the cron job
		err := j.Func()

		if j.PostExecution != nil {
			j.PostExecution(err)
		}

		if err != nil && j.OnError != nil {
			j.OnError(err)
		}

		if err == nil && j.OnSuccess != nil {
			j.OnSuccess()
		}

		if s.Storage != nil {
			if err := s.Storage.RegisterExecution(newExecutionLog(source, name, now, err)); err != nil {
				errors.Add(fmt.Errorf("failed to register execution for %v: %v", name, err))
			}

			if err := s.Storage.RegisterJob(s.Source, name, freq, descr, JobStatusDone, err); err != nil {
				errors.Add(fmt.Errorf("failed to set job %v to done: %v", name, err))
			}
		}

	}, name))

	if err != nil {
		return err
	}

	// Add the job to the list of registered jobs
	s.Jobs = append(s.Jobs, j)

	return errors.ToErr()
}

// jobLock is a mutex lock that prevents the execution of a
// job if it is already running.
type jobLock struct {
	jobFunc  func()
	jobName  string
	jobMutex sync.Mutex
}

func newJobLock(jobFunc func(), name string) *jobLock {
	return &jobLock{jobName: name, jobFunc: jobFunc}
}

func (j *jobLock) Run() {
	if j.tryLock() {
		defer j.jobMutex.Unlock()
		j.jobFunc()
	} else {
		fmt.Printf("job %v already running. Skipping...\n", j.jobName)
	}
}

func (j *jobLock) tryLock() bool { return j.jobMutex.TryLock() }

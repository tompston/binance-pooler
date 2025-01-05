package syro

import (
	"binance-pooler/pkg/lib/mongodb"
	"binance-pooler/pkg/syro/errgroup"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CronScheduler is a wrapper around the robfig/cron package that allows for the
// registration of jobs and the optional storage of job status and
// execution logs.
type CronScheduler struct {
	cron        *cron.Cron  // cron is the cron CronScheduler which will run the jobs
	Source      string      // Source is used to identify the source of the job
	Jobs        []*Job      // Jobs is a list of all registered jobs
	CronStorage CronStorage // Storage is an optional storage interface for the CronScheduler
}

type CronStorage interface {
	// SetOptions sets the storage options
	SetOptions(CronStorageOptions) CronStorage
	// GetStorageOptions returns the storage options
	GetStorageOptions() CronStorageOptions
	// FindCronJobs returns a list of all registered jobs
	FindCronJobs() ([]CronJob, error)
	// RegisterJob registers the details of the selected job
	RegisterJob(source, name, frequency, description string, status JobStatus, err error) error
	// RegisterExecution registers the execution of a job if the storage is specified
	RegisterExecution(*CronExecLog) error
	// FindExecutions returns a list of job executions that match the filter
	FindExecutions(filter CronExecFilter) ([]CronExecLog, error)
	// SetJobsToInactive updates the status of the jobs for the given source. Useful when the app exits.
	SetJobsToInactive(source string) error
}

type CronStorageOptions struct {
	LogRuntime bool
}

func NewCronScheduler(cron *cron.Cron, source string) *CronScheduler {
	return &CronScheduler{cron: cron, Source: source}
}

// WithStorage sets the storage for the CronScheduler.
func (s *CronScheduler) WithStorage(storage CronStorage) *CronScheduler {
	s.CronStorage = storage
	return s
}

// Register adds a new job to the cron CronScheduler and wraps the job function with a
// mutex lock to prevent the execution of the job if it is already running.
// If a storage interface is provided, the job and job execution logs
// will be stored using it
func (s *CronScheduler) Register(j *Job) error {
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

	// if the name of the job is already taken, return an error
	for _, job := range s.Jobs {
		if job != nil && job.Name == j.Name {
			return fmt.Errorf("job with name %v already exists", j.Name)
		}
	}

	// fmt.Printf("adding job %v with frequency %v\n", j.Name, j.Freq)

	name := j.Name
	freq := j.Freq
	source := s.Source
	descr := j.Description

	// NOTE: there is a slight inefficiency in the data that is written by
	// the query because the (source, name, freq, descr) params are
	// written each time in order to update the status.

	if s.CronStorage != nil {
		if err := s.CronStorage.RegisterJob(source, name, freq, descr, JobStatusInitialized, nil); err != nil {
			return err
		}
	}

	// Accumulate errors in the c.AddJob function, because the cron.Job param does not return anything
	errors := errgroup.New()

	_, err := s.cron.AddJob(freq, newJobLock(func() {

		if s.CronStorage != nil {
			if err := s.CronStorage.RegisterJob(s.Source, name, freq, descr, JobStatusRunning, nil); err != nil {
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

		if s.CronStorage != nil {
			if err := s.CronStorage.RegisterExecution(newCronExecutionLog(source, name, now, err)); err != nil {
				errors.Add(fmt.Errorf("failed to register execution for %v: %v", name, err))
			}

			if err := s.CronStorage.RegisterJob(s.Source, name, freq, descr, JobStatusDone, err); err != nil {
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

// Start starts the cron CronScheduler.
//
// NOTE: Need to specify for how long the CronScheduler should run after
// calling this function (e.g. time.Sleep(1 * time.Hour) or forever)
func (s *CronScheduler) Start() { s.cron.Start() }

// Job represents a cron job that can be registered with the CronScheduler.
type Job struct {
	Source      string       // Source of the job (like the name of application which registered the job)
	Freq        string       // Frequency of the job in cron format
	Name        string       // Name of the job
	Func        func() error // Function to be executed by the job
	Description string       // Optional. Description of the job
	// TODO: add these in the logic and test them
	// TODO: add a context input so that it would be possible to optionally cancel the job if it takes longer than x to run
	OnSuccess     func()      // Optional. Function to be executed after the job executes without errors
	OnError       func(error) // Optional. Function to be executed if the job returns an error
	PostExecution func(error) // Optional. Combined version of OnError and OnSuccess functions.
}

// CronJob stores information about the registered job
type CronJob struct {
	ID              string     `json:"_id" bson:"_id"`
	CreatedAt       time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" bson:"updated_at"`
	Source          string     `json:"source" bson:"source"`
	Name            string     `json:"name" bson:"name"`
	Status          string     `json:"status" bson:"status"`
	Frequency       string     `json:"frequency" bson:"frequency"`
	Description     string     `json:"description" bson:"description"`
	Error           string     `json:"error" bson:"error"`
	ExitedWithError bool       `json:"exited_with_error" bson:"exited_with_error"`
	FinishedAt      *time.Time `json:"finished_at" bson:"finished_at"`
}

// CronExecLog stores information about the job execution
type CronExecLog struct {
	Source        string        `json:"source" bson:"source"`
	Name          string        `json:"name" bson:"name"`
	InitializedAt time.Time     `json:"initialized_at" bson:"initialized_at"`
	FinishedAt    time.Time     `json:"finished_at" bson:"finished_at"`
	ExecutionTime time.Duration `json:"execution_time" bson:"execution_time"`
	Error         string        `json:"error" bson:"error"`
}

type CronExecFilter struct {
	TimeseriesFilter `json:"timeseries_filter" bson:"timeseries_filter"`
	Source           string        `json:"source" bson:"source"`
	Name             string        `json:"name" bson:"name"`
	ExecutionTime    time.Duration `json:"execution_time" bson:"execution_time"`
}

func newCronExecutionLog(source, name string, initializedAt time.Time, err error) *CronExecLog {
	log := &CronExecLog{
		Source:        source,
		Name:          name,
		InitializedAt: initializedAt,
		FinishedAt:    time.Now().UTC(),
		ExecutionTime: time.Since(initializedAt),
	}

	// Avoid panics if the error is nil
	if err != nil {
		log.Error = err.Error()
	}

	return log
}

type JobStatus string

const (
	JobStatusInitialized JobStatus = "initialized"
	JobStatusRunning     JobStatus = "running"
	JobStatusDone        JobStatus = "done"
	JobStatusInactive    JobStatus = "inactive"
)

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
	if j.jobMutex.TryLock() {
		defer j.jobMutex.Unlock()
		j.jobFunc()
	} else {
		fmt.Printf("job %v already running. Skipping...\n", j.jobName)
	}
}

// --- Mongo implementation ---

// MongoStorage implementation of the Storage interface
type MongoCronStorage struct {
	Options         CronStorageOptions
	cronListColl    *mongo.Collection
	cronHistoryColl *mongo.Collection
}

// NOTE: add optional auto delete index?
func NewMongoCronStorage(cronListColl, cronHistoryColl *mongo.Collection) (*MongoCronStorage, error) {
	if cronListColl == nil || cronHistoryColl == nil {
		return nil, fmt.Errorf("collections cannot be nil")
	}

	// Create indexes for the collections
	if err := mongodb.NewIndexes().
		Add("name").
		Add("status").
		Add("frequency").
		Create(cronListColl); err != nil {
		return nil, err
	}

	// Create indexes for the collections
	if err := mongodb.NewIndexes().
		Add("name").
		Add("initialized_at").
		Add("execution_time").
		Create(cronHistoryColl); err != nil {
		return nil, err
	}

	return &MongoCronStorage{
		cronListColl:    cronListColl,
		cronHistoryColl: cronHistoryColl,
	}, nil
}

func (m *MongoCronStorage) SetOptions(opt CronStorageOptions) CronStorage {
	m.Options = opt
	return m
}

func (m *MongoCronStorage) GetStorageOptions() CronStorageOptions {
	return m.Options
}

// TODO: refactor so that filter is a variadic parameter
func (m *MongoCronStorage) FindCronJobs() ([]CronJob, error) {
	var docs []CronJob
	err := mongodb.GetAllDocumentsWithTypes(m.cronListColl, bson.M{}, nil, &docs)
	return docs, err
}

// TODO: test this function
func (m *MongoCronStorage) SetJobsToInactive(source string) error {
	filter := bson.M{"source": source}
	update := bson.M{"$set": bson.M{"status": JobStatusInactive}}
	_, err := m.cronListColl.UpdateMany(context.Background(), filter, update)
	return err
}

// RegisterJob upsert the job name in the database based on the source
// and the job name. If the job does not exist, set the created_at
// field to the current time. If the job already exists,
// update the updated_at field to the current time.
func (m *MongoCronStorage) RegisterJob(source, name, freq, descr string, status JobStatus, fnErr error) error {
	filter := bson.M{
		"source": source,
		"name":   name,
	}

	set := bson.M{
		"frequency":   freq,
		"status":      status,
		"description": descr,
		"updated_at":  time.Now().UTC(),
	}

	if fnErr != nil {
		set["exited_with_error"] = true
		set["error"] = fnErr.Error()
	} else {
		set["exited_with_error"] = false
		set["error"] = ""
	}

	if status == JobStatusDone {
		set["finished_at"] = time.Now().UTC()
	}

	_, err := m.cronListColl.UpdateOne(context.Background(), filter, bson.M{
		"$set":         set,
		"$setOnInsert": bson.M{"created_at": time.Now().UTC()},
	}, mongodb.UpsertOpt)

	return err
}

// Register the execution of a job in the database
func (m *MongoCronStorage) RegisterExecution(ex *CronExecLog) error {
	if ex == nil {
		return fmt.Errorf("job execution cannot be nil")
	}

	_, err := m.cronHistoryColl.InsertOne(context.Background(), ex)
	return err
}

// FindExecutions returns a list of executions based on the filter
func (m *MongoCronStorage) FindExecutions(filter CronExecFilter) ([]CronExecLog, error) {
	queryFilter := bson.M{}

	// if the from and to fields are not zero, add them to the query filter
	if !filter.From.IsZero() && !filter.To.IsZero() {
		if filter.From.After(filter.To) {
			return nil, errors.New("from date cannot be after to date")
		}

		queryFilter["time"] = bson.M{"$gte": filter.From, "$lte": filter.To}
	}

	if filter.Source != "" {
		queryFilter["source"] = filter.Source
	}

	if filter.Name != "" {
		queryFilter["name"] = filter.Name
	}

	if filter.ExecutionTime > 0 {
		queryFilter["execution_time"] = bson.M{"$gte": filter.ExecutionTime}
	}

	// limit := filter.TimeseriesFilter.Limit
	limit := 200

	opts := options.Find().
		SetSort(bson.D{{Key: "initialized_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(filter.TimeseriesFilter.Skip)

	var docs []CronExecLog
	err := mongodb.GetAllDocumentsWithTypes(m.cronHistoryColl, queryFilter, opts, &docs)
	return docs, err
}

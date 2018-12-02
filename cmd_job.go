package main

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"sort"
	"sync"
	"time"
)

type Job struct {
	Id         string    `json:"id"`
	Status     JobStatus `json:"status"`
	Error      string    `json:"error"` // Error msg when fork & exec
	Cmd        string    `json:"cmd"`
	Dir        string    `json:"dir"`
	Env        []string  `json:"env"`
	Stdout     string    `json:"stdout"`
	Stderr     string    `json:"stderr"`
	ExitCode   int       `json:"exit_code"`
	Pid        int       `json:"pid"`
	CreateTime time.Time `json:"create_time"`
	FinishTime time.Time `json:"finish_time"`
	cancelFunc context.CancelFunc
}

type Jobs []*Job

func (o Jobs) Len() int {
	return len(o)
}
func (o Jobs) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}
func (o Jobs) Less(i, j int) bool {
	return o[i].CreateTime.Before(o[j].CreateTime)
}

type JobBookkeeper struct {
	expireDays int

	jobs map[string]*Job

	quitC  chan struct{}
	cancel context.CancelFunc

	sync.RWMutex
}

func NewJobBookkeeper(expireDays int) *JobBookkeeper {
	if expireDays <= 0 {
		expireDays = 7
	}

	ctx, cancel := context.WithCancel(context.Background())

	o := JobBookkeeper{
		expireDays: expireDays,
		jobs:       make(map[string]*Job),
		quitC:      make(chan struct{}),
		cancel:     cancel,
	}
	go o.checkExpire(ctx)
	return &o
}

func (o *JobBookkeeper) Close() error {
	o.cancel()
	<-o.quitC
	return nil
}

// Record the job's info
func (o *JobBookkeeper) Add(j *Job) {
	o.Lock()
	defer o.Unlock()
	if j == nil {
		return
	}
	o.jobs[j.Id] = j
}

// Get the job info by id
func (o *JobBookkeeper) Get(id string) *Job {
	o.Lock()
	defer o.Unlock()
	if j, ok := o.jobs[id]; ok {
		return j
	} else {
		return nil
	}
}

// Get all jobs ordered by create time desc
func (o *JobBookkeeper) GetAll() []*Job {
	o.Lock()
	defer o.Unlock()
	// The job that
	var jobs Jobs
	for _, j := range o.jobs {
		jobs = append(jobs, j)
	}
	sort.Sort(sort.Reverse(jobs))
	return jobs
}

func (o *JobBookkeeper) checkExpire(ctx context.Context) {
	defer close(o.quitC)
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			o.expire()
		case <-ctx.Done():
			return
		}
	}
}

func (o *JobBookkeeper) expire() {
	o.Lock()
	defer o.Unlock()
	purgedCnt := 0
	for k, j := range o.jobs {
		if j.Status == JSRunning {
			continue
		}
		if time.Duration(o.expireDays)*time.Hour*24 < time.Now().Sub(j.FinishTime) {
			delete(o.jobs, k)
			purgedCnt++
		}
	}
	log.Debugf("purged %d outdated jobs", purgedCnt)
}

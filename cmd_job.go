package main

import (
	"context"
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
	Jobs map[string]*Job
	Lock sync.RWMutex
}

func NewJobBookkeeper() *JobBookkeeper {
	return &JobBookkeeper{
		Jobs: map[string]*Job{},
		Lock: sync.RWMutex{},
	}
}

// Record the job's info
func (o *JobBookkeeper) Add(j *Job) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	if j == nil {
		return
	}
	o.Jobs[j.Id] = j
}

// Get the job info by id
func (o *JobBookkeeper) Get(id string) *Job {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	if j, ok := o.Jobs[id]; ok {
		return j
	} else {
		return nil
	}
}

// Get all jobs ordered by create time desc
func (o *JobBookkeeper) GetAll() []*Job {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	// The job that
	var jobs Jobs
	for _, j := range o.Jobs {
		jobs = append(jobs, j)
	}
	sort.Sort(sort.Reverse(jobs))
	return jobs
}

func (o *JobBookkeeper) Expire() {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	for k, v := range o.Jobs {
		if time.Hour*24*3 < time.Now().Sub(v.FinishTime) {
			delete(o.Jobs, k)
		}
	}
}

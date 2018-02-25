package main

import (
	"bytes"
	"context"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/nu7hatch/gouuid"
	"io/ioutil"
	"net/http"
	"os/exec"
	"sync"
	"syscall"
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

type RunCmdReq struct {
	Cmd   string   `json:"cmd"`
	Async bool     `json:"async,omitempty"`
	Dir   string   `json:"dir,omitempty"`
	Env   []string `json:"env,omitempty"`
}

type SyncRunCmdRes Job
type AsyncRuncmdRes struct {
	Id         string    `json:"id"`
	CreateTime time.Time `json:"create_time"`
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

func (o *JobBookkeeper) Add(j *Job) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	if j == nil {
		return
	}
	o.Jobs[j.Id] = j
}

var (
	gJobBookkeeper *JobBookkeeper
)

func init() {
	gHttpServer.AddToInit(InitUserHandler)
	gHttpServer.AddToUninit(UninitUserHandler)
}

func InitUserHandler() error {
	gJobBookkeeper = NewJobBookkeeper()
	return nil
}

func UninitUserHandler() {
}

func RunCmdHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var req RunCmdReq
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Failed to read r.Body: %s", err)
		ServeJSON(w, NewResponse().SetError(ECUnknown, "Failed to read r.Body"))
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &req); err != nil {
		log.Errorf("Failed to unmarshall data: %s", err)
		ServeJSON(w, NewResponse().SetError(ECUnknown, "Failed to unmarshall data"))
		return
	}

	if req.Cmd == "" {
		ServeJSON(w, NewResponse().SetError(ECInvalidParam, "Param cmd is empty"))
		return
	}

	var job Job
	job.Cmd = req.Cmd
	job.Dir = req.Dir
	job.Env = req.Env
	job.Status = JSRunning
	job.CreateTime = time.Now()
	job.FinishTime = time.Unix(0, 0)

	u4, err := uuid.NewV4()
	if err != nil {
		log.Errorf("Failed to genereate uuid: %s", err)
		ServeJSON(w, NewResponse().SetError(ECUnknown, "Failed to generate uuid"))
		return
	}
	job.Id = u4.String()

	ctx, cancel := context.WithCancel(context.Background())
	job.cancelFunc = cancel

	gJobBookkeeper.Add(&job)

	var resp interface{}
	if !req.Async {
		cmdWorker(ctx, &job)
		resp = (*SyncRunCmdRes)(&job)
	} else {
		go cmdWorker(ctx, &job)
		resp = &AsyncRuncmdRes{
			Id:         job.Id,
			CreateTime: job.CreateTime,
		}
	}
	ServeJSON(w, NewResponse().SetData(resp))

}

func cmdWorker(ctx context.Context, job *Job) {
	var err error
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	defer func() {
		job.FinishTime = time.Now()
		job.Stdout = stdout.String()
		job.Stderr = stderr.String()
	}()

	cmd := exec.Command("sh", "-c", job.Cmd)
	cmd.Dir = job.Dir
	cmd.Env = append(cmd.Env, job.Env...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Infof("Running cmd: %s", job.Cmd)
	err = cmd.Start()
	if err != nil {
		log.Errorf("cmd.Start failed: %s", err)
		job.Error = err.Error()
		job.Status = JSFailed
		return
	}

	job.Pid = cmd.Process.Pid

	doneC := make(chan struct{})
	canceled := false
	// Wait for context cancel
	go func() {
		select {
		case <-ctx.Done():
			canceled = true
			cmd.Process.Kill()
			log.Info("Canceling the process: ", cmd.Process.Pid)
		case <-doneC:
		}
	}()

	// Wait until the process exits or be killed
	err = cmd.Wait()
	close(doneC)
	if err != nil {
		// The process has been killed, exit with non-zero, or termiated by some signal
		log.Error("c.Process.Wait failed: ", err)

		if ee, ok := err.(*exec.ExitError); ok && ee.Exited() {
			exitCode := ee.Sys().(syscall.WaitStatus).ExitStatus()
			log.Error("process exited with non-zero exit code: ", exitCode)
			job.ExitCode = exitCode
		}

		job.Error = err.Error()
		job.Status = JSFailed

	} else {
		log.Info("Process finished: ", cmd.Process.Pid)
		job.Status = JSFinished
	}

	// If has been canceled by user
	if canceled {
		log.Warn("Process canceled: ", cmd.Process.Pid)
		job.Error = err.Error()
		job.Status = JSCanceled
	}

}

/*
func (o *InstanceController) CancelConsistencyCheck() {
	var err error
	var job models.ConsistencyCheckJob
	db := orm.NewOrm()
	id, err := strconv.Atoi(o.Ctx.Input.Param(":instance_id"))
	if id <= 0 || err != nil {
		o.Data["json"] = NewResponse().SetError(ECInvalidParam, "invalid instance")
		o.ServeJSON()
		return
	}
	if cancelFunc, ok := globalConsistencyCheckJobContext[job.Id]; ok {
		cancelFunc()
		o.Data["json"] = NewResponse()
		o.ServeJSON()
		return
	} else {
		o.Data["json"] = NewResponse().SetError(ECUnknown, "the job is lost")
		o.ServeJSON()
		return
	}
}
*/

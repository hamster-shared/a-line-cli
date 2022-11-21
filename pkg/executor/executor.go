package executor

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	action2 "github.com/hamster-shared/a-line-cli/pkg/action"
	"github.com/hamster-shared/a-line-cli/pkg/logger"
	"github.com/hamster-shared/a-line-cli/pkg/model"
	"github.com/hamster-shared/a-line-cli/pkg/output"
	"github.com/hamster-shared/a-line-cli/pkg/service"
	"github.com/hamster-shared/a-line-cli/pkg/utils"
	"gopkg.in/yaml.v2"
)

type IExecutor interface {

	// FetchJob 获取任务
	FetchJob(name string) (io.Reader, error)

	// Execute 执行任务
	Execute(id int, job *model.Job) error

	// HandlerLog 处理日志
	HandlerLog(jobId int)

	//SendResultToQueue 发送结果到队列
	SendResultToQueue(channel chan any)

	Cancel(name string)
}

type Executor struct {
	cancelMap  map[string]func()
	jobService service.IJobService
}

// FetchJob 获取任务
func (e *Executor) FetchJob(name string) (io.Reader, error) {

	//TODO... 根据 name 从 rpc 或 直接内部调用获取 job 的 pipeline 文件
	job := e.jobService.GetJob(name)
	data, err := yaml.Marshal(job)
	return strings.NewReader(string(data)), err
}

// Execute 执行任务
func (e *Executor) Execute(id int, job *model.Job) error {

	// 1. 解析对 pipeline 进行任务排序
	stages, err := job.StageSort()
	jobWrapper := &model.JobDetail{
		Id:     id,
		Job:    *job,
		Status: model.STATUS_NOTRUN,
		Stages: stages,
	}

	if err != nil {
		return err
	}

	// 2. 初始化 执行器的上下文

	env := make([]string, 0)
	env = append(env, "NAME="+job.Name)

	engineContext := make(map[string]interface{})
	engineContext["hamsterRoot"] = "/tmp/example"
	engineContext["workdir"] = "/tmp/example"
	engineContext["name"] = job.Name
	engineContext["env"] = env

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), "stack", engineContext))

	// 将取消 hook 记录到内存中，用于中断程序
	e.cancelMap[job.Name] = cancel

	// 队列堆栈
	var stack utils.Stack[action2.ActionHandler]

	jobWrapper.Status = model.STATUS_RUNNING
	jobWrapper.StartTime = time.Now()

	executeAction := func(ah action2.ActionHandler, job *model.JobDetail) error {
		if jobWrapper.Status != model.STATUS_RUNNING {
			return nil
		}
		err := ah.Pre()
		if err != nil {
			job.Status = model.STATUS_FAIL
			fmt.Println(err)
			return err
		}
		stack.Push(ah)
		_, err = ah.Hook()
		if err != nil {
			job.Status = model.STATUS_FAIL
			return err
		}
		return nil
	}

	job.Output = output.New(job.Name, jobWrapper.Id)

	for index, stageWapper := range jobWrapper.Stages {
		//TODO ... stage 的输出也需要换成堆栈方式
		logger.Info("stage: {")
		logger.Infof("   // %s", stageWapper.Name)
		stageWapper.Status = model.STATUS_RUNNING
		stageWapper.StartTime = time.Now()
		jobWrapper.Stages[index] = stageWapper
		e.jobService.SaveJobDetail(jobWrapper.Name, jobWrapper)
		for _, step := range stageWapper.Stage.Steps {
			var ah action2.ActionHandler
			if step.RunsOn != "" {
				ah = action2.NewDockerEnv(step, ctx, job.Output)
				err = executeAction(ah, jobWrapper)
			}
			if step.Uses == "" {
				ah = action2.NewShellAction(step, ctx, job.Output)
				err = executeAction(ah, jobWrapper)
			}
			if step.Uses == "git-checkout" {
				ah = action2.NewGitAction(step, ctx, job.Output)
				err = executeAction(ah, jobWrapper)
			}
			if step.Uses == "ipfs" {
				ah = action2.NewIpfsAction(step, ctx, job.Output)
				err = executeAction(ah, jobWrapper)
			}
			if strings.Contains(step.Uses, "/") {
				ah = action2.NewRemoteAction(step, ctx)
				err = executeAction(ah, jobWrapper)
			}

		}
		for !stack.IsEmpty() {
			ah, _ := stack.Pop()
			_ = ah.Post()
		}

		if err != nil {
			stageWapper.Status = model.STATUS_FAIL
		} else {
			stageWapper.Status = model.STATUS_SUCCESS
		}
		stageWapper.Duration = time.Now().Sub(stageWapper.StartTime)
		jobWrapper.Stages[index] = stageWapper
		e.jobService.SaveJobDetail(jobWrapper.Name, jobWrapper)
		logger.Info("}")
		if err != nil {
			cancel()
			break
		}

	}
	job.Output.Done()

	delete(e.cancelMap, job.Name)
	if err == nil {
		jobWrapper.Status = model.STATUS_SUCCESS
	} else {
		jobWrapper.Status = model.STATUS_FAIL
	}

	jobWrapper.Duration = time.Now().Sub(jobWrapper.StartTime)
	e.jobService.SaveJobDetail(jobWrapper.Name, jobWrapper)

	//TODO ... 发送结果到队列
	e.SendResultToQueue(nil)
	_ = os.RemoveAll(path.Join(engineContext["hamsterRoot"].(string), job.Name))

	return err

}

// HandlerLog 处理日志
func (e *Executor) HandlerLog(jobId int) {

	//TODO ...
}

// SendResultToQueue 发送结果到队列
func (e *Executor) SendResultToQueue(channel chan any) {
	//TODO ...
}

// Cancel 取消
func (e *Executor) Cancel(name string) {

	cancel, ok := e.cancelMap[name]
	if ok {
		cancel()
	}
}

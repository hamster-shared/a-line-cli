package model

import (
	"fmt"
	"github.com/hamster-shared/a-line-cli/pkg/utils"
	"io"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/hamster-shared/a-line-cli/pkg/output"
)

type Status int

const (
	STATUS_NOTRUN  Status = 0
	STATUS_RUNNING Status = 1
	STATUS_FAIL    Status = 2
	STATUS_SUCCESS Status = 3
	STATUS_STOP    Status = 4
)

type Job struct {
	Version string           `yaml:"version" json:"version"`
	Name    string           `yaml:"name" json:"name"`
	Stages  map[string]Stage `yaml:"stages" json:"stages"`
}

type JobVo struct {
	Version          string           `yaml:"version" json:"version"`
	Name             string           `yaml:"name" json:"name"`
	Stages           map[string]Stage `yaml:"stages" json:"stages"`
	Status           Status           `json:"status"`
	StartTime        time.Time        `yaml:"startTime"json:"startTime"`
	Duration         int64            `json:"duration"`
	TriggerMode      string           `yaml:"triggerMode"json:"triggerMode"`
	PipelineDetailId int              `json:"pipelineDetailId"`
}

type JobDetail struct {
	Id int `json:"id"`
	Job
	Status       Status        `json:"status"`
	TriggerMode  string        `yaml:"triggerMode"json:"triggerMode"`
	Stages       []StageDetail `json:"stages"`
	StartTime    time.Time     `yaml:"startTime"json:"startTime"`
	Duration     int64         `json:"duration"`
	ActionResult `yaml:"actionResult" json:"actionResult"`
	Output       *output.Output `json:"output"`
}

func (jd *JobDetail) ToString() string {
	str := ""
	for _, s := range jd.Stages {
		str += s.ToString() + "\n"
	}
	return fmt.Sprintf("job: %s, Status: %d, StartTime: %s , Duration: %d, stages: [\n%s]", jd.Name, jd.Status, jd.StartTime, jd.Duration, str)
}

// StageSort job 排序
func (job *Job) StageSort() ([]StageDetail, error) {
	stages := make(map[string]Stage)
	for key, stage := range job.Stages {
		stages[key] = stage
	}

	sortedMap := make(map[string]any)

	stageList := make([]StageDetail, 0)
	for len(stages) > 0 {
		last := len(stages)
		for key, stage := range stages {
			allContains := true
			for _, needs := range stage.Needs {
				_, ok := sortedMap[needs]
				if !ok {
					allContains = false
				}
			}
			if allContains {
				sortedMap[key] = ""
				delete(stages, key)
				stageList = append(stageList, NewStageDetail(key, stage))
			}
		}

		if len(stages) == last {
			return nil, fmt.Errorf("cannot resolve dependency, %v", stages)
		}

	}

	return stageList, nil
}

func (jd *JobDetail) AddArtifactory(file *os.File) error {
	arti := Artifactory{
		Name: file.Name(),
		Url:  fmt.Sprintf("/artifactory/%s/%d/%s", jd.Name, jd.Id, file.Name()),
	}
	dir := path.Join(utils.DefaultConfigDir(), "artifactory", jd.Name, strconv.Itoa(jd.Id))
	_ = os.MkdirAll(dir, os.ModePerm)

	fullPath := path.Join(dir, file.Name())

	destination, err := os.Create(fullPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(destination, file)
	if err != nil {
		return err
	}

	if len(jd.Artifactorys) > 0 {
		jd.Artifactorys = append(jd.Artifactorys, arti)
	} else {
		jd.Artifactorys = make([]Artifactory, 0)
		jd.Artifactorys = append(jd.Artifactorys, arti)
	}
	return nil
}

type JobLog struct {
	// 开始时间
	StartTime time.Time `json:"startTime"`
	// 持续时间
	Duration time.Duration `json:"duration"`

	//日志内容
	Content string `json:"content"`

	//最后一行 行号
	LastLine int `json:"lastLine"`
}

type JobStageLog struct {
	// 开始时间
	StartTime time.Time `json:"startTime"`
	// 持续时间
	Duration time.Duration `json:"duration"`

	//日志内容
	Content string `json:"content"`

	//最后一行 行号
	LastLine int `json:"lastLine"`
}

type JobPage struct {
	Data     []JobVo `json:"data"`
	Total    int     `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"pageSize"`
}

type JobDetailPage struct {
	Data     []JobDetail `json:"data"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

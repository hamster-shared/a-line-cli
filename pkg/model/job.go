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
	Version string           `yaml:"version"`
	Name    string           `yaml:"name"`
	Stages  map[string]Stage `yaml:"stages"`
}

type JobDetail struct {
	Id int
	Job
	Status      Status
	TriggerMode string //触发方式
	Stages      []StageDetail
	StartTime   time.Time
	Duration    time.Duration
	ActionResult
	Output *output.Output
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
	StartTime time.Time
	// 持续时间
	Duration time.Duration

	//日志内容
	Content string

	//最后一行 行号
	LastLine int
}

type JobStageLog struct {
	// 开始时间
	StartTime time.Time
	// 持续时间
	Duration time.Duration

	//日志内容
	Content string

	//最后一行 行号
	LastLine int
}

type JobPage struct {
	Data     []Job
	Total    int
	Page     int
	PageSize int
}

type JobDetailPage struct {
	Data     []JobDetail
	Total    int
	Page     int
	PageSize int
}

package action

import (
	"context"
	"errors"
	"fmt"
	"github.com/hamster-shared/a-line-cli/pkg/consts"
	"github.com/hamster-shared/a-line-cli/pkg/logger"
	"github.com/hamster-shared/a-line-cli/pkg/model"
	"github.com/hamster-shared/a-line-cli/pkg/output"
	"github.com/hamster-shared/a-line-cli/pkg/utils"
	"os"
	path2 "path"
	"path/filepath"
	"strings"
)

// ArtifactoryAction Storage building
type ArtifactoryAction struct {
	name   string
	path   []string
	output *output.Output
	ctx    context.Context
}

func NewArtifactoryAction(step model.Step, ctx context.Context, output *output.Output) *ArtifactoryAction {
	var path string
	s := step.With["path"]
	if s != "" {
		if s[len(s)-1:] == "\n" {
			path = s[:len(s)-1]
		} else {
			path = s
		}
	}
	return &ArtifactoryAction{
		name:   step.With["name"],
		path:   strings.Split(path, "\n"),
		ctx:    ctx,
		output: output,
	}
}

func (a *ArtifactoryAction) Pre() error {
	if !(len(a.path) > 0 && a.path[0] != "") {
		return errors.New("the path parameter of the save artifact is required")
	}
	if a.name == "" {
		return errors.New("the name parameter of the save artifact is required")
	}
	split := strings.Split(a.name, ".")
	if split[len(split)-1] != "zip" {
		return errors.New("can only be saved as zip")
	}
	stack := a.ctx.Value(STACK).(map[string]interface{})
	workdir, ok := stack["workdir"].(string)
	if !ok {
		return errors.New("get workdir error")
	}
	var fullPathList []string
	for _, path := range a.path {
		absPath := path2.Join(workdir, path)
		fullPathList = append(fullPathList, absPath)
	}
	var absPathList []string
	a.path = GetFiles(workdir, fullPathList, absPathList)
	return nil
}

func (a *ArtifactoryAction) Hook() (*model.ActionResult, error) {
	stack := a.ctx.Value(STACK).(map[string]interface{})
	jobName, ok := stack["name"].(string)
	if !ok {
		return nil, errors.New("get job name error")
	}
	jobId, ok := stack["id"].(string)
	if !ok {
		return nil, errors.New("get job id error")
	}
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Errorf("Failed to get home directory, the file will be saved to the current directory, err is %s", err.Error())
		userHomeDir = "."
	}
	dest := path2.Join(userHomeDir, consts.ARTIFACTORY_DIR, jobName, jobId, a.name)
	var files []*os.File
	for _, path := range a.path {
		file, err := os.Open(path)
		if err != nil {
			return nil, errors.New("file open fail")
		}
		files = append(files, file)
	}
	err = utils.CompressZip(files, dest)
	if err != nil {
		return nil, errors.New("compression failed")
	}
	logger.Infof("File saved to %s", dest)
	actionResult := model.ActionResult{
		Artifactorys: []model.Artifactory{
			{
				Name: a.name,
				Url:  dest,
			},
		},
		Reports: []model.Report{
			{
				Id:   1,
				Url:  dest,
				Type: 1,
			},
		},
	}
	return &actionResult, nil
}

func (a *ArtifactoryAction) Post() error {
	fmt.Println("artifactory Post end")
	return nil
}

func GetFiles(workdir string, fuzzyPath []string, pathList []string) []string {
	files, _ := os.ReadDir(workdir)
	flag := false
	for _, file := range files {
		currentPath := workdir + "/" + file.Name()
		for _, path := range fuzzyPath {
			matched, err := filepath.Match(path, currentPath)
			flag = matched
			if matched && err == nil {
				pathList = append(pathList, currentPath)
			}
		}
		if file.IsDir() && !flag {
			pathList = GetFiles(currentPath, fuzzyPath, pathList)
		}
	}
	return pathList
}

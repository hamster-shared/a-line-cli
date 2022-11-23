package action

import (
	"context"
	"errors"
	"fmt"
	"github.com/hamster-shared/a-line-cli/pkg/consts"
	"github.com/hamster-shared/a-line-cli/pkg/model"
	"github.com/hamster-shared/a-line-cli/pkg/output"
	shell "github.com/ipfs/go-ipfs-api"
	"os"
	path2 "path"
)

// IpfsAction Upload files/directories to ipfs
type IpfsAction struct {
	path   string
	output *output.Output
	ctx    context.Context
}

func NewIpfsAction(step model.Step, ctx context.Context, output *output.Output) *IpfsAction {
	return &IpfsAction{
		path:   step.With["path"],
		ctx:    ctx,
		output: output,
	}
}

func (a *IpfsAction) Pre() error {
	newShell := shell.NewShell(consts.IPFS_SHELL)
	version, s, err := newShell.Version()
	if err != nil {
		return errors.New("get workdir error")
	}
	fmt.Println(fmt.Sprintf("ipfs version is %s, commit sha is %s", version, s))
	return nil
}

func (a *IpfsAction) Hook() (*model.ActionResult, error) {
	stack := a.ctx.Value(STACK).(map[string]interface{})

	workdir, ok := stack["workdir"].(string)
	if !ok {
		return nil, errors.New("get workdir error")
	}

	path := path2.Join(workdir, a.path)
	fmt.Println(path)
	fi, err := os.Stat(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("get path fail, err is  %s", err.Error()))
	}
	isDir := fi.IsDir()
	newShell := shell.NewShell(consts.IPFS_SHELL)
	var cid string
	if isDir {
		cid, err = newShell.AddDir(path)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("ipfs add dir fail, err is  %s", err.Error()))
		}
	} else {
		file, err := os.Open(path)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("open file fail, err is  %s", err.Error()))
		}
		cid, err = newShell.Add(file)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("ipfs add file fail, err is  %s", err.Error()))
		}
	}

	url := consts.IPFS_WEB_URL + cid
	actionResult := model.ActionResult{
		Artifactorys: []model.Artifactory{
			{
				Name: a.path,
				Url:  url,
			},
		},
		Reports: []model.Report{
			{
				Id:   1,
				Url:  url,
				Type: 1,
			},
		},
	}
	fmt.Println(actionResult)
	return &actionResult, nil
}

func (a *IpfsAction) Post() error {
	stack := a.ctx.Value(STACK).(map[string]interface{})
	workdir, ok := stack["workdir"].(string)
	if !ok {
		return errors.New("get workdir error")
	}
	return os.RemoveAll(workdir)
}

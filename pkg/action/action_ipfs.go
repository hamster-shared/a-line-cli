package action

import (
	"context"
	"fmt"
	"github.com/hamster-shared/a-line-cli/pkg/model"
	"github.com/hamster-shared/a-line-cli/pkg/output"
	"os"
)

// 将目录上传到ipfs
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
	a.output.NewStage("ipfs")

	//TODO ... 检验 ipfs 命令是否存在
	fmt.Println("ipfs hello world pre")
	return nil
}

func (a *IpfsAction) Hook() (*model.ActionResult, error) {
	a.output.NewStage("ipfs")

	//TODO ... 上传目录到ipfs

	fmt.Println("ipfs hello world hook")
	return nil, nil
}

func (a *IpfsAction) Post() error {
	return os.RemoveAll(a.path)
}

package command

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

//Command 执行命令行任务
func SimpleExec(name string, timeout int, arg ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arg...)
	var errbuf bytes.Buffer
	var outbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	if err := cmd.Start(); err != nil {
		return errbuf.String(), err
	}
	if err := cmd.Wait(); err != nil {
		return errbuf.String(), err
	}
	return outbuf.String(), nil
}

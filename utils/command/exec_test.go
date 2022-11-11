package command

import (
	"fmt"
	"testing"
)

func TestCommand(t *testing.T) {
	fmt.Println(SimpleExec("rm", 2, "/root/docker", "-rf"))
}

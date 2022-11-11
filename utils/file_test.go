package utils

import "testing"

func TestFile(t *testing.T) {
	b, err := FileExists("/root")
	t.Log(b, err)
	files, err := ListDir("/root/tan", "")
	t.Log(files, err)
}

func TestFileStat(t *testing.T) {
	t.Log(GetFileStat("/root/go_workspace/main.sh"))
	t.Log(GetMutiFileStat("/root"))
}

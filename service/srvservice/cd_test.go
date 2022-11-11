package srvservice

import (
	"fmt"
	"strings"
	"testing"
)

var ms = NewStack("/root/tan", 0, 20)

func TestCd(t *testing.T) {

	fmt.Println(ms.ParseCdCmd("   cd  docker   "))
	fmt.Println(ms.ParseCdCmd("   cd  ..   "))
}

func TestSplit(t *testing.T) {
	fmt.Println(strings.Split("root/asd", "/"))
	fmt.Println(strings.Split("root/asd/", "/"))
	fmt.Println(strings.Split("/root/asd/", "/"))
}

func TestLs(t *testing.T) {
	t.Log(ms.ParseLsCmd("   ls  test.html  -l "))
	//ms.ParseCdCmd("   cd  docker   ")
	//t.Log(ms.ParseLsCmd("   ls  ..   "))
	// t.Log(ms.ParseLsCmd("   ls  ../1.txt  -l  "))
}

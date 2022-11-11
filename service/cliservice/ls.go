package cliservice

import (
	"encoding/json"
	"fmt"
	"myftp/model"
	"myftp/zlog"
	"net"
)

func LsCommand(conn net.Conn, cmd string) {
	b, err := Pack(0, cmd)
	if err != nil {
		zlog.Error("Pack failed")
		return
	}
	conn.Write(b)
	msg, err := Unpack(conn)
	if err != nil {
		zlog.Error("Unpack failed")
		return
	}
	var resp model.LsResp
	json.Unmarshal(msg.GetData(), &resp)
	for _, res := range resp.Data {
		fmt.Println(res)
	}
}

func CdCommand(conn net.Conn, cmd string) {
	b, err := Pack(6, cmd)
	if err != nil {
		zlog.Error("Pack failed")
		return
	}
	conn.Write(b)
	msg, err := Unpack(conn)
	if err != nil {
		zlog.Error("Unpack failed")
		return
	}
	fmt.Println(string(msg.GetData()))
}

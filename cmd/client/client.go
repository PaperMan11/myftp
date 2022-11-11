package main

import (
	"bufio"
	"fmt"
	"myftp/model"
	"myftp/service/cliservice"
	"myftp/utils"
	"myftp/zlog"
	"net"
	"os"
	"strings"
	"sync"
)

var globalWait sync.WaitGroup

func main() {

	fmt.Println("Client_11 Test Start...")
	// time.Sleep(3 * time.Second)
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("net.Dial err: ", err)
		return
	}
	defer conn.Close()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		scanner.Scan()
		if scanner.Text() == "exit" {
			break
		}
		cmd := scanner.Text()
		switch {
		case strings.Contains(cmd, "put"):
			files := strings.Fields(cmd)
			files = files[1:]
			UploadFiles(conn, files)
		case strings.Contains(cmd, "get"):
			files1 := strings.Fields(cmd)
			files1 = files1[1:]
			DownloadFiles(conn, files1)
		case strings.Contains(cmd, "ls"):
			cliservice.LsCommand(conn, cmd)
		case strings.Contains(cmd, "cd"):
			cliservice.CdCommand(conn, cmd)
		default:
			continue
		}
	}
}

func UploadFiles(conn net.Conn, filePaths []string) {
	for _, file := range filePaths {
		globalWait.Add(1)
		go uploadfile(conn, file)
	}
	globalWait.Wait()
}

func uploadfile(conn net.Conn, filepath string) {
	defer globalWait.Done()

	var err error
	fileSize, _ := utils.FileSize(filepath)
	if fileSize <= model.SmallFileSize {
		err = cliservice.UploaderSmallFile(conn, filepath)
	} else {
		uploader := cliservice.GetUploader(conn, filepath, model.SliceBytes)

		if uploader == nil {
			zlog.Info("这是一个全新要上传的文件")
			uploader = cliservice.NewUploader(conn, filepath, model.SliceBytes)
		}

		if uploader == nil {
			zlog.Error("创建上传器失败，上传文件失败")
			return
		}

		if uploader.StopUp {
			zlog.Errorf("服务端文件与客户端文件uuid不对应")
			return
		}

		err = uploader.UploadFileBySlice()
	}
	if err != nil {
		zlog.Errorf("上传%s文件失败 %s", filepath, err)
	}
}

func DownloadFiles(conn net.Conn, fileNames []string) {
	storeDir := model.StoreDir
	for _, file := range fileNames {
		globalWait.Add(1)
		go downloadfile(conn, file, storeDir)
	}
	globalWait.Wait()
}

func downloadfile(conn net.Conn, fileName, storeDir string) {
	defer globalWait.Done()

	downloader := cliservice.NewDownloader(conn, fileName, storeDir)
	if downloader == nil {
		zlog.Errorf("%s download failed", fileName)
		fmt.Printf("%s 下载失败", fileName)
		return
	}
	err := downloader.DownloadFile(conn, fileName, storeDir)
	if err != nil {
		zlog.Errorf("%s 文件下载失败", fileName)
		return
	}
	zlog.Infof("%s 文件下载成功", fileName)
}

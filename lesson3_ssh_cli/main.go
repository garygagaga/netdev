package main

import (
	"bytes"
	"fmt"
	"log"
	"sync"

	"golang.org/x/crypto/ssh"
)

func pushOne(addr, user, passwd, cmd string) string {
	// 建立SSH客户端连接
	client, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(passwd)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		log.Fatalf("SSH dial error: %s", err.Error())
	}

	// 建立新会话
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("new session error: %s", err.Error())
	}
	defer session.Close()

	var out bytes.Buffer
	session.Stdout = &out

	session.Run(cmd)
	return out.String()
}

func main() {
	// 创建一个设备的结构体
	type device struct {
		IP     string
		User   string
		Passwd string
	}

	// 创建一个设备的切片
	devices := []device{
		{"192.168.5.201", "nett", "Huawei@123"},
		{"192.168.5.202", "nett", "Huawei@123"},
		{"192.168.5.203", "nett", "Huawei@123"}}

	// 循环遍历设备切片，执行命令
	// for _, dev := range devices {
	// 	fmt.Printf("设备IP: %s, 回显为：\n", dev.IP)
	// 	fmt.Println(pushOne(dev.IP+":22", dev.User, dev.Passwd, "display clock"))
	// }

	// 在goroutine中执行命令
	wg := sync.WaitGroup{}
	wg.Add(len(devices))
	for _, dev := range devices {
		go func(dev device) {
			defer wg.Done()
			fmt.Printf("设备IP: %s, 回显为：\n", dev.IP)
			fmt.Println(pushOne(dev.IP+":22", dev.User, dev.Passwd, "display clock"))
		}(dev)
	}
	wg.Wait()
}

package main

import (
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

func main() {
	// 建立SSH客户端连接
	client, err := ssh.Dial("tcp", "192.168.5.201:22", // 交换机的IP和ssh端口，一般为22
		&ssh.ClientConfig{
			User:            "nett",                                       // 交换机的username
			Auth:            []ssh.AuthMethod{ssh.Password("Huawei@123")}, // 交换机的password
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		})
	if err != nil {
		log.Fatalf("SSH dial error: %s", err.Error())
	}

	// 通过客户端建立新会话
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("new session error: %s", err.Error())
	}
	defer session.Close()

	session.Stdout = os.Stdout // 会话输出关联到系统标准输出设备
	session.Stderr = os.Stderr // 会话错误输出关联到系统标准错误输出设备
	session.Stdin = os.Stdin   // 会话输入关联到系统标准输入设备

	// 创建一个伪终端
	terminalModes := ssh.TerminalModes{
		ssh.ECHO:          0,     // 禁用回显（0禁用，1启动）
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, //output speed = 14.4kbaud
	}
	// 配置伪终端参数，包括终端类型、终端的高和宽
	if err = session.RequestPty("xterm", 32, 160, terminalModes); err != nil {
		log.Fatalf("request pty error: %s", err.Error())
	}
	// 开启伪终端
	if err = session.Shell(); err != nil {
		log.Fatalf("start shell error: %s", err.Error())
	}
	// 等待会话终止信号
	if err = session.Wait(); err != nil {
		log.Fatalf("return error: %s", err.Error())
	}
}

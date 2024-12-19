# SSH伪终端

> 本章节介绍下网络工程师比较熟悉的SSH，日常中我们经常会用到终端软件和设备进行交互，最常见的是SecureCRT、Xshell、Putty以及MobaXterm，个人比较喜欢用Moba，可能是因为它集成的sftp比较方便，另外内置的认证算法也比较多，总的来说兼容感比较好。
>
> 不过这些软件的底层都是很相似的，针对ssh来说，都是提供了一个模拟的伪终端与设备进行交互。本章节使用了尽可能短的代码实现了一个伪终端，可以实现与设备的实时交互。

## 环境

使用了eve-ng跑了几台华为的交换机镜像：  
![image-20241209211536651](https://gary-picture.oss-cn-hongkong.aliyuncs.com/202412092115751.png)

全都桥接到了家里的Wi-Fi网络中，网段为192.168.5.0/24，设备下方名称后缀为主机位，例如201代表该设备的管理口地址为192.168.5.201，这个实验演示的是实时交互的伪终端，意味着同时只和一台设备进行实时交互，因此只开启了一台设备，后面的章节会继续介绍同时操作多台设备，我们一步一步来，循序渐进
## 代码

```go
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
```

代码量不是很多，注释尽可能的详细解释了就不展开了，如果有疑问的欢迎交流和指正

## 演示

在代码路径下直接运行go run main.go

![image-20241209213253578](https://gary-picture.oss-cn-hongkong.aliyuncs.com/202412092132608.png)

可以看到后台迅速与设备建立了连接，且可以持续的实时交互，该会话会持续监听，当你在程序端有了终止的动作，例如按了ctrl+C，或者你向通道里发送了quit命令，都会导致会话结束：

```
<HWSW>disp version
disp version
Huawei Versatile Routing Platform Software
VRP (R) software, Version 8.180 (CE6800 V200R005C10SPC607B607)
Copyright (C) 2012-2018 Huawei Technologies Co., Ltd.
HUAWEI CE6800 uptime is 0 day, 0 hour, 32 minutes 
SVRP Platform Version 1.0
<HWSW>

<HWSW>disp clock
disp clock
2024-12-09 21:33:25
Monday
Time Zone(DefaultZoneName) : UTC
<HWSW>

<HWSW>quit
quit
Info: The max number of VTY users is 5, and the number of current VTY users on line is 0.2024/12/09 21:33:55 return error: wait: remote command exited without exit status or exit signal
exit status 1
gary@192 lesson2_ssh_terminal % 
```

## 注意事项

我们日常使用的终端中可以直接进行补全，例如当你输入display ver然后按tab的时候，ver会自动补全成version，且当你直接按问号的时候会出现提示，但是上述代码实现的终端却无法做到：

![image-20241209213830224](https://gary-picture.oss-cn-hongkong.aliyuncs.com/202412092138253.png)

你发现需要在你按了tab之后多加一个enter才可以，同样的问号也是。根本原因是因为当前的会话是单通道，同一时刻要么是被输入使用要么被输出使用，而enter键则被看作是单次交互的标志位。

自动化场景中比较少有这种实时交互终端的需求，以往的工作经历中，很多都是通过堡垒机jumpserver或者vdi等云桌面去实现会话的鉴权与审计。

如果你很想自己去实现一个属于自己的终端软件，你可以通过go的channel去实现输入与输出的实时交互，这里只提供思路，具体实现感兴趣的可以自己研究



# 批量推送命令到多台设备

>日常工作中使用频次最高的功能，也是大部份网工入门自动化首先遇到的刚需，当接到类似‘采集1000台交换机的SN’的任务时，我们第一反应会想到，是不是可以使用自动化去批量实现，本文会提供思路以及demo帮助你完成这个想法

- [批量推送命令到多台设备](#批量推送命令到多台设备)
	- [需求案例](#需求案例)
	- [案例分析](#案例分析)
	- [实验拓扑](#实验拓扑)
	- [基础Demo](#基础demo)
		- [代码](#代码)
		- [运行结果](#运行结果)
		- [代码解读](#代码解读)
	- [最终实践](#最终实践)
		- [代码](#代码-1)
		- [执行结果](#执行结果)
		- [代码优化](#代码优化)
	- [总结与思考](#总结与思考)
## 需求案例

因近期收到较多时钟不同步的告警，工程师小明接到领导需求，查看所有交换机的时间是否正常

## 案例分析

- 常规做法：准备一个设备清单Excel，逐个登陆设备，输入查看时间的命令进行核验，将结果登记在Excel上
- 自动化做法：拿到设备的清单，批量登陆设备，发送查看时间的命令，核验结果，汇总到Excel上

## 实验拓扑

>仍然使用3台CE6800作为目标设备，IP地址如图所示，我的代码则在我个人PC上执行，通过Net与3台交换机通信

![20241219115000](https://gary-picture.oss-cn-hongkong.aliyuncs.com/20241219115000.png)

## 基础Demo

### 代码

```go
package main

import (
	"bytes"
	"fmt"
	"log"

	"golang.org/x/crypto/ssh"
)

func main() {
	// 建立SSH客户端连接
	client, err := ssh.Dial("tcp", "192.168.5.201:22", &ssh.ClientConfig{
		User:            "nett",
		Auth:            []ssh.AuthMethod{ssh.Password("Huawei@123")},
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

	session.Run("display clock\n")
	fmt.Println("Output: ", out.String())
}
```
### 运行结果

![20241219154217](https://gary-picture.oss-cn-hongkong.aliyuncs.com/20241219154217.png)

### 代码解读

代码导入了`golang.org/x/crypto/ssh`包，该包提供了SSH客户端和服务器的实现。然后，在main函数中，代码尝试建立一个SSH客户端连接。使用ssh.Dial函数，指定连接类型为tcp，目标地址为192.168.5.201:22，并传入一个ssh.ClientConfig配置对象。配置对象中包含了用户名nett、认证方法（这里使用密码认证Huawei@123）以及一个忽略主机密钥验证的回调函数ssh.InsecureIgnoreHostKey。

如果连接失败，代码会使用log.Fatalf函数输出错误信息并终止程序执行。成功连接后，代码通过client.NewSession方法创建一个新的SSH会话。如果会话创建失败，同样会输出错误信息并终止程序。

在成功创建会话后，代码使用defer关键字确保会话在函数结束时关闭。接下来，代码定义了一个bytes.Buffer类型的变量out，并将会话的标准输出重定向到该变量。

然后，代码使用session.Run方法在远程服务器上执行命令display clock\n。该命令的输出会被捕获到之前定义的out变量中。最后，代码使用fmt.Println函数将命令的输出打印到控制台。

## 最终实践
通过上述的demo实例，我们可以实现单台的去执行命令，并得到命令的执行结果，那接下来批量执行就容易的多。
我们将上述demo的**通用变量**提取出来，作为外参，将**回显**作为函数的返回结果，然后使用循环或者协程（推荐）依次推送到多个目标设备

### 代码
```go
package main

import (
	"bytes"
	"fmt"
	"log"

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
	for _, dev := range devices {
		fmt.Printf("设备IP: %s, 回显为：\n", dev.IP)
		fmt.Println(pushOne(dev.IP+":22", dev.User, dev.Passwd, "display clock"))
	}
}
```
### 执行结果

![20241219160913](https://gary-picture.oss-cn-hongkong.aliyuncs.com/20241219160913.png)

可以看到3台设备不到一秒的时间里就执行完成了，且回显按顺序返回。

### 代码优化
如果仅仅止步于此，那其实是没有必要使用Golang的，使用Python也可以很容易做到，而且可能代码量更少。因此我们需要将上述的for循环使用协程去改进，具体做法就是将for循环代码替换成下面的代码：

```go
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
```

这段代码展示了如何使用Go语言的sync.WaitGroup来管理并发操作。首先，代码创建了一个WaitGroup实例wg。然后，调用wg.Add(len(devices))方法，传入设备列表的长度，表示将有多少个并发操作需要等待完成。

接下来，代码使用for循环遍历设备列表devices，对每个设备启动一个新的goroutine。在每个goroutine中，首先使用defer wg.Done()确保在goroutine结束时调用wg.Done()，这会减少WaitGroup的计数器。然后，代码使用fmt.Printf打印设备的IP地址，并调用pushOne函数通过SSH连接到设备并执行命令display clock，最后打印命令的输出结果。

所有goroutine启动后，代码调用wg.Wait()方法，这会阻塞主goroutine，直到所有并发操作完成，即WaitGroup的计数器变为零。这样可以确保所有设备的命令执行完成后，主程序才会继续执行。

执行结果：
```shell
gary@192 lesson3_ssh_cli % go run main.go
设备IP: 192.168.5.203, 回显为：
设备IP: 192.168.5.202, 回显为：
设备IP: 192.168.5.201, 回显为：

Info: The max number of VTY users is 5, the number of current VTY users online is 1, and total number of terminal users online is 2.
      The current login time is 2024-12-19 16:13:12.
      The last login time is 2024-12-19 16:12:45 from 192.168.5.25 through SSH.
<HWSW>
2024-12-19 16:13:12
Thursday
Time Zone(DefaultZoneName) : UTC
<HWSW>
Info: The max number of VTY users is 5, and the number of current VTY users on line is 0.

Info: The max number of VTY users is 5, the number of current VTY users online is 1, and total number of terminal users online is 2.
      The current login time is 2024-12-19 16:13:12.
      The last login time is 2024-12-19 16:12:45 from 192.168.5.25 through SSH.
<HWSW>
2024-12-19 16:13:12
Thursday
Time Zone(DefaultZoneName) : UTC
<HWSW>
Info: The max number of VTY users is 5, and the number of current VTY users on line is 0.

Info: The max number of VTY users is 5, the number of current VTY users online is 1, and total number of terminal users online is 2.
      The current login time is 2024-12-19 16:13:12.
      The last login time is 2024-12-19 16:12:45 from 192.168.5.25 through SSH.
<HWSW>
2024-12-19 16:13:12
Thursday
Time Zone(DefaultZoneName) : UTC
<HWSW>
Info: The max number of VTY users is 5, and the number of current VTY users on line is 0.
gary@192 lesson3_ssh_cli % 
```

由于协程的特性，for循环中的`fmt.Printf("设备IP: %s, 回显为：\n", dev.IP)`被首先执行，然后才跟上了每个设备的回显，相信读者都有能力去做进一步的优化
这里只有3台设备，因此在体感上可能速度与纯使用for循环差距不大，但有条件的读者可以试下，当数量越来越多，甚至超过千台的时候，go的优势就很明显了

## 总结与思考

通过本文的介绍，读者大体应该可以掌握批量的去向多台设备推送命令以及获取回显。
- 细心的读者可能发现，当执行其他命令例如`display interface brief`时，程序会卡住，可以先思考下为什么卡住？
- 如果你是一个‘懒到极致’的工程师，你可能又会想，我虽然省去了逐个登陆设备这一步骤，但是我还是需要去查看结果以及确认没台的时钟是否正常，回显太多了，我看着都眼花了，有没有优化的手段，可以进一步的让我‘躺平’？

别急，后面会慢慢道来 \^.^
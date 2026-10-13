package tool

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss/win"
)

func shellStart(filename string) error {
	cmd := exec.Command("cmd.exe")
	//start把第一个带引号的参数视为窗口标题,需前置空标题
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: `/c start "" ` + startArg(filename)}
	return cmd.Start()
}

// startArg 处理start命令的程序路径,防止路径包含空格时解析错误
func startArg(cmdStr string) string {
	cmdStr = strings.TrimSpace(cmdStr)
	//调用方已自带引号,不做处理
	if strings.HasPrefix(cmdStr, `"`) {
		return cmdStr
	}
	//整体是存在的路径(无参数),含空格时加引号
	if _, err := os.Stat(cmdStr); err == nil {
		if strings.ContainsAny(cmdStr, " \t") {
			return `"` + cmdStr + `"`
		}
		return cmdStr
	}
	//携带参数时,按可执行文件后缀定位程序部分,含空格则加引号
	for _, ext := range []string{".exe", ".bat", ".cmd", ".com"} {
		if i := strings.Index(strings.ToLower(cmdStr), ext); i > 0 {
			name := cmdStr[:i+len(ext)]
			if _, err := os.Stat(name); err == nil {
				if strings.ContainsAny(name, " \t") {
					return `"` + name + `"` + cmdStr[i+len(ext):]
				}
				return name + cmdStr[i+len(ext):]
			}
		}
	}
	return cmdStr
}

func shellRun(filename string) error {
	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: fmt.Sprintf("/c \"%s\"", filename)}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func PowerShellRun(filename string) error {
	fmt.Printf("执行: %s\n", filename)
	cmd := exec.Command("powershell.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: fmt.Sprintf("-Command \"%s\"", filename)}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func PublishNotice(message *notice.Message) error {
	return notice.NewWindows().Publish(message)
}

func APPPath(arg string) ([]string, error) {
	return win.APPPath(arg)
}

// Shortcut 创建快捷方式,例Shortcut("xx/Desktop/google.lnk","https://google.cn")
func Shortcut(filename, target string) error {
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	oleShellObject, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return err
	}
	defer oleShellObject.Release()
	wshell, err := oleShellObject.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer wshell.Release()
	cs, err := oleutil.CallMethod(wshell, "CreateShortcut", filename)
	if err != nil {
		return err
	}

	idispatch := cs.ToIDispatch()
	_, err = oleutil.PutProperty(idispatch, "IconLocation", "%SystemRoot%\\System32\\SHELL32.dll,0")
	if err != nil {
		return err
	}
	_, err = oleutil.PutProperty(idispatch, "TargetPath", target)
	if err != nil {
		return err
	}
	_, err = oleutil.PutProperty(idispatch, "Arguments", "")
	if err != nil {
		return err
	}
	_, err = oleutil.PutProperty(idispatch, "Description", "")
	if err != nil {
		return err
	}
	_, err = oleutil.PutProperty(idispatch, "Hotkey", "")
	if err != nil {
		return err
	}
	_, err = oleutil.PutProperty(idispatch, "WindowStyle", "1")
	if err != nil {
		return err
	}
	_, err = oleutil.PutProperty(idispatch, "WorkingDirectory", "")
	if err != nil {
		return err
	}
	_, err = oleutil.CallMethod(idispatch, "Save")
	if err != nil {
		return err
	}
	return nil
}

package tool

import (
	"fmt"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss/win"
	"os"
	"os/exec"
	"syscall"
)

func shellStart2(filename string) error {
	filename = `"" "` + filename + `"`
	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: "/c start " + filename}
	return cmd.Start()
}

func shellRun(filename string) error {
	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: fmt.Sprintf("/c \"%s\"", filename)}
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

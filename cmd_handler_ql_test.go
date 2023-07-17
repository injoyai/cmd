package main

import (
	"github.com/injoyai/goutil/oss/shell"
	"testing"
)

func Test_openBrowser(t *testing.T) {
	t.Log(shell.Exec("cmd", "/c start http://192.168.3.100:10001"))
	t.Log(shell.OpenBrowser("http://192.168.3.100:10001"))
}

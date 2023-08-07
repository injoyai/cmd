package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"testing"
)

func Test_handleVersion(t *testing.T) {
	<-dial.RedialTCP("127.0.0.1:80", func(c *io.Client) {
		c.Debug()
	}).DoneAll()
}

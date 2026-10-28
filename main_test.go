package main

import (
	"bytes"
	"github.com/gorilla/websocket"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCLIProcess(t *testing.T) {
	if os.Getenv("CMD_TEST_PROCESS") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"i"}, os.Args[i+1:]...)
			break
		}
	}
	main()
	os.Exit(0)
}

func cliProcess(args ...string) *exec.Cmd {
	c := exec.Command(os.Args[0], append([]string{"-test.run=^TestCLIProcess$", "--"}, args...)...)
	c.Env = append(os.Environ(), "CMD_TEST_PROCESS=1")
	return c
}

func TestCLIExitStatus(t *testing.T) {
	for _, tt := range []struct {
		args []string
		fail bool
	}{
		{[]string{"version"}, false}, {[]string{"--help"}, false},
		{[]string{"unknown-command"}, true}, {[]string{"dl"}, true},
		{[]string{"upload", "minio", filepath.Join(t.TempDir(), "missing")}, true},
		{[]string{"http"}, true},
		{[]string{"test", "download", "://bad"}, true},
		{[]string{"test", "upload", "://bad"}, true}, {[]string{"server", "tcp", "--port=-1"}, true},
	} {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			c := cliProcess(tt.args...)
			var out, stderr bytes.Buffer
			c.Stdout = &out
			c.Stderr = &stderr
			err := c.Run()
			if (err != nil) != tt.fail {
				t.Fatalf("error=%v stdout=%s stderr=%s", err, &out, &stderr)
			}
			if tt.fail && stderr.Len() == 0 {
				t.Fatalf("failure missing stderr: %s", &out)
			}
		})
	}
}

func TestDialWSSKeepsTLS(t *testing.T) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	listener.SetDeadline(time.Now().Add(5 * time.Second))
	c := cliProcess("dial", "wss://"+listener.Addr().String()+"/ws", "--debug=false", "--redial=false")
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { c.Process.Kill(); c.Wait() }()
	conn, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var first [1]byte
	if _, err := conn.Read(first[:]); err != nil {
		t.Fatal(err)
	}
	if first[0] != 0x16 {
		t.Fatalf("expected TLS handshake, got first byte %#x", first[0])
	}
}

func TestWebsocketServerAcceptsExtensions(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	_, port, _ := net.SplitHostPort(address)
	listener.Close()
	c := cliProcess("server", "websocket", "--port="+port, "--debug=false")
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { c.Process.Kill(); c.Wait() }()
	ready := false
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			ready = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !ready {
		t.Fatal("server did not start")
	}
	dialer := websocket.Dialer{EnableCompression: true, HandshakeTimeout: time.Second}
	conn, resp, err := dialer.Dial("ws://"+address+"/ws", nil)
	if err != nil {
		t.Fatalf("handshake failed: %v response=%v", err, resp)
	}
	conn.Close()
}

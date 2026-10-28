package main

import (
	"bytes"
	"crypto/rand"
	"github.com/gorilla/websocket"
	"io"
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

func TestDialTCPStdinAndNoRedial(t *testing.T) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	listener.SetDeadline(time.Now().Add(5 * time.Second))
	c := cliProcess("dial", "tcp", listener.Addr().String(), "--debug=false", "--redial=false")
	c.Stdin = strings.NewReader("0x6869\n")
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	defer c.Process.Kill()
	wait := make(chan error, 1)
	go func() { wait <- c.Wait() }()
	conn, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	got := make([]byte, 2)
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if string(got) != "hi" {
		t.Fatalf("stdin payload = %q, want hi", got)
	}
	conn.Close()
	select {
	case <-wait:
	case <-time.After(5 * time.Second):
		t.Fatal("dial did not exit after disconnect with --redial=false")
	}
}

func TestDeployV2RoundTrip(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	_, port, _ := net.SplitHostPort(address)
	listener.Close()
	server := cliProcess("server", "deploy", "--port="+port)
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { server.Process.Kill(); server.Wait() }()
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
		t.Fatal("deploy server did not start")
	}
	// 不可压缩的数据让请求超过进度复制的缓冲区，验证多次写入仍是一帧。
	data := make([]byte, 96*1024)
	if _, err := rand.Read(data); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "payload.bin")
	target := filepath.Join(t.TempDir(), "payload.bin")
	if err := os.WriteFile(source, data, 0600); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name   string
		target string
		fail   bool
	}{
		{"success", target, false},
		{"server failure", filepath.Join(source, "payload.bin"), true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := cliProcess("dial", "deploy", address, "--source="+source, "--target="+tt.target, "--restart=false")
			var out bytes.Buffer
			client.Stdout, client.Stderr = &out, &out
			if err := client.Start(); err != nil {
				t.Fatal(err)
			}
			defer client.Process.Kill()
			wait := make(chan error, 1)
			go func() { wait <- client.Wait() }()
			select {
			case err := <-wait:
				if (err != nil) != tt.fail {
					t.Fatalf("deploy error = %v, want failure %v\n%s", err, tt.fail, &out)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("deploy timed out")
			}
			code := `"code":200`
			if tt.fail {
				code = `"code":500`
			}
			if !strings.Contains(out.String(), code) {
				t.Fatalf("missing response %s: %s", code, &out)
			}
		})
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("deployed file differs from source")
	}
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

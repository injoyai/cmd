package resource

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"syscall"
	"testing"
)

func TestNetworkErrorClassification(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
		want bool
	}{
		{"connection reset", &url.Error{Op: "Get", Err: &net.OpError{Op: "read", Err: syscall.ECONNRESET}}, true},
		{"connection refused", &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}, true},
		{"invalid address", &net.OpError{Op: "dial", Err: &net.AddrError{Err: "invalid port", Addr: "host:99999"}}, false},
		{"blocked network", &net.OpError{Op: "dial", Err: syscall.EACCES}, false},
		{"timeout", &url.Error{Op: "Get", Err: context.DeadlineExceeded}, true},
		{"temporary DNS", &net.DNSError{IsTemporary: true}, true},
		{"missing host", &net.DNSError{IsNotFound: true}, false},
		{"certificate", &net.OpError{Op: "dial", Err: x509.UnknownAuthorityError{}}, false},
		{"unsupported protocol", &url.Error{Op: "Get", Err: errors.New("unsupported protocol")}, false},
		{"truncated body", io.ErrUnexpectedEOF, true},
		{"connection EOF", io.EOF, true},
		{"file permission", &os.PathError{Op: "write", Err: os.ErrPermission}, false},
		{"local EOF", &os.PathError{Op: "read", Err: io.EOF}, false},
		{"cancelled", &url.Error{Op: "Get", Err: context.Canceled}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := NetworkError(tt.err)
			if got := IsRetryable(fmt.Errorf("download: %w", err)); got != tt.want {
				t.Fatalf("IsRetryable() = %v, want %v", got, tt.want)
			}
			if !errors.Is(err, tt.err) {
				t.Fatal("classification lost the original cause")
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	for _, tt := range []struct {
		code int
		want bool
	}{
		{400, false}, {401, false}, {403, false}, {404, false},
		{408, true}, {425, true}, {429, true}, {500, true},
		{501, false}, {502, true}, {503, true}, {504, true}, {505, false},
	} {
		t.Run(fmt.Sprint(tt.code), func(t *testing.T) {
			if got := IsRetryable(&HTTPError{StatusCode: tt.code}); got != tt.want {
				t.Fatalf("HTTP %d: retryable = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
	transient := NetworkError(io.ErrUnexpectedEOF)
	if !IsRetryable(errors.Join(transient, &HTTPError{StatusCode: 503})) {
		t.Fatal("all transient causes must remain retryable")
	}
	for _, err := range []error{
		nil, errors.New("invalid archive"), io.ErrUnexpectedEOF,
		errors.Join(transient, os.ErrPermission), errors.Join(transient, context.Canceled),
	} {
		if IsRetryable(err) {
			t.Fatalf("permanent or unclassified error is retryable: %v", err)
		}
	}
}

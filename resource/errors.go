package resource

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

// HTTPError 保留响应状态，避免通过错误文本判断是否需要重试。
type HTTPError struct {
	StatusCode int
	URL        string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d %s %s", e.StatusCode, http.StatusText(e.StatusCode), RedactURL(e.URL))
}

func (e *HTTPError) Retryable() bool {
	switch e.StatusCode {
	case 408, 425, 429, 500, 502, 503, 504:
		return true
	}
	return false
}

type networkError struct{ err error }

func (e *networkError) Error() string   { return e.err.Error() }
func (e *networkError) Unwrap() error   { return e.err }
func (e *networkError) Retryable() bool { return true }

// NetworkError 仅在请求或读取响应失败时标记临时故障；解析和本地文件错误不在此处处理。
func NetworkError(err error) error {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, os.ErrPermission) {
		return err
	}
	var path *os.PathError
	var link *os.LinkError
	if errors.As(err, &path) || errors.As(err, &link) {
		return err
	}
	var address *net.AddrError
	var invalidAddress net.InvalidAddrError
	var network net.UnknownNetworkError
	if errors.As(err, &address) || errors.As(err, &invalidAddress) || errors.As(err, &network) {
		return err
	}
	var dns *net.DNSError
	if errors.As(err, &dns) {
		if !dns.IsNotFound && (dns.IsTimeout || dns.IsTemporary) {
			return &networkError{err: err}
		}
		return err
	}
	var certificate *tls.CertificateVerificationError
	var authority x509.UnknownAuthorityError
	var hostname x509.HostnameError
	var invalid x509.CertificateInvalidError
	var record tls.RecordHeaderError
	if errors.As(err, &certificate) || errors.As(err, &authority) ||
		errors.As(err, &hostname) || errors.As(err, &invalid) || errors.As(err, &record) {
		return err
	}
	var op *net.OpError
	var timeout interface{ Timeout() bool }
	if errors.As(err, &op) || (errors.As(err, &timeout) && timeout.Timeout()) ||
		errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return &networkError{err: err}
	}
	return err
}

// IsRetryable 默认不重试未知错误；多个原因中只要包含永久错误，就停止本次任务。
func IsRetryable(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) {
		return false
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		found := false
		for _, cause := range joined.Unwrap() {
			if cause == nil {
				continue
			}
			if !IsRetryable(cause) {
				return false
			}
			found = true
		}
		return found
	}
	if classified, ok := err.(interface{ Retryable() bool }); ok {
		return classified.Retryable()
	}
	return IsRetryable(errors.Unwrap(err))
}

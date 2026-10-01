package mail_service

import (
	"net"
	"testing"
	"time"
)

func TestConnectPlainDoesNotHangOnSilentServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			// Simulate an implicit-TLS endpoint reached over plaintext:
			// the connection is accepted but no SMTP greeting is ever sent.
			_ = conn
		}
	}()

	host, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}

	e := &EmailSender{Host: host, Port: port}

	done := make(chan error, 1)
	go func() {
		done <- e.connectPlain()
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected connectPlain to fail against a silent server, got nil")
		}
	case <-time.After(40 * time.Second):
		t.Fatal("connectPlain hung on a silent server instead of timing out")
	}
}

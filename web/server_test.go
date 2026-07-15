package web

import (
	"net"
	"testing"
	"time"

	"github.com/nabilulilalbab/promote/utils"
)

func TestStartServer_PortInUseReturnsError(t *testing.T) {
	t.Chdir(t.TempDir())

	s := &DashboardServer{logger: utils.NewLogger("TEST", false)}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("cannot reserve port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	defer ln.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.StartServer(port)
	}()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error when port is already in use, got nil")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("StartServer did not return error for an occupied port (possible silent failure)")
	}
}

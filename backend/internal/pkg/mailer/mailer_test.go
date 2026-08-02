package mailer

import (
	"bufio"
	"context"
	"log/slog"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeSMTPServer implements just enough of RFC 5321 to capture the message a
// client submits via smtp.SendMail.
type fakeSMTPServer struct {
	ln     net.Listener
	msgs   chan string
	closed chan struct{}
}

func startFakeSMTPServer(t *testing.T) (*fakeSMTPServer, string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	s := &fakeSMTPServer{ln: ln, msgs: make(chan string, 4), closed: make(chan struct{})}
	go s.serve()
	t.Cleanup(func() {
		close(s.closed)
		ln.Close()
	})
	return s, ln.Addr().String()
}

func (s *fakeSMTPServer) serve() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *fakeSMTPServer) handle(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	write := func(line string) {
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		conn.Write([]byte(line + "\r\n"))
	}
	write("220 fake-smtp ESMTP")

	var inData bool
	var data strings.Builder
	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		if inData {
			if line == "." {
				inData = false
				select {
				case s.msgs <- data.String():
				default:
				}
				data.Reset()
				write("250 OK: queued")
				continue
			}
			data.WriteString(line)
			data.WriteString("\n")
			continue
		}
		switch {
		case strings.HasPrefix(line, "EHLO"):
			write("250-fake-smtp")
			write("250 OK")
		case strings.HasPrefix(line, "MAIL FROM"):
			write("250 OK")
		case strings.HasPrefix(line, "RCPT TO"):
			write("250 OK")
		case strings.HasPrefix(line, "DATA"):
			inData = true
			write("354 End data with <CR><LF>.<CR><LF>")
		case strings.HasPrefix(line, "QUIT"):
			write("221 Bye")
			return
		default:
			write("250 OK")
		}
	}
}

func TestMailerDisabledLogsInsteadOfSending(t *testing.T) {
	t.Parallel()
	var logs []string
	logger := slog.New(slog.NewTextHandler(&bufferWriter{fn: func(s string) { logs = append(logs, s) }}, nil))

	m := New(Config{SMTPHost: "", FrontendURL: "http://localhost:3000", Log: logger})
	if m.Enabled() {
		t.Fatal("Enabled() = true, want false with empty host")
	}
	if err := m.SendVerificationEmail(context.Background(), "a@b.com", "tok123"); err != nil {
		t.Fatalf("disabled mailer must not error: %v", err)
	}
	if len(logs) == 0 {
		t.Error("expected dev-mode log output")
	}
}

func TestMailerSendVerificationEmail(t *testing.T) {
	t.Parallel()
	srv, addr := startFakeSMTPServer(t)
	host, port, _ := net.SplitHostPort(addr)

	m := New(Config{
		SMTPHost:    host,
		SMTPPort:    atoi(port),
		From:        "no-reply@schemahub.dev",
		FromName:    "SchemaHub",
		FrontendURL: "http://localhost:3000",
		Log:         slog.New(slog.DiscardHandler),
	})
	if !m.Enabled() {
		t.Fatal("Enabled() = false, want true")
	}

	if err := m.SendVerificationEmail(context.Background(), "dev@example.com", "verify_abc123"); err != nil {
		t.Fatalf("send: %v", err)
	}

	select {
	case msg := <-srv.msgs:
		if !strings.Contains(msg, "Subject: Verify your email") {
			t.Errorf("missing subject in:\n%s", msg)
		}
		if !strings.Contains(msg, "http://localhost:3000/verify-email?token=verify_abc123") {
			t.Errorf("missing verify link in:\n%s", msg)
		}
		if !strings.Contains(msg, "To: dev@example.com") {
			t.Errorf("missing recipient in:\n%s", msg)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestMailerSendPasswordResetEmail(t *testing.T) {
	t.Parallel()
	srv, addr := startFakeSMTPServer(t)
	host, port, _ := net.SplitHostPort(addr)

	m := New(Config{SMTPHost: host, SMTPPort: atoi(port), FrontendURL: "http://localhost:3000"})
	if err := m.SendPasswordResetEmail(context.Background(), "dev@example.com", "reset_x"); err != nil {
		t.Fatalf("send: %v", err)
	}

	select {
	case msg := <-srv.msgs:
		if !strings.Contains(msg, "Subject: Reset your password") {
			t.Errorf("missing subject in:\n%s", msg)
		}
		if !strings.Contains(msg, "http://localhost:3000/forgot-password/reset?token=reset_x") {
			t.Errorf("missing reset link in:\n%s", msg)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}

type bufferWriter struct {
	mu sync.Mutex
	fn func(string)
}

func (w *bufferWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.fn(string(p))
	return len(p), nil
}

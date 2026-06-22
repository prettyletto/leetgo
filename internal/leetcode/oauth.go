package leetcode

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

const (
	leetcodeLoginURL = "https://leetcode.com/accounts/login/"
	callbackPath     = "/callback"
)

func (c *Client) browserSessionAuth(ctx context.Context) error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("start callback server: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	callbackURL := fmt.Sprintf("http://127.0.0.1:%d%s", port, callbackPath)

	resultCh := make(chan sessionAuthResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		csrf := r.URL.Query().Get("csrf_token")
		session := r.URL.Query().Get("session_id")

		if csrf == "" || session == "" {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte("Missing tokens. Please try again.")); err != nil {
				resultCh <- sessionAuthResult{err: fmt.Errorf("write callback response: %w", err)}
				return
			}
			resultCh <- sessionAuthResult{err: fmt.Errorf("missing tokens in callback")}
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(`
			<html>
			<body style="font-family: sans-serif; text-align: center; padding: 50px;">
				<h1>✅ Leetgo Connected!</h1>
				<p>You can close this tab and return to the terminal.</p>
			</body>
			</html>
		`)); err != nil {
			resultCh <- sessionAuthResult{err: fmt.Errorf("write callback response: %w", err)}
			return
		}

		resultCh <- sessionAuthResult{
			csrf:    csrf,
			session: session,
		}
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)
	defer server.Shutdown(ctx)

	instructions := fmt.Sprintf(
		"\n🔐 LeetCode Authentication\n\n"+
			"1. Open this URL in your browser:\n   %s\n\n"+
			"2. Log in to LeetCode\n"+
			"3. Copy your CSRF token and session ID from browser cookies\n"+
			"4. Visit: %s?csrf_token=<CSRF>&session_id=<SESSION>\n\n"+
			"Waiting for callback...\n",
		leetcodeLoginURL, callbackURL,
	)
	fmt.Print(instructions)

	if err := openBrowser(leetcodeLoginURL); err != nil {
		fmt.Printf("Could not open browser automatically: %v\n", err)
	}

	select {
	case result := <-resultCh:
		if result.err != nil {
			return result.err
		}
		c.session = &Session{
			CSRFToken: result.csrf,
			SessionID: result.session,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}
		return c.saveSession()

	case <-ctx.Done():
		return ctx.Err()

	case <-time.After(5 * time.Minute):
		return fmt.Errorf("authentication timed out")
	}
}

type sessionAuthResult struct {
	csrf    string
	session string
	err     error
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		return fmt.Errorf("unsupported platform")
	}

	if err := exec.Command(cmd, args...).Start(); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	return nil
}

package notify

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Notifier sends desktop or webhook notifications for port events.
type Notifier struct {
	webhookURL string
	desktop    bool
}

// New creates a Notifier. If webhookURL is empty, webhook notifications are
// disabled. If desktop is true, OS-level notifications are attempted.
func New(webhookURL string, desktop bool) *Notifier {
	return &Notifier{webhookURL: webhookURL, desktop: desktop}
}

// Opened sends a notification that a new port was detected.
func (n *Notifier) Opened(port int) error {
	msg := fmt.Sprintf("portwatch: port %d opened", port)
	return n.send("Port Opened", msg)
}

// Closed sends a notification that a port is no longer open.
func (n *Notifier) Closed(port int) error {
	msg := fmt.Sprintf("portwatch: port %d closed", port)
	return n.send("Port Closed", msg)
}

func (n *Notifier) send(title, body string) error {
	var errs []string

	if n.desktop {
		if err := sendDesktop(title, body); err != nil {
			errs = append(errs, fmt.Sprintf("desktop: %v", err))
		}
	}

	if n.webhookURL != "" {
		if err := sendWebhook(n.webhookURL, title, body); err != nil {
			errs = append(errs, fmt.Sprintf("webhook: %v", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("notify errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func sendDesktop(title, body string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("notify-send", title, body).Run()
	case "darwin":
		script := fmt.Sprintf(`display notification %q with title %q`, body, title)
		return exec.Command("osascript", "-e", script).Run()
	default:
		return fmt.Errorf("desktop notifications not supported on %s", runtime.GOOS)
	}
}

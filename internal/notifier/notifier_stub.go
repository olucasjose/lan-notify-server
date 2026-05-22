//go:build !linux

package notifier

import (
	"fmt"
)

type StubNotifier struct{}

func newNativeNotifier() (Notifier, error) {
	return &StubNotifier{}, nil
}

func (n *StubNotifier) Notify(title, message string, opts NotifyOptions) error {
	fmt.Printf("[Mock Notification] %s: %s (Urgency: %s)\n", title, message, opts.Urgency)
	return nil
}

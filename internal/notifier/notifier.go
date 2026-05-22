package notifier

// NotifyOptions contains advanced options for desktop notifications.
type NotifyOptions struct {
	Urgency  string // "low", "normal", "critical"
	Category string // Optional D-Bus category (e.g., "im.received")
}

// Notifier defines the common interface for showing notifications.
type Notifier interface {
	Notify(title, message string, opts NotifyOptions) error
}

// New returns the appropriate native Notifier for the host OS.
func New() (Notifier, error) {
	return newNativeNotifier()
}

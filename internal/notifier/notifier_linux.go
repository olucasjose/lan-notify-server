//go:build linux

package notifier

import (
	"github.com/godbus/dbus/v5"
)

type LinuxNotifier struct {
	conn *dbus.Conn
}

func newNativeNotifier() (Notifier, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}
	return &LinuxNotifier{conn: conn}, nil
}

func (n *LinuxNotifier) Notify(title, message string, opts NotifyOptions) error {
	obj := n.conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")

	hints := make(map[string]dbus.Variant)
	
	// Map urgency
	var urgency byte = 1 // Normal
	if opts.Urgency == "low" {
		urgency = 0
	} else if opts.Urgency == "critical" {
		urgency = 2
	}
	hints["urgency"] = dbus.MakeVariant(urgency)

	if opts.Category != "" {
		hints["category"] = dbus.MakeVariant(opts.Category)
	}

	call := obj.Call("org.freedesktop.Notifications.Notify", 0,
		"lan-notify",              // app_name
		uint32(0),                 // replaces_id
		"dialog-information",      // app_icon
		title,                     // summary
		message,                   // body
		[]string{},                // actions
		hints,                     // hints
		int32(-1),                 // expire_timeout (-1 for default)
	)

	if call.Err != nil {
		return call.Err
	}
	return nil
}

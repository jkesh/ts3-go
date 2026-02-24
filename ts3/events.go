package ts3

import (
	"context"
	"strconv"
	"strings"
)

// Register registers a raw notification handler by notify event name.
//
// For example:
//   - notifytextmessage
//   - notifycliententerview
//   - notifyclientleftview
func (c *Client) Register(eventName string, callback func(string)) {
	if strings.TrimSpace(eventName) == "" || callback == nil {
		return
	}

	c.notifyMu.Lock()
	c.notifications[eventName] = append(c.notifications[eventName], callback)
	c.notifyMu.Unlock()
}

// Unregister removes all handlers bound to one event name.
func (c *Client) Unregister(eventName string) {
	c.notifyMu.Lock()
	delete(c.notifications, eventName)
	c.notifyMu.Unlock()
}

// dispatchNotify parses a raw notify line and dispatches it to handlers.
func (c *Client) dispatchNotify(rawLine string) {
	parts := strings.SplitN(rawLine, " ", 2)
	eventName := parts[0]
	eventData := ""
	if len(parts) == 2 {
		eventData = parts[1]
	}

	c.notifyMu.RLock()
	handlers := append([]func(string){}, c.notifications[eventName]...)
	c.notifyMu.RUnlock()

	for _, h := range handlers {
		go h(eventData)
	}
}

// RegisterServerEvents subscribes to server-level client enter/leave/move events.
func (c *Client) RegisterServerEvents(ctx context.Context) error {
	_, err := c.Exec(ctx, "servernotifyregister event=server")
	return err
}

// RegisterChannelEvents subscribes to channel-level events.
//
// channelID is optional:
//   - 0: current/default behavior
//   - >0: explicit channel id
func (c *Client) RegisterChannelEvents(ctx context.Context, channelID int) error {
	cmd := "servernotifyregister event=channel"
	if channelID > 0 {
		cmd += " id=" + strconv.Itoa(channelID)
	}
	_, err := c.Exec(ctx, cmd)
	return err
}

// RegisterTextEvents subscribes to private, channel and server text message events.
func (c *Client) RegisterTextEvents(ctx context.Context) error {
	if _, err := c.Exec(ctx, "servernotifyregister event=textprivate"); err != nil {
		return err
	}
	if _, err := c.Exec(ctx, "servernotifyregister event=textserver"); err != nil {
		return err
	}
	if _, err := c.Exec(ctx, "servernotifyregister event=textchannel"); err != nil {
		return err
	}
	return nil
}

// UnregisterNotify unsubscribes current query client from notifications.
func (c *Client) UnregisterNotify(ctx context.Context) error {
	_, err := c.Exec(ctx, "servernotifyunregister")
	return err
}

// OnClientEnter registers a handler for "notifycliententerview".
func (c *Client) OnClientEnter(ctx context.Context, handler func(string)) error {
	if err := c.RegisterServerEvents(ctx); err != nil {
		return err
	}
	c.Register("notifycliententerview", handler)
	return nil
}

// OnClientLeave registers a handler for "notifyclientleftview".
func (c *Client) OnClientLeave(ctx context.Context, handler func(string)) error {
	if err := c.RegisterServerEvents(ctx); err != nil {
		return err
	}
	c.Register("notifyclientleftview", handler)
	return nil
}

// OnTextMessage registers a handler for "notifytextmessage".
func (c *Client) OnTextMessage(ctx context.Context, handler func(string)) error {
	if err := c.RegisterTextEvents(ctx); err != nil {
		return err
	}
	c.Register("notifytextmessage", handler)
	return nil
}

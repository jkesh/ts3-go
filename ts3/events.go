package ts3

import (
	"strings"
)

// Register 注册一个事件处理器
func (c *Client) Register(eventName string, callback func(string)) {
	c.notifyMu.Lock()
	defer c.notifyMu.Unlock()

	c.notifications[eventName] = append(c.notifications[eventName], callback)
}

// dispatchNotify 解析并分发事件
func (c *Client) dispatchNotify(rawLine string) {
	// 1. 提取事件名称
	parts := strings.SplitN(rawLine, " ", 2)
	eventName := parts[0]
	eventData := ""
	if len(parts) > 1 {
		eventData = parts[1]
	}

	c.notifyMu.RLock()
	handlers, ok := c.notifications[eventName]
	c.notifyMu.RUnlock()

	if ok {
		for _, h := range handlers {
			go h(eventData)
		}
	}
}

// OnClientEnter 注册用户进入频道事件
func (c *Client) OnClientEnter(handler func(string)) error {
	_, err := c.Exec("servernotifyregister event=server")
	if err != nil {
		return err
	}
	c.Register("notifycliententerview", handler)
	return nil
}

// OnTextMessage 注册接收消息事件
func (c *Client) OnTextMessage(handler func(string)) error {
	_, err := c.Exec("servernotifyregister event=textprivate")
	if err != nil {
		return err
	}
	_, err = c.Exec("servernotifyregister event=textserver")
	if err != nil {
		return err
	}
	_, err = c.Exec("servernotifyregister event=textchannel")
	if err != nil {
		return err
	}
	c.Register("notifytextmessage", handler)
	return nil
}

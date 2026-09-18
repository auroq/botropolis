package ui

import "time"

// NoticeFor is how long a status line stays on the footer before the
// key row has it to itself again.
const NoticeFor = 4 * time.Second

// Notice is a passing status: what the last key did, an error, a
// prompt. It reads for NoticeFor after it is set, then reads as nothing.
type Notice struct {
	text string
	at   time.Time
}

func (n *Notice) Set(text string, now time.Time) {
	n.text, n.at = text, now
}

func (n *Notice) Clear() {
	n.text = ""
}

// Text is the notice while it is fresh, else "".
func (n *Notice) Text(now time.Time) string {
	if n.text == "" || !now.Before(n.at.Add(NoticeFor)) {
		return ""
	}
	return n.text
}

package proto

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/auroq/botropolis/pkg/state"
)

const (
	OpSnapshot  = "snapshot"
	OpSubscribe = "subscribe"
	OpEvent     = "event"

	socketDir  = "botropolis"
	socketName = "botropolis.sock"
)

type Request struct {
	Op    string          `json:"op"`
	Event json.RawMessage `json:"event,omitempty"`
}

type Response struct {
	Snapshot *state.Snapshot `json:"snapshot,omitempty"`
	Error    string          `json:"error,omitempty"`
}

func SocketPath() string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return filepath.Join(dir, socketDir, socketName)
	}
	return filepath.Join(os.TempDir(), socketDir+"-"+strconv.Itoa(os.Getuid()), socketName)
}

func Write(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

func NewDecoder(r io.Reader) *json.Decoder {
	return json.NewDecoder(r)
}

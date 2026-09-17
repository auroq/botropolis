package proto

import (
	"encoding/json"
	"errors"
	"net"
	"time"

	"github.com/auroq/botropolis/pkg/state"
)

const dialTimeout = 500 * time.Millisecond

type Client struct {
	conn    net.Conn
	decoder *json.Decoder
}

func Dial(sock string) (*Client, error) {
	conn, err := net.DialTimeout("unix", sock, dialTimeout)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, decoder: json.NewDecoder(conn)}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Snapshot() (state.Snapshot, error) {
	if err := Write(c.conn, Request{Op: OpSnapshot}); err != nil {
		return state.Snapshot{}, err
	}
	return c.readSnapshot()
}

func (c *Client) Subscribe() (<-chan state.Snapshot, error) {
	if err := Write(c.conn, Request{Op: OpSubscribe}); err != nil {
		return nil, err
	}
	updates := make(chan state.Snapshot, 1)
	go func() {
		defer close(updates)
		for {
			snapshot, err := c.readSnapshot()
			if err != nil {
				return
			}
			updates <- snapshot
		}
	}()
	return updates, nil
}

func (c *Client) readSnapshot() (state.Snapshot, error) {
	var response Response
	if err := c.decoder.Decode(&response); err != nil {
		return state.Snapshot{}, err
	}
	if response.Error != "" {
		return state.Snapshot{}, errors.New(response.Error)
	}
	if response.Snapshot == nil {
		return state.Snapshot{}, errors.New("response carried no snapshot")
	}
	return *response.Snapshot, nil
}

func SendEvent(sock string, event json.RawMessage) error {
	client, err := Dial(sock)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()
	if err := Write(client.conn, Request{Op: OpEvent, Event: event}); err != nil {
		return err
	}
	var response Response
	if err := client.decoder.Decode(&response); err != nil {
		return err
	}
	if response.Error != "" {
		return errors.New(response.Error)
	}
	return nil
}

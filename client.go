package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

const RTMMessage = "message"

// respRtmStart is the structure of the introductory response
// from Slack
type respRtmStart struct {
	Ok   bool   `json:"ok"`
	URL  string `json:"url"`
	Self struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Prefs struct {
		} `json:"prefs"`
		Created        int    `json:"created"`
		ManualPresence string `json:"manual_presence"`
	} `json:"self"`
	Team struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		EmailDomain string `json:"email_domain"`
		Domain      string `json:"domain"`
		Icon        struct {
		} `json:"icon"`
		MsgEditWindowMins int  `json:"msg_edit_window_mins"`
		OverStorageLimit  bool `json:"over_storage_limit"`
		Prefs             struct {
		} `json:"prefs"`
		Plan string `json:"plan"`
	} `json:"team"`
	Users    []interface{} `json:"users"`
	Channels []interface{} `json:"channels"`
	Groups   []interface{} `json:"groups"`
	Mpims    []interface{} `json:"mpims"`
	Ims      []interface{} `json:"ims"`
	Bots     []interface{} `json:"bots"`
	Error    string        `json:"error"`
}

// Message is the conversation data structure
type Message struct {
	ID    uint64 `json:"id"`
	Type  string `json:"type"`
	Error struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"error"`
	ReplyTo int    `json:"reply_to"`
	Channel string `json:"channel"`
	Ts      string `json:"ts"`
	User    string `json:"user"`
	Text    string `json:"text"`
}

// Client is the slack client object
type Client struct {
	id      string
	conn    *websocket.Conn
	apiURL  string
	token   string
	counter uint64
	mux     *EventMux
}

// NewClient returns an initialised Client pointer
func NewClient(token string, mux *EventMux) *Client {
	return &Client{
		apiURL: "https://slack.com/api",
		token:  token,
		mux:    mux,
	}
}

// start performs a rtm.start, and returns a websocket URL and user ID.
func (c *Client) start() (wsurl string, err error) {
	url := fmt.Sprintf("%s/rtm.start?token=%s", c.apiURL, c.token)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}
	var respObj respRtmStart
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return
	}

	if !respObj.Ok {
		err = fmt.Errorf("Slack error: %s", respObj.Error)
		return
	}

	wsurl = respObj.URL
	c.id = respObj.Self.ID
	return
}

// Connect opens a websocket connection with Slack.
func (c *Client) Connect() error {
	wsurl, err := c.start()
	if err != nil {
		log.Fatal(err)
		return errors.Errorf("Client Connect error: %s", err.Error)
	}

	c.conn, _, err = websocket.DefaultDialer.Dial(wsurl, nil)
	if err != nil {
		log.Fatal(err)
		return errors.Errorf("Client Websocket Connect error: %s", err.Error)
	}

	return nil
}

// Dispatch reads the events from Slack and sends them to the correct Handler
func (c *Client) Dispatch() {
	defer c.conn.Close()

	for {
		msg, err := c.getMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		if v, ok := c.mux.m[msg.Type]; ok {
			go v.handler.ServeEvent(msg, c)
		} else {
			log.Printf("recv: %+v", msg)
		}
	}
}

// Shutdown cleanly closes a connection. Shutdown sends a close
// frame and waits for the server to close the connection.
func (c *Client) Shutdown() error {
	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return errors.Errorf("Client Shutdown error: %s", err.Error)
	}
	return nil
}

//Close immediately closes the websocket. For a graceful shutdown, call Shutdown first.
func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) getMessage() (*Message, error) {
	msg := &Message{}
	err := c.conn.ReadJSON(msg)
	if err != nil {
		return msg, errors.Errorf("Client getMessage error: %s", err.Error)
	}
	return msg, nil
}

func (c *Client) PostMessage(m *Message) error {
	m.ID = atomic.AddUint64(&c.counter, 1)
	err := c.conn.WriteJSON(m)
	if err != nil {
		return errors.Errorf("Client PostMessage error: %s", err.Error)
	}
	return nil
}

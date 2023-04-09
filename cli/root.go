package cli

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

func NewRootCmd() (cmd *cobra.Command) {
	opt := &rootOption{}
	cmd = &cobra.Command{
		Use:   "gows",
		Short: "Echo the message from a websoket server",
		Args:  cobra.MinimumNArgs(1),
		RunE:  opt.runE,
	}
	flags := cmd.Flags()
	flags.StringVarP(&opt.server, "server", "s", "", "The server address")
	cmd.MarkFlagRequired("server")
	return
}

type rootOption struct {
	server string
}

func (o *rootOption) runE(cmd *cobra.Command, args []string) (err error) {
	service := args[0]
	server := url.URL{Scheme: "ws", Host: o.server, Path: service}

	header := http.Header{}
	header.Set("Cookie", "oAuthLoginInfo=; token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODA2MTE1OTYsImlhdCI6MTY4MDYwNDM5NiwiaXNzIjoia3ViZXNwaGVyZSIsInN1YiI6ImFkbWluIiwidG9rZW5fdHlwZSI6ImFjY2Vzc190b2tlbiIsInVzZXJuYW1lIjoiYWRtaW4ifQ.kY7jpe4If37_9Kz_U9R5s6q2jIwzZ7tTaYI6KbXee_Y; expire=1680611597080; refreshToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODA2MTg3OTYsImlhdCI6MTY4MDYwNDM5NiwiaXNzIjoia3ViZXNwaGVyZSIsInN1YiI6ImFkbWluIiwidG9rZW5fdHlwZSI6InJlZnJlc2hfdG9rZW4iLCJ1c2VybmFtZSI6ImFkbWluIn0.8niwD-WQK7LMDb-T9B177v0rVpWWP21F7qc8ve2qNrc; lang=zh")

	var conn *websocket.Conn
	var resp *http.Response
	if conn, resp, err = websocket.DefaultDialer.Dial(server.String(), header); err != nil {
		cmd.Println("failed to dial, response is", resp)
		return
	}
	defer conn.Close()

	done := make(chan struct{}, 1)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	conn.WriteMessage(websocket.TextMessage, []byte("\n"))
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			data := []byte(scanner.Text() + "\n")
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				fmt.Println("failed to send message", err)
			}
		}
	}()

	go func() {
		defer close(done)
		for {
			if _, message, msgErr := conn.ReadMessage(); msgErr != nil {
				err = msgErr
				return
			} else {
				cmd.Print(string(message))
			}
		}
	}()

	for {
		select {
		case <-interrupt:
			return
		case <-done:
		}
	}
}

package cli

import (
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

	var conn *websocket.Conn
	var resp *http.Response
	if conn, resp, err = websocket.DefaultDialer.Dial(server.String(), nil); err != nil {
		cmd.Println("failed to dial, response is", resp)
		return
	}
	defer conn.Close()

	done := make(chan struct{}, 1)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		defer close(done)
		for {
			if _, message, msgErr := conn.ReadMessage(); msgErr != nil {
				err = msgErr
				return
			} else {
				cmd.Println(string(message))
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

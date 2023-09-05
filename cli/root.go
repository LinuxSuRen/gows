package cli

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

const quitCommand = "quit-gows"

func NewRootCmd() (cmd *cobra.Command) {
	opt := &rootOption{}
	cmd = &cobra.Command{
		Use: "gows",
		Example: fmt.Sprintf(`gwws ws://your-server.com

You can exit the program by inputting %q and press enter.`, quitCommand),
		Short:   "Echo the message from a websoket server",
		Args:    cobra.MinimumNArgs(1),
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}
	flags := cmd.Flags()
	flags.StringVarP(&opt.cookie, "cookie", "", "", "The cookie to connect to the server")
	return
}

type rootOption struct {
	cookie string
}

func (o *rootOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	o.cookie = os.ExpandEnv(o.cookie)
	return
}

func (o *rootOption) runE(cmd *cobra.Command, args []string) (err error) {
	service := args[0]

	header := http.Header{}
	if o.cookie != "" {
		header.Set("Cookie", o.cookie)
	}

	var conn *websocket.Conn
	var resp *http.Response
	if conn, resp, err = websocket.DefaultDialer.Dial(service, header); err != nil {
		cmd.Println("failed to dial, response is", resp)
		return
	}
	defer conn.Close()

	done := make(chan struct{}, 1)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	conn.WriteMessage(websocket.TextMessage, []byte("\n"))

	go func() {
		defer close(done)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			data := []byte(scanner.Text() + "\n")
			if strings.HasPrefix(string(data), quitCommand) {
				break
			}

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

	// handle the speical command
	go func() {
		for {
			select {
			case <-interrupt:
				// Get more control characters
				// https://www.decisivetactics.com/support/view?article=control-characters
				conn.WriteMessage(websocket.TextMessage, []byte{'\003'})
			}
		}
	}()

	for {
		select {
		case <-done:
		}
		return
	}
}

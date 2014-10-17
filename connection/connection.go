package connection

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/miketheprogrammer/thrust-go/commands"
	. "github.com/miketheprogrammer/thrust-go/common"
)

// Single Connection
var conn net.Conn

type In struct {
	Commands chan *commands.Command
	Quit     chan int
}
type Out struct {
	CommandResponses chan commands.CommandResponse
	Errors           chan error
}

var in In
var out Out

/*
Initializes threads with Channel Structs
Opens Connection
*/
func InitializeThreads(proto, address string) error {
	c, err := net.Dial(proto, address)
	conn = c

	in = In{
		Commands: make(chan *commands.Command),
		Quit:     make(chan int),
	}

	out = Out{
		CommandResponses: make(chan commands.CommandResponse),
		Errors:           make(chan error),
	}

	go Reader(&out, &in)
	go Writer(&out, &in)

	return err
}

func GetConnection() *net.Conn {
	return &conn
}

func GetOutputChannels() *Out {
	return &out
}

func GetInputChannels() *In {
	return &in
}

func GetCommunicationChannels() (*Out, *In) {
	return GetOutputChannels(), GetInputChannels()
}

func Reader(out *Out, in *In) {
	defer conn.Close()

	r := bufio.NewReader(conn)
	for {
		select {
		case quit := <-in.Quit:
			Log.Errorf("Connection Reader Received a Quit message from somewhere ... Exiting Now")
			os.Exit(quit)
		default:
			//a := <-in.Quit
			//fmt.Println(a)
			line, err := r.ReadString(byte('\n'))
			if err != nil {
				fmt.Println(err)
				panic(err)
			}
			Log.Debug("SOCKET::Line", line)
			if !strings.Contains(line, SOCKET_BOUNDARY) {
				response := commands.CommandResponse{}
				json.Unmarshal([]byte(line), &response)
				//Log.Debug(response)
				out.CommandResponses <- response
			}

		}
		time.Sleep(time.Microsecond * 100)

	}

}

func Writer(out *Out, in *In) {
	for {
		select {
		case command := <-in.Commands:
			ActionId += 1
			command.ID = ActionId

			//fmt.Println(command)
			cmd, _ := json.Marshal(command)
			Log.Debug("Writing", string(cmd), "\n", SOCKET_BOUNDARY)

			conn.Write(cmd)
			conn.Write([]byte("\n"))
			conn.Write([]byte(SOCKET_BOUNDARY))
		}
		time.Sleep(time.Microsecond * 100)
	}
}
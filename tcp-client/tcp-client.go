package main

import (
	"bufio"
	"fmt"
	"strings"
	"log"
	"net"
	"os"

	serverinfo "github.com/edibl/go/server-info"
)

var comm_key string

var username string

func dialServer() {
	conn, err := net.Dial(serverinfo.TYPE, serverinfo.HOST+":"+serverinfo.PORT)

	if err != nil {
		log.Fatal("Error in dialing TCP Server: ", err)
	}

	fmt.Print("Welcome, please provide auth credentials in this following format: <username>-<password>\n")

	handleAuth(conn)
	defer handleUserInput(conn)
}

func handleAuth(conn net.Conn) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("<Guest> ")
		text, _ := reader.ReadString('\n')

		text = strings.TrimSpace(text)

		fmt.Fprintf(conn, text+"\n")

		received, _ := bufio.NewReader(conn).ReadString('\n')

		received = strings.TrimSpace(received)

		if received == ">Auth unsuccessful<" {
			fmt.Print("<Server> Authentication is unsuccessful. Please provide credential in this following format: <username>-<password>---\n")
		} else if strings.Contains(received, ">Authenticated") {
			arr := strings.Split(received, "==")
			comm_key = arr[1]
			username = arr[2]
			fmt.Print("<Server> ", strings.TrimSuffix(arr[3], "<"), "\n")
			break
		} else {
			fmt.Print("<Server> ", received, "\n")
		}

		if text == "!quit" {
			os.Exit(0)
		}
	}
}

func handleUserInput(conn net.Conn) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("<%s> ", username)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		fmt.Fprintf(conn, comm_key+"=="+text+"\n")

		received, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("<Server> ", received)

		if text == "!quit" {
			break
		}
	}
}

func main() {
	dialServer()
}
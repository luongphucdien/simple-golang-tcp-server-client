package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	gamedata "github.com/edibl/go/game-data"
	serverinfo "github.com/edibl/go/server-info"
	userinfo "github.com/edibl/go/user-info"
)

var PASSWORD_PATTERN = "^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$"
var USERNAME_PATTERN = "^[a-zA-Z0-9_]*$"
var AUTH_PATTERN = "^[a-zA-Z0-9_]+-.+$"
var COMMAND_PATTERN = "![a-zA-Z0-9-]+$"

var comm_key string

var rand_seed *rand.Rand

var user_info userinfo.User

func tcpListener() {
	listener, err := net.Listen(serverinfo.TYPE, serverinfo.HOST+":"+serverinfo.PORT)
	if err != nil {
		log.Fatal("Error in listening: ", err)
	}

	fmt.Printf("Server is listening at %s:%s\n", serverinfo.HOST, serverinfo.PORT)

	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		log.Fatal("Error in accepting connection: ", err)
	}

	user_info = userinfo.LoadUser("../user-info/user-info.SAVE")

	handleAuth(conn)					// Handle authentication
	defer handleCommunication(conn)		// Handle normal comm. Runs after auth (and others funcs) finish
}

func handleAuth(conn net.Conn) {
	for {
		data, err := bufio.NewReader(conn).ReadString('\n')		// New reader for conn. Read string until newline
		if err != nil {
			log.Fatal("Error in reading data: ", err)
		}

		fmt.Print("<Guest> ", data)
		data = strings.TrimSpace(data)								// Removes leading whitespace and trailing newline, once

		match, _ := regexp.MatchString(AUTH_PATTERN, data)			// Check if received data matches the auth pattern via reg ex
		cmd_match, _ := regexp.MatchString(COMMAND_PATTERN, data)	// Check if received data matches the command pattern via reg ex

		if match {													// If received data match auth pattern, begins checking credentials
			username_password := strings.Split(string(data), "-")
			username := username_password[0]
			password := userinfo.Encode(username_password[1])

			if username == user_info.Username && password == user_info.Password {
				comm_key = keyGenerator()							// Gen a random public key
				conn.Write([]byte(fmt.Sprintf(						// Send public key, username, and full name to client
					">Authenticated==%s==%s==Welcome, %s<\n", 
					comm_key, 
					user_info.Username, 
					user_info.FullName)))
				break
			} else {
				conn.Write([]byte(">Auth unsuccessful<\n"))
			}

		} else if cmd_match {			// If received data matches command pattern, begins to check the commands from client
			if data == "!quit" {		// Command for quitting the program
				conn.Write([]byte("Thank you for coming. See ya later!\n"))
			} else {					// If no commands are recognized, then the command is unknown
				conn.Write([]byte("Unknown command\n"))
			}
		} else {						// If received data neither matches auth nor command pattern, return auth unsuccessful
			conn.Write([]byte(">Auth unsuccessful<\n"))
		}
	}
}

func handleCommunication(conn net.Conn) {		// Runs after everything else is done
	for {
		received, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatal("Error in reading data: ", err)
		}

		fmt.Printf("<%s> %s", user_info.Username, received)

		received = strings.TrimSpace(received)
		arr := strings.Split(received, "==")						// Split received data at == to separate key and actual data
		received = arr[1]											// Actual data at index 1
		match, _ := regexp.MatchString(COMMAND_PATTERN, received)	// Match actual data for command pattern

		if match {
			if received == "!help" {
				conn.Write([]byte(comm_key + "==" + "!game-guessing: Guessing game | !game-hangman: Hangman game | !quit: Close the connection\n"))
			} else if received == "!game-guessing" {
				handleGuessingGame(conn)
			} else if received == "!game-hangman" {
				handleHangmanGame(conn)
			} else if received == "!quit" {
				conn.Write([]byte("Thank you for coming. See ya next time\n"))
				conn.Close()
			} else {
				conn.Write([]byte(comm_key + "==" + "Unknown command\n"))
			}
		} else {		// If data is not a command, then it's just a normal message
			conn.Write([]byte(fmt.Sprintf("%s==Message received\n", comm_key)))
		}
	}
}

func handleGuessingGame(conn net.Conn) {
	max := 1000
	min := 0
	number := rand_seed.Intn(max-min+1) + min	// Random in range determined above

	isWin := false

	conn.Write([]byte(comm_key + "==" + "Please guess the number\n"))

Rerun:
	for {
		received, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatal("Error in reading data: ", err)
		}

		fmt.Printf("<%s> %s", user_info.Username, received)

		received = strings.TrimSpace(received)
		arr := strings.Split(received, "==")
		received = arr[1]

		match, _ := regexp.MatchString(COMMAND_PATTERN, received)
		if match {
			if received == "!end" {
				conn.Write([]byte("Back to home\n"))
				break
			} else if received == "!continue" {
				handleGuessingGame(conn)
				break
			} else {
				conn.Write([]byte("Unknown command\n"))
			}
		} else {
			user_number, err := strconv.ParseInt(received, 10, 64)		// Parse data to int
			if err != nil {												// If parsing is unsuccessful, then it's not a number
				conn.Write([]byte(comm_key + "==" + "Your input must be a number\n"))
				goto Rerun												// Return back to the label Rerun
			}

			if user_number > int64(number) && !isWin {
				conn.Write([]byte(comm_key + "==" + "The number is smaller\n"))
			} else if user_number < int64(number) && !isWin {
				conn.Write([]byte(comm_key + "==" + "The number is larger\n"))
			} else {		// isWin keeps the player in this condition if they've won the game, preventing the execution of the above conditions
				isWin = true
				conn.Write([]byte(comm_key + "==" + "You've guessed the number! To continue playing, type !continue. To end the game, type !end\n"))
			}
		}
	}
}

func handleHangmanGame(conn net.Conn) {
	var words_list gamedata.WordList
	lives := 8	// Traditional lives number of a Hangman game: 1. Post | 2.Head | 3.Body | 4,5.Two arms | 6,7.Two legs | 8.The rope

	go func() {
		json_bytes, err := os.ReadFile("../game-data/hangman.json")
		if err != nil {
			log.Fatal("Error in reading file: ", err)
		}

		json.Unmarshal(json_bytes, &words_list)
	}()

CheckIfEmpty:
	if reflect.ValueOf(words_list).IsZero() {
		fmt.Println("No list")
		goto CheckIfEmpty
	} else {
		fmt.Println("Yes list")
	}

	min := 0
	max := len(words_list.Words)
	position := rand_seed.Intn(max-min) + min

	word_obj := words_list.Words[position]			// Get a random word with its des in the slice
	var blankify_arr []string
	for range word_obj.Word {
		blankify_arr = append(blankify_arr, "_ ")	// Replace characters in that word with underscores
	}
	blankify_word := strings.Join(blankify_arr, "")	// Joins the underscore array into a string

	conn.Write([]byte(
		comm_key +
			"==" +
			"You have " +
			strconv.FormatInt(int64(lives), 10) +
			" lives left. Good luck ||| " +
			blankify_word +
			"||| " +
			word_obj.Description +
			"\n"))

	for {
		received, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatal("Error in reading data: ", err)
		}

		fmt.Printf("<%s> %s", user_info.Username, received)

		received = strings.TrimSpace(received)
		arr := strings.Split(received, "==")
		received = arr[1]

		match, _ := regexp.MatchString(COMMAND_PATTERN, received)
		if match {
			if received == "!end" {
				conn.Write([]byte("Back to home\n"))
				break
			} else if received == "!continue" {
				handleHangmanGame(conn)
				break
			} else if received == "!hint" {
				conn.Write([]byte(blankify_word + "||| " + word_obj.Description + "\n"))
			} else {
				conn.Write([]byte(comm_key + "==" + "Unknown command\n"))
			}
		} else {
			if lives > 0 {				// If there're still lives left, the game continues
				if len(received) > 1 {	// Only allow player to enter a single character
					conn.Write([]byte(comm_key + "==" + "Please enter a single character only\n"))
				} else {
					if strings.Contains(blankify_word, "_") {										// Check if the word still has blank spaces
						if strings.Contains(word_obj.Word, received) {								// Check if the received char exists in the word
							regexify_received := regexp.MustCompile(received)						// Compile the char into reg ex
							indexes := regexify_received.FindAllIndex([]byte(word_obj.Word), -1)	// Find all occurences of that char in the word

							for i := range indexes {
								copy(blankify_arr[indexes[i][0]:], []string{received + " "})		// Replace the char into the correct pos in the underscore array
							}

							blankify_word = strings.Join(blankify_arr, "")
							conn.Write([]byte(comm_key + "==" + blankify_word + "\n"))
						} else {						// If the char does not exist in the word, run this condition
							if len(received) == 0 {		// If the received data is empty, just print the incompleted word again
								conn.Write([]byte(comm_key + "==" + blankify_word + "\n"))
							} else {
								lives--
								conn.Write([]byte(
									comm_key +
										"==" +
										"The word does not contain the character, You have " +
										strconv.FormatInt(int64(lives), 10) +
										" lives left\n"))
							}
						}
					} else {		// Player will be stuck here if the word is completed since there is no underscore left
						conn.Write([]byte(comm_key + "==" + "You've won the game! To continue playing, type !continue. To end the game, type !end\n"))
					}
				}
			} else {			// Player will be stuck here if their lives are depleted
				conn.Write([]byte(comm_key + "==" + "You've lost the game! To restart, type !continue. To end the game, type !end\n"))
			}
		}
	}
}

func keyGenerator() string {
	rand_seed = rand.New(rand.NewSource(time.Now().UnixNano()))
	min := 100
	max := 1000
	return strconv.FormatInt(int64(rand_seed.Intn(max-min+1)+min), 10)	// Gen a random 3-digit public key
}

func main() {
	tcpListener()
}
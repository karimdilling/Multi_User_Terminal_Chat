package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	user := getUsername()

	conn := connectToServer(&user.username)

	app := tview.NewApplication()
	layout, chat, clientsOnline := setupUI(app, conn, &user)

	serverSideDisconnect := false
	go readMessages(conn, app, chat, clientsOnline, &serverSideDisconnect, &user)

	if err := app.SetRoot(layout, true).Run(); err != nil {
		log.Fatalf("Encountered error: %v\n", err)
	}
	clearTerminal()
	if serverSideDisconnect {
		fmt.Fprintln(os.Stderr, "Connection closed by foreign host.")
	}
}

func getUsername() User {
	fmt.Println("Please enter a username!")
	username := "User"
	fmt.Scanf("%s", &username)

	return User{
		username: username,
		color:    colors[rand.Intn(len(colors))],
	}
}

func connectToServer(username *string) net.Conn {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not connect to server.")
		os.Exit(1)
	}

	msg := Message{Username: *username, MsgType: ClientConnected}
	msgJSON, err := json.Marshal(&msg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not recognize Username. Disconnected from the server.")
		os.Exit(1)
	}
	conn.Write(msgJSON)

	return conn
}

func setupUI(app *tview.Application, conn net.Conn, user *User) (*tview.Flex, *tview.TextView, *tview.TextView) {
	textview := tview.NewTextView()
	createChatWindow(app, textview, user)

	inputField := tview.NewInputField()
	createInputField(inputField, textview, conn, app, user)

	textviewClientsOnline := tview.NewTextView().SetChangedFunc(func() {
		app.Draw()
	}).
		ScrollToEnd().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)
	textviewClientsOnline.SetBorder(true).SetTitle(" Online ")

	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(textview, 0, 9, false).
			AddItem(inputField, 0, 1, true), 0, 3, true).
		AddItem(textviewClientsOnline, 0, 1, false)

	return flex, textview, textviewClientsOnline
}

func readMessages(conn net.Conn, app *tview.Application, textview *tview.TextView, textviewClientsOnline *tview.TextView, serverSideDisconnect *bool, thisClientUser *User) {
	usernameColor := make(map[string]string)
	rdbuff := make([]byte, 250)
	for {
		n, err := conn.Read(rdbuff)
		if err != nil {
			conn.Close()
			*serverSideDisconnect = true
			app.Stop()
		}
		var msg Message
		err = json.Unmarshal(rdbuff[0:n], &msg)
		if err != nil {
			continue
		}
		switch msg.MsgType {
		case ClientMessage:
			textview.Write([]byte(fmt.Sprintf("[%s]%s[-]: %s", usernameColor[msg.Username], msg.Username, msg.Content)))
		case ClientConnected, ClientDisconnected:
			textview.Write([]byte("[::r]" + msg.Content + "[-:-:-:-]")) // [::r] inverts colors and [-:-:-:-] resets formatting

			textviewClientsOnline.Clear()
			usernameListAsString := ""
			slices.Sort(msg.ClientList)
			for _, username := range msg.ClientList {
				_, exists := usernameColor[username]
				if !exists && username != thisClientUser.username {
					usernameColor[username] = colors[rand.Intn(len(colors))]
				} else if username == thisClientUser.username {
					usernameColor[username] = thisClientUser.color
				}
				usernameListAsString += fmt.Sprintf("[%s]%s[-]\n", usernameColor[username], username)
			}
			textviewClientsOnline.Write([]byte(usernameListAsString))
		}
	}
}

func createChatWindow(app *tview.Application, textview *tview.TextView, thisClientUser *User) {
	textview.SetChangedFunc(func() {
		app.Draw()
	}).
		ScrollToEnd().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetMaxLines(10000)
	textview.SetBorder(true).
		SetTitle(fmt.Sprintf(" Chat (online as [%s]%s[-]) ", thisClientUser.color, thisClientUser.username))
}

var colors = [...]string{"green", "blue", "red", "purple", "aqua", "yellow", "pink", "navy"}

type User struct {
	username string
	color    string
}

type MessageType int

const (
	ClientDisconnected MessageType = iota
	ClientConnected
	ClientMessage
)

type Message struct {
	Username   string      `json:"username"`
	MsgType    MessageType `json:"msg_type"`
	Content    string      `json:"content"`
	ClientList []string    `json:"client_list"`
}

func createInputField(inputField *tview.InputField, textview *tview.TextView, conn net.Conn, app *tview.Application, thisClientUser *User) {
	inputField.SetBorder(true).
		SetBackgroundColor(tcell.ColorBlack)
	inputField.
		SetPlaceholder("Type your message here...").
		SetPlaceholderTextColor(tcell.ColorGrey).
		SetPlaceholderStyle(tcell.StyleDefault.Background(tcell.ColorBlack)).
		SetAcceptanceFunc(tview.InputFieldMaxLength(200)).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEnter:
				input := inputField.GetText()
				input = strings.TrimSpace(input)
				var msg Message
				msg.Username = thisClientUser.username
				msg.Content = input
				msg.MsgType = ClientMessage
				switch msg.Content {
				case "":
					break
				case ".quit":
					app.Stop()
				default:
					msg.Content = tview.Escape(msg.Content)
					msg.Content += "\n"
					textview.Write([]byte(fmt.Sprintf("[%s]%s[-]: %s", thisClientUser.color, thisClientUser.username, msg.Content)))
					msgJSON, _ := json.Marshal(msg)
					conn.Write([]byte(msgJSON))
					inputField.SetText("")
				}
			}
		})
}

func clearTerminal() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

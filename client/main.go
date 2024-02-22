package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	fmt.Println("Please enter a username!")
	username := "User"
	fmt.Scanf("%s", &username)

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not connect to server.")
		return
	}

	app := tview.NewApplication()

	textview := tview.NewTextView()
	createChatWindow(app, textview)

	inputField := tview.NewInputField()
	createInputField(inputField, textview, conn, app, username)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textview, 0, 9, false).
		AddItem(inputField, 0, 1, true)

	serverSideDisconnect := false
	go readMessages(conn, app, textview, &serverSideDisconnect)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Fatalf("Encountered error: %v\n", err)
	}
	clearTerminal()
	if serverSideDisconnect {
		fmt.Fprintln(os.Stderr, "Connection closed by foreign host.")
	}
}

func readMessages(conn net.Conn, app *tview.Application, textview *tview.TextView, serverSideDisconnect *bool) {
	rdbuff := make([]byte, 80)
	for {
		n, err := conn.Read(rdbuff)
		if err != nil {
			conn.Close()
			*serverSideDisconnect = true
			app.Stop()
		}
		textview.Write(rdbuff[0:n])
	}
}

func createChatWindow(app *tview.Application, textview *tview.TextView) {
	textview.SetChangedFunc(func() {
		app.Draw()
	}).
		ScrollToEnd().
		SetTextAlign(tview.AlignLeft)
	textview.SetBorder(true).
		SetTitle(" Chat ")
}

func createInputField(inputField *tview.InputField, textview *tview.TextView, conn net.Conn, app *tview.Application, username string) {
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
				if input == ".quit" {
					app.Stop()
				}
				input += "\n"
				if input != "\n" {
					textview.Write([]byte("You: " + input))
					conn.Write([]byte(username + ": " + input))
				}
				inputField.SetText("")
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

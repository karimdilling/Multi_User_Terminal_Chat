package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not connect to server.")
		return
	}

	app := tview.NewApplication()

	textview := tview.NewTextView()
	createChatWindow(app, textview)

	inputField := tview.NewInputField()
	createInputField(inputField, textview, conn)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textview, 0, 9, false).
		AddItem(inputField, 0, 1, true)

	go readMessages(conn, textview)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Fatalf("Encountered error: %v\n", err)
	}
}

func readMessages(conn net.Conn, textview *tview.TextView) {
	rdbuff := make([]byte, 80)
	for {
		n, err := conn.Read(rdbuff)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Connection closed by foreign host.")
			conn.Close()
			os.Exit(1)
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

func createInputField(inputField *tview.InputField, textview *tview.TextView, conn net.Conn) {
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
				input += "\n"
				if input != "\n" {
					textview.Write([]byte("You: " + input))
					conn.Write([]byte(input))
				}
				inputField.SetText("")
			}
		})
}

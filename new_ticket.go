package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gonutz/wui/v2"
)

func main() {
	font, _ := wui.NewFont(wui.FontDesc{
		Name:   "Tahoma",
		Height: -13,
	})
	bold, _ := wui.NewFont(wui.FontDesc{
		Name:   "Tahoma",
		Height: -13,
		Bold:   true,
	})
	window := wui.NewWindow()
	window.SetHasMinButton(false)
	window.SetHasMaxButton(false)
	window.SetResizable(false)
	window.SetFont(font)
	window.SetTitle("Create new Ticket")
	window.SetInnerSize(600, 600)
	title := wui.NewEditLine()
	title.SetText("Title")
	title.SetBounds(10, 10, 580, 25)
	title.SetFont(bold)
	window.Add(title)
	desc := wui.NewTextEdit()
	desc.SetWordWrap(true)
	desc.SetBounds(10, 50, 580, 500)
	desc.SetText("Description\r\n...")
	window.Add(desc)
	ok := wui.NewButton()
	ok.SetBounds(260, 560, 80, 30)
	ok.SetText("OK")
	ok.SetOnClick(func() {
		nextNumPath := filepath.Join(
			filepath.Dir(os.Args[0]),
			"next_ticket_number.txt",
		)
		data, err := ioutil.ReadFile(nextNumPath)
		if err != nil {
			data = []byte("1")
		}
		n, err := strconv.Atoi(string(data))
		if err != nil {
			wui.MessageBoxError(
				"Error",
				"Invalid number in next ticket number file: "+err.Error(),
			)
			return
		}
		text := title.Text() + "\n\n" + desc.Text()
		text = winLines(text)
		err = ioutil.WriteFile(
			fmt.Sprintf("%d.txt", n),
			[]byte(text),
			0777,
		)
		if err != nil {
			wui.MessageBoxError(
				"Error",
				"Unable to write new ticket file: "+err.Error(),
			)
			return
		}
		err = ioutil.WriteFile(nextNumPath, []byte(strconv.Itoa(n+1)), 0777)
		if err != nil {
			wui.MessageBoxError(
				"Error",
				"Unable to update next ticket file: "+err.Error(),
			)
			return
		}
		window.Close()
	})
	window.Add(ok)
	window.SetOnShow(func() {
		title.Focus()
	})
	window.Show()
}

func winLines(s string) string {
	s = strings.Replace(s, "\r\n", "\n", -1)
	s = strings.Replace(s, "\n", "\r\n", -1)
	return s
}

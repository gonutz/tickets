//+build ignore

package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/gonutz/w32"
	"github.com/gonutz/wui"
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
	window.SetFont(font)
	window.SetClientSize(700, 500)
	window.SetTitle("Tickets")
	scrollPos := 0
	scroll := func(delta float64) {
		d := round(delta * 50)
		if scrollPos+d <= 0 {
			window.Scroll(0, d)
			scrollPos += d
		}
	}
	window.SetOnMouseWheel(func(x, y int, delta float64) {
		scroll(delta)
	})
	files, err := ioutil.ReadDir(".")
	sort.Sort(byNumber(files))
	if err != nil {
		wui.MessageBoxError(window, "Error", "Unable to read ticket directory: "+err.Error())
	}
	y := 10
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".txt") {
			number := strings.TrimSuffix(file.Name(), ".txt")
			_, err := strconv.Atoi(number)
			if err == nil {
				data, err := ioutil.ReadFile(file.Name())
				if err == nil {
					s := string(data)
					i := strings.Index(s, "\n")
					if i == -1 {
						i = len(s)
					}
					firstLine := strings.TrimSuffix(s[:i], "\r")

					b := wui.NewButton()
					b.SetText(number)
					b.SetFont(bold)
					b.SetBounds(10, y, 40, 20)
					ticket := file.Name()
					b.SetOnClick(func() {
						output, err := exec.Command("cmd", "/C", "start", ticket).CombinedOutput()
						if err != nil {
							wui.MessageBoxError(
								window,
								"Error",
								winLines("Unable to open ticket file: "+err.Error()+"\n"+string(output)),
							)
						}
					})
					window.Add(b)

					title := wui.NewLabel()
					title.SetBounds(60, y, 600, 20)
					title.SetText(firstLine)
					window.Add(title)

					x := wui.NewButton()
					x.SetText("x")
					x.SetFont(bold)
					x.SetBounds(670, y, 20, 20)
					window.Add(x)
					x.SetOnClick(func() {
						if !wui.MessageBoxYesNo(
							window,
							"Delete Ticket?",
							"Really delete ticket "+number+"?") {
							return
						}
						if err := os.Remove(ticket); err != nil {
							wui.MessageBoxError(
								window,
								"Error",
								winLines("Unable to delete ticket: "+err.Error()),
							)
						} else {
							x.SetEnabled(false)
							b.SetEnabled(false)
							title.SetEnabled(false)
						}
					})

					y += 20
				}
			}
		}
	}
	window.SetShortcut(wui.ShortcutKeys{Key: w32.VK_DOWN}, func() {
		scroll(-0.25)
	})
	window.SetShortcut(wui.ShortcutKeys{Key: w32.VK_UP}, func() {
		scroll(0.25)
	})
	window.SetShortcut(wui.ShortcutKeys{Key: w32.VK_NEXT}, func() {
		scroll(-9)
	})
	window.SetShortcut(wui.ShortcutKeys{Key: w32.VK_PRIOR}, func() {
		scroll(9)
	})
	window.SetShortcut(wui.ShortcutKeys{Key: w32.VK_ESCAPE}, window.Close)
	window.Show()
}

type byNumber []os.FileInfo

func (x byNumber) Len() int {
	return len(x)
}

func (x byNumber) Less(i, j int) bool {
	a, _ := strconv.Atoi(strings.TrimSuffix(x[i].Name(), ".txt"))
	b, _ := strconv.Atoi(strings.TrimSuffix(x[j].Name(), ".txt"))
	return a < b
}

func (x byNumber) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func winLines(s string) string {
	s = strings.Replace(s, "\r\n", "\n", -1)
	s = strings.Replace(s, "\n", "\r\n", -1)
	return s
}

func round(x float64) int {
	if x < 0 {
		return int(x - 0.5)
	}
	return int(x + 0.5)
}

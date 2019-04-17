//+build ignore

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"unicode"

	"github.com/gonutz/w32"
	"github.com/gonutz/wui"
)

type appSettings struct {
	MonitorX, MonitorY  int
	X, Y, Width, Height int
}

func main() {
	var settings appSettings
	settingsPath := filepath.Join(os.Getenv("APPDATA"), "view_tickets.set")
	if data, err := ioutil.ReadFile(settingsPath); err == nil {
		json.Unmarshal(data, &settings)
	}
	defer func() {
		data, err := json.Marshal(&settings)
		if err == nil {
			ioutil.WriteFile(settingsPath, data, 0666)
		}
	}()
	// The settings store the last top-left corner of the monitor which the
	// window was previously on. On the next program start this monitor might be
	// unplugged. Since we do not want to show the window on a non-existing
	// monitor (this would put the window off-screen) we check if there
	// currently is a monitor that has the same top-left corner and put our
	// window on it.
	var monitors []w32.HMONITOR
	cb := syscall.NewCallback(func(m w32.HMONITOR, hdc w32.HDC, r *w32.RECT, l w32.LPARAM) uintptr {
		monitors = append(monitors, m)
		return 1
	})
	w32.EnumDisplayMonitors(0, nil, cb, 0)
	var found bool
	for _, monitor := range monitors {
		var info w32.MONITORINFO
		if w32.GetMonitorInfo(monitor, &info) &&
			int(info.RcWork.Left) == settings.MonitorX &&
			int(info.RcWork.Top) == settings.MonitorY {
			found = true
		}
	}
	if !found {
		// if the last monitor is not available anymore, display top-left
		settings.X, settings.Y = 0, 0
	}

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
	window.SetTitle("Tickets")
	window.SetClientSize(700, 500)
	window.SetBounds(settings.X, settings.Y, settings.Width, settings.Height)
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
		wui.MessageBoxError("Error", "Unable to read ticket directory: "+err.Error())
	}

	type ticket struct {
		path    string
		number  string
		title   string
		content string
	}
	var tickets []ticket
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
					tickets = append(tickets, ticket{
						path:    file.Name(),
						number:  number,
						title:   firstLine,
						content: s,
					})
				}
			}
		}
	}

	searchText := wui.NewEditLine()
	window.Add(searchText)
	searchText.SetBounds(10, 10, 595, 20)
	search := wui.NewButton()
	window.Add(search)
	search.SetBounds(610, 10, 80, 20)
	search.SetText("Search")

	var titles []*wui.Label

	y := 40
	for i := range tickets {
		ticket := tickets[i]
		b := wui.NewButton()
		b.SetText(ticket.number)
		b.SetFont(bold)
		b.SetBounds(10, y, 40, 20)
		b.SetOnClick(func() {
			output, err := exec.Command("cmd", "/C", "start", ticket.path).CombinedOutput()
			if err != nil {
				wui.MessageBoxError(
					"Error",
					winLines("Unable to open ticket file: "+err.Error()+"\n"+string(output)),
				)
			}
		})
		window.Add(b)

		title := wui.NewLabel()
		title.SetBounds(60, y, 600, 20)
		title.SetText(ticket.title)
		window.Add(title)
		titles = append(titles, title)

		x := wui.NewButton()
		x.SetText("x")
		x.SetFont(bold)
		x.SetBounds(670, y, 20, 20)
		window.Add(x)
		x.SetOnClick(func() {
			if !wui.MessageBoxYesNo(
				"Delete Ticket?",
				"Really delete ticket "+ticket.number+"?") {
				return
			}
			if err := os.Remove(ticket.path); err != nil {
				wui.MessageBoxError(
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

	search.SetOnClick(func() {
		terms := searchText.Text()
		if terms == "" {
			for i := range titles {
				titles[i].SetFont(font)
			}
			return
		}
		terms = strings.ToLower(terms)
		terms = strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				return r
			}
			return ' '
		}, terms)
		list := strings.Split(terms, " ")
		var words []string
		for _, term := range list {
			if term != "" {
				words = append(words, term)
			}
		}
		searchText.SetText(strings.Join(words, " "))
		for i, ticket := range tickets {
			text := strings.Map(func(r rune) rune {
				if unicode.IsLetter(r) || unicode.IsDigit(r) {
					return unicode.ToLower(r)
				}
				return -1
			}, ticket.content)
			found := false
			for _, word := range words {
				if strings.Contains(text, word) {
					found = true
					break
				}
			}
			if found {
				titles[i].SetFont(bold)
			} else {
				titles[i].SetFont(font)
			}
		}
	})

	window.SetShortcut(wui.ShortcutKeys{Key: w32.VK_RETURN}, func() {
		if searchText.HasFocus() {
			search.OnClick()()
		}
	})
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
	window.SetOnShow(func() {
		searchText.Focus()
	})
	window.SetOnClose(func() {
		settings.X, settings.Y, settings.Width, settings.Height = window.Bounds()
		monitor := window.Monitor()
		if monitor != 0 {
			var info w32.MONITORINFO
			if w32.GetMonitorInfo(monitor, &info) {
				settings.MonitorX = int(info.RcWork.Left)
				settings.MonitorY = int(info.RcWork.Top)
			}
		}
	})
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

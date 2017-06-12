package frontend

import (
	"fmt"
	"strings"
	"io/ioutil"
	"github.com/gdamore/tcell"
	"github.com/fatih/color"
)

type AsciiScreen struct {
	Content [][]rune
	Width int
	Height int
}

func NewAsciiScreen() *AsciiScreen {
	screenWidth := 80
	screenHeight := 24

	// Initialize to a collection of spaces
	content := make([][]rune, screenHeight)
	for i := 0; i < screenHeight; i++ {
		content[i] = make([]rune, screenWidth)
		for j := 0; j < screenWidth; j++ {
			content[i][j] = ' '
		}
	}

	return &AsciiScreen{
		Content: content,
		Width: screenWidth,
		Height: screenHeight,
	}
}

func (screen AsciiScreen) Init() error {
	return nil
}
func (screen AsciiScreen) Fini() {}
func (screen AsciiScreen) Clear() {
}
func (screen AsciiScreen) Fill(r rune, s tcell.Style) {
}
func (screen AsciiScreen) SetStyle(style tcell.Style) {}
func (screen AsciiScreen) ShowCursor(x int, y int) {}
func (screen AsciiScreen) HideCursor() {}
func (screen AsciiScreen) Size() (i int, j int) {
	return screen.Width, screen.Height
}
func (screen AsciiScreen) PollEvent() tcell.Event {
	neverResolves := make(chan tcell.Event, 1)
	return <-neverResolves
}
func (screen AsciiScreen) PostEvent(ev tcell.Event) error { return nil }
func (screen AsciiScreen) PostEventWait(ev tcell.Event) {}

func (screen AsciiScreen) HasMouse() bool {
	return false
}
func (screen AsciiScreen) EnableMouse() {}
func (screen AsciiScreen) DisableMouse() {}
func (screen AsciiScreen) Colors() int {
	return 0
}
func (screen AsciiScreen) Show() {
}
func (screen AsciiScreen) CharacterSet() string {
	return "" // TODO: is this the right value?
}
func (screen AsciiScreen) RegisterRuneFallback(r rune, substr string) {
}
func (screen AsciiScreen) UnregisterRuneFallback(r rune) {
}
func (screen AsciiScreen) CanDisplay(r rune, checksFallback bool) bool {
	return true
}
func (screen AsciiScreen) Resize(a, b, c, d int) {}
func (screen AsciiScreen) HasKey(key tcell.Key) bool {
	return true
}


func (screen AsciiScreen) Sync() {
}
func (screen AsciiScreen) GetContent(x int, y int) (mainc rune, combc []rune, style tcell.Style, width int) {
	return screen.Content[y][x], []rune{}, tcell.StyleDefault, 1
}
func (screen AsciiScreen) SetContent(x int, y int, mainc rune, combc []rune, style tcell.Style) {
	screen.Content[y][x] = mainc
}
func (screen AsciiScreen) SetCell(x int, y int, style tcell.Style, ch ...rune) {
	screen.Content[y][x] = ch[0]
}

func (screen AsciiScreen) Compare(filename string) (string, bool) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err.Error(), false
	}
	lines := strings.Split(string(data), "\n")

	var total string
	var ok bool = true
	for i := 0; i < screen.Height; i++ {
		var line string
		for j := 0; j < screen.Width; j++ {
			if screen.Content[i][j] == ' ' {
				line += " "
			} else if i < len(lines) && j < len(lines[i]) {
				actual := rune(lines[i][j])
				test := rune(screen.Content[i][j])
				if actual == test {
					line += color.GreenString(string(actual))
				} else {
					line += color.RedString("X")
					ok = false
				}
			} else {
				line += color.RedString(".")
			}
		}
		total += fmt.Sprintf("%s\n", line)
	}
	return total, ok
}

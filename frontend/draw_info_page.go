package frontend

import (
	"fmt"
	"github.com/1egoman/slick/version"
	"strings"
)

var INFO_PAGE = fmt.Sprintf(`
Slick %s

By Ryan Gaus and contributors.
Open source and MIT licensed.
type '/q' to exit
type '/connect "slack token"' to connect to a slack team
type '/help' to learn more
(no slack teams are currently connected.)
`, version.Version())

// Render above launch info when the user isn't conencted to an active connection.
func (term *TerminalDisplay) DrawInfoPage() {
	width, height := term.screen.Size()

	rows := strings.Split(INFO_PAGE, "\n")
	firstRowPosition := (height - len(rows)) / 2

	for index, row := range rows {
		xPos := (width - len(row)) / 2
		term.WriteText(xPos, firstRowPosition+index, row)
	}
}

package modal

type Modal struct {
	Title string
	Body string
	Editable bool
	CursorPosition int

	ScrollPosition int
}

// Called when the modal is originally opened, to reset the state from the previous use.
func (m *Modal) Reset() {
	m.ScrollPosition = 0
	m.Editable = false
}

func (m *Modal) SetContent(data string) {
	m.Body = data
	m.CursorPosition = len(data) - 1
}

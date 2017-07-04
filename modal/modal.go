package modal

type Modal struct {
	Title string
	Body string

	ScrollPosition int
}

// Called when the modal is originally opened, to reset the state from the previous use.
func (m *Modal) Reset() {
	m.ScrollPosition = 0
}

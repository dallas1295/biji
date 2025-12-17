package tui

func (m model) View() string {
	s := "Notes\n\n"

	for i, n := range m.notes {
		if i == m.listIndex {
			s += ">> "
		} else {
			s += "   "
		}

		s += n.Name
	}

	return s
}

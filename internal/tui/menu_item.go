package tui

/*
MENU ITEM
*/
type item struct {
	title, desc, key string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }
func (i item) Key() string         { return i.key }

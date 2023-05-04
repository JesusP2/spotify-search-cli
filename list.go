package main

import "github.com/charmbracelet/bubbles/list"

type Item struct {
	title, desc string
}

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.desc }
func (i Item) FilterValue() string { return i.title }

func NewListModel(title string, items []list.Item) list.Model {
	searchResults := list.New(items, list.NewDefaultDelegate(), 0, 0)
	searchResults.Title = title
	return searchResults
}

package main

import (
  "strings"
)

// A struct that wraps a collection of string items and norma items, and sorts both based on the
// closeness of each of `StringItems` to `Needle`.
type FuzzySorter struct {
  Visible     bool

	Items       []interface{}
	StringItems []string
	Needle      string
  OnSelected  func(*State)

  SelectedItem int
  BottomItem int
}

func (p FuzzySorter) Len() int {
	return len(p.StringItems)
}

func (p FuzzySorter) Less(i, j int) bool {
  iItem := p.StringItems[i]
  jItem := p.StringItems[j]
  return strings.Index(iItem, p.Needle) > strings.Index(jItem, p.Needle)
}

func (p FuzzySorter) Swap(i, j int) {
	p.Items[i], p.Items[j] = p.Items[j], p.Items[i]
	p.StringItems[i], p.StringItems[j] = p.StringItems[j], p.StringItems[i]
}

// Show the fuzzy picker
func (p *FuzzySorter) Show(callbackOnSelected func(*State)) {
  p.Visible = true
  p.OnSelected = callbackOnSelected
  p.SelectedItem = 0
  p.BottomItem = 0
}

// Hide the fuzzy picker and reset to initial state
func (p *FuzzySorter) Hide() {
  p.Visible = false
  p.Items = []interface{}{}
  p.StringItems = []string{}
}

type FuzzyPickerConnectionChannelItem struct {
	Channel    string
	Connection string
}

type FuzzyPickerSlashCommandItem struct {
  Name string
  Description string
  Type CommandType
  Permutations []string
  Arguments string
}

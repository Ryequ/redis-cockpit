package command

import "testing"

func TestHistoryAddAndList(t *testing.T) {
	h := NewHistory(3)
	h.Add("GET one")
	h.Add("SET one 1")
	h.Add("DEL one")

	all := h.All()
	if len(all) != 3 || all[0] != "GET one" || all[2] != "DEL one" {
		t.Fatalf("history = %#v", all)
	}
}

func TestHistoryEnforcesLimitAndMovesRepeatedCommandToEnd(t *testing.T) {
	h := NewHistory(2)
	h.Add("GET one")
	h.Add("GET two")
	h.Add("GET one")

	all := h.All()
	if len(all) != 2 || all[0] != "GET two" || all[1] != "GET one" {
		t.Fatalf("history = %#v, want [GET two GET one]", all)
	}
}

package lexer

import "testing"

func TestRuneStringMatch(t *testing.T) {
	rs := NewRuneString("THIS IS A TEST MATCH")
	if rs.MatchesFrom(0, "TEST") {
		t.Error("erroneous postitive match at TEST from 0")
	}
	if !rs.MatchesFrom(0, "THIS") {
		t.Error("erroneous negative match at TEST from 0")
	}
	if !rs.MatchesFrom(5, "IS") {
		t.Error("erroneous negative match at IS from 5")
	}
}

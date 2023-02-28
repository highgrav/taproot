package lexer

type RuneString struct {
	Runes      []rune
	Length     int32
	CurrentPos int32
}

func NewRuneString(str string) *RuneString {
	rs := &RuneString{
		Runes:      []rune(str),
		CurrentPos: 0,
	}
	rs.Length = int32(len(rs.Runes))
	return rs
}

func (s *RuneString) MatchesFrom(from int32, toMatch string) bool {
	matchRunes := NewRuneString(toMatch)
	for x := int32(0); x < matchRunes.Length; x++ {
		if s.Get(from+x) != matchRunes.Get(x) {
			return false
		}
	}
	return true
}

func (s *RuneString) Get(pos int32) rune {
	if pos >= s.Length {
		return rune(0)
	}
	return s.Runes[pos]
}

func (s *RuneString) Next() rune {
	s.CurrentPos++
	return s.Get(s.CurrentPos - 1)
}

func (s *RuneString) Peek() rune {
	if s.CurrentPos < s.Length {
		return s.Get(s.CurrentPos)
	}
	return rune(0x0)
}

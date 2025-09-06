package main

/*
 * Export everything for Compiler
 */
type AtomPosition struct {
	LineStart int
	LineEnded int
	ColmStart int
	ColmEnded int
}

func (p *AtomPosition) Merge(other AtomPosition) AtomPosition {
	return AtomPosition{
		LineStart: p.LineStart,
		LineEnded: other.LineEnded,
		ColmStart: p.ColmStart,
		ColmEnded: other.ColmEnded,
	}
}

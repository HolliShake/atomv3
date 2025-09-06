package main

/*
 * Export everything for Compiler
 */
type Position struct {
	LineStart int
	LineEnded int
	ColmStart int
	ColmEnded int
}

func (p *Position) Merge(other Position) Position {
	return Position{
		LineStart: p.LineStart,
		LineEnded: other.LineEnded,
		ColmStart: p.ColmStart,
		ColmEnded: other.ColmEnded,
	}
}

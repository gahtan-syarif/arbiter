package games

import (
	"errors"
	"fmt"
	"math/bits"
	"strconv"
	"strings"
)

func HasAtaxxGameEnded(fen string, moves []string) Result {
	var pos Position
	pos.SetFen(fen)

	for _, mov := range moves {
		move, _ := NewMove(mov)
		pos.MakeMove(*move)
	}

	stm := pos.turn
	xtm := pos.turn ^ 1

	// Halfmove clock
	if pos.halfmoves >= 100 {
		return Draw
	}

	// No pieces left
	if pos.pieces[stm].Data == 0 {
		return XtmWins
	} else if pos.pieces[xtm].Data == 0 {
		return StmWins
	}

	// No moves left
	empty := all ^ pos.pieces[0].Data ^ pos.pieces[1].Data ^ pos.gaps.Data
	both := Bitboard{pos.pieces[0].Data | pos.pieces[1].Data}
	if (both.Singles().Data|both.Doubles().Data)&empty == 0 {
		stm_n := bits.OnesCount64(pos.pieces[stm].Data)
		xtm_n := bits.OnesCount64(pos.pieces[xtm].Data)

		if stm_n > xtm_n {
			return StmWins
		} else if xtm_n > stm_n {
			return XtmWins
		} else {
			return Draw
		}
	}

	return Ongoing
}

const (
	all uint64 = 0x1FFFFFFFFFFFF
	// Files
	fileA uint64 = 0x0040810204081
	fileB uint64 = 0x0081020408102
	fileF uint64 = 0x0810204081020
	fileG uint64 = 0x1020408102040
	// Ranks
	// Not Files
	notFileA uint64 = 0x1fbf7efdfbf7e
	notFileB uint64 = 0x1f7efdfbf7efd
	notFileF uint64 = 0x17efdfbf7efdf
	notFileG uint64 = 0x0fdfbf7efdfbf
	// Stuff
	fileAB    uint64 = fileA | fileB
	fileFG    uint64 = fileF | fileG
	notFileAB uint64 = notFileA & notFileB
	notFileFG uint64 = notFileF & notFileG
)

// Bitboard ...
type Bitboard struct {
	Data uint64
}

// Set ...
func (bb *Bitboard) Set(sq Square) {
	bb.Data |= 1 << sq.Data
}

// Unset ...
func (bb *Bitboard) Unset(sq Square) {
	bb.Data &= ^(1 << sq.Data)
}

// Get ...
func (bb Bitboard) Get(sq Square) bool {
	return bb.Data&(1<<sq.Data) != 0
}

// Count ...
func (bb Bitboard) Count() int {
	return bits.OnesCount64(bb.Data)
}

// LSB ...
func (bb Bitboard) LSB() uint8 {
	return uint8(bits.TrailingZeros64(bb.Data))
}

// North ...
func (bb Bitboard) North() Bitboard {
	return Bitboard{(bb.Data << 7) & all}
}

// South ...
func (bb Bitboard) South() Bitboard {
	return Bitboard{bb.Data >> 7}
}

// East ...
func (bb Bitboard) East() Bitboard {
	return Bitboard{(bb.Data << 1) & notFileA}
}

// West ...
func (bb Bitboard) West() Bitboard {
	return Bitboard{(bb.Data >> 1) & notFileG}
}

// NorthEast ...
func (bb Bitboard) NorthEast() Bitboard {
	return Bitboard{(bb.Data << 8) & notFileA}
}

// NorthWest ...
func (bb Bitboard) NorthWest() Bitboard {
	return Bitboard{(bb.Data << 6) & notFileG}
}

// SouthEast ...
func (bb Bitboard) SouthEast() Bitboard {
	return Bitboard{(bb.Data >> 6) & notFileA}
}

// SouthWest ...
func (bb Bitboard) SouthWest() Bitboard {
	return Bitboard{(bb.Data >> 8) & notFileG}
}

// Singles ...
func (bb Bitboard) Singles() Bitboard {
	return Bitboard{
		bb.NorthEast().Data |
			bb.NorthWest().Data |
			bb.SouthEast().Data |
			bb.SouthWest().Data |
			bb.North().Data |
			bb.South().Data |
			bb.East().Data |
			bb.West().Data}
}

// Doubles ...
func (bb Bitboard) Doubles() Bitboard {
	var moves uint64 = 0
	var asd = bb.Data
	moves |= (asd << 12) & notFileFG // North North West West
	moves |= (asd << 13) & notFileG  // North North West
	moves |= (asd << 14)             // North North
	moves |= (asd << 15) & notFileA  // North North East
	moves |= (asd << 16) & notFileAB // North North East East

	moves |= (asd >> 16) & notFileFG // South South West West
	moves |= (asd >> 15) & notFileG  // South South West
	moves |= (asd >> 14)             // South South
	moves |= (asd >> 13) & notFileA  // South South East
	moves |= (asd >> 12) & notFileAB // South South East East

	moves |= (asd << 9) & notFileAB // East East North
	moves |= (asd << 2) & notFileAB // East East
	moves |= (asd >> 5) & notFileAB // East East South

	moves |= (asd << 5) & notFileFG // West West North
	moves |= (asd >> 2) & notFileFG // West West
	moves |= (asd >> 9) & notFileFG // West West South

	return Bitboard{moves}
}

// Square ...
type Square struct {
	Data uint8
}

// File ...
func (sq *Square) File() uint8 {
	return sq.Data % 7
}

// Rank ...
func (sq *Square) Rank() uint8 {
	return sq.Data / 7
}

// Position ...
type Position struct {
	pieces    [2]Bitboard
	gaps      Bitboard
	turn      int
	halfmoves int
}

// NewPosition ...
func NewPosition(fen string) (*Position, error) {
	var position Position
	position.SetFen(fen)
	return &position, nil
}

// Turn ...
func (pos *Position) Turn() int {
	return pos.turn
}

// Us ...
func (pos *Position) Us() Bitboard {
	return pos.pieces[pos.turn]
}

// Them ...
func (pos *Position) Them() Bitboard {
	return pos.pieces[1-pos.turn]
}

// Set ...
func (pos *Position) Set(sq Square, piece int) {
	switch piece {
	case 0:
		pos.pieces[0].Set(sq)
		pos.pieces[1].Unset(sq)
		pos.gaps.Unset(sq)
	case 1:
		pos.pieces[0].Unset(sq)
		pos.pieces[1].Set(sq)
		pos.gaps.Unset(sq)
	case 2:
		pos.pieces[0].Unset(sq)
		pos.pieces[1].Unset(sq)
		pos.gaps.Set(sq)
	default:
		pos.pieces[0].Unset(sq)
		pos.pieces[1].Unset(sq)
		pos.gaps.Unset(sq)
	}
}

// Get ...
func (pos *Position) Get(sq Square) int {
	if pos.pieces[0].Get(sq) {
		return 0
	}
	if pos.pieces[1].Get(sq) {
		return 1
	}
	if pos.gaps.Get(sq) {
		return 2
	}
	return 3
}

// Print ...
func (pos Position) Print() {
	for i := 42; i >= 0; i++ {
		sq := Square{Data: uint8(i)}
		switch pos.Get(sq) {
		case 0:
			fmt.Print("x ")
		case 1:
			fmt.Print("o ")
		case 2:
			fmt.Print("  ")
		default:
			fmt.Print("- ")
		}

		if i%7 == 6 {
			fmt.Println()
			i -= 14
		}
	}
}

// SetFen ...
func (pos *Position) SetFen(fen string) {
	// Default
	pos.pieces[0].Data = 0
	pos.pieces[1].Data = 0
	pos.gaps.Data = 0
	pos.turn = 0
	pos.halfmoves = 0

	results := strings.Split(fen, " ")

	// Pieces
	if len(results) >= 1 {
		var sq uint8 = 42
		for i := 0; i < len(results[0]); i++ {
			switch results[0][i] {
			case 'x':
				pos.Set(Square{sq}, 0)
				sq++
			case 'o':
				pos.Set(Square{sq}, 1)
				sq++
			case '-':
				pos.Set(Square{sq}, 2)
				sq++
			case '1':
				sq++
			case '2':
				sq += 2
			case '3':
				sq += 3
			case '4':
				sq += 4
			case '5':
				sq += 5
			case '6':
				sq += 6
			case '7':
				sq += 7
			case '/':
				sq -= 14
			}
		}
	}

	// Turn
	if len(results) >= 2 {
		if results[1] == "x" {
			pos.turn = 0
		} else {
			pos.turn = 1
		}
	}

	// Halfmove clock
	if len(results) >= 3 {
		pos.halfmoves, _ = strconv.Atoi(results[2])
	}
}

// Move ...
type Move struct {
	From Square
	To   Square
}

// NULLMOVE ...
var NULLMOVE = Move{Square{49}, Square{49}}

// NewMove ...
func NewMove(movestr string) (*Move, error) {
	if movestr == "0000" {
		return &NULLMOVE, nil
	} else if len(movestr) == 2 {
		f := uint8(movestr[0] - 'a')
		r := uint8(movestr[1] - '1')
		to := Square{r*7 + f}
		return &Move{to, to}, nil
	} else if len(movestr) == 4 {
		f1 := uint8(movestr[0] - 'a')
		r1 := uint8(movestr[1] - '1')
		fr := Square{r1*7 + f1}
		f2 := uint8(movestr[2] - 'a')
		r2 := uint8(movestr[3] - '1')
		to := Square{r2*7 + f2}
		return &Move{fr, to}, nil
	}
	return nil, errors.New("Failed to parse move string")
}

// IsSingle ...
func (move *Move) IsSingle() bool {
	return move.From == move.To
}

// IsDouble ...
func (move *Move) IsDouble() bool {
	return move.From != move.To
}

func (move Move) String() string {
	if move == NULLMOVE {
		return "0000"
	} else if move.IsSingle() {
		return fmt.Sprintf("%c%c", 'a'+move.To.File(), '1'+move.To.Rank())
	}
	return fmt.Sprintf("%c%c%c%c", 'a'+move.From.File(), '1'+move.From.Rank(), 'a'+move.To.File(), '1'+move.To.Rank())
}

// MakeMove ...
func (pos *Position) MakeMove(move Move) {
	// Nullmove
	if move == NULLMOVE {
		pos.turn = 1 - pos.turn
		return
	}

	bbTo := Bitboard{uint64(1) << move.To.Data}
	bbFrom := Bitboard{uint64(1) << move.From.Data}
	neighbours := bbTo.Singles().Data

	// Move our piece
	pos.pieces[pos.turn].Data ^= bbTo.Data | bbFrom.Data

	// Flip captured pieces
	captured := pos.pieces[1-pos.turn].Data & neighbours
	pos.pieces[pos.turn].Data ^= captured
	pos.pieces[1-pos.turn].Data ^= captured

	// Halfmove counter
	pos.halfmoves++
	if captured != 0 || move.IsSingle() {
		pos.halfmoves = 0
	}

	// Adjust the hashKey for flipped pieces
	for captured != 0 {
		captured &= captured - 1
	}

	// Flip turn
	pos.turn = 1 - pos.turn
}

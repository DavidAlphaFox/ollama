package readline

import (
	"fmt"

	"github.com/emirpasic/gods/lists/arraylist"
	"golang.org/x/term"
)

type Buffer struct {
	Pos       int
	Buf       *arraylist.List
	Prompt    *Prompt
	LineWidth int
	Width     int
	Height    int
}

func NewBuffer(prompt *Prompt) (*Buffer, error) {
	width, height, err := term.GetSize(0)
	if err != nil {
		fmt.Println("Error getting size:", err)
		return nil, err
	}

	b := &Buffer{
		Pos:       0,
		Buf:       arraylist.New(),
		Prompt:    prompt,
		Width:     width,
		Height:    height,
		LineWidth: width - len(prompt.Prompt),
	}

	return b, nil
}

func (b *Buffer) MoveLeft() {
	if b.Pos > 0 {
		if b.Pos%b.LineWidth == 0 {
			fmt.Printf(CursorUp + CursorBOL + fmt.Sprintf(CursorRightN, b.Width))
		} else {
			fmt.Printf(CursorLeft)
		}
		b.Pos -= 1
	}
}

func (b *Buffer) MoveRight() {
	if b.Pos < b.Size() {
		b.Pos += 1
		if b.Pos%b.LineWidth == 0 {
			fmt.Printf(CursorDown + CursorBOL + fmt.Sprintf(CursorRightN, len(b.Prompt.Prompt)))
		} else {
			fmt.Printf(CursorRight)
		}
	}
}

func (b *Buffer) MoveToStart() {
	if b.Pos > 0 {
		currLine := b.Pos / b.LineWidth
		if currLine > 0 {
			for cnt := 0; cnt < currLine; cnt++ {
				fmt.Printf(CursorUp)
			}
		}
		fmt.Printf(CursorBOL + fmt.Sprintf(CursorRightN, len(b.Prompt.Prompt)))
		b.Pos = 0
	}
}

func (b *Buffer) MoveToEnd() {
	if b.Pos < b.Size() {
		currLine := b.Pos / b.LineWidth
		totalLines := b.Size() / b.LineWidth
		if currLine < totalLines {
			for cnt := 0; cnt < totalLines-currLine; cnt++ {
				fmt.Printf(CursorDown)
			}
			remainder := b.Size() % b.LineWidth
			fmt.Printf(CursorBOL + fmt.Sprintf(CursorRightN, len(b.Prompt.Prompt)+remainder))
		} else {
			fmt.Printf(fmt.Sprintf(CursorRightN, b.Size()-b.Pos))
		}

		b.Pos = b.Size()
	}
}

func (b *Buffer) Size() int {
	return b.Buf.Size()
}

func min(n, m int) int {
	if n > m {
		return m
	}
	return n
}

func (b *Buffer) Add(r rune) {
	if b.Pos == b.Buf.Size() {
		fmt.Printf("%c", r)
		b.Buf.Add(r)
		b.Pos += 1
		if b.Pos > 0 && b.Pos%b.LineWidth == 0 {
			fmt.Printf("\n... ")
		}
	} else {
		fmt.Printf("%c", r)
		b.Buf.Insert(b.Pos, r)
		b.Pos += 1
		if b.Pos > 0 && b.Pos%b.LineWidth == 0 {
			fmt.Printf("\n... ")
		}
		b.drawRemaining()
	}
}

func (b *Buffer) drawRemaining() {
	var place int
	remainingText := b.StringN(b.Pos)
	if b.Pos > 0 {
		place = b.Pos % b.LineWidth
	}

	// render the rest of the current line
	currLine := remainingText[:min(b.LineWidth-place, len(remainingText))]
	if len(currLine) > 0 {
		fmt.Printf(ClearToEOL + currLine)
		fmt.Print(fmt.Sprintf(CursorLeftN, len(currLine)))
	} else {
		fmt.Printf(ClearToEOL)
	}

	// render the other lines
	if len(remainingText) > len(currLine) {
		fmt.Printf(CursorSave)

		remaining := []rune(remainingText[len(currLine):])
		for i, c := range remaining {
			if i%b.LineWidth == 0 {
				fmt.Printf("\n... ")
			}
			fmt.Printf("%c", c)
		}
		fmt.Printf(ClearToEOL + CursorRestore)
	}
}

func (b *Buffer) Remove() {
	if b.Buf.Size() > 0 && b.Pos > 0 {
		if b.Pos%b.LineWidth == 0 {
			// if the user backspaces over the word boundary, do this magic to clear the line
			// and move to the end of the previous line
			fmt.Printf(CursorBOL + ClearToEOL)
			fmt.Printf(CursorUp + CursorBOL + fmt.Sprintf(CursorRightN, b.Width) + " " + CursorLeft)
		} else {
			fmt.Printf(CursorLeft + " " + CursorLeft)
		}
		var eraseExtraLine bool
		if (b.Size()-1)%b.LineWidth == 0 {
			eraseExtraLine = true
		}

		b.Pos -= 1
		b.Buf.Remove(b.Pos)

		if b.Pos < b.Size() {
			b.drawRemaining()
			// this erases a line which is left over when backspacing in the middle of a line and there
			// are trailing characters which go over the line width boundary
			if eraseExtraLine {
				currPos := b.Pos
				fmt.Printf(CursorSave)
				b.MoveToEnd()
				fmt.Printf(CursorBOL + ClearToEOL + CursorRestore)
				b.Pos = currPos
			}
		}
	}
}

func (b *Buffer) Delete() {
	if b.Buf.Size() > 0 && b.Pos < b.Size() {
		b.Buf.Remove(b.Pos)
		b.drawRemaining()
		if b.Size()%b.LineWidth == 0 {
			if b.Pos == b.Size() {
				fmt.Printf(CursorRight)
			} else {
				currPos := b.Pos
				fmt.Printf(CursorSave)
				b.MoveToEnd()
				fmt.Printf(CursorBOL + ClearToEOL + CursorRestore)
				b.Pos = currPos
			}
		}
	}
}

func (b *Buffer) DeleteBefore() {
	if b.Pos > 0 {
		for cnt := b.Pos - 1; cnt >= 0; cnt-- {
			b.Remove()
		}
	}
}

func (b *Buffer) DeleteRemaining() {
	if b.Size() > 0 && b.Pos < b.Size() {
		fmt.Printf(ClearToEOL)
		totalLines := b.Size() / b.LineWidth
		if totalLines > 0 {
			fmt.Printf(CursorSave)
			for cnt := 0; cnt < totalLines; cnt++ {
				fmt.Printf(CursorDown + CursorBOL + ClearToEOL)
			}
			fmt.Printf(CursorRestore)
		}
		for cnt := b.Buf.Size() - 1; cnt >= b.Pos; cnt-- {
			b.Buf.Remove(cnt)
		}
	}
}

func (b *Buffer) DeleteWord() {
	if b.Buf.Size() > 0 && b.Pos > 0 {
		var foundNonspace bool
		for {
			v, _ := b.Buf.Get(b.Pos - 1)
			if v == ' ' {
				if !foundNonspace {
					b.Remove()
				} else {
					break
				}
			} else {
				foundNonspace = true
				b.Remove()
			}

			if b.Pos == 0 {
				break
			}
		}
	}
}

func (b *Buffer) ClearScreen() {
	fmt.Printf(ClearScreen + CursorReset + b.Prompt.Prompt)
	if b.Empty() {
		ph := b.Prompt.Placeholder
		fmt.Printf(ColorGrey + ph + fmt.Sprintf(CursorLeftN, len(ph)) + ColorDefault)
	} else {
		currPos := b.Pos
		b.Pos = 0
		b.drawRemaining()

		if currPos > 0 {
			targetLine := currPos / b.LineWidth
			if targetLine > 0 {
				for cnt := 0; cnt < targetLine; cnt++ {
					fmt.Printf(CursorDown)
				}
			}
			remainder := currPos % b.LineWidth
			fmt.Printf(fmt.Sprintf(CursorRightN, remainder))
			if currPos%b.LineWidth == 0 {
				fmt.Printf(CursorBOL + "... ")
			}
		}

		b.Pos = currPos
	}
}

func (b *Buffer) Empty() bool {
	return b.Buf.Empty()
}

func (b *Buffer) String() string {
	return b.StringN(0)
}

func (b *Buffer) StringN(n int) string {
	return b.StringNM(n, 0)
}

func (b *Buffer) StringNM(n, m int) string {
	var s string
	if m == 0 {
		m = b.Size()
	}
	for cnt := n; cnt < m; cnt++ {
		c, _ := b.Buf.Get(cnt)
		s += string(c.(rune))
	}
	return s
}

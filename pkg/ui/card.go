package ui

import (
	"github.com/auroq/botropolis/pkg/city"
)

// Card is a hover or selection card laid out on a panel: a title at the
// title size and lines at the body size.
type Card struct {
	Rect    city.Rect
	Title   Text
	Lines   []Text
	Buttons []Button
}

// LayoutCard sizes a card to its content and parks it in the top-right
// corner of bounds, a margin in; lines too wide are clipped and lines
// that would run past the bottom are dropped.
func LayoutCard(th Theme, card city.Card, bounds city.Rect, measure Measure) Card {
	pad, margin := 2*th.Grid(), 2*th.Grid()
	maxWidth := bounds.Width() - 2*margin - 2*pad
	title := clipTo(card.Title, maxWidth, Title, measure)
	width, titleH := measure(title, Title)
	_, lineH := measure("", Body)
	maxLines := int((bounds.Height() - 2*margin - 2*pad - titleH) / lineH)
	var lines []string
	for _, line := range card.Lines {
		if len(lines) >= maxLines {
			break
		}
		line = clipTo(line, maxWidth, Body, measure)
		if w, _ := measure(line, Body); w > width {
			width = w
		}
		lines = append(lines, line)
	}
	c := Card{Rect: city.RectAt(0, 0, width+2*pad, 2*pad+titleH+lineH*float64(len(lines)))}
	c.Title = Text{Text: title, At: city.Point{X: pad, Y: pad}, Size: Title}
	for i, line := range lines {
		c.Lines = append(c.Lines, Text{Text: line, At: city.Point{X: pad, Y: pad + titleH + lineH*float64(i)}, Size: Body})
	}
	if len(card.Actions) > 0 {
		row := LayoutButtons(th, card.Actions, city.Point{X: pad, Y: c.Rect.Max.Y - pad + th.Grid()}, measure)
		last := row[len(row)-1].Rect
		c.Rect.Max.Y = last.Max.Y + pad
		if last.Max.X+pad > c.Rect.Max.X {
			c.Rect.Max.X = last.Max.X + pad
		}
		c.Buttons = row
	}
	x := bounds.Max.X - margin - c.Rect.Width()
	if x < bounds.Min.X+margin {
		x = bounds.Min.X + margin
	}
	return c.MoveTo(city.Point{X: x, Y: bounds.Min.Y + margin})
}

// PinTo places the card beside a screen rect: to its right when there is
// room, else to its left, kept inside bounds.
func (c Card) PinTo(beside city.Rect, bounds city.Rect, gap float64) Card {
	at := city.Point{X: beside.Max.X + gap, Y: beside.Min.Y}
	if at.X+c.Rect.Width() > bounds.Max.X {
		at.X = beside.Min.X - gap - c.Rect.Width()
	}
	if at.X < bounds.Min.X {
		at.X = bounds.Min.X
	}
	if at.Y+c.Rect.Height() > bounds.Max.Y {
		at.Y = bounds.Max.Y - c.Rect.Height()
	}
	if at.Y < bounds.Min.Y {
		at.Y = bounds.Min.Y
	}
	return c.MoveTo(at)
}

// MoveTo puts the card's top-left corner at a point, carrying its text.
func (c Card) MoveTo(at city.Point) Card {
	by := at.Sub(c.Rect.Min)
	moved := Card{Rect: city.Rect{Min: at, Max: c.Rect.Max.Add(by)}, Title: c.Title.moved(by)}
	for _, line := range c.Lines {
		moved.Lines = append(moved.Lines, line.moved(by))
	}
	for _, b := range c.Buttons {
		moved.Buttons = append(moved.Buttons, Button{Rect: city.Rect{Min: b.Rect.Min.Add(by), Max: b.Rect.Max.Add(by)}, Label: b.Label.moved(by)})
	}
	return moved
}

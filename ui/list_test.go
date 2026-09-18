package ui

import (
	"claude-squad/session"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/require"
)

func newTestList(titles ...string) *List {
	s := spinner.New()
	l := NewList(&s, false)
	for _, t := range titles {
		inst, _ := session.NewInstance(session.InstanceOptions{
			Title:   t,
			Path:    ".",
			Program: "echo",
		})
		l.AddInstance(inst)
	}
	return l
}

func TestMoveUp(t *testing.T) {
	l := newTestList("a", "b", "c")
	l.SetSelectedInstance(1) // select "b"

	moved := l.MoveUp()
	require.True(t, moved)
	require.Equal(t, 0, l.selectedIdx)
	require.Equal(t, "b", l.items[0].Title)
	require.Equal(t, "a", l.items[1].Title)
	require.Equal(t, "c", l.items[2].Title)
}

func TestMoveUp_AtTop(t *testing.T) {
	l := newTestList("a", "b", "c")
	l.SetSelectedInstance(0)

	moved := l.MoveUp()
	require.False(t, moved)
	require.Equal(t, 0, l.selectedIdx)
	require.Equal(t, "a", l.items[0].Title)
}

func TestMoveDown(t *testing.T) {
	l := newTestList("a", "b", "c")
	l.SetSelectedInstance(1) // select "b"

	moved := l.MoveDown()
	require.True(t, moved)
	require.Equal(t, 2, l.selectedIdx)
	require.Equal(t, "a", l.items[0].Title)
	require.Equal(t, "c", l.items[1].Title)
	require.Equal(t, "b", l.items[2].Title)
}

func TestMoveDown_AtBottom(t *testing.T) {
	l := newTestList("a", "b", "c")
	l.SetSelectedInstance(2)

	moved := l.MoveDown()
	require.False(t, moved)
	require.Equal(t, 2, l.selectedIdx)
	require.Equal(t, "c", l.items[2].Title)
}

func TestMoveWithSingleItem(t *testing.T) {
	l := newTestList("only")
	l.SetSelectedInstance(0)

	require.False(t, l.MoveUp())
	require.False(t, l.MoveDown())
}

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

// visibleTitles returns the instance titles which the list actually rendered.
func visibleTitles(t *testing.T, l *List, titles []string) []string {
	t.Helper()
	out := ansiRe.ReplaceAllString(l.String(), "")
	var visible []string
	for i, title := range titles {
		if strings.Contains(out, fmt.Sprintf("%d.  %s", i+1, title)) {
			visible = append(visible, title)
		}
	}
	return visible
}

func TestStringFitsInHeight(t *testing.T) {
	titles := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	l := newTestList(titles...)

	// A short terminal (e.g. after increasing the font size) can't fit every instance.
	for _, height := range []int{4, 8, 10, 15, 20, 30, 60} {
		l.SetSize(40, height)
		require.LessOrEqual(t, lipgloss.Height(l.String()), height, "height %d", height)
	}
}

func TestStringKeepsFirstInstanceVisible(t *testing.T) {
	titles := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	l := newTestList(titles...)
	l.SetSize(40, 14)
	l.SetSelectedInstance(0)

	visible := visibleTitles(t, l, titles)
	require.NotEmpty(t, visible)
	require.Equal(t, "a", visible[0])
}

func TestStringScrollsToSelectedInstance(t *testing.T) {
	titles := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	l := newTestList(titles...)
	l.SetSize(40, 14)

	// Walking down the list keeps the selected instance on screen.
	for i := range titles {
		l.SetSelectedInstance(i)
		require.Contains(t, visibleTitles(t, l, titles), titles[i], "selected %s", titles[i])
		require.LessOrEqual(t, lipgloss.Height(l.String()), 14)
	}

	// The last instance is selected, so the list is scrolled to the bottom.
	visible := visibleTitles(t, l, titles)
	require.Equal(t, "h", visible[len(visible)-1])

	// Walking back up scrolls back to the top.
	for i := len(titles) - 1; i >= 0; i-- {
		l.SetSelectedInstance(i)
		require.Contains(t, visibleTitles(t, l, titles), titles[i], "selected %s", titles[i])
	}
	require.Equal(t, "a", visibleTitles(t, l, titles)[0])
}

func TestStringTinyHeight(t *testing.T) {
	titles := []string{"a", "b", "c"}
	l := newTestList(titles...)
	l.SetSelectedInstance(2)

	// Not even the title plus one instance fits in these. Nothing useful can be shown,
	// but the list must still stay inside the height it was given.
	for _, height := range []int{1, 2, 3, 5} {
		l.SetSize(40, height)
		require.Equal(t, height, lipgloss.Height(l.String()), "height %d", height)
	}
}

func TestStringUnboundedHeightRendersEverything(t *testing.T) {
	titles := []string{"a", "b", "c"}
	l := newTestList(titles...)
	l.SetSize(40, 0)

	require.Equal(t, titles, visibleTitles(t, l, titles))
}

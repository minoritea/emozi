package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
	"github.com/pkg/errors"
	"github.com/sahilm/fuzzy"
	"log"
)

var sNames []string

func init() {
	for sn := range emojiCodeMap {
		sNames = append(sNames, sn)
	}
}

func view(s tcell.Screen, search string, matches fuzzy.Matches, cursor uint) {
	s.Clear()

	putln(s, 0, "find: "+search)

	for i, m := range matches {
		str := "  "
		if uint(i) == cursor {
			str = "> "
		}

		emoji, ok := emojiCodeMap[m.Str]
		if !ok {
			emoji = ""
		}
		str += emoji
		putln(s, i+1, str)
	}
}

func find(search string) fuzzy.Matches {
	return fuzzy.Find(search, sNames)
}

var (
	errInterrupt = errors.New("interrupt")
	errNotFound  = errors.New("not found")
)

func putln(s tcell.Screen, y int, str string) {
	var x int

	style := tcell.StyleDefault
	var cc []rune // combined charactors

	for _, r := range str {
		if runewidth.RuneWidth(r) > 0 {
			switch len(cc) {
			case 0:
				// nothing
			case 1:
				s.SetContent(x, y, cc[0], nil, style)
			default:
				s.SetContent(x, y, cc[0], cc[1:], style)
			}
			x++
			cc = cc[0:0]
		}
		cc = append(cc, r)
	}

	switch len(cc) {
	case 0:
		// nothing
	case 1:
		s.SetContent(x, y, cc[0], nil, style)
	default:
		s.SetContent(x, y, cc[0], cc[1:], style)
	}
}

func run() (string, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return "", err
	}
	if err := s.Init(); err != nil {
		return "", err
	}
	defer s.Fini()

	s.Clear()

	search := ""
	matches := fuzzy.Matches{}
	var cursor uint // = 0
	view(s, search, matches, cursor)
	s.Sync()

	for {
		switch ev := s.PollEvent().(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyRune:
				search += string(ev.Rune())
				matches = find(search)
				goto RENDERING

			case tcell.KeyEscape, tcell.KeyCtrlC:
				return "", errInterrupt

			case tcell.KeyBS, tcell.KeyDEL:
				if len(search) >= 1 {
					search = search[0 : len(search)-1]
					matches = find(search)
				}
				goto RENDERING

			case tcell.KeyEnter:
				if len(matches) == 0 {
					return "", errNotFound
				}
				if uint(len(matches)) <= cursor {
					return "", errors.New("index out of range")
				}
				sn := matches[cursor].Str
				emoji, ok := emojiCodeMap[sn]
				if !ok {
					return "", errors.Errorf("emoji not found by key: `%s`", sn)
				}
				return emoji, nil

			case tcell.KeyCtrlK, tcell.KeyUp:
				cursor--
				goto RENDERING

			case tcell.KeyDown, tcell.KeyCtrlJ:
				cursor++
				if l := uint(len(matches)); cursor >= l {
					cursor = l - 1
				}
				goto RENDERING
			}
		case *tcell.EventResize:
			s.Sync()
		}
		continue

	RENDERING:
		s.Clear()
		view(s, search, matches, cursor)
		s.Sync()
	}
}

func main() {
	emoji, err := run()
	if err != nil {
		if err == errInterrupt || err == errNotFound {
			return
		}
		log.Fatal(err)
	}
	fmt.Println(emoji)
}

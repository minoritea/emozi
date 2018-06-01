package main

import (
	"bytes"
	"fmt"
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	"github.com/pkg/errors"
	"github.com/sahilm/fuzzy"
	"io"
	"log"
	"os"
)

var sNames []string

func init() {
	for sn := range emojiCodeMap {
		sNames = append(sNames, sn)
	}
}

func view(search string, matches fuzzy.Matches, cursol uint) {
	c := termbox.ColorDefault
	termbox.Clear(c, c)
	for i, r := range "find: " + search {
		termbox.SetCell(i, 0, r, c, c)
	}

	for i, m := range matches {
		if uint(i) == cursol {
			termbox.SetCell(0, i+1, '>', c, c)
		}
		emoji, ok := emojiCodeMap[m.Str]
		if !ok {
			emoji = ""
		}
		log.Printf("%s %s\t%+v", m.Str, emoji, []rune(emoji))
		shift := i + 1
		for j, r := range emoji {
			log.Printf("shift: %d", shift)
			termbox.SetCell(j+2, shift, r, c, c)
			shift += runewidth.RuneWidth(r)
		}
	}
	termbox.Flush()
}

func find(search string) fuzzy.Matches {
	return fuzzy.Find(search, sNames)
}

var (
	errInterrupt = errors.New("interrupt")
	errNotFound  = errors.New("not found")
)

func run() (string, error) {
	if err := termbox.Init(); err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()

	search := ""
	matches := fuzzy.Matches{}
	var cursol uint // = 0
	view(search, matches, cursol)

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Ch != 0 {
				search += string(ev.Ch)
				matches = find(search)
				goto RENDERING
			}

			switch ev.Key {
			case termbox.KeyCtrlC:
				return "", errInterrupt
			case termbox.KeyBackspace, termbox.KeyBackspace2, termbox.KeyCtrlD:
				if len(search) >= 1 {
					search = search[0 : len(search)-1]
					matches = find(search)
					goto RENDERING
				}
			case termbox.KeyEnter:
				if len(matches) == 0 {
					return "", errNotFound
				}
				if uint(len(matches)) <= cursol {
					return "", errors.New("index out of range")
				}
				sn := matches[cursol].Str
				emoji, ok := emojiCodeMap[sn]
				if !ok {
					return "", errors.Errorf("emoji not found by key: `%s`", sn)
				}
				return emoji, nil
			case termbox.KeyArrowUp, termbox.KeyCtrlK:
				cursol--
				goto RENDERING
			case termbox.KeyArrowDown, termbox.KeyCtrlJ:
				cursol++
				if l := uint(len(matches)); cursol >= l {
					cursol = l - 1
				}
				goto RENDERING
			}
			continue
		RENDERING:
			view(search, matches, cursol)
		}
	}
}

func main() {
	l := setupLog()
	defer l.Flush()
	emoji, err := run()
	if err != nil {
		if err == errInterrupt || err == errNotFound {
			return
		}
		log.Fatal(err)
	}
	fmt.Println(emoji)
}

type logBuffer struct{ *bytes.Buffer }

func (l *logBuffer) Flush() {
	io.Copy(os.Stderr, l)
}

func setupLog() *logBuffer {
	l := &logBuffer{bytes.NewBuffer(nil)}
	log.SetOutput(l)
	return l
}

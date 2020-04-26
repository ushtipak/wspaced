package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/ajstarks/svgo"
	"io/ioutil"
	"log"
	"os"
)

var (
	whitespaceCode = flag.String("ws", "/tmp/code.ws", "hex color base")
	output		   = flag.String("output", "/tmp/logo.svg", "output svg")
	colorBase      = flag.String("base", "586e75", "base hex color")
	colorAccent    = flag.String("accent", "cb4b16", "accent hex color")
	tabScale       = flag.Int("tab-scale", 7, "ration of space to tab")
	blockSize      = flag.Int("block", 2, "square block size in px")
)

// readWCode reads Whitespace code from input file and returns it as byte slice, with longest sequence of chars
func readWCode() (code []byte, longestSequence int) {
	f, err := os.Open(*whitespaceCode)
	if err != nil {
		log.Fatalf("os.Open [%s]", err)
	}
	code, err = ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("ioutil.ReadAll [%s]", err)
	}
	err = f.Close()
	if err != nil {
		log.Fatalf("f.Close [%s]", err)
	}

	// split code into sequences of tab-space chars, terminated by a linefeed
	sequences := bytes.Split(code, []byte{10})
	longestSequence = 0
	for _, sequence := range sequences {
		// get the longest by comparing chains of unified ones of replaced tab bytes with spaces with target ratio
		unified := bytes.Replace(sequence, []byte{9}, bytes.Repeat([]byte{32}, *tabScale), -1)
		l := len(unified)
		if l > longestSequence {
			longestSequence = l
		}
	}

	// remove tailing linefeed chars and seal with a single one
	for {
		last := code[len(code)-1]
		if last != 10 {
			break
		}
		code = code[:len(code)-1]
	}
	code = append(code, []byte{10}...)
	return
}

// drawSVG creates a SVG file from input bytes by drawing tabs and spaces in different sizes
func drawSVG(code []byte, longestSequence int) {
	blockSpace := *blockSize
	blockTab := *tabScale * blockSpace
	rows := bytes.Count(code, []byte{10})

	// allocate canvas of sufficient width
	width := rows*blockSpace
	height := longestSequence * blockSpace

	f, _ := os.Create(*output)
	canvas := svg.New(f)
	canvas.Start(width, height)

	var color string
	x := 0
	y := 0
	for _, c := range code {

		// linefeed, switch next column
		if c == 10 {
			x += blockSpace
			y = 0
		}

		// tab, draw a square
		if c == 9 {
			color = fmt.Sprintf("fill:#%v", *colorBase)
			canvas.Rect(x, y, blockSpace, blockTab, color)
			y += blockTab
		}

		// space, draw a rect
		if c == 32 {
			color = fmt.Sprintf("fill:#%v", *colorAccent)
			canvas.Square(x, y, blockSpace, color)
			y += blockSpace
		}

	}
	canvas.End()
}

func main() {
	flag.Parse()
	log.Println("wspaced start")

	log.Printf("read whitespace code [%v] ...", *whitespaceCode)
	code, l := readWCode()
	log.Printf("  code(:12) %v", code[:12])
	log.Printf("  longest sequence [%v]", l)

	log.Printf("draw svg [%v] ...", *output)
	drawSVG(code, l)

	log.Println("all done \\o/")
}

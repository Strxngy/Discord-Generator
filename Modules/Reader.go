package module

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
)

// Reader IDONNO struct
type Reader struct {
	File  string
	Lines []string
}

func (p *Reader) Read() {
	f, err := os.Open(p.File)
	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		p.Lines = append(p.Lines, scanner.Text())
	}
}

// Remove specific line
func (p *Reader) Remove(item string) {
	newitems := []string{}

	for _, i := range p.Lines {
		if i != item {
			newitems = append(newitems, i)
		}
	}
	p.Lines = newitems
}

// Add new line
func (p *Reader) Add(line string) {
	p.Lines = append(p.Lines, line)
}

// Rand returns random p
func (p Reader) Rand() string {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(p.Lines))))
	if err != nil {
		fmt.Println(err)
	}
	return p.Lines[n.Int64()+0]
}

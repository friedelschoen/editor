package iorw

import (
	"io"
	"unicode/utf8"
)

func ReadRuneAt(r ReaderAt, i int) (rune, int, error) {
	n := utf8.UTFMax
	b, err := r.ReadFastAt(i, n)
	if err != nil {
		return 0, 0, err
	}
	ru, size := utf8.DecodeRune(b)
	if size == 0 {
		return 0, 0, io.EOF
	}
	return ru, size, nil
}

func ReadLastRuneAt(r ReaderAt, i int) (rune, int, error) {
	if i == 0 {
		return 0, 0, io.EOF
	}

	// handle left limit
	n := utf8.UTFMax
	min := r.Min()
	if i >= min && i-n < min {
		n = i - min
	}

	b, err := r.ReadFastAt(i-n, n)
	if err != nil {
		return 0, 0, err
	}
	ru, size := utf8.DecodeLastRune(b)
	if size == 0 {
		return 0, 0, io.EOF
	}
	return ru, size, nil
}

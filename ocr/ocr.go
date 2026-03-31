package ocr

import internalocr "github.com/libi/ko-browser/internal/ocr"

type Engine = internalocr.Engine

func NewEngine(langs ...string) (*Engine, error) {
	return internalocr.NewEngine(langs...)
}

package frontend

import (
	"os"

	"github.com/speedata/boxesandglue/backend/bag"
	"github.com/speedata/boxesandglue/backend/color"
	"github.com/speedata/boxesandglue/backend/document"
	"github.com/speedata/boxesandglue/backend/font"
	"github.com/speedata/boxesandglue/pdfbackend/pdf"
	"github.com/speedata/textlayout/harfbuzz"
)

// Document holds convenience functions.
type Document struct {
	FontFamilies    map[string]*FontFamily
	Doc             *document.PDFDocument
	DefaultFeatures []harfbuzz.Feature
	FindFile        func(string) string
	usedcolors      map[string]*color.Color
	usedSpotcolors  map[*color.Color]bool
	usedFonts       map[*pdf.Face]map[bag.ScaledPoint]*font.Font
	dirstack        []string
}

func initDocument() *Document {
	d := &Document{
		usedSpotcolors: make(map[*color.Color]bool),
		usedcolors:     make(map[string]*color.Color),
		usedFonts:      make(map[*pdf.Face]map[bag.ScaledPoint]*font.Font),
		FontFamilies:   make(map[string]*FontFamily),
	}
	d.FindFile = d.findFile
	return d
}

// New creates a new PDF file. After Doc.Finish() is called, the file is closed.
func New(filename string) (*Document, error) {
	w, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	fe := initDocument()
	fe.Doc = document.NewDocument(w)
	fe.Doc.Filename = filename
	return fe, nil
}

// Finish writes all necessary objects for the PDF.
func (fe *Document) Finish() error {
	for col := range fe.usedSpotcolors {
		fe.Doc.Spotcolors = append(fe.Doc.Spotcolors, col)
	}
	if len(fe.usedSpotcolors) > 0 {
		if fe.Doc.ColorProfile == nil {
			_, err := fe.Doc.LoadDefaultColorprofile()
			if err != nil {
				return err
			}
		}
	}
	return fe.Doc.Finish()
}

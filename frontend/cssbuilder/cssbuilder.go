package cssbuilder

import (
	"os"
	"path/filepath"

	"github.com/PuerkitoBio/goquery"
	"github.com/speedata/boxesandglue/backend/bag"
	"github.com/speedata/boxesandglue/backend/node"
	"github.com/speedata/boxesandglue/csshtml"
	"github.com/speedata/boxesandglue/frontend"
	"github.com/speedata/boxesandglue/frontend/pdfdraw"
	"github.com/speedata/boxesandglue/htmlstyle"
	"golang.org/x/net/html"
)

var onecm = bag.MustSp("1cm")

// CSSBuilder handles HTML chunks and CSS instructions.
type CSSBuilder struct {
	pagebox               []node.Node
	currentPageDimensions PageDimensions
	frontend              *frontend.Document
	css                   *csshtml.CSS
	stylesStack           htmlstyle.StylesStack
}

// New creates an instance of the CSSBuilder.
func New(fd *frontend.Document, c *csshtml.CSS) *CSSBuilder {
	cb := CSSBuilder{
		css:         c,
		frontend:    fd,
		stylesStack: make(htmlstyle.StylesStack, 0),
		pagebox:     []node.Node{},
	}
	cb.css.FrontendDocument = fd

	return &cb
}

// PageDimensions contains the page size and the margins of the page.
type PageDimensions struct {
	Width         bag.ScaledPoint
	Height        bag.ScaledPoint
	MarginLeft    bag.ScaledPoint
	MarginRight   bag.ScaledPoint
	MarginTop     bag.ScaledPoint
	MarginBottom  bag.ScaledPoint
	PageAreaLeft  bag.ScaledPoint
	PageAreaTop   bag.ScaledPoint
	ContentWidth  bag.ScaledPoint
	ContentHeight bag.ScaledPoint
	masterpage    *csshtml.Page
}

func (cb *CSSBuilder) getPageType() *csshtml.Page {
	if first, ok := cb.css.Pages[":first"]; ok && len(cb.frontend.Doc.Pages) == 0 {
		return &first
	}
	isRight := len(cb.frontend.Doc.Pages)%2 == 0
	if right, ok := cb.css.Pages[":right"]; ok && isRight {
		return &right
	}
	if left, ok := cb.css.Pages[":left"]; ok && !isRight {
		return &left
	}
	if allPages, ok := cb.css.Pages[""]; ok {
		return &allPages
	}
	return nil
}

// InitPage makes sure that there is a valid page in the frontend.
func (cb *CSSBuilder) InitPage() error {
	var err error
	if cb.frontend.Doc.CurrentPage == nil {
		if defaultPage := cb.getPageType(); defaultPage != nil {
			wdStr, htStr := csshtml.PapersizeWdHt(defaultPage.Papersize)
			var wd, ht, mt, mb, ml, mr bag.ScaledPoint
			if wd, err = bag.Sp(wdStr); err != nil {
				return err
			}
			if ht, err = bag.Sp(htStr); err != nil {
				return err
			}
			if str := defaultPage.MarginTop; str == "" {
				mt = onecm
			} else {
				if mt, err = bag.Sp(str); err != nil {
					return err
				}
			}
			if str := defaultPage.MarginBottom; str == "" {
				mb = onecm
			} else {
				if mb, err = bag.Sp(str); err != nil {
					return err
				}
			}
			if str := defaultPage.MarginLeft; str == "" {
				ml = onecm
			} else {
				if ml, err = bag.Sp(str); err != nil {
					return err
				}
			}
			if str := defaultPage.MarginRight; str == "" {
				mr = onecm
			} else {
				if mr, err = bag.Sp(str); err != nil {
					return err
				}
			}

			res, _ := csshtml.ResolveAttributes(defaultPage.Attributes)
			styles := cb.stylesStack.PushStyles()
			if err = htmlstyle.StylesToStyles(styles, res, cb.frontend, cb.stylesStack.CurrentStyle().Fontsize); err != nil {
				return err
			}
			vl := node.NewVList()
			vl.Width = wd - ml - mr - styles.BorderLeftWidth - styles.BorderRightWidth - styles.PaddingLeft - styles.PaddingRight
			vl.Height = ht - mt - mb - styles.PaddingTop - styles.PaddingBottom - styles.BorderTopWidth - styles.BorderBottomWidth
			hv := frontend.HTMLValues{
				BorderLeftWidth:         styles.BorderLeftWidth,
				BorderRightWidth:        styles.BorderRightWidth,
				BorderTopWidth:          styles.BorderTopWidth,
				BorderBottomWidth:       styles.BorderBottomWidth,
				BorderTopStyle:          styles.BorderTopStyle,
				BorderLeftStyle:         styles.BorderLeftStyle,
				BorderRightStyle:        styles.BorderRightStyle,
				BorderBottomStyle:       styles.BorderBottomStyle,
				BorderTopColor:          styles.BorderTopColor,
				BorderLeftColor:         styles.BorderLeftColor,
				BorderRightColor:        styles.BorderRightColor,
				BorderBottomColor:       styles.BorderBottomColor,
				PaddingLeft:             styles.PaddingLeft,
				PaddingRight:            styles.PaddingRight,
				PaddingBottom:           styles.PaddingBottom,
				PaddingTop:              styles.PaddingTop,
				BorderTopLeftRadius:     styles.BorderTopLeftRadius,
				BorderTopRightRadius:    styles.BorderTopRightRadius,
				BorderBottomLeftRadius:  styles.BorderBottomLeftRadius,
				BorderBottomRightRadius: styles.BorderBottomRightRadius,
			}
			vl = cb.frontend.HTMLBorder(vl, hv)
			cb.stylesStack.PopStyles()

			// set page width / height
			cb.frontend.Doc.DefaultPageWidth = wd
			cb.frontend.Doc.DefaultPageHeight = ht
			cb.currentPageDimensions = PageDimensions{
				Width:         wd,
				Height:        ht,
				PageAreaLeft:  ml + styles.BorderLeftWidth + styles.PaddingLeft,
				PageAreaTop:   mt - styles.BorderTopWidth - styles.PaddingTop,
				ContentWidth:  wd - styles.BorderRightWidth - styles.PaddingRight - ml - mr - styles.BorderLeftWidth - styles.PaddingLeft,
				ContentHeight: ht - styles.BorderBottomWidth - styles.PaddingBottom - mt - mb - styles.BorderTopWidth - styles.PaddingTop,
				MarginTop:     mt,
				MarginBottom:  mb,
				MarginLeft:    ml,
				MarginRight:   mr,
				masterpage:    defaultPage,
			}
			cb.frontend.Doc.NewPage()
			if styles.BackgroundColor != nil {
				r := node.NewRule()
				x := pdfdraw.NewStandalone().ColorNonstroking(*styles.BackgroundColor).Rect(0, 0, wd, -ht).Fill()
				r.Pre = x.String()
				rvl := node.Vpack(r)
				rvl.Attributes = node.H{"origin": "page background color"}
				cb.frontend.Doc.CurrentPage.OutputAt(0, ht, rvl)
			}
			cb.frontend.Doc.CurrentPage.OutputAt(ml, ht-mt, vl)
		} else {
			// no page master found
			cb.frontend.Doc.DefaultPageWidth = bag.MustSp("210mm")
			cb.frontend.Doc.DefaultPageHeight = bag.MustSp("297mm")

			cb.currentPageDimensions = PageDimensions{
				Width:         cb.frontend.Doc.DefaultPageWidth,
				Height:        cb.frontend.Doc.DefaultPageHeight,
				ContentWidth:  cb.frontend.Doc.DefaultPageWidth - 2*onecm,
				ContentHeight: cb.frontend.Doc.DefaultPageHeight - 2*onecm,
				PageAreaLeft:  onecm,
				PageAreaTop:   onecm,
				MarginTop:     onecm,
				MarginBottom:  onecm,
				MarginLeft:    onecm,
				MarginRight:   onecm,
			}
			cb.frontend.Doc.NewPage()
		}

	}
	return err
}

// PageSize returns a struct with the dimensions of the current page.
func (cb *CSSBuilder) PageSize() (PageDimensions, error) {
	err := cb.InitPage()
	if err != nil {
		return PageDimensions{}, err
	}
	return cb.currentPageDimensions, nil
}

// ParseCSSString reads CSS instructions from a string.
func (cb *CSSBuilder) ParseCSSString(css string) error {
	var err error
	if err = cb.css.AddCSSText(css); err != nil {
		return err
	}
	return nil
}

// NewPage puts the current page into the PDF document and starts with a new page.
func (cb *CSSBuilder) NewPage() error {
	if err := cb.InitPage(); err != nil {
		return err
	}
	if err := cb.BeforeShipout(); err != nil {
		return err
	}
	cb.frontend.Doc.CurrentPage.Shipout()
	cb.frontend.Doc.NewPage()
	return nil
}

// ParseHTMLFromNode interprets the HTML string and applies all previously read CSS data.
func (cb *CSSBuilder) ParseHTMLFromNode(input *html.Node) (*frontend.Text, error) {
	doc := goquery.NewDocumentFromNode(input)
	sel, err := cb.css.ApplyCSS(doc)
	if err != nil {
		return nil, err
	}
	var te *frontend.Text
	if te, err = htmlstyle.ParseSelection(sel, cb.stylesStack, cb.frontend); err != nil {
		return nil, err
	}

	return te, nil
}

// ParseHTML interprets the HTML string and applies all previously read CSS data.
func (cb *CSSBuilder) ParseHTML(html string) (*frontend.Text, error) {
	doc, err := cb.css.ReadHTMLChunk(html)
	if err != nil {
		return nil, err
	}

	sel, err := cb.css.ApplyCSS(doc)
	if err != nil {
		return nil, err
	}
	var te *frontend.Text
	if te, err = htmlstyle.ParseSelection(sel, cb.stylesStack, cb.frontend); err != nil {
		return nil, err
	}

	return te, nil
}

// AddCSS reads the CSS instructions in css.
func (cb *CSSBuilder) AddCSS(css string) {
	cb.css.Stylesheet = append(cb.css.Stylesheet, csshtml.ConsumeBlock(csshtml.ParseCSSString(css), false))
}

type info struct {
	sumht  bag.ScaledPoint
	newY   bag.ScaledPoint
	hv     frontend.HTMLValues
	height bag.ScaledPoint
}

func hasContents(areaAttributes map[string]string) bool {
	return areaAttributes["content"] != "none" && areaAttributes["content"] != "normal"
}

type pageMarginBox struct {
	minWidth    bag.ScaledPoint
	maxWidth    bag.ScaledPoint
	areaWidth   bag.ScaledPoint
	hasContents bool
	widthAuto   bool
	halign      frontend.HorizontalAlignment
	x           bag.ScaledPoint
	y           bag.ScaledPoint
	wd          bag.ScaledPoint
	ht          bag.ScaledPoint
}

// ParseCSSFile reads the given file name and tries to parse the CSS contents
// from the file.
func (cb *CSSBuilder) ParseCSSFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return err
	}
	cb.css.Dirstack = []string{abs}
	return cb.css.AddCSSText(string(data))

}
package csshtml

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// OpenHTMLFile opens an HTML file
func (c *CSS) OpenHTMLFile(filename string) (*goquery.Document, error) {
	dir, fn := filepath.Split(filename)
	c.dirstack = append(c.dirstack, dir)
	dirs := strings.Join(c.dirstack, "")
	r, err := os.Open(filepath.Join(dirs, fn))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	c.document, err = goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	var errcond error
	c.document.Find(":root > head link").Each(func(i int, sel *goquery.Selection) {
		if stylesheetfile, attExists := sel.Attr("href"); attExists {
			block, err := c.ParseCSSFile(stylesheetfile)
			if err != nil {
				errcond = err
			}
			parsedStyles := ConsumeBlock(block, false)
			c.Stylesheet = append(c.Stylesheet, parsedStyles)
		}
	})
	if errcond != nil {
		return nil, errcond
	}
	c.processAtRules()
	_, err = c.ApplyCSS()
	if err != nil {
		return nil, err
	}
	return c.document, nil
}

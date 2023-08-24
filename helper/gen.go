// This is based on "gen.go" for the Go fonts
//
// The original file has this copyright info:
//
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/speedata/textlayout/fonts/truetype"
)

func updatefonts() error {
	for _, dir := range []string{"camingocode", "crimsonpro", "texgyreheros"} {
		fmt.Println(dir)
		fmt.Println(os.Getwd())
		fontdir := filepath.Join("fontsource", dir)

		ttfs, err := os.Open(fontdir)
		if err != nil {
			return err
		}

		infos, err := ttfs.ReadDir(-1)
		if err != nil {
			return err
		}
		defer ttfs.Close()

		for _, info := range infos {
			fontname := info.Name()
			fmt.Println("fontname", fontname)
			if !(strings.HasSuffix(fontname, ".otf") || strings.HasSuffix(fontname, ".ttf")) {
				continue
			}
			if err = do(filepath.Join(fontdir, fontname)); err != nil {
				return err
			}
		}
	}

	return nil
}

func do(ttfName string) error {
	fontName, err := fontName(ttfName)
	if err != nil {
		return err
	}
	pkgName := pkgName(ttfName)
	outDir := filepath.Join("fonts", pkgName)
	if err := os.MkdirAll(outDir, 0777); err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	src, err := ioutil.ReadFile(ttfName)
	if err != nil {
		log.Fatal(err)
	}

	b := new(bytes.Buffer)
	fmt.Fprintf(b, "// generated by rake updatefonts; DO NOT EDIT\n\n")
	fmt.Fprintf(b, "// Package %s provides the %q font\n", pkgName, fontName)
	fmt.Fprintf(b, "package %s\n\n", pkgName)
	fmt.Fprintf(b, "// TTF is the data for the %q font.\n", fontName)
	fmt.Fprintf(b, "var TTF = []byte{")
	for i, x := range src {
		if i&15 == 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(b, "%#02x,", x)
	}
	fmt.Fprintf(b, "\n}\n")

	dst, err := format.Source(b.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(outDir, "data.go"), dst, 0666); err != nil {
		log.Fatal(err)
	}
	return nil
}

// fontName maps "Go-Regular.ttf" to "Go Regular".
func fontName(ttfName string) (string, error) {
	r, err := os.Open(ttfName)
	if err != nil {
		return "", err
	}
	fnt, err := truetype.Parse(r)
	if err != nil {
		return "", err
	}
	return fnt.Names.SelectEntry(truetype.NameFontFamily).String(), nil
}

// pkgName maps "Go-Regular.ttf" to "goregular".
func pkgName(ttfName string) string {
	suffix := filepath.Ext(ttfName)
	ttfName = filepath.Base(ttfName)
	s := ttfName[:len(ttfName)-len(suffix)]
	s = strings.Replace(s, "-", "", -1)
	s = strings.ToLower(s)
	return s
}
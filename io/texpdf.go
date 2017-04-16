// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"bytes"
	"strings"
)

// FcnConvertNum is a function to convert number to string
type FcnConvertNum func(row int, x float64) string

// FcnRow is a function that returns the row value as string
type FcnRow func(row int) string

// Report holds data to generate LaTeX and PDF files
type Report struct {

	// configuration
	Title     string // title of pdf
	Author    string // author of pdf
	Landscape bool   // to format paper

	// default options
	TablePos    string  // default table positioning key; e.g. !t (to be written as [!t])
	NumFmt      string  // default number formatting string; e.g. "%v", "%g" or "%.3f"
	TableFontSz string  // default table fontsize string; e.g. \scriptsize
	TableColSep float64 // default table column separation in 'em'; e.g. 0.5 => \setlength{\tabcolsep}{0.5em}

	// options
	DoNotAlignTable   bool // align coluns in TeX table (has to loop over rows first...)
	DoNotUseGeomPkg   bool // do not use package geometry for margins
	DoNotGeneratePDF  bool // do not generate pdf when writing tex files
	DoNotShowMessages bool // do not show messages

	// internal
	buffer *bytes.Buffer
}

// Reset clears report
func (o *Report) Reset() {
	o.buffer.Reset()
}

// AddSection adds section and subsections to report
func (o *Report) AddSection(name string, level int) {
	sec := "section"
	for i := 0; i < level; i++ {
		if i < 2 {
			sec = "sub" + sec
		}
	}
	if o.buffer == nil {
		o.buffer = new(bytes.Buffer)
	}
	Ff(o.buffer, "\n")
	Ff(o.buffer, "\\%s{%s}\n", sec, name)
}

// AddTex adds TeX commands
func (o *Report) AddTex(commands string) {
	if o.buffer == nil {
		o.buffer = new(bytes.Buffer)
	}
	Ff(o.buffer, "\n%s\n", commands)
}

// AddTable adds tex table to report
//   caption -- caption of table
//   label -- label of table
//   keys -- column keys
//   T -- table values
//   key2tex -- maps key to tex formatted text of this key (i.e. equation). may be nil
//   key2convert -- maps key to function to convert numbers to string in that column. may be nil
func (o *Report) AddTable(caption, label string, keys []string, T map[string][]float64, key2tex map[string]string, key2numfmt map[string]FcnConvertNum) {

	// new buffer
	if o.buffer == nil {
		o.buffer = new(bytes.Buffer)
	}

	// fix default parameters
	o.fixDefaults()

	// find column widths and set formatting string
	strfmt := make([]string, len(keys)) // for each column
	if !o.DoNotAlignTable {
		widths := make([]int, len(keys)) // column widths
		for j, key := range keys {
			if key2tex == nil {
				widths[j] = imax(widths[j], len(key))
			} else {
				widths[j] = imax(widths[j], len(key2tex[key]))
			}
			for i, v := range T[key] {
				if key2numfmt == nil {
					widths[j] = imax(widths[j], len(Sf(o.NumFmt, v)))
				} else {
					widths[j] = imax(widths[j], len(Sf(key2numfmt[key](i, v))))
				}
			}
		}
		for j, width := range widths {
			strfmt[j] = "%" + Sf("%d", width) + "s"
		}
	} else {
		for j := 0; j < len(keys); j++ {
			strfmt[j] = "%s"
		}
	}

	// start table
	Ff(o.buffer, "\n")
	Ff(o.buffer, "\\begin{table*} [%s] \\centering\n", o.TablePos)
	Ff(o.buffer, "\\caption{%s}\n", caption)

	// set fontsize and column separation
	Ff(o.buffer, o.TableFontSz)
	Ff(o.buffer, " \\setlength{\\tabcolsep}{%gem}\n", o.TableColSep)

	// start tabular
	cc := ""
	for range keys {
		cc += "c"
	}
	Ff(o.buffer, "\\begin{tabular}[c]{%s} \\toprule\n", cc)

	// header
	for j, key := range keys {
		if j > 0 {
			Ff(o.buffer, " & ")
		}
		if key2tex == nil {
			Ff(o.buffer, strfmt[j], key)
		} else {
			Ff(o.buffer, strfmt[j], key2tex[key])
		}
	}
	Ff(o.buffer, " \\\\ \\hline\n")

	// rows
	nrows := len(T[keys[0]])
	for i := 0; i < nrows; i++ {
		if i > 0 {
			Ff(o.buffer, "\n")
		}
		for j, key := range keys {
			if j > 0 {
				Ff(o.buffer, " & ")
			}
			if key2numfmt == nil {
				Ff(o.buffer, strfmt[j], Sf(o.NumFmt, T[key][i]))
			} else {
				Ff(o.buffer, strfmt[j], key2numfmt[key](i, T[key][i]))
			}
		}
		Ff(o.buffer, " \\\\")
	}

	// end tabular and table
	Ff(o.buffer, "\n")
	Ff(o.buffer, "\\bottomrule\n")
	Ff(o.buffer, "\\end{tabular}\n")
	Ff(o.buffer, "\\label{tab:%s}\n", label)
	Ff(o.buffer, "\\end{table*}")
}

// WriteTexPdf writes tex file and generates pdf file
//  extra -- extra LaTeX commands; may be nil
func (o *Report) WriteTexPdf(dirout, fnkey string, extra *bytes.Buffer) (err error) {

	// header
	pdf := new(bytes.Buffer)
	if o.Landscape {
		Ff(pdf, "\\documentclass[a4paper,landscape]{article}\n")
	} else {
		Ff(pdf, "\\documentclass[a4paper]{article}\n")
	}
	Ff(pdf, "\\usepackage{amsmath}\n")
	Ff(pdf, "\\usepackage{amssymb}\n")
	Ff(pdf, "\\usepackage{booktabs}\n")
	if !o.DoNotUseGeomPkg {
		Ff(pdf, "\\usepackage[margin=1.5cm,footskip=0.5cm]{geometry}\n")
	}

	// title and author
	hasTitleOrAuthor := false
	if o.Title != "" {
		Ff(pdf, "\n")
		Ff(pdf, "\\title{%s}\n", o.Title)
		hasTitleOrAuthor = true
	}
	if o.Author != "" {
		Ff(pdf, "\\author{%s}\n", o.Author)
		hasTitleOrAuthor = true
	}

	// begin document
	Ff(pdf, "\n")
	Ff(pdf, "\\begin{document}\n")
	if hasTitleOrAuthor {
		Ff(pdf, "\\maketitle\n")
	}

	// buffer
	if o.buffer != nil {
		Ff(pdf, "%v\n", o.buffer)
	}

	// extra LaTeX commands
	if extra != nil {
		Ff(pdf, "\n%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%% extra commands %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%\n\n")
		Ff(pdf, "%v\n", extra)
	}

	// end document
	Ff(pdf, "\n")
	Ff(pdf, "\\end{document}\n")

	// write TeX file
	fn := fnkey + ".tex"
	WriteFileD(dirout, fn, pdf)

	// run pdflatex
	if !o.DoNotGeneratePDF {
		_, err = RunCmd(false, "pdflatex", "-interaction=batchmode", "-halt-on-error", "-output-directory="+dirout, fn)
		if err != nil {
			if !o.DoNotShowMessages {
				PfRed("file <%s/%s> generated\n", dirout, fn)
			}
			return
		}
		if !o.DoNotShowMessages {
			PfBlue("file <%s/%s.pdf> generated\n", dirout, fnkey)
		}
	}
	return
}

// fixDefaults fix default values
func (o *Report) fixDefaults() {

	// default table positioning key
	if o.TablePos == "" {
		o.TablePos = "h"
	}

	// default number formatting string
	if o.NumFmt == "" {
		o.NumFmt = "%g"
	}

	// default table fontsize string
	if o.TableFontSz == "" {
		//o.TableFontSz = "\\scriptsize"
	}

	// default table column separation in 'em'; e.g. 0.5 =>
	if o.TableColSep <= 0 {
		o.TableColSep = 0.5
	}
}

// TexNum returns a string representation in TeX format of a real number.
// scientificNotation:
//   peforms the conversion of numbers into scientific notation where
//   the exponent notation with e{+-}{##} is converted into \cdot 10^{{+-}##}
func TexNum(fmt string, num float64, scientificNotation bool) (l string) {
	if fmt == "" {
		fmt = "%g"
	}
	l = Sf(fmt, num)
	if scientificNotation {
		s := strings.Split(l, "e-")
		if len(s) == 2 {
			e := s[1]
			if e == "00" {
				l = s[0]
				return
			}
			if e[0] == '0' {
				e = string(e[1])
			}
			l = s[0] + "\\cdot 10^{-" + e + "}"
		}
		s = strings.Split(l, "e+")
		if len(s) == 2 {
			e := s[1]
			if e == "00" {
				l = s[0]
				return
			}
			if e[0] == '0' {
				e = string(e[1])
			}
			l = s[0] + "\\cdot 10^{" + e + "}"
		}
	}
	return
}

// auxiliary //////////////////////////////////////////////////////////////////////////////////////

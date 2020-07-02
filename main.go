// Evernote2md is a cli tool to convert Evernote notes exported in *.enex format
// to a directory with markdown files.
//
// Usage:
//   evernote2md <file> [-o <outputDir>]
//
// If outputDir is not specified, current workdir is used.
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
	"github.com/integrii/flaggy"
	"github.com/wormi4ok/evernote2md/encoding/enex"
	"github.com/wormi4ok/evernote2md/file"
	"github.com/wormi4ok/evernote2md/internal"
)

var version = "dev"

func main() {
	var input string
	var outputDir = filepath.FromSlash("./notes")
	var outputOverride string

	flaggy.SetName("evernote2md")
	flaggy.SetDescription(" Convert Evernote notes exported in *.enex format to markdown files")
	flaggy.SetVersion(version)

	flaggy.AddPositionalValue(&input, "input", 1, true, "Evernote export file")
	flaggy.AddPositionalValue(&outputDir, "output", 2, false, "Output directory")
	flaggy.String(&outputOverride, "o", "outputDir", "Directory where markdown files will be created")

	flaggy.DefaultParser.ShowHelpOnUnexpected = false
	flaggy.DefaultParser.AdditionalHelpPrepend = "http://github.com/wormi4ok/evernote2md"

	flaggy.Parse()

	if len(outputOverride) > 0 {
		outputDir = outputOverride
	}

	run(input, outputDir)
}

func run(input, output string) {
	i, err := os.Open(input)
	failWhen(err)

	export, err := enex.Decode(i)
	failWhen(err)

	err = i.Close()
	failWhen(err)

	err = os.MkdirAll(output, os.ModePerm)
	failWhen(err)

	progress := pb.StartNew(len(export.Notes))
	progress.SetTemplateString(`Notes: {{counters .}} {{bar . "[" "=" ">" "_" "]" }} {{percent .}} {{etime .}}`)

	n := export.Notes
	for i := range n {
		md, err := internal.Convert(&n[i])
		failWhen(err)
		mdFile := filepath.FromSlash(output + "/" + file.BaseName(n[i].Title) + ".md")
		f, err := os.Create(mdFile)
		failWhen(err)
		_, err = io.Copy(f, bytes.NewReader(md.Content))
		failWhen(err)
		for _, res := range md.Media {
			err = file.Save(output+"/"+string(res.Type), res.Name, bytes.NewReader(res.Content))
			failWhen(err)
		}
		err = f.Close()
		failWhen(err)
		progress.Increment()
	}
	progress.Finish()
	fmt.Println("Done!")
}

func failWhen(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

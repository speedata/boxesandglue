package main

import (
	"fmt"
	"os"

	"github.com/speedata/optionparser"
)

func dothings() error {
	op := optionparser.NewOptionParser()
	op.Command("genpatterns", "Create patterns")
	op.Command("updatefonts", "Update font files")
	err := op.Parse()
	if err != nil {
		return err
	}

	if len(op.Extra) != 1 {
		op.Help()
		return nil
	}
	fmt.Println(op.Extra[0])
	switch op.Extra[0] {
	case "genpatterns":
		return createPatterns()
	case "updatefonts":
		return updatefonts()
	}
	return nil
}

func main() {
	if err := dothings(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

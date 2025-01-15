package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const TplBuildVersion = "${BUILD_VERSION}"

const TemplatesToCompileDirPath = "templates_to_compile"

var BuildVersion = "n/a"

func main() {
	fmt.Printf("Started Build Assistant. App version: '%s'\n", BuildVersion)

	err := filepath.Walk("./"+TemplatesToCompileDirPath, visit)
	if err != nil {
		panic(err)
	}
}

func visit(path string, fi os.FileInfo, err error) error {
	if err != nil {
		fmt.Printf("Error: '%s', path: '%s'\n", err, path)
		return err
	}

	if fi.IsDir() {
		return nil
	}

	fmt.Printf("Opened tpl file: '%s'\n", path)

	matched, err := filepath.Match("*.*", fi.Name())

	if err != nil {
		fmt.Printf("Error matching file name: '%s', path: '%s'\n", err, path)

		panic(err)
	}

	if matched {
		read, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}

		newContents := strings.Replace(string(read), TplBuildVersion, BuildVersion, -1)

		err = os.WriteFile(path, []byte(newContents), 0)
		if err != nil {
			fmt.Printf("Error writing file: '%s', path: '%s'\n", err, path)

			panic(err)
		}

	}

	fmt.Printf("Processed tpl file: '%s'\n", path)

	return nil
}

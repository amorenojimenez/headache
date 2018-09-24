package header

import (
	"bufio"
	. "bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Insert(config *configuration) {
	for _, includePattern := range config.Includes {
		matches, err := filepath.Glob(includePattern)
		if err != nil {
			panic(err)
		}
		insertInMatchedFiles(config, matches)
	}
}

func insertInMatchedFiles(config *configuration, files []string) {
	for _, file := range files {
		bytes, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}

		if HasPrefix(bytes, []byte(config.HeaderContents)) {
			continue
		}

		newContents := append([]byte(fmt.Sprintf("%s%s", config.HeaderContents, "\n")), bytes...)
		var writer = config.writer
		if writer == nil {
			openFile, err := os.Open(file)
			if err != nil {
				panic(err)
			}
			defer openFile.Close()
			writer = bufio.NewWriter(openFile)
		}

		writer.Write(newContents)

	}
}
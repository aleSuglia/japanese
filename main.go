package main

import (
	"fmt"
	"github.com/aleSuglia/japanese/base"
	"github.com/aleSuglia/japanese/retriever"
	"io/ioutil"
	"os"
)

func main() {

	for i := 9; i <= 15; i++ {
		fileName := ""

		if i > 9 {
			fileName = fmt.Sprintf("/home/asuglia/dblp_results/20%d.json", i)
		} else {
			fileName = fmt.Sprintf("/home/asuglia/dblp_results/200%d.json", i)
		}

		dblpList, err := base.RetrieveDBLPList(fileName)

		if err != nil {
			fmt.Println(err)
		} else {
			content, err := retriever.DownloadEntryList(&dblpList)

			if err != nil {
				fmt.Println("Some errors while downloading bib: ", err)
			} else {
				writeFile := fmt.Sprintf("/home/asuglia/dblp_results/%d.bib", i)
				fmt.Println("Writing: ", writeFile)
				err := ioutil.WriteFile(writeFile, []byte(content), os.ModePerm)
				if err != nil {
					fmt.Println("ERROR: ", err)
				}
			}
		}

	}

}

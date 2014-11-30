package retriever

import (
	"code.google.com/p/go.net/html"
	"errors"
	"fmt"
	"github.com/aleSuglia/japanese/base"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Used when a specific HTML tag isn't found
type TagNotFoundError struct {
	errorMsg string
}

func (e TagNotFoundError) Error() string {
	return e.errorMsg
}

type Downloader interface {
	Download() (string, error)
}

func downloadBib(basicUrl string) (string, error, int64) {
	bibRealURL, errUrl := getCorrectBibtex(basicUrl)

	fmt.Println(bibRealURL)

	if errUrl != nil {
		return "", errUrl, 0
	}

	resp, errResp := http.Get(bibRealURL)

	if errResp != nil {
		return "", errResp, 0
	}
	timeout, errConv := strconv.ParseInt(resp.Header.Get("Retry-After"), 10, 64)

	// correct conversion: Retry-After field is setted
	if errConv == nil {
		return "", nil, timeout
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err, 0
	}

	firstRef, errRef := getFirstRef(string(body))

	return firstRef, errRef, 0

}

func getCorrectBibtex(bibURL string) (string, error) {
	var urlTag, innetText string = "a", "download as .bib file"

	resp, err := http.Get(bibURL)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	doc, err := html.Parse(io.Reader(resp.Body))
	if err != nil {
		return "", err
	}

	bibRealURL := ""
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == urlTag {

			if n.FirstChild.Data == innetText {
				for _, a := range n.Attr {
					if a.Key == "href" {
						bibRealURL = a.Val
						break
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if bibRealURL == "" {
		return bibRealURL, TagNotFoundError{"Unable to find 'meta content' tag in HTML page"}
	}

	return bibRealURL, nil

}

func getFirstRef(bibRef string) (string, error) {
	atSign := strings.Index(bibRef, "@")
	if atSign == -1 {
		fmt.Println(bibRef)
		return "", errors.New("'@' not found in string")
	}

	firstBracket := strings.Index(bibRef, "}\n\n@")

	if firstBracket == -1 {
		firstBracket = len(bibRef)
	} else {
		firstBracket += 1
	}
	return bibRef[atSign:firstBracket], nil

}

func DownloadEntryList(dblpList *base.DBLPList) (string, error) {
	completeFile := ""

	for i := 0; i < len(dblpList.HitsList); {
		content, err, timeout := downloadBib(dblpList.HitsList[i].Url)

		if err != nil {
			return "", err
		}

		if timeout == 0 {
			completeFile += content + "\n"
			i++
		} else {
			time.Sleep(time.Duration(timeout) * time.Second)
		}

	}

	return completeFile, nil

}

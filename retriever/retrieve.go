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

/**
Downloads the specific BibTex entry from the specified URL.
An error is returned if something is wrong during the download and
if the DBLP website doesn't accept any other request, the last parameter
represents the number of seconds that you need to wait before any
other request will be accepted.
*/
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

/**
Gets the specific URL which contains the BibTex representation
for the given pubblication from the URL of the page that contains it.
An error is returned if the download process is wrong.
*/
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

/**
Parses the specified string and returns only the first
bib element (of two of them are specified, only the first one represents
an article or a pubblication).

An error is returned if something is wrong in the specified content.

*/
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

/**
Given a list of DBLP pubblication returns a single string which contains
the whole ".bib" file.

The error is not "nil" if one of the entries generates an error.
*/
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

package base

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type titleType struct {
	Ee   string `json:"@ee"`
	Text string `json:"text"`
}

type venueType struct {
	Url        string `json:"@url"`
	Conference string `json:"@conference"`
	Pages      string `json:"@pages"`
	Text       string `json:"text"`
}

type InfoEntry struct {
	Authors struct {
		Author []string `json:"author"`
	} `json:"authors"`
	Title       titleType         `json:"title"`
	Venue       map[string]string `json:"venue"`
	Year        string            `json:"year"`
	ArticleType string            `json:"type"`
}

type DBLPEntry struct {
	Score string    `json:"@score"`
	Id    string    `json:"@id"`
	Info  InfoEntry `json:"info"`
	Url   string    `json:"url"`
}

type DBLPList struct {
	HitsList []DBLPEntry `json:"hit"`
}

/**
Given a filename of the specific file which contains the results from DBLP
in JSON format, it extracts a list of structured record that can be
easily manipulated by the program. In case of error, the function returns
a second value that it's not nil.
*/
func RetrieveDBLPList(jsonFileName string) (DBLPList, error) {
	var dblpEntries DBLPList

	fileContent, readError := ioutil.ReadFile(jsonFileName)

	if readError != nil {
		return DBLPList{}, readError
	}

	errJson := json.Unmarshal(fileContent, &dblpEntries)

	if errJson != nil {
		return DBLPList{}, errJson
	}

	return dblpEntries, nil

}

type ElementConverter interface {
	Convert() string
}

/**
Converts a DBLP paper entry in BibTex format according to the paper's type
Returns the formatted BibTex strings related to the specified entry.
*/
func (entry *DBLPEntry) Convert() string {
	authorsList := ""

	authorsLen := len(entry.Info.Authors.Author)

	for pos, elem := range entry.Info.Authors.Author {
		if pos == (authorsLen - 1) {
			continue
		}

		authorsList += elem + ","
	}

	if authorsLen != 0 {
		authorsList += entry.Info.Authors.Author[authorsLen-1]
	}

	formatString := ""
	venueString := ""

	if entry.Info.ArticleType == "inproceedings" || entry.Info.ArticleType == "proceedings" {
		formatString =
			"@%s{%s,\nauthor={%s},\ntitle={%s},\nbooktitle={%s},\npages={%s},\nyear={%s},\nurl={%s}\n}"
		venueString = entry.Info.Venue["@conference"]
	} else if entry.Info.ArticleType == "incollection" {
		formatString =
			"@%s{%s,\nauthor={%s},\ntitle={%s},\nbook={%s},\npages={%s},\nyear={%s},\nurl={%s}\n}"
		venueString = entry.Info.Venue["@conference"]

	} else if entry.Info.ArticleType == "article" {
		formatString =
			"@%s{%s,\nauthor={%s},\ntitle={%s},\njournal={%s},\npages={%s},\nyear={%s},\nurl={%s}\n}"
		venueString = entry.Info.Venue["@journal"]
	} else {
		formatString =
			"@%s{%s,\nauthor={%s},\ntitle={%s},\nbooktitle={%s},\npages={%s},\nyear={%s},\nurl={%s}\n}"
		venueString = entry.Info.Venue["@conference"]

	}

	return fmt.Sprintf(formatString,
		entry.Info.ArticleType, entry.Id, authorsList,
		entry.Info.Title.Text, venueString, entry.Info.Venue["@pages"],
		entry.Info.Year, entry.Url)
}

/**
Returns a string containing all the entries in BibTex format.
*/
func (entryList DBLPList) GetBibtex() string {
	finalBib := ""

	for _, entry := range entryList.HitsList {
		finalBib += entry.Convert() + "\n"

	}

	return finalBib
}

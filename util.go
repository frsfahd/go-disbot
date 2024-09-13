package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/PuerkitoBio/goquery"
)

func decodeFromReader(r io.Reader, res interface{}) error {

	dec := json.NewDecoder(r)
	err := dec.Decode(&res)
	if err != nil {
		return err
	}
	return nil

}

func extractTitle(r io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", err
	}

	// Remove script and style tags.
	doc.Find("script, style").Remove()

	// Extract the title text.
	title := doc.Find("title").Text()

	return title, nil
}

func updateDataset(filepath string, newData challenge) error {
	var listOfChallenges challenges
	f, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	r := bytes.NewBuffer(f)

	err = decodeFromReader(r, &listOfChallenges)
	if err != nil {
		return err
	}

	listOfChallenges.Challenges = append(listOfChallenges.Challenges, newData)

	updatedData, err := json.MarshalIndent(listOfChallenges, "", " ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath, updatedData, 0666)
	// writer := bufio.NewWriter(f)
	// _, err = writer.Write(updatedData)
	if err != nil {
		return err
	}

	// err = writer.Flush()
	// if err != nil {
	// 	return err
	// }

	return nil
}

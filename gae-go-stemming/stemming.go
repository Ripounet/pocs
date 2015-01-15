package stemming

import (
	"appengine"
	"appengine/search"
	"fmt"
	"net/http"
	"strconv"
)

func init() {
	http.HandleFunc("/", handle)
}

func handle(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	//
	// Put some items in the index.
	//
	fmt.Fprintln(w, "Indexing :")
	for i, sentence := range []string{
		"I have a dog.",
		"I like dogs.",
		"I have a cat.",
		"I like cats.",
	} {
		docId := strconv.Itoa(i)
		err := writeToIndex(c, docId, sentence)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%q \n", sentence)
	}
	fmt.Fprintln(w)

	//
	// Search the index.
	//
	// The stemmed queries (with ~) are expected to hit more results.
	//
	fmt.Fprintln(w, "Searching :")
	for _, queryValue := range []string{
		"dog",
		"~dog",
		"cat",
		"~cat",
	} {
		hits, err := searchInIndex(c, queryValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%v => %v \n", queryValue, hits)
	}
}

// This sample struct contains only 1 string field Bulk.
// Bulk will will automatically tokenized, etc.
type searchableDoc struct {
	Bulk string
}

func writeToIndex(c appengine.Context, docId, text string) error {
	index, err := search.Open("my-index")
	if err != nil {
		return err
	}
	doc := &searchableDoc{
		Bulk: text,
	}
	_, err = index.Put(c, docId, doc)
	return err
}

func searchInIndex(c appengine.Context, query string) ([]string, error) {

	index, err := search.Open("my-index")
	if err != nil {
		return nil, err
	}
	sentences := make([]string, 0)
	it := index.Search(c, query, nil)
	for {
		var hit searchableDoc
		_, err := it.Next(&hit)
		if err == search.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		sentences = append(sentences, hit.Bulk)
	}
	return sentences, nil
}

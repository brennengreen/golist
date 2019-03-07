package main

import (
	"fmt"
	"net/http"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func main() {
	resp, err := http.Get("https://lexington.craigslist.org/search/sya?")
	if err != nil {
		panic(err)
	}
	root, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	articleMatcher := func(n *html.Node) bool {
		if n.DataAtom == atom.A && n.Parent != nil {
			return scrape.Attr(n.Parent, "class") == "result-info"
		}
		return false
	}

	findChildByClass := func(parent *html.Node, childAttrType string) (*html.Node, bool) {
		matcher := func(n *html.Node) bool {
			if n.DataAtom == atom.Span && n.Parent != nil {
				return scrape.Attr(n, "class") == childAttrType
			}
			return false
		}

		nodes, found := scrape.Find(parent, matcher)
		return nodes, found
	}

	findPrice := func(parent *html.Node) string {
		price, found := findChildByClass(parent, "result-price")
		if found == false {
			return "nil"
		}
		return scrape.Text(price)
	}

	articles := scrape.FindAll(root, articleMatcher)
	
	for i, article := range articles {
		price := findPrice(article.Parent)
		fmt.Printf("%2d %s (%s)\n", i, scrape.Text(article), price)
	}
		
}


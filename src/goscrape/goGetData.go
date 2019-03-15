package goscrape

import (
	// Standard
	//"fmt"
	"net/http"
	"strconv"	

	// Third-Party
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)
// Item export
type Item struct {
	Title string
	Price int
	Link string
}

// GetData returns scraping data
func GetData(pageURL string) []Item {
	// Process HTTP Request into the root of page in HTML Node Form
	httpReq := func(reqPath string) *html.Node {
		resp, err := http.Get(reqPath)
		if err != nil {
			panic(err)
		}
		root, err := html.Parse(resp.Body)
		if err != nil {
			panic(err)
		}
		
		return root
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

	findPrice := func(parent *html.Node) int {
		price, found := findChildByClass(parent, "result-price")
		if found == false {
			return 0
		}
		priceInt,_ := strconv.Atoi(scrape.Text(price)[1:])
		return priceInt
	}

	articles := scrape.FindAll(httpReq(pageURL), articleMatcher)

	var items []Item

	for _, article := range articles {
		price := findPrice(article.Parent)
		title := scrape.Text(article)
		link := scrape.Attr(article, "href")
		item := Item{Title: title, Price:price, Link:link}
		//fmt.Printf("%2d %s ($%d)\n", i, scrape.Text(article), price)
		items = append(items, item)
	}

	return items	
}


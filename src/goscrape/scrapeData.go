package goscrape

import (
	"fmt"
	"database/sql"
	//"io/ioutil"
	"strconv"
	"strings"
	"os"

	_ "github.com/lib/pq"
)

// Brand : Training Data
type Brand struct {
	BrandName string
}

func ScrapeData() int {
	count := 0
	// connect() uses internal settings not publicly shared to connect to a RWS Postgres Database
	database := connect()
	defer database.Close()
	/*******************************************************************************************
	* GoScrape is a custom library that utilizes github.com/yhat/scrape in order to scrape the
	* first page of the specified craigslist category
	*******************************************************************************************/
	items := GetData("https://lexington.craigslist.org/d/computers/search/sya")

	var brands []string
	brandRows, err := database.Query("SELECT BrandName FROM Brand")
	if err != nil {
		panic(err)
	}
	defer brandRows.Close()
	
	for brandRows.Next() {
		var brand string
		err := brandRows.Scan(&brand)
		if err != nil {
			continue
		}
		brands = append(brands, brand)
	}


	for _, item := range items {


		var queryLink string
		err := database.QueryRow("SELECT link FROM Posts WHERE link=$1", item.Link).Scan(&queryLink)
		switch {	
		// IF this case is true that means the current item is not already stored inside our database
		case err == sql.ErrNoRows:
			isMatch, brand := checkMatch(brands, item.Title)
			if isMatch {
				itemName := getItemName(brand, item.Title)
				_, err := database.Exec("INSERT INTO Items (brandname, name) VALUES ($1, $2)", brand, itemName)
				if err != nil {
					fmt.Println(itemName, " already exists")
				}
				count = count +1
				fmt.Printf("Adding %s to database..", item.Link)
				_, err = database.Exec("INSERT INTO Posts (title, name, price, link) VALUES ($1, $2, $3, $4)", item.Title, brand, item.Price, item.Link)
				if err != nil {
					panic(err)
				}
				var avg int
				avgErr := database.QueryRow("SELECT AVG(avgprice) FROM Items WHERE brandname=$1", brand).Scan(&avg)
				switch {
				case avgErr == sql.ErrNoRows:
					fmt.Println("No data on brand")
				default:
					if avg == 0 {
						
					} else {
						if (100 - ((item.Price/avg)*100)) >= 25 {
							fmt.Printf("Notifying admin...")
						}
					}
					
				}

			} else {
				invalidErr := database.QueryRow("SELECT * FROM InvalidPosts WHERE link=$1", item.Link).Scan()
				switch {
				case invalidErr == sql.ErrNoRows:
					_, err = database.Exec("INSERT INTO InvalidPosts (link) VALUES ($1)", item.Link)
				default:
					fmt.Println("InvalidPost already exists in database.")
				}
			}

		// The only case we don't want :(
		// If this case is true that means the Query has resulted in an error and needs to be handled properly
		case err != nil:
			panic(err)

		// If default is true that means the query returned a result, signifying our item already exists in the database
		default:
			fmt.Println("Found ", item.Link)
			continue;
		}
	}	
	return count
}

func UpdatePrices() {

}

func connect() *sql.DB {
	db ,err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}

func checkMatch(keywords []string, matchingString string) (bool, string) {
	tokenizedStr := strings.Split(matchingString, " ")
	for _, token := range tokenizedStr {
		for _, key := range keywords {
			if strings.ToLower(token) == strings.ToLower(key) {
				fmt.Println(key)
				return true, key
			}
		}
	}
	return false, ""
}

func getItemName(brand string, postTitle string) string {
	tokenizedStr := strings.Split(postTitle, " ")
	for i, token := range tokenizedStr {
		if strings.ToLower(token) == strings.ToLower(brand) {
			return tokenizedStr[i+1]
		}
	}
	return ""
}
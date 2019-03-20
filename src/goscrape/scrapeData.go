package goscrape

import (
	"fmt"
	"database/sql"
	//"io/ioutil"
	"strconv"
	"strings"

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
				count = count +1
				fmt.Printf("Adding %s to database..", item.Link)
				_, err := database.Exec("INSERT INTO Posts (title, name, price, link) VALUES ($1, $2, $3, $4)", item.Title, brand, item.Price, item.Link)
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

	// Looping through rows example 
	fmt.Printf("\n\nUpdating database!\n")
	var (names []string)
	rows, err := database.Query("SELECT name FROM Items")
	if err != nil {
		panic(err)
	} 
	defer rows.Close()
	for rows.Next() {
		var (
			name string
		)
		err := rows.Scan(&name)
		if err != nil {
			panic(err)
		}
		names = append(names, name)
	}

	for _, name := range names {
		priceRows, err := database.Query("SELECT AVG(price) FROM Posts WHERE name=$1", name)
		if err != nil {
			panic(err)
		}
		defer priceRows.Close()

		var avg []uint8
		for priceRows.Next() {
			err := priceRows.Scan(&avg)
			if err != nil {
				panic(err)
			}
		}
		avgStr := string(avg)
		avgFloat,err := strconv.ParseFloat(avgStr,64)
		if err != nil {
			avgFloat = 0.0
		} else {
			if avgFloat > 0.0 {
				_, err := database.Exec("UPDATE Items SET avgprice=$1 WHERE name=$2", int(avgFloat), name)
				if err != nil {
					panic(err)
				}
				fmt.Println(name, ": ", int(avgFloat))
			} else {
			}
		}
	}
	return count
}

func connect() *sql.DB {
	dbStr := "host=learn-this-shit.clj9pnmafa0l.us-east-2.rds.amazonaws.com port=5432 user=brenneng password=brodog12 dbname=learning_shit sslmode=disable"
	db ,err := sql.Open("postgres", dbStr)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}

/*func returnFileData(filePath string) string {
	f, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	return string(f)
}*/

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
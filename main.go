/***********************************************************************************
* Title: GOlist
* Author: Brennen Green (wwww.github.com/brennengreen)
* Description: A bot that is ran periodally to scrape the first page of the computers
* 	category of my local craigslist and inform me if something is posted with a price 
* 	of whis is significantly(25%) less than similar items.
* Link: www.github.com/brennengreen/golist
************************************************************************************/

package main

import (
	"fmt"
	"net/http"
	"os"
	"log"

	"github.com/brennengreen/golist/src/goscrape"

)


func main() {
	port := ":"+os.Getenv("PORT")
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(port, nil))
	
}

func handler(w http.ResponseWriter, r *http.Request) {
	count := goscrape.ScrapeData()
	fmt.Fprintln(w, "Found ", count, " items to add to database")

}



package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/gocolly/colly"
)

type PokemonProduct struct {
	url, image, name, price string
}

////////////////////////////////
//webscrape using colly
//1. enter user input
//2. user enters url
//3. user enters what they are looking for
//4. create webscrape goroutine that will scrap
//5. data is then returned to user in an html file

func initWebscraper() *colly.Collector {
	scraper := colly.NewCollector()
	return scraper
}

func toCsv(pokemonproducts []PokemonProduct) {

	var csv_name string
	//takes valid csv file name
	for {
		fmt.Print("Please enter a name for the csv file\n")
		_, err := fmt.Scanln(&csv_name)
		if err != nil {
			fmt.Print("Invalid Name or already used.\n Please Enter a different Name: ", err)
		} else {
			break
		}
	}
	//creates a new csv file
	csv_file, err := os.Create(fmt.Sprintf("%s.csv", csv_name))

	if err != nil {
		log.Fatalln("Pokemon.csv could not be created", err)
	}
	defer csv_file.Close()

	csv_writer := csv.NewWriter(csv_file)

	//define header fields
	headers := []string{
		"url",
		"image",
		"name",
		"price",
	}

	csv_writer.Write(headers)
	//add webscraped values to csv
	for _, pokemon := range pokemonproducts {
		record := []string{
			pokemon.url,
			pokemon.image,
			pokemon.name,
			pokemon.price,
		}

		csv_writer.Write(record)
	}
	defer csv_writer.Flush()
}

func main() {

	fmt.Println("Starting webscrape...")
	collector := initWebscraper()

	var pkproducts []PokemonProduct
	//enter a url for webscraping
	var url string
	var err error

	for {
		fmt.Print("Please enter a url to scrape\n")
		_, err = fmt.Scanln(&url)
		if err != nil {
			fmt.Print("Invalid type, Try again: ", err)
		} else {
			break
		}
	}

	//callbacks
	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting {}", url)
	})

	collector.OnError(func(_ *colly.Response, err error) {
		log.Println("Request failed: ", err)
	})

	collector.OnResponse(func(r *colly.Response) {
		fmt.Println("Successful visit to {}", url)
	})
	collector.OnHTML("li.product", func(e *colly.HTMLElement) {

		pkproduct := PokemonProduct{}
		pkproduct.url = e.ChildAttr("a", "href")
		pkproduct.image = e.ChildAttr("img", "src")
		pkproduct.name = e.ChildText("h2")
		pkproduct.price = e.ChildText(".price")

		pkproducts = append(pkproducts, pkproduct)
		//store in csv
		toCsv(pkproducts)
	})
	collector.OnScraped(func(r *colly.Response) {
		fmt.Println(r.Request.URL, " scraped!")
	})
	//challenge, incorporate a db microservice to store these tag information in a database
	collector.Visit(url)

}

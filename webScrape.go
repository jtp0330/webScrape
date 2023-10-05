package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gocolly/colly"
	_ "github.com/lib/pq"
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

func connectToDB() *sql.DB {
	//var db_pwd string
	fmt.Println("Starting DB Connection...")
	//fmt.Println("Enter database password: ")
	//fmt.Scan(&db_pwd)
	//connStr := fmt.Sprintf("postgres://postgres:%s@localhost/pokemon_data", db_pwd)
	connStr := "postgres://postgres:p0kem0n@localhost/pokemon_data?sslmode=disable"

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected to database: Pokemon_data")

	return db
}

// meant to be called while a connection is established
// stores data entry into database
func storeScrapedPokemonData(db *sql.DB, entry PokemonProduct) {

	pk_price, _ := strconv.ParseFloat(entry.price[2:], 64)
	var pkid = 0

	//calculate new pkid based on latest entry from database or set to 0
	var rowCount int
	err := db.QueryRow(`SELECT COUNT(*) FROM pokemon_entries`).Scan(&rowCount)

	if err != nil {
		panic(err)
	}

	if rowCount > 0 {
		getLatestPkid := `SELECT pkid FROM pokemon_entries order by pkid desc limit 1`
		get_err := db.QueryRow(getLatestPkid).Scan(&pkid)

		if get_err != nil {
			panic(get_err)
		}
	}

	insertQuery := `INSERT INTO "pokemon_entries"("pkid","pk_name","pk_price","pk_url") VALUES($1,$2,$3,$4)`
	_, insert_err := db.Exec(insertQuery, pkid+1, entry.name, pk_price, entry.url)

	if insert_err != nil {
		panic(insert_err)
	}
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
		//toCsv(pkproducts)

		db := connectToDB()
		//store data in pokemon_data table in PokemonDB Database
		fmt.Println("Inserting Data into Database: Pokemon_data")
		for _, pkproduct := range pkproducts {
			storeScrapedPokemonData(db, pkproduct)
		}
		db.Close()
	})
	collector.OnScraped(func(r *colly.Response) {
		fmt.Println(r.Request.URL, " scraped!")
	})
	//challenge, incorporate a db microservice to store these tag information in a database
	collector.Visit(url)

}

package main

// Import all the packages
import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/lib/pq"
	"github.com/tidwall/gjson"

	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/spf13/viper"
)

// Defining all the structures to Unmarshall json data
type sports struct {
	gorm.Model
	Success bool `json:"success"`
	Data    []struct {
		Key     string `json:"key"`
		Active  bool   `json:"active"`
		Group   string `json:"group"`
		Details string `json:"details"`
		Title   string `json:"title"`
	} `json:"data"`
}

type matchOdds struct {
	Success bool `json:"success"`
	Data    []struct {
		ID           string   `json:"id"`
		SportKey     string   `json:"sport_key"`
		SportNice    string   `json:"sport_nice"`
		Teams        []string `json:"teams"`
		CommenceTime int      `json:"commence_time"`
		HomeTeam     string   `json:"home_team"`
		Sites        []struct {
			SiteKey    string `json:"site_key"`
			SiteNice   string `json:"site_nice"`
			LastUpdate int    `json:"last_update"`
			Odds       struct {
				H2H []float64 `json:"h2h"`
			} `json:"odds"`
		} `json:"sites"`
		SitesCount int `json:"sites_count"`
	} `json:"data"`
}

// creating structures for database
type newmatchData struct {
	ID           string         `json:"id" gorm:"primary_key"`
	SportKey     string         `json:"sport_key"`
	SportNice    string         `json:"sport_nice"`
	Teams        pq.StringArray `json:"teams" gorm:"type:text[]"`
	CommenceTime int            `json:"commence_time"`
	HomeTeam     string         `json:"home_team"`
	Upcoming     bool           `gorm:"type:boolean" gorm:"defult=false"`
}

type newSiteData struct {
	MatchID    string          `gorm:"type:text"`
	SiteKey    string          `json:"site_key"`
	SiteNice   string          `json:"site_nice"`
	LastUpdate int             `json:"last_update"`
	Odds       pq.Float64Array `gorm:"type:float[]"`
}
type sportsData struct {
	Key     string `json:"key"`
	Active  bool   `json:"active"`
	Group   string `json:"group"`
	Details string `json:"details"`
	Title   string `json:"title"`
}

// initializeEnvVariables with Viper can also use os.GetEnv("variable_name") to get system env variables
func initializeEnvVariables() {
	viper.SetConfigName("config")

	// Set the path to look for the configurations file
	viper.AddConfigPath(".")

	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()

	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file, %s", err)
	}

}

// function to get all the available sports
func fetchAllSports() (resp *http.Response) {

	resp, err := http.Get("https://api.the-odds-api.com/v3/sports/?apiKey=" + viper.Get("API_KEY").(string))
	if err != nil {
		log.Fatalln(err)
	}
	return

}

// Using gjson instead of directly unmarshalling wanted to explore few libraries
func getAllKeys(resp *http.Response) (value gjson.Result, body []byte) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	sb := string(body)
	value = gjson.Get(sb, "data.#.key")
	return
}

// TO clear database if needed
func clearAllSportsData(db *gorm.DB) {

	db.Exec("DELETE FROM SPORTS_DATA")
}

// Update Sports Data in database
func updateSportsData(body []byte) {
	result := sports{}
	err := json.Unmarshal(body, &result)
	if err != nil {
		log.Println("penik")
	}
	db, err := gorm.Open(viper.Get("DB.type").(string), "host="+viper.Get("DB.host").(string)+" port="+viper.Get("DB.port").(string)+" user="+viper.Get("DB.user").(string)+" dbname="+viper.Get("DB.dbname").(string)+" sslmode="+viper.Get("DB.sslmode").(string)+" password="+viper.Get("DB.password").(string)+"")
	if err != nil {
		panic("failed to connect database")
	}

	defer db.Close()
	db.AutoMigrate(&sportsData{})
	for index := range result.Data {
		sportsEvent := sportsData{Key: result.Data[index].Key, Active: result.Data[index].Active, Group: result.Data[index].Group, Details: result.Data[index].Details, Title: result.Data[index].Title}
		db.Create(sportsEvent)
	}
	log.Println("Update Sports Event in database")
}

// Wrapper function for fetching all the odds

func fetchAllOdds(keys gjson.Result) {
	for _, key := range keys.Array() {
		fetchOddsSport(key.Str)
	}
	log.Println("ALL Data Updated")
}

// Calls api based on keys passed could have used better naming though
func fetchOddsSport(key string) {
	resp, err := http.Get("https://api.the-odds-api.com/v3/odds/?apiKey=" + viper.Get("API_KEY").(string) + "&sport=" + key + "&region=uk&mkt=h2h")
	if err != nil {
		log.Fatalln(err)
	}
	body, err2 := io.ReadAll(resp.Body)

	if err2 != nil {
		log.Fatalln(err)
	}
	newResult := matchOdds{}
	err = json.Unmarshal(body, &newResult)
	if err != nil {
		log.Println("penik")
	}
	db, err := gorm.Open(viper.Get("DB.type").(string), "host="+viper.Get("DB.host").(string)+" port="+viper.Get("DB.port").(string)+" user="+viper.Get("DB.user").(string)+" dbname="+viper.Get("DB.dbname").(string)+" sslmode="+viper.Get("DB.sslmode").(string)+" password="+viper.Get("DB.password").(string)+"")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()
	// clearAllOddsData(db)

	updateDatabase(newResult, db, key)
}

// Updating database from unmarshalled json
func updateDatabase(result matchOdds, db *gorm.DB, key string) {
	var upcoming bool
	if key == "upcoming" {
		upcoming = true
	} else {
		upcoming = false
	}
	//autoMigrate creates tables based on structures
	db.AutoMigrate(&newmatchData{}, &newSiteData{})
	// iterate over each matches and sites
	for index := range result.Data {
		matchdata := newmatchData{ID: result.Data[index].ID, SportKey: result.Data[index].SportKey, SportNice: result.Data[index].SportNice, Teams: result.Data[index].Teams, CommenceTime: result.Data[index].CommenceTime, HomeTeam: result.Data[index].HomeTeam, Upcoming: upcoming}
		if db.Model(&matchdata).Where("ID = ?", result.Data[index].ID).Updates(&matchdata).RowsAffected == 0 {
			db.Create(&matchdata)
		}
		for newIndex := range result.Data[index].Sites {
			sitedata := newSiteData{MatchID: string(result.Data[index].ID), SiteKey: result.Data[index].Sites[newIndex].SiteKey, SiteNice: result.Data[index].Sites[newIndex].SiteNice, LastUpdate: result.Data[index].Sites[newIndex].LastUpdate, Odds: result.Data[index].Sites[newIndex].Odds.H2H}
			if db.Model(&sitedata).Where("MATCH_ID = ? AND SITE_KEY = ?", string(result.Data[index].ID), result.Data[index].Sites[newIndex].SiteKey).Updates(&sitedata).RowsAffected == 0 {
				db.Create(&sitedata)

			}
		}
	}
	log.Println("Odds Data updated")
}

// to clear all odds data
func clearAllOddsData(db *gorm.DB) {
	log.Println("CLEARNING OLD DATA")
	db.Exec("DELETE FROM NEW_SITE_DATA")
	db.Exec("DELETE FROM NEWMATCH_DATA")

}

func main() {
	initializeEnvVariables()
	response := fetchAllSports()
	keys, body := getAllKeys(response)
	updateSportsData(body)
	//delays for all matches and inplay matches can be set here
	delay := viper.Get("DELAY").(int)
	delay_upcoming := viper.Get("DELAY_UPCOMING").(int)
	allMatchesTicker := time.NewTicker(time.Duration(delay) * time.Minute)
	upcomingTicker := time.NewTicker(time.Duration(delay_upcoming) * time.Minute)

	for {
		select {
		case <-allMatchesTicker.C:
			fetchAllOdds(keys)
		case <-upcomingTicker.C:
			fetchOddsSport("upcoming") //to save all in-play matches use this function
		}
	}

}

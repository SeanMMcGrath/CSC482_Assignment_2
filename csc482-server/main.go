package main

import (
	"fmt"
	"net/http"
	"log"
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type statusResponse struct{
	Tbl table `json:"table"`
}

type table struct {
	TableName string `json:"table_name"`
	ItemCount int `json:"item_count"`
}

type Item struct {
	ID string `json:"username"`
	Acc Account `json:"account"`
}

//one per user account, accountId is encrypted and used to send more queries
type Account struct {
	ID            string `json:"id"`            //does not change
	AccountID     string `json:"accountId"`     //does not change
	Puuid         string `json:"puuid"`         //does not change
	Name          string `json:"name"`          //player can change this if they want, will mess up api requests if they do
	SummonerLevel int    `json:"summonerLevel"` //changes as player plays, increasing by 1 after an amount of games are played
	ChampionData  ChampMastery `json:"champData"`
}

type ChampMastery []struct { //multiple of these per account up to 143 MAX, some may return empty if champion has never ben played
	ChampionID                   int    `json:"championId"`     //does not change
	ChampionLevel                int    `json:"championLevel"`  //from 0 to 7, can go up but not down
	ChampionPoints               int    `json:"championPoints"` //number indicating how much this champion has been played, higher number = higher playtime
	LastPlayTime                 int64  `json:"lastPlayTime"`   //number indicating last time this champion was played by user
	ChampionPointsSinceLastLevel int    `json:"championPointsSinceLastLevel"`
	ChampionPointsUntilNextLevel int    `json:"championPointsUntilNextLevel"`
	ChestGranted                 bool   `json:"chestGranted"` //t/f
	TokensEarned                 int    `json:"tokensEarned"` //from 0 to 3
	SummonerID                   string `json:"summonerId"`   //connects back to summoner
}

const tableName = "smcgrat3_table"
const username = "TF_Blade"

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/smcgrat3/all", allHandler).Methods("GET")
	r.HandleFunc("/smcgrat3/status", statusHandler).Methods("GET")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", r))
}

func allHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Println("Connect Successful")
	item, _ := update()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
} //return everything in dynamoDB table

func statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Connect Successful")
	_, t := update()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
} //display table name and number of objects in table

func update() (Item, statusResponse){
	//update dynamodb tables
	var item Item
		sess, err := session.NewSession(&aws.Config{
			Region:                aws.String("us-east-1"),
		})
		if err != nil {
			fmt.Println("aws Session Error: ", err)
			fmt.Println("Update failed")
		}
		svc := dynamodb.New(sess)

		result, err := svc.GetItem(&dynamodb.GetItemInput{
    	TableName: aws.String(tableName),
    	Key: map[string]*dynamodb.AttributeValue{
				"username":{
					S: aws.String(username),
				},
			},
		})
		if err != nil {
			fmt.Println(err)
		}

		err = dynamodbattribute.UnmarshalMap(result.Item, &item)
		if err != nil {
    	panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
		}

		if  item.ID == "" {
			fmt.Println("username not found")
		}

		response := statusResponse{
			Tbl: table{
				TableName: tableName,
				ItemCount: 1,
			},
		}

		return item, response
}

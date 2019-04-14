package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/JamesPEarly/loggly"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Item struct {
	ID string `json:"username"`
	Acc       Account `json:"account"`
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

func main() {
	client := loggly.New("My_Project") //second string is tag

	//usernames := [5]string{"","","","",""}//usernames

	var linkHTTP = "https://"
	var server = "na1"                //na1 = north america, kr = korea
	var name = "TF_Blade"            //username
	var apiKey = os.Getenv("API_KEY") //new key needs to be generated every day
	var linkP1 = ".api.riotgames.com/lol/summoner/v4/summoners/by-name/"
	var linkP2 = "?api_key=" //link = linkP1 + myUsername + linkP2 + apiKey
	var linkP3 = ".api.riotgames.com/lol/champion-mastery/v4/champion-masteries/by-summoner/"


	fmt.Println("API_KEY: " + apiKey)
	if apiKey == "" {
		fmt.Println("No API key")
		os.Exit(1)
	}

	link1 := []string{linkHTTP, server, linkP1, name, linkP2, apiKey} // user Account

	for { //sends new set of requests every 5 min, 10 times(for now, can be any number)

		resp, err := http.Get(strings.Join(link1, ""))
		if err == nil {
			//send error to loggly
			err := client.EchoSend("info", "First API request sucessful.")
			if err != nil {
				fmt.Println("err: ", err)
			}
		} else {
			client.EchoSend("error", err.Error())
		}

		//defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			//send error to loggly
			client.EchoSend("error", err.Error())
		} else {
			//let loggly know it was sucesful
			err := client.EchoSend("info", "Body sucesfully retrieved.")
			if err != nil {
				client.EchoSend("error", err.Error())
			}
		}

		//turn body into struct
		var user Account

		err = json.Unmarshal(body, &user)

		if err != nil {
			//send error to loggly
			client.EchoSend("error", err.Error())
		}

		link2 := []string{linkHTTP, server, linkP3, user.ID, linkP2, apiKey} // ChampMastery

		resp, err = http.Get(strings.Join(link2, ""))
		if err != nil {
			//send error to loggly
			client.EchoSend("error", err.Error())
		}

		defer resp.Body.Close()

		body2, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			//send error to loggly
			client.EchoSend("error", err.Error())
		} else {
			//send raw data to loggly
			err = client.EchoSend("info", "Second API request sucessful.")
			if err != nil {
				client.EchoSend("error", err.Error())
			}
		}

		err = json.Unmarshal(body2, &user.ChampionData)

		printAccount(user)
		sess := session.Must(session.NewSession(&aws.Config{
			Region:                aws.String("us-east-1"),
		}))
		svc := dynamodb.New(sess)

		if err != nil {
			client.EchoSend("error", err.Error())
		} else {
			client.EchoSend("info", "Session started.")
		}
		item := Item{
			ID: name,
			Acc:       user,
		}

		av, err := dynamodbattribute.MarshalMap(item)

		if err != nil {
				fmt.Println("Got error marshalling map:")
				client.EchoSend("error", err.Error())
				os.Exit(1)
		}

		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String("smcgrat3_table"),
		}

		_, err = svc.PutItem(input)

		if err != nil {
			fmt.Println("Got error calling PutItem:")
			client.EchoSend("error", err.Error())
			os.Exit(1)
		}
		client.EchoSend("info", "Item successfully placed in table.")

		fmt.Println("Sleeping for 5 minutes before polling again")
		time.Sleep(5 * time.Minute) //limited to 20 requests every 1 seconds and 100 requests every 2 minutes per server
		fmt.Println("Waking up.")
	}
}

func printAccount(user Account) {
	fmt.Println("ID: ", user.ID)
	fmt.Println("Account ID: ", user.AccountID)
	fmt.Println("Puuid: ", user.Puuid)
	fmt.Println("Username: ", user.Name)
	fmt.Println("Summoner Level: ", user.SummonerLevel)
	printMasteries(user.ChampionData)
}

func printMasteries(masteries ChampMastery) {
	for i := 0; i < len(masteries); i++ { //print out all of the mastery information for an account
		fmt.Println("Champion ID: ", masteries[i].ChampionID)
		fmt.Println("Champion Level: ", masteries[i].ChampionLevel)
		fmt.Println("Champion Points: ", masteries[i].ChampionPoints)
		fmt.Println("Last Play Time: ", masteries[i].LastPlayTime)
		fmt.Println("Champion Points Since Last Level: ", masteries[i].ChampionPointsSinceLastLevel)
		fmt.Println("Champion Points Until Next Level: ", masteries[i].ChampionPointsUntilNextLevel)
		fmt.Println("Chest Granted: ", masteries[i].ChestGranted)
		fmt.Println("Tokens Earned: ", masteries[i].TokensEarned)
	}
}

package main

import (
  "fmt"
  "net/http"
  "io/ioutil"
  //"os"
  "encoding/json"
  "strings"
  "time"
  "github.com/segmentio/go-loggly"
)

type Account struct {//one per user account, accountId is encrypted and used to send more queries
	ID            string `json:"id"`//does not change
	AccountID     string `json:"accountId"`//does not change
	Puuid         string `json:"puuid"`//does not change
	Name          string `json:"name"`//player can change this if they want, will mess up api requests if they do
	SummonerLevel int    `json:"summonerLevel"`//changes as player plays, increasing by 1 after an amount of games are played
  ChampionData ChampMastery
}

type ChampMastery []struct {//multiple of these per account up to 143 MAX, some may return empty if champion has never ben played
  ChampionID                   int    `json:"championId"`//does not change
	ChampionLevel                int    `json:"championLevel"`//from 0 to 7, can go up but not down
	ChampionPoints               int    `json:"championPoints"`//number indicating how much this champion has been played, higher number = higher playtime
	LastPlayTime                 int64  `json:"lastPlayTime"`//number indicating last time this champion was played by user
	ChampionPointsSinceLastLevel int    `json:"championPointsSinceLastLevel"`
	ChampionPointsUntilNextLevel int    `json:"championPointsUntilNextLevel"`
	ChestGranted                 bool   `json:"chestGranted"`//t/f
	TokensEarned                 int    `json:"tokensEarned"`//from 0 to 3
	SummonerID                   string `json:"summonerId"`//connects back to summoner
}

func main(){

  logToLoggly := loggly.New("fc53e471-a824-4e4c-86cc-5359555efb1a", "project2")//second string is tag

  //usernames := [5]string{"","","","",""}//usernames


  var linkHttp string = "https://"
  var server string = "na1"//na1 = north america, kr = korea
  var myUsername string = "Kýu"//username
  var apiKey string = ""//new key needs to be generated every few days, due to Riot's policy it can not be public
  var linkP1 string = ".api.riotgames.com/lol/summoner/v4/summoners/by-name/"
  var linkP2 string = "?api_key="//link = linkP1 + myUsername + linkP2 + apiKey
  var linkP3 = ".api.riotgames.com/lol/champion-mastery/v4/champion-masteries/by-summoner/"

    link1 := []string{linkHttp, server, linkP1, myUsername, linkP2, apiKey}// user Account

  for i := 0; i < 20; i++{//sends new set of requests every 5 min, 10 times(for now, can be any number)

      resp, err := http.Get(strings.Join(link1, ""))
      if err != nil{
        //send error to loggly
        logToLoggly.Error(err.Error())
        fmt.Println(err.Error())
      }

      defer resp.Body.Close()

      body, err := ioutil.ReadAll(resp.Body)

      if err != nil{
        //send error to loggly
        logToLoggly.Error(err.Error())
        fmt.Println(err.Error())
      } else {
        //send raw data to loggly
        _, err = logToLoggly.Write(body)
        if err != nil{
          fmt.Println(err.Error())
        }
      }

      //turn body into struct
      var user Account

      err = json.Unmarshal(body, &user)

      if err != nil{
        //send error to loggly
        logToLoggly.Error(err.Error())
        fmt.Println(err.Error())
      }

      link2 := []string{linkHttp, server, linkP3, user.ID , linkP2, apiKey}// ChampMastery

      resp, err = http.Get(strings.Join(link2, ""))
      if err != nil{
        //send error to loggly
        logToLoggly.Error(err.Error())
        fmt.Println(err.Error())
      }

      defer resp.Body.Close()

      body2, err := ioutil.ReadAll(resp.Body)

      if err != nil{
        //send error to loggly
        logToLoggly.Error(err.Error())
        fmt.Println(err.Error())
      } else {
        //send raw data to loggly
        _, err = logToLoggly.Write(body2)
        if err != nil{
          fmt.Println(err.Error())
        }
      }

      err = json.Unmarshal(body2, &user.ChampionData)


      printAccount(user)

      fmt.Println("Sleeping for 2 minutes")
      time.Sleep(2*time.Minute)//limited to 20 requests every 1 seconds and 100 requests every 2 minutes per server
      fmt.Println("Waking up. Program starting API requests again.")
  }
}

func printAccount(user Account){
  fmt.Println("ID: ", user.ID)
  fmt.Println("Account ID: ", user.AccountID)
  fmt.Println("Puuid: ", user.Puuid)
  fmt.Println("Username: ", user.Name)
  fmt.Println("Summoner Level: ", user.SummonerLevel)
  printMasteries(user.ChampionData)
}

func printMasteries(masteries ChampMastery){
  for i := 0; i< len(masteries); i++{//print out all of the mastery information for an account
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

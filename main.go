package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"./misc"
)

var stages []string

var config misc.Configuration

const startMonth int = 8 // august
const endMonth int = 6   // june

func intFormat(x int) string {
	s := strconv.Itoa(x)
	if x < 10 {
		s = "0" + s
	}
	return s
}

func main() {

	start := time.Now()
	config = misc.ReadConfig("config.json")
	stages = strings.Split(config.Stages, " ")
	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) > 0 {

		switch argsWithoutProg[0] {
		case "load":
			downloadAllSeasons("season", argsWithoutProg[1])
		case "get":
			parseDir()
		}

	}

	fmt.Println("Execution time:", time.Now().Sub(start))

}

func check(err error) {
	if err != nil {
		os.Exit(1)
	}
}

func requestToSite(url string) string {
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	req.Header.Add("Host", "ru.whoscored.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Accept", "text/plain, */*; q=0.01")
	req.Header.Add("Model-Last-Mode", config.Metka)
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Cookie", config.Cookie)
	req.Header.Add("Upgrade-Insecure-Requests", "1")

	req.Header.Add("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Referer", config.Refferer)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36 OPR/52.0.2871.40")
	check(err)

	resp, err := client.Do(req)

	check(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	b := bytes.NewBuffer(body)

	var r io.Reader
	r, err = gzip.NewReader(b)
	check(err)

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	check(err)

	resData := resB.Bytes()

	bufStr := "{matchs:" + string(resData) + "}"

	return bufStr

}

func getMatchesFromFile(path string) {
	re := regexp.MustCompile("[0-9]+")
	season := re.FindAllString(path, -1)

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	var res []string
	scanner := bufio.NewScanner(file)
	re = regexp.MustCompile("[0-9]+")

	for scanner.Scan() {
		if len(scanner.Text()) > 3 {
			res = re.FindAllString(scanner.Text(), -1)
			if len(res) != 0 && len(season) != 0 {
				loadMatch("matches/"+season[0], res[0])
			}

		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func loadMatch(path string, id string) {
	pagesType := map[string]string{"Betting": `(?ms:matchHeaderJson = JSON.parse\('(.*?)'\);)`, "Live": `(?ms:MatchDataForScrappers = (.*?);)`}
	var url, pageHTML, data string
	var r *regexp.Regexp
	var allFind, homePlayers, awayPlayers [][]string
	os.Mkdir(path+"/"+id, 0777)
	for key, value := range pagesType {
		url = "https://ru.whoscored.com/Matches/" + id + "/" + key + "/"
		pageHTML = requestToSite(url)
		if len(pageHTML) == 0 {
			panic(0)
		}
		r = regexp.MustCompile(value)
		allFind = r.FindAllStringSubmatch(pageHTML, -1)

		if key == "Live" { // broad search
			if len(allFind) == 0 {
				r = regexp.MustCompile(`(?ms:matchCentreData = (.*?);)`)
				allFind = r.FindAllStringSubmatch(pageHTML, -1)
			}

		}

		if key == "Betting" { // broad search

			r = regexp.MustCompile(`(?ms:homePlayers = JSON.parse\('\[(.*?)\]'\);)`)
			homePlayers = r.FindAllStringSubmatch(pageHTML, -1)
			if len(homePlayers) != 0 {
				err := ioutil.WriteFile(path+"/"+id+"/Homeplayers.json", []byte(homePlayers[0][1]), 0777)
				check(err)
			}

			r = regexp.MustCompile(`(?ms:awayPlayers = JSON.parse\('\[(.*?)\]'\);)`)
			awayPlayers = r.FindAllStringSubmatch(pageHTML, -1)
			if len(awayPlayers) != 0 {
				err := ioutil.WriteFile(path+"/"+id+"/Awayplayers.json", []byte(awayPlayers[0][1]), 0777)
				check(err)
			}

		}

		if len(allFind) != 0 && len(allFind[0]) != 0 {
			data = "{'data':" + allFind[0][1] + "}"
		} else {
			data = "{'data':[]}"
		}

		err := ioutil.WriteFile(path+"/"+id+"/"+key+".json", []byte(data), 0777)
		check(err)
		time.Sleep(1 * time.Second)
	}

}

func parseDir() {
	os.Mkdir("season", 0777)

	var url, bufStr, path string
	var err error

	for i := 0; i < len(stages); i++ {
		path = "season/" + strconv.Itoa(i+config.StartYear)
		os.Mkdir(path, 0777)

		for index := startMonth; index <= 12; index++ {
			url = "https://ru.whoscored.com/tournamentsfeed/" + stages[i] + "/Fixtures/?d=" + strconv.Itoa(i+config.StartYear) + intFormat(index) + "&isAggregate=false"

			fmt.Println(string(url))

			bufStr = requestToSite(url)

			err = ioutil.WriteFile(path+"/"+strconv.Itoa(i+config.StartYear+1)+strconv.Itoa(index)+".json", []byte(bufStr), 0777)

			check(err)

			time.Sleep(1 * time.Second)
		}

		for index := 1; index <= endMonth; index++ {
			url = "https://ru.whoscored.com/tournamentsfeed/" + stages[i] + "/Fixtures/?d=" + strconv.Itoa(i+config.StartYear+1) + intFormat(index) + "&isAggregate=false"

			fmt.Println(string(url))

			bufStr = requestToSite(url)

			err = ioutil.WriteFile(path+"/"+strconv.Itoa(i+config.StartYear+1)+strconv.Itoa(index)+".json", []byte(bufStr), 0777)

			check(err)

			time.Sleep(1 * time.Second)
		}

	}
}

func openDir(path string, isGetFiles bool) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		check(err)
	}

	res := make([]string, 0)
	for _, file := range files {
		if isGetFiles && !file.IsDir() || !isGetFiles && file.IsDir() {
			res = append(res, file.Name())
		}
	}
	return res
}

func downloadAllSeasons(path string, onlydir string) {

	os.Mkdir("matches", 0777)
	dirs := openDir(path, false)
	// set next season
	for _, namedir := range dirs {
		if namedir == onlydir {
			path = path.Combine(path, namedir)
			os.Mkdir("matches/"+namedir, 0777)

			files := openDir(path, true)

			tasks, _ := misc.SplitLinks(files, config.WorkersNumber)
			var manager misc.Manager
			for i := 0; i < config.WorkersNumber; i++ {

				var worker misc.Worker
				worker.Id = i
				worker.Status = false
				worker.Links = tasks[i]
				manager.Workers = append(manager.Workers, &worker)

				// parallel
				go func(worker *misc.Worker, i int, path string) {

					for _, file := range worker.Links {
						getMatchesFromFile(path + "/" + file)
					}
					time.Sleep(100 * time.Millisecond)
					worker.Status = true

				}(&worker, i, path)
			}

			for { // while true
				time.Sleep(100 * time.Millisecond)
				status := true
				for _, worker := range manager.Workers {
					if !worker.Status {
						status = false
						break
					}
				}
				if status {
					break
				}
			}

		}

	}

}

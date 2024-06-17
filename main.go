package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	value    string
	title    string
	location string
	company  string
}

var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?=&searchword=python"

func main() {
	var jobs []extractedJob
	c := make(chan []extractedJob)
	totalPages := getPages()
	// fmt.Println(totalPages)

	for i := 1; i < totalPages+1; i++ {
		go getPage(i, c)

	}

	for i := 1; i < totalPages+1; i++ {
		extractedJobs := <-c
		jobs = append(jobs, extractedJobs...)
	}

	// fmt.Println(jobs)
	writeJobs(jobs)
	fmt.Println("Done, extracted", len(jobs))
}

func getPage(page int, mainC chan<- []extractedJob) {

	var jobs []extractedJob
	c := make(chan extractedJob)

	pageURL := baseURL + "&&recruitPage=" + strconv.Itoa(page)
	fmt.Println("Requesting", pageURL)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".item_recruit")
	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
	})

	for i := 0; i < searchCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}

	mainC <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	value, _ := card.Attr("value")

	title := cleanString(card.Find(".area_job>.job_tit>a").Text())

	location := cleanString(card.Find(".area_job>.job_condition>span>a").Text())

	company := cleanString(card.Find(".area_corp>.corp_name>a").Text())
	// fmt.Println(value, title, location, company)
	c <- extractedJob{
		value:    value,
		title:    title,
		location: location,
		company:  company}

}

func getPages() int {
	pages := 1
	res, err := http.Get(baseURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Link", "Title", "Location", "Company"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{"https://www.saramin.co.kr/zf_user/jobs/relay/view?isMypage=no&rec_idx=" + job.value, job.title, job.location, job.company}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status", res.StatusCode)
	}
}

func cleanString(str string) string {
	// return strings.Fields(strings.TrimSpace(str))
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

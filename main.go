package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
)

var (
	companyID   string
	companyName string
	username    string
	password    string

	inputFileName  string
	outputFileName string

	format string

	delimeter      string
	delay          int
	userAgent      string
	start          int
	count          int
	emptyThreshold int
	cookieURL      *url.URL
	totalResults   int = 0
)

const (
	csvHeader = "firstname,lastname,occupation,link"
)

func init() {
	flag.StringVar(&companyID, "id", "", "Company ID (required)")
	flag.StringVar(&companyName, "name", "", "Company Name to lookup")
	flag.StringVar(&username, "u", "", "LinkedIn Username (required)")
	flag.StringVar(&password, "p", "", "LinkedIn Password (required)")
	flag.StringVar(&inputFileName, "i", "", "Input File")
	flag.StringVar(&outputFileName, "o", "", "Output File")
	flag.IntVar(&delay, "d", 1, "Number of seconds to delay between requests")
	flag.StringVar(&format, "f", "", "Format of output (csv, {first}.{last})")
	flag.IntVar(&start, "s", 0, "What part of list to start at")
	flag.IntVar(&count, "c", 20, "Results per request")
	flag.StringVar(&userAgent, "a", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:70.0) Gecko/20100101 Firefox/70.0", "User agent")
	flag.IntVar(&emptyThreshold, "e", 3, "Number of empty responses before quitting worker")
	flag.StringVar(&delimeter, "delim", ",", "Delimeter")
	flag.Parse()

	if inputFileName == "" {
		if companyID == "" && companyName == "" {
			fmt.Println("-id required!")
			flag.Usage()
			os.Exit(1)
		}
		if username == "" {
			fmt.Println("-u required")
			flag.Usage()
			os.Exit(1)
		}
		if password == "" {
			fmt.Println("-p required")
			flag.Usage()
			os.Exit(1)
		}
	}

	cookieURL, _ = url.Parse("https://www.linkedin.com/uas/login")
}

func main() {
	personList := []Person{}

	// Open output file
	var outFile *os.File
	var err error
	if outputFileName != "" {
		outFile, err = os.OpenFile(outputFileName,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		defer outFile.Close()
	}

	// If formatting input file
	if inputFileName != "" {
		fileBytes, err := ioutil.ReadFile(inputFileName)
		if err != nil {
			log.Fatal(err)
		}

		fileLoaded := false

		// Check if file is json
		err = json.Unmarshal(fileBytes, &personList)
		if err == nil {
			fileLoaded = true
		}

		// Check if file is csv
		if !fileLoaded {
			fileString := string(fileBytes)
			lines := strings.Split(fileString, "\n")

			lineFields := strings.Split(lines[0], delimeter)
			if len(lineFields) >= 2 {
				for _, line := range lines {
					if line == csvHeader {
						continue
					}

					fields := strings.Split(line, delimeter)
          if len(fields) < 2 {
            continue
          }
					p := Person{
						FirstName: fields[0],
						LastName:  fields[1],
					}

					if len(fields) >= 3 {
						p.Occupation = fields[2]
					}
					if len(fields) >= 4 {
						p.PublicIdentifier = fields[3]
					}

					// If formatting name
					if !strings.Contains(format, "raw") {
						p, err = FormatName(p)
						if err != nil {
							log.Printf("Unable to format %s %s. %s\n", p.FirstName, p.LastName, err.Error())
							continue
						}
					}

					personList = appendIfMissing(personList, p)
				}

				fileLoaded = true
			}
		}

		if !fileLoaded {
			fmt.Printf("%s is in an unknown format, quitting...\n", inputFileName)
			return
		}

		output, err := FormatOutput(personList)
		if err != nil {
			log.Fatal(err)
		}

		if outputFileName != "" {
			outFile.Write(output)
		} else {
			fmt.Println(string(output))
		}

		// file, err := os.Open(inputFileName)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// defer file.Close()

		// // Read file line by line
		// scanner := bufio.NewScanner(file)
		// for scanner.Scan() {
		// 	line := scanner.Text()

		// 	// Split line by delimeter
		// 	fields := strings.Split(line, delimeter)
		// 	if len(fields) >= 2 {
		// 		tmpLastName := strings.Join(fields[1:], " ")
		// 		p := Person{
		// 			FirstName: fields[0],
		// 			LastName:  tmpLastName,
		// 		}

		// 		//
		// 		if strings.Contains(format, "raw") {
		// 			personList = append(personList, p)
		// 		} else {
		// 			formatted, err := FormatName(p)
		// 			if err != nil {
		// 				log.Printf("Error formatting %s %s - %s\n", p.FirstName, p.LastName, err.Error())
		// 			} else {

		// 			}
		// 			fmt.Println(formatted)
		// 		}

		// 		// if outputFileName == "" {
		// 		// } else {
		// 		// 	if format == "" || format == "csv" || format == "raw_csv" {
		// 		// 		if _, err := outFile.WriteString(formatted + "\n"); err != nil {
		// 		// 			log.Println(err)
		// 		// 		}
		// 		// 	} else if format == "json" || format == "raw_json" {
		// 		// 		// jsonOut, err := json.Marshal()
		// 		// 	}
		// 		// }
		// 	}
		// }

		// if err := scanner.Err(); err != nil {
		// 	log.Fatal(err)
		// }

		return
	}

	// If scraping...
	usernames := strings.Split(username, ",")
	passwords := strings.Split(password, ",")

	// Check username vs password lists
	if len(usernames) != len(passwords) {
		flag.Usage()
		fmt.Println("\nPlease enter same number of usernames and passwords.")
		os.Exit(0)
	}

	// Login each client
	clients := []*http.Client{}
	for i, u := range usernames {
		jar, _ := cookiejar.New(nil)
		c := http.Client{Jar: jar}

		err := login(&c, u, passwords[i])
		if err != nil {
			log.Printf("%s Error logging in. - %s\n", u, err.Error())
			continue
		}

		log.Printf("%s Logged in successfully!\n", u)

		clients = append(clients, &c)
	}

	if len(clients) == 0 {
		log.Println("No valid users. Quitting...")
		return
	}

	// If no id supplied, get from company name
	if companyID == "" {
		id, err := getCompanyID(clients[0], companyName)
		if err != nil {
			log.Printf("Could not get company id: %s\n", err.Error())
			return
		}

		companyID = id
	}

	wg := sync.WaitGroup{}

	for i, c := range clients {
		wg.Add(1)
		go func(i int, c *http.Client) {
			startIndex := start
			emptyResponses := 0

			for emptyResponses < emptyThreshold {
				people, err := getPeople(c, startIndex)
				if err != nil {
					log.Printf("Error getting people, %s\n", err.Error())
				}

				for _, p := range people {
					log.Printf("Account %d found %s %s - %s\n", i, p.FirstName, p.LastName, p.Occupation)

					if !strings.Contains(format, "raw") {
						p, err = FormatName(p)
						if err != nil {
							log.Printf("Unable to format %s %s. %s\n", p.FirstName, p.LastName, err.Error())
							continue
						}
					}

					personList = appendIfMissing(personList, p)
				}

				startIndex += count * len(clients)
				if len(people) == 0 {
					emptyResponses++
				}
			}

			wg.Done()
		}(i, c)
	}

	wg.Wait()

	output, err := FormatOutput(personList)
	if err != nil {
		log.Fatal(err)
	}

	if outputFileName != "" {
		_, err := outFile.Write(output)
		if err != nil {
			log.Fatalf("Unable to write to output file %s. %s\n", outputFileName, err.Error())
		}
	} else {
		fmt.Println(string(output))
	}

	// people, err := getPeople(clients[0], 0)
	// if err != nil {
	// 	panic(err)
	// }

	// m := map[Person]bool{}

	// for _, p := range people {
	// 	fmt.Println(p.FirstName, p.LastName)
	// }

}

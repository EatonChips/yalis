package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"

	"github.com/fatih/color"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	companyID    string
	companyName  string
	companyIDs   []string
	companyNames []string
	companyFile  string
	username     string
	password     string

	inputFileName  string
	outputFileName string

	format string

	delimeter      string
	delay          int
	userAgent      string
	start          int
	count          int
	emptyThreshold int

	configFileName string

	cookieURL    *url.URL
	totalResults int = 0
)

const (
	csvHeader = "companyID,firstname,lastname,occupation,link"
)

func init() {
	flag.StringSliceVar(&companyIDs, "id", []string{}, "Company IDs (required)")
	flag.StringSliceVar(&companyNames, "name", []string{}, "Company names to lookup")
	flag.StringVarP(&configFileName, "config", "c", "", "Configuration File")
	flag.StringVarP(&username, "username", "u", "", "LinkedIn Username (required)")
	flag.StringVarP(&password, "password", "p", "", "LinkedIn Password (required)")
	flag.StringVarP(&inputFileName, "input-file", "i", "", "Input File")
	flag.StringVarP(&outputFileName, "output-file", "o", "", "Output File")
	flag.IntVarP(&delay, "delay", "d", 1, "Number of seconds to delay between requests")
	flag.StringVarP(&format, "format", "f", "", "Format of output (csv, {first}.{last})")
	flag.IntVarP(&start, "start", "s", 0, "What part of list to start at")
	flag.IntVarP(&count, "count", "n", 20, "Results per request")
	flag.StringVarP(&userAgent, "user-agent", "a", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:70.0) Gecko/20100101 Firefox/70.0", "User agent")
	flag.IntVarP(&emptyThreshold, "empty-threshold", "e", 3, "Number of empty responses before quitting worker")
	flag.StringVarP(&delimeter, "delimeter", "m", ",", "Delimeter")
	flag.Parse()

	fmt.Println(configFileName)

	viper.BindPFlags(flag.CommandLine)
	viper.SetConfigType("yaml")

	// If no config provided, attempt to load from ./config.yml
	if configFileName == "" {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.ReadInConfig()
	} else {
		log.Printf("Using config file: %s\n", configFileName)
		f, err := os.Open(configFileName)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		viper.ReadConfig(f)
	}

	// Use config variables
	companyIDs = viper.GetStringSlice("id")
	companyNames = viper.GetStringSlice("name")
	username = viper.GetString("username")
	password = viper.GetString("password")
	inputFileName = viper.GetString("input-file")
	outputFileName = viper.GetString("output-file")
	format = viper.GetString("format")
	delimeter = viper.GetString("delimeter")
	delay = viper.GetInt("delay")
	userAgent = viper.GetString("user-agent")
	start = viper.GetInt("start")
	count = viper.GetInt("count")
	emptyThreshold = viper.GetInt("empty-threshold")

	if inputFileName == "" {
		if len(companyIDs) == 0 && len(companyNames) == 0 && companyFile == "" {
			fmt.Println("No input option specified. Supports:\n")
			fmt.Println("\t--id Company id")
			fmt.Println("\t--name Company name")
			fmt.Println("\t--id-file File of company ids or names\n")
			flag.Usage()
			os.Exit(1)
		}
		if username == "" {
			fmt.Println("--username or -u required")
			flag.Usage()
			os.Exit(1)
		}
		if password == "" {
			fmt.Println("--password or -p required")
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
			printError(err.Error())
		}

		defer outFile.Close()
	}

	// If formatting input file
	if inputFileName != "" {
		fileBytes, err := ioutil.ReadFile(inputFileName)
		if err != nil {
			printError(err.Error())
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
					if len(fields) >= 5 {
						p.CompanyID = fields[4]
					}

					// If formatting name
					if !strings.Contains(format, "raw") {
						p, err = FormatName(p)
						if err != nil {
							printError(fmt.Sprintf("Unable to format %s %s. %s\n", p.FirstName, p.LastName, err.Error()))
							continue
						}
					}

					personList = appendIfMissing(personList, p)
				}

				fileLoaded = true
			}
		}

		if !fileLoaded {
			printError(fmt.Sprintf("%s is in an unknown format, quitting...\n", inputFileName))
			return
		}

		output, err := FormatOutput(personList)
		if err != nil {
			printError(fmt.Sprintf("%s - %s", "Unable to format", err.Error()))
		}

		if outputFileName != "" {
			outFile.Write(output)
		} else {
			fmt.Println(string(output))
		}

		return
	}

	// Login to LinkedIn
	jar, _ := cookiejar.New(nil)
	c := http.Client{Jar: jar}

	err = login(&c, username, password)
	if err != nil {
		printError(fmt.Sprintf("%s Error logging in. - %s\n", username, err.Error()))
		os.Exit(1)
	}

	printInfo(fmt.Sprintf("%s Logged in successfully!\n", username))

	// Get company IDs from names
	for _, name := range companyNames {
		id, err := getCompanyID(&c, name)
		if err != nil {
			printError(err.Error())
		}
		companyIDs = append(companyIDs, id)
	}

	// Loop through each company
	for _, id := range companyIDs {
		startIndex := 0
		emptyResponses := 0

		// Until no more results
		for emptyResponses < emptyThreshold {
			// Make API call
			people, err := getPeople(&c, id, startIndex)
			if err != nil {
				printError(fmt.Sprintf("Error getting people, %s", err.Error()))
			}

			// Loop over results
			for _, p := range people {
				printSuccess(fmt.Sprintf("Discovered %s %s - %s", p.FirstName, p.LastName, p.Occupation))

				// Don't format name if 'raw' format
				if !strings.Contains(format, "raw") {
					p, err = FormatName(p)
					if err != nil {
						printError(fmt.Sprintf("Unable to format %s %s. %s", p.FirstName, p.LastName, err.Error()))
						continue
					}
				}

				// Add new people to list
				personList = appendIfMissing(personList, p)
			}

			startIndex += count
			if len(people) == 0 {
				emptyResponses++
			}
		}
	}

	// Format scraped person list
	output, err := FormatOutput(personList)
	if err != nil {
		log.Fatal(err)
	}

	// Write output
	if outputFileName != "" {
		_, err := outFile.Write(output)
		if err != nil {
			printError(fmt.Sprintf("Unable to write to output file %s. %s\n", outputFileName, err.Error()))
		}
	} else {
		fmt.Println(string(output))
	}

}

func printError(s string) {
	fmt.Println(color.RedString("[!]"), s)
}

func printSuccess(s string) {
	fmt.Println(color.GreenString("[+]"), s)
}

func printInfo(s string) {
	fmt.Println(color.BlueString("[*]"), s)
}

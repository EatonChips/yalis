package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// FormatName returns the Person object with a sanitized/formatted
// firstName and lastName
func FormatName(p Person) (Person, error) {
	var newFirstName string
	var newLastName string

	// Format first name
	firstNameFields := strings.Fields(p.FirstName)
	if len(firstNameFields) >= 1 {
		newFirstName = firstNameFields[0]
	} else {
		return p, errors.New("No firstname found")
	}

	newFirstName = strings.ToLower(newFirstName)
	newFirstName = sanitizeString(newFirstName)

	// Format last name
	newLastName = p.LastName
	cutStrings := []string{",", "(", "|", "."}
	for _, s := range cutStrings {
		newLastName = strings.Split(newLastName, s)[0]
	}

	tmp := ""
	for _, w := range strings.Fields(newLastName) {
		if len(w) > len(tmp) {
			tmp = w
		}
	}
	newLastName = tmp

	newLastName = strings.ToLower(newLastName)
	newLastName = sanitizeString(newLastName)

	// Update Person's name with formatted name
	p.FirstName = newFirstName
	p.LastName = newLastName

	return p, nil
}

func formatPerson(p Person) (string, error) {
	var formattedName string
	var firstName string
	var lastName string

	// Only take first element
	firstNameFields := strings.Fields(p.FirstName)
	if len(firstNameFields) >= 1 {
		firstName = firstNameFields[0]
	} else {
		return "", errors.New("no firstname provided")
	}

	lastName = p.LastName
	lastName = strings.TrimSpace(lastName)
	cutStrings := []string{",", "(", "|", ".", " "}
	for _, s := range cutStrings {
		splitName := strings.Split(lastName, s)
		lastName = splitName[0]
		// fmt.Printf("%s | %s\n", firstName, splitName)
	}

	firstName = strings.ToLower(firstName)
	firstName = sanitizeString(firstName)
	lastName = strings.ToLower(lastName)
	lastName = sanitizeString(lastName)

	// lastSplit := strings.Fields(lastName)
	// if len(lastSplit) > 1 {
	// 	lastName = lastSplit[len(lastSplit)-1]
	// }

	switch format {
	case "":
		formattedName = fmt.Sprintf("%s %s", firstName, lastName)
	case "raw":
		formattedName = fmt.Sprintf("%s %s", p.FirstName, p.LastName)
	case "csv":
		formattedName = fmt.Sprintf("%s,%s", strings.Title(firstName), strings.Title(lastName))
	case "raw_csv":
		formattedName = fmt.Sprintf("%s,%s", p.FirstName, p.LastName)
	// case "json":
	// 	people[i] = Person{
	// 		FirstName: firstName,
	// 		LastName:  lastName,
	// 	}
	// case "raw_json":
	// 	break
	default:
		username := format
		username = strings.ReplaceAll(username, "{f}", string(firstName[0]))
		username = strings.ReplaceAll(username, "{first}", firstName)

		if len(lastName) != 0 {
			username = strings.ReplaceAll(username, "{l}", string(lastName[0]))
			username = strings.ReplaceAll(username, "{last}", lastName)
		} else {
			username = strings.ReplaceAll(username, "{l}", "")
			username = strings.ReplaceAll(username, "{last}", "")

		}

		formattedName = username
	}

	return formattedName, nil
}

func sanitizeString(s string) string {
	re := regexp.MustCompile("[^A-Za-z ]+")
	return re.ReplaceAllString(s, "")
}

func FormatOutput(personList []Person) ([]byte, error) {
	output := []byte{}

	if format == "csv" || format == "raw_csv" {
		output = append(output, []byte(csvHeader+"\n")...)

		for _, p := range personList {
			line := fmt.Sprintf("%s,%s,%s,%s,%s\n", p.FirstName, p.LastName, p.Occupation, p.PublicIdentifier, p.CompanyID)
			output = append(output, []byte(line)...)
		}
	} else if format == "json" || format == "raw_json" {
		jsonBytes, err := json.MarshalIndent(personList, "", "  ")
		if err != nil {
			return output, err
		}

		output = jsonBytes

		// _, err = outFile.Write(jsonBytes)
		// if err != nil {
		// 	log.Fatalf("Unable to write to output file %s. %s\n", outputFileName, err.Error())
		// }
	} else if format != "" {
		// format stuff
		for _, p := range personList {
			line := format
			// fmt.Printf("%s %s\n", p.FirstName, p.LastName)

			line = strings.ReplaceAll(line, "{f}", string(p.FirstName[0]))
			line = strings.ReplaceAll(line, "{first}", p.FirstName)

			if p.LastName != "" {
				line = strings.ReplaceAll(line, "{l}", string(p.LastName[0]))
				line = strings.ReplaceAll(line, "{last}", p.LastName)
			}

			output = append(output, []byte(line+"\n")...)

			// _, err := outFile.WriteString(line)
			// if err != nil {
			// 	log.Fatalf("Unable to write to output file %s. %s\n", outputFileName, err.Error())
			// }
		}
	} else {
		for _, p := range personList {
			line := fmt.Sprintf("%s %s\n", p.FirstName, p.LastName)

			output = append(output, []byte(line)...)
		}
	}

	return output, nil
}

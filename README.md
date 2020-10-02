# yalis

Yet Another LinkedIn Scraper

## Usage

```
yalis -id <CompanyID> -f <format> -o <outfile>
```

```
Usage of ./yalis:
  -a string
    	User agent (default "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:70.0) Gecko/20100101 Firefox/70.0")
  -c int
    	Results per request (default 20)
  -d int
    	Number of seconds to delay between requests (default 1)
  -delim string
    	Delimeter (default ",")
  -e int
    	Number of empty responses before quitting worker (default 3)
  -f string
    	Format of output (csv, {first}.{last})
  -i string
    	Input File
  -id string
    	Company ID (required)
  -name string
    	Company Name to lookup
  -o string
    	Output File
  -p string
    	LinkedIn Password (required)
  -s int
    	What part of list to start at
  -u string
    	LinkedIn Username (required)
```

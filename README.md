# yalis

Yet Another LinkedIn Scraper

## Typical Workflow

#### Scraping Employee Names

```
$ yalis -u legitimate@email.com -p Password123 --id 1234 -f csv -o yalis-1234.csv

$ yalis -u legitimate@email.com -p Password123 --name "Company Name Inc." -f csv -o yalis-company-name.csv
```

#### Formatting Employee Names

```
yalis -u legitimate@email.com -p Password123 -f {f}{last} -o yalis-users.txt
```

## Usage

```
Usage of ./yalis:
  -c, --config string         Configuration File
  -n, --count int             Results per request (default 20)
  -d, --delay int             Number of seconds to delay between requests (default 1)
  -m, --delimeter string      Delimeter (default ",")
  -e, --empty-threshold int   Number of empty responses before quitting worker (default 3)
  -f, --format string         Format of output (csv, {first}.{last})
      --id strings            Company IDs (required)
  -i, --input-file string     Input File
      --name strings          Company names to lookup
  -o, --output-file string    Output File
  -p, --password string       LinkedIn Password (required)
  -s, --start int             What part of list to start at
  -a, --user-agent string     User agent (default "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:70.0) Gecko/20100101 Firefox/70.0")
  -u, --username string       LinkedIn Username (required)
```

### Config

Use the example config.yml file in the repo to avoid specifying credentials via command line flags. 

```yml
username: ""
password: ""
format: "csv"
delimeter: ","
delay: 1
user-agent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:70.0) Gecko/20100101 Firefox/70.0"
start: 0
count: 20
empty-threshold: 3
# input-file: ""
# output-file: "linkedin.csv"
# id:
#   - 123
# name:
#   - "Company LLC."
```

### Format Strings

`-f` flag will replace replace the following strings

| Format String | Replacement   |
| ------------- | ------------- |
| `{f}`         | First Initial |
| `{first}`     | First Name    |
| `{l}`         | Last Initial  |
| `{last}`      | Last Name     |

#### Example: Rick Astley

| Format String          | Output             | 
| ---------------------- | ------------------ |
| `{first}.{last}`       | rick.astley        |
| `{f}{last}@domain.tld` | rastley@domain.tld |
| `{l}{first}`           | arick              |



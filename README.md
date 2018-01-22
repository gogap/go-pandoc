go-pandoc
============

# Run as a service

## Run at local

```bash
> go get github.com/gogap/go-pandoc
> cd $GOPATH/src/github.com/gogap/go-pandoc
> go build
> ./go-pandoc run -c app.conf
```

## Run at docker

 ```bash
 docker pull idocking/go-pandoc:latest
 docker run -it -d -p 8080:8080 idocking/go-pandoc:latest ./go-pandoc run
 ```

or

 ```bash
 docker-compose up -d
 ```

> then you could access the 8080 port
> in osx, you could get the docker ip by command `docker-machine ip`, 
> and the access service by IP:8080

## Config

`app.conf`

```
{

	service {
		path = "/v1"
		
		cors {
			allowed-origins = ["*"]
		}

		gzip-enabled = true

		graceful {
			timeout = 10s
		}

		http {
			address = ":8080"
			enabled = true
		}

		https {
			address = ":443"
			enabled = false
			cert    = ""
			key     = ""
		}

		templates  {
			render-html {
				template = "templates/render_html.tmpl"
			}

			binary {
				template = "templates/binary.tmpl"
			}
		}
	}

	pandoc {

		verbose     = false
		trace       = false
		dump-args   = false
		ignore-args = false

		safe-dir = "/app"

		fetchers {
			http {
				driver = http
				options {}
			}

			data {
				driver = data
				options {}
			}
		}
	}
}
```


## API

```json
{
    "fetcher": {
        "name": "data",
        "params": {
            "data": "base64String"
        }
    },
    "converter": {
        "from": "markdown",
        "to": "pdf",
        "standalone": true,
        "variable": {
            "CJKmainfont": "Source Han Sans SC",
            "mainfont": "Source Han Sans SC",
            "sansfont": "Source Han Sans SC",
            "geometry:margin": "1cm",
            "subject": "gsjbxx"
        },
        "template": "/app/data/docs.template"
    },
    "template": "binary"
}
```


### Request Args

Field|Values|Usage
:--|:--|:--
fetcher ||if is nil, converter.uri could not be empty, it will pass to pandoc
fetcher.name||fetcher name in `app.conf`
fetcher.params ||different fetcher driver has different options
converter||the options for converter


### converter

the converter is the following json struct


```json
{
  "from":"markdown",
  "to": "pdf",
  "pdf_engine": "xelatex"
   ...
}
```

> use `pandoc --help` command to list options


### Use curl

```bash
curl -X POST \
  http://IP:8080/v1/convert \
  -H 'accept-encoding: gzip' \
  -H 'cache-control: no-cache' \
  -H 'content-type: application/json' \
  -d '{
	"fetcher": {
		"name": "data",
		"params": {
			"data": "IyMjIEhlbGxvCgo+IEdvLVBhbmRvYw=="
		}
	},
	"converter":{
		"from": "markdown",
	    "to" : "pdf",
	    "standalone": true,
	    "variable":{
	    	"CJKmainfont":"Source Han Sans SC",
	    	"mainfont":"Source Han Sans SC",
	    	"sansfont": "Source Han Sans SC",
	    	"geometry:margin":"1cm",
	    	"subject":"gsjbxx"
	    },
	    "template": "/app/data/docs.template"
	},
	"template": "binary"
}' --compressed -o test.pdf
```

> if you enabled gzip, you should add arg `--compressed` to curl

### Template

The defualt template is 

```
{"code":{{.Code}},"message":"{{.Message}}"{{if .Result}},"result":{{.Result|Jsonify}}{{end}}}
```

response example:


```json
{"code":0,"message":"","result":{"data":"bGl.............}}
```


we could add `template` to render as different response, we have another example template named `render-data`


```json
{
	"converter":{
		...
	},
	"template": "render-html"
}
```

the response is 

```html
<html>
	<body>
	     	<img src="data:application/pdf;base64,bGl............"/> 
 	</body>
</html>
```

So, the template will render at brower directly. you could add more your templates

#### Template funcs

Func|usage
:--|:--
base64Encode|encode value to base64 string
base64Decode|decode base64 string to string
jsonify|marshal object
md5|string md5 hash
toBytes|convert value to []byte
htmlEscape|for html safe
htmlUnescape|unescape html

#### Template Args

```go
type TemplateArgs struct {
	From string
	To   string
	ConvertResponse
	Response *RespHelper
}

type ConvertResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}
```

#### Internal templates

> at templates dir


Name|Usage
:--|:--
 |default template, retrun `code`,`message`, `result`
render-html|render data to html
binary|you cloud use curl to download directly

##### use render-html

```bash
curl -X POST \
  http://IP:8080/v1/convert \
  -H 'accept-encoding: gzip' \
  -H 'cache-control: no-cache' \
  -H 'content-type: application/json' \
  -d '{
	"converter":{
		...
	},
	"template": "render-html"
}' --compressed -o bing.html
```


##### use binary

```bash
curl -X POST \
  http://IP:8080/v1/convert \
  -H 'accept-encoding: gzip' \
  -H 'cache-control: no-cache' \
  -H 'content-type: application/json' \
  -d '{
	"converter":{
		...
	},
	"template": "binary"
}' --compressed -o test.pdf
```

### Fetcher

fetcher is an external source input, sometimes we could not fetch data by url, or the go-pandoc could not access the url because of some auth options

##### Data fetcher

the request contain data


```bash
curl -X POST \
  http://IP:8080/v1/convert \
  -H 'accept-encoding: gzip' \
  -H 'cache-control: no-cache' \
  -H 'content-type: application/json' \
  -d '{
	"fetcher": {
		"name": "data",
		"params": {
			"data": "IyMjIEhlbGxvCgo+IEdvLVBhbmRvYw=="
		}
	},
	"converter":{
		"from": "markdown",
	    "to" : "pdf",
	    "standalone": true,
	    "variable":{
	    	"CJKmainfont":"Source Han Sans SC",
	    	"mainfont":"Source Han Sans SC",
	    	"sansfont": "Source Han Sans SC",
	    	"geometry:margin":"1cm",
	    	"subject":"gsjbxx"
	    },
	    "template": "/app/data/docs.template"
	},
	"template": "binary"
}' --compressed -o test.pdf
```

```bash
> echo IyMjIEhlbGxvCgo+IEdvLVBhbmRvYw== | base64 -D


### Hello

> Go-Pandoc

```

params:


```json
{
    "data":"base64string"
}
```


#### HTTP fetcher

Fetch data by http driver

```bash
curl -X POST \
  http://IP:8080/v1/convert \
  -H 'cache-control: no-cache' \
  -H 'content-type: application/json' \
  -d '{
	    "fetcher": {
	        "name": "http",
	        "params": {
	            "url": "https://raw.githubusercontent.com/golang/go/master/README.md"
	        }
	    },
	    "converter": {
	        "from": "markdown",
	        "to": "pdf",
	        "standalone": true,
	        "template": "/app/data/docs.template",
	        "variable": {
	            "CJKmainfont": "Source Han Sans SC",
	            "mainfont": "Source Han Sans SC",
	            "sansfont": "Source Han Sans SC",
	            "geometry:margin": "1cm",
	            "subject": "gsjbxx"
	        }
	    },
	    "template": "render-html"
}' -o golang-readme.html
```

> if the source contain image urls, it will not display correct, the image resource should be base64 format like:

```markdown

### Title

- content

#### Examle Image: 
![](data:image/png;base64,iVBORw.......)

```



#### Code your own fetcher

step 1: Implement the following interface

```go
type Fetcher interface {
	Fetch(FetchParams) ([]byte, error)
}

func NewDataFetcher(conf config.Configuration) (dataFetcher fetcher.Fetcher, err error) {
	dataFetcher = &DataFetcher{}
	return
}

```

step 2: Reigister your driver

```go
func init() {
	err := fetcher.RegisterFetcher("data", NewDataFetcher)

	if err != nil {
		panic(err)
	}
}
```

step 3: import driver and rebuild

```go
import (
	_ "github.com/gogap/go-pandoc/pandoc/fetcher/data"
	_ "github.com/gogap/go-pandoc/pandoc/fetcher/http"
)
```

> make sure the register name is unique



# Use this package as libary

Just import `github.com/gogap/go-pandoc/pandoc`

```go
pdoc, err := pandoc.New(conf)
//...
//...
convData, err := pdoc.Convert(fetcherOpts, convertOpts)
```
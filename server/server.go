package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/TV4/graceful"
	"github.com/gogap/config"
	"github.com/gogap/go-pandoc/pandoc"
	"github.com/gorilla/mux"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/rs/cors"
	"github.com/spf13/cast"
	"github.com/urfave/negroni"
)

const (
	defaultTemplateText = `{"code":{{.Code}},"message":"{{.Message}}"{{if .Result}},"result":{{.Result|jsonify}}{{end}}}`
)

var (
	pdoc *pandoc.Pandoc

	renderTmpls = make(map[string]*template.Template)

	defaultTmpl *template.Template
)

type ConvertData struct {
	Data []byte `json:"data"`
}

type ConvertArgs struct {
	Fetcher   *pandoc.FetcherOptions `json:"fetcher"`
	Converter *pandoc.ConvertOptions `json:"converter"`
	Template  string                 `json:"template"`
}

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

type serverWrapper struct {
	tls      bool
	certFile string
	keyFile  string

	reqNumber int64
	addr      string
	n         *negroni.Negroni

	timeout time.Duration
}

func (p *serverWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&p.reqNumber, 1)
	defer atomic.AddInt64(&p.reqNumber, -1)

	p.n.ServeHTTP(w, r)
}

func (p *serverWrapper) ListenAndServe() (err error) {

	if p.tls {
		err = http.ListenAndServeTLS(p.addr, p.certFile, p.keyFile, p)
	} else {
		err = http.ListenAndServe(p.addr, p)
	}

	return
}

func (p *serverWrapper) Shutdown(ctx context.Context) error {
	num := atomic.LoadInt64(&p.reqNumber)

	schema := "HTTP"

	if p.tls {
		schema = "HTTPS"
	}

	beginTime := time.Now()

	for num > 0 {
		time.Sleep(time.Second)
		timeDiff := time.Now().Sub(beginTime)
		if timeDiff > p.timeout {
			break
		}
	}

	log.Printf("[%s] Shutdown finished, Address: %s\n", schema, p.addr)

	return nil
}

type PandocServer struct {
	conf    config.Configuration
	servers []*serverWrapper
}

func New(conf config.Configuration) (srv *PandocServer, err error) {

	serviceConf := conf.GetConfig("service")

	pandocConf := conf.GetConfig("pandoc")

	pdoc, err = pandoc.New(pandocConf)

	if err != nil {
		return
	}

	// init templates

	defaultTmpl, err = template.New("default").Funcs(funcMap).Parse(defaultTemplateText)

	if err != nil {
		return
	}

	err = loadTemplates(
		serviceConf.GetConfig("templates"),
	)

	if err != nil {
		return
	}

	// init http server
	c := cors.New(
		cors.Options{
			AllowedOrigins:     serviceConf.GetStringList("cors.allowed-origins"),
			AllowedMethods:     serviceConf.GetStringList("cors.allowed-methods"),
			AllowedHeaders:     serviceConf.GetStringList("cors.allowed-headers"),
			ExposedHeaders:     serviceConf.GetStringList("cors.exposed-headers"),
			AllowCredentials:   serviceConf.GetBoolean("cors.allow-credentials"),
			MaxAge:             int(serviceConf.GetInt64("cors.max-age")),
			OptionsPassthrough: serviceConf.GetBoolean("cors.options-passthrough"),
			Debug:              serviceConf.GetBoolean("cors.debug"),
		},
	)

	r := mux.NewRouter()

	pathPrefix := serviceConf.GetString("path", "/")

	r.PathPrefix(pathPrefix).Path("/convert").
		Methods("POST").
		HandlerFunc(handlePandocToX)

	r.PathPrefix(pathPrefix).Path("/ping").
		Methods("GET", "HEAD").HandlerFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
			rw.Write([]byte("pong"))
		},
	)

	n := negroni.Classic()

	n.Use(c) // use cors

	if serviceConf.GetBoolean("gzip-enabled", true) {
		n.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	n.UseHandler(r)

	gracefulTimeout := serviceConf.GetTimeDuration("graceful.timeout", time.Second*3)

	enableHTTP := serviceConf.GetBoolean("http.enabled", true)
	enableHTTPS := serviceConf.GetBoolean("https.enabled", false)

	var servers []*serverWrapper

	if enableHTTP {

		listenAddr := serviceConf.GetString("http.address", "127.0.0.1:8080")

		httpServer := &serverWrapper{
			n:       n,
			timeout: gracefulTimeout,
			addr:    listenAddr,
		}

		servers = append(servers, httpServer)
	}

	if enableHTTPS {

		listenAddr := serviceConf.GetString("http.address", "127.0.0.1:443")
		certFile := serviceConf.GetString("https.cert")
		keyFile := serviceConf.GetString("https.key")

		httpsServer := &serverWrapper{
			n:        n,
			timeout:  gracefulTimeout,
			addr:     listenAddr,
			tls:      true,
			certFile: certFile,
			keyFile:  keyFile,
		}

		servers = append(servers, httpsServer)
	}

	srv = &PandocServer{
		conf:    conf,
		servers: servers,
	}

	return
}

func (p *PandocServer) Run() (err error) {

	wg := sync.WaitGroup{}

	wg.Add(len(p.servers))

	for i := 0; i < len(p.servers); i++ {
		go func(srv *serverWrapper) {
			defer wg.Done()
			shcema := "HTTP"
			if srv.tls {
				shcema = "HTTPS"
			}
			log.Printf("[%s] Listening on %s\n", shcema, srv.addr)
			graceful.ListenAndServe(srv)
		}(p.servers[i])
	}

	wg.Wait()

	return
}

func writeResp(rw http.ResponseWriter, convertArgs ConvertArgs, resp ConvertResponse) {

	var tmpl *template.Template
	if len(convertArgs.Template) == 0 {
		tmpl = defaultTmpl
	} else {
		var exist bool

		tmpl, exist = renderTmpls[convertArgs.Template]
		if !exist {
			tmpl = defaultTmpl
		}
	}

	respHelper := newRespHelper(rw)

	args := TemplateArgs{
		From:            convertArgs.Converter.From,
		To:              convertArgs.Converter.To,
		ConvertResponse: resp,
		Response:        respHelper,
	}

	buf := bytes.NewBuffer(nil)

	err := tmpl.Execute(buf, args)

	if err != nil {
		log.Println(err)
	}

	if !respHelper.Holding() {
		rw.Write(buf.Bytes())
	}
}

func handlePandocToX(rw http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)

	decoder.UseNumber()

	args := ConvertArgs{}

	err := decoder.Decode(&args)

	if err != nil {
		writeResp(rw, args, ConvertResponse{http.StatusBadRequest, err.Error(), nil})
		return
	}

	if args.Converter == nil {
		writeResp(rw, args, ConvertResponse{http.StatusBadRequest, "converter options is nil", nil})
		return
	}

	if args.Fetcher == nil {
		writeResp(rw, args, ConvertResponse{http.StatusBadRequest, "fetcher options is nil", nil})
		return
	}

	var convData []byte

	convData, err = pdoc.Convert(*args.Fetcher, *args.Converter)

	if err != nil {
		writeResp(rw, args, ConvertResponse{http.StatusBadRequest, err.Error(), nil})
		return
	}

	writeResp(rw, args, ConvertResponse{0, "", ConvertData{Data: convData}})

	return
}

func loadTemplates(tmplsConf config.Configuration) (err error) {
	if tmplsConf == nil {
		return
	}

	tmpls := tmplsConf.Keys()

	for _, name := range tmpls {

		file := tmplsConf.GetString(name + ".template")

		tmpl := template.New(name).Funcs(funcMap)

		var data []byte
		data, err = ioutil.ReadFile(file)

		if err != nil {
			return
		}

		tmpl, err = tmpl.Parse(string(data))

		if err != nil {
			return
		}

		renderTmpls[name] = tmpl
	}

	return
}

type RespHelper struct {
	rw   http.ResponseWriter
	hold bool
}

func newRespHelper(rw http.ResponseWriter) *RespHelper {
	return &RespHelper{
		rw:   rw,
		hold: false,
	}
}

func (p *RespHelper) SetHeader(key, value interface{}) error {
	k := cast.ToString(key)
	v := cast.ToString(value)

	p.rw.Header().Set(k, v)

	return nil
}

func (p *RespHelper) Hold(v interface{}) error {
	h := cast.ToBool(v)
	p.hold = h

	return nil
}

func (p *RespHelper) Holding() bool {
	return p.hold
}

func (p *RespHelper) Write(data []byte) error {
	p.rw.Write(data)
	return nil
}

func (p *RespHelper) WriteHeader(code interface{}) error {
	c, err := cast.ToIntE(code)
	if err != nil {
		return err
	}

	p.rw.WriteHeader(c)

	return nil
}

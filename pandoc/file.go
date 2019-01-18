package pandoc

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type File struct {
	Url     string
	SafeDir string

	TempDirPrefix string

	path          string
	lastError     error
	shouldCleanup bool

	initOnce sync.Once
}

func (p *File) Path() (filename string, err error) {
	p.initOnce.Do(
		func() {
			if len(p.Url) == 0 {
				return
			}

			var u *url.URL
			u, p.lastError = url.Parse(p.Url)
			switch u.Scheme {
			case "http", "https":
				p.path, p.lastError = p.downloadToFile()
				p.shouldCleanup = true
			case "data":
				p.path, p.lastError = p.base64dataToFile()
				p.shouldCleanup = true
			case "file", "":
				if !strings.HasPrefix(u.Path, p.SafeDir) {
					p.lastError = fmt.Errorf("file path is not in safe dir")
				} else {
					p.path = u.Path
				}
			default:
				p.lastError = fmt.Errorf("unknown path schema, %s", u.Scheme)
			}
		},
	)

	return p.path, p.lastError
}

func (p *File) downloadToFile() (fname string, err error) {

	cli := http.DefaultClient

	resp, err := cli.Get(p.Url)

	if err != nil {
		err = fmt.Errorf("download file failure for url %s, error: %s", p.Url, err)
		return
	}

	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")

	filename, err := p.urlToFileName(p.Url, ct)

	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("read body from %s, error: %s", p.Url, err)
		return
	}

	fname, err = p.writeToTempFile(filename, data)

	if err != nil {
		return
	}

	return
}

func (p *File) urlToFileName(fileUrl, contentType string) (ret string, err error) {

	u, err := url.Parse(fileUrl)
	if err != nil {
		err = fmt.Errorf("parse url of %s failure, error: %s", fileUrl, err)
		return
	}

	filename := filepath.Base(u.Path)

	if len(u.Path) == 0 {
		filename = uuid.New().String()
	}

	ext := filepath.Ext(filename)

	if len(ext) == 0 {
		if len(contentType) > 0 {
			exts, _ := mime.ExtensionsByType(contentType)
			if len(exts) > 0 {
				filename = filename + exts[0]
			}
		}
	}

	ret = filename

	return
}

func (p *File) writeToTempFile(filename string, data []byte) (fname string, err error) {
	tmpDir := os.TempDir()

	dir := filepath.Join(tmpDir, p.TempDirPrefix)

	err = os.MkdirAll(dir, 0755)

	if err != nil {
		err = fmt.Errorf("make temp dir failure: %s, error: %s", dir, err)
		return
	}

	tempFileName := filepath.Join(dir, filename)

	err = ioutil.WriteFile(tempFileName, data, 0644)

	if err != nil {
		err = fmt.Errorf("write file %s failure", tempFileName)
		return
	}

	fname = tempFileName

	return
}

func (p *File) parseBase64Data() (contentType string, data []byte, err error) {
	regExp := `data:(.*?);(.*?),(.*)`

	reg, err := regexp.Compile(regExp)

	if err != nil {
		return
	}

	matched := reg.FindAllStringSubmatch(p.Url, -1)

	if len(matched) != 1 && len(matched[0]) != 4 {
		err = fmt.Errorf("base64 data format error, the format should be: %s", "data:content-type;encoding,base64string")
		return
	}

	d, err := base64.StdEncoding.DecodeString(matched[0][3])
	if err != nil {
		err = fmt.Errorf("parse base64 data failure: %s", err)
		return
	}

	contentType = matched[0][1]
	data = d

	return
}

func (p *File) base64dataToFile() (fname string, err error) {

	ct, data, err := p.parseBase64Data()

	if err != nil {
		return
	}

	filename, err := p.urlToFileName(uuid.New().String(), ct)

	if err != nil {
		return
	}

	fname, err = p.writeToTempFile(filename, data)

	if err != nil {
		return
	}

	return
}

func (p *File) Cleanup() {
	if p.shouldCleanup && p.lastError == nil && len(p.path) > 0 {
		os.Remove(p.path)
	}
	return
}

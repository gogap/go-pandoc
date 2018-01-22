package pandoc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gogap/config"
	"github.com/pborman/uuid"

	"github.com/gogap/go-pandoc/pandoc/fetcher"
)

type Metadata map[string]string
type Variable map[string]string
type RequestHeader map[string]string

type ConvertOptions struct {
	From                  string        `json:"from"`
	To                    string        `json:"to"`
	DataDir               string        `json:"data_dir"`
	Smart                 bool          `json:"smart"`
	BaseHeaderLevel       int           `json:"base_header_level"`
	StripEmptyParagraphs  bool          `json:"strip_empty_paragraphs"`
	IndentedCodeClasses   string        `json:"indented_code_classes"`
	Filter                string        `json:"filter"`
	LuaFilter             string        `json:"lua_filter"`
	PreserveTabs          bool          `json:"preserve_tabs"`
	TabStop               int           `json:"tab_stop"`
	TrackChanges          string        `json:"track_changes"` // accept|reject|all
	FileScope             bool          `json:"file_scope"`
	ExtractMedia          string        `json:"extract_media"`
	Standalone            bool          `json:"standalone"`
	Template              string        `json:"template"`
	Metadata              Metadata      `json:"metadata"`
	Variable              Variable      `json:"variable"`
	PrintDefaultTemplate  string        `json:"print_default_template"`
	PrintDefaultDataFile  string        `json:"print_default_data_file"`
	PrintHighlightStyle   string        `json:"print_highlight_style"`
	DPI                   int           `json:"dpi"`
	EOL                   string        `json:"eol"`  // crlf|lf|native
	Wrap                  string        `json:"wrap"` // auto|none|preserve
	Columns               int           `json:"columns"`
	StripComments         bool          `json:"strip_comments"`
	TOC                   bool          `json:"toc"`
	TOCDepth              int           `json:"toc_depth"`
	NoHighlight           bool          `json:"no_highlight"`
	HighlightStyle        string        `json:"highlight_style"`
	SyntaxDefinition      string        `json:"syntax_definition"`
	IncludeInHeader       string        `json:"include_in_header"`
	IncludeBeforeBody     string        `json:"include_before_body"`
	IncludeAfterBody      string        `json:"include_after_body"`
	ResourcePath          string        `json:"resource_path"`
	RequestHeader         RequestHeader `json:"request_header"`
	SelfContained         bool          `json:"self_contained"`
	HtmlQTags             bool          `json:"html_q_tags"`
	Ascii                 bool          `json:"ascii"`
	ReferenceLinks        bool          `json:"reference_links"`
	ReferenceLocation     string        `json:"reference_location"` // block|section|document
	AtxHeaders            bool          `json:"atx_headers"`
	TopLevelDivision      string        `json:"top_level_division"` // section|chapter|part
	NumberSections        bool          `json:"number_sections"`
	NumberOffset          int           `json:"number_offset"`
	Listings              bool          `json:"listings"`
	Incremental           bool          `json:"incremental"`
	SlideLevel            int           `json:"slide_level"`
	SectionDivs           bool          `json:"section_divs"`
	DefaultImageExtension string        `json:"default_image_extension"`
	EmailObfuscation      string        `json:"email_obfuscation"` // none|javascript|references
	IdPrefix              string        `json:"id_prefix"`
	TitlePrefix           string        `json:"title_prefix"`
	CSS                   string        `json:"css"`
	ReferenceDoc          string        `json:"reference_doc"`
	EpubSubdirectory      string        `json:"epub_subdirectory"`
	EpubCoverImage        string        `json:"epub_cover_image"`
	EpubMetadata          string        `json:"epub_metadata"`
	EpubEmbedFont         string        `json:"epub_embed_font"`
	EpubChapterLevel      int           `json:"epub_chapter_level"`
	PDFEngine             string        `json:"pdf_engine"`
	PDFEngineOpt          string        `json:"pdf_engine_opt"`
	Bibliography          string        `json:"bibliography"`
	CSL                   string        `json:"csl"`
	CitationAbbreviations string        `json:"citation_abbreviations"`
	Natbib                bool          `json:"natbib"`
	Biblatex              bool          `json:"biblatex"`
	Mathml                bool          `json:"mathml"`
	Webtex                string        `json:"webtex"`
	Mathjax               string        `json:"mathjax"`
	Katex                 string        `json:"katex"`
	Latexmathml           string        `json:"latexmathml"`
	Mimetex               string        `json:"mimetex"`
	Jsmath                string        `json:"jsmath"`
	Gladtex               bool          `json:"gladtex"`
	Abbreviations         string        `json:"abbreviations"`
	FailIfWarnings        bool          `json:"fail_if_warnings"`

	verbose    bool
	trace      bool
	dumpArgs   bool
	ignoreArgs bool
}

func (p *ConvertOptions) toCommandArgs() []string {
	var args []string

	if p.Smart {
		args = append(args, "+smart")
	} else {
		args = append(args, "-smart")
	}

	if p.StripEmptyParagraphs {
		args = append(args, "--strip-empty-paragraphs")
	}

	if p.PreserveTabs {
		args = append(args, "--preserve-tabs")
	}

	if p.FileScope {
		args = append(args, "--file-scope")
	}

	if p.Standalone {
		args = append(args, "--standalone")
	}

	if p.StripComments {
		args = append(args, "--strip-comments")
	}

	if p.TOC {
		args = append(args, "--toc")
	}

	if p.NoHighlight {
		args = append(args, "--no-highlight")
	}

	if p.SelfContained {
		args = append(args, "--self-contained")
	}

	if p.HtmlQTags {
		args = append(args, "--html-q-tags")
	}

	if p.Ascii {
		args = append(args, "--ascii")
	}

	if p.ReferenceLinks {
		args = append(args, "--reference-links")
	}

	if p.AtxHeaders {
		args = append(args, "--atx-headers")
	}

	if p.NumberSections {
		args = append(args, "--number-sections")
	}

	if p.Listings {
		args = append(args, "--listings")
	}

	if p.Incremental {
		args = append(args, "--incremental")
	}

	if p.SectionDivs {
		args = append(args, "--section-divs")
	}

	if p.Natbib {
		args = append(args, "--natbib")
	}

	if p.Biblatex {
		args = append(args, "--biblatex")
	}

	if p.Mathml {
		args = append(args, "--mathml")
	}

	if p.Gladtex {
		args = append(args, "--gladtex")
	}

	if p.FailIfWarnings {
		args = append(args, "--fail-if-warnings")
	}

	for k, v := range p.Variable {
		args = append(args, "--variable", strings.Join([]string{k, v}, "="))
	}

	for k, v := range p.Metadata {
		args = append(args, "--metadata", strings.Join([]string{k, v}, "="))
	}

	for k, v := range p.RequestHeader {
		args = append(args, "--request-header", strings.Join([]string{k, v}, "="))
	}

	if p.PDFEngine == "" {
		p.PDFEngine = "xelatex"
	}

	if len(p.From) != 0 {
		args = append(args, "--from", p.From)
	}

	if len(p.To) != 0 && strings.ToUpper(p.To) != "PDF" {
		args = append(args, "--to", p.To)
	}

	if len(p.DataDir) != 0 {
		args = append(args, "--data-dir", p.DataDir)
	}

	if p.BaseHeaderLevel != 0 {
		args = append(args, "--base-header-level", strconv.Itoa(p.BaseHeaderLevel))
	}

	if len(p.IndentedCodeClasses) != 0 {
		args = append(args, "--indented-code-classes", p.IndentedCodeClasses)
	}

	if len(p.Filter) != 0 {
		args = append(args, "--filter", p.Filter)
	}

	if len(p.LuaFilter) != 0 {
		args = append(args, "--lua-filter", p.LuaFilter)
	}

	if p.TabStop != 0 {
		args = append(args, "--tab-stop", strconv.Itoa(p.TabStop))
	}

	if len(p.TrackChanges) != 0 {
		args = append(args, "--track-changes", p.TrackChanges)
	}

	if len(p.ExtractMedia) != 0 {
		args = append(args, "--extract-media", p.ExtractMedia)
	}

	if len(p.Template) != 0 {
		args = append(args, "--template", p.Template)
	}

	if len(p.PrintDefaultTemplate) != 0 {
		args = append(args, "--print-default-template", p.PrintDefaultTemplate)
	}

	if len(p.PrintDefaultDataFile) != 0 {
		args = append(args, "--print-default-data-file", p.PrintDefaultDataFile)
	}

	if len(p.PrintHighlightStyle) != 0 {
		args = append(args, "--print-highlight-style", p.PrintHighlightStyle)
	}

	if p.DPI != 0 {
		args = append(args, "--dpi", strconv.Itoa(p.DPI))
	}

	if len(p.EOL) != 0 {
		args = append(args, "--eol", p.EOL)
	}

	if len(p.Wrap) != 0 {
		args = append(args, "--wrap", p.Wrap)
	}

	if p.Columns != 0 {
		args = append(args, "--columns", strconv.Itoa(p.Columns))
	}

	if p.TOCDepth != 0 {
		args = append(args, "--toc-depth", strconv.Itoa(p.TOCDepth))
	}

	if len(p.HighlightStyle) != 0 {
		args = append(args, "--highlight-style", p.HighlightStyle)
	}

	if len(p.SyntaxDefinition) != 0 {
		args = append(args, "--syntax-definition", p.SyntaxDefinition)
	}

	if len(p.IncludeInHeader) != 0 {
		args = append(args, "--include-in-header", p.IncludeInHeader)
	}

	if len(p.IncludeBeforeBody) != 0 {
		args = append(args, "--include-before-body", p.IncludeBeforeBody)
	}

	if len(p.IncludeAfterBody) != 0 {
		args = append(args, "--include-after-body", p.IncludeAfterBody)
	}

	if len(p.ResourcePath) != 0 {
		args = append(args, "--resource-path", p.ResourcePath)
	}

	if len(p.ReferenceLocation) != 0 {
		args = append(args, "--reference-location", p.ReferenceLocation)
	}

	if len(p.TopLevelDivision) != 0 {
		args = append(args, "--top-level-division", p.TopLevelDivision)
	}

	if p.NumberOffset != 0 {
		args = append(args, "--number-offset", strconv.Itoa(p.NumberOffset))
	}

	if p.SlideLevel != 0 {
		args = append(args, "--slide-level", strconv.Itoa(p.SlideLevel))
	}

	if len(p.DefaultImageExtension) != 0 {
		args = append(args, "--default-image-extension", p.DefaultImageExtension)
	}

	if len(p.EmailObfuscation) != 0 {
		args = append(args, "--email-obfuscation", p.EmailObfuscation)
	}

	if len(p.IdPrefix) != 0 {
		args = append(args, "--id-prefix", p.IdPrefix)
	}

	if len(p.TitlePrefix) != 0 {
		args = append(args, "--title-prefix", p.TitlePrefix)
	}

	if len(p.CSS) != 0 {
		args = append(args, "--css", p.CSS)
	}

	if len(p.ReferenceDoc) != 0 {
		args = append(args, "--reference-doc", p.ReferenceDoc)
	}

	if len(p.EpubSubdirectory) != 0 {
		args = append(args, "--epub-subdirectory", p.EpubSubdirectory)
	}

	if len(p.EpubCoverImage) != 0 {
		args = append(args, "--epub-cover-image", p.EpubCoverImage)
	}

	if len(p.EpubMetadata) != 0 {
		args = append(args, "--epub-metadata", p.EpubMetadata)
	}

	if len(p.EpubEmbedFont) != 0 {
		args = append(args, "--epub-embed-font", p.EpubEmbedFont)
	}

	if p.EpubChapterLevel != 0 {
		args = append(args, "--epub-chapter-level", strconv.Itoa(p.EpubChapterLevel))
	}

	if len(p.PDFEngine) != 0 {
		args = append(args, "--pdf-engine", p.PDFEngine)
	}

	if len(p.PDFEngineOpt) != 0 {
		args = append(args, "--pdf-engine-opt", p.PDFEngineOpt)
	}

	if len(p.Bibliography) != 0 {
		args = append(args, "--bibliography", p.Bibliography)
	}

	if len(p.CSL) != 0 {
		args = append(args, "--csl", p.CSL)
	}

	if len(p.CitationAbbreviations) != 0 {
		args = append(args, "--citation-abbreviations", p.CitationAbbreviations)
	}

	if len(p.Webtex) != 0 {
		args = append(args, "--webtex", p.Webtex)
	}

	if len(p.Mathjax) != 0 {
		args = append(args, "--mathjax", p.Mathjax)
	}

	if len(p.Katex) != 0 {
		args = append(args, "--katex", p.Katex)
	}

	if len(p.Latexmathml) != 0 {
		args = append(args, "--latexmathml", p.Latexmathml)
	}

	if len(p.Mimetex) != 0 {
		args = append(args, "--mimetex", p.Mimetex)
	}

	if len(p.Jsmath) != 0 {
		args = append(args, "--jsmath", p.Jsmath)
	}

	if len(p.Abbreviations) != 0 {
		args = append(args, "--abbreviations", p.Abbreviations)
	}

	if p.verbose {
		args = append(args, "--verbose")
	}

	if p.dumpArgs {
		args = append(args, "--dump-args")
	}

	if p.ignoreArgs {
		args = append(args, "--ignore-args")
	}

	if p.trace {
		args = append(args, "--trace")
	}

	return args
}

type FetcherOptions struct {
	Name   string          `json:"name"`   // http, oss, data
	Params json.RawMessage `json:"params"` // Optional
}

type Pandoc struct {
	timeout  time.Duration
	fetchers map[string]fetcher.Fetcher

	verbose    bool
	trace      bool
	dumpArgs   bool
	ignoreArgs bool

	safeDir string
}

func New(conf config.Configuration) (pandoc *Pandoc, err error) {

	pdoc := &Pandoc{
		fetchers: make(map[string]fetcher.Fetcher),
	}

	commandTimeout := conf.GetTimeDuration("timeout", time.Second*300)

	pdoc.timeout = commandTimeout

	fetchersConf := conf.GetConfig("fetchers")

	if fetchersConf == nil || len(fetchersConf.Keys()) == 0 {
		pandoc = pdoc
		return
	}

	fetcherList := fetchersConf.Keys()

	for _, fName := range fetcherList {

		if len(fName) == 0 || fName == "default" {
			err = fmt.Errorf("fetcher name could not be '' or 'default'")
			return
		}

		_, exist := pdoc.fetchers[fName]

		if exist {
			err = fmt.Errorf("fetcher of %s already exist", fName)
			return
		}

		fetcherConf := fetchersConf.GetConfig(fName)
		fDriver := fetcherConf.GetString("driver")

		if len(fDriver) == 0 {
			err = fmt.Errorf("the fetcher of %s's driver is empty", fName)
			return
		}

		fOptions := fetcherConf.GetConfig("options")

		var f fetcher.Fetcher
		f, err = fetcher.New(fDriver, fOptions)

		if err != nil {
			return
		}

		pdoc.fetchers[fName] = f
	}

	pdoc.verbose = conf.GetBoolean("verbose")
	pdoc.trace = conf.GetBoolean("trace")
	pdoc.dumpArgs = conf.GetBoolean("dump-args")
	pdoc.ignoreArgs = conf.GetBoolean("ignore-args")

	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	pdoc.safeDir = conf.GetString("safe-dir", cwd)

	pandoc = pdoc

	return
}

func (p *Pandoc) Convert(fetcherOpts FetcherOptions, convertOpts ConvertOptions) (ret []byte, err error) {

	var data []byte

	if len(convertOpts.DataDir) > 0 && !filepath.HasPrefix(convertOpts.DataDir, p.safeDir) {
		err = fmt.Errorf("DataDir: '%s' is not is safe dir: '%s'", convertOpts.DataDir, p.safeDir)
		return
	}

	if len(fetcherOpts.Name) == 0 {
		err = fmt.Errorf("non input method, please check your fetcher options or uri param")
		return
	}

	if len(fetcherOpts.Name) > 0 {
		data, err = p.fetch(fetcherOpts)
		if err != nil {
			return
		}
	}

	tmpDir, err := ioutil.TempDir("", "go-pandoc")
	if err != nil {
		return
	}

	tmpInput := filepath.Join(tmpDir, uuid.New()) + "." + convertOpts.From
	tmpOutpout := filepath.Join(tmpDir, uuid.New()) + "." + convertOpts.To

	err = ioutil.WriteFile(tmpInput, data, 0644)
	if err != nil {
		return
	}

	defer os.Remove(tmpInput)

	convertOpts.verbose = p.verbose
	convertOpts.trace = p.trace
	convertOpts.dumpArgs = p.dumpArgs
	convertOpts.ignoreArgs = p.ignoreArgs

	args := convertOpts.toCommandArgs()

	args = append(args, []string{"--quiet", tmpInput, "--output", tmpOutpout}...)

	_, err = execCommand(p.timeout, "pandoc", args...)

	if err != nil {
		return
	}

	defer os.Remove(tmpOutpout)

	var result []byte
	result, err = ioutil.ReadFile(tmpOutpout)

	ret = result

	return
}

func (p *Pandoc) fetch(fetcherOpts FetcherOptions) (data []byte, err error) {
	fetcher, exist := p.fetchers[fetcherOpts.Name]
	if !exist {
		err = fmt.Errorf("fetcher %s not exist", fetcherOpts.Name)
		return
	}

	data, err = fetcher.Fetch([]byte(fetcherOpts.Params))

	return
}

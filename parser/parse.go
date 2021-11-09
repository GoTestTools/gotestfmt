package parser

import (
	"fmt"
	"regexp"

	"github.com/haveyoudebuggedit/gotestfmt/tokenizer"
)

// Parse creates a new formatter that reads the go test output from the input reader, formats it and writes
// to the output according to the passed configuration.
// The result are three channels: the prefix text before any recognized action, the download channel will receive either
// zero or one result and then be closed. Once the downloads channel is closed the parsed package results will be
// streamed over the second result.
func Parse(
	evts <-chan tokenizer.Event,
) (<-chan string, <-chan *Downloads, <-chan *Package) {
	prefixChannel := make(chan string)
	downloadsChannel := make(chan *Downloads)
	packagesChannel := make(chan *Package)
	go parse(evts, prefixChannel, downloadsChannel, packagesChannel)
	return prefixChannel, downloadsChannel, packagesChannel
}

var noModuleProvidesRegexp = regexp.MustCompile(`no required module provides package (?P<package>[^\s]+);`)

func parse(
	evts <-chan tokenizer.Event,
	prefixChannel chan string,
	downloadsChannel chan *Downloads,
	packagesChannel chan *Package,
) {
	downloadTracker := &downloadsTracker{
		prefixChannel:       prefixChannel,
		downloadResultsList: nil,
		downloadsByPackage:  map[string]*Download{},
		downloadsFinished:   false,
		target:              downloadsChannel,
	}
	pkgTracker := &packageTracker{
		currentPackage:  nil,
		packages:        map[string]*Package{},
		testCases:       nil,
		lastTestCase:    nil,
		testCasesByName: map[string]*TestCase{},
		target:          packagesChannel,
	}

	defer func() {
		downloadTracker.Write()
		pkgTracker.Write()
		close(packagesChannel)
	}()

	var lastAction tokenizer.Action
	for {
		evt, ok := <-evts
		if !ok {
			return
		}
		if evt.Action != tokenizer.ActionDownload &&
			evt.Action != tokenizer.ActionDownloadFailed &&
			evt.Action != tokenizer.ActionStdout {

			downloadTracker.Write()
			if evt.Package != "" {
				pkgTracker.SetPackage(
					&Package{
						Name: evt.Package,
					},
				)
			}
		}
	prevaction:
		switch evt.Action {
		case tokenizer.ActionRun:
			fallthrough
		case tokenizer.ActionCont:
			pkgTracker.SetLastTest(evt.Test)
		case tokenizer.ActionFail:
			result := ResultFail
			finish(evt, pkgTracker, result)
		case tokenizer.ActionPass:
			result := ResultPass
			finish(evt, pkgTracker, result)
		case tokenizer.ActionSkip:
			result := ResultSkip
			finish(evt, pkgTracker, result)
		case tokenizer.ActionDownload:
			downloadTracker.Add(
				&Download{
					Package: evt.Package,
					Version: evt.Version,
				},
			)
		case tokenizer.ActionDownloadFailed:
			downloadTracker.Add(
				&Download{
					Package: evt.Package,
					Version: evt.Version,
					Failed:  true,
					Reason:  string(evt.Output),
				},
			)
		case tokenizer.ActionPackage:
			pkgTracker.SetPackage(
				&Package{
					Name: evt.Package,
				},
			)
		case tokenizer.ActionStdout:
			switch lastAction {
			case "":
				// Special case: error message right before any output indicates that there was an error downloading a
				// dependency. We will try to identify it here.
				pkg := ""
				if match := noModuleProvidesRegexp.FindSubmatch(evt.Output); len(match) != 0 {
					pkg = string(match[1])
				} else {
					prefixChannel <- string(evt.Output)
					break prevaction
				}
				evt.Action = tokenizer.ActionDownloadFailed
				downloadTracker.Add(
					&Download{
						Package: pkg,
					},
				)
				fallthrough
			case tokenizer.ActionDownloadFailed:
				fallthrough
			case tokenizer.ActionDownload:
				lastDownload := downloadTracker.GetLast()
				lastDownload.Failed = true
				if lastDownload.Reason == "" {
					lastDownload.Reason = string(evt.Output)
				} else {
					lastDownload.Reason = fmt.Sprintf(
						"%s\n%s",
						lastDownload.Reason,
						string(evt.Output),
					)
				}
			case tokenizer.ActionRun:
				fallthrough
			case tokenizer.ActionPass:
				fallthrough
			case tokenizer.ActionFail:
				fallthrough
			case tokenizer.ActionSkip:
				fallthrough
			case tokenizer.ActionCont:
				lastTestCase := pkgTracker.GetLastTestCase()
				if lastTestCase.Output == "" {
					lastTestCase.Output = string(evt.Output)
				} else {
					lastTestCase.Output = fmt.Sprintf(
						"%s\n%s",
						lastTestCase.Output,
						string(evt.Output),
					)
				}
			case tokenizer.ActionPackage:
				lastPackage := pkgTracker.GetPackage()
				if lastPackage.Output == "" {
					lastPackage.Output = string(evt.Output)
				} else {
					lastPackage.Output = fmt.Sprintf(
						"%s\n%s",
						lastPackage.Output,
						string(evt.Output),
					)
				}
			case tokenizer.ActionFailFinal:
				fallthrough
			case tokenizer.ActionPassFinal:
				fallthrough
			case tokenizer.ActionSkipFinal:
				pkgTracker.Write()
			default:
				if len(evt.Output) > 0 {
					panic(fmt.Errorf("unexpected output after %s event: %s", lastAction, evt.Output))
				}
			}
		}
		if evt.Action != tokenizer.ActionStdout {
			lastAction = evt.Action
		}
	}
}

func finish(evt tokenizer.Event, pkgTracker *packageTracker, result Result) {
	if evt.Test != "" {
		pkgTracker.SetLastTest(evt.Test)
		lastTestCase := pkgTracker.GetLastTestCase()
		lastTestCase.Result = result
		lastTestCase.Duration = evt.Elapsed
		lastTestCase.Coverage = evt.Coverage
	}
	if evt.Package != "" {
		pkgTracker.SetPackage(
			&Package{
				Name:     evt.Package,
				Result:   result,
				Duration: evt.Elapsed,
				Reason:   string(evt.Output),
				Coverage: evt.Coverage,
			},
		)
	}
}

type packageTracker struct {
	currentPackage  *Package
	packages        map[string]*Package
	testCases       []*TestCase
	lastTestCase    *TestCase
	testCasesByName map[string]*TestCase
	target          chan<- *Package
}

func (p *packageTracker) SetPackage(pkg *Package) {
	if p.currentPackage == nil {
		pkg.TestCases = p.testCases
		p.currentPackage = pkg
		p.packages[pkg.Name] = pkg
		p.testCases = nil
		return
	} else if p.currentPackage.Name != pkg.Name {
		p.currentPackage = p.packages[pkg.Name]
	}

	if pkg.Result != "" {
		p.currentPackage.Result = pkg.Result
	}
	if pkg.Coverage != nil {
		p.currentPackage.Coverage = pkg.Coverage
	}
	if len(pkg.TestCases) != 0 {
		p.currentPackage.TestCases = pkg.TestCases
	}
	if len(pkg.Output) != 0 {
		p.currentPackage.Output = pkg.Output
	}
	if len(pkg.Reason) != 0 {
		p.currentPackage.Reason = pkg.Reason
	}
	if pkg.Duration != 0 {
		p.currentPackage.Duration = pkg.Duration
	}
}

func (p *packageTracker) Add(testCase *TestCase) {
	p.lastTestCase = testCase
	p.testCasesByName[testCase.Name] = testCase
	if p.currentPackage != nil {
		p.currentPackage.TestCases = append(p.currentPackage.TestCases, testCase)
	} else {
		p.testCases = append(p.testCases, testCase)
	}
}

func (p *packageTracker) GetLastTestCase() *TestCase {
	return p.lastTestCase
}

func (p *packageTracker) Write() {
	if p.currentPackage == nil {
		if len(p.testCases) > 0 {
			p.target <- &Package{
				TestCases: p.testCases,
			}
		}
		p.testCases = nil
	}
	for _, pkg := range p.packages {
		p.target <- pkg
	}
	p.packages = map[string]*Package{}
	p.currentPackage = nil
	p.testCasesByName = map[string]*TestCase{}
}

func (p *packageTracker) GetTestCase(test string) *TestCase {
	return p.testCasesByName[test]
}

func (p *packageTracker) SetLastTest(test string) {
	if _, ok := p.testCasesByName[test]; !ok {
		p.Add(
			&TestCase{
				Name: test,
			},
		)
	}
	p.lastTestCase = p.testCasesByName[test]
}

func (p *packageTracker) GetPackage() *Package {
	return p.currentPackage
}

type downloadsTracker struct {
	downloadResultsList []*Download
	downloadsByPackage  map[string]*Download
	downloadsFinished   bool
	lastDownload        *Download
	target              chan *Downloads
	prefixChannel       chan string
}

func (d *downloadsTracker) Add(download *Download) {
	if d.downloadsFinished {
		panic(fmt.Errorf("tried to add download after downloads are already finished (%v)", download))
	}
	packageID := fmt.Sprintf("%s@%s", download.Package, download.Version)
	if _, ok := d.downloadsByPackage[packageID]; ok {
		panic(fmt.Errorf("download already exists in tracker (%v)", download))
	}
	d.downloadsByPackage[packageID] = download
	d.downloadResultsList = append(d.downloadResultsList, download)
	d.lastDownload = download
}

func (d *downloadsTracker) GetLast() *Download {
	return d.lastDownload
}

func (d *downloadsTracker) Write() {
	if d.downloadsFinished {
		return
	}
	close(d.prefixChannel)
	failed := false
	for _, dl := range d.downloadResultsList {
		if dl.Failed {
			failed = true
			break
		}
	}
	d.target <- &Downloads{
		Packages: d.downloadResultsList,
		Failed:   failed,
	}
	d.downloadsFinished = true
	d.downloadResultsList = nil
	d.downloadsByPackage = nil
	close(d.target)
}

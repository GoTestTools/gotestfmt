package parser

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/haveyoudebuggedit/gotestfmt/v2/tokenizer"
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
	downloadsFailureReason := make(chan string)
	packagesChannel := make(chan *Package)
	go parse(evts, prefixChannel, downloadsChannel, downloadsFailureReason, packagesChannel)
	return prefixChannel, downloadsChannel, packagesChannel
}

var downloadErrors = []*regexp.Regexp{
	regexp.MustCompile(`no required module provides package (?P<package>[^\s]+);`),
	regexp.MustCompile(`updates to go.mod needed; to update it:`),
}

func parse(
	evts <-chan tokenizer.Event,
	prefixChannel chan string,
	downloadsChannel chan *Downloads,
	downloadsFailureReason chan string,
	packagesChannel chan *Package,
) {
	outputStarted := false
	downloadTracker := &downloadsTracker{
		prefixChannel:          prefixChannel,
		downloadResultsList:    nil,
		downloadsByPackage:     map[string]*Download{},
		downloadsFinished:      false,
		downloadsFailureReason: downloadsFailureReason,
		target:                 downloadsChannel,
	}
	pkgTracker := &packageTracker{
		packagesByName: map[string]*Package{},
		target:         packagesChannel,
	}

	defer func() {
		downloadTracker.Write()
		pkgTracker.Write()
		close(packagesChannel)
	}()

	var prevErroredDownload string
	var prevErroredPkg string
	for {
		evt, ok := <-evts
		if !ok {
			return
		}

		if evt.Action != tokenizer.ActionStdout {
			outputStarted = true
		}
		if evt.Action != tokenizer.ActionDownload && evt.Action != tokenizer.ActionDownloadFailed && evt.Package != "" {
			pkgTracker.SetTestStartTime(evt.Package, evt.Test, evt.Received)
			if evt.Elapsed != 0 {
				pkgTracker.SetTestElapsed(evt.Package, evt.Test, evt.Elapsed)
			}
			if evt.Coverage != nil {
				pkgTracker.SetCoverage(evt.Package, evt.Test, *evt.Coverage)
			}
			if evt.Cached {
				pkgTracker.SetCached(evt.Package, evt.Test)
			}
		}

		switch evt.Action {
		case tokenizer.ActionFailFinal:
			pkgTracker.SetResult(evt.Package, evt.Test, ResultFail)
			if len(evt.Output) > 0 {
				pkgTracker.AddReason(evt.Package, string(evt.Output))
			}
		case tokenizer.ActionSkipFinal:
			pkgTracker.SetResult(evt.Package, evt.Test, ResultSkip)
			if len(evt.Output) > 0 {
				pkgTracker.AddReason(evt.Package, string(evt.Output))
			}
		case tokenizer.ActionPassFinal:
			pkgTracker.SetResult(evt.Package, evt.Test, ResultPass)
			if len(evt.Output) > 0 {
				pkgTracker.AddReason(evt.Package, string(evt.Output))
			}
		case tokenizer.ActionFail:
			pkgTracker.SetResult(evt.Package, evt.Test, ResultFail)
			if len(evt.Output) > 0 {
				pkgTracker.AddReason(evt.Package, string(evt.Output))
			}
		case tokenizer.ActionPass:
			pkgTracker.SetResult(evt.Package, evt.Test, ResultPass)
			if len(evt.Output) > 0 {
				pkgTracker.AddReason(evt.Package, string(evt.Output))
			}
		case tokenizer.ActionSkip:
			pkgTracker.SetResult(evt.Package, evt.Test, ResultSkip)
			if len(evt.Output) > 0 {
				pkgTracker.AddReason(evt.Package, string(evt.Output))
			}
		case tokenizer.ActionDownload:
			downloadTracker.Add(
				evt.Package,
				evt.Version,
			)
		case tokenizer.ActionDownloadFailed:
			prevErroredDownload = evt.Package
			downloadTracker.SetDownloadFailed(evt.Package, evt.Version)
			downloadTracker.AddReason(evt.Package, evt.Output)
		case tokenizer.ActionPackage:
			pkgTracker.SetResult(evt.Package, "", ResultFail)
			prevErroredPkg = evt.Package
		case tokenizer.ActionStdout:
			if evt.JSON {
				// We have a JSON-encoded output, that makes things much easier.
				pkgTracker.AddOutput(evt.Package, evt.Test, evt.Output)
			} else if len(evt.Output) > 0 {
				// We don't have a JSON output, so this must be an error.
				foundDLError := false
				for _, dlError := range downloadErrors {
					if submatch := dlError.FindSubmatch(evt.Output); len(submatch) > 0 {
						if len(submatch) > 1 {
							pkgName := string(submatch[1])
							downloadTracker.SetDownloadFailed(pkgName, "")
							downloadTracker.AddReason(pkgName, evt.Output)
							prevErroredDownload = pkgName
						} else {
							downloadTracker.SetFailureReason(submatch[0])
							prevErroredDownload = "*"
						}
						foundDLError = true
						outputStarted = true
					}
				}
				if !foundDLError {
					if prevErroredDownload != "" {
						if prevErroredDownload == "*" {
							downloadTracker.SetFailureReason(evt.Output)
						} else {
							downloadTracker.AddReason(prevErroredDownload, evt.Output)
						}
					} else if prevErroredPkg != "" {
						pkgTracker.AddOutput(prevErroredPkg, "", evt.Output)
					} else if !outputStarted {
						prefixChannel <- string(evt.Output)
					} else if strings.HasPrefix(string(evt.Output), "exit status ") {
						// Ignore go-acc exit status reporting.
					} else {
						panic(
							fmt.Errorf(
								"unexpected output encountered: %s (Did you use -json on go test?)",
								evt.Output,
							),
						)
					}
				}
			}
		}
		if evt.Action != tokenizer.ActionStdout &&
			evt.Action != tokenizer.ActionDownloadFailed &&
			evt.Action != tokenizer.ActionPackage {
			prevErroredDownload = ""
			prevErroredPkg = ""
		}
	}
}

type packageTracker struct {
	packages       []*Package
	packagesByName map[string]*Package
	target         chan<- *Package
}

func (p *packageTracker) AddOutput(pkg string, test string, output []byte) {
	pkgObj := p.ensurePackage(pkg)
	if test == "" {
		pkgObj.Output = pkgObj.Output + string(output) + "\n"
		return
	}
	testCase := p.ensureTest(pkgObj, test)
	testCase.Output = testCase.Output + string(output) + "\n"
}

func (p *packageTracker) ensureTest(pkgObj *Package, test string) *TestCase {
	if _, ok := pkgObj.TestCasesByName[test]; !ok {
		tc := &TestCase{
			StartTime: nil,
			Name:      test,
			Result:    "",
			Duration:  0,
			Coverage:  nil,
			Output:    "",
		}
		pkgObj.TestCasesByName[test] = tc
		pkgObj.TestCases = append(pkgObj.TestCases, tc)
	}

	return pkgObj.TestCasesByName[test]
}

func (p *packageTracker) ensurePackage(pkg string) *Package {
	if pkg == "" {
		panic("BUG: Empty package name encountered.")
	}
	if _, ok := p.packagesByName[pkg]; !ok {
		pkgObj := &Package{
			StartTime:       nil,
			Name:            pkg,
			Result:          "",
			Duration:        0,
			Coverage:        nil,
			Output:          "",
			TestCases:       nil,
			TestCasesByName: map[string]*TestCase{},
			Reason:          "",
		}
		p.packagesByName[pkg] = pkgObj
		p.packages = append(p.packages, pkgObj)
	}
	return p.packagesByName[pkg]
}

func (p *packageTracker) SetTestStartTime(pkg string, test string, startTime time.Time) {
	if pkg == "" {
		return
	}
	pkgObj := p.ensurePackage(pkg)
	if test == "" {
		if pkgObj.StartTime == nil {
			pkgObj.StartTime = &startTime
		}
		return
	}
	testCase := p.ensureTest(pkgObj, test)
	if testCase.StartTime == nil {
		testCase.StartTime = &startTime
	}
}

func (p *packageTracker) SetTestElapsed(pkg string, test string, elapsed time.Duration) {
	if pkg == "" {
		return
	}
	pkgObj := p.ensurePackage(pkg)
	if test == "" {
		pkgObj.Duration = elapsed
		return
	}
	testCase := p.ensureTest(pkgObj, test)
	testCase.Duration = elapsed
}

func (p *packageTracker) SetResult(pkg string, test string, result Result) {
	if pkg == "" {
		return
	}
	pkgObj := p.ensurePackage(pkg)
	if test == "" {
		pkgObj.Result = result
		return
	}
	testCase := p.ensureTest(pkgObj, test)
	testCase.Result = result
}

func (p *packageTracker) SetCoverage(pkg string, test string, coverage float64) {
	if pkg == "" {
		return
	}
	pkgObj := p.ensurePackage(pkg)
	if test == "" {
		pkgObj.Coverage = &coverage
		return
	}
	testCase := p.ensureTest(pkgObj, test)
	testCase.Coverage = &coverage
}

func (p *packageTracker) AddReason(pkg string, reason string) {
	if pkg == "" {
		return
	}
	pkgObj := p.ensurePackage(pkg)
	pkgObj.Reason = pkgObj.Reason + reason + "\n"
}

func (p *packageTracker) SetCached(pkg string, test string) {
	if pkg == "" {
		return
	}
	pkgObj := p.ensurePackage(pkg)
	if test == "" {
		pkgObj.Cached = true
		return
	}
	testCase := p.ensureTest(pkgObj, test)
	testCase.Cached = true
}

func (p *packageTracker) Write() {
	sort.SliceStable(
		p.packages, func(i, j int) bool {
			return p.packages[i].Name < p.packages[j].Name
		},
	)
	for _, pkg := range p.packages {
		pkg.Output = strings.TrimRight(pkg.Output, "\n")
		pkg.Reason = strings.TrimRight(pkg.Reason, "\n")
		sort.SliceStable(
			pkg.TestCases, func(i, j int) bool {
				return compareTestCaseNames(pkg.TestCases[i].Name, pkg.TestCases[j].Name)
			},
		)
		for _, tc := range pkg.TestCases {
			tc.Output = strings.TrimRight(tc.Output, "\n")
			if tc.Result == "" {
				tc.Result = ResultFail
			}
		}
	}
	for _, pkg := range p.packages {
		p.target <- pkg
	}
}

func compareTestCaseNames(name1 string, name2 string) bool {
	parts1 := strings.SplitN(name1, "/", -1)
	parts2 := strings.SplitN(name2, "/", -1)

	for i := 0; i < int(math.Min(float64(len(parts1)), float64(len(parts2)))); i++ {
		part1 := parts1[i]
		part2 := parts2[i]
		if part1 < part2 {
			return true
		} else if part1 > part2 {
			return false
		}
	}
	return len(parts1) < len(parts2)
}

type downloadsTracker struct {
	downloadResultsList    []*Download
	downloadsByPackage     map[string]*Download
	downloadsFinished      bool
	lastDownload           *Download
	target                 chan *Downloads
	prefixChannel          chan string
	startTime              *time.Time
	endTime                *time.Time
	downloadsFailureReason chan string
	failureReason          []byte
}

func (d *downloadsTracker) Add(name string, version string) {
	if d.downloadsFinished {
		panic(fmt.Errorf("tried to add download after downloads are already finished (%v)", name))
	}

	pkg := d.ensurePackage(name)
	if version != "" {
		pkg.Version = version
	}
	d.lastDownload = pkg
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
		}
		dl.Reason = strings.TrimRight(dl.Reason, "\n")
	}
	if len(d.failureReason) > 0 {
		failed = true
	}
	d.target <- &Downloads{
		Packages:  d.downloadResultsList,
		Failed:    failed,
		StartTime: d.startTime,
		EndTime:   d.endTime,
		Reason:    strings.TrimSpace(string(d.failureReason)),
	}
	d.downloadsFinished = true
	d.downloadResultsList = nil
	d.downloadsByPackage = nil
	close(d.target)
}

func (d *downloadsTracker) SetDownloadFailed(name string, version string) {
	if d.downloadsFinished {
		panic(fmt.Errorf("tried to add download after downloads are already finished (%v)", name))
	}
	pkg := d.ensurePackage(name)
	if version != "" {
		pkg.Version = version
	}
	pkg.Failed = true
}

func (d *downloadsTracker) ensurePackage(name string) *Download {
	t := time.Now()
	if d.startTime == nil {
		d.startTime = &t
	}
	d.endTime = &t
	if _, ok := d.downloadsByPackage[name]; !ok {
		d.downloadsByPackage[name] = &Download{
			Package: name,
		}
		d.downloadResultsList = append(d.downloadResultsList, d.downloadsByPackage[name])
	}
	return d.downloadsByPackage[name]
}

func (d *downloadsTracker) AddReason(name string, output []byte) {
	if d.downloadsFinished {
		panic(fmt.Errorf("tried to add download after downloads are already finished (%v)", name))
	}
	pkg := d.ensurePackage(name)
	pkg.Reason = pkg.Reason + string(output) + "\n"
}

func (d *downloadsTracker) SetFailureReason(output []byte) {
	if d.downloadsFinished {
		panic(fmt.Errorf("tried to add download failure reason after downloads are already finished"))
	}
	d.failureReason = append(append(d.failureReason, output...), '\n')
}

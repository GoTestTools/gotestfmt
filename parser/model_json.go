package parser

import (
	"encoding/json"
	"fmt"
	"time"
)

type tmpTestCase struct {
	Name string `json:"name"`
	Result Result `json:"result"`
	Duration string `json:"duration"`
	Coverage *float64 `json:"coverage"`
	Output string `json:"output"`
}

func (t *TestCase) MarshalJSON() ([]byte, error) {
	tmp := tmpTestCase{
		Name: t.Name,
		Result: t.Result,
		Duration: t.Duration.String(),
		Coverage: t.Coverage,
		Output: t.Output,
	}
	return json.Marshal(tmp)
}

func (t *TestCase) UnmarshalJSON(data []byte) error {
	var tmp tmpTestCase
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	var duration time.Duration
	var err error
	if tmp.Duration != "" {
		duration, err = time.ParseDuration(tmp.Duration)
		if err != nil {
			return fmt.Errorf("failed to parse duration: %s (%w)", tmp.Duration, err)
		}
	}
	t.Name = tmp.Name
	t.Result = tmp.Result
	t.Duration = duration
	t.Coverage = tmp.Coverage
	t.Output = tmp.Output
	return nil
}

type tmpPackage struct {
	Name string `json:"name"`
	Result Result `json:"result"`
	Duration string `json:"duration"`
	Coverage *float64 `json:"coverage"`
	Output string `json:"output"`
	TestCases []*TestCase `json:"testcases"`
	Reason string `json:"reason"`
}

func (p *Package) MarshalJSON() ([]byte, error) {
	tmp := tmpPackage{
		Name:      p.Name,
		Result:    p.Result,
		Duration:  p.Duration.String(),
		Coverage:  p.Coverage,
		Output:    p.Output,
		TestCases: p.TestCases,
		Reason:    p.Reason,
	}
	return json.Marshal(tmp)
}

func (p *Package) UnmarshalJSON(data []byte) error {
	var tmp tmpPackage
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	var duration time.Duration
	var err error
	if tmp.Duration != "" {
		duration, err = time.ParseDuration(tmp.Duration)
		if err != nil {
			return fmt.Errorf("failed to parse duration: %s (%w)", tmp.Duration, err)
		}
	}
	p.Name = tmp.Name
	p.Result = tmp.Result
	p.Duration = duration
	p.Coverage = tmp.Coverage
	p.Output = tmp.Output
	p.TestCases = tmp.TestCases
	p.Reason = tmp.Reason
	return nil
}
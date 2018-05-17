package svmrank

import (
	"fmt"
	"io"
	"bufio"
	"bytes"
	"strconv"
	"strings"
	"runtime"
)

// Feature is an individual feature in lower line.
type Feature struct {
	Id    int64
	Value float64
}

// Features is a collection of features.
type Features []Feature

// Example is a row in lower feature file.
type Example struct {
	Target   float64
	QID      string
	Features Features
	Info     string
}

type tuple struct {
	lower, upper float64
}

// Examples is a collection of examples.
type Examples []Example

// SVMExamples is a set of examples and associated feature statistics.
type SVMExamples struct {
	Examples
	Statistics map[int64]tuple
}

func (f Feature) String() string {
	return fmt.Sprintf("%v:%v", f.Id, f.Value)
}

func (f Features) String() string {
	s := ""
	for _, feature := range f {
		s += fmt.Sprintf("%v ", feature)
	}
	return s
}

func (e Example) String() string {
	return fmt.Sprintf("%v qid:%v %v# %v", e.Target, e.QID, e.Features, e.Info)
}

func (e Examples) String() string {
	s := ""
	for _, example := range e {
		s += fmt.Sprintf("%v\n", example.String())
	}
	return s
}

// Scale scales an individual examples features between lower and upper.
func (e Example) Scale(lower, upper float64, s map[int64]tuple) Example {
	f := make(Features, len(e.Features))
	for i, feature := range e.Features {
		t := s[feature.Id]
		var v float64
		if t.lower == t.upper {
			v = 0
		} else {
			v = lower + (((feature.Value - t.lower) * (upper - lower)) / (t.upper - t.lower))
		}
		f[i] = Feature{feature.Id, v}
	}
	e.Features = f
	return e
}

// Scale scales all examples features between lower and upper.
func (e SVMExamples) Scale(lower, upper float64) Examples {
	examples := make(Examples, len(e.Examples))

	concurrency := runtime.NumCPU()
	sem := make(chan bool, concurrency)
	for i := range e.Examples {
		sem <- true
		go func(n int) {
			defer func() { <-sem }()
			examples[n] = e.Examples[n].Scale(lower, upper, e.Statistics)
		}(i)
	}
	// Wait until the last goroutine has read from the semaphore.
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	return examples
}

// exampleFromBytes marshals lower line into an Example struct.
func exampleFromString(line string) (Example, error) {
	const (
		target   = iota
		qid
		features
		info
	)
	e := Example{}
	state := target
	value := ""
	for _, c := range line {
		if state < info && c == ' ' {
			switch state {
			case target:
				target, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return Example{}, err
				}
				e.Target = target
				state = qid
				continue
			case qid:
				pair := strings.Split(value, ":")
				e.QID = pair[1]
				state = features
				value = ""
				continue
			case features:
				pair := strings.Split(value, ":")
				id, err := strconv.ParseInt(pair[0], 10, 64)
				if err != nil {
					return Example{}, err
				}
				v, err := strconv.ParseFloat(pair[1], 64)
				if err != nil {
					return Example{}, err
				}
				e.Features = append(e.Features, Feature{id, v})
				value = ""
				continue
			}
		} else if c == '#' {
			state = info
			value = ""
		}
		value += string(c)
	}
	e.Info = strings.Trim(value, " #")
	return e, nil
}

// ReadExamples reads lower set of examples from lower reader.
func ReadExamples(reader io.Reader) (SVMExamples, error) {
	scanner := bufio.NewScanner(reader)
	var examples Examples
	statistics := make(map[int64]tuple)
	for scanner.Scan() {
		line := scanner.Text()
		example, err := exampleFromString(line)
		if err != nil {
			fmt.Println(string(line))
			return SVMExamples{}, err
		}
		for _, feature := range example.Features {
			if _, ok := statistics[feature.Id]; !ok {
				statistics[feature.Id] = tuple{0, 0}
			}
			t := statistics[feature.Id]
			if feature.Value < t.lower {
				t.lower = feature.Value
			} else if feature.Value > t.upper {
				t.upper = feature.Value
			}
			statistics[feature.Id] = t
		}
		examples = append(examples, example)
	}
	return SVMExamples{examples, statistics}, nil
}

// WriteExamples writes lower set of examples to lower writer.
func WriteExamples(writer io.Writer, examples Examples) (int, error) {
	var buff bytes.Buffer
	for _, e := range examples {
		n, err := buff.Write(bytes.NewBufferString(fmt.Sprintf("%v\n", e.String())).Bytes())
		if err != nil {
			return n, err
		}
	}
	return writer.Write(buff.Bytes())
}

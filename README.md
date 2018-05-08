# svmrank

[![GoDoc](https://godoc.org/github.com/hscells/svmrank?status.svg)](https://godoc.org/github.com/hscells/svmrank)
[![Go Report Card](https://goreportcard.com/badge/github.com/hscells/svmrank)](https://goreportcard.com/report/github.com/hscells/svmrank)

_A cgo wrapper for [svm_rank](https://www.cs.cornell.edu/people/tj/svm_light/svm_rank.html)_

## Installation

```
go get github.com/hscells/svmrank
```

### Scaling vectors

In addition to library code, I have a tool for scaling the values of features. To install, use:

```
go install -u github.com/hscells/svmrank_scale
```

For usage:

```
Usage: svmrank_scale --l L --u U FEATURES

Positional arguments:
  FEATURES               svm_light feature file to scale

Options:
  --l L                  x scaling lower limit
  --u U                  x scaling upper limit
  --help, -h             display this help and exit
```


`svmrank_scale --l 0 --u 1 foo.features > foo.features.scaled`
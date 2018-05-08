package main

import (
	"github.com/alexflint/go-arg"
	"os"
	"github.com/hscells/svmrank"
)

type args struct {
	L        float64 `arg:"required,help:x scaling lower limit"`
	U        float64 `arg:"required,help:x scaling upper limit"`
	Features string  `arg:"positional,required,help:svm_light feature file to scale"`
}

func main() {
	// Parse the command line arguments.
	var args args
	arg.MustParse(&args)

	f, err := os.Open(args.Features)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	examples, err := svmrank.ReadExamples(f)
	if err != nil {
		panic(err)
	}

	if args.L >= args.U {
		panic("lower scaling factor too large")
	}

	_, err = svmrank.WriteExamples(os.Stdout, examples.Scale(args.L, args.U))
	if err != nil {
		panic(err)
	}

	return
}

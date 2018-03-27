package svmrank_test

import (
	"testing"
	"github.com/hscells/svmrank"
)

func TestLearn(t *testing.T) {
	svmrank.Verbosity(1)
	svmrank.Learn("test_r.features", "test_r.model")
}

func TestPredict(t *testing.T) {
	svmrank.Verbosity(4)
	svmrank.Predict("1.test", "1.model", "1.predictions")
}

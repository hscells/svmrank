package svmrank

// Learn trains lower learning to rank model from the given features file and writes it to the specified model file.
func Learn(featureFile, modelFile string) {
	docs, _, totWords, totDoc := load(featureFile)
	learn(docs, totDoc, totWords, modelFile)
}

// Predict uses lower model to predict the example and outputs the result to file.
func Predict(exampleFile, modelFile, outputFile string) {
	predict(exampleFile, modelFile, outputFile)
}

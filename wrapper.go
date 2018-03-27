// +build darwin linux
// TODO is there a specific windows build?

package svmrank

/*
#include "svm_light/svm_common.c"
#include "svm_light/svm_learn.c"
#include "svm_light/svm_hideo.c"

#cgo CFLAGS:-I.
#cgo LDFLAGS:-L. -lm

// used for prediction.
char docfile[200];
char modelfile[200];
char predictionsfile[200];

// set_verbosity sets the verbosity for svm_rank.
void set_verbosity(int v) {
	verbosity = v;
}

// load_docs loads a feature file into a DOC struct.
void load_docs(char *docfile, double *label, long *totwords, long *totdoc, DOC ***docs) {
	read_documents(docfile, docs, &label, totwords, totdoc);
	printf("final check: %ld\n", *totdoc);
}

// learn takes a loaded DOC struct and learns a model, then outputs the model to a file.
void learn(DOC **docs, double *rankvalue, long totdoc, long totwords, char* modelfile) {
  	// Parameters to svm_rank.
  	KERNEL_CACHE *kernel_cache;
  	LEARN_PARM learn_parm;
  	KERNEL_PARM kernel_parm;
  	MODEL *model = (MODEL *)my_malloc(sizeof(MODEL));

	// Initiate svm_rank with default values.
	set_learning_defaults(&learn_parm, &kernel_parm);
	learn_parm.type = RANKING;
	if(learn_parm.svm_iter_to_shrink == -9999) {
		if(kernel_parm.kernel_type == LINEAR)
		  	learn_parm.svm_iter_to_shrink=2;
    	else
      		learn_parm.svm_iter_to_shrink=100;
  	}

	// Set the kernel type of the model.
  	if(kernel_parm.kernel_type == LINEAR) {
		kernel_cache=NULL;
	}
	else {
		kernel_cache=kernel_cache_init(totdoc,learn_parm.kernel_cache_size);
	}

	// Learn the model.
	svm_learn_ranking(docs, rankvalue, totdoc, totwords, &learn_parm, &kernel_parm, &kernel_cache, model);

	// Write the model to the specified file.
	write_model(modelfile,model);

	// Free memory.
	free_model(model,0);
	for(int i=0;i<totdoc;i++)
		free_example(docs[i],1);
	free(docs);
}

// --- here on out is just a straight copy of svm_classify.c --- //

void read_input_parameters(int argc, char **argv, char *docfile, char *modelfile, char *predictionsfile, long int *verbosity, long int *pred_format) {
  	long i;
	strcpy (modelfile, "svm_model");
	strcpy (predictionsfile, "svm_predictions");
	(*verbosity)=2;
	(*pred_format)=1;

	for(i=1;(i<argc) && ((argv[i])[0] == '-');i++) {
		switch ((argv[i])[1]) {
			case 'h':  exit(0);
			case 'v': i++; (*verbosity)=atol(argv[i]); break;
			case 'f': i++; (*pred_format)=atol(argv[i]); break;
			default: printf("\nUnrecognized option %s!\n\n",argv[i]);
			exit(0);
		}
	}
	if((i+1)>=argc) {
		printf("\nNot enough input parameters!\n\n");
		exit(0);
	}
	strcpy (docfile, argv[i]);
	strcpy (modelfile, argv[i+1]);
	if((i+2)<argc) {
		strcpy (predictionsfile, argv[i+2]);
	}
	if(((*pred_format) != 0) && ((*pred_format) != 1)) {
		printf("\nOutput format can only take the values 0 or 1!\n\n");
		exit(0);
	}
}

int make_prediction(int argc, char* argv[]) {
	DOC *doc;
	WORD *words;
	long max_docs,max_words_doc,lld;
	long totdoc=0,queryid,slackid;
	long correct=0,incorrect=0,no_accuracy=0;
	long res_a=0,res_b=0,res_c=0,res_d=0,wnum,pred_format;
	long j;
	double t1,runtime=0;
	double dist,doc_label,costfactor;
	char *line,*comment;
	FILE *predfl,*docfl;
	MODEL *model;

	read_input_parameters(argc,argv,docfile,modelfile,predictionsfile, &verbosity,&pred_format);

	nol_ll(docfile,&max_docs,&max_words_doc,&lld);
	max_words_doc+=2;
	lld+=2;

	line = (char *)my_malloc(sizeof(char)*lld);
	words = (WORD *)my_malloc(sizeof(WORD)*(max_words_doc+10));

	model=read_model(modelfile);

	if(model->kernel_parm.kernel_type == 0) {
		add_weight_vector_to_linear_model(model);
	}

	if(verbosity>=2) {
		printf("Classifying test examples.."); fflush(stdout);
	}

	if ((docfl = fopen (docfile, "r")) == NULL) {
		perror (docfile); exit (1);
	}
	if ((predfl = fopen (predictionsfile, "w")) == NULL) {
		perror (predictionsfile); exit (1);
	}

	while ((!feof(docfl)) && fgets(line,(int)lld,docfl)) {
		if(line[0] == '#') continue;
		parse_document(line,words,&doc_label,&queryid,&slackid,&costfactor,&wnum,max_words_doc,&comment);
		totdoc++;
		if(model->kernel_parm.kernel_type == LINEAR) {
			for(j=0;(words[j]).wnum != 0;j++) {
				if((words[j]).wnum>model->totwords)
					(words[j]).wnum=0;
			}
		}

		doc = create_example(-1,0,0,0.0,create_svector(words,comment,1.0));
		t1=get_runtime();

		if(model->kernel_parm.kernel_type == LINEAR) {
			dist=classify_example_linear(model,doc);
		} else {
			dist=classify_example(model,doc);
		}

		runtime+=(get_runtime()-t1);
		free_example(doc,1);

		if(dist>0) {
			if(pred_format==0) {
				fprintf(predfl,"%.8g:+1 %.8g:-1\n",dist,-dist);
			}
			if(doc_label>0) correct++; else incorrect++;
			if(doc_label>0) res_a++; else res_b++;
		} else {
			if(pred_format==0) {
				fprintf(predfl,"%.8g:-1 %.8g:+1\n",-dist,dist);
			}
			if(doc_label<0) correct++; else incorrect++;
			if(doc_label>0) res_c++; else res_d++;
		}
		if(pred_format==1) {
			fprintf(predfl,"%.8g\n",dist);
		}
		if((int)(0.01+(doc_label*doc_label)) != 1) {
			no_accuracy=1;
		}
		if(verbosity>=2) {
			if(totdoc % 100 == 0) {
				printf("%ld..",totdoc); fflush(stdout);
			}
		}
	}
	fclose(predfl);
	fclose(docfl);
	free(line);
	free(words);
	free_model(model,1);

	if(verbosity>=2) {
		printf("done\n");
		printf("Runtime (without IO) in cpu-seconds: %.2f\n", (float)(runtime/100.0));
	}
	if((!no_accuracy) && (verbosity>=1)) {
		printf("Accuracy on test set: %.2f%% (%ld correct, %ld incorrect, %ld total)\n",(float)(correct)*100.0/totdoc,correct,incorrect,totdoc);
		printf("Precision/recall on test set: %.2f%%/%.2f%%\n",(float)(res_a)*100.0/(res_a+res_b),(float)(res_a)*100.0/(res_a+res_c));
	}

	return(0);
}
*/
import "C"
import "unsafe"

// Verbosity sets the verbosity level for svm_rank.
func Verbosity(v int) {
	C.set_verbosity(C.int(v))
}

// load loads feature vectors from a file.
func load(filename string) (**C.DOC, C.double, C.long, C.long) {
	var docs **C.DOC
	var label C.double
	var totWords C.long
	var totDoc C.long
	docFile := C.CString(filename)
	C.load_docs(docFile, &label, &totWords, &totDoc, &docs)
	return docs, label, totWords, totDoc
}

// learn takes feature vectors and learns a model, then outputs the model to file.
func learn(docs **C.DOC, totDoc, totWords C.long, filename string) {
	var rankValue C.double
	modelFile := C.CString(filename)
	C.learn(docs, &rankValue, totDoc, totWords, modelFile)
}

// predict takes an example file and a model file and produces some prediction in the output file.
func predict(exampleFile, modelFile, outputFile string) {
	// Create a C array.
	argv := C.malloc(C.size_t(4) * C.size_t(unsafe.Sizeof(uintptr(0))))

	// Convert the C array to a Go Array so we can index it.
	a := (*[1<<30 - 1]*C.char)(argv)

	// Populate arguments to function.
	a[0] = C.CString("")
	a[1] = C.CString(exampleFile)
	a[2] = C.CString(modelFile)
	a[3] = C.CString(outputFile)

	C.make_prediction(C.int(4), (**C.char)(argv))
}

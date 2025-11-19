package retrier

type DefaultRetryClassifier struct{}

func NewDefaultRetryClassifier() *DefaultRetryClassifier {
	return &DefaultRetryClassifier{}
}

// DefaultRetryClassifier все ошибки класифицирует как Retriable
func (*DefaultRetryClassifier) ClassifyError(err error) ErrorClass {
	if err == nil {
		return NonRetriable
	}
	return Retriable
}

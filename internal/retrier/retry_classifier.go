package retrier

type DefaultRetryClassifier struct{}

// NewDefaultRetryClassifier creates a new DefaultRetryClassifier instance.
//
// Returns a *DefaultRetryClassifier ready for use in classifying errors.
func NewDefaultRetryClassifier() *DefaultRetryClassifier {
	return &DefaultRetryClassifier{}
}

// ClassifyError classifies errors as Retriable or NonRetriable.
//
// Parameters:
//   - err: The error to classify
//
// Returns Retriable for all non-nil errors, and NonRetriable for nil.
func (*DefaultRetryClassifier) ClassifyError(err error) ErrorClass {
	if err == nil {
		return NonRetriable
	}
	return Retriable
}

package scorecard

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"

	ks "github.com/zegl/kube-score/domain"
)

const (
	ignoredChecksAnnotation = "kube-score/ignore"
)

type Scorecard map[string]*ScoredObject

// New creates and initializes a new Scorecard
func New() Scorecard {
	return make(Scorecard)
}

func (s Scorecard) NewObject(typeMeta metav1.TypeMeta, objectMeta metav1.ObjectMeta) *ScoredObject {
	o := &ScoredObject{
		TypeMeta:   typeMeta,
		ObjectMeta: objectMeta,
		Checks:     make([]TestScore, 0),
	}

	// If this object already exists, return the previous version
	if object, ok := s[o.resourceRefKey()]; ok {
		return object
	}

	o.setIgnoredTests()

	s[o.resourceRefKey()] = o
	return o
}

func (s Scorecard) AnyBelowOrEqualToGrade(threshold Grade) bool {
	for _, o := range s {
		for _, s := range o.Checks {
			if s.Grade <= threshold {
				return true
			}
		}
	}
	return false
}

type ScoredObject struct {
	TypeMeta   metav1.TypeMeta
	ObjectMeta metav1.ObjectMeta
	Checks     []TestScore

	ignoredChecks map[string]struct{}
}

func (so *ScoredObject) setIgnoredTests() {
	ignoredMap := make(map[string]struct{})
	if ignoredCSV, ok := so.ObjectMeta.Annotations[ignoredChecksAnnotation]; ok {
		for _, ignored := range strings.Split(ignoredCSV, ",") {
			ignoredMap[strings.TrimSpace(ignored)] = struct{}{}
		}
	}
	so.ignoredChecks = ignoredMap
}

func (so ScoredObject) resourceRefKey() string {
	return so.TypeMeta.Kind + "/" + so.TypeMeta.APIVersion + "/" + so.ObjectMeta.Namespace + "/" + so.ObjectMeta.Name
}

func (so ScoredObject) HumanFriendlyRef() string {
	s := so.ObjectMeta.Name
	if so.ObjectMeta.Namespace != "" {
		s += "/" + so.ObjectMeta.Namespace
	}
	s += " " + so.TypeMeta.APIVersion + "/" + so.TypeMeta.Kind
	return s
}

func (so *ScoredObject) Add(ts TestScore, check ks.Check) {
	// This test is ignored, don't save it
	if _, ok := so.ignoredChecks[check.ID]; ok {
		return
	}

	ts.Check = check
	so.Checks = append(so.Checks, ts)
}

type TestScore struct {
	Check    ks.Check
	Grade    Grade
	Comments []TestScoreComment
}

type Grade int

const (
	GradeCritical Grade = 1
	GradeWarning  Grade = 5
	GradeAlmostOK Grade = 7
	GradeAllOK    Grade = 10
)

type TestScoreComment struct {
	Path        string
	Summary     string
	Description string
}

func (ts *TestScore) AddComment(path, summary, description string) {
	ts.Comments = append(ts.Comments, TestScoreComment{
		Path:        path,
		Summary:     summary,
		Description: description,
	})
}

package main

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"io"
	"sort"
	"strings"

	"github.com/zegl/kube-score/scorecard"
)

func outputMillenial(scoreCard *scorecard.Scorecard, okThreshold, warningThreshold int) io.Reader {
	// Print the items sorted by item kind and scorecard key
	var keys []string
	for k := range *scoreCard {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// var lastKind string

	type KindStatus struct {
		Kind string

		CountOK       int
		CountWarning  int
		CountCritical int

		Objects []scorecard.ScoredObject
	}

	var statuses []KindStatus

	var currentStatus KindStatus

	for _, key := range keys {
		scoredObject := (*scoreCard)[key]

		if scoredObject.TypeMeta.Kind != currentStatus.Kind {
			// fmt.Fprintf(w, "%s\t\t\t 3 critical, 8 warning, 10 OK\n", scoredObject.TypeMeta.Kind)
			// fmt.Fprintf(w, "%s\n", strings.Repeat("-", 80))
			// lastKind = scoredObject.TypeMeta.Kind

			if currentStatus.Kind != "" {
				statuses = append(statuses, currentStatus)
			}

			currentStatus = KindStatus{
				Kind: scoredObject.TypeMeta.Kind,
			}
		}

		if scoredObject.Grade() == scorecard.GradeCritical {
			currentStatus.CountCritical++
		} else if scoredObject.Grade() == scorecard.GradeWarning {
			currentStatus.CountWarning++
		} else {
			currentStatus.CountOK++
		}

		currentStatus.Objects = append(currentStatus.Objects, *scoredObject)

		// Headers for each object
		/*color.New(color.FgMagenta).Fprintf(w, "%s/%s %s", scoredObject.TypeMeta.APIVersion, scoredObject.TypeMeta.Kind, scoredObject.ObjectMeta.Name)
		if scoredObject.ObjectMeta.Namespace != "" {
			color.New(color.FgMagenta).Fprintf(w, " in %s\n", scoredObject.ObjectMeta.Namespace)
		} else {
			fmt.Fprintln(w)
		}

		for _, card := range scoredObject.Checks {
			r := outputMillenialStep(card, okThreshold, warningThreshold)
			io.Copy(w, r)
		}*/

	}
	statuses = append(statuses, currentStatus)

	w := bytes.NewBufferString("")

	fmt.Fprintln(w, `
 ___  __    ___  ___  ________  _______                  ________  ________  ________  ________  _______      
|\  \|\  \ |\  \|\  \|\   __  \|\  ___ \                |\   ____\|\   ____\|\   __  \|\   __  \|\  ___ \     
\ \  \/  /|\ \  \\\  \ \  \|\ /\ \   __/|   ____________\ \  \___|\ \  \___|\ \  \|\  \ \  \|\  \ \   __/|    
 \ \   ___  \ \  \\\  \ \   __  \ \  \_|/__|\____________\ \_____  \ \  \    \ \  \\\  \ \   _  _\ \  \_|/__  
  \ \  \\ \  \ \  \\\  \ \  \|\  \ \  \_|\ \|____________|\|____|\  \ \  \____\ \  \\\  \ \  \\  \\ \  \_|\ \ 
   \ \__\\ \__\ \_______\ \_______\ \_______\               ____\_\  \ \_______\ \_______\ \__\\ _\\ \_______\
    \|__| \|__|\|_______|\|_______|\|_______|              |\_________\|_______|\|_______|\|__|\|__|\|_______|
                                                           \|_________|                                       
                                                                                                              
                                                                                                              `)

	pluralized := map[string]string{
		"Pod":                 "Dods",
		"Deployment":          "Deployments",
		"StatefulSet":         "StatefulSets",
		"Job":                 "Jobs",
		"CronJob":             "CronJobs",
		"NetworkPolicy":       "NetworkPolicies",
		"PodDisruptionBudget": "PodDisruptionBudgets",
	}

	// TODO: Don't print ignored checks

	for _, status := range statuses {
		fmt.Fprintln(w)

		kind := status.Kind
		if plural, ok := pluralized[kind]; ok {
			kind = plural
		}

		l, _ := fmt.Fprintf(w, "%s", kind)
		rowSuffix := fmt.Sprintf("%d critical, %d warning, %d OK", status.CountCritical, status.CountWarning, status.CountOK)
		fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", 80-l-len(rowSuffix)), rowSuffix)
		fmt.Fprintf(w, "%s\n", strings.Repeat("-", 80))

		for _, obj := range status.Objects {
			l, _ := fmt.Fprintf(w, "  %s", obj.ObjectMeta.Name)

			fmt.Fprintf(w, "%s%s\n", strings.Repeat(".", 78-l), obj.Grade().Emoji())

			if obj.AnyBelowOrEqualToGrade(scorecard.Grade(okThreshold)) {
				for _, check := range obj.Checks {
					if check.Grade <= scorecard.Grade(okThreshold) {
						if len(check.MillenialComment) > 0 {

							color.New(check.Grade.Color()).Fprintf(w, "    · %s\n", check.MillenialComment)
						} else {
							for _, comment := range check.Comments {
								color.New(check.Grade.Color()).Fprintf(w, "    · %s\n", comment.Summary)
							}
						}
					}
				}
			}
		}
	}

	return w
}

/*
func outputMillenialStepCountStatus(card scorecard.ScoredObject, okThreshold, warningThreshold int) (oks int, warnings int, criticals int) {

	if card.Grade >= scorecard.Grade(okThreshold) {
		oks++
	} else if card.Grade >= scorecard.Grade(warningThreshold) {
		warnings++
	} else {
		criticals++
	}
	return
}

func outputMillenialStep(card scorecard.TestScore, okThreshold, warningThreshold int) io.Reader {
	var col color.Attribute

	if card.Grade >= scorecard.Grade(okThreshold) {
		// Higher than or equal to --threshold-ok
		col = color.FgGreen
	} else if card.Grade >= scorecard.Grade(warningThreshold) {
		// Higher than or equal to --threshold-warning
		col = color.FgYellow
	} else {
		// All lower than both --threshold-ok and --threshold-warning are critical
		col = color.FgRed
	}

	w := bytes.NewBufferString("")

	color.New(col).Fprintf(w, "    [%s] %s\n", statusString(card.Grade, okThreshold, warningThreshold), card.Check.Name)

	for _, comment := range card.Comments {
		fmt.Fprintf(w, "        * ")

		if len(comment.Path) > 0 {
			fmt.Fprintf(w, "%s -> ", comment.Path)
		}

		fmt.Fprint(w, comment.Summary)

		if len(comment.Description) > 0 {
			fmt.Fprintf(w, "\n             %s", comment.Description)
		}

		fmt.Fprintln(w)
	}

	return w
}
*/

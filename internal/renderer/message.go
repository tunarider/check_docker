package renderer

import (
	"fmt"
	"github.com/tunarider/nagios-go-sdk/nagios"
)

type MessageResolver func(interface{}, []nagios.Performance) string

func NoProblemMessage(_ interface{}, performances []nagios.Performance) string {
	return nagios.MessageWithPerformance(
		"No problem",
		performances,
	)
}

func OutputFromPerformances(performances []nagios.Performance) []string {
	var messages []string
	for _, p := range performances {
		messages = append(
			messages,
			fmt.Sprintf("%s(%d/%d)", p.Label, p.Value, p.Max),
		)
	}
	return messages
}

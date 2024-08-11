package client

import (
	"context"
	"google.golang.org/appengine/log"
	"strings"
)

func ParseArgs(args []string) Args {
	var parsed []*Arg
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			if strings.Contains(arg, "=") {
				parts := strings.Split(arg, "=")
				parsed = append(parsed, &Arg{Name: parts[0][2:], Value: parts[1]})
			} else {
				parsed = append(parsed, &Arg{Name: arg[2:], Value: ""})
			}
		} else {
			log.Debugf(context.Background(), "Skipping arg %q", arg)
		}
	}

	return parsed

}

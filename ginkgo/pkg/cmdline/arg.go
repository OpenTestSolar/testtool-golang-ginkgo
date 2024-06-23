package cmdline

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/shlex"
)

type CommandArg struct {
	Key   string
	Value string
}

type CommandArgs struct {
	Args []*CommandArg
}

func NewCmdArgs() *CommandArgs {
	ca := &CommandArgs{}
	return ca
}

func NewCmdArgsWithDefauleArgs(cmdArgs []*CommandArg) *CommandArgs {
	ca := &CommandArgs{
		cmdArgs,
	}
	return ca
}

func NewCmdArgsParseByCmdLine(cmdline string) (*CommandArgs, error) {
	args, err := shlex.Split(cmdline)
	if err != nil {
		log.Printf("[ERROR] parse input extra cmdline args [%s] failed, err: %v", cmdline, err)
		return nil, err
	}
	if len(args) == 0 {
		log.Printf("[ERROR] unvalid input extra cmdline args [%s]", cmdline)
		return nil, fmt.Errorf("no args in cmdline: %s", cmdline)
	}
	var cmdArgs []*CommandArg
	for i := 0; i < len(args); i++ {
		if args[i] == "ginkgo" {
			cmdArgs = append(cmdArgs, &CommandArg{Key: "ginkgo", Value: ""})
			continue
		}
		key := ""
		value := ""
		if strings.HasPrefix(args[i], "-") {
			key = args[i]
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				value = "\"" + args[i+1] + "\""
			}
		}
		if key != "" {
			cmdArgs = append(cmdArgs, &CommandArg{Key: key, Value: value})
		}
	}
	return &CommandArgs{cmdArgs}, nil
}

func (ca *CommandArgs) Add(arg *CommandArg) {
	replace := false
	for _, a := range ca.Args {
		if a.Key == arg.Key {
			a.Value = arg.Value
			replace = true
		}
	}
	if !replace {
		ca.Args = append(ca.Args, arg)
	}
}

func (ca *CommandArgs) Extend(args []*CommandArg) {
	for _, arg := range args {
		ca.Add(arg)
	}
}

func (ca *CommandArgs) GetValueByKey(key string) string {
	for _, arg := range ca.Args {
		if arg.Key == key {
			return arg.Value
		}
	}
	return ""
}

func (ca *CommandArgs) AddOrReplaceArgs(args []*CommandArg) {
	for _, arg := range args {
		replace := false
		for _, a := range ca.Args {
			if a.Key == arg.Key {
				a.Value = arg.Value
				replace = true
				break
			}
		}
		if !replace {
			ca.Args = append(ca.Args, arg)
		}
	}
}

func (ca *CommandArgs) AddIfNotExists(args []*CommandArg) {
	for _, arg := range args {
		var exits bool
		for _, a := range ca.Args {
			if a.Key == arg.Key {
				exits = true
				break
			}
		}
		if !exits {
			ca.Args = append(ca.Args, arg)
		}
	}
}

func (ca *CommandArgs) Merge(mergedCmdArgs *CommandArgs) {
	ca.AddOrReplaceArgs(mergedCmdArgs.Args)
}

func (ca *CommandArgs) GenerateCmdLineStr() string {
	cmdLineItems := make([]string, 0)
	for _, arg := range ca.Args {
		if arg.Value == "" {
			cmdLineItems = append(cmdLineItems, arg.Key)
		} else if arg.Key == "" {
			cmdLineItems = append(cmdLineItems, arg.Value)
		} else {
			tmp := strings.Join([]string{arg.Key, arg.Value}, " ")
			cmdLineItems = append(cmdLineItems, tmp)
		}
	}
	cmdline := strings.Join(cmdLineItems, " ")
	return cmdline
}

func (ca *CommandArgs) NeedFocus() bool {
	if ok, err := strconv.ParseBool(os.Getenv("TESTSOLAR_TTP_FOCUS")); err != nil {
		return true
	} else {
		return ok
	}
}

package scrape

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/rsteube/carapace-spec/pkg/command"
	"gopkg.in/yaml.v3"
)

func Scrape(ctx kong.Context) {
	cmd := scrape(ctx.Model.Node)
	m, err := yaml.Marshal(cmd)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(string(m))
}

func scrape(node *kong.Node) command.Command {
	cmd := command.Command{
		Name:        node.Name,
		Aliases:     node.Aliases,
		Description: node.Help,
		Flags:       make(map[string]string),
		Commands:    make([]command.Command, 0),
	}
	cmd.Completion.Flag = make(map[string][]string)

	if group := node.Group; group != nil {
		cmd.Group = group.Key
	}

	for _, flag := range node.Flags {
		formatted := ""

		if flag.Short != 0 {
			formatted += fmt.Sprintf("-%v, ", string(flag.Short))
		}
		formatted += fmt.Sprintf("--%v", flag.Name)

		switch {
		case flag.IsBool():
		//case optionalArgument:
		//	formatted += "?"
		default:
			formatted += "="
		}

		if flag.IsCounter() || flag.IsCumulative() {
			formatted += "*"
		}
		cmd.Flags[formatted] = flag.Help

		if flag.Enum != "" {
			splitted := strings.Split(flag.Enum, ",")
			for index, v := range splitted {
				splitted[index] = strings.TrimSpace(v)
			}
			cmd.Completion.Flag[flag.Name] = splitted
		} else if tag := flag.Flag.Tag; tag != nil {
			switch tag.Type {
			case "path", "existingfile":
				cmd.Completion.Flag[flag.Name] = []string{"$files"}
			case "existingdir":
				cmd.Completion.Flag[flag.Name] = []string{"$directories"}
			}
		}
	}

	for _, subcmd := range node.Children {
		if !subcmd.Hidden {
			cmd.Commands = append(cmd.Commands, scrape(subcmd))
		}
	}
	return cmd
}

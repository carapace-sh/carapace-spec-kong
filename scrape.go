package scrape

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/rsteube/carapace-spec/pkg/command"
	"gopkg.in/yaml.v3"
)

type spec struct{}

func (s spec) Run(ctx *kong.Context) (err error) {
	var m []byte
	if m, err = yaml.Marshal(Command(ctx.Model.Node)); err == nil {
		fmt.Fprintln(ctx.Stdout, string(m))
	}
	return
}

var Plugin struct {
	Carapace struct {
		Spec spec `cmd:"" name:"spec"`
	} `cmd:"" name:"_carapace" hidden:""`
}

func Command(node *kong.Node) command.Command {
	cmd := command.Command{
		Name:        node.Name,
		Aliases:     node.Aliases,
		Description: node.Help,
		Commands:    make([]command.Command, 0),
	}
	cmd.Completion.Flag = make(map[string][]string)

	if group := node.Group; group != nil {
		cmd.Group = group.Key
	}

	for _, flag := range node.Flags {
		f := command.Flag{
			Longhand:   "--" + flag.Name,
			Value:      !flag.IsBool(),
			Repeatable: flag.IsCounter() || flag.IsCumulative(),
			Required:   flag.Required,
			Usage:      flag.Help,
		}
		if flag.Short != 0 {
			f.Shorthand = "-" + string(flag.Short)
		}

		cmd.AddFlag(f)

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
			cmd.Commands = append(cmd.Commands, Command(subcmd))
		}
	}
	return cmd
}

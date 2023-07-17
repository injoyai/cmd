package main

import (
	"github.com/injoyai/conv"
	"github.com/spf13/cobra"
)

type Flags struct {
	m map[string]*Flag
	conv.Extend
}

func newFlags(list []*Flag) *Flags {
	f := &Flags{m: make(map[string]*Flag)}
	for _, v := range list {
		f.m[v.Name] = v
	}
	f.Extend = conv.NewExtend(f)
	return f
}

func (this *Flags) Range(fn func(key string, val *Flag) bool) {
	for k, v := range this.m {
		if !fn(k, v) {
			break
		}
	}
}

func (this *Flags) GetVar(key string) *conv.Var {
	val, ok := this.m[key]
	if ok && len(val.Value) > 0 {
		return conv.New(val.Value)
	}
	return conv.Nil()
}

type Flag struct {
	Name     string
	Short    string
	DefValue string
	Memo     string
	Value    string
}

type Command struct {
	Flag []*Flag
	*cobra.Command

	Use     string
	Short   string
	Long    string
	Example string
	Run     RunFunc
	Child   []*Command
}

func (this *Command) command() *cobra.Command {
	if this.Command == nil {
		this.Command = &cobra.Command{}
	}
	for _, v := range this.Flag {
		this.Command.PersistentFlags().StringVarP(&v.Value, v.Name, v.Short, v.DefValue, v.Memo)
	}

	this.Command.Use = conv.SelectString(this.Command.Use == "", this.Use, this.Command.Use)
	this.Command.Short = conv.SelectString(this.Command.Short == "", this.Short, this.Command.Short)
	this.Command.Long = conv.SelectString(this.Command.Long == "", this.Long, this.Command.Long)
	this.Command.Example = conv.SelectString(this.Command.Example == "", this.Example, this.Command.Example)
	this.Command.Run = func(cmd *cobra.Command, args []string) {
		if this.Run != nil {
			this.Run(cmd, args, newFlags(this.Flag))
		}
	}
	for _, v := range this.Child {
		this.Command.AddCommand(v.command())
	}
	return this.Command
}

type RunFunc func(cmd *cobra.Command, args []string, flag *Flags)

type ICommand interface {
	command() *cobra.Command
}

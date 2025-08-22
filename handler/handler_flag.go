package handler

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

func NewFlags(list []*Flag) *Flags {
	return newFlags(list)
}

func newFlagRunType() *Flags {
	return newFlags([]*Flag{{Name: "runType", Value: "start"}})
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
	Name         string
	Short        string
	DefaultValue string
	Memo         string
	Value        string
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

func (this *Command) Deal(flags ...*Flag) *cobra.Command {
	if this.Command == nil {
		this.Command = &cobra.Command{}
	}
	for _, v := range this.Flag {
		this.Command.PersistentFlags().StringVarP(&v.Value, v.Name, v.Short, v.DefaultValue, v.Memo)
	}
	flags = append(this.Flag, flags...)

	this.Command.Use = conv.Select(this.Command.Use == "", this.Use, this.Command.Use)
	this.Command.Short = conv.Select(this.Command.Short == "", this.Short, this.Command.Short)
	this.Command.Long = conv.Select(this.Command.Long == "", this.Long, this.Command.Long)
	this.Command.Example = conv.Select(this.Command.Example == "", this.Example, this.Command.Example)
	this.Command.Run = func(cmd *cobra.Command, args []string) {
		if this.Run != nil {
			this.Run(cmd, args, newFlags(flags))
		}
	}
	for _, v := range this.Child {
		this.Command.AddCommand(v.Deal(flags...))
	}
	return this.Command
}

type RunFunc func(cmd *cobra.Command, args []string, flag *Flags)

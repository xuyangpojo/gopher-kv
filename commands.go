package main

import (
	"strings"
)

// Command 命令结构
// @author xuyang
// @datetime 2025-6-24 7:00
type Command struct {
	Name        string
	Description string
	Usage       string
}

// Commands 命令列表
// @author xuyang
// @datetime 2025-6-24 7:00
var Commands = []Command{
	{
		Name:        "set",
		Description: "设置键值对",
		Usage:       "set \"key\" \"value\"",
	},
	{
		Name:        "get",
		Description: "获取键对应的值",
		Usage:       "get \"key\"",
	},
	{
		Name:        "setnx",
		Description: "仅当键不存在时设置值",
		Usage:       "setnx \"key\" \"value\"",
	},
	{
		Name:        "setxx",
		Description: "仅当键存在时设置值",
		Usage:       "setxx \"key\" \"value\"",
	},
	{
		Name:        "del",
		Description: "删除键值对",
		Usage:       "del \"key\"",
	},
	{
		Name:        "keys",
		Description: "获取所有键",
		Usage:       "keys",
	},
	{
		Name:        "kvs",
		Description: "获取所有键值对",
		Usage:       "kvs",
	},
	{
		Name:        "settime",
		Description: "设置键的过期时间（毫秒）",
		Usage:       "settime \"key\" (milliseconds)",
	},
	{
		Name:        "getlasttime",
		Description: "获取键的剩余生存时间（毫秒）",
		Usage:       "getlasttime \"key\"",
	},
	{
		Name:        "help",
		Description: "显示帮助信息",
		Usage:       "help",
	},
	{
		Name:        "quit",
		Description: "退出程序",
		Usage:       "quit",
	},
}

// FindSimilarCommands 查找相似命令
// @author xuyang
// @datetime 2025-6-24 7:00
func FindSimilarCommands(input string) []Command {
	var similar []Command
	inputLower := strings.ToLower(input)
	for _, cmd := range Commands {
		cmdLower := strings.ToLower(cmd.Name)
		if strings.Contains(cmdLower, inputLower) || strings.Contains(inputLower, cmdLower) {
			similar = append(similar, cmd)
		}
	}
	return similar
}

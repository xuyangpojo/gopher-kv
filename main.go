package main

import (
	"bufio"
	"fmt"
	"gopherkv/data"
	"os"
	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("-------------------------------------------------------")
	fmt.Println("   _____             _                 _  ____      __")
	fmt.Println("  / ____|           | |               | |/ /\\ \\    / /")
	fmt.Println(" | |  __  ___  _ __ | |__   ___ _ __  | ' /  \\ \\  / / ")
	fmt.Println(" | | |_ |/ _ \\| '_ \\| '_ \\ / _ \\ '__| |  <    \\ \\/ /  ")
	fmt.Println(" | |__| | (_) | |_) | | | |  __/ |    | . \\    \\  /   ")
	fmt.Println("  \\_____|\\___/| .__/|_| |_|\\___|_|    |_|\\_\\    \\/    ")
	fmt.Println("              | |                                     ")
	fmt.Println("              |_|                                     ")
	fmt.Println("欢迎使用 GopherKV 键值型内存数据库! :)")
	fmt.Println("访问 https://github.com/xuyangpojo/gopher-kv 以获取帮助")
	fmt.Println("-------------------------------------------------------")
	for {
		fmt.Print("gkv> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		fields := parseFields(line)
		if len(fields) == 0 {
			continue
		}
		switch strings.ToLower(fields[0]) {
		case "set":
			if len(fields) != 3 {
				fmt.Println("参数错误!")
				fmt.Println("用法: set \"key\" \"value\"")
				continue
			}
			data.DataGkvString.Set(fields[1], []byte(fields[2]))
			fmt.Println("OK")
		case "get":
			if len(fields) != 2 {
				fmt.Println("参数错误!")
				fmt.Println("用法: get \"key\"")
				continue
			}
			v, ok := data.DataGkvString.Get(fields[1])
			if ok {
				fmt.Println(string(v))
			} else {
				fmt.Println("(nil)")
			}
		case "setnx":
			if len(fields) != 3 {
				fmt.Println("参数错误!")
				fmt.Println("用法: setnx \"key\" \"value\"")
				continue
			}
			if data.DataGkvString.SetNX(fields[1], []byte(fields[2])) {
				fmt.Println("OK")
			} else {
				fmt.Println("插入失败,Key已存在")
			}
		case "setxx":
			if len(fields) != 3 {
				fmt.Println("参数错误!")
				fmt.Println("用法: setxx \"key\" \"value\"")
				continue
			}
			if data.DataGkvString.SetXX(fields[1], []byte(fields[2])) {
				fmt.Println("OK")
			} else {
				fmt.Println("插入失败,Key不存在")
			}
		case "del":
			if len(fields) != 2 {
				fmt.Println("参数错误!")
				fmt.Println("用法: del \"key\"")
				continue
			}
			data.DataGkvString.Delete(fields[1])
			fmt.Println("OK")
		case "keys":
			if len(fields) != 1 {
				fmt.Println("参数错误!")
				fmt.Println("用法: keys")
				continue
			}
			keys := data.DataGkvString.GetAllKeys()
			if len(keys) == 0 {
				fmt.Println("(empty list or set)")
			} else {
				for i, key := range keys {
					if i > 0 {
						fmt.Print(" ")
					}
					fmt.Printf("\"%s\"", key)
				}
				fmt.Println()
			}
		case "kvs":
			if len(fields) != 1 {
				fmt.Println("参数错误!")
				fmt.Println("用法: kvs")
				continue
			}
			kvs := data.DataGkvString.GetAllKVs()
			if len(kvs) == 0 {
				fmt.Println("(empty list or set)")
			} else {
				for k, v := range kvs {
					fmt.Println(k, " -> ", v)
				}
			}
		case "settime":
			if len(fields) != 3 {
				fmt.Println("参数错误!")
				fmt.Println("用法: settime \"key\" (milliseconds)")
				continue
			}
			num, _ := strconv.Atoi(fields[2])
			data.DataGkvString.SetTime(fields[1], num)
		case "getlasttime":
			if len(fields) != 2 {
				fmt.Println("参数错误!")
				fmt.Println("用法: getlasttime \"key\"")
				continue
			}
			ttl := data.DataGkvString.GetTTL(fields[1])
			switch ttl {
			case -1:
				fmt.Println("(nil)")
			case -2:
				fmt.Println("已过期")
			default:
				fmt.Printf("%d\n", ttl)
			}
		case "help":
			showHelp()
		case "quit":
			fmt.Println("再见! :D")
			return
		default:
			fmt.Println("未知命令: ", fields[0])
			showSimilarCommands(fields[0])
		}
	}
}

// showHelp 显示帮助信息
// @author xuyang
// @datetime 2025-6-24 7:00
func showHelp() {
	fmt.Println("GopherKV 目前支持的命令:")
	fmt.Println("================================================")
	for _, cmd := range Commands {
		fmt.Printf("%-10s - %s\n", cmd.Name, cmd.Description)
		fmt.Printf("  用法: %s\n", cmd.Usage)
		fmt.Println()
	}
}

// showSimilarCommands 显示相似命令建议
// @author xuyang
// @datetime 2025-6-24 7:00
// @param input string 未知命令
func showSimilarCommands(input string) {
	similar := FindSimilarCommands(input)
	if len(similar) > 0 {
		fmt.Println("您是否在查找:")
		for _, cmd := range similar {
			fmt.Printf("  %-10s - %s\n", cmd.Name, cmd.Description)
		}
	} else {
		fmt.Println("输入 'help' 以查看所有可用命令")
	}
}

// parseFields 解析命令行
// @author xuyang
// @datetime 2025-6-24 7:00
// @param line string 整行输入
// @return []string 拆分命令
func parseFields(line string) []string {
	var fields []string
	var buf strings.Builder
	inQuotes := false
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '"' {
			inQuotes = !inQuotes
			continue
		}
		if c == ' ' && !inQuotes {
			if buf.Len() > 0 {
				fields = append(fields, buf.String())
				buf.Reset()
			}
			continue
		}
		buf.WriteByte(c)
	}
	if buf.Len() > 0 {
		fields = append(fields, buf.String())
	}
	return fields
}

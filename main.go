package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	store := make(map[string]collection.MyString)
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("--------------------------------------------------")
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
	fmt.Println("--------------------------------------------------")
	for {
		fmt.Print("gopherKV> ")
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
				fmt.Println("用法: set \"key\" \"value\"")
				continue
			}
			store[fields[1]] = collection.NewMyString(fields[2])
			fmt.Println("OK")
		case "get":
			if len(fields) != 2 {
				fmt.Println("用法: get \"key\"")
				continue
			}
			v, ok := store[fields[1]]
			if ok {
				fmt.Println(v.Value())
			} else {
				fmt.Println("(nil)")
			}
		case "setnx":
			if len(fields) != 3 {
				fmt.Println("用法: setnx \"key\" \"value\"")
				continue
			}
			_, exists := store[fields[1]]
			if !exists {
				store[fields[1]] = collection.NewMyString(fields[2])
				fmt.Println("OK")
			} else {
				fmt.Println("0")
			}
		case "quit":
			fmt.Println("Bye!")
			store = nil
			return
		default:
			fmt.Println("未知命令：", fields[0])
		}
	}
}

// parseFields 解析命令行，支持用双引号包裹的参数
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

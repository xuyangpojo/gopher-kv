package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"golang.org/x/term"
)

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

// showWithPager 使用分页器显示内容
// @author xuyang
// @datetime 2025-6-24 7:00
// @param content string 要显示的内容
// @return error 错误信息
func showWithPager(content string) error {
	// 尝试使用 less，如果不存在则使用 more
	var cmd *exec.Cmd
	if _, err := exec.LookPath("less"); err == nil {
		cmd = exec.Command("less", "-R") // -R 支持ANSI颜色
	} else if _, err := exec.LookPath("more"); err == nil {
		cmd = exec.Command("more")
	} else {
		return fmt.Errorf("未找到分页器")
	}
	// 设置标准输入输出
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// 执行命令
	return cmd.Run()
}

// showHelp 显示帮助信息
// @author xuyang
// @datetime 2025-6-24 7:00
func showHelp() {
	// 构建帮助内容
	var helpContent strings.Builder
	helpContent.WriteString("GopherKV 命令参考手册\n")
	helpContent.WriteString("===============================================\n\n")
	helpContent.WriteString("GopherKV 是一个轻量级的键值型内存数据库，支持以下命令：\n\n")
	for _, cmd := range Commands {
		helpContent.WriteString(fmt.Sprintf("命令: %s\n", cmd.Name))
		helpContent.WriteString(fmt.Sprintf("描述: %s\n", cmd.Description))
		helpContent.WriteString(fmt.Sprintf("用法: %s\n", cmd.Usage))
		helpContent.WriteString("-----------------------------------------------\n")
	}
	helpContent.WriteString("更多信息请访问: https://github.com/xuyangpojo/gopher-kv\n")
	// 尝试使用分页器显示
	if err := showWithPager(helpContent.String()); err != nil {
		// 如果分页器失败，回退到直接打印
		fmt.Println(helpContent.String())
	}
}

// InputHandler 输入处理器
type InputHandler struct {
	history     []string
	historyPos  int
	currentLine string
	reader      *bufio.Reader
}

// NewInputHandler 创建新的输入处理器
func NewInputHandler() *InputHandler {
	return &InputHandler{
		history:    make([]string, 0),
		historyPos: -1,
		reader:     bufio.NewReader(os.Stdin),
	}
}

// ReadLine 读取一行输入，支持历史记录
func (h *InputHandler) ReadLine(prompt string) (string, error) {
	fmt.Print(prompt)
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	var line strings.Builder
	cursorPos := 0
	for {
		char, _, err := h.reader.ReadRune()
		if err != nil {
			return "", err
		}
		switch char {
		case '\r', '\n': // 回车键
			result := line.String()
			if strings.TrimSpace(result) != "" {
				h.addToHistory(result)
			}
			// 确保光标回到行首，然后输出换行符
			fmt.Print("\r")
			fmt.Println() // 输出换行符，确保后续输出从新行开始
			return result, nil
		case 3: // Ctrl+C
			fmt.Println("^C")
			return "", fmt.Errorf("用户中断")
		case 4: // Ctrl+D
			fmt.Println("^D")
			return "", fmt.Errorf("EOF")
		case 127: // 退格键
			if cursorPos > 0 {
				lineStr := line.String()
				line.Reset()
				line.WriteString(lineStr[:cursorPos-1])
				line.WriteString(lineStr[cursorPos:])
				cursorPos--
				h.redrawLine(prompt, line.String(), cursorPos)
			}

		case 27: // ESC键
			// 检查是否是方向键
			nextChar, _, err := h.reader.ReadRune()
			if err != nil {
				continue
			}
			if nextChar == '[' {
				arrowChar, _, err := h.reader.ReadRune()
				if err != nil {
					continue
				}
				switch arrowChar {
				case 'A': // 上箭头
					h.navigateHistory(1, prompt, &line, &cursorPos)
				case 'B': // 下箭头
					h.navigateHistory(-1, prompt, &line, &cursorPos)
				case 'C': // 右箭头
					if cursorPos < line.Len() {
						cursorPos++
						h.moveCursor(1)
					}
				case 'D': // 左箭头
					if cursorPos > 0 {
						cursorPos--
						h.moveCursor(-1)
					}
				}
			}
		default:
			// 普通字符
			if char >= 32 && char <= 126 { // 可打印字符
				lineStr := line.String()
				line.Reset()
				line.WriteString(lineStr[:cursorPos])
				line.WriteRune(char)
				line.WriteString(lineStr[cursorPos:])
				cursorPos++
				// 重新显示当前行
				h.redrawLine(prompt, line.String(), cursorPos)
			}
		}
	}
}

// addToHistory 添加命令到历史记录
func (h *InputHandler) addToHistory(cmd string) {
	// 避免重复添加相同的命令
	if len(h.history) == 0 || h.history[len(h.history)-1] != cmd {
		h.history = append(h.history, cmd)
	}
	h.historyPos = -1
}

// navigateHistory 导航历史记录
func (h *InputHandler) navigateHistory(direction int, prompt string, line *strings.Builder, cursorPos *int) {
	if len(h.history) == 0 {
		return
	}
	newPos := h.historyPos + direction
	if newPos >= -1 && newPos < len(h.history) {
		h.historyPos = newPos
		if h.historyPos == -1 {
			// 回到当前输入
			line.Reset()
			line.WriteString(h.currentLine)
			*cursorPos = line.Len()
		} else {
			// 显示历史记录
			line.Reset()
			line.WriteString(h.history[len(h.history)-1-h.historyPos])
			*cursorPos = line.Len()
		}
		h.redrawLine(prompt, line.String(), *cursorPos)
	}
}

// redrawLine 重新绘制当前行
func (h *InputHandler) redrawLine(prompt, line string, cursorPos int) {
	// 清除当前行并重新开始
	fmt.Print("\r\033[K")
	fmt.Print(prompt)
	fmt.Print(line)
	// 将光标移动到正确位置
	if cursorPos < len(line) {
		fmt.Printf("\033[%dD", len(line)-cursorPos)
	}
}

// moveCursor 移动光标
func (h *InputHandler) moveCursor(direction int) {
	if direction > 0 {
		fmt.Printf("\033[%dC", direction)
	} else {
		fmt.Printf("\033[%dD", -direction)
	}
}

// saveCurrentLine 保存当前输入行
func (h *InputHandler) saveCurrentLine(line string) {
	h.currentLine = line
}

// SaveHistoryToFile 将历史命令保存到文件
func (h *InputHandler) SaveHistoryToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, cmd := range h.history {
		_, err := file.WriteString(cmd + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadAndExecHistoryFromFile 从文件加载命令并执行
func LoadAndExecHistoryFromFile(filename string, execFunc func(string)) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			execFunc(line)
		}
	}
	return scanner.Err()
}

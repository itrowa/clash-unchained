package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"github.com/itrowa/clash-unchained/internal/config"
	"github.com/itrowa/clash-unchained/internal/generator"
)

const configFilePath = "config.yaml"

// RunWizard runs the interactive setup wizard. It collects user input, saves
// config.yaml, generates the injection script, and prints next-step instructions.
// If outputFile is non-empty (from -o flag), it skips asking the user for a filename.
func RunWizard(outputFile string) error {
	scanner := bufio.NewScanner(os.Stdin)

	printHeader("clash-unchained 配置向导")
	fmt.Println("  按回车键使用 [括号内] 的默认值")
	fmt.Println()

	// ── Step 1: Residential proxy credentials ──────────────────────────────
	printStep(1, "住宅代理信息", "这是 AI 服务最终看到的出口 IP")

	server := mustPrompt(scanner, "代理服务器地址", "", "例：1.2.3.4 或 proxy.example.com")
	portStr := promptLine(scanner, "端口", "443")
	port, err := strconv.Atoi(strings.TrimSpace(portStr))
	if err != nil || port <= 0 || port > 65535 {
		return fmt.Errorf("端口无效：%s", portStr)
	}
	username := mustPrompt(scanner, "用户名", "", "")
	password, err := promptPassword(scanner, "密码")
	if err != nil {
		return fmt.Errorf("读取密码失败：%w", err)
	}

	// ── Step 2: Subscription linkage ───────────────────────────────────────
	printStep(2, "Clash 订阅配置", "通过哪个订阅节点组连接到住宅 IP")

	dialerProxy := promptLine(scanner, "订阅里的主代理组名称", "Proxies")
	fmt.Println("  提示：打开 Clash Verge，找到最顶层的手动选择组，通常叫 Proxies 或 节点选择")
	fmt.Println()

	// ── Step 3: Display names ──────────────────────────────────────────────
	printStep(3, "Clash UI 显示名称", "会出现在 Clash 节点列表和策略组里，可直接回车使用默认")

	nodeName := promptLine(scanner, "住宅节点名称", "My-Residential-IP")
	groupName := promptLine(scanner, "AI 路由组名称", "AI-Services")

	// ── Step 4: Optional features ──────────────────────────────────────────
	printStep(4, "可选功能", "")

	tailscaleAnswer := promptLine(scanner, "启用 Tailscale 直连（*.ts.net 走 DIRECT）", "y")
	tailscale := strings.ToLower(strings.TrimSpace(tailscaleAnswer)) == "y"

	// ── Step 5: Output file ────────────────────────────────────────────────
	if outputFile == "" {
		fmt.Println()
		outputFile = promptLine(scanner, "注入脚本输出文件名", "clash-script-injection.js")
	}

	// ── Build config ───────────────────────────────────────────────────────
	cfg := &config.Config{
		Nodes: []config.NodeConfig{
			{
				Name:        nodeName,
				Type:        "socks5",
				Server:      server,
				Port:        port,
				Username:    username,
				Password:    password,
				DialerProxy: dialerProxy,
			},
		},
		ProxyGroups: []config.ProxyGroupConfig{
			{
				Name:    groupName,
				Type:    "select",
				Proxies: []string{nodeName},
			},
		},
		AIDomains: config.AIDomainsConfig{
			ProxyGroup: groupName,
			UseBuiltin: true,
		},
	}
	if tailscale {
		cfg.ProxyGroups = append(cfg.ProxyGroups, config.ProxyGroupConfig{
			Name:            "Tailscale",
			Type:            "direct",
			TailscaleBypass: true,
		})
	}

	// ── Summary ────────────────────────────────────────────────────────────
	fmt.Println()
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("  配置摘要")
	fmt.Println("─────────────────────────────────────────")
	fmt.Printf("  住宅节点    : %s (%s:%d)\n", nodeName, server, port)
	fmt.Printf("  经由订阅组  : %s\n", dialerProxy)
	fmt.Printf("  AI 路由组   : %s\n", groupName)
	fmt.Printf("  Tailscale   : %s\n", yesNo(tailscale))
	fmt.Printf("  输出文件    : %s\n", outputFile)
	fmt.Println()

	confirm := promptLine(scanner, "确认并生成？", "y")
	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		fmt.Println("已取消。")
		return nil
	}

	// ── Save config.yaml ───────────────────────────────────────────────────
	if err := saveConfig(cfg); err != nil {
		return fmt.Errorf("保存配置文件失败：%w", err)
	}
	fmt.Printf("\n✓ 配置已保存 → %s\n", configFilePath)

	// ── Generate script ────────────────────────────────────────────────────
	script, err := generator.Generate(cfg)
	if err != nil {
		return fmt.Errorf("生成脚本失败：%w", err)
	}
	if err := os.WriteFile(outputFile, []byte(script), 0644); err != nil {
		return fmt.Errorf("写入脚本失败：%w", err)
	}
	fmt.Printf("✓ 注入脚本已生成 → %s\n", outputFile)

	// ── Next steps ─────────────────────────────────────────────────────────
	fmt.Println()
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("  下一步：安装到 Clash Verge")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("  1. 打开 Clash Verge → 配置")
	fmt.Println("  2. 找到你的订阅 → 右键 → 扩展脚本")
	fmt.Printf("  3. 将 %s 的内容粘贴到脚本编辑器\n", outputFile)
	fmt.Println("  4. 保存 → 刷新订阅 → 完成！")
	fmt.Println()

	return nil
}

// ── helpers ────────────────────────────────────────────────────────────────

func printHeader(title string) {
	fmt.Println()
	fmt.Println("─────────────────────────────────────────")
	fmt.Printf("  %s\n", title)
	fmt.Println("─────────────────────────────────────────")
	fmt.Println()
}

func printStep(n int, title, subtitle string) {
	fmt.Printf("第%d步：%s\n", n, title)
	if subtitle != "" {
		fmt.Printf("（%s）\n", subtitle)
	}
	fmt.Println()
}

// promptLine prints "  label [default]: " and returns the user's input,
// falling back to defaultVal if input is empty.
func promptLine(scanner *bufio.Scanner, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("  %s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("  %s: ", label)
	}
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			return line
		}
	}
	return defaultVal
}

// mustPrompt keeps asking until a non-empty value is provided.
func mustPrompt(scanner *bufio.Scanner, label, defaultVal, hint string) string {
	for {
		val := promptLine(scanner, label, defaultVal)
		if val != "" {
			return val
		}
		if hint != "" {
			fmt.Printf("  ! 必填，%s\n", hint)
		} else {
			fmt.Println("  ! 此项为必填")
		}
	}
}

// promptPassword reads a password from stdin with echo disabled.
// Falls back to the shared scanner when stdin is not a TTY (e.g. piped input).
func promptPassword(scanner *bufio.Scanner, label string) (string, error) {
	fmt.Printf("  %s: ", label)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // newline after hidden input
	if err != nil {
		// Non-TTY (piped input): reuse the shared scanner so we don't fight over stdin
		if scanner.Scan() {
			return strings.TrimSpace(scanner.Text()), nil
		}
		return "", fmt.Errorf("读取密码失败")
	}
	return strings.TrimSpace(string(b)), nil
}

func yesNo(b bool) string {
	if b {
		return "是"
	}
	return "否"
}

// saveConfig marshals cfg to YAML and writes it to config.yaml.
func saveConfig(cfg *config.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configFilePath, data, 0600)
}

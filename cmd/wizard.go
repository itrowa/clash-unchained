package cmd

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"github.com/itrowa/clash-unchained/internal/config"
	"github.com/itrowa/clash-unchained/internal/generator"
)

const configFilePath = "config.yaml"

// proxyCredentials holds parsed residential proxy connection info.
type proxyCredentials struct {
	server   string
	port     int
	username string
	password string
}

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

	creds := promptProxyCredentials(scanner)

	// ── Step 2: Subscription linkage ───────────────────────────────────────
	printStep(2, "Clash 订阅配置", "通过哪个订阅节点组连接到住宅 IP")

	fmt.Println("  打开 Clash Verge → 配置 → 点开你的订阅，找最顶层的手动选择组")
	fmt.Println("  常见名称：Proxies、节点选择、PROXY（注意大小写需完全一致）")
	fmt.Println()
	dialerProxy := promptValidated(scanner, field{
		label:      "订阅里的主代理组名称",
		defaultVal: "Proxies",
		example:    "Proxies  或  节点选择",
	}, validateNonEmpty)

	// ── Step 3: Display names ──────────────────────────────────────────────
	printStep(3, "Clash UI 显示名称", "随便起，会出现在 Clash 的节点列表和策略组里")

	nodeName := promptValidated(scanner, field{
		label:      "住宅节点名称",
		defaultVal: "My-Residential-IP",
		example:    "My-Residential-IP  或  住宅节点",
	}, validateNonEmpty)

	groupName := promptValidated(scanner, field{
		label:      "AI 路由组名称",
		defaultVal: "AI-Services",
		example:    "AI-Services  或  AI专用",
	}, validateNonEmpty)

	// ── Step 4: Optional features ──────────────────────────────────────────
	printStep(4, "可选功能", "")

	fmt.Println("  Tailscale 直连：开启后 *.ts.net 流量走 DIRECT，不经过任何代理")
	tailscale := promptYesNo(scanner, "启用 Tailscale 直连", "y")

	// ── Step 5: Output file ────────────────────────────────────────────────
	if outputFile == "" {
		fmt.Println()
		outputFile = promptValidated(scanner, field{
			label:      "注入脚本输出文件名",
			defaultVal: "clash-script-injection.js",
			example:    "clash-script-injection.js",
		}, validateFilename)
	}

	// ── Build config ───────────────────────────────────────────────────────
	cfg := &config.Config{
		Nodes: []config.NodeConfig{
			{
				Name:        nodeName,
				Type:        "socks5",
				Server:      creds.server,
				Port:        creds.port,
				Username:    creds.username,
				Password:    creds.password,
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
	fmt.Printf("  住宅节点    : %s  (%s:%d)\n", nodeName, creds.server, creds.port)
	fmt.Printf("  用户名      : %s\n", creds.username)
	fmt.Printf("  经由订阅组  : %s\n", dialerProxy)
	fmt.Printf("  AI 路由组   : %s\n", groupName)
	fmt.Printf("  Tailscale   : %s\n", yesNo(tailscale))
	fmt.Printf("  输出文件    : %s\n", outputFile)
	fmt.Println()

	if !promptYesNo(scanner, "确认并生成", "y") {
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

// ── proxy credential input ─────────────────────────────────────────────────

// promptProxyCredentials first offers a one-line paste, then falls back to
// field-by-field entry if the user presses Enter or the input can't be parsed.
func promptProxyCredentials(scanner *bufio.Scanner) proxyCredentials {
	fmt.Println("  代理供应商后台通常提供一键复制的连接字符串，可直接粘贴：")
	fmt.Println()
	fmt.Println("    格式一：host:port:user:pass")
	fmt.Println("            50.3.64.223:443:MyUser:MyPass")
	fmt.Println("    格式二：socks5://user:pass@host:port")
	fmt.Println("            socks5://MyUser:MyPass@50.3.64.223:443")
	fmt.Println("    格式三：user:pass@host:port")
	fmt.Println("            MyUser:MyPass@50.3.64.223:443")
	fmt.Println("    格式四：host:port@user:pass")
	fmt.Println("            50.3.64.223:443@MyUser:MyPass")
	fmt.Println()
	fmt.Print("  粘贴连接字符串（直接回车改为逐项填写）: ")

	var raw string
	if scanner.Scan() {
		raw = strings.TrimSpace(scanner.Text())
	}

	if raw != "" {
		creds, err := parseProxyString(raw)
		if err == nil {
			fmt.Println()
			fmt.Printf("  ✓ 解析成功：%s:%d  用户名：%s\n", creds.server, creds.port, creds.username)
			fmt.Println()
			return *creds
		}
		fmt.Printf("  ✗ 无法解析：%s\n", err)
		fmt.Println("  改为逐项填写：")
		fmt.Println()
	}

	// Fallback: field by field
	return promptProxyFields(scanner)
}

// promptProxyFields collects server, port, username, password individually.
func promptProxyFields(scanner *bufio.Scanner) proxyCredentials {
	server := promptValidated(scanner, field{
		label:   "代理服务器地址",
		hint:    "只填地址，不含端口和协议前缀",
		example: "50.3.64.223  或  proxy.example.com",
	}, validateServer)

	portStr := promptValidated(scanner, field{
		label:      "端口",
		defaultVal: "443",
		example:    "443  或  1080  或  8080",
	}, validatePort)
	port, _ := strconv.Atoi(portStr)

	username := promptValidated(scanner, field{
		label:   "用户名",
		example: "MyUser",
	}, validateNonEmpty)

	var password string
	for {
		var err error
		password, err = promptPassword(scanner, "密码")
		if err != nil {
			fmt.Println("  ✗ 读取密码失败，请重试")
			continue
		}
		if password == "" {
			fmt.Println("  ✗ 密码不能为空")
			continue
		}
		fmt.Println()
		break
	}

	return proxyCredentials{server: server, port: port, username: username, password: password}
}

// ── connection string parser ───────────────────────────────────────────────

// parseProxyString understands the four common formats residential proxy
// providers use for one-click copy:
//
//	Format 1: host:port:user:pass
//	Format 2: socks5://user:pass@host:port  (any scheme accepted)
//	Format 3: user:pass@host:port
//	Format 4: host:port@user:pass
func parseProxyString(s string) (*proxyCredentials, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("空字符串")
	}

	// Format 2: scheme://user:pass@host:port — strip scheme first
	if idx := strings.Index(s, "://"); idx != -1 {
		s = s[idx+3:]
		return parseUserAtHost(s)
	}

	// Formats 3 & 4: contain exactly one @
	if strings.Contains(s, "@") {
		atIdx := strings.Index(s, "@")
		left := s[:atIdx]
		// If left side looks like host:port (contains a dot or is an IP), it's format 4
		hostPart := strings.SplitN(left, ":", 2)[0]
		if strings.Contains(hostPart, ".") || net.ParseIP(hostPart) != nil {
			// Format 4: host:port@user:pass
			return parseHostAtUser(s)
		}
		// Format 3: user:pass@host:port
		return parseUserAtHost(s)
	}

	// Format 1: host:port:user:pass — exactly four colon-separated fields
	// Use SplitN with 4 so a password containing ":" is kept intact
	parts := strings.SplitN(s, ":", 4)
	if len(parts) == 4 {
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("端口必须是数字，得到：%q", parts[1])
		}
		if err := validatePort(parts[1]); err != nil {
			return nil, err
		}
		if parts[0] == "" || parts[2] == "" || parts[3] == "" {
			return nil, fmt.Errorf("字段不能为空")
		}
		return &proxyCredentials{
			server:   parts[0],
			port:     port,
			username: parts[2],
			password: parts[3],
		}, nil
	}

	return nil, fmt.Errorf("无法识别格式，请检查是否完整（需包含地址、端口、用户名、密码）")
}

// parseUserAtHost parses "user:pass@host:port"
func parseUserAtHost(s string) (*proxyCredentials, error) {
	atIdx := strings.Index(s, "@")
	if atIdx < 0 {
		return nil, fmt.Errorf("缺少 @ 分隔符")
	}
	userPart := s[:atIdx]
	hostPart := s[atIdx+1:]

	user, pass, err := splitUserPass(userPart)
	if err != nil {
		return nil, err
	}
	host, port, err := splitHostPort(hostPart)
	if err != nil {
		return nil, err
	}
	return &proxyCredentials{server: host, port: port, username: user, password: pass}, nil
}

// parseHostAtUser parses "host:port@user:pass"
func parseHostAtUser(s string) (*proxyCredentials, error) {
	atIdx := strings.Index(s, "@")
	if atIdx < 0 {
		return nil, fmt.Errorf("缺少 @ 分隔符")
	}
	hostPart := s[:atIdx]
	userPart := s[atIdx+1:]

	host, port, err := splitHostPort(hostPart)
	if err != nil {
		return nil, err
	}
	user, pass, err := splitUserPass(userPart)
	if err != nil {
		return nil, err
	}
	return &proxyCredentials{server: host, port: port, username: user, password: pass}, nil
}

func splitHostPort(s string) (host string, port int, err error) {
	h, p, e := net.SplitHostPort(s)
	if e != nil {
		return "", 0, fmt.Errorf("地址格式错误 %q（应为 host:port）", s)
	}
	n, e := strconv.Atoi(p)
	if e != nil {
		return "", 0, fmt.Errorf("端口必须是数字，得到：%q", p)
	}
	if n < 1 || n > 65535 {
		return "", 0, fmt.Errorf("端口范围是 1–65535，得到：%d", n)
	}
	return h, n, nil
}

func splitUserPass(s string) (user, pass string, err error) {
	// Split on first colon only; password may contain colons
	idx := strings.Index(s, ":")
	if idx < 0 {
		return "", "", fmt.Errorf("用户名密码格式错误 %q（应为 user:pass）", s)
	}
	u, p := s[:idx], s[idx+1:]
	if u == "" {
		return "", "", fmt.Errorf("用户名不能为空")
	}
	if p == "" {
		return "", "", fmt.Errorf("密码不能为空")
	}
	return u, p, nil
}

// ── field descriptor ───────────────────────────────────────────────────────

type field struct {
	label      string
	defaultVal string
	hint       string
	example    string
}

// ── prompt helpers ─────────────────────────────────────────────────────────

func promptValidated(scanner *bufio.Scanner, f field, validate func(string) error) string {
	if f.hint != "" {
		fmt.Printf("  ℹ  %s\n", f.hint)
	}
	if f.example != "" {
		fmt.Printf("  示例：%s\n", f.example)
	}
	for {
		if f.defaultVal != "" {
			fmt.Printf("  %s [%s]: ", f.label, f.defaultVal)
		} else {
			fmt.Printf("  %s: ", f.label)
		}
		var raw string
		if scanner.Scan() {
			raw = strings.TrimSpace(scanner.Text())
		}
		if raw == "" {
			raw = f.defaultVal
		}
		if err := validate(raw); err != nil {
			fmt.Printf("  ✗  %s\n", err)
			continue
		}
		fmt.Println()
		return raw
	}
}

func promptYesNo(scanner *bufio.Scanner, label, defaultVal string) bool {
	fmt.Printf("  %s [%s]: ", label, defaultVal)
	var raw string
	if scanner.Scan() {
		raw = strings.ToLower(strings.TrimSpace(scanner.Text()))
	}
	if raw == "" {
		raw = strings.ToLower(defaultVal)
	}
	fmt.Println()
	return raw == "y" || raw == "yes"
}

func promptPassword(scanner *bufio.Scanner, label string) (string, error) {
	fmt.Printf("  %s: ", label)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		// Non-TTY fallback (piped input)
		if scanner.Scan() {
			return strings.TrimSpace(scanner.Text()), nil
		}
		return "", fmt.Errorf("读取密码失败")
	}
	fmt.Println()
	return strings.TrimSpace(string(b)), nil
}

// ── validators ────────────────────────────────────────────────────────────

func validateServer(s string) error {
	if s == "" {
		return fmt.Errorf("不能为空")
	}
	for _, prefix := range []string{"socks5://", "socks4://", "http://", "https://", "://"} {
		if strings.Contains(strings.ToLower(s), prefix) {
			return fmt.Errorf("不要包含协议前缀（如 socks5://），只填地址部分")
		}
	}
	if _, _, err := net.SplitHostPort(s); err == nil {
		return fmt.Errorf("不要把端口一起填进来，端口在下一项单独填写")
	}
	if strings.ContainsAny(s, " \t/\\?#@") {
		return fmt.Errorf("地址格式不正确，只填纯域名或 IP 地址")
	}
	return nil
}

func validatePort(s string) error {
	if s == "" {
		return fmt.Errorf("不能为空")
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("端口必须是数字，例如 443 或 1080")
	}
	if n < 1 || n > 65535 {
		return fmt.Errorf("端口范围是 1–65535")
	}
	return nil
}

func validateNonEmpty(s string) error {
	if s == "" {
		return fmt.Errorf("此项不能为空")
	}
	return nil
}

func validateFilename(s string) error {
	if s == "" {
		return fmt.Errorf("文件名不能为空")
	}
	if strings.ContainsAny(s, "/\\:*?\"<>|") {
		return fmt.Errorf("文件名包含非法字符")
	}
	if !strings.HasSuffix(s, ".js") {
		return fmt.Errorf("文件名应以 .js 结尾，例如 clash-script-injection.js")
	}
	return nil
}

// ── misc ──────────────────────────────────────────────────────────────────

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

func yesNo(b bool) string {
	if b {
		return "是"
	}
	return "否"
}

func saveConfig(cfg *config.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configFilePath, data, 0600)
}

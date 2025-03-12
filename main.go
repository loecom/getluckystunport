package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"net/url"
)

type Response struct {
	ModuleEnable bool `json:"ModuleEnable"`
	List         []StunNode `json:"list"`
	Ret          int `json:"ret"`
	Statistics   map[string]Statistics `json:"statistics"`
}

type StunNode struct {
	Key                string `json:"Key"`
	Name               string `json:"Name"`
	StunType           string `json:"StunType"`
	Enable             bool `json:"Enable"`
	DisablePortForward bool `json:"DisablePortForward"`
	LastLogs           string `json:"LastLogs"`
	StunLocalAddr      string `json:"StunLocalAddr"`
	TargetAddrList     []string `json:"TargetAddrList"`
	PublicAddr         string `json:"PublicAddr"`
	PublicAddrInfo     string `json:"PublicAddrInfo"`
	PublicAddrHistroy  []AddrRecord `json:"PublicAddrHistroy"`
	WebhookEnable      bool `json:"WebhookEnable"`
	WebhookProxy       string `json:"WebhookProxy"`
	WebhookCallTime    string `json:"WebhookCallTime"`
	WebhookCallResult  bool `json:"WebhookCallResult"`
	WebhookCallErrorMsg string `json:"WebhookCallErrorMsg"`
	WebhookCallHistroy []string `json:"WebhookCallHistroy"`
	GlobalWebhook      bool `json:"GlobalWebhook"`
	GlobalWebhookCallTime    string `json:"GlobalWebhookCallTime"`
	GlobalWebhookCallResult  bool `json:"GlobalWebhookCallResult"`
	GlobalWebhookCallErrorMsg string `json:"GlobalWebhookCallErrorMsg"`
	GlobalWebhookCallHistroy []string `json:"GlobalWebhookCallHistroy"`
	Options            Options `json:"Options"`
}

type AddrRecord struct {
	AddrRecord  string `json:"AddrRecord"`
	UpdateTime  string `json:"UpdateTime"`
}

type Options struct {
	SingleProxyMaxTCPConnections         int `json:"SingleProxyMaxTCPConnections"`
	SingleProxyMaxUDPReadTargetDatagoroutineCount int `json:"SingleProxyMaxUDPReadTargetDatagoroutineCount"`
	UDPSessionTimeout                   int `json:"UDPSessionTimeout"`
	SafeMode                            string `json:"SafeMode"`
	TCPListenTLS                        bool `json:"TCPListenTLS"`
	TCPRelayTLS                         bool `json:"TCPRelayTLS"`
	TCPRelayTLSServerName               string `json:"TCPRelayTLSServerName"`
	TCPRelayTLSInsecureSkipVerify       bool `json:"TCPRelayTLSInsecureSkipVerify"`
	TCPStreamEncryptionSource           bool `json:"TCPStreamEncryptionSource"`
	TCPStreamEncryptionAccept           bool `json:"TCPStreamEncryptionAccept"`
	TCPStreamEncryptionKey              string `json:"TCPStreamEncryptionKey"`
	SinglePortSpeedLimit                bool `json:"SinglePortSpeedLimit"`
	SinglePortSendSpeedLimit            int `json:"SinglePortSendSpeedLimit"`
	SinglePortReceSpeedLimit            int `json:"SinglePortReceSpeedLimit"`
	RuleSpeedLimit                      bool `json:"RuleSpeedLimit"`
	RuleSendSpeedLimit                  int `json:"RuleSendSpeedLimit"`
	RuleReceSpeedLimit                  int `json:"RuleReceSpeedLimit"`
	UDPPacketSize                       int `json:"UDPPacketSize"`
	UDPPacketSourceEncryption           bool `json:"UDPPacketSourceEncryption"`
	UDPPacketAcceptEncryption           bool `json:"UDPPacketAcceptEncryption"`
	UDPPacketEncryptionKey              string `json:"UDPPacketEncryptionKey"`
}

type Statistics struct {
	TrafficIn             int `json:"TrafficIn"`
	TrafficOut            int `json:"TrafficOut"`
	TCPCurrentConnections int `json:"TCPCurrentConnections"`
	UDPCurrentConnections int `json:"UDPCurrentConnections"`
}

func GetPortFromPublicAddr(publicAddr string) (string, error) {
	parts := strings.Split(publicAddr, ":")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid PublicAddr format: %s", publicAddr)
	}
	return parts[len(parts)-1], nil
}

func GenerateTargetURL(templateURL string, port string) (string, error) {
	// 替换port占位符
	newURL := strings.ReplaceAll(templateURL, "port", port)
	// 解析URL以确保合法
	parsedURL, err := url.Parse(newURL)
	if err != nil {
		return "", fmt.Errorf("invalid target URL: %w", err)
	}
	return parsedURL.String(), nil
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run main.go <Name> <First URL> <Third URL Template>")
		os.Exit(1)
	}

	name := os.Args[1]
	firstURL := os.Args[2]
	thirdURLTemplate := os.Args[3]

	// 发送第一个GET请求
	firstResp, err := http.Get(firstURL)
	if err != nil {
		fmt.Printf("Failed to send first request: %v\n", err)
		os.Exit(1)
	}
	defer firstResp.Body.Close()

	if firstResp.StatusCode != http.StatusOK {
		fmt.Printf("First request failed with status code: %d\n", firstResp.StatusCode)
		os.Exit(1)
	}

	// 解析JSON响应
	var response Response
	if err := json.NewDecoder(firstResp.Body).Decode(&response); err != nil {
		fmt.Printf("Failed to parse JSON response: %v\n", err)
		os.Exit(1)
	}

	// 查找指定Name的节点
	var targetNode *StunNode
	for _, node := range response.List {
		if node.Name == name {
			targetNode = &node
			break
		}
	}

	if targetNode == nil {
		fmt.Printf("No node found with Name '%s' in first response\n", name)
		os.Exit(1)
	}

	publicAddr := targetNode.PublicAddr
	port, err := GetPortFromPublicAddr(publicAddr)
	if err != nil {
		fmt.Printf("Failed to extract port from PublicAddr: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Port from PublicAddr (%s): %s\n", name, port)

	// 构造目标URL并发送GET请求
	targetURL, err := GenerateTargetURL(thirdURLTemplate, port)
	if err != nil {
		fmt.Printf("Failed to generate target URL: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated third URL: %s\n", targetURL)

	// 发送第三轮GET请求
	thirdResp, err := http.Get(targetURL)
	if err != nil {
		fmt.Printf("Failed to send third request to %s: %v\n", targetURL, err)
		os.Exit(1)
	}
	defer thirdResp.Body.Close()

	fmt.Printf("Third request status code: %d\n", thirdResp.StatusCode)
}
package gateway

import "net"

// LocalIP は自ホストの非ループバック IPv4 アドレスを返す
// (v0.2 の get_ip 相当: 最初の inet / 127.0.0.1 以外)。
// 取得できない場合は空文字を返す (webhook payload の情報項目のため実行は止めない)。
func LocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		if ip := ipNet.IP.To4(); ip != nil {
			return ip.String()
		}
	}
	return ""
}

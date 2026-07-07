package project

import "sort"

// InventoryGroupAll は全グループ横断の予約グループ名 (v0.2 互換)。
const InventoryGroupAll = "all"

// SelectInventoryHosts はグループに属するホスト一覧を昇順・重複排除で返す。
// グループ名 all は全グループのホストを対象とする (v0.2 の sort | uniq と同じ出力規則)。
func SelectInventoryHosts(groups map[string][]string, group string) []string {
	var hosts []string
	if group == InventoryGroupAll {
		for _, groupHosts := range groups {
			hosts = append(hosts, groupHosts...)
		}
	} else {
		hosts = append(hosts, groups[group]...)
	}

	seen := map[string]bool{}
	uniq := make([]string, 0, len(hosts))
	for _, h := range hosts {
		if h == "" || seen[h] {
			continue
		}
		seen[h] = true
		uniq = append(uniq, h)
	}
	sort.Strings(uniq)
	return uniq
}

// InventoryGroupExists はグループの存在を判定する。
// v0.2 と同じく、ホスト取得結果の有無で判定する (定義済みでも空グループは false)。
func InventoryGroupExists(groups map[string][]string, group string) bool {
	return len(SelectInventoryHosts(groups, group)) > 0
}

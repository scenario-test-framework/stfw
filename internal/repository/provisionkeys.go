package repository

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
)

// flattenProcessConfigExcept はプロセスの config/config.yml をフラット化して dst に上書きする。
// ただし stfw.process.{type} 直下のキーのうち keys (plugin.yml の provision_keys) は、
// 値ごと (リスト・マップの配下も含めて) 取り除いてから重ねる (AS-BUILT §4.7)。
// キー名の完全一致で選ぶため、名前の先頭が同じ別のキー (例: tool_version_label) は残る。
// 取り除いたキー名を昇順で返す。ファイルが無ければ何もしない。
func flattenProcessConfigExcept(path, processType string, keys []string, dst map[string]string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var root map[string]any
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	var removed []string
	if target := mapPath(root, "stfw", "process", processType); target != nil {
		for _, k := range keys {
			if _, ok := target[k]; ok {
				delete(target, k)
				removed = append(removed, k)
			}
		}
	}
	// 別の YAML パスが同じ env 名に展開される場合 (例: logfilter: {version: ..} → logfilter_version)
	// も provision key を上書きさせない。env 名がキーそのもの、またはリストの要素 ({key}_{数字}…) なら捨てる。
	local := map[string]string{}
	flattenValue("", root, local)
	for k, v := range local {
		if isProvisionEnvName(k, processType, keys) {
			continue
		}
		dst[k] = v
	}
	sort.Strings(removed)
	return removed, nil
}

// isProvisionEnvName は env 名が provision key のもの (キーそのもの、またはリストの要素
// {key}_{数字}…) かを返す。名前の先頭が同じ別のキー (例: {key}_label) は含めない。
func isProvisionEnvName(envName, processType string, keys []string) bool {
	for _, k := range keys {
		prefix := "stfw_process_" + processType + "_" + k
		if envName == prefix {
			return true
		}
		rest, ok := strings.CutPrefix(envName, prefix+"_")
		if !ok {
			continue
		}
		digits := rest
		if i := strings.IndexByte(rest, '_'); i >= 0 {
			digits = rest[:i]
		}
		if digits != "" && strings.Trim(digits, "0123456789") == "" {
			return true
		}
	}
	return false
}

// mapPath は YAML をデコードした値から、キーをたどったマップを返す (無ければ nil)。
func mapPath(root map[string]any, keys ...string) map[string]any {
	cur := root
	for _, k := range keys {
		next, ok := cur[k].(map[string]any)
		if !ok {
			return nil
		}
		cur = next
	}
	return cur
}

// ProvisionKeyInProcessConfig はプロセスの config/config.yml に書かれた provision key 1 件。
// 書かれていても使われないため、validate が警告する (AS-BUILT §3.5)。
type ProvisionKeyInProcessConfig struct {
	ProcessPath string // プロジェクトルート相対のプロセスパス
	ProcessType string
	Key         string // stfw.process.{type} の下のキー名
}

// CheckProvisionKeysInProcessConfig は、シナリオの各プロセス (parallel の子を含む) の
// config/config.yml に、プラグインの plugin.yml が provision_keys に宣言したキーが
// 書かれていないかを調べる (値が空のリスト・null でも書かれていれば検出する)。
func CheckProvisionKeysInProcessConfig(projDir string, views []scenario.ScenarioView) ([]ProvisionKeyInProcessConfig, error) {
	var found []ProvisionKeyInProcessConfig
	for _, sv := range views {
		for _, bv := range sv.Bizdates {
			for _, pv := range bv.Processes {
				f, err := checkProcessProvisionKeys(projDir, []string{scenario.RootDirName, sv.Name, bv.DirName}, pv)
				if err != nil {
					return nil, err
				}
				found = append(found, f...)
			}
		}
	}
	return found, nil
}

// checkProcessProvisionKeys はプロセス 1 件 (parallel の場合は子を含む) を調べる。
func checkProcessProvisionKeys(projDir string, parents []string, pv scenario.ProcessView) ([]ProvisionKeyInProcessConfig, error) {
	var found []ProvisionKeyInProcessConfig
	segments := append(append([]string(nil), parents...), pv.DirName)
	// ディレクトリ名の形式違反 (ProcessType="") は構造検証側で error になるため対象外
	if pv.ProcessType != "" {
		if loc, err := ResolveProcessPlugin(projDir, pv.ProcessType); err == nil {
			meta, err := LoadPluginMeta(loc)
			if err != nil {
				return nil, err
			}
			if len(meta.ProvisionKeys) > 0 {
				processDir := filepath.Join(append([]string{projDir}, segments...)...)
				removed, err := flattenProcessConfigExcept(filepath.Join(processDir, "config", "config.yml"),
					pv.ProcessType, meta.ProvisionKeys, map[string]string{})
				if err != nil {
					return nil, err
				}
				display := strings.Join(segments, "/")
				for _, key := range removed {
					found = append(found, ProvisionKeyInProcessConfig{ProcessPath: display, ProcessType: pv.ProcessType, Key: key})
				}
			}
		}
	}
	for _, cv := range pv.Children {
		f, err := checkProcessProvisionKeys(projDir, segments, cv)
		if err != nil {
			return nil, err
		}
		found = append(found, f...)
	}
	return found, nil
}

// stripProvisionKeys は既定 config.yml (stfw.process.{type} 配下) から provision key を取り除いた
// YAML を返す。雛形 (stfw new process) でプロセスの設定に版を複製しないため (AS-BUILT §4.7)。
// 取り除くキーが無ければ元のバイト列をそのまま返す (コメント・書式を保つ)。
// 取り除くキーのアンカーを他の値が参照していても壊れないよう、別名 (*alias) は実値に展開してから出力する。
func stripProvisionKeys(raw []byte, processType string, keys []string) ([]byte, error) {
	if len(keys) == 0 {
		return raw, nil
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("plugin config: %w", err)
	}
	target := mappingPath(&doc, "stfw", "process", processType)
	if target == nil {
		return raw, nil
	}
	// 取り除くキーのアンカーを参照する別名が壊れないよう、先に別名を実値へ展開する
	if err := expandAliases(&doc, map[*yaml.Node]bool{}); err != nil {
		return nil, err
	}
	drop := map[string]bool{}
	for _, k := range keys {
		drop[k] = true
	}
	kept := make([]*yaml.Node, 0, len(target.Content))
	removed := false
	for i := 0; i+1 < len(target.Content); i += 2 {
		if drop[target.Content[i].Value] {
			removed = true
			continue
		}
		kept = append(kept, target.Content[i], target.Content[i+1])
	}
	if !removed {
		return raw, nil
	}
	target.Content = kept
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return nil, fmt.Errorf("plugin config: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("plugin config: %w", err)
	}
	return buf.Bytes(), nil
}

// errAliasCycle は別名が自分自身を含む値を参照している (展開すると終わらない) ことを示す。
var errAliasCycle = fmt.Errorf("plugin config: an alias refers to a value that contains itself")

// expandAliases はノード木の別名 (*alias) を、参照先ノードの複製に置き換える。
// visiting は展開中のノード (循環の検出用)。
func expandAliases(n *yaml.Node, visiting map[*yaml.Node]bool) error {
	if visiting[n] {
		return errAliasCycle
	}
	visiting[n] = true
	defer delete(visiting, n)
	for i, c := range n.Content {
		if c.Kind == yaml.AliasNode && c.Alias != nil {
			cp, err := deepCopyNode(c.Alias, map[*yaml.Node]bool{})
			if err != nil {
				return err
			}
			cp.Anchor = ""
			n.Content[i] = cp
			c = cp
		}
		if err := expandAliases(c, visiting); err != nil {
			return err
		}
	}
	return nil
}

// deepCopyNode はノードを子まで複製する (別名は参照先の複製に展開する)。
// visiting は複製中のノード (循環の検出用)。
func deepCopyNode(n *yaml.Node, visiting map[*yaml.Node]bool) (*yaml.Node, error) {
	if visiting[n] {
		return nil, errAliasCycle
	}
	visiting[n] = true
	defer delete(visiting, n)
	if n.Kind == yaml.AliasNode && n.Alias != nil {
		return deepCopyNode(n.Alias, visiting)
	}
	cp := *n
	cp.Content = make([]*yaml.Node, len(n.Content))
	for i, c := range n.Content {
		cc, err := deepCopyNode(c, visiting)
		if err != nil {
			return nil, err
		}
		cp.Content[i] = cc
	}
	return &cp, nil
}

// mappingPath はドキュメントノードから、キーをたどったマッピングノードを返す (無ければ nil)。
func mappingPath(doc *yaml.Node, keys ...string) *yaml.Node {
	n := doc
	if n.Kind == yaml.DocumentNode {
		if len(n.Content) == 0 {
			return nil
		}
		n = n.Content[0]
	}
	for _, key := range keys {
		if n.Kind != yaml.MappingNode {
			return nil
		}
		var next *yaml.Node
		for i := 0; i+1 < len(n.Content); i += 2 {
			if n.Content[i].Value == key {
				next = n.Content[i+1]
				break
			}
		}
		if next == nil {
			return nil
		}
		n = next
	}
	if n.Kind != yaml.MappingNode {
		return nil
	}
	return n
}

// Package inventory は stfw inventory (list / exists) のビジネスフローを制御する。
package inventory

import (
	"fmt"
	"io"
	"strconv"

	"github.com/scenario-test-framework/stfw/internal/domain/project"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// List はグループに属するホスト一覧を改行区切りで出力する
// (v0.2 の inventory --list と出力互換。未定義グループは空出力)。
func List(out io.Writer, projDir, fileName, group string) error {
	groups, err := repository.LoadInventory(projDir, fileName)
	if err != nil {
		return err
	}
	for _, host := range project.SelectInventoryHosts(groups, group) {
		fmt.Fprintln(out, host)
	}
	return nil
}

// Arch はホストに設定された arch を出力する (未設定・未定義ホストは空行)。
// 収集系プラグインが logfilter 等のバイナリを arch 別に送り分けるために使う。
func Arch(out io.Writer, projDir, fileName, host string) error {
	hostArch, err := repository.LoadInventoryHostArch(projDir, fileName)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, hostArch[host])
	return nil
}

// Exists はグループの存在を true / false で出力する
// (v0.2 の inventory --is-exist と出力互換)。
func Exists(out io.Writer, projDir, fileName, group string) error {
	groups, err := repository.LoadInventory(projDir, fileName)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, strconv.FormatBool(project.InventoryGroupExists(groups, group)))
	return nil
}

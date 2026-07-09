// Package sshtrust は stfw ssh trust のビジネスフローを制御する
// (v0.2 の未配線関数 gen_ssh_server_key の正式コマンド化)。
package sshtrust

import (
	"fmt"
	"log/slog"

	"github.com/scenario-test-framework/stfw/internal/domain/project"
	"github.com/scenario-test-framework/stfw/internal/gateway"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// Trust は対象ホストの SSH サーバキーを ~/.ssh/known_hosts へ登録する。
// target が inventory のグループ名ならグループ内全ホスト、そうでなければ
// 単一ホストとして扱う。ホストごとに ssh-keygen -R (既存キー削除) +
// ssh-keyscan (再登録) を実行する (v0.2 の gen_ssh_server_key と同じ手順)。
func Trust(log *slog.Logger, projDir, inventoryFile, target string) error {
	hosts := resolveHosts(projDir, inventoryFile, target)

	knownHosts, err := gateway.KnownHostsPath()
	if err != nil {
		return err
	}

	for _, host := range hosts {
		exists, err := gateway.KnownHostsContains(knownHosts, host)
		if err != nil {
			return fmt.Errorf("ssh trust: %w", err)
		}
		if exists {
			// 登録済みはスキップ (v0.2 互換: already exists は正常終了)
			log.Info("SSH server key is already exists", "host", host)
			continue
		}

		if err := gateway.RemoveKnownHostKey(host); err != nil {
			return fmt.Errorf("ssh trust: %w", err)
		}
		if err := gateway.ScanHostKey(host, knownHosts); err != nil {
			return fmt.Errorf("ssh trust: %w", err)
		}
		log.Info("SSH server key was added", "host", host)
	}
	return nil
}

// resolveHosts は target を inventory グループ → 単一ホストの順で解決する。
// inventory が読めない場合も単一ホストとして扱う。
func resolveHosts(projDir, inventoryFile, target string) []string {
	groups, err := repository.LoadInventory(projDir, inventoryFile)
	if err == nil && project.InventoryGroupExists(groups, target) {
		return project.SelectInventoryHosts(groups, target)
	}
	return []string{target}
}

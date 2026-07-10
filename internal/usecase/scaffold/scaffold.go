// Package scaffold は stfw new (scenario / bizdate / process) のビジネスフローを制御する。
package scaffold

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"sort"

	"github.com/scenario-test-framework/stfw/internal/domain/project"
	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// Scenario はシナリオ scaffold を生成する。
// 生成先は {projDir}/scenario/{name} に固定 (v0.2 はカレントがシナリオルートで
// あることを要求していたが、プロジェクトにシナリオルートは 1 つなので固定できる)。
func Scenario(log *slog.Logger, out io.Writer, projDir, nameStr string) error {
	name, err := scenario.NewScenarioName(nameStr)
	if err != nil {
		return err
	}

	root := filepath.Join(projDir, scenario.RootDirName)
	if !repository.ProjectConfigExists(projDir, project.ConfigFileName) || !repository.DirExists(root) {
		return fmt.Errorf("%s is not scenario-root-dir", root)
	}

	created, err := repository.CreateNodeScaffold(root, name.String())
	if err != nil {
		return fmt.Errorf("scenario scaffold: %w", err)
	}

	printCreated(out, projDir, created)
	log.Info("scenario initialized", "name", name.String())
	return nil
}

// Bizdate は業務日付 scaffold を生成する。cwd はシナリオディレクトリであること。
func Bizdate(log *slog.Logger, out io.Writer, projDir, cwd, seqStr, bizdateStr string) error {
	seq, err := scenario.NewSeq(seqStr)
	if err != nil {
		return err
	}
	bizdate, err := scenario.NewBizdate(bizdateStr)
	if err != nil {
		return err
	}

	if !isHierarchyDir(projDir, cwd, scenario.IsScenarioDir) {
		return fmt.Errorf("%s is not scenario-dir", cwd)
	}

	dirName := scenario.BizdateDirName(seq, bizdate)
	created, err := repository.CreateNodeScaffold(cwd, dirName)
	if err != nil {
		return fmt.Errorf("bizdate scaffold: %w", err)
	}

	printCreated(out, projDir, created)
	log.Info("bizdate initialized", "dir", dirName)
	return nil
}

// Process はプロセス scaffold を生成する。cwd は業務日付ディレクトリであること。
// プラグインの template/ を展開する (既存の同名ディレクトリは作り直し。v0.2 互換)。
func Process(log *slog.Logger, out io.Writer, projDir, cwd, seqStr, groupStr, processType string) error {
	seq, err := scenario.NewSeq(seqStr)
	if err != nil {
		return err
	}
	group, err := scenario.NewGroup(groupStr)
	if err != nil {
		return err
	}
	if err := scenario.ValidateProcessType(processType); err != nil {
		return err
	}

	if !isHierarchyDir(projDir, cwd, scenario.IsBizdateDir) {
		return fmt.Errorf("%s is not bizdate-dir", cwd)
	}

	// プラグイン解決 (プロジェクト plugins/ → 同梱の順)
	loc, err := repository.ResolveProcessPlugin(projDir, processType)
	if err != nil {
		return err
	}

	dirName := scenario.ProcessDirName(seq, group, processType)
	created, err := repository.CreateProcessScaffold(loc, cwd, dirName)
	if err != nil {
		return fmt.Errorf("process scaffold: %w", err)
	}

	printCreated(out, projDir, created)
	log.Info("process initialized", "dir", dirName, "type", processType)
	return nil
}

// ScaffoldFromSpec は spec (repository.ScenarioSpec) から scenario/bizdate/process の
// ディレクトリ骨格 (metadata.yml + config/config.yml) を生成する (spec → tree、往復の入口)。
// data/scripts/expect 等の葉は生成しない (往復対象は骨格のみ。plan §0)。
//
// 新規シナリオはそのまま生成する。既存シナリオに対しては sync=false ならエラーにし
// (誤上書き防止。既定は fail-safe)、sync=true なら spec との差分同期を行う:
//   - spec にあり disk に無い: 追加
//   - 両方にある: 維持 (metadata.yml / config.yml は spec で上書き、葉は温存)
//   - disk にあり spec に無い: 削除 (実装済みの葉ごと。破壊的)
func ScaffoldFromSpec(log *slog.Logger, out io.Writer, projDir string, spec repository.ScenarioSpec, sync bool) error {
	plan, err := planFromSpec(spec)
	if err != nil {
		return err
	}

	root := filepath.Join(projDir, scenario.RootDirName)
	if !repository.ProjectConfigExists(projDir, project.ConfigFileName) || !repository.DirExists(root) {
		return fmt.Errorf("%s is not scenario-root-dir", root)
	}

	scenarioDir := filepath.Join(root, plan.name)
	existed := repository.DirExists(scenarioDir)
	if existed && !sync {
		return fmt.Errorf("%s already exists (use --sync to update)", scenarioDir)
	}

	created, err := writeSpecPlan(scenarioDir, plan)
	if err != nil {
		return fmt.Errorf("scenario scaffold: %w", err)
	}
	printCreated(out, projDir, created)

	// 差分同期: spec に無い bizdate/process ディレクトリを削除する (追加・上書きの後に実施)。
	if sync && existed {
		removed, err := repository.PruneScenarioTree(scenarioDir, keptBizdateDirs(plan), keptProcessDirs(plan))
		if err != nil {
			return fmt.Errorf("scenario scaffold sync: %w", err)
		}
		printRemoved(out, projDir, removed)
		log.Info("scenario scaffold synced", "scenario", plan.name, "removed", len(removed))
	}

	log.Info("scenario scaffold generated", "scenario", plan.name, "bizdates", len(plan.bizdates))
	return nil
}

// keptBizdateDirs は plan に含まれる bizdate ディレクトリ名の集合を返す (prune で残す対象)。
func keptBizdateDirs(plan specPlan) map[string]bool {
	kept := make(map[string]bool, len(plan.bizdates))
	for _, b := range plan.bizdates {
		kept[b.dirName] = true
	}
	return kept
}

// keptProcessDirs は plan の各 bizdate 配下で残す process ディレクトリ名の集合を返す。
func keptProcessDirs(plan specPlan) map[string]map[string]bool {
	kept := make(map[string]map[string]bool, len(plan.bizdates))
	for _, b := range plan.bizdates {
		ps := make(map[string]bool, len(b.processes))
		for _, p := range b.processes {
			ps[p.dirName] = true
		}
		kept[b.dirName] = ps
	}
	return kept
}

// specPlan は spec を検証済みのディレクトリ名へ変換した中間表現。
// 書き込み前に spec 全体を検証しきることで、途中で VO 検証エラーになった場合の
// 部分書き込みを避ける。
type specPlan struct {
	name     string
	meta     repository.Metadata
	bizdates []specBizdatePlan
}

type specBizdatePlan struct {
	dirName   string
	meta      repository.Metadata
	processes []specProcessPlan
}

type specProcessPlan struct {
	dirName     string
	processType string
	meta        repository.Metadata
	config      map[string]any
}

// planFromSpec は spec を検証し specPlan を組み立てる。
// ディレクトリ名の組み立て・検証は既存の値オブジェクト (NewSeq/NewBizdate/NewGroup/
// ValidateProcessType) をそのまま再利用し、`stfw new` と同じ規則を通す。
func planFromSpec(spec repository.ScenarioSpec) (specPlan, error) {
	name, err := scenario.NewScenarioName(spec.Scenario)
	if err != nil {
		return specPlan{}, err
	}
	plan := specPlan{
		name: name.String(),
		meta: repository.Metadata{
			Description:               spec.Description,
			RequirementSpecifications: spec.RequirementSpecifications,
		},
	}

	// spec 内の bizdate/process が同一ディレクトリ名に衝突していないかを、書き込み前に
	// 検証する (衝突を許すと writeSpecPlan が後勝ちで silent 上書きしてしまうため)。
	seenBizdateDirs := map[string]bool{}

	for _, b := range spec.Bizdates {
		seq, err := scenario.NewSeq(b.Seq)
		if err != nil {
			return specPlan{}, err
		}
		bizdate, err := scenario.NewBizdate(b.Bizdate)
		if err != nil {
			return specPlan{}, err
		}
		bDirName := scenario.BizdateDirName(seq, bizdate)
		if seenBizdateDirs[bDirName] {
			return specPlan{}, fmt.Errorf("duplicate bizdate directory: %s", bDirName)
		}
		seenBizdateDirs[bDirName] = true

		bPlan := specBizdatePlan{
			dirName: bDirName,
			meta: repository.Metadata{
				Description:               b.Description,
				RequirementSpecifications: b.RequirementSpecifications,
			},
		}

		seenProcessDirs := map[string]bool{}
		for _, p := range b.Processes {
			pSeq, err := scenario.NewSeq(p.Seq)
			if err != nil {
				return specPlan{}, err
			}
			group, err := scenario.NewGroup(p.Group)
			if err != nil {
				return specPlan{}, err
			}
			if err := scenario.ValidateProcessType(p.Type); err != nil {
				return specPlan{}, err
			}
			pDirName := scenario.ProcessDirName(pSeq, group, p.Type)
			if seenProcessDirs[pDirName] {
				return specPlan{}, fmt.Errorf("duplicate process directory: %s in %s", pDirName, bDirName)
			}
			seenProcessDirs[pDirName] = true

			bPlan.processes = append(bPlan.processes, specProcessPlan{
				dirName:     pDirName,
				processType: p.Type,
				meta: repository.Metadata{
					Description:               p.Description,
					RequirementSpecifications: p.RequirementSpecifications,
				},
				config: p.Config,
			})
		}
		plan.bizdates = append(plan.bizdates, bPlan)
	}
	return plan, nil
}

// writeSpecPlan は検証済みの specPlan をディスクへ書き出す。
// 作成・上書きしたファイルの絶対パス一覧 (昇順) を返す。
func writeSpecPlan(scenarioDir string, plan specPlan) ([]string, error) {
	var created []string

	if err := repository.CreateSpecNode(scenarioDir, plan.meta); err != nil {
		return nil, err
	}
	created = append(created, filepath.Join(scenarioDir, "metadata.yml"))

	for _, b := range plan.bizdates {
		bDir := filepath.Join(scenarioDir, b.dirName)
		if err := repository.CreateSpecNode(bDir, b.meta); err != nil {
			return nil, err
		}
		created = append(created, filepath.Join(bDir, "metadata.yml"))

		for _, p := range b.processes {
			pDir := filepath.Join(bDir, p.dirName)
			if err := repository.CreateSpecNode(pDir, p.meta); err != nil {
				return nil, err
			}
			if err := repository.WriteProcessConfig(pDir, p.processType, p.config); err != nil {
				return nil, err
			}
			created = append(created, filepath.Join(pDir, "metadata.yml"), filepath.Join(pDir, "config", "config.yml"))
		}
	}

	sort.Strings(created)
	return created, nil
}

// isHierarchyDir は cwd がプロジェクト内の期待する階層かを判定する。
// v0.2 の is_*-dir 判定 (深さ + stfw.yml の存在) に対応する。
func isHierarchyDir(projDir, cwd string, isLevel func(rel string) bool) bool {
	if !repository.ProjectConfigExists(projDir, project.ConfigFileName) {
		return false
	}
	rel, err := filepath.Rel(projDir, cwd)
	if err != nil {
		return false
	}
	return isLevel(filepath.ToSlash(rel))
}

// printCreated は作成ファイルをプロジェクトルートからの相対パスで出力する。
func printCreated(out io.Writer, projDir string, created []string) {
	for _, p := range created {
		rel, err := filepath.Rel(projDir, p)
		if err != nil {
			rel = p
		}
		fmt.Fprintln(out, filepath.ToSlash(rel))
	}
}

// printRemoved は prune で削除したディレクトリを "removed: " プレフィックス付きで
// 出力する (作成行と区別できるようにする)。
func printRemoved(out io.Writer, projDir string, removed []string) {
	for _, p := range removed {
		rel, err := filepath.Rel(projDir, p)
		if err != nil {
			rel = p
		}
		fmt.Fprintln(out, "removed: "+filepath.ToSlash(rel))
	}
}

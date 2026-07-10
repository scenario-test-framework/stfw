package repository

import (
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
)

// ScenarioSpec は scenario ⇄ tree 往復の構造化 YAML (`{name}.yml`) の DTO。
// tree が真実の源であり、spec は往復可能な骨格のみを持つ (data/scripts/expect 等の葉は
// 対象外)。ディレクトリ構造ではなく DTO であるため domain には置かない。
type ScenarioSpec struct {
	Scenario                  string        `yaml:"scenario"`
	Description               string        `yaml:"description"`
	RequirementSpecifications []string      `yaml:"requirement_specifications,omitempty"`
	Bizdates                  []BizdateSpec `yaml:"bizdates,omitempty"`
}

// BizdateSpec は spec 内の業務日付 1 件分。
type BizdateSpec struct {
	Seq                       string        `yaml:"seq"`
	Bizdate                   string        `yaml:"bizdate"`
	Description               string        `yaml:"description"`
	RequirementSpecifications []string      `yaml:"requirement_specifications,omitempty"`
	Processes                 []ProcessSpec `yaml:"processes,omitempty"`
}

// ProcessSpec は spec 内のプロセス 1 件分。Config は config/config.yml の
// `stfw.process.{type}` サブツリーをそのまま持つ (キー順は Marshal 時に yaml.v3 が
// 昇順で決定論的に出力する)。
type ProcessSpec struct {
	Seq                       string         `yaml:"seq"`
	Group                     string         `yaml:"group"`
	Type                      string         `yaml:"type"`
	Description               string         `yaml:"description"`
	RequirementSpecifications []string       `yaml:"requirement_specifications,omitempty"`
	Config                    map[string]any `yaml:"config,omitempty"`
}

// MarshalSpec は ScenarioSpec を決定論的な YAML へ直列化する。
func MarshalSpec(spec ScenarioSpec) ([]byte, error) {
	return yaml.Marshal(spec)
}

// UnmarshalSpec は YAML を ScenarioSpec へパースする。
func UnmarshalSpec(raw []byte) (ScenarioSpec, error) {
	var spec ScenarioSpec
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		return ScenarioSpec{}, err
	}
	return spec, nil
}

// BuildSpecFromTree は走査済みシナリオ (ScenarioView) + metadata.yml + config/config.yml
// から spec を組み立てる (tree → spec、往復の出口)。
func BuildSpecFromTree(projDir string, view scenario.ScenarioView) (ScenarioSpec, error) {
	scenarioDir := filepath.Join(projDir, scenario.RootDirName, view.Name)
	meta, err := ReadNodeMetadata(scenarioDir)
	if err != nil {
		return ScenarioSpec{}, err
	}

	spec := ScenarioSpec{
		Scenario:                  view.Name,
		Description:               meta.Description,
		RequirementSpecifications: meta.RequirementSpecifications,
	}

	for _, b := range view.Bizdates {
		bDir := filepath.Join(scenarioDir, b.DirName)
		bMeta, err := ReadNodeMetadata(bDir)
		if err != nil {
			return ScenarioSpec{}, err
		}
		bSpec := BizdateSpec{
			Seq:                       b.Seq,
			Bizdate:                   b.Bizdate,
			Description:               bMeta.Description,
			RequirementSpecifications: bMeta.RequirementSpecifications,
		}

		for _, p := range b.Processes {
			pDir := filepath.Join(bDir, p.DirName)
			pMeta, err := ReadNodeMetadata(pDir)
			if err != nil {
				return ScenarioSpec{}, err
			}
			cfg, err := ReadProcessConfigSubtree(pDir, p.ProcessType)
			if err != nil {
				return ScenarioSpec{}, err
			}
			bSpec.Processes = append(bSpec.Processes, ProcessSpec{
				Seq:                       p.Seq,
				Group:                     p.Group,
				Type:                      p.ProcessType,
				Description:               pMeta.Description,
				RequirementSpecifications: pMeta.RequirementSpecifications,
				Config:                    cfg,
			})
		}
		spec.Bizdates = append(spec.Bizdates, bSpec)
	}
	return spec, nil
}

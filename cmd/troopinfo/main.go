package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	troopInfoPath     = "C:\\Program Files (x86)\\Steam\\steamapps\\common\\KUF Crusader\\Data\\SOX\\TroopInfo.sox"
	troopInfoYAMLPath = "C:\\Program Files (x86)\\Steam\\steamapps\\common\\KUF Crusader\\Data\\SOX\\TroopInfo.yaml"
)

const defaultLength = 4

var errInvalidSOX = errors.New("not a valid SOX file")

var (
	restore = flag.Bool("restore", false, "Restores TroopInfo.sox file using a backup")
	debug   = flag.Bool("debug", false, "Pretty-prints struct info to stdout")
	diff    = flag.Bool("diff", false, "Prints out a diff of what would be written and the current SOX file")
	write   = flag.Bool("write", false, "Writes TroopInfo.sox back to the source game directory")
	update  = flag.Bool("update", false, "Updates TroopInfo.yaml")
)

type levelUpData struct {
	SkillID       int32   `yaml:"skill_id"`
	SkillPerLevel float32 `yaml:"skill_per_level"`
}

type troopInfo struct {
	Job    int32 `yaml:"job"`     // troop Job type (defined in K2JobDef.h)
	TypeID int32 `yaml:"type_id"` // troop type ID (defined in K2TroopDef.h)

	MoveSpeed        float32 `yaml:"move_speed"`        // max move speed
	RotateRate       float32 `yaml:"rotate_rate"`       // max rotate rate
	MoveAcceleration float32 `yaml:"move_acceleration"` // move acceleration
	MoveDeceleration float32 `yaml:"move_deceleration"` // move deceleration

	SightRange float32 `yaml:"sight_range"` // visible range

	AttackRangeMax   float32 `yaml:"attack_range_max"`
	AttackRangeMin   float32 `yaml:"attack_range_min"`   // ranged attack range (0 if troop lacks ranged attack)
	AttackFrontRange float32 `yaml:"attack_front_range"` // frontal attack range (0 if troop lacks frontal attack)

	DirectAttack   float32 `yaml:"direct_attack"`   // direct attack strength (melee/frontal)
	IndirectAttack float32 `yaml:"indirect_attack"` // indirect attack strength (ranged)
	Defense        float32 `yaml:"defense"`         // defense strength

	BaseWidth float32 `yaml:"base_width"` // base troop size

	// resistance to attack types
	ResistMelee     float32 `yaml:"resist_melee"`
	ResistRanged    float32 `yaml:"resist_ranged"`
	ResistFrontal   float32 `yaml:"resist_frontal"`
	ResistExplosion float32 `yaml:"resist_explosion"`
	ResistFire      float32 `yaml:"resist_fire"`
	ResistIce       float32 `yaml:"resist_ice"`
	ResistLightning float32 `yaml:"resist_lightning"`
	ResistHoly      float32 `yaml:"resist_holy"`
	ResistCurse     float32 `yaml:"resist_curse"`
	ResistPoison    float32 `yaml:"resist_poison"`

	MaxUnitSpeedMultiplier float32 `yaml:"max_unit_speed_multiplier"`
	DefaultUnitHP          float32 `yaml:"default_unit_hp"`
	FormationRandom        int32   `yaml:"formation_random"`
	DefaultUnitNumX        int32   `yaml:"default_unit_num_x"`
	DefaultUnitNumY        int32   `yaml:"default_unit_num_y"`

	UnitHPLevUp float32 `yaml:"unit_hp_lev_up"`

	LevelUpData [3]levelUpData `yaml:"level_up_data"` // needs to be set to a length of 3

	DamageDistribution float32 `yaml:"damage_distribution"`
}

type troopInfoSOX struct {
	Version int32 `yaml:"version"`
	Count   int32 `yaml:"count"`

	TroopInfos [43]troopInfo `yaml:"troop_infos"`

	TheEnd [64]byte `yaml:"-"`
}

func main() {
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	troopNames := []string{
		"Archer",
		"Longbows",
		"Infantry",
		"Spearman",
		"Heavy Infantry",
		"Knight",
		"Paladin",
		"Calvary",
		"Heavy Calvary",
		"Storm Riders",
		"Sappers",
		"Pyro Techs",
		"Bomber Wings",
		"Mortar",
		"Ballista",
		"Harpoon",
		"Catapult",
		"Battaloon",
		"Dark Elves Archer",
		"Dark Elves Calvary Archers",
		"Dark Elves Infantry",
		"Dark Elves Knights",
		"Dark Elves Calvary",
		"Orc Infantry",
		"Orc Riders",
		"Orc Heavy Riders",
		"Orc Axe Man",
		"Orc Heavy Infantry",
		"Orc Sappers",
		"Orc Scorpion",
		"Orc Swamp Mammoth",
		"Orc Dirigible",
		"Orc Black Wyverns",
		"Orc Ghouls",
		"Orc Bone Dragon",
		"Wall Archers (Humans)",
		"Scouts",
		"Ghoul Selfdestruct",
		"Encablossa Monster (Melee)",
		"Encablossa Flying Monster",
		"Encablossa Monster (Ranged)",
		"Wall Archers (Elves)",
		"Encablossa Main",
	}

	file, err := os.Open(troopInfoPath)
	if err != nil {
		log.Fatal().Err(err)
	}
	defer file.Close()

	version := readInt32(file)
	count := readInt32(file)

	if !validSOX(version, count) {
		log.Fatal().Err(errInvalidSOX)
	}

	troopInfos := [43]troopInfo{}

	tis := troopInfoSOX{
		Version:    version,
		Count:      count,
		TroopInfos: troopInfos,
	}

	for i := range troopNames {
		ti := troopInfo{
			Job:    readInt32(file),
			TypeID: readInt32(file),

			MoveSpeed:        readFloat32(file),
			RotateRate:       readFloat32(file),
			MoveAcceleration: readFloat32(file),
			MoveDeceleration: readFloat32(file),

			SightRange: readFloat32(file),

			AttackRangeMax:   readFloat32(file),
			AttackRangeMin:   readFloat32(file),
			AttackFrontRange: readFloat32(file),

			DirectAttack:   readFloat32(file),
			IndirectAttack: readFloat32(file),
			Defense:        readFloat32(file),

			BaseWidth: readFloat32(file),

			ResistMelee:     readFloat32(file),
			ResistRanged:    readFloat32(file),
			ResistFrontal:   readFloat32(file),
			ResistExplosion: readFloat32(file),
			ResistFire:      readFloat32(file),
			ResistIce:       readFloat32(file),
			ResistLightning: readFloat32(file),
			ResistHoly:      readFloat32(file),
			ResistCurse:     readFloat32(file),
			ResistPoison:    readFloat32(file),

			MaxUnitSpeedMultiplier: readFloat32(file),
			DefaultUnitHP:          readFloat32(file),
			FormationRandom:        readInt32(file),
			DefaultUnitNumX:        readInt32(file),
			DefaultUnitNumY:        readInt32(file),

			UnitHPLevUp: readFloat32(file),

			LevelUpData: [3]levelUpData{
				{
					SkillID:       readInt32(file),
					SkillPerLevel: readFloat32(file),
				},
				{
					SkillID:       readInt32(file),
					SkillPerLevel: readFloat32(file),
				},
				{
					SkillID:       readInt32(file),
					SkillPerLevel: readFloat32(file),
				},
			},

			DamageDistribution: readFloat32(file),
		}

		tis.TroopInfos[i] = ti
	}

	buf := &bytes.Buffer{}

	// Read the rest of the file.
	for {
		data := make([]byte, defaultLength)

		_, err := file.Read(data)
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal().Err(err)
		}

		var chars []byte

		reader := bytes.NewReader(data)

		if err := binary.Read(reader, binary.LittleEndian, &chars); err != nil {
			log.Fatal().Err(err)
		}

		buf.Write(chars)
	}

	copy(tis.TheEnd[:], buf.Bytes())

	buf.Reset()

	if err := binary.Write(buf, binary.LittleEndian, &tis); err != nil {
		log.Fatal().Err(err)
	}

	if *restore {
		data, err := ioutil.ReadFile(troopInfoPath + ".bak")
		if err != nil {
			log.Fatal().Err(err)
		}

		if err := ioutil.WriteFile(troopInfoPath, data, 0600); err != nil {
			log.Fatal().Err(err)
		}

		log.Info().Msg("Success!")

		os.Exit(0)
	}

	if *debug {
		spew.Dump(tis)
		os.Exit(0)
	}

	if *update {
		buf.Reset()

		for i, name := range troopNames {
			buf.WriteString(fmt.Sprintf("# %d -- %s\n", i, name))
		}

		data, err := yaml.Marshal(tis)
		if err != nil {
			log.Fatal().Err(err)
		}

		buf.Write(data)

		if err := ioutil.WriteFile(troopInfoYAMLPath, buf.Bytes(), 0600); err != nil {
			log.Fatal().Err(err)
		}

		log.Info().Msg("Success!")
	}

	if *diff {
		tis = troopInfoSOX{}

		data, err := binaryData(tis)
		if err != nil {
			log.Fatal().Err(err)
		}

		if diff := cmp.Diff(data, buf.Bytes()); diff != "" {
			fmt.Printf("binary data mismatch (-want +got):\n%s", diff)
		}

		os.Exit(0)
	}

	if *write {
		tis = troopInfoSOX{}

		data, err := binaryData(tis)
		if err != nil {
			log.Fatal().Err(err)
		}

		if err := ioutil.WriteFile(troopInfoPath, data, 0600); err != nil {
			log.Fatal().Err(err)
		}

		log.Info().Msg("Success!")

		os.Exit(0)
	}
}

func binaryData(sox troopInfoSOX) ([]byte, error) {
	buf := &bytes.Buffer{}

	yamlData, err := ioutil.ReadFile(troopInfoYAMLPath)
	if err != nil {
		return buf.Bytes(), err
	}

	if err := yaml.Unmarshal(yamlData, &sox); err != nil {
		return buf.Bytes(), err
	}

	if err := binary.Write(buf, binary.LittleEndian, &sox); err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

func validSOX(version, count int32) bool {
	if version != 100 || count != 43 {
		return false
	}

	return true
}

func readInt32(file io.Reader) int32 {
	data := readBytes(file)

	return int32FromBytes(data)
}

func readFloat32(file io.Reader) float32 {
	data := readBytes(file)

	return float32FromBytes(data)
}

func int32FromBytes(data []byte) int32 {
	var i32 int32

	buf := bytes.NewReader(data)

	err := binary.Read(buf, binary.LittleEndian, &i32)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("binary.Read failed")
	}

	return i32
}

func float32FromBytes(data []byte) float32 {
	var f32 float32

	buf := bytes.NewReader(data)

	err := binary.Read(buf, binary.LittleEndian, &f32)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("binary.Read failed")
	}

	return f32
}

func readBytes(file io.Reader) []byte {
	data := make([]byte, defaultLength)

	_, err := file.Read(data)
	if err != nil && err != io.EOF {
		log.Fatal().Err(err)
	}

	return data
}

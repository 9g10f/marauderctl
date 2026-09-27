//go:build unix

package start

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"codeberg.org/9g10f/marauderctl/internal/script"
)

func ParseProtonVersion1(path string) (float64, error) {
	return strconv.ParseFloat(filepath.Base(path)[7:], 32)
}

func ParseProtonVersion3(path string) (float64, error) {
	return strconv.ParseFloat(strings.ReplaceAll(filepath.Base(path)[9:], "-", "."), 32)
}

func ParseProtonVersion5(path string) (float64, error) {
	return strconv.ParseFloat(strings.Split(filepath.Base(path)[7:], " ")[0], 32)
}

func GetProtons() ([]string, error) {
	// 1: Proton *.*
	// 2: Proton - Experimental
	// 3: GE-Proton*-*
	// 4: Proton Hotfix
	// 5: Proton *.* (Beta)

	var Protons1 []string
	var Protons2 []string
	var Protons3 []string
	var Protons4 []string
	var Protons5 []string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var protonPaths = []string{
		filepath.Join(homeDir, ".steam", "root", "steamapps", "common"),
		filepath.Join(homeDir, ".local", "share", "Steam", "steamapps", "common"),
		filepath.Join(homeDir, ".local", "share", "Steam", "compatibilitytools.d"),
	}

	for _, protonPath := range protonPaths {
		protons1, err := filepath.Glob(filepath.Join(protonPath, "Proton *.*"))
		if err != nil {
			return nil, err
		}

		for _, proton1 := range protons1 {
			if _, err := os.Stat(filepath.Join(proton1, "proton")); err == nil {
				_, err := ParseProtonVersion1(proton1)
				if err == nil {
					Protons1 = append(Protons1, proton1)
				}
			} else if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				return nil, err
			}
		}

		protons2, err := filepath.Glob(filepath.Join(protonPath, "Proton - Experimental"))
		if err != nil {
			return nil, err
		}

		for _, proton2 := range protons2 {
			if _, err := os.Stat(filepath.Join(proton2, "proton")); err == nil {
				Protons2 = append(Protons2, proton2)
			} else if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				return nil, err
			}
		}

		protons3, err := filepath.Glob(filepath.Join(protonPath, "GE-Proton*-*"))
		if err != nil {
			return nil, err
		}

		for _, proton3 := range protons3 {
			if _, err := os.Stat(filepath.Join(proton3, "proton")); err == nil {
				_, err := ParseProtonVersion3(proton3)
				if err == nil {
					Protons3 = append(Protons3, proton3)
				}
			} else if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				return nil, err
			}
		}

		protons4, err := filepath.Glob(filepath.Join(protonPath, "Proton Hotfix"))
		if err != nil {
			return nil, err
		}

		for _, proton4 := range protons4 {
			if _, err := os.Stat(filepath.Join(proton4, "proton")); err == nil {
				Protons4 = append(Protons4, proton4)
			} else if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				return nil, err
			}
		}

		protons5, err := filepath.Glob(filepath.Join(protonPath, "Proton *.* (Beta)"))
		if err != nil {
			return nil, err
		}

		for _, proton5 := range protons5 {
			if _, err := os.Stat(filepath.Join(proton5, "proton")); err == nil {
				_, err := ParseProtonVersion5(proton5)
				if err == nil {
					Protons5 = append(Protons5, proton5)
				}
			} else if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				return nil, err
			}
		}
	}

	sort.Slice(Protons1, func(i int, j int) bool {
		v1, _ := ParseProtonVersion1(Protons1[i])
		v2, _ := ParseProtonVersion1(Protons1[j])
		return v1 < v2
	})

	sort.Slice(Protons3, func(i int, j int) bool {
		v1, _ := ParseProtonVersion3(Protons3[i])
		v2, _ := ParseProtonVersion3(Protons3[j])
		return v1 < v2
	})

	sort.Slice(Protons5, func(i int, j int) bool {
		v1, _ := ParseProtonVersion5(Protons5[i])
		v2, _ := ParseProtonVersion5(Protons5[j])
		return v1 < v2
	})

	return append(Protons1, append(Protons2, append(Protons3, append(Protons4, Protons5...)...)...)...), err
}

func GetProton() (string, error) {
	protons, err := GetProtons()
	if err != nil {
		return "", err
	}

	if len(protons) == 0 {
		return "", errors.New("Proton not found")
	}

	return filepath.Join(protons[0], "proton"), nil
}

func ProtonRun(gameExe string, proton string, gameMeta cstructs.GameMeta) (*exec.Cmd, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(proton, "run", gameExe)

	gameId := gameMeta.GetId()

	err = os.MkdirAll(filepath.Join(homeDir, ".local", "share", "marauder", "compat", gameId), 0755)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(filepath.Join(homeDir, ".local", "share", "marauder", "logs"), 0755)
	if err != nil {
		return nil, err
	}

	vars, err := script.ParseLocalScriptVariables(gameMeta)
	if err != nil {
		return nil, err
	}

	env := os.Environ()

	env = append(env, "STEAM_COMPAT_DATA_PATH=" + filepath.Join(homeDir, ".local", "share", "marauder", "compat", gameId))
	env = append(env, "STEAM_COMPAT_CLIENT_INSTALL_PATH=" + filepath.Join(homeDir, ".local", "share", "Steam"))
	env = append(env, "PROTON_LOG=1")
	env = append(env, "PROTON_LOG_DIR=" + filepath.Join(homeDir, ".local", "share", "marauder", "logs"))

	steamAppId, ok := vars["steam-app-id"]
	if ok {
		env = append(env, "STEAM_COMPAT_APP_ID=" + steamAppId)
	}

	cmd.Env = env
	cmd.Dir = filepath.Dir(gameExe)

	return cmd, nil
}
//go:build unix

package start

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"codeberg.org/9g10f/marauderctl/internal/script"
)

// marauderctl automatically searches for Proton installations and chooses the best proton version
// It chooses based on which type it is first and then what version it is
// It follows this type priority list:
// 1: Proton *.*
// 2: Proton - Experimental
// 3: GE-Proton*-*
// 4: Proton Hotfix
// 5: Proton *.* (Beta)
// The lower the number the better
// You can obviously choose a custom Proton if either it isn't found or selected with --proton-path

// Returns Proton's version from it's directory name (Proton type 1: Proton *.*)
func ParseProtonVersion1(path string) (float64, error) {
	f, err := strconv.ParseFloat(filepath.Base(path)[7:], 32)
	if err != nil {
		return 0, fmt.Errorf("ParseProtonVersion1: Unable to parse Proton version from path: '%v'", path)
	}

	return f, nil
}

// Returns Proton's version from it's directory name (Proton type 3: GE-Proton*-*)
func ParseProtonVersion3(path string) (float64, error) {
	f, err := strconv.ParseFloat(strings.ReplaceAll(filepath.Base(path)[9:], "-", "."), 32)
	if err != nil {
		return 0, fmt.Errorf("ParseProtonVersion3: Unable to parse Proton version from path: '%v'", path)
	}

	return f, nil
}

// Returns Proton's version from it's directory name (Proton type 5: Proton *.* (Beta))
func ParseProtonVersion5(path string) (float64, error) {
	f, err := strconv.ParseFloat(strings.Split(filepath.Base(path)[7:], " ")[0], 32)
	if err != nil {
		return 0, fmt.Errorf("ParseProtonVersion5: Unable to parse Proton version from path: '%v'", path)
	}

	return f, nil
}

// Returns all Proton installations found sorted from best to worst
func GetProtons() ([]string, error) {
	var Protons1 []string
	var Protons2 []string
	var Protons3 []string
	var Protons4 []string
	var Protons5 []string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("GetProtons: Unable to get home directory")
	}

	var protonPaths = []string{
		filepath.Join(homeDir, ".steam", "root", "steamapps", "common"),
		filepath.Join(homeDir, ".local", "share", "Steam", "steamapps", "common"),
		filepath.Join(homeDir, ".local", "share", "Steam", "compatibilitytools.d"),
	}

	for _, protonPath := range protonPaths {
		protons1, err := filepath.Glob(filepath.Join(protonPath, "Proton *.*"))
		if err != nil {
			return nil, fmt.Errorf("GetProtons: Invalid glob syntax: %v", err)
		}

		for _, proton1 := range protons1 {
			// Check for 'proton' file inside directory
			if stat, err := os.Stat(filepath.Join(proton1, "proton")); err == nil {
				_, err := ParseProtonVersion1(proton1)
				if err == nil && !stat.IsDir() {
					Protons1 = append(Protons1, proton1)
				}
			} else if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				return nil, fmt.Errorf("GetProtons: Unable to inspect path: '%v'", filepath.Join(proton1, "proton"))
			}
		}

		protons2, err := filepath.Glob(filepath.Join(protonPath, "Proton - Experimental"))
		if err != nil {
			return nil, fmt.Errorf("GetProtons: Invalid glob syntax: %v", err)
		}

		for _, proton2 := range protons2 {
			// Check for 'proton' file inside directory
			if stat, err := os.Stat(filepath.Join(proton2, "proton")); err == nil {
				if !stat.IsDir() {
					Protons2 = append(Protons2, proton2)
				}
			} else if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				return nil, fmt.Errorf("GetProtons: Unable to inspect path: '%v'", filepath.Join(proton2, "proton"))
			}
		}

		protons3, err := filepath.Glob(filepath.Join(protonPath, "GE-Proton*-*"))
		if err != nil {
			return nil, fmt.Errorf("GetProtons: Invalid glob syntax: %v", err)
		}

		for _, proton3 := range protons3 {
			// Check for 'proton' file inside directory
			if stat, err := os.Stat(filepath.Join(proton3, "proton")); err == nil {
				_, err := ParseProtonVersion3(proton3)
				if err == nil && !stat.IsDir() {
					Protons3 = append(Protons3, proton3)
				}
			} else if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				return nil, fmt.Errorf("GetProtons: Unable to inspect path: '%v'", filepath.Join(proton3, "proton"))
			}
		}

		protons4, err := filepath.Glob(filepath.Join(protonPath, "Proton Hotfix"))
		if err != nil {
			return nil, fmt.Errorf("GetProtons: Invalid glob syntax: %v", err)
		}

		for _, proton4 := range protons4 {
			// Check for 'proton' file inside directory
			if stat, err := os.Stat(filepath.Join(proton4, "proton")); err == nil {
				if !stat.IsDir() {
					Protons4 = append(Protons4, proton4)
				}
			} else if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				return nil, fmt.Errorf("GetProtons: Unable to inspect path: '%v'", filepath.Join(proton4, "proton"))
			}
		}

		protons5, err := filepath.Glob(filepath.Join(protonPath, "Proton *.* (Beta)"))
		if err != nil {
			return nil, fmt.Errorf("GetProtons: Invalid glob syntax: %v", err)
		}

		for _, proton5 := range protons5 {
			// Check for 'proton' file inside directory
			if stat, err := os.Stat(filepath.Join(proton5, "proton")); err == nil {
				_, err := ParseProtonVersion5(proton5)
				if err == nil && !stat.IsDir() {
					Protons5 = append(Protons5, proton5)
				}
			} else if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				return nil, fmt.Errorf("GetProtons: Unable to inspect path: '%v'", filepath.Join(proton5, "proton"))
			}
		}
	}

	// Sort by versions
	sort.Slice(Protons1, func(i int, j int) bool {
		// We can ignore these errors because they were checked in the for loop
		v1, _ := ParseProtonVersion1(Protons1[i])
		v2, _ := ParseProtonVersion1(Protons1[j])
		return v1 < v2
	})

	sort.Slice(Protons3, func(i int, j int) bool {
		// We can ignore these errors because they were checked in the for loop
		v1, _ := ParseProtonVersion3(Protons3[i])
		v2, _ := ParseProtonVersion3(Protons3[j])
		return v1 < v2
	})

	sort.Slice(Protons5, func(i int, j int) bool {
		// We can ignore these errors because they were checked in the for loop
		v1, _ := ParseProtonVersion5(Protons5[i])
		v2, _ := ParseProtonVersion5(Protons5[j])
		return v1 < v2
	})

	return append(Protons1, append(Protons2, append(Protons3, append(Protons4, Protons5...)...)...)...), nil
}

// Returns the best Proton's executable
func GetProton() (string, error) {
	protons, err := GetProtons()
	if err != nil {
		return "", err
	}

	if len(protons) == 0 {
		return "", fmt.Errorf("GetProton: No Proton installation was found")
	}

	return filepath.Join(protons[0], "proton"), nil
}

// Returns the command for running a specified game with Proton
func ProtonRun(gameExe string, proton string, gameMeta cstructs.GameMeta) (*exec.Cmd, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("ProtonRun: Unable to get home directory")
	}

	cmd := exec.Command(proton, "run", gameExe)

	gameId := gameMeta.GetId()

	err = os.MkdirAll(filepath.Join(homeDir, ".local", "share", "marauder", "compat", gameId), 0755)
	if err != nil {
		return nil, fmt.Errorf("ProtonRun: Unable to create directory: '%v'", filepath.Join(homeDir, ".local", "share", "marauder", "compat", gameId))
	}

	err = os.MkdirAll(filepath.Join(homeDir, ".local", "share", "marauder", "logs"), 0755)
	if err != nil {
		return nil, fmt.Errorf("ProtonRun: Unable to create directory: '%v'", filepath.Join(homeDir, ".local", "share", "marauder", "logs"))
	}

	vars, err := script.ParseLocalScriptVariables(gameMeta)
	if err != nil {
		return nil, err
	}

	env := os.Environ()

	if !IsEnvSet(env, "STEAM_COMPAT_DATA_PATH") {
		env = append(env, "STEAM_COMPAT_DATA_PATH=" + filepath.Join(homeDir, ".local", "share", "marauder", "compat", gameId))
	}

	if !IsEnvSet(env, "STEAM_COMPAT_CLIENT_INSTALL_PATH") {
		env = append(env, "STEAM_COMPAT_CLIENT_INSTALL_PATH=" + filepath.Join(homeDir, ".local", "share", "Steam"))
	}

	if !IsEnvSet(env, "PROTON_LOG") {
		env = append(env, "PROTON_LOG=1")
	}

	if !IsEnvSet(env, "PROTON_LOG_DIR") {
		env = append(env, "PROTON_LOG_DIR=" + filepath.Join(homeDir, ".local", "share", "marauder", "logs"))
	}

	if !IsEnvSet(env, "STEAM_COMPAT_APP_ID") {
		steamAppId, ok := vars["steam-app-id"]
		if ok {
			env = append(env, "STEAM_COMPAT_APP_ID=" + steamAppId)
		}
	}

	cmd.Env = env
	cmd.Dir = filepath.Dir(gameExe)

	return cmd, nil
}

func IsEnvSet(env []string, key string) bool {
	for _, envVar := range env {
		if strings.HasPrefix(envVar, key + "=") {
			return true
		}
	}

	return false
}
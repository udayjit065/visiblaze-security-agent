package collect

import (
	"strings"

	"github.com/visiblaze/sec-agent/agent/internal/util"
)

type Package struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Arch        string `json:"arch"`
	Manager     string `json:"manager"`
	Source      string `json:"source"`
	InstalledAt string `json:"installed_at"`
}

func CollectPackages(osID string) ([]Package, error) {
	var pkgs []Package

	switch osID {
	case "ubuntu", "debian":
		p, err := collectDpkg()
		if err == nil {
			pkgs = append(pkgs, p...)
		}
	case "rhel", "centos", "fedora":
		p, err := collectRPM()
		if err == nil {
			pkgs = append(pkgs, p...)
		}
	case "alpine":
		p, err := collectAPK()
		if err == nil {
			pkgs = append(pkgs, p...)
		}
	default:
		p, _ := collectDpkg()
		pkgs = append(pkgs, p...)
		p, _ = collectRPM()
		pkgs = append(pkgs, p...)
		p, _ = collectAPK()
		pkgs = append(pkgs, p...)
	}

	return pkgs, nil
}

func collectDpkg() ([]Package, error) {
	output, err := util.RunCmd("dpkg-query", "-W", "-f=${Package}\t${Version}\t${Architecture}\n")
	if err != nil {
		return nil, err
	}

	var pkgs []Package
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			pkgs = append(pkgs, Package{
				Name:    parts[0],
				Version: parts[1],
				Arch:    parts[2],
				Manager: "dpkg",
				Source:  "debian",
			})
		}
	}

	return pkgs, nil
}

func collectRPM() ([]Package, error) {
	output, err := util.RunCmd("rpm", "-qa", "--qf", "%{NAME}\t%{VERSION}-%{RELEASE}\t%{ARCH}\n")
	if err != nil {
		return nil, err
	}

	var pkgs []Package
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			pkgs = append(pkgs, Package{
				Name:    parts[0],
				Version: parts[1],
				Arch:    parts[2],
				Manager: "rpm",
				Source:  "rhel",
			})
		}
	}

	return pkgs, nil
}

func collectAPK() ([]Package, error) {
	output, err := util.RunCmd("apk", "info", "-v")
	if err != nil {
		return nil, err
	}

	var pkgs []Package
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.LastIndex(line, "-")
		if parts > 0 {
			pkgs = append(pkgs, Package{
				Name:    line[:parts],
				Version: line[parts+1:],
				Arch:    "unknown",
				Manager: "apk",
				Source:  "alpine",
			})
		}
	}

	return pkgs, nil
}

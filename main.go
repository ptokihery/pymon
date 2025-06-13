package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ptokihery/pymon/config"
	"github.com/ptokihery/pymon/notify"
)



func main() {

    configPath := flag.String("c", "/etc/pymon/pymon.conf", "Path to config file")
    flag.Parse()

    cfg := config.Load(*configPath)

	if !shouldNotify(cfg.Server.Interval) {
        return 
    }

    if cfg.Disk.Enabled {
        checkDisk(cfg)
    }
	fmt.Printf("Gp")
    if cfg.Services.Enabled {
        checkServices(cfg)
    }
    if cfg.Resources.Enabled {
        checkResources(cfg)
    }
	if cfg.HTTPServer.Enabled {
		checkHTTPPorts(cfg)
	}

	if cfg.CleanUp.Enabled {
	fmt.Printf("CLEAN UPDA")

		cleanUp(cfg)
	}

	updateNotifyState()
}

func cleanUp(cfg config.Config) {
	for _, file := range cfg.CleanUp.Files {
		path := file.Path
		sizeStr := file.Size // e.g., "10Mb", "2Go"

		maxSize, err := parseSize(sizeStr)
		if err != nil {
			fmt.Printf("Erreur lors de l'analyse de la taille pour %s: %v\n", path, err)
			continue
		}

		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue 
			}
			fmt.Printf("Erreur lors de la vérification du fichier %s: %v\n", path, err)
			continue
		}

		if info.Size() > maxSize {
			err := os.Remove(path)
			if err != nil {
				fmt.Printf("Erreur lors de la suppression du fichier %s: %v\n", path, err)
				continue
			}
			// Recreate empty file
			_, err = os.Create(path)
			if err != nil {
				fmt.Printf("Erreur lors de la création du fichier %s: %v\n", path, err)
			}
		}
	}
}

func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ToLower(sizeStr))
	multiplier := int64(1)

	switch {
	case strings.HasSuffix(sizeStr, "kb"):
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "kb")
	case strings.HasSuffix(sizeStr, "mb"):
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "mb")
	case strings.HasSuffix(sizeStr, "gb"):
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "gb")
	case strings.HasSuffix(sizeStr, "go"):
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "go")
	case strings.HasSuffix(sizeStr, "b"):
		multiplier = 1
		sizeStr = strings.TrimSuffix(sizeStr, "b")
	}

	val, err := strconv.ParseFloat(strings.TrimSpace(sizeStr), 64)
	if err != nil {
		return 0, err
	}
	return int64(val * float64(multiplier)), nil
}


func getStateDir() string {
    home, err := os.UserHomeDir()
    if err != nil {
        return "./pymon"
    }
    return filepath.Join(home, ".pymon")
}

func shouldNotify(cooldownMinutes int) bool {
	path := getStateDir()

	if err := os.MkdirAll(path, 0755); err != nil {
		fmt.Println("Erreur lors de la création du dossier:", err)
		return false
	}

	stateFile := path + "/notify_state"
	info, err := os.Stat(stateFile)
	if os.IsNotExist(err) {
		_ = os.WriteFile(stateFile, []byte(time.Now().Format(time.RFC3339)), 0644)
		return false
	}
	if err != nil {
		fmt.Println("Erreur lors de la vérification de l'état de notification:", err)
		return false
	}

	elapsed := time.Since(info.ModTime())
	return elapsed.Minutes() >= float64(cooldownMinutes)
}

func updateNotifyState() {
	path := getStateDir()
    os.WriteFile(path +"/notify_state", []byte(time.Now().Format(time.RFC3339)), 0644)
}


func checkDisk(cfg config.Config) {
    output, err := exec.Command("df", "-h", "--output=pcent,target").Output()
    if err != nil {
        notify.SendMail(cfg, "Erreur vérif disque", "Impossible de vérifier le disque", err.Error())
        return
    }

    lines := strings.Split(string(output), "\n")[1:]
    for _, line := range lines {
        fields := strings.Fields(line)
        if len(fields) < 2 {
            continue
        }

        mountPoint := fields[1]
        if shouldExcludeMount(mountPoint, cfg.Disk.Exclude) {
            continue
        }

        usageStr := strings.TrimSuffix(fields[0], "%")
        usage, err := strconv.Atoi(usageStr)
        if err != nil {
            continue
        }

        fmt.Println("Disk usage for", mountPoint, "is", usage, "%")
        if usage >= cfg.Disk.Threshold {
            msg := fmt.Sprintf("Le disque %s est utilisé à %d%%", mountPoint, usage)
            notify.SendMail(cfg, "Alerte espace disque", msg, string(output))
        }
    }
}

func shouldExcludeMount(mount string, excluded []string) bool {
    for _, ex := range excluded {
        if strings.HasPrefix(mount, strings.TrimSpace(ex)) {
            return true
        }
    }
    return false
}


func checkServices(cfg config.Config) {
    for _, service := range cfg.Services.List {
        cmd := exec.Command("systemctl", "is-active", service)
        output, err := cmd.CombinedOutput()
        if err != nil || strings.TrimSpace(string(output)) != "active" {
            logMsg := fmt.Sprintf("Service %s inactif: %s", service, string(output))
            notify.SendMail(cfg, fmt.Sprintf("Service %s KO", service), "Service arrêté ou en échec", logMsg)
        }
    }
}

func checkHTTPPorts(cfg config.Config) {
    for _, srv := range cfg.HTTPServer.Servers {
        address := fmt.Sprintf("127.0.0.1:%d", srv.Port)
        conn, err := net.DialTimeout("tcp", address, 3*time.Second)
        if err != nil {
            msg := fmt.Sprintf("Le service %s (port %d) est inaccessible", srv.Name, srv.Port)
            notify.SendMail(cfg, fmt.Sprintf("Alerte %s : KO", srv.Name), msg, err.Error())
        } else {
            conn.Close()
        }
    }
}

func checkResources(cfg config.Config) {
    cpuOut, _ := exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)' | awk '{print 100 - $8}'").Output()
    cpuStr := strings.TrimSpace(string(cpuOut))
    cpu, _ := strconv.ParseFloat(cpuStr, 64)

    memOut, _ := exec.Command("sh", "-c", "free | grep Mem | awk '{print ($3/$2)*100}'").Output()
    memStr := strings.TrimSpace(string(memOut))
    mem, _ := strconv.ParseFloat(memStr, 64)
	fmt.Printf("CPU Usage: %.2f%%, RAM Usage: %.2f%%\n", cpu, mem)
    if int(cpu) >= cfg.Resources.CPUThreshold {
        notify.SendMail(cfg, "Alerte CPU", fmt.Sprintf("Utilisation CPU à %.2f%%", cpu), string(cpuOut))
    }
    if int(mem) >= cfg.Resources.RAMThreshold {
        notify.SendMail(cfg, "Alerte RAM", fmt.Sprintf("Utilisation RAM à %.2f%%", mem), string(memOut))
    }
}


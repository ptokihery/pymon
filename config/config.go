package config

import (
	"fmt"
	"log"
	"strings"

	"gopkg.in/ini.v1"
)

type Config struct {
	Server struct {
		Name string
		Interval int // Interval in minutes
	}
    Disk struct {
		Enabled   bool
		Threshold int
		Exclude   []string
	}

    Services struct {
        Enabled bool
        List    []string
    }
    Resources struct {
        Enabled      bool
        CPUThreshold int
        RAMThreshold int
    }
	HTTPServer struct {
        Enabled bool
        Servers []NamedPort
    }
	CleanUp struct {
        Enabled bool
		Files []File
	}
	
    Email EmailConfig
}
type File struct {
		Name string
		Size string
		Path string
}

type NamedPort struct {
    Name string
    Port int
}

type EmailConfig struct {
    Enabled    bool
    To         []string
    SMTPServer string
    SMTPPort   int
    Username   string
    Password   string
	Server string
}

func Load(path string) Config {
    cfgFile, err := ini.Load(path)
    if err != nil {
        log.Fatalf("Impossible de charger le fichier de config: %v", err)
    }

    var cfg Config
	

    cfg.Disk.Enabled = cfgFile.Section("disk").Key("enabled").MustBool()
    cfg.Disk.Threshold = cfgFile.Section("disk").Key("threshold").MustInt(80)

    cfg.Services.Enabled = cfgFile.Section("services").Key("enabled").MustBool()
    cfg.Services.List = strings.Split(cfgFile.Section("services").Key("list").MustString(""), ",")

    cfg.Resources.Enabled = cfgFile.Section("resources").Key("enabled").MustBool()
    cfg.Resources.CPUThreshold = cfgFile.Section("resources").Key("cpu_threshold").MustInt(95)
    cfg.Resources.RAMThreshold = cfgFile.Section("resources").Key("ram_threshold").MustInt(95)

    cfg.Email.Enabled = cfgFile.Section("email").Key("enabled").MustBool()
    cfg.Email.To = strings.Split(cfgFile.Section("email").Key("to").MustString(""), ",")
    cfg.Email.SMTPServer = cfgFile.Section("email").Key("smtp_server").String()
    cfg.Email.SMTPPort = cfgFile.Section("email").Key("smtp_port").MustInt(587)
    cfg.Email.Username = cfgFile.Section("email").Key("username").String()
    cfg.Email.Password = cfgFile.Section("email").Key("password").String()
	cfg.Disk.Exclude = strings.Split(cfgFile.Section("disk").Key("exclude_mounts").MustString(""), ",")
	cfg.Server.Name = cfgFile.Section("server").Key("name").String()
	cfg.Server.Interval = cfgFile.Section("server").Key("interval").MustInt(60) // Default to 60 minutes
	cfg.HTTPServer.Enabled = cfgFile.Section("http.server").Key("enabled").MustBool()

	// CleanUp

	cfg.CleanUp.Enabled = cfgFile.Section("cleanup").Key("enabled").MustBool()
	for i := 0; ; i++ {
		sectionName := fmt.Sprintf("cleanup.%d", i)
		if !cfgFile.HasSection(sectionName) {
			break
		}
		section := cfgFile.Section(sectionName)
		name := section.Key("name").String()
		size := section.Key("size").String()
		path := section.Key("path").String()
		cfg.CleanUp.Files = append(cfg.CleanUp.Files, File{Name: name, Size: size, Path: path})
	}

	for i := 0; ; i++ {
		sectionName := fmt.Sprintf("http.server.%d", i)
		if !cfgFile.HasSection(sectionName) {
			break
		}
		section := cfgFile.Section(sectionName)
		name := section.Key("name").String()
		port := section.Key("port").MustInt()
		cfg.HTTPServer.Servers = append(cfg.HTTPServer.Servers, NamedPort{Name: name, Port: port})
	}


    for i := range cfg.Email.To {
        cfg.Email.To[i] = strings.TrimSpace(cfg.Email.To[i])
    }

    return cfg
}


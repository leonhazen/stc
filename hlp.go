package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tenox7/stc/api"
)

func usage() {
	o := flag.CommandLine.Output()
	fmt.Fprintf(o, "stc [flags] [commands]\n\nflags:\n")
	flag.PrintDefaults()
	fmt.Fprintf(o, "\ncommands:\n"+
		"  log           - print syncthing 'recent' log\n"+
		"  restart       - restart syncthing daemon\n"+
		"  shutdown      - shutdown syncthing daemon\n"+
		"  errors        - print errors visible in web UI\n"+
		"  clear_errors  - clear errors in the web UI\n"+
		"  post_error    - posts a custom error message in the web UI\n"+
		"  folder_errors - prints folder errors from scan or pull\n"+
		"  folder_pause  - pause specified folder\n"+
		"  folder_resume - unpause specified folder\n"+
		"  id            - print ID of this node\n"+
		"  reset_db      - reset the database / file index\n"+
		"  rescan        - rescan a folder or 'all'\n"+
		"  override      - override remote changed for a send-only folder (OoSync)\n"+
		"  revert        - revert local changes for a receive-only folder (LocAdds)\n",
	)
}

func cfg(apiKey, target, configfile string) (string, string, error) {
	if apiKey == "" {
		apiKey = os.Getenv("APIKEY")
	}
	if apiKey != "" && target != "" {
		return apiKey, target, nil
	}

	if configfile == "" {
		cfgDir, err := os.UserConfigDir()
		if err != nil {
			return "", "", err
		}

		configfile = filepath.Join(cfgDir, "syncthing", "config.xml")
	}

	var err error
	var f []byte
	f, err = os.ReadFile(configfile)
	if err != nil {
		return "", "", err
	}

	x := struct {
		XMLName xml.Name `xml:"configuration"`
		GUI     struct {
			ApiKey  string `xml:"apikey,omitempty"`
			Address string `xml:"address" default:"127.0.0.1:8384"`
			UseTLS  bool   `xml:"tls,attr"`
		} `xml:"gui"`
	}{}

	err = xml.Unmarshal(f, &x)
	if err != nil {
		return "", "", err
	}

	p := "http://"
	if x.GUI.UseTLS {
		p = "https://"
	}

	return x.GUI.ApiKey, p + x.GUI.Address, nil
}

func isConn(paused, conn bool, ID, myID string) string {
	if ID == myID {
		return "Myself"
	}
	if paused {
		return "Paused"
	}
	if conn {
		return "OK"
	}
	return "Offline"
}

func fStatus(paused bool, ty, st string, err, loChg, needItms uint64) string {
	if paused {
		return "Paused"
	}
	if err > 0 {
		return "Errors"
	}
	if ty == "sendonly" && needItms > 0 {
		return "OoSync"
	}
	if ty == "receiveonly" && loChg > 0 {
		return "LocAdds"
	}
	return st
}

func folderID(fName string) (string, error) {
	cfg, err := api.GetConfig()
	if err != nil {
		return "", err
	}
	fID := ""
	for _, f := range cfg.Folders {
		if f.Label != fName {
			continue
		}
		fID = f.ID
	}
	return fID, nil
}

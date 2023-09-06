package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path"
	"runtime"
	"strconv"
	"time"

	"github.com/gochan-org/gochan/pkg/gcutil"
)

const (
	GC_DIR_MODE  fs.FileMode = 0775
	GC_FILE_MODE fs.FileMode = 0664
)

var (
	uid int
	gid int
)

// MissingField represents a field missing from the configuration file
type MissingField struct {
	Name        string
	Critical    bool
	Description string
}

// InvalidValueError represents a GochanConfig field with a bad value
type InvalidValueError struct {
	Field   string
	Value   any
	Details string
}

func (iv *InvalidValueError) Error() string {
	str := fmt.Sprintf("invalid %s value: %#v", iv.Field, iv.Value)
	if iv.Details != "" {
		str += " - " + iv.Details
	}
	return str
}

func TakeOwnership(fp string) (err error) {
	if runtime.GOOS == "windows" || fp == "" || cfg.Username == "" {
		// Chown returns an error in Windows so skip it, also skip if Username isn't set
		// because otherwise it'll think we want to switch to uid and gid 0 (root)
		return nil
	}
	return os.Chown(fp, uid, gid)
}

func TakeOwnershipOfFile(f *os.File) error {
	if runtime.GOOS == "windows" || f == nil || cfg.Username == "" {
		// Chown returns an error in Windows so skip it, also skip if Username isn't set
		// because otherwise it'll think we want to switch to uid and gid 0 (root)
		return nil
	}
	return f.Chown(uid, gid)
}

// InitConfig loads and parses gochan.json on startup and verifies its contents
func InitConfig(versionStr string) {
	cfg = defaultGochanConfig
	if flag.Lookup("test.v") != nil {
		// create a dummy config for testing if we're using go test
		cfg = defaultGochanConfig
		cfg.ListenIP = "127.0.0.1"
		cfg.Port = 8080
		cfg.UseFastCGI = true
		cfg.DebugMode = true
		cfg.testing = true
		cfg.TemplateDir = "templates"
		cfg.DBtype = "sqlite3"
		cfg.DBhost = "./testdata/gochantest.db"
		cfg.DBname = "gochan"
		cfg.DBusername = "gochan"
		cfg.SiteDomain = "127.0.0.1"
		cfg.RandomSeed = "test"
		cfg.Version = ParseVersion(versionStr)
		cfg.SiteSlogan = "Gochan testing"
		cfg.Verbosity = 1
		cfg.Captcha.OnlyNeededForThreads = true
		cfg.Cooldowns = BoardCooldowns{0, 0, 0}
		cfg.BanColors = []string{
			"admin:#0000A0",
			"somemod:blue",
		}
		return
	}
	cfgPath = gcutil.FindResource(
		"gochan.json",
		"/usr/local/etc/gochan/gochan.json",
		"/etc/gochan/gochan.json")
	if cfgPath == "" {
		fmt.Println("gochan.json not found")
		os.Exit(1)
	}

	cfgFile, err := os.Open(cfgPath)
	if err != nil {
		fmt.Printf("Error reading %s: %s\n", cfgPath, err.Error())
		os.Exit(1)
	}
	defer cfgFile.Close()

	if err = json.NewDecoder(cfgFile).Decode(cfg); err != nil {
		fmt.Printf("Error parsing %s: %s", cfgPath, err.Error())
		os.Exit(1)
	}
	cfg.jsonLocation = cfgPath

	if err = cfg.ValidateValues(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if runtime.GOOS != "windows" {
		var gcUser *user.User
		if cfg.Username != "" {
			gcUser, err = user.Lookup(cfg.Username)
		} else {
			gcUser, err = user.Current()
		}
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if uid, err = strconv.Atoi(gcUser.Uid); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if gid, err = strconv.Atoi(gcUser.Gid); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	if _, err = os.Stat(cfg.DocumentRoot); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if _, err = os.Stat(cfg.TemplateDir); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if _, err = os.Stat(cfg.LogDir); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	cfg.LogDir = gcutil.FindResource(cfg.LogDir, "log", "/var/log/gochan/")
	if err = gcutil.InitLogs(cfg.LogDir, cfg.DebugMode, uid, gid); err != nil {
		fmt.Println("Error opening logs:", err.Error())
		os.Exit(1)
	}

	if cfg.Port == 0 {
		cfg.Port = 80
	}

	if len(cfg.FirstPage) == 0 {
		cfg.FirstPage = []string{"index.html", "1.html", "firstrun.html"}
	}

	if cfg.WebRoot == "" {
		cfg.WebRoot = "/"
	}

	if cfg.WebRoot[0] != '/' {
		cfg.WebRoot = "/" + cfg.WebRoot
	}
	if cfg.WebRoot[len(cfg.WebRoot)-1] != '/' {
		cfg.WebRoot += "/"
	}

	if cfg.EnableGeoIP {
		if _, err = os.Stat(cfg.GeoIPDBlocation); err != nil {
			gcutil.LogError(err).
				Str("location", cfg.GeoIPDBlocation).
				Msg("Unable to load GeoIP file location set in gochan.json, disabling GeoIP")
		}
		cfg.EnableGeoIP = false
	}

	_, zoneOffset := time.Now().Zone()
	cfg.TimeZone = zoneOffset / 60 / 60

	cfg.Version = ParseVersion(versionStr)
	cfg.Version.Normalize()
}

// WebPath returns an absolute path, starting at the web root (which is "/" by default)
func WebPath(part ...string) string {
	return path.Join(cfg.WebRoot, path.Join(part...))
}

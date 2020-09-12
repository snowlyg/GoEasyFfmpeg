package utils

import (
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/teris-io/shortid"

	"github.com/go-ini/ini"
)

func LocalIP() string {
	ip := ""
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsMulticast() && !ipnet.IP.IsLinkLocalUnicast() && !ipnet.IP.IsLinkLocalMulticast() && ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
			}
		}
	}
	return ip
}

func MD5(str string) string {
	encoder := md5.New()
	encoder.Write([]byte(str))
	return hex.EncodeToString(encoder.Sum(nil))
}

func CWD() string {
	path, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(path)
}

func EXEName() string {
	path, err := os.Executable()
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func HomeDir() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.HomeDir
}

func LogDir() string {
	dir := filepath.Join(CWD(), "logs")
	EnsureDir(dir)
	return dir
}

func ErrorLogFilename() string {
	return filepath.Join(LogDir(), fmt.Sprintf("%s-error.log", strings.ToLower(EXEName())))
}

func DataDir() string {
	dir := CWD()
	_dir := Conf().Section("").Key("data_dir").Value()
	if _dir != "" {
		dir = _dir
	}
	dir = ExpandHomeDir(dir)
	EnsureDir(dir)
	return dir
}

var FlagVarConfFile string

func ConfFile() string {
	if FlagVarConfFile != "" {
		return FlagVarConfFile
	}
	if Exist(ConfFileDev()) {
		return ConfFileDev()
	}
	return filepath.Join(CWD(), strings.ToLower(EXEName())+".ini")
}

func ConfFileDev() string {
	return filepath.Join(CWD(), strings.ToLower(EXEName())+".dev.ini")
}

var FlagVarDBFile string

func DBFile() string {
	if FlagVarDBFile != "" {
		return FlagVarDBFile
	}
	if Exist(DBFileDev()) {
		return DBFileDev()
	}
	return filepath.Join(CWD(), strings.ToLower(EXEName()+".db"))
}

func DBFileDev() string {
	return filepath.Join(CWD(), strings.ToLower(EXEName())+".dev.db")
}

var conf *ini.File

func Conf() *ini.File {
	if conf != nil {
		return conf
	}
	if _conf, err := ini.InsensitiveLoad(ConfFile()); err != nil {
		_conf, _ = ini.LoadSources(ini.LoadOptions{Insensitive: true}, []byte(""))
		conf = _conf
	} else {
		conf = _conf
	}
	return conf
}

func ReloadConf() *ini.File {
	if _conf, err := ini.InsensitiveLoad(ConfFile()); err != nil {
		_conf, _ = ini.LoadSources(ini.LoadOptions{Insensitive: true}, []byte(""))
		conf = _conf
	} else {
		conf = _conf
	}
	return conf
}

func ExpandHomeDir(path string) string {
	if len(path) == 0 {
		return path
	}
	if path[0] != '~' {
		return path
	}
	if len(path) > 1 && path[1] != '/' && path[1] != '\\' {
		return path
	}
	return filepath.Join(HomeDir(), path[1:])
}

func EnsureDir(dir string) (err error) {
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return
		}
	}
	return
}

func Exist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func ShortID() string {
	return shortid.MustGenerate()
}

func PauseExit() {
	log.Println("Press any to exit")
	keyboard.GetSingleKey()
	os.Exit(0)
}

func IsPortInUse(port int) bool {
	if conn, err := net.DialTimeout("tcp", net.JoinHostPort("", fmt.Sprintf("%d", port)), 3*time.Second); err == nil {
		conn.Close()
		return true
	}
	return false
}

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register(StringArray(""))
	ini.PrettyFormat = false
}

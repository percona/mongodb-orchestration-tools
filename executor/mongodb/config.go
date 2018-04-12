package mongodb

const (
	DefaultBinDir            = "/usr/bin"
	DefaultTmpDirFallback    = "/tmp"
	DefaultConfigDirFallback = "/etc"
	DefaultUser              = "mongodb"
	DefaultGroup             = "root"
)

type Config struct {
	ConfigDir string
	BinDir    string
	TmpDir    string
	User      string
	Group     string
}

package main

import (
	"context"
	"flag"
	"fmt"
	mfs "github.com/Hacky-DH/moosefs-client"
	"github.com/golang/glog"
	"github.com/google/subcommands"
	"os"
	"path/filepath"
)

var (
	version    bool
	versionstr string
)

func init() {
	flag.BoolVar(&version, "version", false, "show version")
	flag.Set("logtostderr", "true")
	flag.Set("v", "1")
}

type uploadCmd struct {
	dst string
}

func (*uploadCmd) Name() string     { return "put" }
func (*uploadCmd) Synopsis() string { return "upload files to mfs" }
func (s *uploadCmd) Usage() string {
	return fmt.Sprintf("%s [-d path] <src1> <src2> ...\n\t%s\n", s.Name(), s.Synopsis())
}
func (s *uploadCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&s.dst, "d", "", "upload dst path")
}
func (s *uploadCmd) Execute(_ context.Context, f *flag.FlagSet,
	_ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitUsageError
	}
	c, err := mfs.NewCLient()
	if err != nil {
		glog.Error(err)
		return subcommands.ExitFailure
	}
	defer c.Close()
	for _, src := range f.Args() {
		dst := filepath.Join(s.dst, filepath.Base(src))
		glog.Infof("start upload file %s to %s", src, dst)
		err = c.WriteFile(src, dst)
		if err != nil {
			glog.Error(err)
			return subcommands.ExitFailure
		}
	}
	return subcommands.ExitSuccess
}

type downloadCmd struct {
	src string
}

func (*downloadCmd) Name() string     { return "get" }
func (*downloadCmd) Synopsis() string { return "download file from mfs" }
func (s *downloadCmd) Usage() string {
	return fmt.Sprintf("%s -s <path> [src path]\n\t%s\n", s.Name(), s.Synopsis())
}
func (s *downloadCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&s.src, "s", "", "download src file")
}
func (s *downloadCmd) Execute(_ context.Context, f *flag.FlagSet,
	_ ...interface{}) subcommands.ExitStatus {
	if len(s.src) < 1 {
		f.Usage()
		return subcommands.ExitUsageError
	}
	c, err := mfs.NewCLient()
	if err != nil {
		glog.Error(err)
		return subcommands.ExitFailure
	}
	defer c.Close()
	dst := filepath.Base(s.src)
	if f.NArg() > 0 {
		dst = filepath.Join(f.Arg(0), filepath.Base(s.src))
	}
	glog.Infof("start download file %s to %s", s.src, dst)
	err = c.ReadFile(s.src, dst)
	if err != nil {
		glog.Error(err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

type lsCmd struct {
}

func (*lsCmd) Name() string     { return "ls" }
func (*lsCmd) Synopsis() string { return "list files and dirs" }
func (s *lsCmd) Usage() string {
	return fmt.Sprintf("%s [path]\n\t%s\n", s.Name(), s.Synopsis())
}
func (s *lsCmd) SetFlags(f *flag.FlagSet) {
}
func (s *lsCmd) Execute(_ context.Context, f *flag.FlagSet,
	_ ...interface{}) subcommands.ExitStatus {
	c, err := mfs.NewCLient()
	if err != nil {
		glog.Error(err)
		return subcommands.ExitFailure
	}
	defer c.Close()
	path := c.Cwd
	if f.NArg() > 0 {
		path = f.Arg(0)
	}
	info, err := c.Readdir(path)
	if err != nil {
		glog.Error(err)
		return subcommands.ExitFailure
	}
	if len(info) == 0 {
		glog.Info("empty dir")
	} else {
		var i int
		for _, f := range info {
			if f.Name == ".." || f.Name == "." {
				continue
			}
			fmt.Printf("%s\t", f.Name)
			i++
			if i%5 == 0 {
				fmt.Println()
			}
		}
		if i%5 != 0 {
			fmt.Println()
		}
	}
	return subcommands.ExitSuccess
}

type removeCmd struct {
}

func (*removeCmd) Name() string     { return "rm" }
func (*removeCmd) Synopsis() string { return "remove files" }
func (s *removeCmd) Usage() string {
	return fmt.Sprintf("%s <path>\n\t%s\n", s.Name(), s.Synopsis())
}
func (s *removeCmd) SetFlags(f *flag.FlagSet) {
}
func (s *removeCmd) Execute(_ context.Context, f *flag.FlagSet,
	_ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitUsageError
	}
	c, err := mfs.NewCLient()
	if err != nil {
		glog.Error(err)
		return subcommands.ExitFailure
	}
	defer c.Close()
	for _, path := range f.Args() {
		err = c.Unlink(path)
		if err != nil {
			glog.Error(err)
			return subcommands.ExitFailure
		}
	}
	return subcommands.ExitSuccess
}

type mkdirCmd struct {
}

func (*mkdirCmd) Name() string     { return "mkdir" }
func (*mkdirCmd) Synopsis() string { return "make dirs" }
func (s *mkdirCmd) Usage() string {
	return fmt.Sprintf("%s <path>\n\t%s\n", s.Name(), s.Synopsis())
}
func (s *mkdirCmd) SetFlags(f *flag.FlagSet) {
}
func (s *mkdirCmd) Execute(_ context.Context, f *flag.FlagSet,
	_ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitUsageError
	}
	c, err := mfs.NewCLient()
	if err != nil {
		glog.Error(err)
		return subcommands.ExitFailure
	}
	defer c.Close()
	for _, path := range f.Args() {
		err = c.Mkdir(path)
		if err != nil {
			glog.Error(err)
			return subcommands.ExitFailure
		}
	}
	return subcommands.ExitSuccess
}

type rmdirCmd struct {
}

func (*rmdirCmd) Name() string     { return "rmdir" }
func (*rmdirCmd) Synopsis() string { return "remove dirs" }
func (s *rmdirCmd) Usage() string {
	return fmt.Sprintf("%s <path>\n\t%s\n", s.Name(), s.Synopsis())
}
func (s *rmdirCmd) SetFlags(f *flag.FlagSet) {
}
func (s *rmdirCmd) Execute(_ context.Context, f *flag.FlagSet,
	_ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitUsageError
	}
	c, err := mfs.NewCLient()
	if err != nil {
		glog.Error(err)
		return subcommands.ExitFailure
	}
	defer c.Close()
	for _, path := range f.Args() {
		err = c.Rmdir(path)
		if err != nil {
			glog.Error(err)
			return subcommands.ExitFailure
		}
	}
	return subcommands.ExitSuccess
}

func mainRecover() {
	if err := recover(); err != nil {
		glog.Fatal("Error: ", err)
	}
}

func main() {
	defer mainRecover()
	defer glog.Flush()
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(&uploadCmd{}, "mfs")
	subcommands.Register(&downloadCmd{}, "mfs")
	subcommands.Register(&lsCmd{}, "mfs")
	subcommands.Register(&removeCmd{}, "mfs")
	subcommands.Register(&mkdirCmd{}, "mfs")
	subcommands.Register(&rmdirCmd{}, "mfs")

	flag.Parse()
	if version {
		fmt.Println(versionstr)
		return
	}
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}

package dotgit

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StupidHod/go-git/core"
	"github.com/StupidHod/go-git/fixtures"
	osfs "github.com/StupidHod/go-git/utils/fs/os"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type SuiteDotGit struct {
	fixtures.Suite
}

var _ = Suite(&SuiteDotGit{})

func (s *SuiteDotGit) TestSetRefs(c *C) {
	tmp, err := ioutil.TempDir("", "dot-git")
	c.Assert(err, IsNil)
	defer os.RemoveAll(tmp)

	fs := osfs.NewOS(tmp)
	dir := New(fs)

	err = dir.SetRef(core.NewReferenceFromStrings(
		"refs/heads/foo",
		"e8d3ffab552895c19b9fcf7aa264d277cde33881",
	))

	c.Assert(err, IsNil)

	err = dir.SetRef(core.NewReferenceFromStrings(
		"refs/heads/symbolic",
		"ref: refs/heads/foo",
	))

	c.Assert(err, IsNil)

	refs, err := dir.Refs()
	c.Assert(err, IsNil)
	c.Assert(refs, HasLen, 2)

	ref := findReference(refs, "refs/heads/foo")
	c.Assert(ref, NotNil)
	c.Assert(ref.Hash().String(), Equals, "e8d3ffab552895c19b9fcf7aa264d277cde33881")

	ref = findReference(refs, "refs/heads/symbolic")
	c.Assert(ref, NotNil)
	c.Assert(ref.Target().String(), Equals, "refs/heads/foo")
}

func (s *SuiteDotGit) TestRefsFromPackedRefs(c *C) {
	fs := fixtures.Basic().ByTag(".git").One().DotGit()
	dir := New(fs)

	refs, err := dir.Refs()
	c.Assert(err, IsNil)

	ref := findReference(refs, "refs/remotes/origin/branch")
	c.Assert(ref, NotNil)
	c.Assert(ref.Hash().String(), Equals, "e8d3ffab552895c19b9fcf7aa264d277cde33881")

}
func (s *SuiteDotGit) TestRefsFromReferenceFile(c *C) {
	fs := fixtures.Basic().ByTag(".git").One().DotGit()
	dir := New(fs)

	refs, err := dir.Refs()
	c.Assert(err, IsNil)

	ref := findReference(refs, "refs/remotes/origin/HEAD")
	c.Assert(ref, NotNil)
	c.Assert(ref.Type(), Equals, core.SymbolicReference)
	c.Assert(string(ref.Target()), Equals, "refs/remotes/origin/master")

}

func (s *SuiteDotGit) TestRefsFromHEADFile(c *C) {
	fs := fixtures.Basic().ByTag(".git").One().DotGit()
	dir := New(fs)

	refs, err := dir.Refs()
	c.Assert(err, IsNil)

	ref := findReference(refs, "HEAD")
	c.Assert(ref, NotNil)
	c.Assert(ref.Type(), Equals, core.SymbolicReference)
	c.Assert(string(ref.Target()), Equals, "refs/heads/master")
}

func (s *SuiteDotGit) TestConfig(c *C) {
	fs := fixtures.Basic().ByTag(".git").One().DotGit()
	dir := New(fs)

	file, err := dir.Config()
	c.Assert(err, IsNil)
	c.Assert(filepath.Base(file.Filename()), Equals, "config")
}

func findReference(refs []*core.Reference, name string) *core.Reference {
	n := core.ReferenceName(name)
	for _, ref := range refs {
		if ref.Name() == n {
			return ref
		}
	}

	return nil
}

func (s *SuiteDotGit) TestObjectsPack(c *C) {
	f := fixtures.Basic().ByTag(".git").One()
	fs := f.DotGit()
	dir := New(fs)

	hashes, err := dir.ObjectPacks()
	c.Assert(err, IsNil)
	c.Assert(hashes, HasLen, 1)
	c.Assert(hashes[0], Equals, f.PackfileHash)
}

func (s *SuiteDotGit) TestObjectPack(c *C) {
	f := fixtures.Basic().ByTag(".git").One()
	fs := f.DotGit()
	dir := New(fs)

	pack, err := dir.ObjectPack(f.PackfileHash)
	c.Assert(err, IsNil)
	c.Assert(filepath.Ext(pack.Filename()), Equals, ".pack")
}

func (s *SuiteDotGit) TestObjectPackIdx(c *C) {
	f := fixtures.Basic().ByTag(".git").One()
	fs := f.DotGit()
	dir := New(fs)

	idx, err := dir.ObjectPackIdx(f.PackfileHash)
	c.Assert(err, IsNil)
	c.Assert(filepath.Ext(idx.Filename()), Equals, ".idx")
}

func (s *SuiteDotGit) TestObjectPackNotFound(c *C) {
	fs := fixtures.Basic().ByTag(".git").One().DotGit()
	dir := New(fs)

	pack, err := dir.ObjectPack(core.ZeroHash)
	c.Assert(err, Equals, ErrPackfileNotFound)
	c.Assert(pack, IsNil)

	idx, err := dir.ObjectPackIdx(core.ZeroHash)
	c.Assert(idx, IsNil)
}

func (s *SuiteDotGit) TestNewObject(c *C) {
	tmp, err := ioutil.TempDir("", "dot-git")
	c.Assert(err, IsNil)
	defer os.RemoveAll(tmp)

	fs := osfs.NewOS(tmp)
	dir := New(fs)
	w, err := dir.NewObject()
	c.Assert(err, IsNil)

	err = w.WriteHeader(core.BlobObject, 14)
	n, err := w.Write([]byte("this is a test"))
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 14)

	c.Assert(w.Hash().String(), Equals, "a8a940627d132695a9769df883f85992f0ff4a43")

	err = w.Close()
	c.Assert(err, IsNil)

	i, err := fs.Stat("objects/a8/a940627d132695a9769df883f85992f0ff4a43")
	c.Assert(err, IsNil)
	c.Assert(i.Size(), Equals, int64(34))
}

func (s *SuiteDotGit) TestObjects(c *C) {
	fs := fixtures.ByTag(".git").ByTag("unpacked").One().DotGit()
	dir := New(fs)

	hashes, err := dir.Objects()
	c.Assert(err, IsNil)
	c.Assert(hashes, HasLen, 187)
	c.Assert(hashes[0].String(), Equals, "0097821d427a3c3385898eb13b50dcbc8702b8a3")
	c.Assert(hashes[1].String(), Equals, "01d5fa556c33743006de7e76e67a2dfcd994ca04")
	c.Assert(hashes[2].String(), Equals, "03db8e1fbe133a480f2867aac478fd866686d69e")
}

func (s *SuiteDotGit) TestObject(c *C) {
	fs := fixtures.ByTag(".git").ByTag("unpacked").One().DotGit()
	dir := New(fs)

	hash := core.NewHash("03db8e1fbe133a480f2867aac478fd866686d69e")
	file, err := dir.Object(hash)
	c.Assert(err, IsNil)
	c.Assert(strings.HasSuffix(
		file.Filename(), "objects/03/db8e1fbe133a480f2867aac478fd866686d69e"),
		Equals, true,
	)
}

func (s *SuiteDotGit) TestObjectNotFound(c *C) {
	fs := fixtures.ByTag(".git").ByTag("unpacked").One().DotGit()
	dir := New(fs)

	hash := core.NewHash("not-found-object")
	file, err := dir.Object(hash)
	c.Assert(err, NotNil)
	c.Assert(file, IsNil)
}

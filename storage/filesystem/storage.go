// Package filesystem is a storage backend base on filesystems
package filesystem

import (
	"github.com/StupidHod/go-git/config"
	"github.com/StupidHod/go-git/core"
	"github.com/StupidHod/go-git/storage/filesystem/internal/dotgit"
	"github.com/StupidHod/go-git/utils/fs"
)

type Storage struct {
	dir *dotgit.DotGit
	fs  fs.Filesystem

	o *ObjectStorage
	r *ReferenceStorage
	c *ConfigStorage
}

func NewStorage(fs fs.Filesystem) (*Storage, error) {
	dir := dotgit.New(fs)
	o, err := newObjectStorage(dir)
	if err != nil {
		return nil, err
	}

	return &Storage{dir: dir, fs: fs, o: o}, nil
}

func (s *Storage) ObjectStorage() core.ObjectStorage {
	return s.o
}

func (s *Storage) ReferenceStorage() core.ReferenceStorage {
	if s.r != nil {
		return s.r
	}

	s.r = &ReferenceStorage{dir: s.dir}
	return s.r
}

func (s *Storage) ConfigStorage() config.ConfigStorage {
	if s.c != nil {
		return s.c
	}

	s.c = &ConfigStorage{dir: s.dir}
	return s.c
}

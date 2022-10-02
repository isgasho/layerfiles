package layerfile_graph

import (
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/layerfile"
	"io/fs"
	"io/ioutil"
	"path/filepath"
)

func FindLayerfiles(dir string) ([]*layerfile.Layerfile, error) {
	var res []*layerfile.Layerfile
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if filepath.Base(path) == ".git" {
			return filepath.SkipDir
		}
		if filepath.Base(path) == "Layerfile" && !d.IsDir() {
			contents, err := ioutil.ReadFile(filepath.Join(dir, path))
			if err != nil {
				return errors.Wrapf(err, "could not read Layerfile at %v", path)
			}
			lf, err := layerfile.ReadLayerfile(string(contents))
			if err != nil {
				return errors.Wrapf(err, "could not parse Layerfile at %v", path)
			}
			lf.FilePath = path
			res = append(res, lf)
		}
		return nil
	})
	return res, err
}

package snapshot_handler

import (
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/snapshot_handler/file_state_tree"
	"os"
	"path/filepath"
)

type Snapshot struct {
	StateTree file_state_tree.Node
	CurrInstructionHash string
}

type SnapshotHandler struct {
	CurrSnapshot Snapshot
	LayerfilePath string
}

func (s *SnapshotHandler) handleFileChange(path string) error {
	if _, err := os.Stat(filepath.Join(s.LayerfilePath, path)); os.IsNotExist(err) {
		return nil //read a non-existent file
	} else if err != nil {
		return errors.Wrapf(err, "could not read file at %v", path)
	}
	s.CurrSnapshot.StateTree.Name = s.LayerfilePath
	node := s.CurrSnapshot.StateTree.NodeFromPath(path)
	err := node.SetHashFromContent()
	if err != nil {
		return errors.Wrapf(err, "could not hash file at %v", path)
	}
	return nil
}
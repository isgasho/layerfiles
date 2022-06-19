package layerfile

import (
	"bytes"
	"testing"
)

func TestParsesParent(t *testing.T) {
	t.Parallel()

	layerfile := bytes.NewBufferString(
		`FROM github/ColinChartier/layer-kubernetes-base/Layerfile@130d951b92c9
RUN kubeadm init
`)
	tokenStream, err := tokenizeLayerfile(layerfile)
	if err != nil {
		t.Fatal(err)
	}

	instrs, err := parseInstructions(tokenStream)
	if err != nil {
		t.Fatal(err)
	}

	if len(instrs) != 2 {
		t.Fatalf("Expected 2 instructions, got %d", len(instrs))
	}

	from, instrs, err := parseFrom(instrs)
	if err != nil {
		t.Fatal(err)
	}
	if from != "github/ColinChartier/layer-kubernetes-base/Layerfile@130d951b92c9" {
		t.Fatalf("From was incorrect, got %v", from)
	}
}

package layerfile

import (
	"bytes"
	"fmt"
	"github.com/webappio/layerfiles/pkg/layerfile/instructions"
	"github.com/webappio/layerfiles/pkg/layerfile/lexer"
	"testing"
)

const ExampleLayerfile = `
FROM vm/ubuntu:18.04

LABEL status=merge display_name=my_cool_name

# install apt dependencies
RUN apt-get update
CACHE /var/lib/apt
RUN apt-get -y install python3 curl
CHECKPOINT

ENV a=b c d
ENV NODE_OPTIONS="--max-old-space-size=8192"
ENV a=` + "`echo hello`" + ` b=$(echo hello) c='echo hello' d="echo hello"
BUILD ENV GIT_BRANCH

# install app
WORKDIR /tmp/hello
COPY example.txt ./
RUN BACKGROUND python3 -m http.server 8080
RUN python3 -m http.server 8080& sleep 5
EXPOSE WEBSITE 8080 /api
EXPOSE WEBSITE https://localhost
EXPOSE TCP :60
EXPOSE TCP :1024 :8080
MEMORY 1G

CLONE "git@github.com:hello/my repo has spaces.git" /clone-dest
CLONE "a@a.a/git" /a /b DEFAULT='hello world' /clone-dest
CLONE https://github.com/webappio/docs.git services/web/app/routes/docs/docs

CHECKPOINT test-and-push
BUTTON deploy?

SECRET ENV thesecret variables

RUN if [ "$(curl localhost:8080/example.txt)" = "data from example.txt" ]; then \
      echo 'success!'; \
    else \
      echo 'failed!'; \
    fi

USER testuser---z00_

AWS link --region=us-east-1
AWS create-db-instance --cli-input-json=input.json

SKIP REMAINING IF API_EXTRA="" AND LAYERCI != true
SKIP REMAINING IF GIT_BRANCH =~ "m.*ster spaces" AND JOB_ID !=~ "layerci/.*"
SKIP REMAINING IF GIT_COMMIT_TITLE =~ "\[skip tests\]"

WAIT  some/other/Layerfile
SPLIT 5
`

func TestTokenization(t *testing.T) {
	t.Parallel()

	layerfile := bytes.NewBufferString(ExampleLayerfile)
	tokenStream, err := tokenizeLayerfile(layerfile)
	if err != nil {
		t.Error(err.Error())
		return
	}

	expectedTokens := []int{
		lexer.LayerfileFROM, lexer.LayerfileFROM_DATA,

		lexer.LayerfileLABEL, lexer.LayerfileLABEL_PAIR, lexer.LayerfileLABEL_PAIR, lexer.LayerfileLABEL_EOL,

		lexer.LayerfileRUN, lexer.LayerfileRUN_DATA,
		lexer.LayerfileCACHE, lexer.LayerfileFILE, lexer.LayerfileEND_OF_FILES,
		lexer.LayerfileRUN, lexer.LayerfileRUN_DATA,
		lexer.LayerfileCHECKPOINT, lexer.LayerfileCHECKPOINT_EOL,

		lexer.LayerfileENV, lexer.LayerfileENV_VALUE, lexer.LayerfileENV_VALUE, lexer.LayerfileENV_VALUE, lexer.LayerfileENV_EOL,
		lexer.LayerfileENV, lexer.LayerfileENV_VALUE, lexer.LayerfileENV_EOL,
		lexer.LayerfileENV, lexer.LayerfileENV_VALUE, lexer.LayerfileENV_VALUE, lexer.LayerfileENV_VALUE, lexer.LayerfileENV_VALUE, lexer.LayerfileENV_EOL,
		lexer.LayerfileBUILD_ENV, lexer.LayerfileBUILD_ENV_VALUE, lexer.LayerfileBUILD_ENV_EOL,

		lexer.LayerfileWORKDIR, lexer.LayerfileFILE, lexer.LayerfileEND_OF_FILES,
		lexer.LayerfileCOPY, lexer.LayerfileFILE, lexer.LayerfileFILE, lexer.LayerfileEND_OF_FILES,
		lexer.LayerfileRUN_BACKGROUND, lexer.LayerfileRUN_DATA,
		lexer.LayerfileRUN, lexer.LayerfileRUN_DATA,
		lexer.LayerfileEXPOSE_WEBSITE, lexer.LayerfileWEBSITE_ITEM, lexer.LayerfileWEBSITE_ITEM, lexer.LayerfileWEBSITE_EOL,
		lexer.LayerfileEXPOSE_WEBSITE, lexer.LayerfileWEBSITE_ITEM, lexer.LayerfileWEBSITE_EOL,
		lexer.LayerfileEXPOSE_TCP, lexer.LayerfileEXPOSE_TCP_ITEM, lexer.LayerfileEXPOSE_TCP_EOL,
		lexer.LayerfileEXPOSE_TCP, lexer.LayerfileEXPOSE_TCP_ITEM, lexer.LayerfileEXPOSE_TCP_ITEM, lexer.LayerfileEXPOSE_TCP_EOL,
		lexer.LayerfileMEMORY, lexer.LayerfileMEMORY_AMOUNT,

		lexer.LayerfileCLONE, lexer.LayerfileCLONE_VALUE, lexer.LayerfileCLONE_VALUE, lexer.LayerfileCLONE_EOL,
		lexer.LayerfileCLONE, lexer.LayerfileCLONE_VALUE, lexer.LayerfileCLONE_VALUE, lexer.LayerfileCLONE_VALUE, lexer.LayerfileCLONE_VALUE, lexer.LayerfileCLONE_VALUE, lexer.LayerfileCLONE_EOL,
		lexer.LayerfileCLONE, lexer.LayerfileCLONE_VALUE, lexer.LayerfileCLONE_VALUE, lexer.LayerfileCLONE_EOL,

		lexer.LayerfileCHECKPOINT, lexer.LayerfileCHECKPOINT_VALUE, lexer.LayerfileCHECKPOINT_EOL,
		lexer.LayerfileBUTTON, lexer.LayerfileBUTTON_DATA,

		lexer.LayerfileSECRET_ENV, lexer.LayerfileSECRET_ENV_VALUE, lexer.LayerfileSECRET_ENV_VALUE, lexer.LayerfileSECRET_ENV_EOL,

		lexer.LayerfileRUN, lexer.LayerfileRUN_DATA,
		lexer.LayerfileUSER, lexer.LayerfileUSER_NAME,

		lexer.LayerfileAWS, lexer.LayerfileAWS_VALUE, lexer.LayerfileAWS_VALUE, lexer.LayerfileAWS_EOL,
		lexer.LayerfileAWS, lexer.LayerfileAWS_VALUE, lexer.LayerfileAWS_VALUE, lexer.LayerfileAWS_EOL,

		lexer.LayerfileSKIP_REMAINING_IF, lexer.LayerfileSKIP_REMAINING_IF_VALUE, lexer.LayerfileSKIP_REMAINING_IF_AND, lexer.LayerfileSKIP_REMAINING_IF_VALUE, lexer.LayerfileSKIP_REMAINING_IF_EOL,
		lexer.LayerfileSKIP_REMAINING_IF, lexer.LayerfileSKIP_REMAINING_IF_VALUE, lexer.LayerfileSKIP_REMAINING_IF_AND, lexer.LayerfileSKIP_REMAINING_IF_VALUE, lexer.LayerfileSKIP_REMAINING_IF_EOL,

		lexer.LayerfileSKIP_REMAINING_IF, lexer.LayerfileSKIP_REMAINING_IF_VALUE, lexer.LayerfileSKIP_REMAINING_IF_EOL,

		lexer.LayerfileWAIT, lexer.LayerfileFILE, lexer.LayerfileEND_OF_FILES,
		lexer.LayerfileSPLIT, lexer.LayerfileSPLIT_NUMBER,
	}

	i := 0
	for tokenStream.HasToken() {
		token := tokenStream.Pop()
		if token.GetTokenType() != expectedTokens[i] {
			t.Errorf("Got a token of unexpected type at %d:%d, was %d, should be %d", token.GetLine(), token.GetColumn(), token.GetTokenType(), expectedTokens[i])
		}
		i += 1
	}
	if i != len(expectedTokens) {
		t.Fatalf("Did not get the right number of tokens, got %d, should be %d", i, len(expectedTokens))
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	layerfile := bytes.NewBufferString(ExampleLayerfile)
	tokenStream, err := tokenizeLayerfile(layerfile)
	if err != nil {
		t.Error(err.Error())
		return
	}
	instrs, err := parseInstructions(tokenStream)
	if err != nil {
		t.Error(err)
		return
	}

	expectedInstrs := []instructions.Instruction{
		&instructions.From{ImageId: "vm/ubuntu:18.04"},
		&instructions.Label{Label: []string{"status=merge", "display_name=my_cool_name"}},
		&instructions.Run{Command: "apt-get update"},
		&instructions.Cache{Dirs: []string{"/var/lib/apt"}},
		&instructions.Run{Command: "apt-get -y install python3 curl"},
		&instructions.Checkpoint{},
		&instructions.Env{Env: []string{"a=b", "c=d"}},
		&instructions.Env{Env: []string{"NODE_OPTIONS=\"--max-old-space-size=8192\""}},
		&instructions.Env{Env: []string{
			"a=`echo hello`",
			"b=$(echo hello)",
			"c='echo hello'",
			`d="echo hello"`,
		}},
		&instructions.BuildEnv{BuildEnv: []string{"GIT_BRANCH"}},
		&instructions.Workdir{Dir: "/tmp/hello"},
		&instructions.Copy{SourceFiles: []string{"example.txt"}, TargetFile: "./"},
		&instructions.Run{Command: "python3 -m http.server 8080", Type: instructions.RunTypeBackground},
		&instructions.Run{Command: "python3 -m http.server 8080& sleep 5"},
		&instructions.ExposeWebsite{Scheme: "http", Domain: "localhost", Port: 8080, Path: "/api"},
		&instructions.ExposeWebsite{Scheme: "https", Domain: "localhost", Port: 443, Path: "/"},
		&instructions.ExposeTcp{SourcePort: 60, DestPort: 60},
		&instructions.ExposeTcp{SourcePort: 1024, DestPort: 8080},
		&instructions.Memory{Amount: 1, Unit: "G"},
		&instructions.Clone{CloneURL: `"git@github.com:hello/my repo has spaces.git"`, Dest: "/clone-dest"},
		&instructions.Clone{CloneURL: `"a@a.a/git"`, Sources: []string{"/a", "/b"}, DefaultBranch: "'hello world'", Dest: "/clone-dest"},
		&instructions.Clone{CloneURL: `https://github.com/webappio/docs.git`, Dest: "services/web/app/routes/docs/docs"},
		&instructions.Checkpoint{Name: "test-and-push"},
		&instructions.Button{Message: "deploy?"},
		&instructions.SecretEnv{Secrets: []string{"thesecret", "variables"}},
		&instructions.Run{Command: "if [ \"$(curl localhost:8080/example.txt)\" = \"data from example.txt\" ]; then       echo 'success!';     else       echo 'failed!';     fi"},
		&instructions.User{Username: "testuser---z00_"},
		&instructions.AWS{Command: "link", Map: map[string]string{"region": "us-east-1"}},
		&instructions.AWS{Command: "create-db-instance", Map: map[string]string{"cli-input-json": "input.json"}},
		&instructions.SkipRemainingIf{SkipRemainingIf: []string{"API_EXTRA=", "AND", "LAYERCI!=true"}},
		&instructions.SkipRemainingIf{SkipRemainingIf: []string{"GIT_BRANCH=~m.*ster spaces", "AND", "JOB_ID!=~layerci/.*"}},
		&instructions.SkipRemainingIf{SkipRemainingIf: []string{"GIT_COMMIT_TITLE=~\\[skip tests\\]"}},
		&instructions.Wait{Targets: []string{"some/other/Layerfile"}},
		&instructions.Split{Count: 5},
	}

	if len(instrs) != len(expectedInstrs) {
		t.Fatalf("Did not get the right number of instructions. Got %d, should be %d", len(instrs), len(expectedInstrs))
	}

	for i := 0; i < len(instrs); i++ {
		if fmt.Sprintf("%+v", instrs[i]) != fmt.Sprintf("%+v", expectedInstrs[i]) {
			t.Fatalf("Two instrs were not equal. Got '%+v', should be '%+v'", instrs[i], expectedInstrs[i])
		}
		parsed, err := ParseInstruction(fmt.Sprintf("%+v", instrs[i]))
		if err != nil {
			t.Fatalf("Could not re-parse instruction (intr -> string -> instr) for %v: %v", instrs[i], err)
		}
		if fmt.Sprintf("%+v", parsed) != fmt.Sprintf("%+v", instrs[i]) {
			t.Fatalf("Parsed version of instruction is not the same (instr -> string -> instr): %v, %v", instrs[i], expectedInstrs[i])
		}
	}
}

func TestErrorsProperly(t *testing.T) {
	t.Parallel()

	layerfile := bytes.NewBufferString(
		`FROM vm/ubuntu:18.04
INVALIDDIRECTIVE hello world
`)
	_, err := tokenizeLayerfile(layerfile)
	if err == nil {
		t.Error("expected parsing invalid layerfile to return error")
	}
}

func TestParseCopyFails(t *testing.T) {
	t.Parallel()
	tokenStream, err := tokenizeLayerfile(bytes.NewBufferString("FROM vm/ubuntu:18.04\nCOPY a b c"))
	if err != nil {
		t.Error(err.Error())
		return
	}
	_, err = parseInstructions(tokenStream)
	if err == nil {
		t.Error("expected copying two directories to a target not ending with '/' to fail.")
		return
	}
}

func TestParseFrom(t *testing.T) {
	image, instrs, err := parseFrom([]instructions.Instruction{
		&instructions.From{ImageId: "vm/ubuntu:18.04"},
	})
	if err != nil {
		t.Error(err.Error())
	}
	if len(instrs) != 0 {
		t.Error("Expected there to be no non-from instructions")
	}
	if image != "vm/ubuntu:18.04" {
		t.Errorf("Expected %s to be vm/ubuntu:18.04", image)
	}

	image, instrs, err = parseFrom([]instructions.Instruction{
		&instructions.Run{Command: "echo hello"},
		&instructions.From{ImageId: "a"},
	})
	if err == nil {
		t.Error("Expected error if FROM is not first instruction in Layerfile")
	}

	image, instrs, err = parseFrom([]instructions.Instruction{
		&instructions.From{ImageId: "a"},
		&instructions.Run{Command: "echo 1"},
		&instructions.Run{Command: "echo 2"},
	})
	if err != nil {
		t.Error(err.Error())
	}
	if len(instrs) != 2 {
		t.Error("Expected there to be two instructions after removing the 'FROM' group")
	}
}

func TestSkipRemainingIfFailsFirstArgumentAnd(t *testing.T) {
	t.Parallel()
	layerfile := bytes.NewBufferString(
		`FROM vm/ubuntu:18.04
		SKIP REMAINING IF AND GIT_BRANCH=master
		`)

	tokenStream, err := tokenizeLayerfile(layerfile)
	if err != nil {
		t.Error(err.Error())
		return
	}
	_, err = parseInstructions(tokenStream)
	if err == nil {
		t.Error("expected SKIP REMAINING IF instruction with first value AND to fail.")
		return
	}
}

func TestSkipRemainingIfFailsAdjacentAnd(t *testing.T) {
	t.Parallel()
	layerfile := bytes.NewBufferString(
		`FROM vm/ubuntu:18.04
		SKIP REMAINING IF GIT_BRANCH=master AND AND LAYERCI=true
		`)

	tokenStream, err := tokenizeLayerfile(layerfile)

	if err != nil {
		t.Error(err.Error())
		return
	}
	_, err = parseInstructions(tokenStream)
	if err == nil {
		t.Error("expected SKIP REMAINING IF instruction with adjacent ANDs to fail.")
		return
	}
}

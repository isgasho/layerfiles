# Checklist for adding a new directive

- [ ] Add it to this grammar
- [ ] run "sanic run build_grammar"
- [ ] Create a file in linkedlib/layerfile/instructions
- [ ] Link to that file in linkedlib/layerfile/parser # parseInstruction
- [ ] Create a test in parser_test.go
- [ ] Add it as a keyword to services/web/app/commonjs/layerfile-editor/layerfile-mode.js
- [ ] Add documentation for it to https://github.com/webappio/docs/
- [ ] Add functionality for it in services/vm-worker/pkg/run_layerfile_job/instruction_processors
- [ ] Link that functionality in services/vm-worker/pkg/run_layerfile_job/instruction_processors/process.go # runInstruction
lexer grammar Layerfile;

WS : [ \t\r\n]+ -> skip;
COMMENT: '#' ~[\r\n]* -> skip;

BUTTON: 'BUTTON' -> pushMode(BUTTON_INSTR);
CACHE: 'CACHE' -> pushMode(READ_FILES);
CHECKPOINT: 'CHECKPOINT' -> pushMode(CHECKPOINT_INSTR);
CLONE: 'CLONE ' -> pushMode(CLONE_INSTR);
COPY: 'COPY' -> pushMode(READ_FILES);
ENV: 'ENV' -> pushMode(ENV_INSTR);
BUILD_ENV: 'BUILD ENV ' -> pushMode(BUILD_ENV_INSTR);
FROM: 'FROM ' -> pushMode(FROM_INSTR);
MEMORY: 'MEMORY ' -> pushMode(MEMORY_INSTR);
RUN: 'RUN ' -> pushMode(RUN_INSTR);
RUN_BACKGROUND: 'RUN BACKGROUND ' -> pushMode(RUN_INSTR);
RUN_REPEATABLE: 'RUN REPEATABLE ' -> pushMode(RUN_INSTR);
SECRET_ENV: 'SECRET ENV ' -> pushMode(SECRET_ENV_INSTR);
SETUP_FILE: 'SETUP FILE ' -> pushMode(READ_FILES);
SKIP_REMAINING_IF: 'SKIP REMAINING IF ' -> pushMode(SKIP_REMAINING_IF_INSTR);
SPLIT: 'SPLIT ' -> pushMode(SPLIT_INSTR);
EXPOSE_WEBSITE: 'EXPOSE WEBSITE ' -> pushMode(EXPOSE_WEBSITE_INSTR);
USER: 'USER ' -> pushMode(USER_INSTR);
LABEL: 'LABEL ' -> pushMode(LABEL_INSTR);
WAIT: 'WAIT ' -> pushMode(READ_FILES);
WORKDIR: 'WORKDIR ' -> pushMode(READ_FILES);
AWS: 'AWS ' -> pushMode(AWS_INSTR);


mode BUILD_ENV_INSTR;
BUILD_ENV_VALUE: ~[ \r\n]+;
BUILD_ENV_WS: [ \t]+ -> skip;
BUILD_ENV_EOL: (('\r'? '\n') | '\r' | EOF) -> popMode;
BUILD_ENV_COMMENT: '#' ~[\r\n]* -> skip;


mode BUTTON_INSTR;
BUTTON_DATA: (('\r'? '\n') | '\r' | EOF) -> popMode;
BUTTON_MORE: . -> more;
BUTTON_COMMENT: '#' ~[\r\n]* -> skip;


mode CHECKPOINT_INSTR;
CHECKPOINT_VALUE: ~[ \t\r\n]+;
CHECKPOINT_WS : [ \t] -> skip;
CHECKPOINT_EOL: (('\r'? '\n') | '\r' | EOF) -> popMode;
CHECKPOINT_COMMENT: '#' ~[\r\n]* -> skip;


mode CLONE_INSTR;

CLONE_VALUE:
        '"' .*? '"'
        | '\'' .*? '\''
        | 'DEFAULT="' .*? '"'
        | 'DEFAULT=\'' .*? '\''
        | ~[ \t\r\n]+
        ;

CLONE_WS: [ \t]+ -> skip;
CLONE_EOL: (('\r'? '\n') | '\r' | EOF) -> popMode;


mode ENV_INSTR;
fragment ENV_VALUE_FRAG: '"' .*? '"'
         | '\'' .*? '\''
         | '`' .*? '`'
         | '$(' .*? ')'
         | ~[ \r\n]+
         ;
fragment ENV_KEY: ('0'..'9' | 'a'..'z' | 'A'..'Z' | '_' | '-')+;

ENV_VALUE:
    ENV_KEY '=' ENV_VALUE_FRAG
    | ENV_VALUE_FRAG
    ;

ENV_WS: [ \t]+ -> skip;
ENV_EOL: (('\r'? '\n') | '\r' | EOF) -> popMode;
ENV_COMMENT: '#' ~[\r\n]* -> skip;

mode LABEL_INSTR;
LABEL_ID: [A-Za-z0-9_-]+;
LABEL_PAIR: LABEL_ID '=' LABEL_ID;
LABEL_WS: [ \t]+ -> skip;
LABEL_EOL: (('\r'? '\n') | '\r' | EOF) -> popMode;
LABEL_COMMENT: '#' ~[\r\n]* -> skip;

mode EXPOSE_WEBSITE_INSTR;
WEBSITE_EOL: (('\r'? '\n') | '\r' | EOF) -> popMode;
WEBSITE_ITEM: ~[ \r\n\t]+ ;
WEBSITE_WS: [ \t]+ -> skip;
WEBSITE_COMMENT: '#' ~[\r\n]* -> skip;


mode FROM_INSTR;
FROM_DATA: (('\r'? '\n') | '\r' | EOF) -> popMode;
FROM_MORE: . -> more;
FROM_COMMENT: '#' ~[\r\n]* -> skip;


mode AWS_INSTR;
AWS_VALUE:
    ('0'..'9'|'A'..'Z'|'a'..'z'|'.'|'_'|'-'|'='|'"'|'\'')+ ;
AWS_WS: [ \t]+ -> skip;
AWS_EOL: (('\r'? '\n') | '\r' | EOF) -> popMode;
AWS_COMMENT: '#' ~[\r\n]* -> skip;


mode MEMORY_INSTR;
MEMORY_EOF: (('\r'? '\n') | '\r' | EOF) -> skip, popMode;
MEMORY_AMOUNT: ('0'..'9')+ ('G' | 'g' | 'M' | 'm' | 'K' | 'k')?;
MEMORY_COMMENT: '#' ~[\r\n]* -> skip;


mode RUN_INSTR;
RUN_DATA: (('\r'? '\n') | '\r' | EOF) -> popMode;
RUN_COMMAND: . -> more;


mode SECRET_ENV_INSTR;
SECRET_ENV_VALUE: ~[ \r\n=]+;
SECRET_ENV_WS: [ \t]+ -> skip;
SECRET_ENV_EOL: (('\r'? '\n') | '\r' | EOF) -> popMode;
SECRET_ENV_COMMENT: '#' ~[\r\n]* -> skip;

mode SKIP_REMAINING_IF_INSTR;
SKIP_REMAINING_IF_VALUE:
    ('0'..'9'|'A'..'Z'|'a'..'z'|'_'|'-')+ [ \t]* ('!=~' | '!=' | '=~' | '=') [ \t]*
    ('"' .*? '"'
     | '\'' .*? '\''
     | ~[ \r\n]+
    );
SKIP_REMAINING_IF_AND: 'AND';
SKIP_REMAINING_IF_WS: [ \t]+ -> skip;
SKIP_REMAINING_IF_EOL: (('\r'? '\n') | '\r' | EOF) -> popMode;

mode SPLIT_INSTR;
SPLIT_NUMBER: ('0'..'9')+ -> popMode;
SPLIT_WS: [ \t]+ -> skip;


mode USER_INSTR;
USER_NAME: ('0'..'9'|'A'..'Z'|'a'..'z'|'.'|'_'|'-')+ -> popMode;
USER_COMMENT: '#' ~[\r\n]* -> skip;


mode READ_FILES;
END_OF_FILES: (('\r'? '\n') | '\r' | EOF) -> popMode;
FILE: ~[ \r\n\t]+
    | '"' .*? '"';
FILE_WS: [ \t]+ -> skip;
FILE_COMMENT: '#' ~[\r\n]* -> skip;

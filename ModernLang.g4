grammar ModernLang;

// ---------------- LEXICAL ----------------

IDENTIFIER  : [a-zA-Z_] [a-zA-Z0-9_]* ;
NUMBER      : [0-9]+ ;
STRING      : '"' (~["\\] | '\\' .)* '"' ;
BOOLEAN     : 'true' | 'false' ;
NULL        : 'null' ;

WS          : [ \t\r\n]+ -> skip ;
COMMENT     : '//' ~[\r\n]* -> skip ;
BLOCK_COMMENT : '/*' .*? '*/' -> skip ;

// ---------------- PROGRAM ----------------

program
    : (moduleDecl | statement)* EOF
    ;

// ---------------- MODULES ----------------

moduleDecl
    : 'module' IDENTIFIER block
    ;

importDecl
    : 'import' qualifiedName ('as' IDENTIFIER)? ';'
    ;

qualifiedName
    : IDENTIFIER ('.' IDENTIFIER)*
    ;

// ---------------- STATEMENTS ----------------

statement
    : variableDecl
    | functionDecl
    | classDecl
    | structDecl
    | enumDecl
    | contractDecl
    | importDecl
    | matchStmt
    | block
    | expression ';'
    ;

block
    : '{' statement* '}'
    ;

// ---------------- VARIABLES ----------------

variableDecl
    : ('let' | 'var') IDENTIFIER (':' type_)? '=' expression ';'
    ;

// ---------------- FUNCTIONS ----------------

functionDecl
    : 'fn' IDENTIFIER typeParams? '(' paramList? ')' (':' type_)? ('async')? contractBlock? (block | '=' expression ';')
    ;

paramList
    : param (',' param)*
    ;

param
    : IDENTIFIER ':' type_
    ;

typeParams
    : '<' IDENTIFIER (',' IDENTIFIER)* '>'
    ;

// ---------------- CONTRACTS ----------------

contractBlock
    : 'contract' '{' contractClause* '}'
    ;

contractClause
    : 'returns' '(' expression? ')' '=>' expression ';'
    ;

// ---------------- CLASSES ----------------

classDecl
    : 'class' IDENTIFIER '(' paramList? ')' (':' IDENTIFIER)? block?
    ;

// ---------------- STRUCTS ----------------

structDecl
    : 'struct' IDENTIFIER '(' paramList? ')' block?
    ;

// ---------------- ENUMS ----------------

enumDecl
    : 'enum' IDENTIFIER typeParams? '{' enumCase* '}'
    ;

enumCase
    : IDENTIFIER ('(' paramList? ')')? ';'
    ;

// ---------------- MATCH ----------------

matchStmt
    : 'match' expression '{' matchCase+ '}'
    ;

matchCase
    : pattern '=>' statement
    ;

pattern
    : literalPattern
    | IDENTIFIER
    | structPattern
    | objectPattern
    ;

literalPattern
    : NUMBER
    | STRING
    | BOOLEAN
    | NULL
    ;

structPattern
    : IDENTIFIER '(' (pattern (',' pattern)*)? ')'
    ;

objectPattern
    : '{' pairPattern (',' pairPattern)* '}'
    ;

pairPattern
    : (STRING | IDENTIFIER) ':' pattern
    ;

// ---------------- EXPRESSIONS ----------------

expression
    : assignment
    ;

assignment
    : conditional ('=' expression)?
    ;

conditional
    : logicalOr ('?:' expression)?
    ;

logicalOr
    : logicalAnd ('||' logicalAnd)*
    ;

logicalAnd
    : equality ('&&' equality)*
    ;

equality
    : comparison (('==' | '!=') comparison)*
    ;

comparison
    : additive (( '<' | '<=' | '>' | '>=' | 'is' | '!is' ) additive)*
    ;

additive
    : multiplicative (('+' | '-') multiplicative)*
    ;

multiplicative
    : unary (('*' | '/' | '%') unary)*
    ;

unary
    : ('!' | '-') unary
    | postfix
    ;

postfix
    : primary ( 
        '.' IDENTIFIER
      | '[' expression ']'
      | '?.' IDENTIFIER
      | '!!'
      | '(' argumentList? ')'
    )*
    ;

argumentList
    : expression (',' expression)*
    ;

primary
    : NUMBER
    | STRING
    | BOOLEAN
    | NULL
    | IDENTIFIER
    | arrayLiteral
    | objectLiteral
    | awaitExpr
    | '(' expression ')'
    ;

arrayLiteral
    : '[' (expression (',' expression)*)? ']'
    ;

objectLiteral
    : '{' (pair (',' pair)*)? '}'
    ;

pair
    : (STRING | IDENTIFIER) ':' expression
    ;

awaitExpr
    : 'await' expression
    ;

// ---------------- TYPES ----------------

type_
    : IDENTIFIER typeArgs? '?'?
    ;

typeArgs
    : '<' type_ (',' type_)* '>'
    ;

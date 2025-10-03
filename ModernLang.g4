grammar ModernLang;

// ---------------- LEXICAL ----------------

PACKAGE    : 'package';
MODULE     : 'module';
IMPORT     : 'import';
AS         : 'as';
LET        : 'let';
VAR        : 'var';
FN         : 'fn';
ASYNC      : 'async';
CONTRACT   : 'contract';
RETURNS    : 'returns';
CLASS      : 'class';
STRUCT     : 'struct';
ENUM       : 'enum';
INTERFACE  : 'interface';
EXT        : 'ext';
MATCH      : 'match';
IF         : 'if';
ELSE       : 'else';
WHILE      : 'while';
FOR        : 'for';
USING      : 'using';
TRY        : 'try';
CATCH      : 'catch';
FINALLY    : 'finally';
THROW      : 'throw';
RETURN     : 'return';
BREAK      : 'break';
CONTINUE   : 'continue';
CONDITION  : 'condition';
WHEN       : 'when';
AWAIT      : 'await';
TRUE       : 'true';
FALSE      : 'false';
NULL       : 'null';
IS         : 'is';
NOTIS      : '!is';

AMPERSAND  : '&';

PLUS_ASSIGN    : '+=';
MINUS_ASSIGN   : '-=';
STAR_ASSIGN    : '*=';
SLASH_ASSIGN   : '/=';
PERCENT_ASSIGN : '%=';
ELVIS          : '?:';
SAFE_DOT       : '?.';
NON_NULL       : '!!';

IDENTIFIER : [a-zA-Z_] [a-zA-Z0-9_]* ;
NUMBER     : [0-9]+ ('.' [0-9]+)? ;
STRING     : '"' (~['"'\\] | '\\' .)* '"';
FORMATSTRING
           : 'f"' (~['"'\\] | '\\' .)* '"'
           | 'f"""' .*? '"""'
           ;
RAWSTRING  : 'r"' (~['"'\\])* '"'
           | 'r"""' .*? '"""'
           | '`' .*? '`'
           ;
BOOLEAN    : TRUE | FALSE;

WS           : [ \t\r\n]+ -> skip;
COMMENT      : '//' ~[\r\n]* -> skip;
BLOCK_COMMENT: '/*' .*? '*/' -> skip;

// ---------------- PROGRAM ----------------

program
    : packageDecl? (moduleDecl | statement)* EOF
    ;

packageDecl
    : PACKAGE IDENTIFIER ';'?
    ;

// ---------------- MODULES ----------------

moduleDecl
    : MODULE IDENTIFIER block?
    ;

importDecl
    : IMPORT (
          IDENTIFIER STRING (AS IDENTIFIER)?
        | importPath (AS IDENTIFIER)?
      ) ';'
    ;

importPath
    : IDENTIFIER ('.' IDENTIFIER)*
    | STRING
    ;

qualifiedName
    : IDENTIFIER ('.' IDENTIFIER)*
    ;

// ---------------- STATEMENTS ----------------

statement
    : variableDecl
    | functionDecl
    | extensionDecl
    | classDecl
    | structDecl
    | enumDecl
    | interfaceDecl
    | contractDecl
    | importDecl
    | matchStmt
    | ifStmt
    | whileStmt
    | forStmt
    | usingStmt
    | tryStmt
    | throwStmt
    | returnStmt
    | breakStmt
    | continueStmt
    | conditionStmt
    | block
    | expression ';'
    ;

block
    : '{' statement* '}'
    ;

// ---------------- VARIABLES ----------------

variableDecl
    : (LET | VAR) IDENTIFIER (':' type_)? '=' expression ';'
    ;

// ---------------- FUNCTIONS ----------------

functionDecl
    : FN IDENTIFIER typeParams? '(' paramList? ')' (':' type_)? (ASYNC)? contractBlock? (block | ('=' | '=>') expression ';'?)
    ;

extensionDecl
    : EXT FN typeAnnotation '.' IDENTIFIER typeParams? '(' paramList? ')' (':' type_)? (ASYNC)? contractBlock? (block | ('=' | '=>') expression ';'?)
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
    : CONTRACT '{' contractClause* '}'
    ;

contractClause
    : RETURNS '(' expression? ')' '=>' expression ';'
    ;

// ---------------- TYPES ----------------

type_
    : typeAnnotation
    ;

typeAnnotation
    : IDENTIFIER typeArgs? '?'?
    ;

typeArgs
    : '<' typeAnnotation (',' typeAnnotation)* '>'
    ;

// ---------------- CLASSES ----------------

classDecl
    : CLASS IDENTIFIER '(' paramList? ')' (':' IDENTIFIER)? block?
    ;

structDecl
    : STRUCT IDENTIFIER '(' paramList? ')' block?
    ;

enumDecl
    : ENUM IDENTIFIER typeParams? '{' enumCase* '}'
    ;

enumCase
    : IDENTIFIER ('(' paramList? ')')? ';'
    ;

interfaceDecl
    : INTERFACE IDENTIFIER '{' interfaceMember* '}'
    ;

interfaceMember
    : FN IDENTIFIER '(' paramList? ')' (':' type_)? ';'
    ;

contractDecl
    : CONTRACT IDENTIFIER block
    ;

// ---------------- CONTROL FLOW ----------------

ifStmt
    : IF expression block (ELSE statement)?
    ;

whileStmt
    : WHILE expression block
    ;

forStmt
    : FOR '(' forInit? ';' expression? ';' expression? ')' block
    ;

forInit
    : variableBinding
    | expression
    ;

variableBinding
    : (LET | VAR) IDENTIFIER (':' type_)? '=' expression
    ;

usingStmt
    : USING (IDENTIFIER '=')? expression block
    ;

tryStmt
    : TRY block catchClause? finallyClause?
    ;

catchClause
    : CATCH ('(' IDENTIFIER? ')')? block
    ;

finallyClause
    : FINALLY block
    ;

throwStmt
    : THROW expression ';'?
    ;

returnStmt
    : RETURN expression? ';'?
    ;

breakStmt
    : BREAK ';'?
    ;

continueStmt
    : CONTINUE ';'?
    ;

conditionStmt
    : CONDITION '{' conditionClause* conditionElse? '}'
    ;

conditionClause
    : WHEN expression '=>' block
    ;

conditionElse
    : ELSE '=>' block
    ;

// ---------------- MATCH ----------------

matchStmt
    : MATCH expression '{' matchCase+ '}'
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
    | FORMATSTRING
    | RAWSTRING
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
    : conditional (( '=' | PLUS_ASSIGN | MINUS_ASSIGN | STAR_ASSIGN | SLASH_ASSIGN | PERCENT_ASSIGN ) expression)?
    ;

conditional
    : logicalOr (ELVIS expression)?
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
    : additive (( '<' | '<=' | '>' | '>=' | IS | NOTIS ) additive)*
    ;

additive
    : multiplicative (('+' | '-') multiplicative)*
    ;

multiplicative
    : unary (('*' | '/' | '%') unary)*
    ;

unary
    : ('!' | '-' | AMPERSAND | '*') unary
    | postfix
    ;

AMPERSAND : '&';

postfix
    : primary (
        '.' IDENTIFIER
      | '[' expression ']'
      | SAFE_DOT IDENTIFIER
      | NON_NULL
      | '(' argumentList? ')'
    )*
    ;

argumentList
    : expression (',' expression)*
    ;

primary
    : NUMBER
    | STRING
    | FORMATSTRING
    | RAWSTRING
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
    : AWAIT expression
    ;

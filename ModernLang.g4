grammar ModernLang;

// ---------------- LEXICAL ----------------

PACKAGE   : 'package';
MODULE    : 'module';
IMPORT    : 'import';
AS        : 'as';
LET       : 'let';
VAR       : 'var';
FN        : 'fn';
ASYNC     : 'async';
CONTRACT  : 'contract';
RETURNS   : 'returns';
CLASS     : 'class';
STRUCT    : 'struct';
ENUM      : 'enum';
INTERFACE : 'interface';
EXT       : 'ext';
MATCH     : 'match';
IF        : 'if';
ELSE      : 'else';
WHILE     : 'while';
FOR       : 'for';
USING     : 'using';
TRY       : 'try';
CATCH     : 'catch';
FINALLY   : 'finally';
THROW     : 'throw';
RETURN    : 'return';
BREAK     : 'break';
CONTINUE  : 'continue';
CONDITION : 'condition';
WHEN      : 'when';
AWAIT     : 'await';
TRUE      : 'true';
FALSE     : 'false';
NULL      : 'null';
IS        : 'is';
NOTIS     : '!is';

ARROW          : '=>';
ELVIS          : '?:';
SAFE_DOT       : '?.';
NON_NULL       : '!!';
PLUS_ASSIGN    : '+=';
MINUS_ASSIGN   : '-=';
STAR_ASSIGN    : '*=';
SLASH_ASSIGN   : '/=';
PERCENT_ASSIGN : '%=';
ASSIGN         : '=';
PLUS           : '+';
MINUS          : '-';
ASTERISK       : '*';
SLASH          : '/';
PERCENT        : '%';
BANG           : '!';
QUESTION       : '?';
COLON          : ':';
AMPERSAND      : '&';
EQ             : '==';
NOT_EQ         : '!=';
LT             : '<';
LTE            : '<=';
GT             : '>';
GTE            : '>=';
OR             : '||';
AND            : '&&';
COMMA          : ',';
DOT            : '.';
SEMICOLON      : ';';
LPAREN         : '(';
RPAREN         : ')';
LBRACE         : '{';
RBRACE         : '}';
LBRACKET       : '[';
RBRACKET       : ']';

IDENTIFIER : LETTER (LETTER | DIGIT)* ;
NUMBER     : DIGIT+ ('.' DIGIT+)? ;
STRING     : '"' STRING_CHAR* '"'
           | TRIPLE_QUOTE .*? TRIPLE_QUOTE
           ;
FORMATSTRING
           : 'f"' STRING_CHAR* '"'
           | 'f' TRIPLE_QUOTE .*? TRIPLE_QUOTE
           ;
RAWSTRING  : 'r"' RAW_CHAR* '"'
           | 'r' TRIPLE_QUOTE .*? TRIPLE_QUOTE
           | '`' RAW_BACKTICK* '`'
           ;

fragment STRING_CHAR : '\\' . | ~['"\\] ;
fragment RAW_CHAR    : ~['"\\] ;
fragment RAW_BACKTICK: ~'`' ;
fragment LETTER      : [a-zA-Z_];
fragment DIGIT       : [0-9];
fragment TRIPLE_QUOTE: '"""';

WS            : [ \t\r\n]+ -> skip;
COMMENT       : '//' ~[\r\n]* -> skip;
BLOCK_COMMENT : '/*' .*? '*/' -> skip;

// ---------------- PROGRAM ----------------

program
    : packageDecl? topLevelItem* EOF
    ;

topLevelItem
    : moduleDecl
    | declaration
    | statement
    ;

packageDecl
    : PACKAGE IDENTIFIER SEMICOLON?
    ;

moduleDecl
    : MODULE IDENTIFIER block
    ;

importDecl
    : IMPORT (IDENTIFIER STRING (AS IDENTIFIER)? | importPath (AS IDENTIFIER)?) SEMICOLON
    ;

importPath
    : IDENTIFIER (DOT IDENTIFIER)*
    | STRING
    ;

declaration
    : variableDecl
    | functionDecl
    | extensionDecl
    | classDecl
    | structDecl
    | enumDecl
    | interfaceDecl
    | contractDecl
    | importDecl
    ;

statement
    : declaration
    | flowStmt
    | block
    | expressionStmt
    ;

flowStmt
    : matchStmt
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
    ;

expressionStmt
    : expression SEMICOLON?
    ;

block
    : LBRACE statement* RBRACE
    ;

// ---------------- DECLARATIONS ----------------

variableDecl
    : (LET | VAR) IDENTIFIER (COLON type_)? ASSIGN expression SEMICOLON
    ;

functionDecl
    : FN IDENTIFIER typeParams? parameterClause returnType? ASYNC? contractBlock? functionBody
    ;

extensionDecl
    : EXT FN typeAnnotation DOT IDENTIFIER typeParams? parameterClause returnType? ASYNC? contractBlock? functionBody
    ;

parameterClause
    : LPAREN paramList? RPAREN
    ;

returnType
    : COLON type_
    ;

paramList
    : param (COMMA param)*
    ;

param
    : IDENTIFIER COLON type_
    ;

typeParams
    : LT IDENTIFIER (COMMA IDENTIFIER)* GT
    ;

functionBody
    : block
    | (ASSIGN | ARROW) expression SEMICOLON?
    ;

contractBlock
    : CONTRACT LBRACE contractClause* RBRACE
    ;

contractClause
    : RETURNS LPAREN expression? RPAREN ARROW expression SEMICOLON
    ;

type_
    : typeAnnotation
    ;

typeAnnotation
    : IDENTIFIER typeArgs? QUESTION?
    ;

typeArgs
    : LT typeAnnotation (COMMA typeAnnotation)* GT
    ;

classDecl
    : CLASS IDENTIFIER parameterClause (COLON IDENTIFIER)? block?
    ;

structDecl
    : STRUCT IDENTIFIER parameterClause block?
    ;

enumDecl
    : ENUM IDENTIFIER typeParams? LBRACE enumCase* RBRACE
    ;

enumCase
    : IDENTIFIER (LPAREN paramList? RPAREN)? SEMICOLON
    ;

interfaceDecl
    : INTERFACE IDENTIFIER LBRACE interfaceMember* RBRACE
    ;

interfaceMember
    : FN IDENTIFIER parameterClause returnType? SEMICOLON
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
    : FOR LPAREN forInit? SEMICOLON expression? SEMICOLON expression? RPAREN block
    ;

forInit
    : variableBinding
    | expression
    ;

variableBinding
    : (LET | VAR) IDENTIFIER (COLON type_)? ASSIGN expression
    ;

usingStmt
    : USING (IDENTIFIER ASSIGN)? expression block
    ;

tryStmt
    : TRY block catchClause? finallyClause?
    ;

catchClause
    : CATCH (LPAREN IDENTIFIER? RPAREN)? block
    ;

finallyClause
    : FINALLY block
    ;

throwStmt
    : THROW expression SEMICOLON?
    ;

returnStmt
    : RETURN expression? SEMICOLON?
    ;

breakStmt
    : BREAK SEMICOLON?
    ;

continueStmt
    : CONTINUE SEMICOLON?
    ;

conditionStmt
    : CONDITION LBRACE conditionClause* conditionElse? RBRACE
    ;

conditionClause
    : WHEN expression ARROW block
    ;

conditionElse
    : ELSE ARROW block
    ;

// ---------------- MATCH ----------------

matchStmt
    : MATCH expression LBRACE matchCase+ RBRACE
    ;

matchCase
    : pattern ARROW statement
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
    | TRUE
    | FALSE
    | NULL
    ;

structPattern
    : IDENTIFIER LPAREN (pattern (COMMA pattern)*)? RPAREN
    ;

objectPattern
    : LBRACE pairPattern (COMMA pairPattern)* RBRACE
    ;

pairPattern
    : (STRING | IDENTIFIER) COLON pattern
    ;

// ---------------- EXPRESSIONS ----------------

expression
    : assignment
    ;

assignment
    : conditional (assignmentOperator expression)?
    ;

assignmentOperator
    : ASSIGN
    | PLUS_ASSIGN
    | MINUS_ASSIGN
    | STAR_ASSIGN
    | SLASH_ASSIGN
    | PERCENT_ASSIGN
    ;

conditional
    : logicalOr (ELVIS expression)?
    ;

logicalOr
    : logicalAnd (OR logicalAnd)*
    ;

logicalAnd
    : equality (AND equality)*
    ;

equality
    : comparison ((EQ | NOT_EQ) comparison)*
    ;

comparison
    : additive ((LT | LTE | GT | GTE | IS | NOTIS) additive)*
    ;

additive
    : multiplicative ((PLUS | MINUS) multiplicative)*
    ;

multiplicative
    : unary ((ASTERISK | SLASH | PERCENT) unary)*
    ;

unary
    : (BANG | MINUS | AMPERSAND | ASTERISK) unary
    | postfix
    ;

postfix
    : primary postfixPart*
    ;

postfixPart
    : DOT IDENTIFIER
    | SAFE_DOT IDENTIFIER
    | LBRACKET expression RBRACKET
    | NON_NULL
    | LPAREN argumentList? RPAREN
    ;

argumentList
    : expression (COMMA expression)*
    ;

primary
    : NUMBER
    | STRING
    | FORMATSTRING
    | RAWSTRING
    | TRUE
    | FALSE
    | NULL
    | IDENTIFIER
    | arrayLiteral
    | objectLiteral
    | awaitExpr
    | LPAREN expression RPAREN
    ;

arrayLiteral
    : LBRACKET (expression (COMMA expression)*)? RBRACKET
    ;

objectLiteral
    : LBRACE (pair (COMMA pair)*)? RBRACE
    ;

pair
    : (STRING | IDENTIFIER) COLON expression
    ;

awaitExpr
    : AWAIT expression
    ;

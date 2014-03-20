%%{
    machine scanner;
    write data nofinal;

    include "action.rl";

    char_visible   = ^0..32;
    char_inline    = char_visible | [ \t];
    char_break     = [\r\n];
    char_any       = char_inline | char_break;
    char_invalid   = ^char_any;
    char_space     = [ \t] | char_break;
    char_delimiter = [{},];

    newline            = char_break | '\r\n';
    inline_comment     = '//' char_inline* (newline | '');
#   comment_char       = char_any*;
#   general_comment    = '/*' comment_char :> '*/';
#   comment            = inline_comment | general_comment;

#   raw_char           = char_any - '`';
#   raw_string         = '`' raw_char* '`';
    interpreted_char   = (char_inline - '"') | '\\"';
    interpreted_string = '"' interpreted_char* '"';
    unquoted_char      = char_visible - char_delimiter;
    unquoted_string    = unquoted_char+;
#   string             = (raw_string | interpreted_string | unquoted_string) ':'?;
    string             = (interpreted_string | unquoted_string) ':'?;

    main := |*
       char_invalid    => onInvalid;
       inline_comment  => onComment;
#      general_comment => onComment;
       '{'             => onLeftBrace;
       '}' ':'?        => onRightBrace;
       ','             => onComma;
       string          => onString;
       char_space+;
    *|;
}%%

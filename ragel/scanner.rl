%%{
    machine scanner;
    write data nofinal;

    include "action.rl";

    char_nonctrl   = ^0..32;
    char_inline    = char_nonctrl | [ \t];
    char_lbreak    = [\r\n];
    char_any       = char_inline | char_lbreak;
    char_invalid   = ^char_any;
    char_space     = [ \t] | char_lbreak;
    char_delimiter = [{},];

    newline            = char_lbreak | '\r\n';
    inline_comment     = '//' char_inline* (newline | '');
#   comment_char       = char_any*;
#   general_comment    = '/*' comment_char :> '*/';
#   comment            = inline_comment | general_comment;

#   raw_char           = char_any - '`';
#   raw_string         = '`' raw_char* '`';
    quoted_char        = (char_inline - '"') | '\\"';
    quoted_string      = '"' quoted_char* '"';
    unquoted_char      = char_nonctrl - char_delimiter;
    unquoted_string    = unquoted_char+;
#   string             = (raw_string | quoted_string | unquoted_string);
    string             = (quoted_string | unquoted_string);

    main := |*
       char_invalid    => onInvalid;
       inline_comment  => onComment;
#      general_comment => onComment;
       '{'             => onLeftBrace;
       '}'             => onRightBrace;
       ','             => onComma;
       string          => onString;
       char_space+;
    *|;
}%%

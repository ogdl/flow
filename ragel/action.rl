%%{
    machine scanner;

    action onInvalid {
        err = fmt.Errorf("Invalid character.")
        fbreak;
    }

    action onLeftBrace {
        tok = m.newToken(tokenLeftBrace)
        fbreak;
    }

    action onRightBrace {
        tok = m.newToken(tokenRightBrace)
        fbreak;
    }

    action onComma {
        tok = m.newToken(tokenComma)
        fbreak;
    }

    action onString {
        tok = m.newToken(tokenString)
        fbreak;
    }

    action onComment {
        tok = m.newToken(tokenComment)
        fbreak;
    }
}%%

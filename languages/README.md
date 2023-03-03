# Taproot DSLs

Taproot makes use of some basic XMLish DSLs, specifically the Acacia Policy Language and the Javascript Markup Language. 
Both Acacia and JSML use the same lexer logic (in the `./lexer` and `./token` directories), but separate parsers and 
transpilers/managers (in the `./acacia*` and `./jsml*` directories).
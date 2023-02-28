
Lexing GF files is straightforward, since GF itself doesn't currently do any interpretive heavy lifting; like Babel, it 
simply transpiles to JavaScript. Therefore, our primary concern is knowing the following:

- Are we generating text that should be output as-is (HTML)?
- Are we generating text that should be executed (JavaScript)?
- Are we processing a GF tag that generates specific JavaScript (go.* tag)?

For lexing, this simply means that we are generating the following tokens:

- Start of File
- Text (either HTML or JavaScript)
- Inside a start or self-closing tag
- Ending a start tag
- Ending a self-closing tag
- Processing an attribute ID, string, or numeric value
- Inside a closing tag
- Ending a closing tag

Parsing will then separate out the function of each token, and transpiling will generate the JS.